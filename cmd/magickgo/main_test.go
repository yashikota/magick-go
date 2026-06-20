package main

import (
	"reflect"
	"testing"
)

func TestNormalizeFlagOrder(t *testing.T) {
	got := normalizeFlagOrder([]string{"in.png", "out.webp", "--width", "1200", "--strip", "--quality=80"})
	want := []string{"--width", "1200", "--strip", "--quality=80", "in.png", "out.webp"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeFlagOrder() = %#v, want %#v", got, want)
	}
}

func TestNormalizeFlagOrderSingleDash(t *testing.T) {
	got := normalizeFlagOrder([]string{"in.png", "out.webp", "-width", "1200", "-strip", "-quality=80"})
	want := []string{"-width", "1200", "-strip", "-quality=80", "in.png", "out.webp"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeFlagOrder() = %#v, want %#v", got, want)
	}
}
