// serve serves the .slides files in a directory over HTTP.
//
// Usage:
//
//	serve [flags] DIR
//
// The homepage ("/") lists links to every .slides file in DIR. Each file is
// served at "/NAME.slides". Anything under the static directory (see -static)
// is served at "/static/..." with the proper content type.
package main

import (
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	addr      = flag.String("addr", ":8000", "address to listen on")
	staticDir = flag.String("static", "static", "directory of static assets served at /static/")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: serve [flags] DIR\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	dir := flag.Arg(0)
	if err := run(dir); err != nil {
		log.Fatal(err)
	}
}

func run(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	mux := http.NewServeMux()

	// Homepage: list of links to the .slides files.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		files, err := slidesFiles(dir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<!DOCTYPE html>\n<html>\n<head><title>Slides</title></head>\n<body>\n<h1>Slides</h1>\n<ul>\n")
		for _, f := range files {
			name := html.EscapeString(f)
			fmt.Fprintf(w, "<li><a href=%q>%s</a></li>\n", "/"+name, name)
		}
		fmt.Fprintf(w, "</ul>\n</body>\n</html>\n")
	})

	// Serve individual .slides files as HTML.
	mux.HandleFunc("/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if !strings.HasSuffix(name, ".slides") || strings.Contains(name, "/") {
			http.NotFound(w, r)
			return
		}
		p := filepath.Join(dir, name)
		if !fileExists(p) {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, p)
	})

	// Serve static assets with proper content types (handled by ServeFileFS).
	mux.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir(*staticDir))))

	log.Printf("serving %s on %s", dir, *addr)
	return http.ListenAndServe(*addr, mux)
}

// slidesFiles returns the names of the .slides files in dir, sorted.
func slidesFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".slides") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}
