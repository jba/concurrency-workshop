#!/bin/bash -e

function build_slides {
	build_mutexes
	build_channels
	build_patterns
}

function build_mutexes {
	(set -x ; go run ./cmd/code2slides -o mutexes.slides slides/mutexes/10-mutexes.go)
}

function build_channels {
	 (set -x ; go run ./cmd/code2slides -o channels.slides slides/channels/channels.go)
}

function build_patterns {
	 (set -x ; go run ./cmd/code2slides -o patterns.slides slides/patterns/[0-9]*.go)
}

function test_solutions {
	for d in exercises/*; do
		go test ./$d/solution
	done
}

build_slides

test_solutions


