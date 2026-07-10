package main

// Tests for install.sh — the curl-path plugin vendor script (kodhama/trellis#124,
// corrected design per spec-0005-curl-install-mechanical-vendoring; see the script's
// own header for why it supersedes the closed #128 attempt). Unlike #128's
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
// pointer to /trellis:setup and must NOT carry the project-only trust-dialog note.
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
	if !strings.Contains(res.stdout, "/trellis:setup") {
		t.Errorf("stdout should carry the next-step pointer to /trellis:setup; got:\n%s", res.stdout)
	}
	if strings.Contains(res.stdout, "trust-dialog") || strings.Contains(res.stdout, "workspace-trust dialog") {
		t.Errorf("personal scope must NOT print the project-only trust-dialog note; got:\n%s", res.stdout)
	}
	if strings.Contains(res.stdout, "git add .claude/skills/trellis") {
		t.Errorf("personal scope must NOT print the project-only commit suggestion; got:\n%s", res.stdout)
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
	// item 5: the next-step pointer to /trellis:setup.
	if !strings.Contains(res.stdout, "/trellis:setup") {
		t.Errorf("item 5 (next step): stdout missing the /trellis:setup pointer; got:\n%s", res.stdout)
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
	claudeMD := "# Project A\n\n<!-- trellis:begin (managed by trellis) -->\nSome existing overlay content that a decision-logic script might try to detect or patch.\n<!-- trellis:end -->\n"
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

	// Fixture B: no .trellis/ was created, and AGENTS.md is untouched.
	if _, err := os.Stat(filepath.Join(repoB, ".trellis")); !os.IsNotExist(err) {
		t.Errorf("install.sh created .trellis/ in fixture B, which never had one — it must never touch .trellis/")
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
