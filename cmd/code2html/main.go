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

type section struct {
	isCode  bool
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
	inCode := false
	inNote := false
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		switch line {
		case "// code":
			if inNote {
				return nil, fmt.Errorf("%s:%d: code inside note", filename, lineNum)
			}
			inCode = true
			current.Reset()
		case "// !code":
			if !inCode {
				return nil, fmt.Errorf("%s:%d: !code without matching code", filename, lineNum)
			}
			slide.sections = append(slide.sections, section{isCode: true, content: current.String()})
			inCode = false
		case "// note":
			if inCode {
				return nil, fmt.Errorf("%s:%d: note inside code", filename, lineNum)
			}
			inNote = true
			current.Reset()
		case "// !note":
			if !inNote {
				return nil, fmt.Errorf("%s:%d: !note without matching note", filename, lineNum)
			}
			slide.sections = append(slide.sections, section{isCode: false, content: current.String()})
			inNote = false
		default:
			if h, ok := strings.CutPrefix(line, "// heading "); ok {
				slide.heading = h
			} else if inCode {
				current.WriteString(line)
				current.WriteByte('\n')
			} else if inNote {
				text, _ := strings.CutPrefix(line, "// ")
				current.WriteString(text)
				current.WriteByte('\n')
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if inCode {
		return nil, fmt.Errorf("%s:%d: unclosed code section", filename, lineNum)
	}
	if inNote {
		return nil, fmt.Errorf("%s:%d: unclosed note section", filename, lineNum)
	}

	return slide, nil
}

func writeSlideHTML(w io.Writer, slide *Slide) {
	fmt.Fprintf(w, "<h1>%s</h1>\n", html.EscapeString(slide.heading))
	for _, sec := range slide.sections {
		if sec.isCode {
			fmt.Fprintf(w, "<code><pre>%s</pre></code>\n", html.EscapeString(sec.content))
		} else {
			fmt.Fprintf(w, "<p>%s</p>\n", html.EscapeString(sec.content))
		}
	}
}
