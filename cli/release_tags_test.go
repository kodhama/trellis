package main

// The tag logic behind .github/workflows/release-tag.yml lives in
// scripts/release-tags.sh, so these tests drive it against throwaway git
// repositories: `plan` lists the tags the workflow would create, and `verify`
// fails on a version at or above the floor whose tag is missing or names the
// wrong commit (decision-2026-10-20-trellis-delivers-rules-as-committed-files
// point 11). Nothing here reaches a remote.

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type tagRepo struct {
	t   *testing.T
	dir string
}

func newTagRepo(t *testing.T) *tagRepo {
	t.Helper()
	r := &tagRepo{t: t, dir: t.TempDir()}
	r.git("init", "-q", "-b", "main")
	r.git("config", "user.email", "test@example.invalid")
	r.git("config", "user.name", "test")
	r.git("config", "commit.gpgsign", "false")
	r.git("config", "tag.gpgsign", "false")
	return r
}

// isolatedGitEnv keeps the maintainer's own git config, signing included, out
// of the throwaway repositories.
func isolatedGitEnv() []string {
	return append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
}

func (r *tagRepo) git(args ...string) string {
	r.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir
	cmd.Env = isolatedGitEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

// commit writes package.json with version (no field when empty) and a note
// file, so a commit can leave the version as it was, and returns the commit.
func (r *tagRepo) commit(version, note string) string {
	r.t.Helper()
	pkg := "{\n  \"name\": \"trellis\",\n  \"private\": true\n}\n"
	if version != "" {
		pkg = "{\n  \"name\": \"trellis\",\n  \"version\": \"" + version + "\",\n  \"private\": true\n}\n"
	}
	return r.commitPackage(pkg, note)
}

// commitPackage commits package.json as given, deleting it when pkg is empty.
func (r *tagRepo) commitPackage(pkg, note string) string {
	r.t.Helper()
	path := filepath.Join(r.dir, "package.json")
	if pkg == "" {
		removeFileT(r.t, path)
	} else {
		writeFileT(r.t, path, pkg)
	}
	writeFileT(r.t, filepath.Join(r.dir, "note.txt"), note+"\n")
	r.git("add", "-A")
	r.git("commit", "-q", "-m", note)
	return r.git("rev-parse", "HEAD")
}

// run runs the script in mode with the floor set as given, returning its
// combined output and exit code.
func (r *tagRepo) run(mode, floor string) (string, int) {
	r.t.Helper()
	script, err := filepath.Abs(filepath.Join("..", "scripts", "release-tags.sh"))
	if err != nil {
		r.t.Fatal(err)
	}
	cmd := exec.Command(script, mode)
	cmd.Dir = r.dir
	cmd.Env = append(isolatedGitEnv(), "TRELLIS_TAG_FLOOR="+floor)
	out, err := cmd.CombinedOutput()
	code := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		code = exitErr.ExitCode()
	} else if err != nil {
		r.t.Fatalf("release-tags.sh %s: %v", mode, err)
	}
	return string(out), code
}

func (r *tagRepo) wantPlan(floor, want string) {
	r.t.Helper()
	out, code := r.run("plan", floor)
	if code != 0 {
		r.t.Fatalf("plan exited %d:\n%s", code, out)
	}
	if got := planLines(out); got != want {
		r.t.Errorf("plan listed %q, want %q", got, want)
	}
}

