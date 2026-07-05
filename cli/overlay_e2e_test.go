package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The M1 overlay is deterministic, so we can drive every target/detection path and
// assert the exact files it writes (decision-0029 / follow-up B). Each case runs with
// NO `claude` on PATH, proving the overlay needs no harness binary.

func noClaude(t *testing.T) {
	t.Helper()
	withLookPath(t, func(string) (string, error) { return "", exec.ErrNotFound })
}

func seedFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustApplyM1(t *testing.T, dir, target string) string {
	t.Helper()
	out, err := run2("", "setup", "--dir", dir, "--profile", "seed", "--mode", "m1", "--target", target, "--apply")
	if err != nil {
		t.Fatalf("m1 apply --target %s: %v", target, err)
	}
	return out
}

func mustHave(t *testing.T, s string, subs ...string) {
	t.Helper()
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			t.Errorf("missing %q in:\n%s", sub, s)
		}
	}
}

func mustLack(t *testing.T, s string, subs ...string) {
	t.Helper()
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			t.Errorf("unexpected %q in:\n%s", sub, s)
		}
	}
}

// CLAUDE.md target → a one-line @import; the rules stay in .trellis/, not inlined.
func TestE2E_M1_ClaudeTargetImports(t *testing.T) {
	noClaude(t)
	dir := t.TempDir()
	seedFile(t, dir, "CLAUDE.md", "# House\n\nKeep this.\n")

	out := mustApplyM1(t, dir, "CLAUDE.md")
	mustHave(t, out, "imports .trellis/trellis.md")

	c := readFile(t, filepath.Join(dir, "CLAUDE.md"))
	mustHave(t, c, "Keep this.", trellisBegin, "@.trellis/trellis.md")
	mustLack(t, c, "inv-directional-flow") // rules live in .trellis/, imported not inlined
	if _, err := os.Stat(filepath.Join(dir, ".trellis", "invariants.md")); err != nil {
		t.Errorf("did not bundle the invariant reference: %v", err)
	}
}

// AGENTS.md target → the rules are inlined (no @import), preserving existing content.
func TestE2E_M1_AgentsTargetInlines(t *testing.T) {
	noClaude(t)
	dir := t.TempDir()
	seedFile(t, dir, "AGENTS.md", "# Agents\n\nHouse rule.\n")

	out := mustApplyM1(t, dir, "AGENTS.md")
	mustHave(t, out, "inlines the rules (no @import)")

	a := readFile(t, filepath.Join(dir, "AGENTS.md"))
	mustHave(t, a, "House rule.", trellisBegin, "without its human approval", "inv-directional-flow")
	mustLack(t, a, "@.trellis/trellis.md")
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); !os.IsNotExist(err) {
		t.Error("targeting AGENTS.md must not create CLAUDE.md")
	}
}

// No instruction file present → the target is created.
func TestE2E_M1_CreatesTargetWhenAbsent(t *testing.T) {
	noClaude(t)
	dir := t.TempDir()
	mustApplyM1(t, dir, "CLAUDE.md")
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err != nil {
		t.Errorf("M1 should create the target when none exists: %v", err)
	}
}

// Both files present → --target picks one; the other is left untouched.
func TestE2E_M1_TargetFlagPicksAmongDetected(t *testing.T) {
	noClaude(t)
	dir := t.TempDir()
	seedFile(t, dir, "CLAUDE.md", "# c\n")
	seedFile(t, dir, "AGENTS.md", "# a\n")

	mustApplyM1(t, dir, "AGENTS.md")
	mustHave(t, readFile(t, filepath.Join(dir, "AGENTS.md")), trellisBegin)
	mustLack(t, readFile(t, filepath.Join(dir, "CLAUDE.md")), trellisBegin)
}

// Re-run is idempotent — the managed block is replaced, never duplicated.
func TestE2E_M1_IdempotentReRun(t *testing.T) {
	noClaude(t)
	dir := t.TempDir()
	seedFile(t, dir, "CLAUDE.md", "# House\n\nKeep.\n")

	mustApplyM1(t, dir, "CLAUDE.md")
	first := readFile(t, filepath.Join(dir, "CLAUDE.md"))
	mustApplyM1(t, dir, "CLAUDE.md")
	second := readFile(t, filepath.Join(dir, "CLAUDE.md"))

	if first != second {
		t.Errorf("re-run not idempotent:\n--- first\n%s\n--- second\n%s", first, second)
	}
	if n := strings.Count(second, trellisBegin); n != 1 {
		t.Errorf("re-run duplicated the managed block (%d begin markers)", n)
	}
}

// An unknown --target is a loud error, not a silent guess (D1).
func TestE2E_M1_UnknownTargetErrors(t *testing.T) {
	noClaude(t)
	if _, err := run2("", "setup", "--dir", t.TempDir(), "--mode", "m1", "--target", "READrME.txt"); err == nil {
		t.Fatal("an unknown --target should be a loud error")
	}
}
