package main

// heading An example: generating unique IDs

// div.flex
// html <div>
// code
import (
	"fmt"
	// em
	"sync"
	// !em
)

type IDGenerator struct {
	prefix string
	mu     sync.Mutex
	num    int
}

func NewIDGenerator(prefix string) *IDGenerator {
	return &IDGenerator{prefix: prefix}
}

func (g *IDGenerator) NewID() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.num++
	return fmt.Sprintf("%s%d", g.prefix, g.num)
}

// !code

// code
func use() {
	g := NewIDGenerator("gopher-")
	fmt.Println(g.NewID())
	fmt.Println(g.NewID())
}

// !code
/* output
TODO CHECK
gopher-1
gopher-2
*/
// html </div>

// text
// `num` must be synchronized to make `NewID` concurrency-safe
//
// `prefix` is written _before_ all calls to `NewID`
//
// The "mutex hat" convention: declare the mutex field
// above the fields it protects

// ![mutex hat](mutex-hat.png)

// !text
// !div.flex
