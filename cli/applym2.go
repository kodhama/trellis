package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// runMorph is indirected so tests exercise the M2 flow without invoking claude.
var runMorph = execClaudeMorph

const morphBranch = "trellis/morph"

// applyM2 performs the model-driven morph: on a fresh git branch, invoke the
// harness to rewrite the project's own instructions to bake in the profile. It
// refuses without git (the rewrite must be reviewable + revertable) and never
// edits the working branch in place.
func applyM2(dir string, plan Plan) (string, error) {
	if !isGitRepo(dir) {
		return "", fmt.Errorf("M2 rewrites the project on a git branch, but %q is not a git repository — run `git init` first, or choose M1 (alongside)", dir)
	}
	if err := gitCheckoutNewBranch(dir, morphBranch); err != nil {
		return "", err
	}
	if err := runMorph(dir, plan.Model.Alias, morphPrompt(plan)); err != nil {
		return "", fmt.Errorf("morph failed on branch %q: %w", morphBranch, err)
	}
	return fmt.Sprintf("applied (M2 morph) on branch %q — review the diff and open a PR before merging.\n", morphBranch), nil
}

func isGitRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}

func gitCheckoutNewBranch(dir, branch string) error {
	cmd := exec.Command("git", "-C", dir, "checkout", "-b", branch)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("creating branch %s: %v: %s", branch, err, string(out))
	}
	return nil
}

func morphPrompt(plan Plan) string {
	return fmt.Sprintf(`Apply the Trellis governance layer to THIS project by rewriting its own agent instructions (the "morph").

Posture: %s. Active invariants: %s. Enforcement lean: %s.

Rewrite the project's instruction files (e.g. CLAUDE.md and any rule/convention files) to bake in these invariants, in the project's own voice and structure. Preserve the project's existing behaviors unless they directly conflict.

The single most important behavior to encode: surface any human-gated handover performed without its human approval (invariant B2). Agent-gated handovers proceed silently. Respect whatever gatekeeping the project already declares — detect it, do not impose it.

Make the edits directly and keep them reviewable.`, plan.Profile.Name, activeList(plan), plan.Profile.C1Lean)
}

// execClaudeMorph runs the harness headlessly in dir to perform the rewrite.
// Flags verified against the Claude Code CLI docs (claude-code-guide):
//   -p                         headless — runs the full agentic loop then exits
//   --permission-mode acceptEdits  auto-approves file edits (no prompts), still
//                              blocks dangerous shell/network
//   --model <alias>            sonnet|opus|haiku|… (the chosen tier)
// The working directory is implicit — setting cmd.Dir is enough.
func execClaudeMorph(dir, modelAlias, prompt string) error {
	var args []string
	if modelAlias != "" {
		args = append(args, "--model", modelAlias)
	}
	args = append(args, "--permission-mode", "acceptEdits", "-p", prompt)
	cmd := exec.Command("claude", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%v: %s", err, string(out))
	}
	return nil
}
