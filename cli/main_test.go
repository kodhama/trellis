package main

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	var buf bytes.Buffer
	if err := run(&buf, []string{"version"}); err != nil {
		t.Fatalf("version returned error: %v", err)
	}
	if got := buf.String(); !strings.Contains(got, "trellis ") {
		t.Errorf("version output = %q, want it to contain %q", got, "trellis ")
	}
}

func TestRunHelpAndNoArgs(t *testing.T) {
	for _, args := range [][]string{nil, {"help"}} {
		var buf bytes.Buffer
		if err := run(&buf, args); err != nil {
			t.Fatalf("run(%v) returned error: %v", args, err)
		}
		if !strings.Contains(buf.String(), "trellis setup") {
			t.Errorf("run(%v) usage did not mention the setup command: %q", args, buf.String())
		}
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var buf bytes.Buffer
	if err := run(&buf, []string{"nope"}); err == nil {
		t.Fatal("expected an error for an unknown command, got nil")
	}
}

func TestSetupNoHarness(t *testing.T) {
	// No harness executable: setup must fail loudly rather than guess (D1).
	withLookPath(t, func(string) (string, error) { return "", exec.ErrNotFound })
	var buf bytes.Buffer
	if err := run(&buf, []string{"setup", "--dir", t.TempDir()}); err == nil {
		t.Fatal("setup should error when no harness executable is present")
	}
}

func TestSetupDetectsHarness(t *testing.T) {
	withLookPath(t, func(string) (string, error) { return "/usr/local/bin/claude", nil })
	var buf bytes.Buffer
	if err := run(&buf, []string{"setup", "--dir", t.TempDir()}); err != nil {
		t.Fatalf("setup should succeed when claude is present, got %v", err)
	}
	if !strings.Contains(buf.String(), "detected harness") {
		t.Errorf("setup output should report the detected harness, got %q", buf.String())
	}
}
