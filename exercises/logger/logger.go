// Exercise: find the race.

package lograce

import (
	"bytes"
	"fmt"
	"io"
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
	data = l.buf.Bytes()
	l.mu.Unlock()

	_, _ = l.out.Write(data)
}
