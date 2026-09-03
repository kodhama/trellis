package main

// Tests for install.sh — the curl-path plugin vendor script (corrected design per
// spec-0005-curl-install-mechanical-vendoring, retired with `specs/` by
// `decision-0079`; the `spec-0005 AC#` markers below name requirements whose only
// surviving statement is these tests, with the retired text in git history. See the
// script's own header for why it supersedes the earlier attempt). Unlike that
// attempt's
// install.sh, this script makes exactly one decision (scope) and composes nothing
// else — so these tests check vendoring mechanics (fetch, verify, write, scope
// resolution), never the setup skill's decision logic (that lives in
// plugins/trellis/skills/setup/SKILL.md and is out of scope here). The harness shape
// (exec the script against throwaway dirs, TRELLIS_BUNDLE_SOURCE pointed at the
// vendored bundle so tests run offline) is salvaged from #128's
// cli/install_script_test.go. Upstream anchors:
//
//   - kodhama/trellis#124 (corrected design): the script vendors the WHOLE
//     plugins/trellis/ tree, verifies every byte before writing anything, resolves
//     project scope via `git rev-parse --show-toplevel` (never $PWD — the exact bug
//     class this design exists to avoid), never mutates git, and is idempotent.
//   - decision-0043 §4 (annotated in this same PR): this is a different, much
//     smaller artifact class than the retired end-user binary installer that used to
//     live at this path.
//   - TestInstallScriptBundleManifestIsCurrent is the pin-advance mechanism: it
//     regenerates the manifest from plugins/trellis/ on disk and fails whenever
//     install.sh's baked-in copy differs in content OR file set, so the two move
//     atomically on main (mirrors #128's TestInstallScriptPinIsCurrent, scoped to
//     the whole bundle instead of just the M1 payload).
//
// POST-GATE REVISION (spec-0005, NEEDS-REVISION verdict addressed here): the
// env var is $TRELLIS_SKILLS_SCOPE (was $TRELLIS_SCOPE); the ambiguous-scope,
// no-tty, no-git-repo case is a fail-closed hard error, never a silent fallback to
// personal scope (spec-0005 AC5 — TestVendorAmbiguousScopeNoTTYFailsClosed replaces
// the old, wrongly-asserting TestVendorDefaultFallsBackToPersonalOutsideGitRepo);
// AC9 (no git mutation, on every path — not just the happy one) and AC2 (zero
// decision logic, proven by instructions-file-content invariance, not just static
// grep) each get their own dedicated coverage below, and the AC10 project-fresh-
// install row now asserts all five §4 guidance items, not just the first.
//
// Real /dev/tty prompting (rustup-style, reading from the terminal even though
// stdin is consumed by the curl|sh pipe) is verified by hand with a real pty
// (`expect`), not by this Go suite — see the PR body for the transcripts. `go test`
// subprocesses have no controlling terminal in CI, so a Go-only test of that path
// would only prove "no tty -> no prompt", which the --non-interactive tests below
// already cover; it would not prove the prompt itself works.

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// --- helpers -------------------------------------------------------------------

func installScriptPath(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs("../install.sh")
	if err != nil {
		t.Fatalf("resolving install.sh path: %v", err)
	}
	return abs
}

// vendoredBundleDir is the plugin bundle install.sh vends — the whole tree
// (kodhama/trellis#117's vendoredPayloadDir in payload_test.go is its reference/
// subdirectory only).
const vendoredBundleDir = "../plugins/trellis"

func vendoredBundleAbs(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(vendoredBundleDir)
	if err != nil {
		t.Fatalf("resolving vendored bundle path: %v", err)
	}
	return abs
}

type vendorResult struct {
	stdout string
	stderr string
	code   int
}

// runVendor execs install.sh with cwd, HOME, and bundle-source overrides.
// --non-interactive is always passed: tests must behave identically in CI (no
// /dev/tty) and on a developer's machine (a live /dev/tty would otherwise turn a
// should-default-or-fail case into a hang waiting for input). home == "" leaves
// $HOME untouched (used by tests that only exercise project scope and never write
// under $HOME).
func runVendor(t *testing.T, dir, home, bundleSrc string, args ...string) vendorResult {
	t.Helper()
	all := append([]string{installScriptPath(t), "--non-interactive"}, args...)
	cmd := exec.Command("/bin/sh", all...)
	cmd.Dir = dir
	env := os.Environ()
	env = append(env, "TRELLIS_BUNDLE_SOURCE="+bundleSrc)
	if home != "" {
		env = append(env, "HOME="+home)
	}
	cmd.Env = env
	var so, se bytes.Buffer
	cmd.Stdout, cmd.Stderr = &so, &se
	err := cmd.Run()
	code := 0
	if err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("running install.sh: %v (stderr: %s)", err, se.String())
		}
		code = ee.ExitCode()
	}
	return vendorResult{stdout: so.String(), stderr: se.String(), code: code}
}

func readFileT(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(b)
}

func writeFileT(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	cmd := exec.Command("git", "init", "-q")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init in %s: %v: %s", dir, err, out)
	}
}

// walkFiles lists every regular file under dir, relative to dir, sorted.
func walkFiles(t *testing.T, dir string) []string {
	t.Helper()
	var rels []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			rels = append(rels, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", dir, err)
	}
	sort.Strings(rels)
	return rels
}

// snapshotTree maps relative path -> content for every regular file under dir.
// Returns an empty map (not an error) if dir does not exist yet.
func snapshotTree(t *testing.T, dir string) map[string]string {
	t.Helper()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return map[string]string{}
	}
	snap := map[string]string{}
	for _, rel := range walkFiles(t, dir) {
		snap[rel] = readFileT(t, filepath.Join(dir, rel))
	}
	return snap
}

// assertBundleVendored checks the one contract this script owes: every file
// vendored under targetTrellisDir is byte-identical to the corresponding file
// under the real plugins/trellis/, the file set matches exactly, and the
// executable bit on hooks/staleness.sh survived the copy.
func assertBundleVendored(t *testing.T, targetTrellisDir string) {
	t.Helper()
	bundle := vendoredBundleAbs(t)
	want := walkFiles(t, bundle)
	got := walkFiles(t, targetTrellisDir)
	if strings.Join(want, "\n") != strings.Join(got, "\n") {
		t.Fatalf("vendored file set differs from plugins/trellis/\nwant: %v\ngot:  %v", want, got)
	}
	for _, rel := range want {
		wantContent := readFileT(t, filepath.Join(bundle, rel))
		gotContent := readFileT(t, filepath.Join(targetTrellisDir, rel))
		if gotContent != wantContent {
			t.Errorf("%s is not byte-identical to the vendored plugins/trellis/%s", rel, rel)
		}
	}
	info, err := os.Stat(filepath.Join(targetTrellisDir, "hooks", "staleness.sh"))
	if err != nil {
		t.Fatalf("stat hooks/staleness.sh: %v", err)
	}
	if info.Mode()&0o111 == 0 {
		t.Errorf("hooks/staleness.sh lost its executable bit when vendored")
	}
}

// --- the bundle-manifest advance guard ------------------------------------------

var bundleManifestHeredocRe = regexp.MustCompile(`(?s)<<'TRELLIS_BUNDLE_MANIFEST'\n(.*?)\nTRELLIS_BUNDLE_MANIFEST\n`)

// TestInstallScriptBundleManifestIsCurrent is the pin-advance mechanism (#124: "the
// script itself is versioned/pinned and checksummed like any writer artifact",
// adapted from #128's TestInstallScriptPinIsCurrent). install.sh's baked-in
// TRELLIS_BUNDLE_MANIFEST must always equal the sha256 of every file actually under
// plugins/trellis/ — both content and file set. Because this fails on any bundle
// change that does not also update install.sh, the manifest advances in the same
// commit that changes the bundle — script and bundle move atomically on main.
func TestInstallScriptBundleManifestIsCurrent(t *testing.T) {
	script := readFileT(t, installScriptPath(t))
	m := bundleManifestHeredocRe.FindStringSubmatch(script)
	if m == nil {
		t.Fatal("install.sh must bake a TRELLIS_BUNDLE_MANIFEST heredoc (<<'TRELLIS_BUNDLE_MANIFEST' ... TRELLIS_BUNDLE_MANIFEST)")
	}
	lines := strings.Split(m[1], "\n")
	got := map[string]string{} // relpath -> sha256
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			t.Fatalf("malformed manifest line %q — expected \"<sha256>  <relpath>\"", line)
		}
		got[fields[1]] = fields[0]
	}

	bundle := vendoredBundleAbs(t)
	want := map[string]string{}
	for _, rel := range walkFiles(t, bundle) {
		content := readFileT(t, filepath.Join(bundle, rel))
		want[rel] = fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
	}

	for rel, wantHash := range want {
		gotHash, ok := got[rel]
		if !ok {
			t.Errorf("install.sh's manifest is missing %s (present in plugins/trellis/) — advance the manifest in this same commit", rel)
			continue
		}
		if gotHash != wantHash {
			t.Errorf("install.sh's manifest for %s is stale: baked-in %s, actual %s — advance the manifest in this same commit", rel, gotHash, wantHash)
		}
	}
	for rel := range got {
		if _, ok := want[rel]; !ok {
			t.Errorf("install.sh's manifest names %s, which no longer exists under plugins/trellis/ — trim the manifest in this same commit", rel)
		}
	}
}

// --- fresh installs --------------------------------------------------------------

// TestVendorPersonalScopeFreshInstall (#124: personal scope needs no git repo and
// writes to $HOME/.claude/skills/trellis). Extended per spec-0005's test-coverage
// table (personal fresh-vendor row, AC1/AC4/AC10): stdout must carry the next-step
// pointer to .trellis/rules.toml (it named /trellis:setup until decision-0072 retired
// the skill) and must NOT carry the project-only trust-dialog note.
//
// The trailing line-count check (kodhama/trellis#132) closes a real gap: AC10 requires
// "exactly the five §4 items… and nothing more", but for personal scope only items 1 and
// 5 apply (items 2-4 are project-scope only) — and the prior version of this test only
// asserted the *absence* of the two project-only strings, which let an unenumerated
// extra line ("Personal scope needs no trust dialog…") ship undetected, since any other
// added line would still pass every Contains check above. Pinning the exact stdout line
// count means any future addition — under any wording — fails loudly instead of drifting
// silently.
func TestVendorPersonalScopeFreshInstall(t *testing.T) {
	cwd := t.TempDir() // deliberately NOT a git repo — personal scope must not care
	home := t.TempDir()

	res := runVendor(t, cwd, home, vendoredBundleAbs(t), "--scope", "personal")
	if res.code != 0 {
		t.Fatalf("expected success, got exit %d\nstdout: %s\nstderr: %s", res.code, res.stdout, res.stderr)
	}
	assertBundleVendored(t, filepath.Join(home, ".claude", "skills", "trellis"))
	if !strings.Contains(res.stdout, "scope: personal") {
		t.Errorf("stdout should say which scope was chosen; got:\n%s", res.stdout)
	}
	// decision-0072 retired /trellis:setup; the next step is the file itself.
	if !strings.Contains(res.stdout, ".trellis/rules.toml") {
		t.Errorf("stdout should carry the next-step pointer to .trellis/rules.toml; got:\n%s", res.stdout)
	}
	if strings.Contains(res.stdout, "/trellis:setup") {
		t.Errorf("stdout still points at the setup skill, retired by decision-0072; got:\n%s", res.stdout)
	}
	if strings.Contains(res.stdout, "trust-dialog") || strings.Contains(res.stdout, "workspace-trust dialog") {
		t.Errorf("personal scope must NOT print the project-only trust-dialog note; got:\n%s", res.stdout)
	}
	if strings.Contains(res.stdout, "git add .claude/skills/trellis") {
		t.Errorf("personal scope must NOT print the project-only commit suggestion; got:\n%s", res.stdout)
	}

	lines := strings.Split(strings.TrimRight(res.stdout, "\n"), "\n")
	const wantLines = 8 // item 1 (scope + vendored-to + files-written + rules = 4
	// lines) + blank separator + item 5 (3-line next-step pointer) = 8; items 2-4
	// never fire for personal scope. See install.sh's post-write block (guarded by
	// `[ "$scope" = "project" ]`) for the source of this count.
	//
	// 7 -> 8 per decision-0068 D1: item 1 now names what was rendered, and on
	// personal scope that is "no rules file (project scope only)". The line is
	// deliberate, not incidental — silently delivering nothing is the exact defect
	// this change fixes, so the path that still delivers nothing has to say so.
	if len(lines) != wantLines {
		t.Errorf("personal-scope stdout has %d lines, want exactly %d (spec-0005 AC10 — items 1 and 5 only, nothing more); got:\n%s", len(lines), wantLines, res.stdout)
	}
}

