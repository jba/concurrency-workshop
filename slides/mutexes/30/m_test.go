package m

import (
	"os"
	"strings"
	"testing"
)

func TestMutex(t *testing.T) {
	wantStdout(t, "40000", main)
}

func wantStdout(t *testing.T, want string, f func()) {
	t.Helper()
	got := stdout(f)
	if got != want {
		t.Errorf("\ngot  %s\nwant %s", got, want)
	}
}

func stdout(f func()) string {
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
