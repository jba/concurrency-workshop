package m

import (
	"fmt"
	"io"
	"slices"
	"sync"
)

// title Advanced Mutex Topics

//////////////////////////////////////////
// heading Copy defensively
//
// text Copy private mutable data

// text `Server.sessions` is managed by `Server`
// text It doesn't want anyone else to change it.
//
// code
// Sessions returns a copy of the server's sessions.
func (s *Server) Sessions() []*ServerSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.sessions)
}

// !code

// text From The [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk/blob/4cdbaaf27132e5356ba13973ae50da4edfa876bb/mcp/server.go)

//////////////////////////////////////////
// heading Avoid locking during I/O

// text I/O is slow.
// text Network peers can be _arbitrarily_ slow.
// text Sometimes you have to copy.

// code
func (s *Server) notifySessions(n string) {
	s.mu.Lock() // required to access s.sessions
	sessions := slices.Clone(s.sessions)
	s.pendingNotifications[n] = nil
	s.mu.Unlock()
	// Do I/O with no locks held.
	notifySessions(sessions, n, changeNotificationParams[n], s.opts.Logger)
}

// !code

// text From The [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk/blob/4cdbaaf27132e5356ba13973ae50da4edfa876bb/mcp/server.go)

type Server struct {
	mu                   sync.Mutex
	sessions             []*ServerSession
	pendingNotifications map[string]*int
	opts                 struct{ Logger int }
}

type ServerSession int

var changeNotificationParams map[string]int

func notifySessions([]*ServerSession, string, int, int) {}

//////////////////////////////////////////
// heading Avoid locking during I/O...unless you need it

// code
type Logger struct {
	outMu sync.Mutex
	out   io.Writer
	// ...
}

// Printf calls l.Output to print to the logger.
func (l *Logger) Printf(format string, v ...any) {
	l.output(0, 2, func(b []byte) []byte {
		return fmt.Appendf(b, format, v...)
	})
}

// output can take either a calldepth or a pc to get source line information.
// It uses the pc if it is non-zero.
func (l *Logger) output(pc uintptr, calldepth int, appendOutput func([]byte) []byte) error {
	var buf []byte
	// ...
	buf = appendOutput(buf)
	// ...
	// em
	l.outMu.Lock()
	defer l.outMu.Unlock()
	_, err := l.out.Write(buf)
	// !em
	return err
}

// !code

// text From The [log package](https://github.com/golang/go/blob/e30e65f7a8bda0351d9def5a6bc91471bddafd3d/src/log/log.go)
