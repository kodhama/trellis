package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func planFor(profileKey string) Plan {
	p, _ := profileByKey(profileKey)
	m, _ := modeByKey("m1")
	mdl, _ := modelByKey("none")
	return Plan{Harness: Harness{Name: "Claude Code"}, Profile: p, Mode: m, Model: mdl}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestApplyM1WritesOverlay(t *testing.T) {
	dir := t.TempDir()
	if _, err := applyM1(dir, planFor("seed")); err != nil {
		t.Fatalf("applyM1: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".trellis", "profile.md")); err != nil {
		t.Errorf(".trellis/profile.md not written: %v", err)
	}
	// CLAUDE.md: minimal — a human line + the header import, and no governance
	// content duplicated into the host's file.
	c := readFile(t, filepath.Join(dir, "CLAUDE.md"))
	for _, want := range []string{trellisBegin, trellisEnd, "@.trellis/trellis.md", "governed by **Trellis**"} {
		if !strings.Contains(c, want) {
			t.Errorf("CLAUDE.md missing %q", want)
		}
	}
	if strings.Contains(c, "without its human approval") {
		t.Error("governance content should live in .trellis/, not in CLAUDE.md")
	}
	// The header carries the behavior and imports the sibling profile.
	header := readFile(t, filepath.Join(dir, ".trellis", "trellis.md"))
	for _, want := range []string{"without its human approval", "@profile.md"} {
		if !strings.Contains(header, want) {
			t.Errorf(".trellis/trellis.md missing %q", want)
		}
	}
	// The profile carries the active set; the bundled reference carries the catalog.
	if prof := readFile(t, filepath.Join(dir, ".trellis", "profile.md")); !strings.Contains(prof, "inv-directional-flow") {
		t.Errorf(".trellis/profile.md missing the active invariants: %q", prof)
	}
	if inv := readFile(t, filepath.Join(dir, ".trellis", "invariants.md")); !strings.Contains(inv, "inv-directional-flow") {
		t.Error(".trellis/invariants.md should contain the bundled invariant reference")
	}
}

func TestApplyM1AugmentNotClobber(t *testing.T) {
	dir := t.TempDir()
	original := "# My Project\n\nSome existing house rules.\n"
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := applyM1(dir, planFor("a")); err != nil {
		t.Fatal(err)
	}
	c := readFile(t, filepath.Join(dir, "CLAUDE.md"))
	if !strings.Contains(c, "Some existing house rules.") {
		t.Error("existing content was clobbered")
	}
	if !strings.Contains(c, trellisBegin) {
		t.Error("trellis block not appended")
	}
}

func TestApplyM1Idempotent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# My Project\n\nHouse rules.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := applyM1(dir, planFor("a")); err != nil {
		t.Fatal(err)
	}
	if _, err := applyM1(dir, planFor("a")); err != nil {
		t.Fatal(err)
	}
	c := readFile(t, filepath.Join(dir, "CLAUDE.md"))
	if n := strings.Count(c, trellisBegin); n != 1 {
		t.Errorf("expected exactly one trellis block after re-run, got %d", n)
	}
	if !strings.Contains(c, "House rules.") {
		t.Error("re-run clobbered existing content")
	}
}

func TestUpsertBlockReplaces(t *testing.T) {
	content := "top\n\n" + trellisBegin + "\nOLD\n" + trellisEnd + "\n\nbottom\n"
	out := upsertBlock(content, trellisBegin+"\nNEW\n"+trellisEnd)
	if strings.Contains(out, "OLD") {
		t.Error("old block not replaced")
	}
	for _, want := range []string{"NEW", "top", "bottom"} {
		if !strings.Contains(out, want) {
			t.Errorf("upsert lost %q: %q", want, out)
		}
	}
	if n := strings.Count(out, trellisBegin); n != 1 {
		t.Errorf("expected one block, got %d", n)
	}
}
