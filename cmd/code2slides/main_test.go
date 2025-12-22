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
		{kind: sectionCode, content: "func foo() {}\n"},
		{kind: sectionNote, content: "Second note.\n"},
		{kind: sectionNote, content: "Third note after blank comment.\n"},
		{kind: sectionNote, content: "Fourth note after blank line.\n"},
		{kind: sectionCode, content: "func bar() {}\n"},
		{kind: sectionQuestion, content: "What is the answer?\n"},
		{kind: sectionAnswer, content: "The answer is 42.\n"},
		{kind: sectionNote, content: "Use `fmt.Println` to print.\n"},
	}

	if !slices.Equal(slide.sections, wantSections) {
		t.Errorf("sections = %v, want %v", slide.sections, wantSections)
	}
}
