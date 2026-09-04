package main

// Tests for .github/scripts/decision-id-guard.sh — the CI guard that refuses a
// pull request claiming an already-taken `decisions/NNNN-*.md` id (TRL-40,
// decision-0089).
//
// The harness shape is the repo's existing one: execute the production script as
// an external file against fixtures, the way cli/plugin_hook_test.go executes
// plugins/trellis/hooks/staleness.sh and cli/install_script_test.go executes
// install.sh. That is also why `go test -count=1` is not optional in this repo —
// the cache keys on Go sources and cannot see a mutation to the script.
//
// Every input the guard needs is injectable, so these run offline: no `gh`, no
// network, no git fetch. The environment handed to the script is built from
// scratch rather than inherited, which is what proves the live `gh` branch is
// never reached — an unset GUARD_PR_FILES with no GUARD_REPO exits 2, and none
// of these tests see that.

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// decisionIDGuardPath is the production script, not a copy.
func decisionIDGuardPath(t *testing.T) string {
	t.Helper()
	p, err := filepath.Abs("../.github/scripts/decision-id-guard.sh")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("guard script missing: %v", err)
	}
	return p
}

// TestDecisionIDGuardIsExecutable — the workflow invokes the script directly
// (`run: .github/scripts/decision-id-guard.sh`), so a lost +x bit is a broken
// check, not a style nit. The tests below exec it directly for the same reason:
// running it as `sh script` would hide the loss.
func TestDecisionIDGuardIsExecutable(t *testing.T) {
	info, err := os.Stat(decisionIDGuardPath(t))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("decision-id-guard.sh is not executable (mode %v); the workflow runs it directly", info.Mode().Perm())
	}
}

// runGuard executes the guard with the given fixture inputs and returns its
// combined output and exit code.
func runGuard(t *testing.T, prNumber, mainFiles, prFiles string) (string, int) {
	t.Helper()
	cmd := exec.Command(decisionIDGuardPath(t))
	cmd.Dir = t.TempDir() // no git repo: nothing here may touch the working tree
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"GUARD_PR_NUMBER=" + prNumber,
		"GUARD_BASE_REF=main",
		"GUARD_MAIN_FILES=" + mainFiles,
		"GUARD_PR_FILES=" + prFiles,
	}
	out, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("running the guard: %v\n%s", err, out)
		}
		code = ee.ExitCode()
	}
	return string(out), code
}

// mustContain keeps every failure message showing the guard's actual output —
// the guard's whole job is to say something a reader can act on, so a test that
// only asserts an exit code would pass on a silent red.
func mustContain(t *testing.T, out, want, why string) {
	t.Helper()
	if !strings.Contains(out, want) {
		t.Errorf("%s\nwant output containing %q, got:\n%s", why, want, out)
	}
}

func mustNotContain(t *testing.T, out, unwanted, why string) {
	t.Helper()
	if strings.Contains(out, unwanted) {
		t.Errorf("%s\nunwanted %q present in:\n%s", why, unwanted, out)
	}
}

// The world every fixture starts from: main holds 0086 through 0088.
const guardMainFiles = "decisions/0086-the-injected-copy-degrades.md\n" +
	"decisions/0087-one-gateway-for-every-payload-read.md\n" +
	"decisions/0088-the-install-render-follows-strictness.md\n" +
	"decisions/README-not-a-decision.md"

// TestDecisionIDGuardFailsOnIDAlreadyOnMain — the first half of the rule. This
// is recurrence #2 from the ticket in miniature: trellis#252 sat open claiming
// an already-merged decision-0076 and nothing said so.
func TestDecisionIDGuardFailsOnIDAlreadyOnMain(t *testing.T) {
	out, code := runGuard(t, "300", guardMainFiles,
		"300 added decisions/0087-a-different-slug.md")

	if code != 1 {
		t.Errorf("an id already on main must be red; exit %d\n%s", code, out)
	}
	mustContain(t, out, "::error file=decisions/0087-a-different-slug.md::",
		"the error must be annotated onto the offending file")
	mustContain(t, out, "decision-0087 is already on main as decisions/0087-one-gateway-for-every-payload-read.md",
		"the error must name the id AND the file that holds it, so the reader does not have to look")
}

