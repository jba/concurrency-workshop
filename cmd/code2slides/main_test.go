package main

import (
	"strings"
	"testing"
)

func sectionsEqual(a, b []section) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].equal(b[i]) {
			return false
		}
	}
	return true
}

func TestScanFileErrors(t *testing.T) {
	tests := []struct {
		file    string
		wantErr string
	}{
		{"testdata/unmatched_endcode.go", "!code without matching code"},
		{"testdata/unmatched_endnote.go", "!note without matching note"},
		{"testdata/code_inside_note.go", "code inside note"},
		{"testdata/note_inside_code.go", "note inside code"},
		{"testdata/unclosed_code.go", "unclosed code section"},
		{"testdata/unclosed_note.go", "unclosed note section"},
		{"testdata/unclosed_question.go", "unclosed answer section"},
		{"testdata/unmatched_endquestion.go", "!question without matching question"},
		{"testdata/question_without_answer.go", "!question without answer"},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			_, err := scanFile(tt.file)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestScanFile(t *testing.T) {
	slides, err := scanFile("testdata/valid.go")
	if err != nil {
		t.Fatal(err)
	}

	if len(slides) != 1 {
		t.Fatalf("got %d slides, want 1", len(slides))
	}
	slide := slides[0]

	if slide.heading != "Test Heading" {
		t.Errorf("heading = %q, want %q", slide.heading, "Test Heading")
	}

	wantSections := []section{
		{kind: sectionNote, content: "First note.\n"},
		{kind: sectionCode, content: "func foo() {}"},
		{kind: sectionNote, content: "Second note.\n\nThird note after blank comment.\n\nFourth note after blank line.\n"},
		{kind: sectionCode, content: "func bar() {}"},
		{kind: sectionQuestion, content: "What is the answer?\n"},
		{kind: sectionAnswer, content: "The answer is 42.\n"},
		{kind: sectionNote, content: "Use `fmt.Println` to print.\n"},
	}

	if !sectionsEqual(slide.sections, wantSections) {
		t.Errorf("got:\n%v\nwant:\n%v", slide.sections, wantSections)
	}
}

func TestElide(t *testing.T) {
	slides, err := scanFile("testdata/elide_test.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(slides) != 1 {
		t.Fatalf("got %d slides, want 1", len(slides))
	}
	slide := slides[0]
	if len(slide.sections) != 1 {
		t.Fatalf("got %d sections, want 1", len(slide.sections))
	}
	sec := slide.sections[0]
	if sec.kind != sectionCode {
		t.Fatalf("got section kind %v, want code", sec.kind)
	}
	want := "func example() {\n\tx := 1\n\t// ...\n\tfmt.Println(x)\n}"
	if sec.content != want {
		t.Errorf("got:\n%q\nwant:\n%q", sec.content, want)
	}
}

func TestInlineEmMulti(t *testing.T) {
	slides, err := scanFile("testdata/inline_em_multi.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(slides) != 1 {
		t.Fatalf("got %d slides, want 1", len(slides))
	}
	slide := slides[0]
	if len(slide.sections) != 1 {
		t.Fatalf("got %d sections, want 1", len(slide.sections))
	}
	sec := slide.sections[0]
	if sec.kind != sectionCode {
		t.Fatalf("got section kind %v, want code", sec.kind)
	}
	// Both foo and bar should be wrapped with em markers
	want := "x, y := \x00em\x00foo\x00/em\x00(), \x00em\x00bar\x00/em\x00()"
	if sec.content != want {
		t.Errorf("got:\n%q\nwant:\n%q", sec.content, want)
	}
}

func TestCodeInAnswer(t *testing.T) {
	slides, err := scanFile("testdata/code_in_answer.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(slides) != 1 {
		t.Fatalf("got %d slides, want 1", len(slides))
	}
	slide := slides[0]
	// Should have: question, answer (before code), code (inAnswer), answer (after code)
	wantSections := []section{
		{kind: sectionQuestion, content: "How do you print hello?\n"},
		{kind: sectionAnswer, content: "Use fmt.Println:\n"},
		{kind: sectionCode, content: "fmt.Println(\"hello\")", inAnswer: true},
		{kind: sectionAnswer, content: "That's it!\n"},
	}
	if !sectionsEqual(slide.sections, wantSections) {
		t.Errorf("got:\n%v\nwant:\n%v", slide.sections, wantSections)
	}
}

func TestCodeInAnswerHTML(t *testing.T) {
	slides, err := scanFile("testdata/code_in_answer.go")
	if err != nil {
		t.Fatal(err)
	}
	slide := slides[0]

	var buf strings.Builder
	w := &indentWriter{w: &buf}
	writeSlideHTML(w, slide, 1, false)
	html := buf.String()

	// The code should appear between <details> and </details>
	detailsStart := strings.Index(html, "<details>")
	detailsEnd := strings.Index(html, "</details>")
	codeStart := strings.Index(html, "<div class='code'>")

	if detailsStart == -1 || detailsEnd == -1 || codeStart == -1 {
		t.Fatalf("missing expected HTML elements in:\n%s", html)
	}
	if !(detailsStart < codeStart && codeStart < detailsEnd) {
		t.Errorf("code block not inside details block:\ndetails starts at %d, code at %d, details ends at %d\n%s",
			detailsStart, codeStart, detailsEnd, html)
	}
}

func TestRenderMarkdown(t *testing.T) {
	got := renderMarkdown("Use `fmt.Println` to print.\n")
	want := "<p>Use <code>fmt.Println</code> to print.</p>\n"
	if got != want {
		t.Errorf("renderMarkdown() = %q, want %q", got, want)
	}
}

func TestSplitFirstWord(t *testing.T) {
	tests := []struct {
		input    string
		wantWord string
		wantRest string
		wantOK   bool
	}{
		{"// code", "code", "", true},
		{"// heading Title", "heading", "Title", true},
		{"/* text", "text", "", true},
		{"// html <div>foo</div>", "html", "<div>foo</div>", true},
		{"//code", "code", "", true},
		{"//  spaced   rest", "spaced", "rest", true},
		{"not a comment", "", "", false},
		{"/ not a comment", "", "", false},
	}
	for _, tt := range tests {
		word, rest, ok := splitFirstWord(tt.input)
		if word != tt.wantWord || rest != tt.wantRest || ok != tt.wantOK {
			t.Errorf("splitFirstWord(%q) = (%q, %q, %v), want (%q, %q, %v)",
				tt.input, word, rest, ok, tt.wantWord, tt.wantRest, tt.wantOK)
		}
	}
}

func TestDivClass(t *testing.T) {
	slides, err := scanFile("testdata/div_test.go")
	if err != nil {
		t.Fatal(err)
	}

	if len(slides) != 1 {
		t.Fatalf("got %d slides, want 1", len(slides))
	}

	wantSections := []section{
		{kind: sectionHTML, content: `<div class="flex">`},
		{kind: sectionCode, content: "x := 1"},
		{kind: sectionHTML, content: "</div> <!-- flex -->"},
	}

	if !sectionsEqual(slides[0].sections, wantSections) {
		t.Errorf("got:\n%v\nwant:\n%v", slides[0].sections, wantSections)
	}
}

func TestDivClassMismatch(t *testing.T) {
	_, err := scanFile("testdata/div_mismatch.go")
	if err == nil {
		t.Fatal("expected error for mismatched div class")
	}
	if !strings.Contains(err.Error(), "mismatched div class") {
		t.Errorf("error = %q, want error containing 'mismatched div class'", err)
	}
}

func TestCodeBad(t *testing.T) {
	slides, err := scanFile("testdata/code_bad.go")
	if err != nil {
		t.Fatal(err)
	}

	if len(slides) != 1 {
		t.Fatalf("got %d slides, want 1", len(slides))
	}

	wantSections := []section{
		{kind: sectionCode, options: []string{"bad"}, content: "x := 1 // wrong"},
	}

	if !sectionsEqual(slides[0].sections, wantSections) {
		t.Errorf("got:\n%v\nwant:\n%v", slides[0].sections, wantSections)
	}
}

func TestInlineEm(t *testing.T) {
	slides, err := scanFile("testdata/inline_em.go")
	if err != nil {
		t.Fatal(err)
	}

	if len(slides) != 1 {
		t.Fatalf("got %d slides, want 1", len(slides))
	}

	wantSections := []section{
		{kind: sectionCode, content: "x := \x00em\x00foo\x00/em\x00()\ny := bar()"},
	}

	if !sectionsEqual(slides[0].sections, wantSections) {
		t.Errorf("got:\n%v\nwant:\n%v", slides[0].sections, wantSections)
	}

	// Verify rendered HTML
	got := renderCode(slides[0].sections[0].content)
	if !strings.Contains(got, "<span class=\"em\">foo</span>") {
		t.Errorf("rendered code does not contain <span class=\"em\">foo</span>: %s", got)
	}
	if strings.Contains(got, "// em") {
		t.Errorf("rendered code still contains // em: %s", got)
	}
}

func TestInlineEmWholeLine(t *testing.T) {
	slides, err := scanFile("testdata/inline_em_whole_line.go")
	if err != nil {
		t.Fatal(err)
	}

	if len(slides) != 1 {
		t.Fatalf("got %d slides, want 1", len(slides))
	}

	wantSections := []section{
		{kind: sectionCode, content: "\x00em\x00x := foo()\x00/em\x00\ny := bar()"},
	}

	if !sectionsEqual(slides[0].sections, wantSections) {
		t.Errorf("got:\n%v\nwant:\n%v", slides[0].sections, wantSections)
	}

	// Verify rendered HTML
	got := renderCode(slides[0].sections[0].content)
	if !strings.Contains(got, "<span class=\"em\">x := foo()</span>") {
		t.Errorf("rendered code does not contain whole line em: %s", got)
	}
	if strings.Contains(got, "// em") {
		t.Errorf("rendered code still contains // em: %s", got)
	}
}

func TestImage(t *testing.T) {
	slides, err := scanFile("testdata/image_test.go")
	if err != nil {
		t.Fatal(err)
	}

	if len(slides) != 1 {
		t.Fatalf("got %d slides, want 1", len(slides))
	}

	wantSections := []section{
		{kind: sectionHTML, content: `<img src="testdata/diagram.png" alt="diagram.png" />`},
		{kind: sectionHTML, content: `<img src="testdata/photo.jpg" alt="photo.jpg" />`},
	}

	if !sectionsEqual(slides[0].sections, wantSections) {
		t.Errorf("got:\n%v\nwant:\n%v", slides[0].sections, wantSections)
	}
}

func TestImageMissingFilename(t *testing.T) {
	_, err := scanFile("testdata/image_missing.go")
	if err == nil {
		t.Fatal("expected error for missing image filename")
	}
	if !strings.Contains(err.Error(), "missing image filename") {
		t.Errorf("error = %q, want error containing 'missing image filename'", err)
	}
}

func TestLink(t *testing.T) {
	slides, err := scanFile("testdata/link_test.go")
	if err != nil {
		t.Fatal(err)
	}

	if len(slides) != 1 {
		t.Fatalf("got %d slides, want 1", len(slides))
	}

	wantSections := []section{
		{kind: sectionHTML, content: `<a href="testdata/doc.html">See the documentation</a>`},
		{kind: sectionHTML, content: `<a href="testdata/other/file.go">View source code</a>`},
	}

	if !sectionsEqual(slides[0].sections, wantSections) {
		t.Errorf("got:\n%v\nwant:\n%v", slides[0].sections, wantSections)
	}
}

func TestLinkMissingFilename(t *testing.T) {
	_, err := scanFile("testdata/link_missing_file.go")
	if err == nil {
		t.Fatal("expected error for missing link filename")
	}
	if !strings.Contains(err.Error(), "missing link filename") {
		t.Errorf("error = %q, want error containing 'missing link filename'", err)
	}
}

func TestLinkMissingText(t *testing.T) {
	_, err := scanFile("testdata/link_missing_text.go")
	if err == nil {
		t.Fatal("expected error for missing link text")
	}
	if !strings.Contains(err.Error(), "missing link text") {
		t.Errorf("error = %q, want error containing 'missing link text'", err)
	}
}

func TestRenderCode(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: "x := 1 // comment\n",
			want:  "<span class='codenum'>1</span>x := 1 <comment>// comment</comment>\n",
		},
		{
			input: "type Foo struct {}\n",
			want:  "<span class='codenum'>1</span>type <defn>Foo</defn> struct {}\n",
		},
		{
			input: "func bar() {}\n",
			want:  "<span class='codenum'>1</span>func <defn>bar</defn>() {}\n",
		},
		{
			input: "func (*Foo) moo() {}\n",
			want:  "<span class='codenum'>1</span>func (*Foo) <defn>moo</defn>() {}\n",
		},
		{
			// Inline em markers (as produced by scanFile)
			input: "x := \x00em\x00foo\x00/em\x00()\n",
			want:  "<span class='codenum'>1</span>x := <span class=\"em\">foo</span>()\n",
		},
		{
			input: "func (f Foo) moo() {}\n",
			want:  "<span class='codenum'>1</span>func (f Foo) <defn>moo</defn>() {}\n",
		},
		{
			// Underscore suffix stripping
			input: "x := foo_3x(bar_v2)\n",
			want:  "<span class='codenum'>1</span>x := foo(bar)\n",
		},
		{
			// Leading underscore preserved
			input: "_private := 1\n",
			want:  "<span class='codenum'>1</span>_private := 1\n",
		},
		{
			// Underscore suffix on func def
			input: "func doThing_2() {}\n",
			want:  "<span class='codenum'>1</span>func <defn>doThing</defn>() {}\n",
		},
	}
	for _, tt := range tests {
		got := renderCode(tt.input)
		if got != tt.want {
			t.Errorf("renderCode(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
