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

// cols
// code
// checkURLs returns the first error found when getting the URLs.
func checkURLs(urls []string) error {
	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Go(func() {
			resp, err := http.Get(url)
			if err != nil {
				// TODO: stop the other goroutines
				// and record the error.
			}
			resp.Body.Close() // always close the body!
		})
	}
	wg.Wait()
	return nil // TODO: return the first error
}

// !code
// nextcol
// question
// html <br/>
// How do we implement the comment?
// answer
// - A channel of errors
// - A context for cancellation
// !question
// !cols
//
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
// func (g *Group) Wait() error  // wait for all, return first error

// func WithContext(ctx context.Context) (*Group, context.Context)
// func (g *Group) SetLimit(n int)
// func (g *Group) TryGo(f func() error) bool
// ```
// !text

////////////////////////////////////
// heading Just collecting errors

// text `Wait` returns the first non-nil error.

// code
func checkURLs_1(urls []string) error {
	var eg errgroup.Group // em
	for _, url := range urls {
		eg.Go(func() error { // em error
			// TODO: stop the other goroutines.
			resp, err := http.Get(url)
			if err != nil {
				return err
			}
			resp.Body.Close()
			return nil
		})
	}
	return eg.Wait() // em eg.Wait
}

// !code

////////////////////////////////////
// heading A useful helper

// text Making a GET request with cancellation.
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
	resp.Body.Close()
	return nil
}

// !code

// ////////////////////////////////
// heading Errors and cancellation

// text The context is canceled on the first non-nil error.

// code
func checkURLs_2(ctx context.Context, urls []string) error {
	eg, gctx := errgroup.WithContext(ctx) // em
	for _, url := range urls {
		eg.Go(func() error {
			return getURLErr_1(gctx, url)
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
	for _, url := range urls {
		// Go blocks if there are more than 4 active goroutines.
		eg.Go(func() error {
			return getURLErr_1(ctx, url)
		})
	}
	return eg.Wait()
}

// !code

// heading A real-world example
//
// text Making multiple, independent RPCs to a service concurrently.
//
// text From https://github.com/golang/pkgsite/blob/master/cmd/internal/pkgsite-cli/module.go
//
// text
// `> pkgsite-cli module -packages -versions github.com/modelcontextprotocol/go-sdk`
// !text
// output
//   Version:          v1.6.1 (latest)
//   Repository:       https://github.com/modelcontextprotocol/go-sdk
//   Has go.mod:       yes
//   Redistributable:  yes

// Versions:
//   v1.6.1
//   v1.6.0
//   ...
//
// Packages:
//   github.com/modelcontextprotocol/go-sdk/auth
//   github.com/modelcontextprotocol/go-sdk/auth/extauth
//   ...
// !output

// //////////////////////////////////
// heading Sequential
// cols
func b(ctx context.Context, path, version string) error {
	// code
	c, err := client.New(server)
	// ...
	mod, err := c.GetModule(ctx, path, version)
	// ...
	result := moduleResult{Module: mod}
	// !code
	// nextcol
	// code small
	if m.versions {
		result.Versions, err = c.GetVersions(ctx, path)
		if err != nil {
			return err
		}
	}
	if m.vulns {
		result.Vulns, err = c.GetVulns(ctx, path, mod.Version)
		if err != nil {
			return err
		}
	}
	if m.packages {
		result.Packages, err = c.GetPackages(ctx, path, mod.Version)
		if err != nil {
			return err
		}
	}
	return write(result)
	// !code
}

// !cols

// //////////////////////////////////
// heading Concurrent

// cols
func a(ctx context.Context, path, version string) error {
	// code
	c, err := client.New(server)
	// elide
	if err != nil {
		return err
	}
	// !elide
	mod, err := c.GetModule(ctx, path, version)
	// ...
	result := moduleResult{Module: mod}

	g, gctx := errgroup.WithContext(ctx) // em
	// !code
	// nextcol
	// code small
	if m.versions {
		g.Go(func() (err error) { // em g.Go
			result.Versions, err = c.GetVersions(gctx, path) // em gctx
			return err
		})
	}
	if m.vulns {
		g.Go(func() (err error) { // em g.Go
			result.Vulns, err = c.GetVulns(gctx, path, mod.Version) // em gctx
			return err
		})
	}
	if m.packages {
		g.Go(func() (err error) { // em g.Go
			result.Packages, err = c.GetPackages(gctx, path, mod.Version) // em gctx
			return err
		})
	}

	if err := g.Wait(); err != nil { // em
		return err
	}

	return write(result)
	// !code
}

// !cols

var m struct {
	versions, vulns, packages bool
}

type cli struct{}

func (cli) New(int) (cli, error)                                     { return cli{}, nil }
func (cli) GetModule(context.Context, string, string) (mod, error)   { return mod{}, nil }
func (cli) GetVersions(context.Context, string) (int, error)         { return 0, nil }
func (cli) GetPackages(context.Context, string, string) (int, error) { return 0, nil }
func (cli) GetVulns(context.Context, string, string) (int, error)    { return 0, nil }

var server, path int

var client cli

type mod struct{ Version string }
type moduleResult struct {
	Module                    mod
	Versions, Vulns, Packages int
}

func write(moduleResult) error { return nil }