// TestDecisionIDGuardFailsOnLowerNumberedOpenPR — the tie-break, losing side.
// decision-0089 requires the check to state the rule and name the winner in its
// own output, so both are asserted here rather than left to the PR body.
func TestDecisionIDGuardFailsOnLowerNumberedOpenPR(t *testing.T) {
	out, code := runGuard(t, "300", guardMainFiles,
		"299 added decisions/0089-the-older-claim.md\n"+
			"300 added decisions/0089-the-newer-claim.md")

	if code != 1 {
		t.Errorf("a lower-numbered open PR holding the same id must be red; exit %d\n%s", code, out)
	}
	mustContain(t, out, "::error file=decisions/0089-the-newer-claim.md::",
		"the error belongs on this PR's file, not the rival's")
	mustContain(t, out, "open PR #299 (decisions/0089-the-older-claim.md)",
		"the losing PR must be told which PR holds the id and where")
	mustContain(t, out, "the older claim wins",
		"the check must state the tie-break rule in its own output")
	mustContain(t, out, "#299 keeps decision-0089",
		"the check must name the winning PR number")
}

// TestDecisionIDGuardReportsHigherNumberedOpenPR — the tie-break, winning side.
// Green, because failing both PRs would leave neither able to merge without
// talking to the other; but not silent, because the collision is real and the
// other branch has to hear about it.
func TestDecisionIDGuardReportsHigherNumberedOpenPR(t *testing.T) {
	out, code := runGuard(t, "300", guardMainFiles,
		"300 added decisions/0089-the-older-claim.md\n"+
			"301 added decisions/0089-the-newer-claim.md")

	if code != 0 {
		t.Errorf("the lower-numbered PR keeps the id and must stay green; exit %d\n%s", code, out)
	}
	mustNotContain(t, out, "::error",
		"a higher-numbered rival is reported, never failed")
	mustContain(t, out, "::notice file=decisions/0089-the-older-claim.md::",
		"the winning PR is still told the collision exists")
	mustContain(t, out, "open PR #301 (decisions/0089-the-newer-claim.md)",
		"the notice must name the rival PR and its file")
	mustContain(t, out, "the older claim wins",
		"the notice must state the same tie-break rule the error does")
	mustContain(t, out, "#301 renumbers",
		"the notice must say who has to move")
}

// TestDecisionIDGuardSilentWithNoNewDecision — the common case. Most PRs add no
// decision record, and the guard must cost them nothing and say nothing about
// ids they never claimed.
func TestDecisionIDGuardSilentWithNoNewDecision(t *testing.T) {
	out, code := runGuard(t, "300", guardMainFiles,
		"300 added cli/decision_id_guard_test.go\n"+
			"299 added decisions/0089-someone-elses.md")

	if code != 0 {
		t.Errorf("a PR that adds no decision record must be green; exit %d\n%s", code, out)
	}
	mustNotContain(t, out, "::error", "nothing was claimed, so nothing can collide")
	mustNotContain(t, out, "::notice", "nothing was claimed, so nothing to report")
	mustContain(t, out, "adds no decisions/NNNN-*.md file",
		"a silent green should still say why it had nothing to do")
}

