package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeOverlayVersion(t *testing.T, dir, v string) {
	t.Helper()
	td := filepath.Join(dir, ".trellis")
	if err := os.MkdirAll(td, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(td, "version"), []byte(v+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestStatusNoOverlay(t *testing.T) {
	out, err := run2("", "status", "--dir", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "no Trellis overlay") {
		t.Errorf("expected a no-overlay message, got: %s", out)
	}
}

func TestStatusCurrent(t *testing.T) {
	dir := t.TempDir()
	writeOverlayVersion(t, dir, version) // same as this binary
	out, err := run2("", "status", "--dir", dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "current") {
		t.Errorf("a matching version should report current, got: %s", out)
	}
}

func TestStatusStale(t *testing.T) {
	dir := t.TempDir()
	writeOverlayVersion(t, dir, "9.9.9-old") // differs from this binary
	out, err := run2("", "status", "--dir", dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "re-run") {
		t.Errorf("a differing version should suggest re-run, got: %s", out)
	}
}

func TestSetupStampsVersion(t *testing.T) {
	dir := t.TempDir()
	if _, err := applyM1(dir, planFor("b")); err != nil {
		t.Fatal(err)
	}
	if v := readFile(t, filepath.Join(dir, ".trellis", "version")); strings.TrimSpace(v) != version {
		t.Errorf(".trellis/version should stamp the binary version %q, got %q", version, v)
	}
}
