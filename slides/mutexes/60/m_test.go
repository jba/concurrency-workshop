package main

import "testing"

func Test(t *testing.T) {
	g := NewIDGenerator("moo")
	got := g.NewID()
	want := "moo1"
	if got != want {
		t.Fatal("bad")
	}
	got = g.NewID()
	want = "moo2"
	if got != want {
		t.Fatal("bad")
	}
}
