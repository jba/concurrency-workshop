# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go concurrency workshop repository that teaches concurrent programming concepts through progressive exercises. The primary focus is implementing `sync.WaitGroup` from scratch to understand synchronization primitives.

## Build and Test Commands

```bash
# Build all packages
go build ./...

# Run tests (when available)
go test ./...

# Run a single package's tests
go test ./waitgroup/ex1

# Check for race conditions
go test -race ./...
```

## Architecture

The repository is organized as progressive exercises under topic directories:

- `waitgroup/ex1` through `waitgroup/ex4`: Progressive implementations of `sync.WaitGroup`, each demonstrating different synchronization approaches and their pitfalls

### WaitGroup Exercise Progression

1. **ex1**: Naive implementation with no synchronization (demonstrates the problem)
2. **ex2**: Uses `atomic.Int64` for the counter
3. **ex3**: Adds a TOCTOU (Time Of Check-Time Of Use) race condition example with atomics
4. **ex4**: Uses `sync.Mutex` for proper synchronization

Each exercise file contains embedded notes (in `// note ... // end note` blocks) explaining the concepts and problems with each approach.
