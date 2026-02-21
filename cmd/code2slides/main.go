// code2slides converts Go source files into HTML slide presentations.
//
// Each input file contains Go code annotated with directives in comments.
// Directives use // or /* prefixes and control how the file is rendered into slides.
// Lines that are not inside a directive block are ignored (unless inside a code or
// other block section).
//
// A single file can produce multiple slides; each "heading" directive starts a new one.
//
// # Directives
//
// heading TEXT
//
//	Set the slide's heading to TEXT. Each heading starts a new slide.
//
// problem
//
//	Mark the current slide as a problem slide. The area around the slide
//	turns red when this slide is displayed.
//
// code / !code
//
//	Begin and end a code block. Lines between these directives are rendered
//	as preformatted source code. Comments in the code are syntax-highlighted.
//	Type and function definitions are highlighted as well.
//
// code bad / !code
//
//	Like code / !code, but the code block is rendered with a red background
//	to indicate incorrect or problematic code.
//
// note / !note
//
//	Begin and end a presenter note block. Lines between these directives are
//	rendered as markdown. Notes are only included in the output when the
//	-notes flag is set.
//
// text / !text
//
//	Begin and end a text block. Lines between these directives are rendered
//	as markdown. A line containing only "*/" also closes a text block (to
//	support /* text ... */ style comments).
//
// text CONTENT (inline form)
//
//	If text is followed by content on the same line, that content is used as
//	a single-line text section rendered as markdown. There is no matching
//	"!text" for this form.
//
// output / !output
//
//	Begin and end an output block. Lines between these directives are rendered
//	as preformatted text with a dark background, representing program output.
//
// question / answer / !question
//
//	Define a question-and-answer section. "question" starts the question text,
//	"answer" ends the question and starts the answer, and "!question" closes
//	the whole block. The answer is hidden behind a <details> toggle. Both
//	question and answer content are rendered as markdown.
//
// html CONTENT
//
//	Emit CONTENT as raw HTML in the slide.
//
// div.CLASS / !div.CLASS
//
//	Open and close a <div> with the given CSS class. The class must match
//	between the opening and closing directives.
//
// em / !em
//
//	Inside a code block, these directives bold (emphasize) the enclosed lines.
//
// em REGEXP (inline form)
//
//	Inside a code block, a trailing "// em REGEXP" on a code line emphasizes
//	all portions of that line (before the "// em") that match the regular
//	expression. The "// em REGEXP" suffix is stripped from the output.
//	There is no matching "// !em" for this form.
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
	"regexp"
	"strings"

	"rsc.io/markdown"
)

type Slide struct {
	heading  string
	problem  bool
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
	sectionCodeBad
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
	case sectionCodeBad:
		return "code bad"
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
	emStyle      = "bold"
)

func main() {
	outputFile := flag.String("o", "output.html", "output file name")
	title := flag.String("title", "Title", "presentation title")
	flag.BoolVar(&includeNotes, "notes", false, "include notes and answers in output")
	flag.BoolVar(&debug, "debug", false, "debug output")
	flag.StringVar(&emStyle, "em", "bold", "emphasis style: 'bold' or a color name")
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

	// Write title slide
	iw.open("<article class='title-slide'>")
	iw.linef("<div class='title-text'>%s</div>", html.EscapeString(title))
	iw.close("</article>")

	pageNum := 2
	for _, filename := range files {
		var err error
		pageNum, err = processFile(iw, filename, pageNum)
		if err != nil {
			return fmt.Errorf("error processing %s: %w", filename, err)
		}
	}

	fmt.Fprintln(iw, bottom)

	return iw.Err()
}

func processFile(w *indentWriter, filename string, pageNum int) (int, error) {
	slides, err := scanFile(filename)
	if err != nil {
		return 0, err
	}

	w.linef("\n<!-- %s -->", filename)
	for _, slide := range slides {
		if debug {
			slide.dump()
		}
		writeSlideHTML(w, slide, pageNum)
		pageNum++
	}
	return pageNum, w.Err()
}

