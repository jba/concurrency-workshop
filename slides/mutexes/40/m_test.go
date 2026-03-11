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

func newIDGenerator1(prefix string) *IDGenerator_1 {
	return &IDGenerator_1{prefix: prefix}
}

func Test1(t *testing.T) {
	g := newIDGenerator1("moo")
	got := g.NewID_2()
	want := "moo1"
	if got != want {
		t.Fatal("bad")
	}
	got = g.NewID_2()
	want = "moo2"
	if got != want {
		t.Fatal("bad")
	}
}

func newIDGenerator2(prefix string) *IDGenerator_2 {
	return &IDGenerator_2{prefix: prefix}
}

func Test2(t *testing.T) {
	g := newIDGenerator2("moo")
	got := g.NewID_3()
	want := "moo1"
	if got != want {
		t.Fatal("bad")
	}
	got = g.NewID_3()
	want = "moo2"
	if got != want {
		t.Fatal("bad")
	}
}
