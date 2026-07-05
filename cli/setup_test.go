package main

import (
	"os/exec"
	"strings"
	"testing"
)

func claudePresent(t *testing.T) {
	t.Helper()
	withLookPath(t, func(string) (string, error) { return "/usr/local/bin/claude", nil })
}

func TestSetupNoHarnessM1Works(t *testing.T) {
	// M1 is a deterministic file overlay — it must NOT require the `claude` binary
	// (decision-0029). The friction fix: you can overlay without installing the harness.
	withLookPath(t, func(string) (string, error) { return "", exec.ErrNotFound })
	out, err := run2("", "setup", "--dir", t.TempDir(), "--mode", "m1")
	if err != nil {
		t.Fatalf("m1 setup should work with no harness binary present: %v", err)
	}
	if !strings.Contains(out, "alongside") {
		t.Errorf("expected an m1 plan (M1 must work with no harness binary), got:\n%s", out)
	}
}

func TestSetupNoHarnessM2Errors(t *testing.T) {
	// M2 drives the harness to rewrite — with no `claude` on PATH it must fail loudly (D1).
	withLookPath(t, func(string) (string, error) { return "", exec.ErrNotFound })
	if _, err := run2("", "setup", "--dir", t.TempDir(), "--mode", "m2", "--model", "high"); err == nil {
		t.Fatal("m2 setup should error when no harness executable is present")
	}
}

func TestSetupWithFlags(t *testing.T) {
	claudePresent(t)
	out, err := run2("", "setup", "--dir", t.TempDir(),
		"--profile", "a", "--mode", "m2", "--model", "high")
	if err != nil {
		t.Fatalf("setup with flags: %v", err)
	}
	for _, want := range []string{"detected harness", "conductor", "rewrite", "high-reasoning", "setup plan"} {
		if !strings.Contains(out, want) {
			t.Errorf("plan missing %q in:\n%s", want, out)
		}
	}
}

func TestSetupDefaultsOnEmptyInput(t *testing.T) {
	claudePresent(t)
	// No flags, empty stdin -> each prompt takes its default (b / m1 / none).
	out, err := run2("", "setup", "--dir", t.TempDir())
	if err != nil {
		t.Fatalf("setup with defaults: %v", err)
	}
	for _, want := range []string{"author-adapt", "alongside", "no model"} {
		if !strings.Contains(out, want) {
			t.Errorf("defaulted plan missing %q in:\n%s", want, out)
		}
	}
}

func TestSetupInteractive(t *testing.T) {
	claudePresent(t)
	// Answer the prompts in order: mode, profile, model (decision-0029: mode first).
	out, err := run2("m2\nb\nbalanced\n", "setup", "--dir", t.TempDir())
	if err != nil {
		t.Fatalf("interactive setup: %v", err)
	}
	for _, want := range []string{"author-adapt", "rewrite", "balanced"} {
		if !strings.Contains(out, want) {
			t.Errorf("interactive plan missing %q in:\n%s", want, out)
		}
	}
}

func TestSetupM2DefaultsHigh(t *testing.T) {
	claudePresent(t)
	// M2 with empty model input -> defaults to the high-reasoning model. Order: mode, profile, model.
	out, err := run2("m2\na\n\n", "setup", "--dir", t.TempDir())
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if !strings.Contains(out, "high-reasoning") {
		t.Errorf("empty model input on m2 should default to high-reasoning, got:\n%s", out)
	}
}

func TestSetupM2RejectsNone(t *testing.T) {
	claudePresent(t)
	// morph has no deterministic path -> --model none must be a loud error.
	if _, err := run2("", "setup", "--dir", t.TempDir(), "--mode", "m2", "--model", "none"); err == nil {
		t.Fatal("mode m2 with --model none should be an error")
	}
}

func TestSetupM1IsDeterministic(t *testing.T) {
	claudePresent(t)
	// M1 asks no model question and is forced to 'none'.
	out, err := run2("", "setup", "--dir", t.TempDir(), "--mode", "m1")
	if err != nil {
		t.Fatalf("m1 setup: %v", err)
	}
	if !strings.Contains(out, "deterministic overlay") || !strings.Contains(out, "no model") {
		t.Errorf("m1 should report a deterministic, model-less overlay, got:\n%s", out)
	}
	// A real model on m1 is a loud error, not a silently-ignored choice.
	if _, err := run2("", "setup", "--dir", t.TempDir(), "--mode", "m1", "--model", "high"); err == nil {
		t.Fatal("mode m1 with --model high should be an error")
	}
}

func TestSetupInvalidFlag(t *testing.T) {
	claudePresent(t)
	if _, err := run2("", "setup", "--dir", t.TempDir(), "--profile", "zzz"); err == nil {
		t.Fatal("an invalid --profile should be a loud error")
	}
}
