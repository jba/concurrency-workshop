package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"strings"

	"rsc.io/markdown"
)

type Slide struct {
	heading  string
	sections []section
}

func (s *Slide) dump() {
	fmt.Printf("----------------\n")
	fmt.Printf("# %s\n", s.heading)
	for _, sec := range s.sections {
		sec.dump()
	}
}

type sectionKind int

const (
	sectionUndefined sectionKind = iota
	sectionNote
	sectionCode
	sectionQuestion
	sectionAnswer
	sectionText
	sectionHTML
	sectionOutput
)

func (k sectionKind) String() string {
	switch k {
	case sectionNote:
		return "note"
	case sectionCode:
		return "code"
	case sectionQuestion:
		return "question"
	case sectionAnswer:
		return "answer"
	case sectionText:
		return "text"
	case sectionHTML:
		return "html"
	case sectionOutput:
		return "output"
	default:
		return "unknown"
	}
}

var simpleKinds = map[string]sectionKind{
	"note":     sectionNote,
	"code":     sectionCode,
	"text":     sectionText,
	"output":   sectionOutput,
	"question": sectionQuestion,
}

type section struct {
	kind    sectionKind
	content string
}

func (s section) dump() {
	fmt.Printf("-- %s --\n", s.kind)
	fmt.Printf("%s", s.content)
	fmt.Printf("^^^^\n")
}

var (
	includeNotes bool
	debug        bool
)

