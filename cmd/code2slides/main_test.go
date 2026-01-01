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
	slide, err := scanFile("testdata/valid.go")
	if err != nil {
		t.Fatal(err)
	}

	if slide.heading != "Test Heading" {
		t.Errorf("heading = %q, want %q", slide.heading, "Test Heading")
	}

	wantSections := []section{
		{kind: sectionNote, content: "First note.\n"},
		{kind: sectionCode, content: "func foo() {}"},
		{kind: sectionNote, content: "Second note.\n"},
		{kind: sectionNote, content: "Third note after blank comment.\n"},
		{kind: sectionNote, content: "Fourth note after blank line.\n"},
		{kind: sectionCode, content: "func bar() {}"},
		{kind: sectionQuestion, content: "What is the answer?\n"},
		{kind: sectionAnswer, content: "The answer is 42.\n"},
		{kind: sectionNote, content: "Use `fmt.Println` to print.\n"},
	}

	if !slices.Equal(slide.sections, wantSections) {
		t.Errorf("sections = %v, want %v", slide.sections, wantSections)
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
		input     string
		wantWord  string
		wantRest  string
		wantOK    bool
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
			input: "func (f Foo) moo() {}\n",
			want:  "func (f Foo) <defn>moo</defn>() {}\n",
		},
	}
	for _, tt := range tests {
		got := renderCode(tt.input)
		if got != tt.want {
			t.Errorf("renderCode(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
