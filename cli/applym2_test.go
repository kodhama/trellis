package main

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

// withRunMorph swaps the harness invocation for a fake, restoring after the test.
func withRunMorph(t *testing.T, fn func(dir, model, prompt string) error) {
	t.Helper()
	orig := runMorph
	runMorph = fn
	t.Cleanup(func() { runMorph = orig })
}

// initGitRepo makes dir a real git repo with one commit (so a branch can be cut).
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.email", "t@example.com"},
		{"config", "user.name", "t"},
		{"commit", "--allow-empty", "-q", "-m", "init"},
	} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
}

func gitCurrentBranch(t *testing.T, dir string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD").CombinedOutput()
	if err != nil {
		t.Fatalf("git branch: %v: %s", err, out)
	}
	return strings.TrimSpace(string(out))
}

// planM2 is an M2 plan with a real model (high → opus alias).
func planM2() Plan {
	p, _ := profileByKey("a")
	m, _ := modeByKey("m2")
	mdl, _ := modelByKey("high")
	return Plan{Harness: Harness{Name: "Claude Code"}, Profile: p, Mode: m, Model: mdl}
}

func TestApplyM2RefusesWithoutGit(t *testing.T) {
	called := false
	withRunMorph(t, func(_, _, _ string) error { called = true; return nil })
	_, err := applyM2(t.TempDir(), planM2())
	if err == nil || !strings.Contains(err.Error(), "not a git repository") {
		t.Fatalf("expected a git-required error, got %v", err)
	}
	if called {
		t.Error("the morph must not run when there is no git repo")
	}
}

func TestApplyM2CreatesBranchAndMorphs(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	var gotModel, gotPrompt string
	withRunMorph(t, func(_, model, prompt string) error {
		gotModel, gotPrompt = model, prompt
		return nil
	})
	out, err := applyM2(dir, planM2())
	if err != nil {
		t.Fatalf("applyM2: %v", err)
	}
	if !strings.Contains(out, morphBranch) {
		t.Errorf("summary should name the branch, got %q", out)
	}
	if cur := gitCurrentBranch(t, dir); cur != morphBranch {
		t.Errorf("expected to be on %q, got %q", morphBranch, cur)
	}
	if gotModel != "opus" {
		t.Errorf("expected model opus, got %q", gotModel)
	}
	if !strings.Contains(gotPrompt, "without its human approval") {
		t.Errorf("morph prompt should carry the B2 behavior, got %q", gotPrompt)
	}
}

func TestApplyM2SurfacesMorphError(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	withRunMorph(t, func(_, _, _ string) error { return fmt.Errorf("boom") })
	if _, err := applyM2(dir, planM2()); err == nil {
		t.Fatal("a failing morph should surface as an error")
	}
}