// TestVendorProjectScopeFreshInstallFromRoot (#124: project scope resolves to
// <repo-root>/.claude/skills/trellis when run from the root itself). Extended per
// spec-0005's test-coverage table ("Project fresh vendor, run from repo root" row,
// AC1/AC3/AC4/AC9/AC10): asserts all five of §4's post-write guidance items in
// order (scope/path/stamp, the trust-dialog note, the no-walk-up caveat, the commit
// suggestion, and the next-step pointer), and confirms the commit suggestion is only
// ever printed — never executed — by checking the target repo's own git status
// afterward.
func TestVendorProjectScopeFreshInstallFromRoot(t *testing.T) {
	repo := t.TempDir()
	initGitRepo(t, repo)

	res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
	if res.code != 0 {
		t.Fatalf("expected success, got exit %d\nstdout: %s\nstderr: %s", res.code, res.stdout, res.stderr)
	}
	assertBundleVendored(t, filepath.Join(repo, ".claude", "skills", "trellis"))

	target := filepath.Join(repo, ".claude", "skills", "trellis")
	// item 1: scope, target path, bundle stamp.
	if !strings.Contains(res.stdout, "scope: project") {
		t.Errorf("item 1 (scope): stdout missing 'scope: project'; got:\n%s", res.stdout)
	}
	if !strings.Contains(res.stdout, target) {
		t.Errorf("item 1 (path): stdout missing the resolved target path %s; got:\n%s", target, res.stdout)
	}
	if !strings.Contains(res.stdout, "payload@") {
		t.Errorf("item 1 (stamp): stdout missing the bundle stamp; got:\n%s", res.stdout)
	}
	// item 2: the trust-dialog note (project scope only).
	if !strings.Contains(res.stdout, "workspace-trust dialog") {
		t.Errorf("item 2 (trust dialog): stdout missing the workspace-trust-dialog note; got:\n%s", res.stdout)
	}
	// item 3: the no-walk-up caveat.
	if !strings.Contains(res.stdout, "do NOT walk up to the repo root") {
		t.Errorf("item 3 (no-walk-up): stdout missing the no-walk-up caveat; got:\n%s", res.stdout)
	}
	// item 4: the commit suggestion is present in output...
	if !strings.Contains(res.stdout, "add .claude/skills/trellis") || !strings.Contains(res.stdout, "commit -m") {
		t.Errorf("item 4 (commit suggestion): stdout missing the suggested git add/commit line; got:\n%s", res.stdout)
	}
	// ...and confirm via git status that the script itself made no staged/committed
	// change — the suggestion is printed, never executed.
	cmd := exec.Command("git", "status", "--porcelain=v1")
	cmd.Dir = repo
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git status: %v", err)
	}
	status := string(out)
	if !strings.Contains(status, "?? .claude/") {
		t.Errorf("item 4 (no mutation): expected the vendored files to show as untracked, got status:\n%s", status)
	}
	if strings.Contains(status, "A  ") {
		t.Errorf("item 4 (no mutation): nothing should be staged — install.sh must never run git add; status:\n%s", status)
	}
	// item 5: the next-step pointer. decision-0072 retired /trellis:setup, so the
	// pointer is now the file the consumer edits.
	if !strings.Contains(res.stdout, ".trellis/rules.toml") {
		t.Errorf("item 5 (next step): stdout missing the .trellis/rules.toml pointer; got:\n%s", res.stdout)
	}
	if strings.Contains(res.stdout, "/trellis:setup") {
		t.Errorf("item 5: stdout still points at the setup skill, retired by decision-0072; got:\n%s", res.stdout)
	}
}

// TestVendorProjectScopeFromSubdirectoryResolvesToRoot (#124's central bug class:
// the corrected design exists specifically because a script that resolved the
// target via $PWD instead of `git rev-parse --show-toplevel` would silently vendor
// the plugin somewhere Claude Code's skills-directory loader — which does NOT walk
// up to the repo root for project-scope plugins — would never find it). Default
// (no --scope flag) resolution, run from three levels deep, must still land the
// plugin at the true repo root, and must NOT also write anything under the
// subdirectory.
func TestVendorProjectScopeFromSubdirectoryResolvesToRoot(t *testing.T) {
	repo := t.TempDir()
	initGitRepo(t, repo)
	sub := filepath.Join(repo, "deep", "nested", "dir")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", sub, err)
	}

	res := runVendor(t, sub, "", vendoredBundleAbs(t)) // no --scope: default resolution
	if res.code != 0 {
		t.Fatalf("expected success, got exit %d\nstdout: %s\nstderr: %s", res.code, res.stdout, res.stderr)
	}
	assertBundleVendored(t, filepath.Join(repo, ".claude", "skills", "trellis"))
	if _, err := os.Stat(filepath.Join(sub, ".claude")); !os.IsNotExist(err) {
		t.Errorf(".claude must not be written inside the subdirectory %s — it must resolve to the repo root", sub)
	}
	if !strings.Contains(res.stdout, "scope: project") {
		t.Errorf("stdout should report project scope was chosen; got:\n%s", res.stdout)
	}
}

// TestVendorAmbiguousScopeNoTTYFailsClosed (spec-0005 AC5 — replaces an earlier,
// wrong reading of the original issue brief that asserted a silent fallback to
// personal scope here; that was flagged as a real conformance failure in gate
// review). Outside a git repo, with no --scope/$TRELLIS_SKILLS_SCOPE override and no
// controlling tty, project scope has no target and there is no one to ask: the
// script must exit non-zero immediately, name exactly what's missing, and write
// nothing — never silently substitute the *other* scope than the one implied by the
// (absent) request. This is the exact scenario spec-0005's test-coverage table row
// "No controlling tty, scope ambiguous (no git repo, no flag/env)" requires (AC5).
func TestVendorAmbiguousScopeNoTTYFailsClosed(t *testing.T) {
	cwd := t.TempDir() // not a git repo
	home := t.TempDir()

	res := runVendor(t, cwd, home, vendoredBundleAbs(t)) // no --scope, --non-interactive (no tty)
	if res.code == 0 {
		t.Fatalf("expected fail-closed (non-zero exit) when scope is ambiguous and no tty is available; got exit 0\nstdout: %s", res.stdout)
	}
	if !strings.Contains(res.stderr, "git repository") {
		t.Errorf("failure must name the missing git repository; stderr:\n%s", res.stderr)
	}
	if !strings.Contains(res.stderr, "controlling terminal") {
		t.Errorf("failure must name the missing controlling terminal; stderr:\n%s", res.stderr)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude")); !os.IsNotExist(err) {
		t.Errorf("nothing may be written on a fail-closed ambiguous scope, but %s/.claude exists (personal scope was silently substituted)", home)
	}
	if _, err := os.Stat(filepath.Join(cwd, ".claude")); !os.IsNotExist(err) {
		t.Errorf("nothing may be written on a fail-closed ambiguous scope, but %s/.claude exists", cwd)
	}
}

// TestVendorExplicitProjectScopeOutsideRepoFailsLoudly (#124: an explicit request
// for something the environment cannot provide is a hard failure, never a silent
// override — distinct from the no-request default-fallback case above).
func TestVendorExplicitProjectScopeOutsideRepoFailsLoudly(t *testing.T) {
	cwd := t.TempDir() // not a git repo

	res := runVendor(t, cwd, "", vendoredBundleAbs(t), "--scope", "project")
	if res.code == 0 {
		t.Fatal("expected failure when project scope is explicitly requested outside a git repo")
	}
	if !strings.Contains(res.stderr, "git repository") {
		t.Errorf("failure should name the git-repo requirement; stderr:\n%s", res.stderr)
	}
	if _, err := os.Stat(filepath.Join(cwd, ".claude")); !os.IsNotExist(err) {
		t.Errorf("nothing may be written on a failed explicit request, but .claude exists")
	}
}

// --- scope selection: flag vs env, precedence, validation ------------------------

// TestVendorScopeFromEnvVar ($TRELLIS_SKILLS_SCOPE is honored when no --scope flag
// is given — spec-0005 §2; renamed from $TRELLIS_SCOPE per gate review).
func TestVendorScopeFromEnvVar(t *testing.T) {
	cwd := t.TempDir()
	home := t.TempDir()

	cmd := exec.Command("/bin/sh", installScriptPath(t), "--non-interactive")
	cmd.Dir = cwd
	cmd.Env = append(os.Environ(),
		"TRELLIS_BUNDLE_SOURCE="+vendoredBundleAbs(t),
		"HOME="+home,
		"TRELLIS_SKILLS_SCOPE=personal",
	)
	var so, se bytes.Buffer
	cmd.Stdout, cmd.Stderr = &so, &se
	if err := cmd.Run(); err != nil {
		t.Fatalf("expected success: %v (stderr: %s)", err, se.String())
	}
	assertBundleVendored(t, filepath.Join(home, ".claude", "skills", "trellis"))
	if !strings.Contains(so.String(), "$TRELLIS_SKILLS_SCOPE") {
		t.Errorf("stdout should attribute the scope to $TRELLIS_SKILLS_SCOPE; got:\n%s", so.String())
	}
}

// TestVendorScopeFlagWinsOverEnv (#124 assumption: with no hand-owned declaration
// file in play here — unlike the setup skill's expression.md — flag beats env by
// simple precedence, not by conflict error).
func TestVendorScopeFlagWinsOverEnv(t *testing.T) {
	repo := t.TempDir()
	initGitRepo(t, repo)

	cmd := exec.Command("/bin/sh", installScriptPath(t), "--non-interactive", "--scope", "project")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(),
		"TRELLIS_BUNDLE_SOURCE="+vendoredBundleAbs(t),
		"TRELLIS_SKILLS_SCOPE=personal",
	)
	var so, se bytes.Buffer
	cmd.Stdout, cmd.Stderr = &so, &se
	if err := cmd.Run(); err != nil {
		t.Fatalf("expected success: %v (stderr: %s)", err, se.String())
	}
	assertBundleVendored(t, filepath.Join(repo, ".claude", "skills", "trellis"))
}

