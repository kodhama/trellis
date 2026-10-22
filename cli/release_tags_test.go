package main

// The tag logic behind .github/workflows/release-tag.yml lives in
// scripts/release-tags.sh, so these tests drive it against throwaway git
// repositories: `plan` lists the tags the workflow would create, and `verify`
// fails on a version at or above the floor whose tag is missing or names the
// wrong commit (decision-2026-10-20-trellis-delivers-rules-as-committed-files
// point 11). Nothing here reaches a remote.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
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
// of the throwaway repositories. Config passed through the environment
// (GIT_CONFIG_PARAMETERS, GIT_CONFIG_COUNT and its KEY_n and VALUE_n) outranks
// the repository's own, so it is dropped too.
func isolatedGitEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "GIT_CONFIG") {
			env = append(env, kv)
		}
	}
	return append(env, "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
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

// runStreams runs the script in mode with the floor set as given, returning
// its stdout, its stderr and its exit code. The workflow reads plan's stdout
// as the list of tags to create, so the streams are kept apart.
func (r *tagRepo) runStreams(mode, floor string) (string, string, int) {
	r.t.Helper()
	script, err := filepath.Abs(filepath.Join("..", "scripts", "release-tags.sh"))
	if err != nil {
		r.t.Fatal(err)
	}
	cmd := exec.Command(script, mode)
	cmd.Dir = r.dir
	cmd.Env = append(isolatedGitEnv(), "TRELLIS_TAG_FLOOR="+floor)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err = cmd.Run()
	code := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		code = exitErr.ExitCode()
	} else if err != nil {
		r.t.Fatalf("release-tags.sh %s: %v", mode, err)
	}
	return stdout.String(), stderr.String(), code
}

// run is runStreams with both streams together, for checks on what the script
// reports rather than on what plan lists.
func (r *tagRepo) run(mode, floor string) (string, int) {
	r.t.Helper()
	stdout, stderr, code := r.runStreams(mode, floor)
	return stdout + stderr, code
}

