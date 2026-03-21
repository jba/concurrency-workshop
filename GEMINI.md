# GEMINI.md

This file provides instructional context for Gemini CLI when working in this repository.

## Project Overview

This is a **Go Concurrency Workshop** repository. It teaches concurrent programming concepts in Go through a series of progressive exercises, including synchronization primitives (Mutex, WaitGroup), common patterns (Hedging, Timeout), and best practices for writing thread-safe code.

The project also contains presentation slides and documentation for the workshop.

## Directory Structure

- `exercises/`: Contains the core workshop exercises.
    - `account/`: Exercise for making a struct thread-safe using `sync.Mutex`.
    - `waitgroup/`: Progressive exercise to implement `sync.WaitGroup` from scratch.
    - `hedging/`: Exercise for the hedging pattern (sending multiple requests and taking the first response).
    - `timeout/`: Exercise for implementing timeouts using `select` and `time.After`.
- `slides/`: Contains the source files for the workshop presentations.
- `docs/`: Holds documentation and generated HTML for the slides.
- `internal/`: Internal helpers used for testing (e.g., `testhelp`).
- `*.slides`: Raw slide files used for generating the workshop materials.
- `*.html`: Generated presentation slides.

## Building and Running

### Prerequisites

- Go (v1.26.0 or higher as per `go.mod`).
- Note: Some exercises use `testing/synctest`, which may require a specific Go version or experiment flag.

### Commands

- **Run all tests:**
  ```bash
  go test ./...
  ```
- **Run all tests with the race detector:**
  ```bash
  go test -race ./...
  ```
- **Run tests for a specific exercise:**
  ```bash
  go test ./exercises/account
  go test ./exercises/waitgroup/solution
  ```
- **Build all packages:**
  ```bash
  go build ./...
  ```

## Development Conventions

- **Progressive Exercises:** Many exercises have a `solution/` subdirectory. When working on an exercise, refer to the solution only if necessary.
- **Testing:** Always run tests with the `-race` flag when working on concurrency exercises to detect data races.
- **Idiomatic Go:** Follow standard Go formatting (`go fmt`) and naming conventions.
- **Concurrency Primitives:** Prioritize using standard library primitives like `sync.Mutex`, `sync.WaitGroup`, and channels. Some exercises involve implementing these from scratch to understand their internal mechanics.
- **Embedded Notes:** Look for `// note ... // end note` blocks in source files for explanations of concepts and common pitfalls.