// TestVendorInvalidScopeFails (fails fast on a bad value, before any network fetch).
func TestVendorInvalidScopeFails(t *testing.T) {
	cwd := t.TempDir()
	res := runVendor(t, cwd, "", vendoredBundleAbs(t), "--scope", "nowhere")
	if res.code == 0 {
		t.Fatal("expected failure on an invalid --scope value")
	}
	if !strings.Contains(res.stderr, "personal or project") {
		t.Errorf("failure should name the valid values; stderr:\n%s", res.stderr)
	}
}

// --- idempotency -------------------------------------------------------------------

// TestVendorReRunIsIdempotent (#124: a deterministic artifact is safe to re-vend —
// every byte on disk after a second run must equal the first).
// TestVendorUpgradeRemovesFilesThatLeftTheBundle: a Codex P1 on #227. The write
// phase used to copy the manifest's files over the existing target, which only
// ever creates and overwrites — so a file that LEFT the bundle survived every
// upgrade. Concretely: decision-0072 deleted skills/setup/, but an existing curl
// install kept the directory, and Claude Code discovers skills from the
// directory, so /trellis:setup stayed live for exactly the consumers who could
// not be told it was retired. The bundle is now swapped in whole.
//
// The stale path here is a REAL retired one, not an invented marker: it is the
// file this PR deletes, so this test fails against the shipped-before state for
// the same reason a consumer would have hit it.
func TestVendorUpgradeRemovesFilesThatLeftTheBundle(t *testing.T) {
	repo := t.TempDir()
	initGitRepo(t, repo)

	if res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project"); res.code != 0 {
		t.Fatalf("first run failed: %s", res.stderr)
	}
	target := filepath.Join(repo, ".claude", "skills", "trellis")

	// Simulate an install made before the retirement: the skill directory is on
	// disk and is not in the current manifest.
	stale := filepath.Join(target, "skills", "setup")
	if err := os.MkdirAll(stale, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stale, "SKILL.md"), []byte("---"+"\n"+"name: setup"+"\n"+"---"+"\n"+"retired"+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project"); res.code != 0 {
		t.Fatalf("upgrade run failed: %s", res.stderr)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("skills/setup survived the upgrade (err=%v) — a retired skill stays discoverable "+
			"as /trellis:setup for every consumer who installed before it was removed", err)
	}
	// The swap must not cost the files that ARE in the bundle.
	assertBundleVendored(t, target)
	for _, leftover := range []string{target + ".new", target + ".old"} {
		matches, _ := filepath.Glob(leftover + "*")
		if len(matches) != 0 {
			t.Errorf("the swap left scratch directories behind: %v", matches)
		}
	}
}

func TestVendorReRunIsIdempotent(t *testing.T) {
	repo := t.TempDir()
	initGitRepo(t, repo)

	if res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project"); res.code != 0 {
		t.Fatalf("first run failed: %s", res.stderr)
	}
	before := snapshotTree(t, filepath.Join(repo, ".claude", "skills", "trellis"))

	res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
	if res.code != 0 {
		t.Fatalf("second run failed (exit %d): %s", res.code, res.stderr)
	}
	after := snapshotTree(t, filepath.Join(repo, ".claude", "skills", "trellis"))
	if len(before) != len(after) {
		t.Fatalf("re-run changed the file set: %d files before, %d after", len(before), len(after))
	}
	for path, want := range before {
		if got, ok := after[path]; !ok || got != want {
			t.Errorf("re-run changed %s", path)
		}
	}
}

// --- verification failures (kodhama-0007 rule 3's "data, not trust" ethos, applied
// to this script's own bundle manifest) ---------------------------------------------

// TestVendorCorruptedFetchFailsClosedNoPartialWrite: a bundle file that does not
// match install.sh's baked-in manifest aborts before anything is written to the
// target directory at all — not even a partial .claude tree.
func TestVendorCorruptedFetchFailsClosedNoPartialWrite(t *testing.T) {
	tamperedSrc := t.TempDir()
	if err := copyDirT(t, vendoredBundleAbs(t), tamperedSrc); err != nil {
		t.Fatalf("copying bundle to tamper: %v", err)
	}
	victim := filepath.Join(tamperedSrc, "reference", "invariants.md")
	writeFileT(t, victim, readFileT(t, victim)+"tampered\n")

	repo := t.TempDir()
	initGitRepo(t, repo)

	res := runVendor(t, repo, "", tamperedSrc, "--scope", "project")
	if res.code == 0 {
		t.Fatal("expected failure on a bundle file that does not match the baked-in manifest")
	}
	if !strings.Contains(res.stderr, "checksum") {
		t.Errorf("failure must name the checksum check; stderr:\n%s", res.stderr)
	}
	if _, err := os.Stat(filepath.Join(repo, ".claude")); !os.IsNotExist(err) {
		t.Errorf("nothing may be installed on verification failure, but .claude exists")
	}
}

// copyDirT recursively copies src to dst (both must exist/be creatable); used only
// to build a scratch bundle source the test can tamper with without touching the
// real plugins/trellis/.
func copyDirT(t *testing.T, src, dst string) error {
	t.Helper()
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		writeFileT(t, target, readFileT(t, path))
		return nil
	})
}

// --- --non-interactive: the no-tty path, forced explicitly ------------------------

// TestVendorNonInteractiveFlagAppliesDefaultWithoutPrompting (#124: "no-tty
// non-interactive path via flag"). --non-interactive must produce the same
// deterministic default as the ambient no-tty case (go test subprocesses have no
// controlling terminal already), and the run must never block waiting on input —
// exercised here by the mere fact that runVendor's exec.Cmd.Run() returns at all
// under Go's test timeout. The real proof that --non-interactive overrides a
// genuinely *available* tty (not just an absent one) is done by hand with a pty;
// see the PR body.
func TestVendorNonInteractiveFlagAppliesDefaultWithoutPrompting(t *testing.T) {
	repo := t.TempDir()
	initGitRepo(t, repo)

	res := runVendor(t, repo, "", vendoredBundleAbs(t)) // --non-interactive, no --scope
	if res.code != 0 {
		t.Fatalf("expected success, got exit %d\nstdout: %s\nstderr: %s", res.code, res.stdout, res.stderr)
	}
	if strings.Contains(res.stdout, "Vendor the Trellis plugin at which scope?") {
		t.Errorf("--non-interactive must never print the interactive prompt; stdout:\n%s", res.stdout)
	}
	assertBundleVendored(t, filepath.Join(repo, ".claude", "skills", "trellis"))
}

// --- AC2: zero decision logic, proven by instructions-file-content invariance -----
//
// A prose grep for `trellis:begin`/`expression.md`/etc. only proves the script
// doesn't *mention* those strings — it can't prove the script doesn't *branch* on
// instructions-file presence or content. This test proves the stronger property
// spec-0005's AC2 actually requires: two otherwise-identical repos that differ only
// in which instructions file they carry (and whether `.trellis/` exists at all)
// produce byte-identical vendoring output, and neither repo's own files are ever
// read-and-rewritten (or read-and-left-alone-by-luck) by the script.

// TestVendorZeroDecisionLogicAcrossInstructionFileVariants (spec-0005 AC2, test-
// coverage table's two-fixture-repo row). Fixture A carries a CLAUDE.md with
// trellis:begin/trellis:end managed-block markers plus a .trellis/expression.md
// declaring a posture (exactly the shape /trellis:setup would have left behind).
// Fixture B carries an AGENTS.md instead, and no .trellis/ at all. A script that
// branched on either — under any name — would produce different stdout (beyond the
// target path) or would touch one repo's own files; this asserts neither happens.
func TestVendorZeroDecisionLogicAcrossInstructionFileVariants(t *testing.T) {
	repoA := t.TempDir()
	initGitRepo(t, repoA)
	// Fixture A used to carry a `trellis:begin` managed block. That block is now
	// a STATIC-DELIVERY CONFLICT signal (decision-0068, AC2's amendment): the
	// installer must refuse to render over it, so stdout legitimately differs.
	// Keeping it here would make this test assert the opposite of the contract.
	//
	// What this test still guards — and what AC2's "zero decision logic" heading
	// still means — is that the script never branches on POSTURE, STYLE, or
	// instructions-file CONTENT. Fixture A keeps the posture bait and the
	// hand-authored prose; only the conflict marker moved out, to
	// TestVendorRefusesToRenderOverAVendoredOverlay where it is asserted directly.
	claudeMD := "# Project A\n\nHand-authored prose a decision-logic script might try to detect or patch.\n"
	writeFileT(t, filepath.Join(repoA, "CLAUDE.md"), claudeMD)
	expressionMD := "---\nprofile: b\n---\n\nOur hand-authored expression — a decision-logic script might try to read this posture.\n"
	writeFileT(t, filepath.Join(repoA, ".trellis", "expression.md"), expressionMD)

	repoB := t.TempDir()
	initGitRepo(t, repoB)
	agentsMD := "# Project B — no trellis markers, no .trellis/ at all\n"
	writeFileT(t, filepath.Join(repoB, "AGENTS.md"), agentsMD)

	resA := runVendor(t, repoA, "", vendoredBundleAbs(t), "--scope", "project")
	if resA.code != 0 {
		t.Fatalf("fixture A run failed (exit %d): %s", resA.code, resA.stderr)
	}
	resB := runVendor(t, repoB, "", vendoredBundleAbs(t), "--scope", "project")
	if resB.code != 0 {
		t.Fatalf("fixture B run failed (exit %d): %s", resB.code, resB.stderr)
	}

	// stdout must be byte-identical once the one legitimate scope-resolution input
	// (the absolute repo path) is normalized away — nothing else may differ.
	normA := strings.ReplaceAll(strings.ReplaceAll(resA.stdout, repoA, "<REPO>"), filepath.Join(repoA, ".claude", "skills", "trellis"), "<REPO>/.claude/skills/trellis")
	normB := strings.ReplaceAll(strings.ReplaceAll(resB.stdout, repoB, "<REPO>"), filepath.Join(repoB, ".claude", "skills", "trellis"), "<REPO>/.claude/skills/trellis")
	if normA != normB {
		t.Errorf("stdout differs between the two fixtures after normalizing the repo path — install.sh is branching on instructions-file presence/content:\nfixture A:\n%s\nfixture B:\n%s", normA, normB)
	}

	assertBundleVendored(t, filepath.Join(repoA, ".claude", "skills", "trellis"))
	assertBundleVendored(t, filepath.Join(repoB, ".claude", "skills", "trellis"))

	// Fixture A's own files: byte-identical before and after — not read-and-
	// rewritten, not read-and-left-alone-by-luck.
	if got := readFileT(t, filepath.Join(repoA, "CLAUDE.md")); got != claudeMD {
		t.Errorf("CLAUDE.md was modified by install.sh — it must never read or write any instructions file:\nwant:\n%s\ngot:\n%s", claudeMD, got)
	}
	if got := readFileT(t, filepath.Join(repoA, ".trellis", "expression.md")); got != expressionMD {
		t.Errorf(".trellis/expression.md was modified by install.sh — it must never touch .trellis/:\nwant:\n%s\ngot:\n%s", expressionMD, got)
	}

	// Fixture B: AGENTS.md untouched, and .trellis/ now contains EXACTLY the seed
	// and nothing else.
	//
	// This assertion inverted with decision-0070 D2. It used to require that
	// install.sh never created .trellis/ at all; the script now seeds
	// .trellis/rules.toml there, because running an installer inside a repository
	// is the adoption act, and without the rows the curl path shipped all
	// the whole rule set into context while only two of them applied. What AC2 still
	// forbids — and what this now checks — is anything BEYOND that one file: no
	// overlay, no internal/, no posture detection.
	seeded := filepath.Join(repoB, ".trellis", "rules.toml")
	if _, err := os.Stat(seeded); err != nil {
		t.Errorf("decision-0070 D2: install.sh must seed .trellis/rules.toml on project scope; got %v", err)
	}
	if got := readFileT(t, seeded); got != readFileT(t, filepath.Join(vendoredBundleAbs(t), "reference", "rules-b.toml")) {
		t.Errorf(".trellis/rules.toml must be the shipped rules-b.toml byte for byte — a composed or edited seed would be decision logic, which AC2 still forbids")
	}
	entries, err := os.ReadDir(filepath.Join(repoB, ".trellis"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "rules.toml" {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf(".trellis/ must contain only the seeded rules.toml; got %v — anything else is the vendored overlay decision-0065 forbids", names)
	}
	if got := readFileT(t, filepath.Join(repoB, "AGENTS.md")); got != agentsMD {
		t.Errorf("AGENTS.md was modified by install.sh — it must never read or write any instructions file")
	}
}

// --- AC9: no git mutation, ever — on every path, not just the happy one -----------

// gitInvocationShim writes a fake `git` onto a fresh directory's PATH that logs
// every invocation's argument line to logPath and then execs the real git (found
// via the ambient PATH before the shim is prepended) — so the script under test
// still gets correct git behavior, but every call it makes is recorded. Returns the
// directory to prepend to PATH and the log file path.
func gitInvocationShim(t *testing.T) (binDir, logPath string) {
	t.Helper()
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatalf("git not found on PATH: %v", err)
	}
	binDir = t.TempDir()
	logPath = filepath.Join(binDir, "git-invocations.log")
	script := "#!/bin/sh\n" +
		"printf '%s\\n' \"$*\" >> " + shQuote(logPath) + "\n" +
		"exec " + shQuote(realGit) + " \"$@\"\n"
	if err := os.WriteFile(filepath.Join(binDir, "git"), []byte(script), 0o755); err != nil {
		t.Fatalf("writing git shim: %v", err)
	}
	return binDir, logPath
}

// shQuote wraps s in single quotes for embedding in a generated POSIX sh script.
func shQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// runVendorWithPATH is runVendor, plus an extra directory prepended to PATH (used
// to put the git invocation shim ahead of the real git).
func runVendorWithPATH(t *testing.T, dir, home, bundleSrc, extraPathDir string, args ...string) vendorResult {
	t.Helper()
	all := append([]string{installScriptPath(t), "--non-interactive"}, args...)
	cmd := exec.Command("/bin/sh", all...)
	cmd.Dir = dir
	env := os.Environ()
	env = append(env, "TRELLIS_BUNDLE_SOURCE="+bundleSrc)
	if home != "" {
		env = append(env, "HOME="+home)
	}
	if extraPathDir != "" {
		env = append(env, "PATH="+extraPathDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	}
	cmd.Env = env
	var so, se bytes.Buffer
	cmd.Stdout, cmd.Stderr = &so, &se
	err := cmd.Run()
	code := 0
	if err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("running install.sh: %v (stderr: %s)", err, se.String())
		}
		code = ee.ExitCode()
	}
	return vendorResult{stdout: so.String(), stderr: se.String(), code: code}
}

