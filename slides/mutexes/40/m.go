package main

// heading Generating unique IDs, unsafe version

import (
	"fmt"
	"io"
	"iter"
	"reflect"
	"slices"
	"sync"
	"sync/atomic"
)

// cols
// code bad
// An IDGenerator generates unique identifiers.
type IDGenerator struct {
	prefix string
	num    int
}

// NewIDGenerator creates an IDGenerator whose
// identifiers begin with prefix.
func NewIDGenerator(prefix string) *IDGenerator {
	return &IDGenerator{prefix: prefix}
}

// NewID generates a unique identifier each time
// it is called.
func (g *IDGenerator) NewID() string {
	g.num++
	return fmt.Sprintf("%s%d", g.prefix, g.num)
}

// !code

// nextcol
// question
// What can happen if two goroutines call `g.NewID()` at the same time?
// answer
// You might get the same ID twice.
// !question
// !cols

//////////////////////////////////////////
// heading Generating unique IDs, safe version

// cols
// code
type IDGenerator_1 struct {
	prefix string
	mu     sync.Mutex // em
	num    int
}

// NewIDGenerator is the same.

func (g *IDGenerator_1) NewID_1() string {
	// em
	g.mu.Lock()
	defer g.mu.Unlock()
	// !em
	g.num++
	return fmt.Sprintf("%s%d", g.prefix, g.num)
}

// !code
// nextcol
// text
// `num` must be synchronized to make `NewID` concurrency-safe
//
// `prefix` is written _before_ all calls to `NewID`
//
// The "mutex hat" convention: declare the mutex field
// above the fields it protects

// !text
// !cols

//////////////////////////////////////////
// heading Limit critical section size

// cols
// code
func (g *IDGenerator_1) NewID_2() string {
	g.mu.Lock()
	g.num++
	g.mu.Unlock()
	return fmt.Sprintf("%s%d", g.prefix, g.num)
}

// !code

// nextcol

// text
// Keep locked regions (_critical sections_) small to avoid contention
//
// `defer` is not always needed or useful
// !text
// question
// Find and fix the bug.
// answer
// code
func (g *IDGenerator_1) NewID_3() string {
	g.mu.Lock()
	g.num++
	n := g.num // em
	g.mu.Unlock()
	return fmt.Sprintf("%s%d", g.prefix, n) // em \bn\b
}

// !code

// !question

// !cols

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

func flog() {
	var l struct {
		outMu sync.Mutex
		out   io.Writer
	}
	var buf *[]byte
	// code
	l.outMu.Lock()
	defer l.outMu.Unlock()
	_, err := l.out.Write(*buf) // avoid interleaved log lines
	// !code
	_ = err
}

// text From The [log package](https://github.com/golang/go/blob/e30e65f7a8bda0351d9def5a6bc91471bddafd3d/src/log/log.go)

//////////////////////////////////////////
// heading Another example of copying

// text We don't know how long the caller will hold on to the iterator.

// code
// Sessions returns an iterator that yields a snapshot of the server sessions.
func (s *Server) Sessions() iter.Seq[*ServerSession] {
	s.mu.Lock()
	clients := slices.Clone(s.sessions)
	s.mu.Unlock()
	return slices.Values(clients)
}

// !code

// text From The [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk/blob/4cdbaaf27132e5356ba13973ae50da4edfa876bb/mcp/server.go)

//////////////////////////////////////////
// heading Atomics

// cols
// text
// Package `sync/atomic` exposes CPU atomic operations
//
// "These functions require great care to be used correctly."
//
// - Faster than mutexes, but much more dangerous.
// - Limited operations
// - Sequences of atomics are _not_ atomic
//
// Recommendation: use only for counters
// !text

// nextcol

// code
type IDGenerator_2 struct {
	prefix string
	num    atomic.Int64 // em
}

// NewIDGenerator is the same.

func (g *IDGenerator_2) NewID_3() string {
	n := g.num.Add(1) // em
	return fmt.Sprintf("%s%d", g.prefix, n)
}

// !code
// !cols

//////////////////////////////////////////
// heading Mutexes and slices

// text Each slice element is a separate memory location.
// text No mutex needed here.

func fslice1() []int {
	// code
	var wg sync.WaitGroup
	s := make([]int, 2)
	wg.Go(func() { s[0] = 1 })
	wg.Go(func() { s[1] = 2 })
	wg.Wait()
	// !code
	return s
}

//////////////////////////////////////////
// heading Mutexes and slices, 2

func fslice2() []int {
	// code
	var wg sync.WaitGroup
	var s []int
	wg.Go(func() { s = append(s, 1) }) // em
	wg.Go(func() { s = append(s, 2) }) // em
	wg.Wait()
	// !code
	return s
}

// question Is a mutex needed here?
// answer
// Yes: there is a data race.
// Both goroutines write to the same location, `s`.
// !question

//////////////////////////////////////////
// heading Mutexes and maps

// cols
// code bad
// IDGenerator generates unique IDs with different prefixes.
type IDGenerator_m1 struct {
	nums map[string]int // prefix to next ID // em
}

func NewIDGenerator_m1(prefix string) *IDGenerator_m1 {
	return &IDGenerator_m1{nums: map[string]int{}}
}

func (g *IDGenerator_m1) NewID_m1(prefix string) string { // em prefix string
	n := g.nums[prefix]
	n++
	g.nums[prefix] = n
	return fmt.Sprintf("%s%d", prefix, n)
}

// !code

// nextcol

// text Maps need synchronization.

// question Why aren't maps like slices?
// answer
// The underlying memory locations can change as the map grows and shrinks.
// !question

// !cols

//////////////////////////////////////////
// heading Mutexes and maps, safely

// code
type IDGenerator_m2 struct {
	mu   sync.Mutex // em
	nums map[string]int
}

// NewIDGenerator is the same.

func (g *IDGenerator_m2) NewID_m2(prefix string) string {
	g.mu.Lock() // em
	n := g.nums[prefix]
	n++
	g.nums[prefix] = n
	g.mu.Unlock() // em
	return fmt.Sprintf("%s%d", prefix, n)
}

// !code

////////////////////////////////
// heading Optimizations

// text
// Atomics, as we've seen.
//
// `sync.RWMutex` if there are many more reads than writes.
//
// `sync.Map` if there are very few writes (often just one).

// !text

////////////////////////////////
// heading sync.Map example

// cols

// code
type userTypeInfo struct{} // fields omitted

var userTypeCache sync.Map // map[reflect.Type]*userTypeInfo // em sync.Map

func validUserType(rt reflect.Type) (*userTypeInfo, error) {
	if ui, ok := userTypeCache.Load(rt); ok {
		return ui.(*userTypeInfo), nil
	}

	// Construct a new userTypeInfo and atomically
	// add it to the userTypeCache.
	ut := new(userTypeInfo)
	// ...

	ui, _ := userTypeCache.LoadOrStore(rt, ut)
	return ui.(*userTypeInfo), nil
}

// !code

// nextcol

// text &nbsp;
// text From the [encoding/gob](https://github.com/golang/go/blob/master/src/encoding/gob/type.go) package.
// question What about the race?
// answer
// "If we lose the race, we'll waste a little CPU and create a little garbage
// but return the existing value anyway."
// !question
// !cols

////////////////////////////////
// heading Find the bug
// TODO find example that Mac found in mcp or jsonschema-go

////////////////////////////////
// heading Exercise
