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

func (r *tagRepo) git(args ...string) string {
	r.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir
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
	for name, body := range map[string]string{"package.json": pkg, "note.txt": note + "\n"} {
		if err := os.WriteFile(filepath.Join(r.dir, name), []byte(body), 0o644); err != nil {
			r.t.Fatal(err)
		}
	}
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
	cmd.Env = append(os.Environ(), "TRELLIS_TAG_FLOOR="+floor)
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

func (r *tagRepo) wantVerify(floor string, wantCode int, wantText ...string) {
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
	r.wantVerify("0.25.0", 1, "v0.25.0", first)
	out, _ := r.run("verify", "0.25.0")
	if strings.Contains(out, "v0.26.0 is") {
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

func TestReleaseTagsRefuseAMalformedFloor(t *testing.T) {
	r := newTagRepo(t)
	r.commit("0.25.0", "first CLI release")
	for _, mode := range []string{"plan", "verify"} {
		if out, code := r.run(mode, "v0.25"); code != 2 {
			t.Errorf("%s with floor v0.25 exited %d, want 2:\n%s", mode, code, out)
		}
	}
}