// assertOnlyRevParseShowToplevel reads the shim's invocation log (if any — an
// absent log means git was never invoked at all, which trivially satisfies "only
// rev-parse --show-toplevel calls") and fails if any logged invocation is anything
// other than exactly `rev-parse --show-toplevel`.
func assertOnlyRevParseShowToplevel(t *testing.T, logPath string) {
	t.Helper()
	data, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		t.Fatalf("reading git invocation log: %v", err)
	}
	trimmed := strings.TrimRight(string(data), "\n")
	if trimmed == "" {
		return
	}
	for _, line := range strings.Split(trimmed, "\n") {
		if line != "rev-parse --show-toplevel" {
			t.Errorf("unexpected git invocation logged: %q — only 'rev-parse --show-toplevel' is ever permitted (spec-0005 AC9)", line)
		}
	}
}

// TestVendorNeverInvokesGitBeyondRevParse (spec-0005 AC9, test-coverage table's
// git-shim row): every scope/error path — personal, project-from-root, project-
// from-subdirectory, the AC5 ambiguous-no-tty fail-closed path, a tampered-fetch
// fail-closed path, an invalid --scope value, and a re-run — is run with the
// logging git shim on PATH; the invocation log must contain only read-only
// `rev-parse --show-toplevel` calls, on the success paths and the failure paths
// alike. Supersedes TestVendorNeverRunsGitAdd (folded into
// TestVendorProjectScopeFreshInstallFromRoot's item-4 assertion for the happy path;
// this test is the comprehensive, cross-path replacement gate review required).
func TestVendorNeverInvokesGitBeyondRevParse(t *testing.T) {
	t.Run("personal_explicit_never_invokes_git_at_all", func(t *testing.T) {
		cwd := t.TempDir() // not a repo — proves personal scope needs no git call
		home := t.TempDir()
		binDir, logPath := gitInvocationShim(t)
		res := runVendorWithPATH(t, cwd, home, vendoredBundleAbs(t), binDir, "--scope", "personal")
		if res.code != 0 {
			t.Fatalf("expected success: %s", res.stderr)
		}
		assertOnlyRevParseShowToplevel(t, logPath)
	})

	t.Run("project_from_root", func(t *testing.T) {
		repo := t.TempDir()
		initGitRepo(t, repo)
		binDir, logPath := gitInvocationShim(t)
		res := runVendorWithPATH(t, repo, "", vendoredBundleAbs(t), binDir, "--scope", "project")
		if res.code != 0 {
			t.Fatalf("expected success: %s", res.stderr)
		}
		assertOnlyRevParseShowToplevel(t, logPath)
	})

	t.Run("project_from_subdirectory", func(t *testing.T) {
		repo := t.TempDir()
		initGitRepo(t, repo)
		sub := filepath.Join(repo, "deep", "nested", "dir")
		if err := os.MkdirAll(sub, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", sub, err)
		}
		binDir, logPath := gitInvocationShim(t)
		res := runVendorWithPATH(t, sub, "", vendoredBundleAbs(t), binDir) // default resolution
		if res.code != 0 {
			t.Fatalf("expected success: %s", res.stderr)
		}
		assertOnlyRevParseShowToplevel(t, logPath)
	})

	t.Run("ambiguous_no_tty_fails_closed", func(t *testing.T) {
		cwd := t.TempDir() // not a repo
		home := t.TempDir()
		binDir, logPath := gitInvocationShim(t)
		res := runVendorWithPATH(t, cwd, home, vendoredBundleAbs(t), binDir) // no --scope
		if res.code == 0 {
			t.Fatalf("expected fail-closed exit; got 0")
		}
		assertOnlyRevParseShowToplevel(t, logPath)
	})

	t.Run("tampered_fetch_fails_closed", func(t *testing.T) {
		tamperedSrc := t.TempDir()
		if err := copyDirT(t, vendoredBundleAbs(t), tamperedSrc); err != nil {
			t.Fatalf("copying bundle to tamper: %v", err)
		}
		victim := filepath.Join(tamperedSrc, "reference", "invariants.md")
		writeFileT(t, victim, readFileT(t, victim)+"tampered\n")
		repo := t.TempDir()
		initGitRepo(t, repo)
		binDir, logPath := gitInvocationShim(t)
		res := runVendorWithPATH(t, repo, "", tamperedSrc, binDir, "--scope", "project")
		if res.code == 0 {
			t.Fatalf("expected failure on a tampered bundle")
		}
		assertOnlyRevParseShowToplevel(t, logPath)
	})

	t.Run("invalid_scope_value", func(t *testing.T) {
		cwd := t.TempDir()
		binDir, logPath := gitInvocationShim(t)
		res := runVendorWithPATH(t, cwd, "", vendoredBundleAbs(t), binDir, "--scope", "nowhere")
		if res.code == 0 {
			t.Fatalf("expected failure on an invalid --scope value")
		}
		assertOnlyRevParseShowToplevel(t, logPath)
	})

	t.Run("re_run", func(t *testing.T) {
		repo := t.TempDir()
		initGitRepo(t, repo)
		binDir, logPath := gitInvocationShim(t)
		if res := runVendorWithPATH(t, repo, "", vendoredBundleAbs(t), binDir, "--scope", "project"); res.code != 0 {
			t.Fatalf("first run failed: %s", res.stderr)
		}
		if res := runVendorWithPATH(t, repo, "", vendoredBundleAbs(t), binDir, "--scope", "project"); res.code != 0 {
			t.Fatalf("second run failed: %s", res.stderr)
		}
		assertOnlyRevParseShowToplevel(t, logPath)
	})
}

