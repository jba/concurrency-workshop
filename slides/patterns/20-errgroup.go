package patterns

import (
	"hash/fnv"
	"io"
	"os"
	"sync"
)

// title A ErrGroup: WaitGroup with Errors

// //////////////////////////////////
// heading Using WaitGroup with errors

// code
// hashFile computes the FNV64a hash of the contents
// of filename.
func hashFile(filename string) (uint64, error) {
	f, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	h := fnv.New64a()
	if _, err := io.Copy(h, f); err != nil {
		return 0, err
	}
	return h.Sum64(), nil
}

// !code

// code
// hashFiles returns the FNV64a hashes of the given files.
func hashFiles(filenames []string) ([]uint64, error) {
	var wg sync.WaitGroup
	result := make([]uint64, len(filenames))
	for i, f := range filenames {
		wg.Go(func() {
			h, err := hashFile(f)
			if err != nil {
				// Stop the other goroutines and return the error.
			}
			result[i] = h
		})
	}
	return result, nil
}

// !code

// question How do we implement the comment?
// answer
// - A channel of `struct { h uint64; err error }`
// - A context for cancellation
// !question

////////////////////////////////////
// heading errgroup.Group

// text golang.org/x/sync/errgroup.Group

// text "errgroup.Group is related to sync.WaitGroup but adds handling of tasks returning errors."

// text
// `package errgroup`

// `type Group`
// `func WithContext(ctx context.Context) (*Group, context.Context)``
// `func (g *Group) Go(f func() error)``
// `func (g *Group) SetLimit(n int)``
// `func (g *Group) TryGo(f func() error) bool`
// `func (g *Group) Wait() error`
// !text

////////////////////////////////////
// heading Just collecting errors

////////////////////////////////////
// heading

////////////////////////////////////
// heading

////////////////////////////////////
// heading

////////////////////////////////////
// heading

////////////////////////////////////
// heading

////////////////////////////////////
// heading

////////////////////////////////////
// heading
