// Package testhelp provides test utilities for capturing stdout.
package testhelp

import (
	"os"
	"strings"
	"testing"
)

// WantStdout runs f and checks that its stdout output matches want.
func WantStdout(t *testing.T, want string, f func()) {
	t.Helper()
	got := Stdout(f)
	if got != want {
		t.Errorf("\ngot  %s\nwant %s", got, want)
	}
}

// Stdout runs f and returns its stdout output with leading/trailing whitespace trimmed.
func Stdout(f func()) string {
	defer func(out *os.File) { os.Stdout = out }(os.Stdout)
	file, err := os.CreateTemp("", "stdout")
	if err != nil {
		panic(err)
	}
	defer os.Remove(file.Name())
	os.Stdout = file
	f()
	if err := file.Close(); err != nil {
		panic(err)
	}
	data, err := os.ReadFile(file.Name())
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(data))
}
