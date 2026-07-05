package main

import (
	"os"
	"os/exec"
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

	out, err := run2("", "setup", "--dir", dir, "--profile", "b", "--mode", "m1", "--apply")
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
	if !strings.Contains(c, trellisBegin) || !strings.Contains(c, "@.trellis/trellis.md") {
		t.Error("e2e did not install the minimal trellis import block")
	}
	if header := readFile(t, filepath.Join(dir, ".trellis", "trellis.md")); !strings.Contains(header, "without its human approval") {
		t.Error("e2e did not write the B2 surfacing behavior into the header")
	}
	if _, err := os.Stat(filepath.Join(dir, ".trellis", "invariants.md")); err != nil {
		t.Error("e2e did not bundle the invariant reference")
	}
}

func TestE2E_M2_RefusesWithoutGit(t *testing.T) {
	claudePresent(t)
	withRunMorph(t, func(_, _, _ string) error { return nil }) // must not be reached
	dir := t.TempDir()                                          // no git
	_, err := run2("", "setup", "--dir", dir, "--profile", "a", "--mode", "m2", "--model", "high", "--apply")
	if err == nil || !strings.Contains(err.Error(), "not a git repository") {
		t.Fatalf("M2 --apply without git should refuse, got %v", err)
	}
}

func TestE2E_M2_MorphsOnGitRepo(t *testing.T) {
	claudePresent(t)
	morphed := false
	withRunMorph(t, func(_, _, _ string) error { morphed = true; return nil })
	dir := t.TempDir()
	initGitRepo(t, dir)
	out, err := run2("", "setup", "--dir", dir, "--profile", "a", "--mode", "m2", "--model", "high", "--apply")
	if err != nil {
		t.Fatalf("M2 --apply on a git repo: %v", err)
	}
	if !morphed {
		t.Error("the morph was not invoked")
	}
	if !strings.Contains(out, morphBranch) {
		t.Errorf("summary should name the morph branch, got:\n%s", out)
	}
	if gitCurrentBranch(t, dir) != morphBranch {
		t.Error("should be on the morph branch")
	}
}

// TestE2E_M2_RealMorphJudged is the model-in-the-loop e2e: it invokes the real
// harness to morph a sample project, then a judge model grades whether the rewrite
// encoded the B2 surfacing behavior. DEFAULT-OFF — set TRELLIS_E2E_MODEL (e.g.
// sonnet) to run it; it spends tokens on your account and is never run in CI.
func TestE2E_M2_RealMorphJudged(t *testing.T) {
	model := os.Getenv("TRELLIS_E2E_MODEL")
	if model == "" {
		t.Skip("set TRELLIS_E2E_MODEL (e.g. sonnet) to run the model-in-the-loop e2e")
	}
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude not on PATH")
	}
	// real runMorph (no mock) — invokes claude.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"),
		[]byte("# Sample\n\nHouse rule: the spec→plan handover requires human approval.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, dir)
	mustGit(t, dir, "add", "-A")
	mustGit(t, dir, "commit", "-q", "-m", "sample")

	if _, err := run2("", "setup", "--dir", dir, "--profile", "a", "--mode", "m2", "--model", model, "--apply"); err != nil {
		t.Fatalf("real morph: %v", err)
	}
	// deterministic pre-check: the morph left changes in the working tree.
	if strings.TrimSpace(mustGit(t, dir, "status", "--porcelain")) == "" {
		t.Fatal("morph produced no changes")
	}
	// LLM judge (sonnet): did the rewrite encode the human-gate surfacing behavior?
	if !judgeSurfacing(t, "sonnet", mustGit(t, dir, "diff", "HEAD")) {
		t.Error("judge: the morph did not clearly encode surfacing a bypassed human gate")
	}
}

func mustGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return string(out)
}

// judgeSurfacing asks a judge model whether the diff encodes the B2 surfacing
// behavior, returning true only on a clear PASS.
func judgeSurfacing(t *testing.T, model, diff string) bool {
	t.Helper()
	rubric := "Below is a diff that installed a governance layer into a project. " +
		"Question: does it make the project's agents SURFACE (flag/warn) when a human-gated handover happens WITHOUT the required human approval? " +
		"Answer with exactly PASS or FAIL as the first word, then one sentence.\n\n" + diff
	out, err := exec.Command("claude", "--model", model, "--permission-mode", "dontAsk", "-p", rubric).CombinedOutput()
	if err != nil {
		t.Fatalf("judge: %v: %s", err, out)
	}
	return strings.HasPrefix(strings.TrimSpace(strings.ToUpper(string(out))), "PASS")
}

// TestE2E_DryRunWritesNothing confirms the default is a no-write dry run.
func TestE2E_DryRunWritesNothing(t *testing.T) {
	claudePresent(t)
	dir := t.TempDir()
	out, err := run2("", "setup", "--dir", dir, "--profile", "b", "--mode", "m1")
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