// planLines keeps only the tag lines of plan's output, dropping its notes.
func planLines(out string) string {
	var lines []string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "v") {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func (r *tagRepo) wantVerify(floor string, wantCode int, wantText ...string) string {
	r.t.Helper()
	out, code := r.run("verify", floor)
	if code != wantCode {
		r.t.Fatalf("verify exited %d, want %d:\n%s", code, wantCode, out)
	}
	for _, text := range wantText {
		if !strings.Contains(out, text) {
			r.t.Errorf("verify output lacks %q:\n%s", text, out)
		}
	}
	return out
}

func TestReleaseTagsDoNothingWithoutAFloor(t *testing.T) {
	r := newTagRepo(t)
	r.commit("", "before the version line")
	r.commit("0.24.8", "U1")
	r.commit("0.25.0", "first CLI release")
	r.wantPlan("", "")
	r.wantVerify("", 0)
}

func TestReleaseTagsListTheMissingTagAndFailVerify(t *testing.T) {
	r := newTagRepo(t)
	r.commit("0.24.8", "U1")
	release := r.commit("0.25.0", "first CLI release")
	r.wantPlan("0.25.0", "v0.25.0 "+release)
	r.wantVerify("0.25.0", 1, "v0.25.0", release)
}

func TestReleaseTagsLeaveAnExistingTagAlone(t *testing.T) {
	r := newTagRepo(t)
	release := r.commit("0.25.0", "first CLI release")
	r.git("tag", "v0.25.0", release)
	r.wantPlan("0.25.0", "")
	r.wantVerify("0.25.0", 0)
}

func TestReleaseTagsNameTheCommitThatSetTheVersion(t *testing.T) {
	// A release whose tag job failed, followed by an unrelated push at the same
	// version, gets its tag on the release commit, not on the later head.
	r := newTagRepo(t)
	release := r.commit("0.25.0", "first CLI release")
	r.commit("0.25.0", "an unrelated change")
	r.wantPlan("0.25.0", "v0.25.0 "+release)
}

func TestReleaseTagsVerifyEveryVersionSinceTheFloor(t *testing.T) {
	// The head's version is tagged, but an earlier one above the floor is not.
	r := newTagRepo(t)
	first := r.commit("0.25.0", "first CLI release")
	second := r.commit("0.26.0", "second CLI release")
	r.git("tag", "v0.26.0", second)
	r.wantPlan("0.25.0", "v0.25.0 "+first)
	if out := r.wantVerify("0.25.0", 1, "v0.25.0", first); strings.Contains(out, "::error::v0.26.0") {
		t.Errorf("verify flagged the tagged head version:\n%s", out)
	}
}

func TestReleaseTagsSkipVersionsBelowTheFloor(t *testing.T) {
	r := newTagRepo(t)
	r.commit("0.24.8", "U1")
	r.commit("0.24.9", "a plugin release before the CLI's first")
	release := r.commit("0.25.0", "first CLI release")
	r.git("tag", "v0.25.0", release)
	r.wantPlan("0.25.0", "")
	r.wantVerify("0.25.0", 0)
}

func TestReleaseTagsCompareTheFloorAsAVersion(t *testing.T) {
	// As text "0.9.0" sorts above "0.10.0"; as a version it is below.
	r := newTagRepo(t)
	r.commit("0.9.0", "below the floor")
	release := r.commit("0.10.0", "at the floor")
	r.wantPlan("0.10.0", "v0.10.0 "+release)
}

func TestReleaseTagsFailATagOnTheWrongCommit(t *testing.T) {
	// A tag made by hand on another commit is never moved, so verify must name it:
	// that version is burned until someone decides what to do.
	r := newTagRepo(t)
	early := r.commit("0.24.8", "U1")
	release := r.commit("0.25.0", "first CLI release")
	r.git("tag", "v0.25.0", early)
	r.wantPlan("0.25.0", "")
	r.wantVerify("0.25.0", 1, "v0.25.0", early, release)
}

func TestReleaseTagsRefuseAMalformedFloorOrMode(t *testing.T) {
	r := newTagRepo(t)
	r.commit("0.25.0", "first CLI release")
	for _, floor := range []string{"v0.25.0", "0.25"} {
		for _, mode := range []string{"plan", "verify"} {
			if out, code := r.run(mode, floor); code != 2 {
				t.Errorf("%s with floor %s exited %d, want 2:\n%s", mode, floor, code, out)
			}
		}
	}
	for _, mode := range []string{"", "tag"} {
		if out, code := r.run(mode, "0.25.0"); code != 2 {
			t.Errorf("mode %q exited %d, want 2:\n%s", mode, code, out)
		}
	}
}

func TestReleaseTagsCompareTheTagsCommit(t *testing.T) {
	// An annotated tag is a tag object; what must match is the commit it names.
	r := newTagRepo(t)
	release := r.commit("0.25.0", "first CLI release")
	r.git("tag", "-a", "-m", "v0.25.0", "v0.25.0", release)
	r.wantPlan("0.25.0", "")
	r.wantVerify("0.25.0", 0, "ok")
}

func TestReleaseTagsIgnoreWhatIsBelowTheFloor(t *testing.T) {
	// Early commits with no version, and the retired v0 CLI's tags, are below it.
	r := newTagRepo(t)
	old := r.commit("", "before the version line")
	r.git("tag", "v0.2.30", old)
	r.commit("0.24.8", "U1")
	release := r.commit("0.25.0", "first CLI release")
	r.wantPlan("0.25.0", "v0.25.0 "+release)
	r.git("tag", "v0.25.0", release)
	r.wantVerify("0.25.0", 0)
}

func TestReleaseTagsTagTheMergeCommit(t *testing.T) {
	// A version set on a side branch is set, on main's first-parent line, by the
	// merge that brought it in.
	r := newTagRepo(t)
	r.commit("0.24.8", "U1")
	r.git("checkout", "-q", "-b", "release")
	r.commit("0.25.0", "first CLI release, on a branch")
	r.git("checkout", "-q", "main")
	r.git("merge", "-q", "--no-ff", "-m", "merge the release", "release")
	merge := r.git("rev-parse", "HEAD")
	r.wantPlan("0.25.0", "v0.25.0 "+merge)
}

func TestReleaseTagsRefuseAPrereleaseAboveTheFloor(t *testing.T) {
	// sort -V puts 0.25.0-rc.1 after 0.25.0, so it cannot be compared.
	r := newTagRepo(t)
	r.commit("0.25.0", "first CLI release")
	r.commit("0.25.1-rc.1", "a prerelease")
	for _, mode := range []string{"plan", "verify"} {
		out, code := r.run(mode, "0.25.0")
		if code != 2 || !strings.Contains(out, "0.25.1-rc.1") {
			t.Errorf("%s exited %d, want 2 naming 0.25.1-rc.1:\n%s", mode, code, out)
		}
	}
}

func TestReleaseTagsRefuseAShallowClone(t *testing.T) {
	r := newTagRepo(t)
	r.commit("0.25.0", "first CLI release")
	r.commit("0.26.0", "second CLI release")
	shallow := &tagRepo{t: t, dir: filepath.Join(t.TempDir(), "shallow")}
	// git ignores --depth for a plain local path, so clone through file://.
	r.git("clone", "-q", "--depth", "1", "file://"+r.dir, shallow.dir)
	for _, mode := range []string{"plan", "verify"} {
		if out, code := shallow.run(mode, "0.25.0"); code != 2 {
			t.Errorf("%s in a shallow clone exited %d, want 2:\n%s", mode, code, out)
		}
	}
}

func TestReleaseTagsReadOnlyTheTopLevelVersion(t *testing.T) {
	// A nested "version" key is not the package's, and a commit that deletes
	// package.json leaves the version as it was.
	r := newTagRepo(t)
	r.commitPackage("{\n  \"name\": \"trellis\",\n  \"engines\": { \"version\": \"0.30.0\" }\n}\n", "a nested key")
	release := r.commit("0.25.0", "first CLI release")
	r.commitPackage("", "package.json deleted")
	r.commit("0.25.0", "package.json restored at the same version")
	r.wantPlan("0.25.0", "v0.25.0 "+release)
}

func TestReleaseTagsFailAVersionThatComesBack(t *testing.T) {
	// One tag names one commit, so a version that returns after main moved past
	// it cannot be tagged. verify says so, naming the way out, while plan still
	// lists the releases after it, so one repeat does not stop later tags.
	r := newTagRepo(t)
	first := r.commit("0.25.0", "first CLI release")
	second := r.commit("0.26.0", "second CLI release")
	back := r.commit("0.25.0", "a revert back to the first version")
	third := r.commit("0.27.0", "third CLI release")
	r.git("tag", "v0.25.0", first)
	r.git("tag", "v0.26.0", second)
	out, code := r.run("plan", "0.25.0")
	if code != 0 || planLines(out) != "v0.27.0 "+third {
		t.Errorf("plan exited %d listing %q, want 0 listing v0.27.0 %s:\n%s", code, planLines(out), third, out)
	}
	for _, text := range []string{"0.25.0", first, back} {
		if !strings.Contains(out, text) {
			t.Errorf("plan's report lacks %q:\n%s", text, out)
		}
	}
	r.git("tag", "v0.27.0", third)
	r.wantVerify("0.25.0", 1, "0.25.0", first, back, "TRELLIS_TAG_FLOOR")
	// The way out the message names: a floor raised past the repeated version.
	r.wantVerify("0.26.1", 0)
}

func TestReleaseTagsIgnoreAChangeThatKeepsTheVersion(t *testing.T) {
	// A Dependabot bump edits package.json at the same version: no new release.
	r := newTagRepo(t)
	release := r.commit("0.25.0", "first CLI release")
	r.commitPackage("{\n  \"name\": \"trellis\",\n  \"version\": \"0.25.0\",\n  \"private\": true,\n  \"devDependencies\": { \"eslint\": \"10.12.0\" }\n}\n", "bump a dev dependency")
	r.wantPlan("0.25.0", "v0.25.0 "+release)
}

// TestReleaseTagWorkflowShape pins the properties of release-tag.yml that keep
// tagging on main and create-only. The repo has no YAML parser, so it reads the
// file as text, job by job, as TestCliCIProvidesNode20BeforeGoTests does.
func TestReleaseTagWorkflowShape(t *testing.T) {
	workflow := readFileT(t, "../.github/workflows/release-tag.yml")
	jobs := strings.Index(workflow, "\njobs:\n")
	planAt := strings.Index(workflow, "\n  plan:\n")
	tagAt := strings.Index(workflow, "\n  tag:\n")
	verifyAt := strings.Index(workflow, "\n  verify:\n")
	if jobs < 0 || planAt < jobs || tagAt < planAt || verifyAt < tagAt {
		t.Fatal("release-tag.yml must define the jobs plan, tag and verify, in that order")
	}
	top, plan, tag, verify := workflow[:jobs], workflow[planAt:tagAt], workflow[tagAt:verifyAt], workflow[verifyAt:]

	// U1 commits the floor unset, so its merge tags nothing; U9 sets it, and
	// changes this line with it.
	if !strings.Contains(top, "\n  TRELLIS_TAG_FLOOR: \"\"\n") {
		t.Error("the floor must be committed unset until the CLI's first release")
	}
	if !strings.Contains(top, "\non:\n  push:\n    branches: [main]\n  workflow_dispatch:\n\n") {
		t.Error("release-tag must run on pushes to main and on dispatch, and on nothing else")
	}
	if !strings.Contains(top, "\npermissions:\n  contents: read\n") {
		t.Error("release-tag's token must be read-only; only the App's token writes")
	}
	if !strings.Contains(top, "group: release-tag-${{ github.ref }}") || !strings.Contains(top, "cancel-in-progress: false") {
		t.Error("release-tag runs must queue per ref, never cancel one another")
	}
	// Every push before the CLI's first release has nothing to tag, and this guard
	// keeps those pushes out of the release environment.
	if !strings.Contains(tag, "needs: plan\n    if: needs.plan.outputs.tags != ''\n") {
		t.Error("the tag job must run only when plan listed a tag")
	}
	if !regexp.MustCompile(`uses: actions/create-github-app-token@[0-9a-f]{40}\n`).MatchString(tag) {
		t.Error("the step that holds the App's key must pin its action by commit")
	}
	if strings.Count(workflow, "environment: release") != 1 || !strings.Contains(tag, "environment: release") {
		t.Error("only the tag job may use the release environment, which holds the App's key")
	}
	if !strings.Contains(tag, `gh api "repos/$GITHUB_REPOSITORY/git/refs" -f ref=`) {
		t.Error("the tag job must create tags with a POST to git/refs")
	}
	for _, verb := range []string{"PATCH", "DELETE", "--method", "-X ", "force"} {
		if strings.Contains(workflow, verb) {
			t.Errorf("release-tag must only create tags, but it contains %q", verb)
		}
	}
	for name, job := range map[string]string{"plan": plan, "verify": verify} {
		if !strings.Contains(job, "fetch-depth: 0") {
			t.Errorf("the %s job must check out the full history", name)
		}
	}
	if !strings.Contains(verify, "needs: [plan, tag]") || !strings.Contains(verify, "if: always()") {
		t.Error("verify must run after tag whatever tag did")
	}
}