func scanFile(filename string) (_ []*Slide, err error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	slide := &Slide{
		heading: filepath.Base(filename),
	}
	var slides []*Slide

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
		// Handle "code bad" before simpleKinds
		if first == "code" && rest == "bad" {
			if kind != sectionUndefined {
				return nil, fmt.Errorf("code bad inside %s", kind)
			}
			kind = sectionCodeBad
			continue
		}
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
			if len(slide.sections) > 0 {
				slides = append(slides, slide)
				slide = &Slide{}
			}
			slide.heading = rest

		case "problem":
			return nil, errors.New("'problem' temporarily not supported")
			// slide.problem = true

		case "text":
			if kind != sectionUndefined {
				return nil, fmt.Errorf("text inside %s", kind)
			}
			if rest != "" {
				add(sectionText, rest+"\n")
			} else {
				kind = sectionText
			}

		case "html":
			add(sectionHTML, rest)

		case "!code":
			if kind != sectionCode && kind != sectionCodeBad {
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
				if kind == sectionCode || kind == sectionCodeBad {
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
						// Check for inline em: code // em PATTERN
						if idx := strings.Index(line, "// em "); idx >= 0 {
							pattern := strings.TrimSpace(line[idx+len("// em "):])
							if pattern != "" {
								re, err := regexp.Compile(pattern)
								if err != nil {
									return nil, fmt.Errorf("invalid em regexp %q: %w", pattern, err)
								}
								codePart := strings.TrimRight(line[:idx], " \t")
								marked := re.ReplaceAllStringFunc(codePart, func(m string) string {
									return "\x00em\x00" + m + "\x00/em\x00"
								})
								current.WriteString(marked)
								current.WriteByte('\n')
								break
							}
						}
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

	slides = append(slides, slide)
	return slides, nil
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
	if slide.problem {
		w.open("<article class='problem'>")
	} else {
		w.open("<article>")
	}
	w.linef("<h1>%s</h1>", html.EscapeString(slide.heading))
	for _, sec := range slide.sections {
		switch sec.kind {
		case sectionCode:
			w.open("<div class='code'><pre>")
			fmt.Fprint(w, renderCode(sec.content))
			fmt.Fprintln(w, "</pre>") // indenting adds a blank line
			w.close("</div>")
		case sectionCodeBad:
			w.open("<div class='code bad'><pre>")
			fmt.Fprint(w, renderCode(sec.content))
			fmt.Fprintln(w, "</pre>")
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

var identRe = regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*`)

// stripUnderscoreSuffixes removes underscore suffixes from identifiers.
// For example, "foo_3x" becomes "foo". Identifiers starting with an
// underscore (like "_private") are left unchanged.
func stripUnderscoreSuffixes(s string) string {
	return identRe.ReplaceAllStringFunc(s, func(m string) string {
		if i := strings.Index(m, "_"); i > 0 {
			return m[:i]
		}
		return m
	})
}

func renderCode(s string) string {
	s = strings.ReplaceAll(s, "\t", "    ")
	lines := strings.Split(s, "\n")

	// Find minimum indentation across all non-empty lines
	minIndent := -1
	for _, line := range lines {
		// Strip em marker prefix to find actual content
		content := line
		for strings.HasPrefix(content, "\x00em\x00") || strings.HasPrefix(content, "\x00/em\x00") {
			if strings.HasPrefix(content, "\x00em\x00") {
				content = content[len("\x00em\x00"):]
			} else {
				content = content[len("\x00/em\x00"):]
			}
		}
		if strings.TrimSpace(content) == "" {
			continue
		}
		indent := len(content) - len(strings.TrimLeft(content, " "))
		if minIndent < 0 || indent < minIndent {
			minIndent = indent
		}
	}
	// Remove common indentation
	if minIndent > 0 {
		for i, line := range lines {
			// Extract em marker prefix
			prefix := ""
			content := line
			for strings.HasPrefix(content, "\x00em\x00") || strings.HasPrefix(content, "\x00/em\x00") {
				if strings.HasPrefix(content, "\x00em\x00") {
					prefix += "\x00em\x00"
					content = content[len("\x00em\x00"):]
				} else {
					prefix += "\x00/em\x00"
					content = content[len("\x00/em\x00"):]
				}
			}
			if len(content) >= minIndent {
				lines[i] = prefix + content[minIndent:]
			}
		}
	}

	var result strings.Builder
	for i, line := range lines {
		if i > 0 {
			result.WriteByte('\n')
		}
		line = stripUnderscoreSuffixes(line)
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
	var emOpen, emClose string
	if emStyle == "bold" {
		emOpen, emClose = "<b>", "</b>"
	} else {
		emOpen = fmt.Sprintf("<span style=\"color: %s\">", emStyle)
		emClose = "</span>"
	}
	out = strings.ReplaceAll(out, "\x00em\x00", emOpen)
	out = strings.ReplaceAll(out, "\x00/em\x00", emClose)
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
    <link rel='icon' type='image/svg+xml' href='static/favicon.svg'>
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