// TestDecisionIDGuardFailsOnlyTheCollidingID — a branch may add two records.
// The one that is free must not be dragged red with the one that is not,
// otherwise the remedy the guard prints ("renumber") is ambiguous.
func TestDecisionIDGuardFailsOnlyTheCollidingID(t *testing.T) {
	out, code := runGuard(t, "300", guardMainFiles,
		"300 added decisions/0088-collides-with-main.md\n"+
			"300 added decisions/0090-perfectly-free.md")

	if code != 1 {
		t.Errorf("one colliding id out of two is still red; exit %d\n%s", code, out)
	}
	mustContain(t, out, "::error file=decisions/0088-collides-with-main.md::",
		"the taken id must be the one flagged")
	mustNotContain(t, out, "::error file=decisions/0090-perfectly-free.md::",
		"the free id must not be flagged")
	mustContain(t, out, "decision-0090 (decisions/0090-perfectly-free.md) — free.",
		"the free id must be reported clear, so the author knows which one to move")
}

// TestDecisionIDGuardIgnoresNonAddedFiles — editing decision-0087 does not take
// 0087, and removing a file takes nothing. This rule lives in the script rather
// than in the workflow's `gh --jq` filter precisely so this test can reach it.
//
// `renamed` is NOT in this list: it used to be, and that was a false pass. See
// TestDecisionIDGuardTreatsRenameDestinationAsAClaim.
func TestDecisionIDGuardIgnoresNonAddedFiles(t *testing.T) {
	out, code := runGuard(t, "300", guardMainFiles,
		"300 modified decisions/0087-one-gateway-for-every-payload-read.md\n"+
			"300 removed decisions/0086-the-injected-copy-degrades.md\n"+
			"300 changed decisions/0088-the-install-render-follows-strictness.md\n"+
			"299 added decisions/0089-the-older-claim.md")

	if code != 0 {
		t.Errorf("modified/removed/changed decision files are not claims; exit %d\n%s", code, out)
	}
	mustNotContain(t, out, "::error", "no file was put at a new decision path")
	mustContain(t, out, "adds no decisions/NNNN-*.md file",
		"the guard should say it found no claim")
}

// TestDecisionIDGuardTreatsRenameDestinationAsAClaim — a confirmed false pass,
// fixed here. The guard used to skip every `renamed` row on the reasoning that
// "a rename is a file that already exists somewhere". That is true of the OLD
// path and false of the NEW one: renaming 0088-old.md to 0090-new.md claims
// 0090 as surely as adding it would, and the guard passed such a branch clean
// while a lower-numbered PR held 0090.
//
// GitHub reports the destination in `filename` and the source in
// `previous_filename`, which is the fourth field of a GUARD_PR_FILES row.
func TestDecisionIDGuardTreatsRenameDestinationAsAClaim(t *testing.T) {
	out, code := runGuard(t, "300", guardMainFiles,
		"299 added decisions/0090-the-older-claim.md\n"+
			"300 renamed decisions/0090-new.md decisions/0086-old.md")

	if code != 1 {
		t.Errorf("a rename INTO a new decision id is a claim and must collide; exit %d\n%s", code, out)
	}
	mustContain(t, out, "::error file=decisions/0090-new.md::",
		"the error belongs on the destination path, which is what claims the id")
	mustContain(t, out, "open PR #299 (decisions/0090-the-older-claim.md)",
		"the rival holding the id must still be named")
}

// TestDecisionIDGuardIgnoresSlugOnlyRename — the other half of the rename rule,
// and the reason it cannot simply be "renamed counts as added". Retitling
// decision-0087 moves `0087-one-gateway.md` to `0087-a-better-slug.md`: same id
// at both ends, so the record claims nothing it did not already hold. Counting
// it would make the guard red on every legitimate retitle, since 0087 is on the
// base branch by definition.
func TestDecisionIDGuardIgnoresSlugOnlyRename(t *testing.T) {
	out, code := runGuard(t, "300", guardMainFiles,
		"300 renamed decisions/0087-a-better-slug.md decisions/0087-one-gateway-for-every-payload-read.md")

	if code != 0 {
		t.Errorf("a rename keeping its own id claims nothing; exit %d\n%s", code, out)
	}
	mustNotContain(t, out, "::error", "the record already held 0087")
	mustContain(t, out, "adds no decisions/NNNN-*.md file",
		"nothing was claimed, so there is nothing to report")
}