// decision-0068 D1/D4/D5 and spec-0005 AC2a. The rendered rules file is the
// whole point of the change: before it, a vendored install delivered NO rules at
// all (measured; issue #201). Everything asserted here is a silent-failure mode
// — none of it errors when wrong, it just governs nothing.
func TestVendorRendersClaudeRulesFile(t *testing.T) {
	repo := t.TempDir()
	initGitRepo(t, repo)
	res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
	if res.code != 0 {
		t.Fatalf("exit %d\nstdout: %s\nstderr: %s", res.code, res.stdout, res.stderr)
	}
	rendered := filepath.Join(repo, ".claude", "rules", "trellis.md")
	raw, err := os.ReadFile(rendered)
	if err != nil {
		t.Fatalf("the rules file is the delivery mechanism; without it the install ships nothing: %v", err)
	}
	got := string(raw)
	lines := strings.Split(got, "\n")
	hasLine := func(want string) bool {
		for _, l := range lines {
			if strings.TrimRight(l, "\r") == want {
				return true
			}
		}
		return false
	}

	// --- the import form. Both spellings are legal markdown and each is correct
	// in exactly one location; in the other it loads NOTHING, with no error and
	// no content. Measured both ways. This file is real (not a symlink) and sits
	// at .claude/rules/, so the ../../ form is the correct one.
	if !hasLine("@../../.trellis/rules.toml") {
		t.Errorf("missing the import line @../../.trellis/rules.toml — the rows would never load:\n%s", got)
	}
	if hasLine("@rules.toml") {
		t.Errorf("emitted the SIBLING import form @rules.toml — correct only from a symlinked file in .trellis/, silently loads nothing from here")
	}
	if hasLine("@.trellis/rules.toml") {
		t.Errorf("emitted the project-root import form — measured NOT to resolve, silently")
	}
	// --- the placeholder the payload header ships must be RESOLVED, not copied.
	// Left in place it resolves to .claude/rules/rules.md, which does not exist,
	// and the entire rules body vanishes while every other assertion still passes.
	if hasLine("@rules.md") {
		t.Errorf("the @rules.md placeholder survived the render — the rules body would silently drop out")
	}

	// --- content anchored to shipped payload bytes, not to literals in this test.
	// This is what makes decision-0053's "the tested wording is the shipped
	// wording" mechanical instead of reviewed.
	files := payloadFiles()
	body := files["rules.md"]
	if !strings.Contains(got, body) {
		t.Errorf("the rules body is not byte-identical to the shipped reference/rules.md")
	}
	// This fixture has no .trellis/rules.toml, so the header is trellis-b's —
	// the hook's absent-strictness branch. TestVendorRenderHeaderFollowsRulesTomlStrictness
	// covers the file being present (TRL-37).
	posture := "**How strictly to follow them:** **By default**"
	if !strings.Contains(got, posture) {
		t.Errorf("missing trellis-b's posture prose; with no rules.toml the render must follow the hook's absent-strictness branch")
	}
	if strings.Contains(got, "**Firmly** — treat these as hard requirements") {
		t.Errorf("emitted trellis-a's firm posture with no rules.toml present; the hook resolves absent strictness to adaptive, and the seed written beside this file is adaptive too")
	}

	// --- ordering. The authority header states the rows are "loaded below the
	// rules"; emitting the import above them makes the shipped text lie.
	iBody := strings.Index(got, "inv-directional-flow")
	iImport := strings.Index(got, "@../../.trellis/rules.toml")
	if iBody < 0 || iImport < 0 || iImport < iBody {
		t.Errorf("the import must come AFTER the rules body (the authority header says rows load below the rules)")
	}

	// --- D5's one sentence of new prose: which source is authoritative. The
	// frozen posture sentence and the live rows can disagree, and a reader must
	// be told which wins rather than left to see a contradiction.
	if !strings.Contains(got, "strictness") || !strings.Contains(got, "authoritative") {
		t.Errorf("D5 requires the file to name .trellis/rules.toml's strictness key as authoritative over the frozen sentence above it")
	}

	// --- the drift surface. Without an embedded stamp the hook can only stand
	// down blindly, and a file rendered by an older installer would govern
	// forever with no signal — decision-0035's floor applied to this artifact,
	// and the gap decision-0068's own Open 4 recorded before Codex found it.
	// Verified by mutation: before this assertion, removing the stamp entirely
	// left the suite green.
	stampRe := regexp.MustCompile(`<!-- trellis:rendered-from payload@[0-9a-f]{12} -->`)
	if !stampRe.MatchString(got) {
		t.Errorf("the rendered file carries no payload stamp — the hook cannot tell a stale install from a current one:\n%s", got)
	}
	shipped := strings.TrimSpace(files["version"])
	if !strings.Contains(got, "<!-- trellis:rendered-from "+shipped+" -->") {
		t.Errorf("the embedded stamp must be the payload actually vendored (%s), or the drift check compares against the wrong thing", shipped)
	}

	// --- the invariants pointer must name a path this install actually creates.
	// The shipped header names .trellis/internal/invariants.md, which the install
	// path never writes (D1: install.sh never touches .trellis/).
	if strings.Contains(got, ".trellis/internal/invariants.md") {
		t.Errorf("the rendered file points at .trellis/internal/invariants.md, which this path never creates — a dead reference")
	}
	if !strings.Contains(got, ".claude/skills/trellis/reference/invariants.md") {
		t.Errorf("the invariants pointer must name the vendored copy this install DOES create")
	}
	// An absolute path would leak the installing machine's filesystem into a file
	// §4 actively suggests committing, and would be wrong on a collaborator's box.
	if strings.Contains(got, repo) {
		t.Errorf("the rendered file embeds an absolute path from the installing machine")
	}
}

// D1 ruled project scope only. A personal install renders nothing and says why:
// ~/.claude/rules/trellis.md would govern EVERY repo on the machine and import
// ~/.trellis/rules.toml, which nothing writes — shipping precisely the
// silent-no-op artifact this change exists to prevent.
func TestVendorPersonalScopeRendersNoRulesFile(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	res := runVendor(t, work, home, vendoredBundleAbs(t), "--scope", "personal")
	if res.code != 0 {
		t.Fatalf("exit %d\nstderr: %s", res.code, res.stderr)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "rules", "trellis.md")); err == nil {
		t.Fatalf("personal scope must NOT render a user-wide rules file")
	}
	if _, err := os.Stat(filepath.Join(work, ".claude", "rules", "trellis.md")); err == nil {
		t.Fatalf("personal scope must not render into the working directory either")
	}
	if !strings.Contains(res.stdout, "no rules file") {
		t.Errorf("silently delivering nothing is the defect this change fixes; personal scope must SAY it rendered no rules file:\n%s", res.stdout)
	}
}

// HIGH, found by independent code review and reproduced: `{ ...; } > file`
// swallows a redirect failure. With a read-only .claude/rules/trellis.md the
// script exited 0, printed "rules: .claude/rules/trellis.md", and left the
// user's prior bytes untouched — the installer lying about what it did, which is
// the same silent-delivery class this whole change exists to fix. It is also
// shell-dependent: bash continues, dash aborts with a bare exit 2.
func TestVendorRenderFailureIsLoudAndFailsClosed(t *testing.T) {
	for _, tc := range []struct{ name, kind string }{
		{"read-only target", "file"},
		{"directory in the way", "dir"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := t.TempDir()
			initGitRepo(t, repo)
			dst := filepath.Join(repo, ".claude", "rules", "trellis.md")
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
				t.Fatal(err)
			}
			switch tc.kind {
			case "file":
				if err := os.WriteFile(dst, []byte("PRIOR USER CONTENT\n"), 0o444); err != nil {
					t.Fatal(err)
				}
			case "dir":
				if err := os.MkdirAll(dst, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
			if tc.kind == "file" {
				// A read-only regular file is REPLACED, and that is correct: the
				// install owns this path by name, and re-running must be
				// idempotent. Render-to-temp-then-mv makes it work where the
				// direct redirect silently did not.
				if res.code != 0 {
					t.Fatalf("a read-only target should be replaced, not fail: exit %d\n%s", res.code, res.stderr)
				}
				b, err := os.ReadFile(dst)
				if err != nil || strings.Contains(string(b), "PRIOR USER CONTENT") {
					t.Errorf("the read-only file was not replaced; got %q (%v)", string(b), err)
				}
				return
			}
			// A non-regular target must fail loudly. `mv file dir` moves the file
			// INTO the directory: measured, exit 0, a success banner, no rules
			// file, and a stray temp file buried inside it.
			if res.code == 0 {
				t.Fatalf("reported SUCCESS with a directory in the way — exit 0 with stdout:\n%s", res.stdout)
			}
			if strings.Contains(res.stdout, "rules: .claude/rules/trellis.md") {
				t.Errorf("claimed it rendered the rules file when it could not")
			}
			if !strings.Contains(res.stdout+res.stderr, "trellis: FAIL:") {
				t.Errorf("failed without the script's own fail() message — a bare shell error is not a diagnosis:\n%s", res.stdout+res.stderr)
			}
		})
	}
}

// The writer and the reader are only tied together by running both. Every other
// test uses a literal fixture for one side or the other; verified by mutation,
// renaming install.sh's output leaves the hook test green.
func TestInstalledRulesFileSilencesTheHookExactlyOnce(t *testing.T) {
	repo := t.TempDir()
	initGitRepo(t, repo)
	if res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project"); res.code != 0 {
		t.Fatalf("install failed: %s", res.stderr)
	}
	if err := os.MkdirAll(filepath.Join(repo, ".trellis"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".trellis", "rules.toml"), []byte(payloadFiles()["rules-b.toml"]), 0o644); err != nil {
		t.Fatal(err)
	}
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(hook)
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+repo,
		"CLAUDE_PLUGIN_ROOT="+filepath.Join(repo, ".claude", "skills", "trellis"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hook exited non-zero: %v: %s", err, out)
	}
	if strings.Contains(string(out), "inv-directional-flow") {
		t.Fatalf("DOUBLE DELIVERY against the REAL rendered file — the hook injected over it:\n%s", out)
	}
	// The stand-down assertion must NOT be a substring both messages share. Both
	// the quiet stand-down and TRELLIS_RULES_NOT_LOADED name the path, so
	// asserting the path alone passed on the very failure this test exists to
	// catch — found by review, not by mutation.
	if strings.Contains(string(out), "TRELLIS_RULES_NOT_LOADED") {
		t.Fatalf("the hook judged the REAL installer's own output incomplete:\n%s", out)
	}
	if !strings.Contains(string(out), "already loaded from .claude/rules/trellis.md") {
		t.Fatalf("expected the quiet stand-down naming the artifact; got:\n%s", out)
	}
}

// Codex P1: a pre-plugin-delivery consumer still has `.trellis/internal/` AND a
// CLAUDE.md managed block importing `@.trellis/internal/trellis.md`. Rendering
// `.claude/rules/trellis.md` there puts BOTH static chains into context, because
// Claude loads them itself before any hook runs. The hook's path-A-first ordering
// suppresses only what the HOOK injects — it cannot un-load a file Claude already
// read. So this combination has to be refused at install time; there is no
// runtime fix for it.
// TestVendorRefusalRemedyNamesTheShapeItFound: a Claude review finding on #227,
// and the SECOND appearance of one defect class. staleness.sh had a remedy
// hard-coded to .trellis/internal/ while its branch fired for two overlay shapes;
// that was fixed earlier in this branch. install.sh had the same bug and was not
// looked at — its $static_conflict takes THREE values, one of which
// ("managed block in <file>") involves no overlay directory at all.
//
// Following a remedy that names the wrong shape deletes nothing, so the refusal
// fires again on every subsequent run: a permanent false-positive refusal that
// the consumer cannot clear by doing what they were told.
func TestVendorRefusalRemedyNamesTheShapeItFound(t *testing.T) {
	cases := []struct {
		name    string
		build   func(t *testing.T, repo string)
		wants   []string
		forbids []string
	}{
		{
			name: "internal overlay",
			build: func(t *testing.T, repo string) {
				mustMkdirAll(t, filepath.Join(repo, ".trellis", "internal"))
				mustWrite(t, filepath.Join(repo, ".trellis", "internal", "version"), "payload@000000000000\n")
			},
			wants: []string{".trellis/internal/"},
		},
		{
			name: "legacy flat overlay",
			build: func(t *testing.T, repo string) {
				mustMkdirAll(t, filepath.Join(repo, ".trellis"))
				mustWrite(t, filepath.Join(repo, ".trellis", "trellis.md"), "legacy vendored prose\n")
			},
			wants:   []string{".trellis/trellis.md"},
			forbids: []string{"delete .trellis/internal/"},
		},
		{
			name: "inline managed block, no overlay directory",
			build: func(t *testing.T, repo string) {
				mustWrite(t, filepath.Join(repo, "CLAUDE.md"), "<!-- trellis:begin -->\nrules\n<!-- trellis:end -->\n")
			},
			wants:   []string{"managed block in CLAUDE.md"},
			forbids: []string{".trellis/internal/", ".trellis/trellis.md"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := t.TempDir()
			initGitRepo(t, repo)
			tc.build(t, repo)

			res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
			combined := res.stdout + res.stderr

			for _, want := range tc.wants {
				if !strings.Contains(combined, want) {
					t.Errorf("the remedy must name %q — the shape actually present; got:\n%s", want, combined)
				}
			}
			for _, forbid := range tc.forbids {
				if strings.Contains(combined, forbid) {
					t.Errorf("the remedy names %q, which this project does not have — following it "+
						"deletes nothing and the refusal fires forever; got:\n%s", forbid, combined)
				}
			}
		})
	}
}

func mustMkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestVendorRefusesToRenderOverAVendoredOverlay(t *testing.T) {
	repo := t.TempDir()
	initGitRepo(t, repo)
	internal := filepath.Join(repo, ".trellis", "internal")
	if err := os.MkdirAll(internal, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(internal, "version"), []byte("payload@000000000000\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")

	if _, err := os.Stat(filepath.Join(repo, ".claude", "rules", "trellis.md")); err == nil {
		t.Fatalf("rendered over a vendored overlay — both static chains would load, and no hook can prevent it")
	}
	// The git-add suggestion must not name a file this path never wrote:
	// `git add` on a missing pathspec exits 128, and the `&&` means the commit
	// never runs either — so the printed command stages nothing at all.
	if strings.Contains(res.stdout, "add .claude/skills/trellis .claude/rules/trellis.md") {
		t.Errorf("the commit suggestion names the rendered file on a path that did not render it — the printed command would fail with exit 128")
	}
	combined := res.stdout + res.stderr
	if !strings.Contains(combined, ".trellis/internal") {
		t.Errorf("refusing silently is its own defect: the output must name the overlay as the reason; got:\n%s", combined)
	}
	// This used to accept `/trellis:setup` OR `/trellis:remove`. Both halves were
	// wrong after decision-0072: the first names a retired skill, and the second is
	// a REMOVAL route, not a migration route — it satisfied the assertion alone,
	// so deleting every line of migration guidance left the test green.
	if !strings.Contains(combined, ".trellis/internal/") || !strings.Contains(combined, ".trellis/rules.toml") {
		t.Errorf("a refusal with no way forward strands the user; name the migration route and what survives it; got:\n%s", combined)
	}
	if strings.Contains(combined, "/trellis:setup") {
		t.Errorf("refusal still points at the setup skill, retired by decision-0072; got:\n%s", combined)
	}
	// The bundle itself must still vendor — the overlay conflicts with the
	// rendered file, not with the plugin package.
	assertBundleVendored(t, filepath.Join(repo, ".claude", "skills", "trellis"))
}

// Codex P2: an unanchored search for the opening marker classified any file that
// merely MENTIONS `<!-- trellis:begin` — contributor guidance, a changelog, this
// project's own docs — as static delivery, suppressing the render and reporting
// a conflict that does not exist. That leaves the curl path without its only
// measured delivery mechanism, for a document that delivers nothing.
func TestVendorRendersDespiteAMereMentionOfTheManagedMarker(t *testing.T) {
	repo := t.TempDir()
	initGitRepo(t, repo)
	writeFileT(t, filepath.Join(repo, "CLAUDE.md"),
		"# Contributing\n\nTrellis writes a managed block delimited by `<!-- trellis:begin` and its\n"+
			"closing marker. Do not hand-edit inside it.\n")

	res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
	if res.code != 0 {
		t.Fatalf("exit %d: %s", res.code, res.stderr)
	}
	if _, err := os.Stat(filepath.Join(repo, ".claude", "rules", "trellis.md")); err != nil {
		t.Fatalf("a document that merely NAMES the marker is not a managed block — the render was suppressed for a conflict that does not exist: %v", err)
	}
}

// AC2c coverage for the two conflict shapes that had none. `.trellis/internal/`
// was tested; the legacy flat overlay and the paired managed block were not —
// only the negative "mere mention" case was.
func TestVendorRefusesForEveryStaticDeliveryShape(t *testing.T) {
	for _, tc := range []struct {
		name, wantReason string
		setup            func(string)
	}{
		{"legacy flat overlay", ".trellis", func(repo string) {
			writeFileT(t, filepath.Join(repo, ".trellis", "trellis.md"), "legacy vendored prose\n")
		}},
		// The INLINE block, reinstated. Deleting this case is what let a
		// regression through: the inline form embeds the rules body in CLAUDE.md
		// and needs no .trellis/internal/, so BOTH existence checks miss it and
		// the render proceeded into live double delivery. Measured against the
		// shipped block, not a hand-written approximation.
		{"inline managed block", "managed block in CLAUDE.md", func(repo string) {
			writeFileT(t, filepath.Join(repo, "CLAUDE.md"), inlineBlockFixture(t))
			writeFileT(t, filepath.Join(repo, ".trellis", "rules.toml"), "# rows only — the inline form needs nothing else under .trellis/\n")
		}},
		// A BOM'd block. Same fail-open direction as the reader-side BOM bug fixed
		// in the same change: setup wrote the block at line 1 of a fresh CLAUDE.md,
		// an editor on a Windows-default checkout rewrote the encoding, and a
		// column-0 grep with no BOM tolerance then missed a REAL block and rendered
		// into live double delivery. Measured before the fix: rules file PRESENT.
		{"inline managed block, UTF-8 BOM", "managed block in CLAUDE.md", func(repo string) {
			writeFileT(t, filepath.Join(repo, "CLAUDE.md"), "\xef\xbb\xbf"+inlineBlockFixture(t))
			writeFileT(t, filepath.Join(repo, ".trellis", "rules.toml"), "# rows only\n")
		}},
		// The CRLF variant is the bug that killed the original check: the old
		// CLOSING grep was $-anchored, so `trellis:end -->\r` never matched and a
		// REAL block went undetected on the Git-for-Windows default. Dropping the
		// closing grep removed the only $-anchor. Without this case the fix is
		// unpinned and the next simplification reintroduces it.
		{"inline managed block, CRLF checkout", "managed block in CLAUDE.md", func(repo string) {
			// The shipped block ends `<!-- trellis:end -->` with NO trailing
			// newline. A naive ReplaceAll over it therefore leaves the CLOSING
			// marker CR-free, `-->$` still matches, and this case cannot detect
			// the very regression it exists to pin — measured: with the
			// $-anchored closing grep reinstated, all four subtests passed. Append
			// the newline first so the closing line is genuinely CRLF-terminated.
			crlf := strings.ReplaceAll(inlineBlockFixture(t)+"\n", "\n", "\r\n")
			if !strings.Contains(crlf, "<!-- trellis:end -->\r\n") {
				t.Fatalf("this fixture is named CRLF but its closing marker is not CR-terminated — the case would pass against the bug it guards")
			}
			writeFileT(t, filepath.Join(repo, "CLAUDE.md"), crlf)
			writeFileT(t, filepath.Join(repo, ".trellis", "rules.toml"), "# rows only\n")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := t.TempDir()
			initGitRepo(t, repo)
			tc.setup(repo)
			res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
			if res.code != 0 {
				t.Fatalf("exit %d: %s", res.code, res.stderr)
			}
			if _, err := os.Stat(filepath.Join(repo, ".claude", "rules", "trellis.md")); err == nil {
				t.Fatalf("rendered over a %s — both static chains would load, and no hook can undo it", tc.name)
			}
			if !strings.Contains(res.stdout, tc.wantReason) {
				t.Errorf("the refusal must NAME the shape it found (%q); got:\n%s", tc.wantReason, res.stdout)
			}
		})
	}
}

// Two guards added in response to earlier review findings shipped with NO test —
// deleting either left the whole suite green. Found by the code reviewer, not by
// mutation, because nothing existed to mutate.
func TestVendorGuardsAddedByReviewAreActuallyPinned(t *testing.T) {
	t.Run("pre-existing rendered file plus an overlay is reported as LIVE", func(t *testing.T) {
		repo := t.TempDir()
		initGitRepo(t, repo)
		// Rendered first, overlay arrives later — a collaborator's commit, or a
		// reverted migration. Refusing does not help: double delivery is already on.
		writeFileT(t, filepath.Join(repo, ".claude", "rules", "trellis.md"), "previously rendered\n")
		writeFileT(t, filepath.Join(repo, ".trellis", "internal", "version"), "payload@000000000000\n")

		res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
		if res.code != 0 {
			t.Fatalf("exit %d: %s", res.code, res.stderr)
		}
		if !strings.Contains(res.stdout, "ALREADY EXISTS") {
			t.Errorf("saying 'no rules file' here is false at the moment it prints — the file is on disk and delivering; got:\n%s", res.stdout)
		}
		if b, err := os.ReadFile(filepath.Join(repo, ".claude", "rules", "trellis.md")); err != nil || !strings.Contains(string(b), "previously rendered") {
			t.Errorf("the installer must not silently remove a file it did not create")
		}
	})

	// The import-form block brings .trellis/internal/ with it — it imports
	// @.trellis/internal/trellis.md — so the EXISTENCE check catches this shape
	// even with the marker grep removed. That is true and worth pinning; what it
	// is not is a reason to remove the grep, which an earlier revision of this
	// branch concluded. The inline shape has no such backstop (see the shape
	// table above), so both mechanisms are load-bearing for different shapes.
	t.Run("an import-form block is caught by its overlay, not by grepping prose", func(t *testing.T) {
		repo := t.TempDir()
		initGitRepo(t, repo)
		writeFileT(t, filepath.Join(repo, "CLAUDE.md"),
			"# Project\n\n<!-- trellis:begin (managed by trellis) -->\n@.trellis/internal/trellis.md\n<!-- trellis:end -->\n")
		writeFileT(t, filepath.Join(repo, ".trellis", "internal", "version"), "payload@000000000000\n")
		res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
		if res.code != 0 {
			t.Fatalf("exit %d: %s", res.code, res.stderr)
		}
		if _, err := os.Stat(filepath.Join(repo, ".claude", "rules", "trellis.md")); err == nil {
			t.Fatalf("rendered over an import-form managed block — live double delivery")
		}
	})

	// The converse: prose that NAMES the delimiters must never suppress the
	// render. Two review rounds were spent on grep forms that got this wrong,
	// which is why the surviving grep is anchored at column 0 — documentation
	// writes the marker mid-sentence, a real block writes it at line start.
	t.Run("prose naming the delimiters never suppresses the render", func(t *testing.T) {
		repo := t.TempDir()
		initGitRepo(t, repo)
		writeFileT(t, filepath.Join(repo, "CLAUDE.md"),
			"# Contributing\n\nA managed region is delimited by `<!-- trellis:begin ... -->`\n"+
				"and `<!-- trellis:end -->`. Never hand-edit between them.\n")
		res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
		if res.code != 0 {
			t.Fatalf("exit %d: %s", res.code, res.stderr)
		}
		if _, err := os.Stat(filepath.Join(repo, ".claude", "rules", "trellis.md")); err != nil {
			t.Fatalf("prose suppressed the render — issue #201 restored for a conflict that does not exist: %v", err)
		}
	})

	// install.sh reads EXACTLY TWO project files' contents: the managed-block
	// opening marker, and the `strictness` key of an existing .trellis/rules.toml.
	// An earlier revision of this subtest asserted zero, which was achieved by
	// deleting the marker check and regressed inline consumers into silent double
	// delivery; it then pinned exactly one, with the note that "a second content
	// read has to come here and argue itself". This is that argument (TRL-37):
	// the rendered file carries a posture sentence, and taking it from a constant
	// while the rows a few lines below say `firm` put the adaptive header over a
	// firm project — the plugin's own hook reads the same key to pick the same
	// header, so the render now reads it too, with the hook's own parser
	// (TestInstallScriptStrictnessParserMatchesHook). It still selects nothing
	// from an instructions file, and still writes nothing under .trellis/ beyond
	// the seed. Asserting the exact count keeps the widening bounded — a third
	// read has to come here and argue itself.
	t.Run("exactly two project file content reads: the marker check and the strictness key", func(t *testing.T) {
		// Classify every grep by its OPERAND, not by whether the line happens to
		// mention `git_root`. That earlier form was bypassed by aliasing:
		//   claude_md="${git_root}/CLAUDE.md"
		//   grep -q '^strictness' "$claude_md" && ...
		// read a declared posture — the exact thing AC2's heading forbids — and
		// left the whole suite green. Here, an operand counts as safe only if it
		// is rooted at the staged bundle or at the temp file this script itself
		// just wrote; every other file operand is a read of pre-existing state.
		//
		// Still lexical, so still not proof: a determined rewrite could launder
		// the path through more indirection. It closes the demonstrated bypass and
		// raises the cost of the next one; the behavioural fixtures above are what
		// actually establish the property.
		// grep's first quoted operand is the PATTERN; every one after it is a
		// FILE. Classifying by position beats matching the line for `git_root`,
		// which an alias walks straight past, and beats "operand contains a
		// slash", which the same alias also defeats.
		// Every command that can read a file, not just grep. The earlier version
		// keyed on `grep`, and a reviewer walked past it with
		//   sed -n '/inv-directional-flow/p' "$git_root/README.md"
		// which reads project content and left the whole suite green.
		//
		// Operands are counted only when they are double-quoted and $-rooted —
		// that is what a path built from this script's variables looks like, and
		// it excludes heredoc delimiters and /dev/tty. Command names are matched
		// at word boundaries, since a substring match makes `chmod` look like
		// `od`. `scripted` commands take their pattern first and files after.
		reader := regexp.MustCompile(`(^|[\s;|(&])(grep|sed|awk|cat|head|tail|wc|cut|tr|sort|od|read)\s|<\s*"`)
		quoted := regexp.MustCompile(`'[^']*'|"[^"]*"`)
		scripted := regexp.MustCompile(`(^|[\s;|(&])(grep|sed|awk)\s`)
		operand := regexp.MustCompile(`"\$\{?[A-Za-z_][A-Za-z0-9_]*[^"]*"`)
		var reads []string
		for _, line := range strings.Split(readFileT(t, installScriptPath(t)), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "#") || strings.Contains(trimmed, "<<") {
				continue
			}
			loc := reader.FindStringIndex(trimmed)
			if loc == nil {
				continue
			}
			cmd := trimmed[loc[0]:]
			// Stop at the redirect so a trailing `|| { x="..."; }` clause does not
			// masquerade as another file operand. Absence of a redirect only ever
			// ADDS operands, and the assertion is an exact count, so the omission
			// fails closed rather than opening a hole.
			if r := strings.Index(cmd, " 2>"); r > 0 {
				cmd = cmd[:r]
			}
			// Two stages. First every quoted token, in order, because for a
			// scripted command the FIRST is the pattern and must not be mistaken
			// for a file. Then keep only $-rooted ones: that is what a path built
			// from this script's variables looks like, and it drops heredoc
			// delimiters, /dev/tty, and literal patterns.
			ops := quoted.FindAllString(cmd, -1)
			if scripted.MatchString(cmd) {
				if len(ops) == 0 {
					continue
				}
				ops = ops[1:]
			}
			for _, f := range ops {
				if !operand.MatchString(f) {
					continue // a literal, not a path built from a variable
				}
				if strings.HasPrefix(f, `"$stage/`) || f == `"$rendered_tmp"` || strings.HasPrefix(f, `"$target/`) {
					continue // the staged bundle, the temp file, or the vendor target
				}
				reads = append(reads, trimmed)
				break
			}
		}
		if len(reads) != 2 {
			t.Fatalf("install.sh makes %d content read(s) of a project file; exactly two are argued (the managed-block marker, and the strictness key of .trellis/rules.toml):\n%s", len(reads), strings.Join(reads, "\n"))
		}
		var marker, strictness string
		for _, r := range reads {
			switch {
			case strings.Contains(r, `<!-- trellis:begin`):
				marker = r
			case strings.Contains(r, `.trellis/rules.toml`):
				strictness = r
			}
		}
		if marker == "" || !strings.Contains(marker, `"^\(`) {
			t.Errorf("one permitted content read must be the column-0-anchored opening marker (optionally BOM-prefixed); reads were:\n%s", strings.Join(reads, "\n"))
		}
		if strings.Contains(marker, "trellis:end") {
			t.Errorf("the CLOSING marker grep is $-anchored and broke on CRLF checkouts — it must not come back: %s", marker)
		}
		// The strictness read is awk over the file as STDIN, so that this
		// lexical guard sees it: an awk program that spans lines carries its file
		// operand on a line with no command name, and a read written that way
		// would slip past the scan above uncounted. The redirect form is the one
		// this scan matches by construction.
		if strictness == "" || !strings.Contains(strictness, `< "$git_root/.trellis/rules.toml"`) {
			t.Errorf("the other permitted content read must be the strictness key of .trellis/rules.toml, read via a `< \"$git_root/.trellis/rules.toml\"` redirect; reads were:\n%s", strings.Join(reads, "\n"))
		}
	})
}

