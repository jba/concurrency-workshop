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

// question
// How do we implement the comment?
// answer
// - A channel of errors
// - A context for cancellation
// !question

////////////////////////////////////
// heading errgroup.Group

// text golang.org/x/sync/errgroup.Group

// text "errgroup.Group is related to sync.WaitGroup but adds handling of tasks returning errors."

// text
// ```
// package errgroup
//
// type Group
// func (g *Group) Go(f func() error)
// func (g *Group) Wait() error

// func WithContext(ctx context.Context) (*Group, context.Context)
// func (g *Group) SetLimit(n int)
// func (g *Group) TryGo(f func() error) bool
// ```
// !text

////////////////////////////////////
// heading A useful helper function

// text This will simplify later code.

// code
func getURLErr(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	resp.Body.Close() // always close the body!
	return nil
}

// !code

////////////////////////////////////
// heading Just collecting errors

// text `Wait` returns the first non-nil error.

// code
func checkURLs_1(urls []string) error {
	var eg errgroup.Group // em
	for _, u := range urls {
		eg.Go(func() error { // em error
			// TODO: stop the other goroutines.
			return getURLErr(u)
		})
	}
	return eg.Wait()
}

// !code

////////////////////////////////////
// heading getURLErr with a context

// code
func getURLErr_1(ctx context.Context, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	// Do returns immediately when req.Context is cancelled
	// or times out.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close() // always close the body!
	return nil
}

// !code

// ////////////////////////////////
// heading Errors and cancellation

// text The context is canceled on the first non-nil error.

// code
func checkURLs_2(ctx context.Context, urls []string) error {
	eg, ctx := errgroup.WithContext(ctx) // em
	for _, u := range urls {
		eg.Go(func() error {
			return getURLErr_1(ctx, u)
		})
	}
	return eg.Wait()
}

// !code

//////////////////////////////////
// heading Setting limits

// text `SetLimit` puts a cap on the number of active goroutines.

// text Useful even if you don't care about errors.

// code
func checkURLs_3(ctx context.Context, urls []string) error {
	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(4) // em
	for _, u := range urls {
		// Go blocks if there are more than 4 active goroutines.
		eg.Go(func() error {
			return getURLErr_1(ctx, u)
		})
	}
	return eg.Wait()
}

// !code