// TestDecisionIDGuardFailsOnDuplicateIDWithinOnePR — the second confirmed false
// pass. Two files claiming one id inside a single diff went green: the base and
// rival checks both compare against OTHER sources, so a branch colliding with
// itself was invisible. No tie-break applies here — both files are on the same
// branch and the author picks which one moves.
func TestDecisionIDGuardFailsOnDuplicateIDWithinOnePR(t *testing.T) {
	out, code := runGuard(t, "300", guardMainFiles,
		"300 added decisions/0090-a.md\n"+
			"300 added decisions/0090-b.md")

	if code != 1 {
		t.Errorf("one id under two filenames in one PR must be red; exit %d\n%s", code, out)
	}
	mustContain(t, out, "adds two files claiming decision-0090",
		"the error must say what is wrong")
	mustContain(t, out, "decisions/0090-a.md decisions/0090-b.md",
		"both colliding files must be named — the author has to pick one to move")
	mustNotContain(t, out, "— free.",
		"an id claimed twice is not free; saying so would contradict the error")
}

// TestDecisionIDGuardIgnoresNonDecisionFilenames — `decisions/` holds more than
// numbered records, and a four-digit prefix is the whole claim shape. Flagging
// a note in that directory would make the guard noise.
func TestDecisionIDGuardIgnoresNonDecisionFilenames(t *testing.T) {
	out, code := runGuard(t, "300", guardMainFiles,
		"300 added decisions/README.md\n"+
			"300 added decisions/0089.md\n"+
			"300 added decisions/089-three-digits.md")

	if code != 0 {
		t.Errorf("nothing here has the decisions/NNNN-*.md shape; exit %d\n%s", code, out)
	}
	mustContain(t, out, "adds no decisions/NNNN-*.md file",
		"none of these is a claim")
}

// TestDecisionIDGuardRequiresItsPRNumber — the guard must not run blind. With no
// PR number it cannot tell its own claim from a rival's, and the tie-break is
// undefined; exit 2 (could not run) rather than 0 (clean).
func TestDecisionIDGuardRequiresItsPRNumber(t *testing.T) {
	out, code := runGuard(t, "", guardMainFiles,
		"300 added decisions/0087-a-different-slug.md")

	if code != 2 {
		t.Errorf("a missing PR number is an inability to run, not a pass; exit %d\n%s", code, out)
	}
	mustContain(t, out, "GUARD_PR_NUMBER is required",
		"the failure must say what is missing")
}

// TestDecisionIDGuardFailsClosedOnProcessingError — `set -u` alone does not make
// a shell script fail closed, and this script is `sh`, so `pipefail` is not
// portably available. The original wrote `awk … | sort -u > mine`: a failing awk
// left `sort` writing an empty file, exit 0, and the script reported "no id
// claimed, nothing to check" — a clean pass produced by an internal error.
//
// Induced here by putting a failing `awk` first on PATH, which is the real
// failure this guards against rather than a mocked return value.
func TestDecisionIDGuardFailsClosedOnProcessingError(t *testing.T) {
	shim := t.TempDir()
	broken := "#!/bin/sh\necho 'awk: simulated failure' >&2\nexit 1\n"
	if err := os.WriteFile(filepath.Join(shim, "awk"), []byte(broken), 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(decisionIDGuardPath(t))
	cmd.Dir = t.TempDir()
	cmd.Env = []string{
		"PATH=" + shim + string(os.PathListSeparator) + os.Getenv("PATH"),
		"GUARD_PR_NUMBER=300",
		"GUARD_BASE_REF=main",
		"GUARD_MAIN_FILES=" + guardMainFiles,
		// An id already on main: were awk working, this would be exit 1.
		"GUARD_PR_FILES=300 added decisions/0087-a-different-slug.md",
	}
	out, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("running the guard: %v\n%s", err, out)
		}
		code = ee.ExitCode()
	}

	if code == 0 {
		t.Fatalf("a processing failure must never read as a clean pass; exit 0\n%s", out)
	}
	if code != 2 {
		t.Errorf("a processing failure is exit 2 (could not run), not %d\n%s", code, out)
	}
	mustNotContain(t, string(out), "adds no decisions/NNNN-*.md file",
		"the empty result of a failed awk must not be reported as 'nothing claimed'")
}