func main() {
	outputFile := flag.String("o", "output.html", "output file name")
	title := flag.String("title", "Title", "presentation title")
	flag.BoolVar(&includeNotes, "notes", false, "include notes and answers in output")
	flag.BoolVar(&debug, "debug", false, "debug output")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: code2slides [-o output.html] [-notes] <file>...")
		os.Exit(1)
	}

	if err := run(*outputFile, *title, flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type indentWriter struct {
	w     io.Writer
	level int
	err   error
}

func (w *indentWriter) indent() {
	for range w.level {
		io.WriteString(w, "  ")
	}
}

func (w *indentWriter) open(s string) {
	w.indent()
	io.WriteString(w, s)
	fmt.Fprintln(w)
	w.level++
}

func (w *indentWriter) close(s string) {
	w.level--
	w.indent()
	io.WriteString(w, s)
	fmt.Fprintln(w)
}

func (w *indentWriter) linef(format string, args ...any) {
	w.indent()
	fmt.Fprintf(w, format, args...)
	fmt.Fprintln(w)
}

func (w *indentWriter) Write(data []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	n, err := w.w.Write(data)
	if err != nil {
		w.err = err
	}
	return n, err
}

func (w *indentWriter) Err() error { return w.err }

func run(outputFile, title string, files []string) (err error) {
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer func() { err = errors.Join(err, outFile.Close()) }()

	iw := &indentWriter{w: outFile}

	fmt.Fprintf(iw, top, title)

	for i, filename := range files {
		if err := processFile(iw, filename, i+1); err != nil {
			return fmt.Errorf("error processing %s: %w", filename, err)
		}
	}

	fmt.Fprintln(iw, bottom)

	return iw.Err()
}

func processFile(w *indentWriter, filename string, pageNum int) error {
	slide, err := scanFile(filename)
	if err != nil {
		return err
	}
	if debug {
		slide.dump()
	}

	w.linef("\n<!-- %s -->", filename)
	writeSlideHTML(w, slide, pageNum)
	return w.Err()
}

func scanFile(filename string) (_ *Slide, err error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	slide := &Slide{
		heading: filepath.Base(filename),
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	var (
		current  strings.Builder
		kind     sectionKind
		divClass string
	)
	lineNum := 0

	defer func() {
		if err != nil {
			err = fmt.Errorf("%s:%d: %v", filename, lineNum, err)
		}
	}()

	add := func(k sectionKind, c string) {
		slide.sections = append(slide.sections, section{kind: k, content: c})
	}

	addCurrent := func(k sectionKind) {
		if current.Len() > 0 {
			add(k, current.String())
			current.Reset()
		}
	}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		first, rest, _ := splitFirstWord(line)
		matchFirst := true
		if sec, ok := simpleKinds[first]; ok {
			if kind != sectionUndefined {
				return nil, fmt.Errorf("%s inside %s", sec, kind)
			}
			kind = sec
			continue
		}
		switch first {
		case "heading":
			if rest == "" {
				return nil, errors.New("missing heading")
			}
			slide.heading = rest

		case "html":
			add(sectionHTML, rest)

		case "!code":
			if kind != sectionCode {
				return nil, errors.New("!code without matching code")
			}
			// Trim trailing blank line
			add(kind, strings.TrimSuffix(current.String(), "\n"))
			current.Reset()
			kind = sectionUndefined
		case "!note":
			if kind != sectionNote {
				return nil, errors.New("!note without matching note")
			}
			addCurrent(sectionNote)
			kind = sectionUndefined

		case "!text":
			if kind != sectionText {
				return nil, errors.New("!text without matching text")
			}
			addCurrent(sectionText)
			kind = sectionUndefined
		case "!output":
			if kind != sectionOutput {
				return nil, fmt.Errorf("output inside %s", kind)
			}
			addCurrent(sectionOutput)
			kind = sectionUndefined

		case "answer":
			if kind != sectionQuestion {
				return nil, errors.New("answer without matching question")
			}
			addCurrent(sectionQuestion)
			kind = sectionAnswer

		case "!question":
			if kind != sectionQuestion && kind != sectionAnswer {
				return nil, errors.New("!question without matching question")
			}
			if kind == sectionQuestion {
				return nil, errors.New("!question without answer")
			}
			addCurrent(sectionAnswer)
			kind = sectionUndefined

		default:
			matchFirst = false
		}
		if !matchFirst {
			if d, c, ok := strings.Cut(first, "."); ok {
				if d == "div" {
					add(sectionHTML, fmt.Sprintf("<div class=%q>", c))
					divClass = c
					continue
				} else if d == "!div" {
					if c != divClass {
						return nil, fmt.Errorf("mismatched div class: start %q, end %q", divClass, c)
					}
					add(sectionHTML, fmt.Sprintf("</div> <!-- %s -->", c))
					divClass = ""
					// fmt.Printf("## !div %q\n", c)
					continue
				}
			}
			switch line {
			case "*/":
				if kind == sectionText {
					addCurrent(sectionText)
					kind = sectionUndefined
					continue
				}
				fallthrough
			default:
				// fmt.Printf("## line %q default section %s\n", line, kind)
				if kind == sectionCode {
					trimmed := strings.TrimLeft(line, " \t")
					switch trimmed {
					case "// em":
						current.WriteString("\x00em\x00")
					case "// !em":
						// Trim trailing blank line before closing em
						s := strings.TrimSuffix(current.String(), "\n")
						current.Reset()
						current.WriteString(s)
						current.WriteString("\x00/em\x00")
						current.WriteByte('\n')
					default:
						current.WriteString(line)
						current.WriteByte('\n')
					}
				} else if kind != sectionUndefined {
					// Strip // prefix if present
					text := strings.TrimSpace(strings.TrimPrefix(line, "//"))
					current.WriteString(text)
					current.WriteByte('\n')
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if kind != sectionUndefined {
		return nil, fmt.Errorf("unclosed %s section", kind)
	}
	if divClass != "" {
		return nil, fmt.Errorf("unclosed div with class %q", divClass)
	}

	return slide, nil
}

func splitFirstWord(s string) (string, string, bool) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "//") && !strings.HasPrefix(s, "/*") {
		return "", "", false
	}
	s = strings.TrimSpace(s[2:])
	i := strings.IndexAny(s, " \t")
	if i < 0 {
		return s, "", true
	}
	return s[:i], strings.TrimSpace(s[i+1:]), true
}

func writeSlideHTML(w *indentWriter, slide *Slide, pageNum int) {
	w.open("<article>")
	w.linef("<h1>%s</h1>", html.EscapeString(slide.heading))
	for _, sec := range slide.sections {
		switch sec.kind {
		case sectionCode:
			w.open("<div class='code'><pre>")
			fmt.Fprint(w, renderCode(sec.content))
			fmt.Fprintln(w, "</pre>") // indenting adds a blank line
			w.close("</div>")
		case sectionText:
			w.open("<div class='text'>")
			fmt.Fprint(w, renderMarkdown(sec.content))
			w.close("</div>")
		case sectionQuestion:
			fmt.Fprint(w, renderMarkdown(sec.content))
			fmt.Fprintln(w, "  <details><summary></summary>")
		case sectionAnswer:
			fmt.Fprint(w, renderMarkdown(sec.content))
			fmt.Fprintln(w, "  </details>")
		case sectionOutput:
			// Avoid two consecutive inline-block divs from appearing
			// next to each other.
			fmt.Fprintln(w, "<div></div>")
			w.open("<div class='output'><pre>")
			fmt.Fprint(w, sec.content)
			fmt.Fprintln(w, "</pre>") // indenting adds a blank line
			w.close("</div>")
		case sectionNote:
			if includeNotes {
				fmt.Fprint(w, renderMarkdown(sec.content))
			}
		case sectionHTML:
			w.linef("%s", sec.content)
		}
	}
	w.linef("<span class='pagenumber'>%d</span>", pageNum)
	w.close("</article>")
}

func renderCode(s string) string {
	s = strings.ReplaceAll(s, "\t", "    ")
	var result strings.Builder
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if i > 0 {
			result.WriteByte('\n')
		}
		// Split off comment if present
		code, comment := line, ""
		if idx := strings.Index(line, "//"); idx >= 0 {
			code, comment = line[:idx], line[idx:]
		}
		// Render code portion with definition highlighting
		result.WriteString(renderCodeLine(code))
		// Render comment if present
		if comment != "" {
			result.WriteString("<comment>")
			result.WriteString(html.EscapeString(comment))
			result.WriteString("</comment>")
		}
	}
	out := result.String()
	out = strings.ReplaceAll(out, "\x00em\x00", "<b>")
	out = strings.ReplaceAll(out, "\x00/em\x00", "</b>")
	return out
}

func renderCodeLine(line string) string {
	// Handle emphasis markers that may prefix the line
	prefix := ""
	if strings.HasPrefix(line, "\x00em\x00") {
		prefix = "\x00em\x00"
		line = strings.TrimPrefix(line, "\x00em\x00")
	} else if strings.HasPrefix(line, "\x00/em\x00") {
		prefix = "\x00/em\x00"
		line = strings.TrimPrefix(line, "\x00/em\x00")
	}

	trimmed := strings.TrimLeft(line, " \t")
	indent := line[:len(line)-len(trimmed)]

	// Check for type definition: "type NAME"
	if name, ok := strings.CutPrefix(trimmed, "type "); ok {
		// Extract the type name (first word)
		parts := strings.Fields(name)
		if len(parts) > 0 {
			typeName := parts[0]
			rest := strings.TrimPrefix(name, typeName)
			return prefix + html.EscapeString(indent) + "type <defn>" + html.EscapeString(typeName) + "</defn>" + html.EscapeString(rest)
		}
	}

	// Check for func/method definition: "func NAME(" or "func (receiver) NAME("
	if rest, ok := strings.CutPrefix(trimmed, "func "); ok {
		// Check if it's a method (starts with receiver)
		if strings.HasPrefix(rest, "(") {
			// Find closing paren of receiver
			if idx := strings.Index(rest, ") "); idx >= 0 {
				receiver := rest[:idx+1]
				afterReceiver := rest[idx+2:]
				// Extract method name
				if parenIdx := strings.Index(afterReceiver, "("); parenIdx >= 0 {
					methodName := afterReceiver[:parenIdx]
					afterName := afterReceiver[parenIdx:]
					return prefix + html.EscapeString(indent) + "func " + html.EscapeString(receiver+" ") + "<defn>" + html.EscapeString(methodName) + "</defn>" + html.EscapeString(afterName)
				}
			}
		} else {
			// Regular function: "func NAME("
			if parenIdx := strings.Index(rest, "("); parenIdx >= 0 {
				funcName := rest[:parenIdx]
				afterName := rest[parenIdx:]
				return prefix + html.EscapeString(indent) + "func <defn>" + html.EscapeString(funcName) + "</defn>" + html.EscapeString(afterName)
			}
		}
	}

	return prefix + html.EscapeString(line)
}

func renderMarkdown(s string) string {
	var p markdown.Parser
	p.Table = true
	doc := p.Parse(s)
	return markdown.ToHTML(doc)
}

const top = `<!DOCTYPE html>
<html>
  <head>
    <title>%s</title>
    <meta charset='utf-8'>
    <script>
      var notesEnabled =  false ;
    </script>
    <script src='static/slides.js'></script>
  </head>

  <body style='display: none'>
    <section class='slides'>
`

const bottom = `
    <div id="help">
      Use the left and right arrow keys or click the left and right
      edges of the page to navigate between slides.<br>
      (Press 'H' or navigate to hide this message.)
    </div>
    <script type="application/javascript" src='static/play.js'></script>
  </body>
</html>`
