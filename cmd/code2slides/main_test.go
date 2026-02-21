package main

import (
	"slices"
	"strings"
	"testing"
)

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

	if !slices.Equal(slide.sections, wantSections) {
		t.Errorf("got:\n%v\nwant:\n%v", slide.sections, wantSections)
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

	if !slices.Equal(slides[0].sections, wantSections) {
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
		{kind: sectionCodeBad, content: "x := 1 // wrong"},
	}

	if !slices.Equal(slides[0].sections, wantSections) {
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

	if !slices.Equal(slides[0].sections, wantSections) {
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

func TestRenderCode(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: "x := 1 // comment\n",
			want:  "x := 1 <comment>// comment</comment>\n",
		},
		{
			input: "type Foo struct {}\n",
			want:  "type <defn>Foo</defn> struct {}\n",
		},
		{
			input: "func bar() {}\n",
			want:  "func <defn>bar</defn>() {}\n",
		},
		{
			input: "func (*Foo) moo() {}\n",
			want:  "func (*Foo) <defn>moo</defn>() {}\n",
		},
		{
			// Inline em markers (as produced by scanFile)
			input: "x := \x00em\x00foo\x00/em\x00()\n",
			want:  "x := <span class=\"em\">foo</span>()\n",
		},
		{
			input: "func (f Foo) moo() {}\n",
			want:  "func (f Foo) <defn>moo</defn>() {}\n",
		},
		{
			// Underscore suffix stripping
			input: "x := foo_3x(bar_v2)\n",
			want:  "x := foo(bar)\n",
		},
		{
			// Leading underscore preserved
			input: "_private := 1\n",
			want:  "_private := 1\n",
		},
		{
			// Underscore suffix on func def
			input: "func doThing_2() {}\n",
			want:  "func <defn>doThing</defn>() {}\n",
		},
	}
	for _, tt := range tests {
		got := renderCode(tt.input)
		if got != tt.want {
			t.Errorf("renderCode(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
