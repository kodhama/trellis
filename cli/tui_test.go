package main

import (
	"bytes"
	"strings"
	"testing"
)

// The selector must engage only on a real terminal. Tests and pipes pass non-*os.File
// readers/writers, so ttyPair returns false and the deterministic line path stays in
// force — this is what keeps the whole suite unchanged (decision-0030).
func TestTTYPair_NonFileIsNotInteractive(t *testing.T) {
	if _, _, ok := ttyPair(strings.NewReader("x"), &bytes.Buffer{}); ok {
		t.Error("a non-file reader/writer must not be treated as a TTY")
	}
}

func TestPalette_RespectsNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if p := newPalette(); p.g("hi") != "hi" {
		t.Errorf("NO_COLOR set should strip color, got %q", p.g("hi"))
	}
	t.Setenv("NO_COLOR", "")
	if p := newPalette(); p.g("hi") == "hi" || !strings.Contains(p.g("hi"), "hi") {
		t.Errorf("color should wrap the text when NO_COLOR is unset, got %q", p.g("hi"))
	}
}
