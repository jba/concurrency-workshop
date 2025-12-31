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
			slide.sections = append(slide.sections, section{kind: sectionCode, content: current.String()})
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
	}
	return "unknown"
}

func writeSlideHTML(w io.Writer, slide *Slide, pageNum int) {
	fmt.Fprintln(w, "<article>")
	fmt.Fprintf(w, "  <h1>%s</h1>\n", html.EscapeString(slide.heading))
	inAnswer := false
	for _, sec := range slide.sections {
		if sec.kind == sectionAnswer && !inAnswer {
			fmt.Fprintln(w, "  <details>")
			fmt.Fprintln(w, "    <summary>Answer</summary>")
			fmt.Fprintln(w, `    <div class="answer">`)
			inAnswer = true
		} else if sec.kind != sectionAnswer && inAnswer {
			fmt.Fprintln(w, "    </div>")
			fmt.Fprintln(w, "  </details>")
			inAnswer = false
		}

		switch sec.kind {
		case sectionCode:
			fmt.Fprintf(w, "    <div class='code'><pre>%s</pre></div>\n", renderCode(sec.content))
		case sectionNote, sectionQuestion, sectionAnswer:
			fmt.Fprint(w, renderMarkdown(sec.content))
		}
	}
	if inAnswer {
		fmt.Fprintln(w, "    </div>")
		fmt.Fprintln(w, "  </details>")
	}

	fmt.Fprintf(w, "<span class='pagenumber'>%d</span>\n", pageNum)
	fmt.Fprintln(w, "</article>")
}

func renderCode(s string) string {
	var result strings.Builder
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if i > 0 {
			result.WriteByte('\n')
		}
		// Check for line comment
		if idx := strings.Index(line, "//"); idx >= 0 {
			result.WriteString(html.EscapeString(line[:idx]))
			result.WriteString("<i>")
			result.WriteString(html.EscapeString(line[idx:]))
			result.WriteString("</i>")
		} else {
			result.WriteString(html.EscapeString(line))
		}
	}
	out := result.String()
	out = strings.ReplaceAll(out, "\x00em\x00", "<b>")
	out = strings.ReplaceAll(out, "\x00/em\x00", "</b>")
	return out
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
