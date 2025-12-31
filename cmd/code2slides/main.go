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

type sectionKind int

const (
	sectionNote sectionKind = iota
	sectionCode
	sectionQuestion
	sectionAnswer
	sectionText
)

type section struct {
	kind    sectionKind
	content string
}

func main() {
	outputFile := flag.String("o", "output.html", "output file name")
	title := flag.String("title", "Title", "presentation title")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: code2slides [-o output.html] <file>...")
		os.Exit(1)
	}

	if err := run(*outputFile, *title, flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type errWriter struct {
	w   io.Writer
	err error
}

func (w *errWriter) Write(data []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	n, err := w.w.Write(data)
	if err != nil {
		w.err = err
	}
	return n, err
}

func (w *errWriter) Err() error { return w.err }

func run(outputFile, title string, files []string) (err error) {
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer func() { err = errors.Join(err, outFile.Close()) }()

	ew := &errWriter{w: outFile}

	fmt.Fprintf(ew, top, title)

	for i, filename := range files {
		if err := processFile(ew, filename, i); err != nil {
			return fmt.Errorf("error processing %s: %w", filename, err)
		}
	}

	fmt.Fprintln(ew, bottom)

	return ew.Err()
}

func processFile(w io.Writer, filename string, pageNum int) error {
	slide, err := scanFile(filename)
	if err != nil {
		return err
	}
	writeSlideHTML(w, slide, pageNum)
	return nil
}

func scanFile(filename string) (*Slide, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	slide := &Slide{
		heading: filepath.Base(filename),
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	var current strings.Builder
	var currentKind sectionKind
	inSection := false
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		switch line {
		case "// code":
			if inSection {
				return nil, fmt.Errorf("%s:%d: code inside %s", filename, lineNum, kindName(currentKind))
			}
			currentKind = sectionCode
			inSection = true
			current.Reset()
		case "// !code":
			if !inSection || currentKind != sectionCode {
				return nil, fmt.Errorf("%s:%d: !code without matching code", filename, lineNum)
			}
			// Trim trailing blank line
			content := strings.TrimSuffix(current.String(), "\n")
			slide.sections = append(slide.sections, section{kind: sectionCode, content: content})
			inSection = false
		case "// note":
			if inSection {
				return nil, fmt.Errorf("%s:%d: note inside %s", filename, lineNum, kindName(currentKind))
			}
			currentKind = sectionNote
			inSection = true
			current.Reset()
		case "// !note":
			if !inSection || currentKind != sectionNote {
				return nil, fmt.Errorf("%s:%d: !note without matching note", filename, lineNum)
			}
			if current.Len() > 0 {
				slide.sections = append(slide.sections, section{kind: sectionNote, content: current.String()})
			}
			inSection = false
		case "// text":
			if inSection {
				return nil, fmt.Errorf("%s:%d: text inside %s", filename, lineNum, kindName(currentKind))
			}
			currentKind = sectionText
			inSection = true
			current.Reset()
		case "// !text":
			if !inSection || currentKind != sectionText {
				return nil, fmt.Errorf("%s:%d: !text without matching text", filename, lineNum)
			}
			if current.Len() > 0 {
				slide.sections = append(slide.sections, section{kind: sectionText, content: current.String()})
			}
			inSection = false
		case "// question":
			if inSection {
				return nil, fmt.Errorf("%s:%d: question inside %s", filename, lineNum, kindName(currentKind))
			}
			currentKind = sectionQuestion
			inSection = true
			current.Reset()
		case "// answer":
			if !inSection || currentKind != sectionQuestion {
				return nil, fmt.Errorf("%s:%d: answer without matching question", filename, lineNum)
			}
			if current.Len() > 0 {
				slide.sections = append(slide.sections, section{kind: sectionQuestion, content: current.String()})
			}
			currentKind = sectionAnswer
			current.Reset()
		case "// !question":
			if !inSection || (currentKind != sectionQuestion && currentKind != sectionAnswer) {
				return nil, fmt.Errorf("%s:%d: !question without matching question", filename, lineNum)
			}
			if currentKind == sectionQuestion {
				return nil, fmt.Errorf("%s:%d: !question without answer", filename, lineNum)
			}
			if current.Len() > 0 {
				slide.sections = append(slide.sections, section{kind: sectionAnswer, content: current.String()})
			}
			inSection = false
		case "//", "":
			if inSection && currentKind == sectionCode {
				current.WriteByte('\n')
			} else if inSection && current.Len() > 0 {
				slide.sections = append(slide.sections, section{kind: currentKind, content: current.String()})
				current.Reset()
			}
		default:
			if h, ok := strings.CutPrefix(line, "// heading "); ok {
				slide.heading = h
			} else if inSection && currentKind == sectionCode {
				trimmed := strings.TrimLeft(line, " \t")
				if trimmed == "// em" {
					current.WriteString("\x00em\x00")
				} else if trimmed == "// !em" {
					// Trim trailing blank line before closing em
					s := strings.TrimSuffix(current.String(), "\n")
					current.Reset()
					current.WriteString(s)
					current.WriteString("\x00/em\x00")
				} else {
					current.WriteString(line)
					current.WriteByte('\n')
				}
			} else if inSection {
				text, _ := strings.CutPrefix(line, "// ")
				current.WriteString(text)
				current.WriteByte('\n')
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if inSection {
		return nil, fmt.Errorf("%s:%d: unclosed %s section", filename, lineNum, kindName(currentKind))
	}

	return slide, nil
}

func kindName(k sectionKind) string {
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
	}
	return "unknown"
}

func writeSlideHTML(w io.Writer, slide *Slide, pageNum int) {
	fmt.Fprintln(w, "<article>")
	fmt.Fprintf(w, "  <h1>%s</h1>\n", html.EscapeString(slide.heading))
	for _, sec := range slide.sections {
		switch sec.kind {
		case sectionCode:
			fmt.Fprintf(w, "    <div class='code'><pre>%s</pre></div>\n", renderCode(sec.content))
		case sectionText:
			fmt.Fprint(w, renderMarkdown(sec.content))
		case sectionQuestion:
			fmt.Fprint(w, renderMarkdown(sec.content))
			fmt.Fprintln(w, "  <details><summary></summary>")
		case sectionAnswer:
			fmt.Fprint(w, renderMarkdown(sec.content))
			fmt.Fprintln(w, "  </details>")
		case sectionNote:
			// Notes are not rendered
		}
	}

	fmt.Fprintf(w, "<span class='pagenumber'>%d</span>\n", pageNum)
	fmt.Fprintln(w, "</article>")
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
    <script src='/static/slides.js'></script>
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