// The render's failure paths shipped with NO test, and that is exactly how a
// dead one got through: the diagnostic expanded $bundle_src, a variable that
// never existed, so under `set -eu` the message was replaced by an expansion
// error — and because an EXIT trap becomes the shell's last command, bash
// reported the whole failed install as exit 0. A silent success for an install
// that refused to complete.
//
// This drives the guard for real by removing a render step, rather than
// asserting the message exists in the source.
func TestRenderFailureIsLoudAndLeavesNothingBehind(t *testing.T) {
	script := readFileT(t, installScriptPath(t))
	broken := strings.Replace(script, "    cat \"$stage/render.body\"\n", "", 1)
	if broken == script {
		t.Fatal("could not find the render-body step to remove; this test's premise has drifted")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "install.sh")
	if err := os.WriteFile(path, []byte(broken), 0o755); err != nil {
		t.Fatal(err)
	}
	repo := t.TempDir()
	initGitRepo(t, repo)
	cmd := exec.Command("/bin/sh", path, "--non-interactive", "--scope", "project")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "TRELLIS_BUNDLE_SOURCE="+vendoredBundleAbs(t))
	out, err := cmd.CombinedOutput()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	}
	if code == 0 {
		t.Errorf("a failed render exited 0 — the installer announced success for an install it refused to complete:\n%s", out)
	}
	if !strings.Contains(string(out), "the rendered rules file is incomplete") {
		t.Errorf("the failure must NAME what was missing; got:\n%s", out)
	}
	if strings.Contains(string(out), "unbound variable") || strings.Contains(string(out), "parameter not set") {
		t.Errorf("the diagnostic expands an undefined variable, so the message never prints:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(repo, ".claude", "rules", "trellis.md")); err == nil {
		t.Errorf("a failed render left a rules file behind")
	}
	if m, _ := filepath.Glob(filepath.Join(repo, ".claude", "rules", ".trellis.md.*")); len(m) > 0 {
		t.Errorf("a failed render left its temp file behind: %v", m)
	}
}

// An explicitly-passed but empty --scope used to fall through to the default,
// silently ignoring a flag the user typed. Unpinned until mutation found it.
func TestVendorRejectsAnEmptyScopeFlag(t *testing.T) {
	for _, arg := range [][]string{{"--scope", ""}, {"--scope="}} {
		repo := t.TempDir()
		initGitRepo(t, repo)
		res := runVendor(t, repo, "", vendoredBundleAbs(t), arg...)
		if res.code == 0 {
			t.Errorf("%v was accepted; an explicitly-passed empty scope must not resolve to the default", arg)
		}
		if !strings.Contains(res.stdout+res.stderr, "scope must be personal or project") {
			t.Errorf("%v failed without naming the reason:\n%s", arg, res.stdout+res.stderr)
		}
	}
}

// TestInstallScriptNamesNoProjectFileInExecutableCode (spec-0005 AC2's
// checkable-by-absence clause). The clause names five strings; before this test
// nothing enforced it, and the clause had silently gone false once already —
// an interim revision of this branch grepped CLAUDE.md/AGENTS.md for a managed
// block while the spec still promised the check passed.
//
// This is the WEAK half of AC2 on purpose. It cannot prove the script doesn't
// branch on instructions-file state under some other name; that is
// TestVendorZeroDecisionLogicAcrossInstructionFileVariants's job. What it does
// prove is that the spec's own stated check still holds, so a reviewer running
// it by hand gets the answer the spec promises.
func TestInstallScriptNamesNoProjectFileInExecutableCode(t *testing.T) {
	// These two terms appear NOWHERE in install.sh — not in code, not in a
	// comment — so this needs no comment parsing and therefore has no comment-
	// parsing bypass. An earlier version of this test split each line at the
	// first `#` and treated the remainder as a comment. Shell disagrees: that
	// truncated four real code lines in the then-current script (`while [ $# -gt
	// 0 ]`, `SCOPE_FLAG="${1#--scope=}"`, and a printf whose entire payload sat
	// after a `##`), and a reviewer demonstrated a genuine content read placed
	// after a `#` that left the test green. A raw substring match cannot be
	// bypassed that way.
	//
	// `trellis:begin`, `CLAUDE.md` and `AGENTS.md` are deliberately NOT here:
	// spec-0005 AC2's amendment permits the managed-block content read and names
	// them; TRL-37 argued a second read, the strictness key. The bounded version
	// of that guard is the exactly-two-reads subtest in
	// TestVendorGuardsAddedByReviewAreActuallyPinned.
	terms := []string{"expression.md", "profile-"}
	for i, line := range strings.Split(readFileT(t, installScriptPath(t)), "\n") {
		for _, term := range terms {
			if strings.Contains(line, term) {
				t.Errorf("install.sh:%d names %q: %s\n\nspec-0005 AC2 promises this grep returns nothing — the script must never branch on a declared POSTURE. Either decision logic came back, or AC2's check list needs amending in the same act.", i+1, term, strings.TrimSpace(line))
			}
		}
	}
}

// inlineBlockFixture returns the INLINE managed block exactly as trellis shipped
// it, read from the bundle rather than hand-written. A hand-written
// approximation is what made this shape look coverable by an existence check:
// the real block embeds the full rules body and imports nothing, so no
// .trellis/internal/ ever appears beside it.
func inlineBlockFixture(t *testing.T) string {
	t.Helper()
	block := readFileT(t, filepath.Join(vendoredBundleAbs(t), "reference", "block-inline-b.md"))
	if !strings.HasPrefix(block, "<!-- trellis:begin") {
		t.Fatalf("the shipped inline block no longer opens with the marker at column 0; this fixture and install.sh's grep both assume it does:\n%.120s", block)
	}
	if strings.Contains(block, "\n@") {
		t.Fatalf("the shipped inline block now carries an @-import; if the inline form gained a .trellis/ dependency, install.sh's existence checks may cover it and this fixture's premise needs re-deriving")
	}
	return block
}

