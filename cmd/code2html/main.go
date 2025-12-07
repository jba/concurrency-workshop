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
	heading      string
	codeSections []string
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
	var currentCode strings.Builder
	inCode := false

	for scanner.Scan() {
		line := scanner.Text()
		switch line {
		case "// code":
			inCode = true
			currentCode.Reset()
		case "// !code":
			if inCode {
				slide.codeSections = append(slide.codeSections, currentCode.String())
				inCode = false
			}
		default:
			if h, ok := strings.CutPrefix(line, "// heading "); ok {
				slide.heading = h
			} else if inCode {
				currentCode.WriteString(line)
				currentCode.WriteByte('\n')
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return slide, nil
}

func writeSlideHTML(w io.Writer, slide *Slide) {
	fmt.Fprintf(w, "<h1>%s</h1>\n", html.EscapeString(slide.heading))
	for _, code := range slide.codeSections {
		fmt.Fprintf(w, "<code><pre>%s</pre></code>\n", html.EscapeString(code))
	}
}