// wantPlan checks that plan's stdout is exactly the tag lines in want, one per
// line, and nothing else.
func (r *tagRepo) wantPlan(floor, want string) string {
	r.t.Helper()
	stdout, stderr, code := r.runStreams("plan", floor)
	if code != 0 {
		r.t.Fatalf("plan exited %d:\n%s%s", code, stdout, stderr)
	}
	if want != "" {
		want += "\n"
	}
	if stdout != want {
		r.t.Errorf("plan printed %q on stdout, want %q", stdout, want)
	}
	return stderr
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

func TestReleaseTagsIgnoreGitConfigFromTheEnvironment(t *testing.T) {
	// A hooks path passed as command-scoped config would run a failing hook on
	// every commit in the throwaway repositories, if it reached them.
	hooks := t.TempDir()
	writeFileT(t, filepath.Join(hooks, "pre-commit"), "#!/bin/sh\nexit 1\n")
	if err := os.Chmod(filepath.Join(hooks, "pre-commit"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "core.hooksPath")
	t.Setenv("GIT_CONFIG_VALUE_0", hooks)
	t.Setenv("GIT_CONFIG_PARAMETERS", "'core.hookspath'='"+hooks+"'")
	r := newTagRepo(t)
	release := r.commit("0.25.0", "first CLI release")
	r.wantPlan("0.25.0", "v0.25.0 "+release)
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
	// The note goes to stderr: the workflow reads all of plan's stdout as tags.
	note := r.wantPlan("0.25.0", "v0.27.0 "+third)
	for _, text := range []string{"0.25.0", first, back} {
		if !strings.Contains(note, text) {
			t.Errorf("plan's note on stderr lacks %q:\n%s", text, note)
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

// The tests below read release-tag.yml as text, since the repo has no YAML
// parser, as TestCliCIProvidesNode20BeforeGoTests does. They split it into jobs
// and steps by indentation, check the keys that decide what runs and where,
// and run each step's shell with stand-ins for the script, gh and git, so a
// change to what the workflow does fails a test, and a comment or a reordered
// key does not.

// readReleaseWorkflow returns release-tag.yml's text before `jobs:`, and each
// job's text by name.
func readReleaseWorkflow(t *testing.T) (string, map[string]string) {
	t.Helper()
	workflow := readFileT(t, "../.github/workflows/release-tag.yml")
	at := strings.Index(workflow, "\njobs:\n")
	if at < 0 {
		t.Fatal("release-tag.yml has no jobs")
	}
	body := workflow[at+len("\njobs:\n"):]
	jobs := map[string]string{}
	locs := regexp.MustCompile(`(?m)^  ([A-Za-z_][\w-]*):[ \t]*$`).FindAllStringSubmatchIndex(body, -1)
	for i, loc := range locs {
		end := len(body)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		jobs[body[loc[2]:loc[3]]] = body[loc[0]:end]
	}
	for _, name := range []string{"plan", "tag", "verify"} {
		if jobs[name] == "" {
			t.Fatalf("release-tag.yml must define the job %s", name)
		}
	}
	return workflow[:at+1], jobs
}

// jobSteps returns each step of job, in order, as written.
func jobSteps(job string) []string {
	at := strings.Index(job, "\n    steps:\n")
	if at < 0 {
		return nil
	}
	body := job[at+len("\n    steps:\n"):]
	locs := regexp.MustCompile(`(?m)^      - `).FindAllStringIndex(body, -1)
	var steps []string
	for i, loc := range locs {
		end := len(body)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		steps = append(steps, body[loc[0]:end])
	}
	return steps
}

var yamlComment = regexp.MustCompile(`(^|\s)#.*$`)

// uncommented drops comments and blank lines, so a check reads what runs, not
// what a comment says about it. No value in release-tag.yml holds a "#"; a
// shell comment inside a run block goes too, which no check looks at.
func uncommented(text string) string {
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		if line = strings.TrimRight(yamlComment.ReplaceAllString(line, ""), " \t"); line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

// keyValues returns the value of every `key:` in text, whether the key opens a
// line (after any "- ") or follows "{" or "," in flow form, quoted or not.
func keyValues(text, key string) []string {
	re := regexp.MustCompile(`(?m)(?:^[ \t]*(?:- )?|[{,][ \t]*)["']?` + regexp.QuoteMeta(key) + `["']?[ \t]*:(?:[ \t]+|$)([^\n]*)`)
	var values []string
	for _, m := range re.FindAllStringSubmatch(text, -1) {
		values = append(values, strings.TrimSpace(m[1]))
	}
	return values
}

// trimmedLines returns text's lines without their indentation.
func trimmedLines(text string) []string {
	var lines []string
	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
		lines = append(lines, strings.TrimSpace(line))
	}
	return lines
}

// stepShell returns the shell a step runs, as GitHub hands it to bash: a
// `run: |` block without its indentation, or a one-line run.
func stepShell(t *testing.T, step string) string {
	t.Helper()
	lines := strings.Split(step, "\n")
	runKey := regexp.MustCompile(`^([ \t]*)(- )?run:[ \t]*(.*)$`)
	for i, line := range lines {
		m := runKey.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		if m[3] != "|" {
			return m[3] + "\n"
		}
		keyIndent := len(m[1]) + len(m[2])
		var block []string
		dedent := -1
		for _, l := range lines[i+1:] {
			if strings.TrimSpace(l) == "" {
				block = append(block, "")
				continue
			}
			indent := len(l) - len(strings.TrimLeft(l, " "))
			if indent <= keyIndent {
				break
			}
			if dedent < 0 {
				dedent = indent
			}
			block = append(block, l[min(dedent, indent):])
		}
		return strings.TrimRight(strings.Join(block, "\n"), "\n") + "\n"
	}
	t.Fatalf("this step runs no shell:\n%s", step)
	return ""
}

// runStepShell runs shell in dir as GitHub runs a step that names no `shell:`,
// with `bash -e`, and returns its combined output and its exit code.
func runStepShell(t *testing.T, dir, shell string, env ...string) (string, int) {
	t.Helper()
	script := filepath.Join(t.TempDir(), "step.sh")
	writeFileT(t, script, shell)
	cmd := exec.Command("bash", "-e", script)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if exitErr, ok := err.(*exec.ExitError); ok {
		return string(out), exitErr.ExitCode()
	} else if err != nil {
		t.Fatalf("running the step: %v", err)
	}
	return string(out), 0
}

// writeStub writes an executable shell script.
func writeStub(t *testing.T, path, script string) {
	t.Helper()
	writeFileT(t, path, "#!/bin/sh\n"+script)
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

// lastStep returns the named job's last step, the one the step tests run.
func lastStep(t *testing.T, jobs map[string]string, name string) string {
	t.Helper()
	steps := jobSteps(jobs[name])
	if len(steps) == 0 {
		t.Fatalf("the %s job has no steps", name)
	}
	return steps[len(steps)-1]
}

// TestReleaseTagWorkflowShape pins the properties of release-tag.yml that keep
// tagging on main and create-only, and that the step tests below rely on.
func TestReleaseTagWorkflowShape(t *testing.T) {
	top, jobs := readReleaseWorkflow(t)
	code := uncommented(top + "jobs:\n" + jobs["plan"] + jobs["tag"] + jobs["verify"])
	for name, job := range jobs {
		if name != "plan" && name != "tag" && name != "verify" {
			code += uncommented(job)
		}
	}
	topCode := uncommented(top)

	// U1 commits the floor unset, so its merge tags nothing; U9 sets it, and
	// changes this line with it. Nothing else may name the floor, or a job or a
	// step could set its own below the one committed here.
	if !strings.Contains(topCode, "\nenv:\n  TRELLIS_TAG_FLOOR: \"\"\n") {
		t.Error("the floor must be committed unset, in the workflow's top-level env, until the CLI's first release")
	}
	if n := strings.Count(code, "TRELLIS_TAG_FLOOR"); n != 1 {
		t.Errorf("only the top-level env may set the floor, but release-tag.yml names it %d times outside comments", n)
	}

	// The triggers: pushes to main and a dispatch, and nothing else.
	triggers := map[string]string{}
	trigger := ""
	if at := strings.Index(topCode, "\non:\n"); at >= 0 {
		for _, line := range strings.Split(topCode[at+len("\non:\n"):], "\n") {
			if !strings.HasPrefix(line, " ") {
				break
			}
			if m := regexp.MustCompile(`^  ([\w-]+):[ \t]*(.*)$`).FindStringSubmatch(line); m != nil {
				trigger = m[1]
				triggers[trigger] = m[2]
			} else {
				triggers[trigger] += strings.TrimSpace(line)
			}
		}
	}
	push := strings.NewReplacer(" ", "", `"`, "", "'", "").Replace(triggers["push"])
	if len(triggers) != 2 || (push != "branches:[main]" && push != "branches:-main") || triggers["workflow_dispatch"] != "" {
		t.Errorf("release-tag must run on pushes to main and on dispatch, and on nothing else; its triggers are %q", triggers)
	}

	// The token is read-only everywhere: a grant at any level would widen it, and
	// only the App's token writes.
	if !strings.Contains(topCode, "\npermissions:\n  contents: read\n") {
		t.Error("release-tag's token must be read-only; only the App's token writes")
	}
	lines := strings.Split(code, "\n")
	permissionsKey := regexp.MustCompile(`^([ \t]*)(?:- )?["']?permissions["']?[ \t]*:[ \t]*(.*)$`)
	for i, line := range lines {
		m := permissionsKey.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		grant := m[2]
		for _, below := range lines[i+1:] {
			if len(below)-len(strings.TrimLeft(below, " \t")) <= len(m[1]) {
				break
			}
			grant += " " + strings.TrimSpace(below)
		}
		if strings.Contains(grant, "write") {
			t.Errorf("only the App's token may write, but release-tag.yml grants %q", strings.TrimSpace(grant))
		}
	}
	if v := keyValues(uncommented(jobs["tag"]), "permissions"); len(v) != 1 || v[0] != "{}" {
		t.Errorf("the tag job's GITHUB_TOKEN must get no permissions (permissions: {}); it has %q", v)
	}

	// GitHub replaces a concurrency group's pending run with any newer one, so
	// a re-run of an old run could drop a push's own run and its version.
	if v := keyValues(code, "concurrency"); len(v) != 0 {
		t.Error("release-tag runs must not share a concurrency group, which can drop a push's run")
	}
	// The step tests run each shell as bash -e, which is GitHub's default.
	for _, key := range []string{"shell", "defaults"} {
		if v := keyValues(code, key); len(v) != 0 {
			t.Errorf("release-tag must run its steps in GitHub's default shell, but it sets %s", key)
		}
	}
	// A failed verify must fail the run.
	for _, v := range keyValues(code, "continue-on-error") {
		if v != "false" {
			t.Errorf("no release-tag step or job may continue on error, but one sets continue-on-error: %s", v)
		}
	}
	// plan's output decides what the App tags, and verify is the check that
	// fails, so no action may change under the same name.
	for _, v := range keyValues(code, "uses") {
		if !regexp.MustCompile(`^[\w./-]+@[0-9a-f]{40}$`).MatchString(v) {
			t.Errorf("every action in release-tag must be pinned by commit, but one uses %q", v)
		}
	}
	// The release environment holds the App's key, so only the tag job enters it.
	if v := keyValues(code, "environment"); len(v) != 1 || keyValues(uncommented(jobs["tag"]), "environment") == nil || v[0] != "release" {
		t.Errorf("only the tag job may use the release environment, which holds the App's key; release-tag.yml has environment %q", v)
	}
	for _, text := range []string{"secrets.", "steps.app.outputs.token"} {
		if n := strings.Count(code, text); n != 1 {
			t.Errorf("the App's key and token may each reach one step, but release-tag.yml names %s %d times", text, n)
		}
	}
	// Create only: nothing may move or delete a tag.
	for _, verb := range []string{`\bPATCH\b`, `DELETE`, `(^|\s)-X`, `--method`, `\bforce\b`, `\bgit\s+push\b`, `\bgraphql\b`} {
		if regexp.MustCompile(verb).MatchString(code) {
			t.Errorf("release-tag must only create tags, but it matches %s", verb)
		}
	}

	// Which job runs when. Every push before the CLI's first release has nothing
	// to tag, and the tag job's guard keeps those pushes out of the release
	// environment; verify runs after tag whatever tag did, so a failed tag job
	// still names what is missing. No step is skipped by a condition.
	for name, want := range map[string][2]string{
		"plan":   {"", ""},
		"tag":    {"plan", "needs.plan.outputs.tags != ''"},
		"verify": {"[plan, tag]", "always()"},
	} {
		job := uncommented(jobs[name])
		needs, ifs := keyValues(job, "needs"), keyValues(job, "if")
		if (want[0] == "" && needs != nil) || (want[0] != "" && (len(needs) != 1 || needs[0] != want[0])) {
			t.Errorf("the %s job must need %q, but it needs %q", name, want[0], needs)
		}
		if (want[1] == "" && ifs != nil) || (want[1] != "" && (len(ifs) != 1 || ifs[0] != want[1])) {
			t.Errorf("the %s job must run if %q, and no step of it may have a condition, but it has if %q", name, want[1], ifs)
		}
	}

	// What each job runs. plan and verify check out the full history of the
	// commit the run is for, then run the script; tag gets the App's token, then
	// creates the tags. The step tests below run each shell.
	checkout := regexp.MustCompile(`^- uses: actions/checkout@[0-9a-f]{40}\nwith:\nfetch-depth: 0$`)
	for _, name := range []string{"plan", "verify"} {
		steps := jobSteps(jobs[name])
		if len(steps) != 2 || !checkout.MatchString(strings.Join(trimmedLines(uncommented(steps[0])), "\n")) {
			t.Errorf("the %s job must check out the full history with a pinned actions/checkout and nothing else, then run one step", name)
		}
	}
	if !strings.Contains(uncommented(jobs["plan"]), "\n    outputs:\n      tags: ${{ steps.plan.outputs.tags }}\n") || !strings.HasPrefix(uncommented(lastStep(t, jobs, "plan")), "      - id: plan\n") {
		t.Error("the plan job's tags output must be what its plan step publishes")
	}
	steps := jobSteps(jobs["tag"])
	if len(steps) != 2 {
		t.Fatalf("the tag job must have two steps, the App's token and the create step; it has %d", len(steps))
	}
	app := trimmedLines(regexp.MustCompile(`@[0-9a-f]{40}\n`).ReplaceAllString(uncommented(steps[0]), "@<commit>\n"))
	app[0] = strings.TrimPrefix(app[0], "- ")
	sort.Strings(app)
	// Without permission-contents the App's token gets every permission the
	// installation holds, not only creating tags.
	wantApp := []string{
		"client-id: ${{ vars.RELEASE_APP_CLIENT_ID }}",
		"id: app",
		"permission-contents: write",
		"private-key: ${{ secrets.RELEASE_APP_PRIVATE_KEY }}",
		"uses: actions/create-github-app-token@<commit>",
		"with:",
	}
	if strings.Join(app, "\n") != strings.Join(wantApp, "\n") {
		t.Errorf("the App's token step must be a pinned create-github-app-token limited to contents: write, but it reads:\n%s", strings.Join(app, "\n"))
	}
	create := uncommented(steps[1])
	for _, line := range []string{"GH_TOKEN: ${{ steps.app.outputs.token }}", "TAGS: ${{ needs.plan.outputs.tags }}"} {
		if !strings.Contains(create, "\n          "+line+"\n") {
			t.Errorf("the create step must get %s", line)
		}
	}
}

// TestReleaseTagWorkflowPlanStepPublishesTheScriptsPlan runs the plan step's
// shell with a stand-in for scripts/release-tags.sh.
func TestReleaseTagWorkflowPlanStepPublishesTheScriptsPlan(t *testing.T) {
	_, jobs := readReleaseWorkflow(t)
	shell := stepShell(t, lastStep(t, jobs, "plan"))
	const commit1, commit2 = "1111111111111111111111111111111111111111", "2222222222222222222222222222222222222222"
	for _, tc := range []struct {
		name, out  string
		code       string
		wantOutput string
	}{
		{"two tags", "v0.25.0 " + commit1 + "\nv0.26.0 " + commit2 + "\n", "0", "tags<<RELEASE_TAGS_EOF\nv0.25.0 " + commit1 + "\nv0.26.0 " + commit2 + "\nRELEASE_TAGS_EOF\n"},
		{"nothing to tag", "", "0", "tags<<RELEASE_TAGS_EOF\nRELEASE_TAGS_EOF\n"},
		{"the script fails", "", "2", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			log, output := filepath.Join(dir, "calls.log"), filepath.Join(dir, "github-output")
			writeStub(t, filepath.Join(dir, "scripts", "release-tags.sh"), `printf '%s floor=%s\n' "$*" "${TRELLIS_TAG_FLOOR-unset}" >> "$STUB_LOG"
printf '%s' "$STUB_OUT"
exit "$STUB_CODE"
`)
			writeFileT(t, output, "")
			writeFileT(t, log, "")
			out, code := runStepShell(t, dir, shell, "STUB_LOG="+log, "STUB_OUT="+tc.out, "STUB_CODE="+tc.code, "GITHUB_OUTPUT="+output, "TRELLIS_TAG_FLOOR=0.25.0")
			if calls := readFileT(t, log); calls != "plan floor=0.25.0\n" {
				t.Errorf("the plan step must run scripts/release-tags.sh plan once, with the workflow's floor; it ran:\n%s", calls)
			}
			if tc.code != "0" {
				if code == 0 {
					t.Errorf("the plan step must fail when the script fails, but it exited 0:\n%s", out)
				}
				return
			}
			if code != 0 {
				t.Fatalf("the plan step exited %d:\n%s", code, out)
			}
			if got := readFileT(t, output); got != tc.wantOutput {
				t.Errorf("the plan step wrote %q to GITHUB_OUTPUT, want %q", got, tc.wantOutput)
			}
		})
	}
}

// TestReleaseTagWorkflowVerifyStepFailsWithTheScript runs the verify step's
// shell with a stand-in for scripts/release-tags.sh.
func TestReleaseTagWorkflowVerifyStepFailsWithTheScript(t *testing.T) {
	_, jobs := readReleaseWorkflow(t)
	shell := stepShell(t, lastStep(t, jobs, "verify"))
	for _, scriptCode := range []string{"0", "1"} {
		dir := t.TempDir()
		log := filepath.Join(dir, "calls.log")
		writeStub(t, filepath.Join(dir, "scripts", "release-tags.sh"), `printf '%s floor=%s\n' "$*" "${TRELLIS_TAG_FLOOR-unset}" >> "$STUB_LOG"
exit "$STUB_CODE"
`)
		writeFileT(t, log, "")
		out, code := runStepShell(t, dir, shell, "STUB_LOG="+log, "STUB_CODE="+scriptCode, "TRELLIS_TAG_FLOOR=0.25.0")
		if calls := readFileT(t, log); calls != "verify floor=0.25.0\n" {
			t.Errorf("the verify step must run scripts/release-tags.sh verify once, with the workflow's floor; it ran:\n%s", calls)
		}
		if (code == 0) != (scriptCode == "0") {
			t.Errorf("the verify step exited %d when the script exited %s:\n%s", code, scriptCode, out)
		}
	}
}

// TestReleaseTagWorkflowCreateStep runs the tag job's create step against a
// stand-in gh, which answers the create with the status each case names and
// the read-back with the commit the existing tag names. The run must create
// each tag with one POST, accept a 422 only when the tag already names the
// planned commit, fail on anything else, and call nothing but those two.
func TestReleaseTagWorkflowCreateStep(t *testing.T) {
	_, jobs := readReleaseWorkflow(t)
	shell := stepShell(t, lastStep(t, jobs, "tag"))
	const planned, other = "1111111111111111111111111111111111111111", "2222222222222222222222222222222222222222"
	const later = "3333333333333333333333333333333333333333"
	post := func(tag, commit string) string {
		return "gh api -i repos/kodhama/trellis/git/refs -f ref=refs/tags/" + tag + " -f sha=" + commit + "\n"
	}
	read := "gh api repos/kodhama/trellis/commits/refs/tags/v0.25.0 --jq .sha\n"
	for _, tc := range []struct {
		name, tags, create, existing string
		wantOK                       bool
		wantCalls                    string
	}{
		{"created", "v0.25.0 " + planned + "\n", "201", "", true, post("v0.25.0", planned)},
		{"every listed tag is created", "v0.25.0 " + planned + "\n\nv0.26.0 " + later + "\n", "201", "", true, post("v0.25.0", planned) + post("v0.26.0", later)},
		{"another run created it on the planned commit", "v0.25.0 " + planned + "\n", "422", planned, true, post("v0.25.0", planned) + read},
		{"the tag names another commit", "v0.25.0 " + planned + "\n", "422", other, false, post("v0.25.0", planned) + read},
		{"the read-back fails", "v0.25.0 " + planned + "\n", "422", "fail", false, post("v0.25.0", planned) + read},
		{"any other status", "v0.25.0 " + planned + "\n", "500", planned, false, post("v0.25.0", planned)},
		{"no status at all", "v0.25.0 " + planned + "\n", "none", planned, false, post("v0.25.0", planned)},
		{"a failure stops the tags after it", "v0.25.0 " + planned + "\nv0.26.0 " + later + "\n", "403", "", false, post("v0.25.0", planned)},
		{"a malformed plan line", "v0.25 " + planned + "\n", "201", "", false, ""},
		{"a short commit", "v0.25.0 1111111\n", "201", "", false, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir, bin := t.TempDir(), t.TempDir()
			log := filepath.Join(dir, "calls.log")
			writeFileT(t, log, "")
			writeStub(t, filepath.Join(bin, "gh"), `printf 'gh %s\n' "$*" >> "$STUB_LOG"
if [ "$2" = -i ]; then
  case "$STUB_CREATE" in
    201) printf 'HTTP/2.0 201 Created\nContent-Type: application/json\n\n{"ref":"%s"}\n' "$5"; exit 0 ;;
    none) echo 'error connecting to api.github.com' >&2; exit 1 ;;
    *) printf 'HTTP/2.0 %s Error\nContent-Type: application/json\n\n{"message":"error"}\n' "$STUB_CREATE"; echo "gh: error (HTTP $STUB_CREATE)" >&2; exit 1 ;;
  esac
fi
if [ "$STUB_EXISTING" = fail ]; then echo 'gh: Not Found (HTTP 404)' >&2; exit 1; fi
printf '%s\n' "$STUB_EXISTING"
`)
			writeStub(t, filepath.Join(bin, "git"), `printf 'git %s\n' "$*" >> "$STUB_LOG"
exit 1
`)
			out, code := runStepShell(t, dir, shell,
				"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
				"STUB_LOG="+log, "STUB_CREATE="+tc.create, "STUB_EXISTING="+tc.existing,
				"GH_TOKEN=app-token", "GITHUB_REPOSITORY=kodhama/trellis", "TAGS="+tc.tags)
			if (code == 0) != tc.wantOK {
				t.Errorf("the create step exited %d, want success %v:\n%s", code, tc.wantOK, out)
			}
			if calls := readFileT(t, log); calls != tc.wantCalls {
				t.Errorf("the create step made these calls:\n%s\nwant:\n%s", calls, tc.wantCalls)
			}
		})
	}
}
