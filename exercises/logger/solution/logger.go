package lograce

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"sync"
)

type Logger struct {
	out io.Writer
	mu  sync.Mutex
	buf bytes.Buffer
}

func NewLogger(out io.Writer) *Logger {
	return &Logger{out: out}
}

func (l *Logger) Logf(format string, args ...any) {
	var data []byte

	l.mu.Lock()
	l.buf.Reset()
	fmt.Fprintf(&l.buf, format, args...)
	// Copy the shared byte slice inside the bytes.Buffer.
	data = slices.Clone(l.buf.Bytes())
	l.mu.Unlock()

	_, _ = l.out.Write(data)
}
