package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupOverlay applies an M1 overlay in dir (optionally over existing host content).
func setupOverlay(t *testing.T, dir, hostContent string) {
	t.Helper()
	if hostContent != "" {
		if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(hostContent), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := applyM1(dir, planFor("seed")); err != nil {
		t.Fatal(err)
	}
}

func TestRemoveM1CleansUpAndPreservesHost(t *testing.T) {
	dir := t.TempDir()
	setupOverlay(t, dir, "# My Project\n\nHouse rules.\n")
	if _, err := run2("y\n", "remove", "--dir", dir); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".trellis")); !os.IsNotExist(err) {
		t.Error(".trellis/ should be gone")
	}
	c := readFile(t, filepath.Join(dir, "CLAUDE.md"))
	if strings.Contains(c, trellisBegin) {
		t.Error("the trellis block should be stripped")
	}
	if !strings.Contains(c, "House rules.") {
		t.Errorf("host content must be preserved, got: %q", c)
	}
}

func TestRemoveStripsNonClaudeTarget(t *testing.T) {
	// setup can attach to AGENTS.md (decision-0029); remove must strip that block too,
	// not just CLAUDE.md — else it leaves a stale Trellis section behind.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# Agents\n\nHouse rule.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	agents, _ := instructionFileByName("AGENTS.md")
	plan := planFor("b")
	plan.Target = agents
	if _, err := applyM1(dir, plan); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(readFile(t, filepath.Join(dir, "AGENTS.md")), trellisBegin) {
		t.Fatal("precondition: the block should be in AGENTS.md")
	}

	if _, err := run2("y\n", "remove", "--dir", dir); err != nil {
		t.Fatalf("remove: %v", err)
	}
	a := readFile(t, filepath.Join(dir, "AGENTS.md"))
	if strings.Contains(a, trellisBegin) {
		t.Error("remove must strip the Trellis block from AGENTS.md")
	}
	if !strings.Contains(a, "House rule.") {
		t.Error("host content in AGENTS.md must be preserved")
	}
	if _, err := os.Stat(filepath.Join(dir, ".trellis")); !os.IsNotExist(err) {
		t.Error(".trellis/ should be gone")
	}
}

func TestRemoveM1Idempotent(t *testing.T) {
	dir := t.TempDir()
	setupOverlay(t, dir, "# P\n\nrules\n")
	if _, err := run2("y\n", "remove", "--dir", dir); err != nil {
		t.Fatal(err)
	}
	out, err := run2("", "remove", "--dir", dir)
	if err != nil {
		t.Fatalf("second remove: %v", err)
	}
	if !strings.Contains(out, "nothing to remove") {
		t.Errorf("second remove should be a clean no-op, got: %s", out)
	}
}

func TestRemoveM1DropsClaudeItCreated(t *testing.T) {
	dir := t.TempDir()
	setupOverlay(t, dir, "") // no host content: setup created CLAUDE.md with only our block
	if _, err := run2("y\n", "remove", "--dir", dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); !os.IsNotExist(err) {
		t.Error("a CLAUDE.md that held only our block should be removed")
	}
}

func TestRemoveM2WarnsWithRollbackRef(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".trellis"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".trellis", "rollback"), []byte("abc123def456\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := run2("y\n", "remove", "--dir", dir)
	if err != nil {
		t.Fatalf("remove: %v", err)
	}
	if !strings.Contains(out, "morph") {
		t.Errorf("expected an M2 morph warning, got: %s", out)
	}
	if !strings.Contains(out, "git reset --hard abc123def456") {
		t.Errorf("expected the rollback command with the ref, got: %s", out)
	}
}

func TestRemoveNothingToRemove(t *testing.T) {
	out, err := run2("", "remove", "--dir", t.TempDir())
	if err != nil {
		t.Fatalf("remove: %v", err)
	}
	if !strings.Contains(out, "nothing to remove") {
		t.Errorf("empty dir should be a no-op, got: %s", out)
	}
}

func TestStripBlockPreservesSurroundings(t *testing.T) {
	content := "top\n\n" + trellisBegin + "\nX\n" + trellisEnd + "\n\nbottom\n"
	out := stripBlock(content)
	if strings.Contains(out, trellisBegin) || strings.Contains(out, "X") {
		t.Error("block not stripped")
	}
	if !strings.Contains(out, "top") || !strings.Contains(out, "bottom") {
		t.Errorf("surroundings lost: %q", out)
	}
}
