package main

// Tests for install.sh — the curl-path plugin vendor script (kodhama/trellis#124,
// corrected design; see the script's own header for why it supersedes the closed
// #128 attempt). Unlike #128's install.sh, this script makes exactly one decision
// (scope) and composes nothing else — so these tests check vendoring mechanics
// (fetch, verify, write, scope resolution), never the setup skill's decision logic
// (that lives in plugins/trellis/skills/setup/SKILL.md and is out of scope here).
// The harness shape (exec the script against throwaway dirs, TRELLIS_BUNDLE_SOURCE
// pointed at the vendored bundle so tests run offline) is salvaged from #128's
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
// writes to $HOME/.claude/skills/trellis).
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
}

// TestVendorProjectScopeFreshInstallFromRoot (#124: project scope resolves to
// <repo-root>/.claude/skills/trellis when run from the root itself).
func TestVendorProjectScopeFreshInstallFromRoot(t *testing.T) {
	repo := t.TempDir()
	initGitRepo(t, repo)

	res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
	if res.code != 0 {
		t.Fatalf("expected success, got exit %d\nstdout: %s\nstderr: %s", res.code, res.stdout, res.stderr)
	}
	assertBundleVendored(t, filepath.Join(repo, ".claude", "skills", "trellis"))
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

// TestVendorDefaultFallsBackToPersonalOutsideGitRepo (#124: "if not inside a git
// repo, project scope isn't available — say so ... or default to personal with a
// clear message" — this script's resolved choice: fall back, loudly, when scope was
// not explicitly requested).
func TestVendorDefaultFallsBackToPersonalOutsideGitRepo(t *testing.T) {
	cwd := t.TempDir() // not a git repo
	home := t.TempDir()

	res := runVendor(t, cwd, home, vendoredBundleAbs(t)) // no --scope
	if res.code != 0 {
		t.Fatalf("expected success (fallback, not failure), got exit %d\nstdout: %s\nstderr: %s", res.code, res.stdout, res.stderr)
	}
	assertBundleVendored(t, filepath.Join(home, ".claude", "skills", "trellis"))
	if !strings.Contains(res.stdout, "not inside a git repository") {
		t.Errorf("the fallback must be stated loudly, not silent; stdout:\n%s", res.stdout)
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

// TestVendorScopeFromEnvVar ($TRELLIS_SCOPE is honored when no --scope flag is given).
func TestVendorScopeFromEnvVar(t *testing.T) {
	cwd := t.TempDir()
	home := t.TempDir()

	cmd := exec.Command("/bin/sh", installScriptPath(t), "--non-interactive")
	cmd.Dir = cwd
	cmd.Env = append(os.Environ(),
		"TRELLIS_BUNDLE_SOURCE="+vendoredBundleAbs(t),
		"HOME="+home,
		"TRELLIS_SCOPE=personal",
	)
	var so, se bytes.Buffer
	cmd.Stdout, cmd.Stderr = &so, &se
	if err := cmd.Run(); err != nil {
		t.Fatalf("expected success: %v (stderr: %s)", err, se.String())
	}
	assertBundleVendored(t, filepath.Join(home, ".claude", "skills", "trellis"))
	if !strings.Contains(so.String(), "$TRELLIS_SCOPE") {
		t.Errorf("stdout should attribute the scope to $TRELLIS_SCOPE; got:\n%s", so.String())
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
		"TRELLIS_SCOPE=personal",
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

// --- never a git mutation ----------------------------------------------------------

// TestVendorNeverRunsGitAdd (#124 rule 4: the script prints a suggested next
// command but never executes a git mutation itself — verified here by checking the
// repo's git status stays clean, i.e. the vendored files remain untracked).
func TestVendorNeverRunsGitAdd(t *testing.T) {
	repo := t.TempDir()
	initGitRepo(t, repo)

	if res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project"); res.code != 0 {
		t.Fatalf("run failed: %s", res.stderr)
	}
	cmd := exec.Command("git", "status", "--porcelain=v1")
	cmd.Dir = repo
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git status: %v", err)
	}
	status := string(out)
	if !strings.Contains(status, "?? .claude/") {
		t.Errorf("expected the vendored files to show as untracked (git never mutated), got status:\n%s", status)
	}
	if strings.Contains(status, "A  ") {
		t.Errorf("nothing should be staged — this script must never run git add; status:\n%s", status)
	}
}
