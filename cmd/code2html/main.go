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
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: code2html [-o output.html] <file>...")
		os.Exit(1)
	}

	if err := run(*outputFile, flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(outputFile string, files []string) (err error) {
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer func() { err = errors.Join(err, outFile.Close()) }()

	fmt.Fprintln(outFile, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Code</title>
    <style>
        p, summary, pre { font-size: larger; }
        .answer { border: 1px solid lightgray; padding: 0.5em; margin: 0.5em 0; }
    </style>
</head>
<body>`)

	for _, filename := range files {
		if err := processFile(outFile, filename); err != nil {
			return fmt.Errorf("error processing %s: %w", filename, err)
		}
	}

	fmt.Fprintln(outFile, `</body>
</html>`)

	return nil
}

func processFile(out *os.File, filename string) error {
	slide, err := scanFile(filename)
	if err != nil {
		return err
	}
	writeSlideHTML(out, slide)
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
			if inSection && currentKind != sectionCode && current.Len() > 0 {
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

func writeSlideHTML(w io.Writer, slide *Slide) {
	fmt.Fprintf(w, "<h1>%s</h1>\n", html.EscapeString(slide.heading))
	inAnswer := false
	for _, sec := range slide.sections {
		if sec.kind == sectionAnswer && !inAnswer {
			fmt.Fprintln(w, "<details>")
			fmt.Fprintln(w, "<summary>Answer</summary>")
			fmt.Fprintln(w, `<div class="answer">`)
			inAnswer = true
		} else if sec.kind != sectionAnswer && inAnswer {
			fmt.Fprintln(w, "</div>")
			fmt.Fprintln(w, "</details>")
			inAnswer = false
		}

		switch sec.kind {
		case sectionCode:
			fmt.Fprintf(w, "<code><pre>%s</pre></code>\n", renderCode(sec.content))
		case sectionNote, sectionQuestion, sectionAnswer:
			fmt.Fprintf(w, "<p>%s</p>\n", renderInlineCode(sec.content))
		}
	}
	if inAnswer {
		fmt.Fprintln(w, "</div>")
		fmt.Fprintln(w, "</details>")
	}
}

func renderCode(s string) string {
	s = html.EscapeString(s)
	s = strings.ReplaceAll(s, "\x00em\x00", "<b>")
	s = strings.ReplaceAll(s, "\x00/em\x00", "</b>")
	return s
}

func renderInlineCode(s string) string {
	var result strings.Builder
	inCode := false
	for _, r := range s {
		if r == '`' {
			if inCode {
				result.WriteString("</code>")
			} else {
				result.WriteString("<code>")
			}
			inCode = !inCode
		} else if inCode {
			result.WriteString(html.EscapeString(string(r)))
		} else {
			result.WriteString(html.EscapeString(string(r)))
		}
	}
	return result.String()
}
