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

func TestPaint_RespectsNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if got := paint(ansiGreen, "hi"); got != "hi" {
		t.Errorf("NO_COLOR set should strip color, got %q", got)
	}
	t.Setenv("NO_COLOR", "")
	if got := paint(ansiGreen, "hi"); !strings.Contains(got, ansiGreen) || !strings.Contains(got, ansiReset) {
		t.Errorf("color should wrap when NO_COLOR is unset, got %q", got)
	}
}
