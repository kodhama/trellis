package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestE2E_M1_ApplyToSampleProject drives the whole CLI (detect → plan → apply)
// against a sample project on disk, then checks the real overlay it wrote. This is
// the deterministic e2e — no model, always runs.
func TestE2E_M1_ApplyToSampleProject(t *testing.T) {
	claudePresent(t)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Sample\n\nExisting rules.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := run2("", "setup", "--dir", dir, "--profile", "seed", "--mode", "m1", "--apply")
	if err != nil {
		t.Fatalf("e2e setup --apply: %v", err)
	}
	if !strings.Contains(out, "applied (M1 overlay)") {
		t.Errorf("expected an applied summary, got:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, ".trellis", "profile.md")); err != nil {
		t.Errorf(".trellis/profile.md not created: %v", err)
	}
	c := readFile(t, filepath.Join(dir, "CLAUDE.md"))
	if !strings.Contains(c, "Existing rules.") {
		t.Error("e2e clobbered the sample's CLAUDE.md")
	}
	if !strings.Contains(c, trellisBegin) || !strings.Contains(c, "without its human approval") {
		t.Error("e2e did not install the trellis block + the B2 surfacing behavior")
	}
}

// TestE2E_M2_ApplyNotYetImplemented pins the honest stub: M2 apply must refuse
// loudly until its own slice, not pretend it did anything.
func TestE2E_M2_ApplyNotYetImplemented(t *testing.T) {
	claudePresent(t)
	dir := t.TempDir()
	_, err := run2("", "setup", "--dir", dir, "--profile", "a", "--mode", "m2", "--model", "high", "--apply")
	if err == nil {
		t.Fatal("M2 --apply should error until implemented")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("expected a 'not implemented' error, got: %v", err)
	}
}

// TestE2E_DryRunWritesNothing confirms the default is a no-write dry run.
func TestE2E_DryRunWritesNothing(t *testing.T) {
	claudePresent(t)
	dir := t.TempDir()
	out, err := run2("", "setup", "--dir", dir, "--profile", "seed", "--mode", "m1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "dry run") {
		t.Errorf("expected a dry-run notice, got:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, ".trellis")); !os.IsNotExist(err) {
		t.Error("dry run should not write .trellis/")
	}
}