// TestDecisionIDGuardEnumeratesEveryOpenPR — a source-level assertion, and
// labelled as one: the live enumeration cannot run offline, so this binds the
// fix rather than the behaviour.
//
// `gh pr list --limit N` caps SILENTLY — past the cap it returns a short list
// and exits 0, so an id held by an unlisted PR reads as free. That contradicts
// the script's own "every open PR" claim, and the failure is invisible. The
// paginated API call has no cap. Verified equivalent against this repository
// live before the swap: both forms returned 266 264 263 262 245 208.
func TestDecisionIDGuardEnumeratesEveryOpenPRWithoutACap(t *testing.T) {
	b, err := os.ReadFile(decisionIDGuardPath(t))
	if err != nil {
		t.Fatal(err)
	}
	src := string(b)

	// Match a COMMAND, not a mention. The script names `gh pr list` in the
	// comment explaining why it does not use it, so a plain Contains flags the
	// fix as the defect — the identical trap agent-workflow-parity.yml
	// documents at its own USES_RE. A line with a `#` before the text is prose.
	command := func(needle string) *regexp.Regexp {
		return regexp.MustCompile(`(?m)^[^#\n]*` + regexp.QuoteMeta(needle))
	}

	if command("gh pr list").MatchString(src) {
		t.Error("the guard enumerates open PRs with `gh pr list`, whose --limit caps silently; use the paginated API instead")
	}
	if !command(`gh api "$repo/pulls"`).MatchString(src) &&
		!command(`gh api "repos/$repo/pulls"`).MatchString(src) {
		t.Error("the guard no longer enumerates open PRs through `gh api .../pulls`")
	}
	if !command("--paginate").MatchString(src) {
		t.Error("the open-PR enumeration must be paginated; without it the list is capped and a silent short list is a false pass")
	}
}

// TestDecisionIDGuardWorkflowRunsTheScript — the pair this change creates: the
// workflow is the only caller of the script, and the script is the only thing
// these tests cover. If the workflow stops invoking it, the suite above goes on
// passing while nothing runs in CI (decision-0028's source/derivative pair).
//
// The assertion is bound to the actual `run:` line, NOT to a substring search.
// A plain Contains for the script path was satisfied by the workflow's own
// comment and its `paths:` list, so rewriting `run:` to `run: true` kept this
// test green while CI invoked nothing.
func TestDecisionIDGuardWorkflowRunsTheScript(t *testing.T) {
	wf, err := os.ReadFile("../.github/workflows/decision-id-guard.yml")
	if err != nil {
		t.Fatalf("the guard's workflow is missing: %v", err)
	}
	body := string(wf)

	runsIt := regexp.MustCompile(`(?m)^[ \t]*run:[ \t]+(\./)?\.github/scripts/decision-id-guard\.sh[ \t]*$`)
	if !runsIt.MatchString(body) {
		t.Error("decision-id-guard.yml has no `run:` step invoking .github/scripts/decision-id-guard.sh — the tested script is not what CI runs")
	}

	for _, want := range []string{
		"GUARD_PR_NUMBER:",    // the input it refuses to run without
		"GUARD_REPO:",         // the one the live gh path needs
		"pull-requests: read", // reading sibling PRs is the whole point
	} {
		if !strings.Contains(body, want) {
			t.Errorf("decision-id-guard.yml no longer carries %q; the guard cannot do its job without it", want)
		}
	}
}
