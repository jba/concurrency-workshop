// Exercise: find the race.

package lograce

import (
	"bytes"
	"fmt"
	"slices"
	"sync"
)

type Logger struct {
	emit func([]byte)
	mu   sync.Mutex
	buf  bytes.Buffer
}

// NewLogger constructs a Logger that calls emit with complete log lines to emit.
func NewLogger(emit func([]byte)) *Logger {
	return &Logger{emit: emit}
}

func (l *Logger) Logf(format string, args ...any) {
	var data []byte

	l.mu.Lock()
	l.buf.Reset()
	fmt.Fprintf(&l.buf, format, args...)
	data = slices.Clone(l.buf.Bytes())
	l.mu.Unlock()

	l.emit(data)
}
