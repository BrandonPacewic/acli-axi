package axi

import (
	"reflect"
	"testing"
)

func TestLimitHints(t *testing.T) {
	got := LimitHints(2, "first", "", "first", "second", "third")
	want := []string{"first", "second"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LimitHints() = %#v, want %#v", got, want)
	}
}

func TestLimitHintsUsesDefaultBound(t *testing.T) {
	got := LimitHints(0, "one", "two", "three", "four")
	if len(got) != DefaultHintLimit {
		t.Fatalf("LimitHints() returned %d hints, want %d", len(got), DefaultHintLimit)
	}
}
