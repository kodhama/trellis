package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

// Harness is a detected agent harness the setup CLI can ride.
type Harness struct {
	Name       string // human name, e.g. "Claude Code"
	Executable string // path to the harness CLI — what we query for models and invoke
	Detail     string // how it was found
}

// lookPath is indirected so tests can simulate a harness being present or absent
// without depending on what happens to be installed on the test machine.
var lookPath = exec.LookPath

// detectHarness finds a supported harness the CLI can ride. v0 supports Claude
// Code, which requires the `claude` executable on PATH: the CLI queries it for a
// model and invokes it to apply changes, so it is the *binary* — not a project
// folder — that makes a harness usable. A .claude/ directory or CLAUDE.md is
// recorded as corroborating context (an existing Claude setup), never the signal.
func detectHarness(dir string) (Harness, bool) {
	path, err := lookPath("claude")
	if err != nil {
		return Harness{}, false
	}
	h := Harness{Name: "Claude Code", Executable: path, Detail: "claude executable on PATH"}
	if hasClaudeProjectFiles(dir) {
		h.Detail += "; existing .claude/ or CLAUDE.md"
	}
	return h, true
}

// hasClaudeProjectFiles reports whether dir already looks like a Claude Code
// project. Used for a tailored "install the CLI" hint and, later, for
// augment-not-clobber — not for detection itself.
func hasClaudeProjectFiles(dir string) bool {
	if fi, err := os.Stat(filepath.Join(dir, ".claude")); err == nil && fi.IsDir() {
		return true
	}
	if fi, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err == nil && !fi.IsDir() {
		return true
	}
	return false
}
