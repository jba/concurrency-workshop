package patterns

import (
	"context"
	"net/http"
	"sync"

	"golang.org/x/sync/errgroup"
)

// title ErrGroup: WaitGroup with Errors

// //////////////////////////////////
// heading Using WaitGroup with errors

// code
// checkURLs returns the first error found when getting the URLs.
func checkURLs(urls []string) error {
	var wg sync.WaitGroup
	for _, u := range urls {
		wg.Go(func() {
			resp, err := http.Get(u)
			if err != nil {
				// TODO: stop the other goroutines and return the error.
			}
			resp.Body.Close() // always close the body!
		})
	}
	wg.Wait()
	return nil
}

// !code

// question How do we implement the comment?
// answer
// - A channel of errors
// - A context for cancellation
// !question

////////////////////////////////////
// heading errgroup.Group

// text golang.org/x/sync/errgroup.Group

// text "errgroup.Group is related to sync.WaitGroup but adds handling of tasks returning errors."

// moo
// ```
// package errgroup

// type Group
// func WithContext(ctx context.Context) (*Group, context.Context)
// func (g *Group) Go(f func() error)
// func (g *Group) SetLimit(n int)
// func (g *Group) TryGo(f func() error) bool
// func (g *Group) Wait() error
// ```
// !moo

////////////////////////////////////
// heading Just collecting errors

// code
func checkURLs_1(urls []string) error {
	var eg errgroup.Group // em
	for _, u := range urls {
		eg.Go(func() error { // em error
			resp, err := http.Get(u)
			if err != nil {
				return err
				// TODO: stop the other goroutines.
			}
			resp.Body.Close() // always close the body!
			return nil
		})
	}
	return eg.Wait()
}

// !code

// //////////////////////////////////
// heading Errors and cancellation
// code
func checkURLs_2(ctx context.Context, urls []string) error {
	eg, ctx := errgroup.WithContext(ctx) // em
	for _, u := range urls {
		eg.Go(func() error {
			// em
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
			if err != nil {
				return err
			}
			resp, err := http.DefaultClient.Do(req)
			// em
			if err != nil {
				return err
			}
			resp.Body.Close() // always close the body!
			return nil
		})
	}
	return eg.Wait()
}

// !code