// TRL-37. The rendered file's posture header was a constant taken from
// trellis-b.md, so a project whose .trellis/rules.toml already said
// strictness = "firm" got the adaptive sentence rendered over its firm rows,
// while the plugin's own SessionStart hook (path B of staleness.sh) served
// trellis-a.md to the same project. The render now selects the header the way
// the hook does: the first `strictness = ...` line, either quote style, exactly
// `firm` selects trellis-a.md and anything else — including no file, no key,
// and a value the hook does not recognise — selects trellis-b.md.
//
// Expected heads are the shipped payload bytes above the @rules.md placeholder,
// not literals, so a reword of either header cannot leave this green by luck.
func TestVendorRenderHeaderFollowsRulesTomlStrictness(t *testing.T) {
	files := payloadFiles()
	headOf := func(name string) string {
		t.Helper()
		src := files[name]
		i := strings.Index(src, "\n@rules.md")
		if i < 0 {
			t.Fatalf("%s carries no @rules.md placeholder; the render's premise has drifted", name)
		}
		return src[:i+1]
	}
	firmHead, adaptiveHead := headOf("trellis-a.md"), headOf("trellis-b.md")
	if firmHead == adaptiveHead {
		t.Fatal("premise drifted: trellis-a.md and trellis-b.md no longer differ above @rules.md, so there is no posture header to select")
	}
	rulesA, rulesB := files["rules-a.toml"], files["rules-b.toml"]
	rowsOnly := rulesB[strings.Index(rulesB, "[rules]"):]
	singleQuoted := strings.Replace(rulesA, `strictness  = "firm"`, `strictness  = 'firm'`, 1)
	if singleQuoted == rulesA {
		t.Fatalf("the shipped rules-a.toml no longer carries `strictness  = \"firm\"`; this test's fixture needs re-deriving:\n%s", rulesA)
	}
	unrecognised := strings.Replace(rulesA, `"firm"`, `"strict"`, 1)

	for _, tc := range []struct {
		name     string
		toml     string // "" means no .trellis/rules.toml at all
		wantHead string
		wantSaid string // the posture the installer reports on stdout
	}{
		{"firm, double quotes (the shipped rules-a.toml)", rulesA, firmHead, "firm"},
		{"firm, single quotes (TOML's other string form)", singleQuoted, firmHead, "firm"},
		{"adaptive (the shipped rules-b.toml)", rulesB, adaptiveHead, "adaptive"},
		{"no rules.toml", "", adaptiveHead, "adaptive"},
		{"a value the hook does not recognise", unrecognised, adaptiveHead, "adaptive"},
		{"rows but no strictness key", rowsOnly, adaptiveHead, "adaptive"},
		{"the first strictness line wins, and a commented one is not a line",
			"# strictness = \"firm\"\nstrictness = \"adaptive\"\nstrictness = \"firm\"\n" + rowsOnly, adaptiveHead, "adaptive"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := t.TempDir()
			initGitRepo(t, repo)
			tomlPath := filepath.Join(repo, ".trellis", "rules.toml")
			if tc.toml != "" {
				writeFileT(t, tomlPath, tc.toml)
			}
			res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
			if res.code != 0 {
				t.Fatalf("exit %d\nstdout: %s\nstderr: %s", res.code, res.stdout, res.stderr)
			}
			got := readFileT(t, filepath.Join(repo, ".claude", "rules", "trellis.md"))
			if want := "<!-- trellis:rendered-begin -->\n" + tc.wantHead; !strings.HasPrefix(got, want) {
				t.Errorf("the rendered file does not open with the header the hook would serve this project:\nwant prefix:\n%s\ngot:\n%.600s", want, got)
			}
			other := firmHead
			if tc.wantHead == firmHead {
				other = adaptiveHead
			}
			if strings.Contains(got, other) {
				t.Errorf("the rendered file carries the OTHER posture header too")
			}
			// The file the header was read from is the project's own and is never
			// rewritten; with no file, the seed and the header must agree.
			if tc.toml != "" {
				if after := readFileT(t, tomlPath); after != tc.toml {
					t.Errorf(".trellis/rules.toml was modified by the install — reading the strictness key must never write it back:\nbefore:\n%s\nafter:\n%s", tc.toml, after)
				}
			} else if seeded := readFileT(t, tomlPath); seeded != rulesB {
				t.Errorf("with no rules.toml the seed must be the adaptive preset, so that it agrees with the adaptive header rendered beside it")
			}
			// Said out loud (floor-transparency): which posture was rendered and
			// why, since a reader of stdout otherwise cannot tell a firm project
			// from an adaptive one, or a recognised value from one that fell
			// through to the default.
			if !strings.Contains(res.stdout, "posture header: "+tc.wantSaid) {
				t.Errorf("stdout must name the posture rendered (%q); got:\n%s", tc.wantSaid, res.stdout)
			}
		})
	}
}

// The two deliveries have to agree on the same project, and the only way to
// know is to run both. For each strictness fixture: install, then move the
// rendered file aside and let the hook take path B (config only) from the
// vendored bundle — the header sentence the hook injects must be the one the
// installer rendered. Verified by mutation: with the render pinned to
// trellis-b.md, the firm fixture fails here.
func TestRenderedHeaderMatchesTheHooksOwnSelection(t *testing.T) {
	files := payloadFiles()
	postureLine := func(name string) string {
		t.Helper()
		for _, l := range strings.Split(files[name], "\n") {
			if strings.HasPrefix(l, "**How strictly to follow them:**") {
				return l
			}
		}
		t.Fatalf("%s carries no posture sentence; this test's premise has drifted", name)
		return ""
	}
	firmLine, adaptiveLine := postureLine("trellis-a.md"), postureLine("trellis-b.md")
	if firmLine == adaptiveLine {
		t.Fatal("premise drifted: both headers carry the same posture sentence")
	}
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, toml string }{
		{"firm", files["rules-a.toml"]},
		{"adaptive", files["rules-b.toml"]},
		{"unrecognised value", strings.Replace(files["rules-a.toml"], `"firm"`, `"strict"`, 1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := t.TempDir()
			initGitRepo(t, repo)
			writeFileT(t, filepath.Join(repo, ".trellis", "rules.toml"), tc.toml)
			if res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project"); res.code != 0 {
				t.Fatalf("install failed: %s", res.stderr)
			}
			rendered := filepath.Join(repo, ".claude", "rules", "trellis.md")
			got := readFileT(t, rendered)
			if err := os.Remove(rendered); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(hook)
			cmd.Dir = repo
			cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+repo,
				"CLAUDE_PLUGIN_ROOT="+filepath.Join(repo, ".claude", "skills", "trellis"))
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("hook exited non-zero: %v: %s", err, out)
			}
			if !strings.Contains(string(out), "inv-directional-flow") {
				t.Fatalf("the hook did not take path B (no rules injected), so there is nothing to compare against:\n%s", out)
			}
			hookFirm, hookAdaptive := strings.Contains(string(out), firmLine), strings.Contains(string(out), adaptiveLine)
			if hookFirm == hookAdaptive {
				t.Fatalf("the hook's output carries neither posture sentence or both (firm=%v adaptive=%v); the header format or the JSON escaping has drifted:\n%s", hookFirm, hookAdaptive, out)
			}
			fileFirm, fileAdaptive := strings.Contains(got, firmLine), strings.Contains(got, adaptiveLine)
			if fileFirm != hookFirm || fileAdaptive != hookAdaptive {
				t.Errorf("the installer and the hook disagree on this project's posture: rendered firm=%v adaptive=%v, hook firm=%v adaptive=%v", fileFirm, fileAdaptive, hookFirm, hookAdaptive)
			}
		})
	}
}

// install.sh cannot share a file with the hook — the hook is inside the bundle
// the script vendors — so the strictness parser is a copy, and a copy drifts.
// decision-0028: a source with a derivative gets a guard per pair. This pins the
// awk program byte for byte, and the two case arms that map its result.
func TestInstallScriptStrictnessParserMatchesHook(t *testing.T) {
	extract := func(path string) string {
		t.Helper()
		src := readFileT(t, path)
		const open = `strictness="$(awk '`
		i := strings.Index(src, open)
		if i < 0 {
			t.Fatalf("%s carries no `%s`; the parser or this guard's premise moved", path, open)
		}
		rest := src[i+len(open):]
		j := strings.Index(rest, "'")
		if j < 0 {
			t.Fatalf("%s: the awk program opened at %q never closes", path, open)
		}
		prog := rest[:j]
		if !strings.Contains(prog, "strictness") {
			t.Fatalf("%s: the extracted awk program does not mention strictness — wrong site:\n%s", path, prog)
		}
		return prog
	}
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	if h, s := extract(hook), extract(installScriptPath(t)); h != s {
		t.Errorf("install.sh's strictness parser has drifted from the hook's (plugins/trellis/hooks/staleness.sh); the two deliveries must read the key identically.\nhook:\n%s\ninstall.sh:\n%s", h, s)
	}
	script := readFileT(t, installScriptPath(t))
	for _, arm := range []string{
		`firm\)\s+header="reference/trellis-a\.md"`,
		`\*\)\s+header="reference/trellis-b\.md"`,
	} {
		if !regexp.MustCompile(arm).MatchString(script) {
			t.Errorf("install.sh lacks the case arm %s — the hook maps exactly `firm` to trellis-a.md and everything else to trellis-b.md", arm)
		}
	}
}

// Found by review, not by the suite: `[ -f ]` is true for a regular file the
// invoking user cannot read, and the `<` redirect then fails INSIDE the command
// substitution — under dash the substitution exits 2, the assignment takes that
// status, and `set -eu` kills the script after the bundle is already in place,
// with no rules file, no seed, and only the shell's own "cannot open" as the
// diagnosis. The hook does not die there (no `set -e`, awk's error swallowed):
// it serves the adaptive header. So the installer must not diverge from it over
// a permission bit. Verified by mutation: dropping the `-r` guard and the
// `|| strictness=""` fails this test with exactly that abort.
//
// Skipped when the test process can read a mode-000 file anyway (root, and CI
// images that run as it) — the fixture cannot exist there.
func TestVendorUnreadableRulesTomlFallsBackToAdaptiveInsteadOfAborting(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: a mode-000 file is still readable, so the fixture cannot be built")
	}
	repo := t.TempDir()
	initGitRepo(t, repo)
	toml := filepath.Join(repo, ".trellis", "rules.toml")
	writeFileT(t, toml, payloadFiles()["rules-a.toml"])
	if err := os.Chmod(toml, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(toml, 0o644) })
	res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
	if res.code != 0 {
		t.Fatalf("an unreadable rules.toml aborted the install (exit %d) instead of falling back the way the hook does\nstdout: %s\nstderr: %s", res.code, res.stdout, res.stderr)
	}
	got := readFileT(t, filepath.Join(repo, ".claude", "rules", "trellis.md"))
	if !strings.Contains(got, "**How strictly to follow them:** **By default**") {
		t.Errorf("the fallback header must be the adaptive one, matching the hook's own behaviour on an unreadable file")
	}
	// Said out loud: silently rendering adaptive over a file that may well say
	// firm is the failure this whole issue is about.
	if !strings.Contains(res.stdout, "could not be read") {
		t.Errorf("the installer must SAY that the posture could not be read rather than quietly defaulting; got:\n%s", res.stdout)
	}
}
