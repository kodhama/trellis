package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func withLookPath(t *testing.T, fn func(string) (string, error)) {
	t.Helper()
	orig := lookPath
	lookPath = fn
	t.Cleanup(func() { lookPath = orig })
}

func TestDetectHarness(t *testing.T) {
	t.Run("no claude executable detects nothing", func(t *testing.T) {
		withLookPath(t, func(string) (string, error) { return "", exec.ErrNotFound })
		if h, ok := detectHarness(t.TempDir()); ok {
			t.Errorf("without the claude executable no harness should be detected, got %+v", h)
		}
	})

	t.Run("claude executable detects Claude Code", func(t *testing.T) {
		withLookPath(t, func(string) (string, error) { return "/usr/local/bin/claude", nil })
		h, ok := detectHarness(t.TempDir())
		if !ok || h.Name != "Claude Code" || h.Executable != "/usr/local/bin/claude" {
			t.Errorf("got %+v ok=%v", h, ok)
		}
	})

	t.Run("existing project files are noted as context", func(t *testing.T) {
		withLookPath(t, func(string) (string, error) { return "/usr/local/bin/claude", nil })
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# host"), 0o644); err != nil {
			t.Fatal(err)
		}
		h, _ := detectHarness(dir)
		if !strings.Contains(h.Detail, "CLAUDE.md") && !strings.Contains(h.Detail, ".claude") {
			t.Errorf("project files should be noted in Detail, got %q", h.Detail)
		}
	})
}

func TestHasClaudeProjectFiles(t *testing.T) {
	t.Run("empty dir", func(t *testing.T) {
		if hasClaudeProjectFiles(t.TempDir()) {
			t.Error("empty dir should have no Claude project files")
		}
	})
	t.Run("CLAUDE.md file", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if !hasClaudeProjectFiles(dir) {
			t.Error("CLAUDE.md should be recognised")
		}
	})
	t.Run(".claude directory", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, ".claude"), 0o755); err != nil {
			t.Fatal(err)
		}
		if !hasClaudeProjectFiles(dir) {
			t.Error(".claude/ should be recognised")
		}
	})
	t.Run("a CLAUDE.md directory is not the file signal", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, "CLAUDE.md"), 0o755); err != nil {
			t.Fatal(err)
		}
		if hasClaudeProjectFiles(dir) {
			t.Error("a CLAUDE.md directory should not count as the CLAUDE.md file")
		}
	})
}
