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
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
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

	// TRL-97. Personal scope's next steps are its own: a plugin outside the
	// repository governs a project once that project has .trellis/rules.toml,
	// and the hook asks in a project that has none. They name the opt-out row
	// and, as the other two branches do, that a true row does not override it,
	// and nothing about strictness.
	for _, want := range []string{"asks whether to adopt Trellis", "<slug> = { active = false }", "a row set to active = true does not override it"} {
		if !strings.Contains(res.stdout, want) {
			t.Errorf("personal-scope next steps must say %q; got:\n%s", want, res.stdout)
		}
	}
	if strings.Contains(res.stdout, "strictness") {
		t.Errorf("personal-scope next steps mention strictness, which selects nothing since TRL-97; got:\n%s", res.stdout)
	}

	lines := strings.Split(strings.TrimRight(res.stdout, "\n"), "\n")
	const wantLines = 8 // item 1 (scope + vendored-to + files-written + rules = 4
	// lines) + blank separator + item 5 (3-line next-step pointer, rewritten for
	// personal scope by TRL-97) = 8; items 2-4 never fire for personal scope. See
	// install.sh's post-write block (guarded by `[ "$scope" = "project" ]`) for
	// the source of this count.
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
	// TRL-97, KTD9. The seed holds only what has effect: no strictness, no
	// seeded_from, and no row set active = true, since a rule with no row
	// applies. The next steps say every rule is active and show the one row form
	// that does something. That sentence states the rule count in words, and
	// TestRowCountProseSitesFollowThePin pins the numeral, so this needle does
	// not repeat it.
	seeded := readFileT(t, filepath.Join(repo, ".trellis", "rules.toml"))
	for _, forbidden := range []string{"strictness", "seeded_from", "active = true"} {
		if strings.Contains(seeded, forbidden) {
			t.Errorf("the seeded .trellis/rules.toml carries %q, which has no effect:\n%s", forbidden, seeded)
		}
	}
	for _, want := range []string{"rules are active", "<slug> = { active = false }"} {
		if !strings.Contains(res.stdout, want) {
			t.Errorf("item 5: the next steps must say %q; got:\n%s", want, res.stdout)
		}
	}
	if !strings.Contains(res.stdout, "add .claude/skills/trellis .claude/rules/trellis.md .trellis/rules.toml") {
		t.Errorf("item 4: the commit suggestion must name the seeded file this run wrote; got:\n%s", res.stdout)
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
	if got := readFileT(t, seeded); got != hookAcceptSeed(t) {
		t.Errorf(".trellis/rules.toml must be the fixed two-line seed byte for byte, the file the hook's accept instruction quotes (TRL-97) — a composed or edited seed would be decision logic, which AC2 still forbids; got:\n%q", got)
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
	// One header ships (TRL-97), and every project receives its By default
	// sentence. TestVendorRendersTheOneShippedHeaderWhateverStrictness covers a
	// project whose file still says strictness = "firm".
	if !strings.Contains(got, "**How strictly to follow them:** **By default**") {
		t.Errorf("missing the shipped header's By default sentence")
	}

	// --- ordering. The authority header states the rows are "loaded below the
	// rules"; emitting the import above them makes the shipped text lie.
	iBody := strings.Index(got, "inv-directional-flow")
	iImport := strings.Index(got, "@../../.trellis/rules.toml")
	if iBody < 0 || iImport < 0 || iImport < iBody {
		t.Errorf("the import must come AFTER the rules body (the authority header says rows load below the rules)")
	}

	// --- the footer is framing only (TRL-97, KTD2). The activation rule lives
	// once, in rules.md, which the body above carries; the footer adds its
	// marker, the activation heading and the import, and no second authority
	// sentence. The one it used to carry named `strictness` as authoritative over
	// a posture sentence that no longer varies. Pinned as the file's exact tail,
	// so a sentence added anywhere between the marker and the stamp fails here.
	wantFooter := "<!-- trellis:rendered-footer -->\n\n## Project rule activation\n\n@../../.trellis/rules.toml\n\n<!-- trellis:rendered-from " + strings.TrimSpace(files["version"]) + " -->\n"
	if !strings.HasSuffix(got, wantFooter) || strings.Count(got, "<!-- trellis:rendered-footer -->") != 1 {
		t.Errorf("the rendered footer must be exactly the marker, the activation heading, the import and the stamp:\nwant tail:\n%s\ngot:\n%s", wantFooter, got)
	}
	if strings.Contains(got, "strictness") {
		t.Errorf("the rendered file names `strictness`, which selects nothing and changes no rule since TRL-97:\n%s", got)
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
	// A project that switched a rule off: the stand-down must not depend on the
	// seed's empty table.
	if err := os.WriteFile(filepath.Join(repo, ".trellis", "rules.toml"), []byte("[rules]\ninv-minimal-first = { active = false }\n"), 0o644); err != nil {
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
	// TRL-97, KTD10. With no rules.toml beside a legacy shape, the way forward is
	// to migrate and re-run, never to write the file by hand: the legacy shape's
	// frozen text applies a rule only when a row says active = true, so the
	// two-line file this installer seeds would govern only the floors there. The
	// output may not quote that file as something to write, and it may not name
	// a preset that no longer ships.
	for _, want := range []string{"Do not write that file by hand", "re-run this installer"} {
		if !strings.Contains(res.stdout, want) {
			t.Errorf("the no-file next step for a legacy shape must say %q; got:\n%s", want, res.stdout)
		}
	}
	if seedComment, _, _ := strings.Cut(hookAcceptSeed(t), "\n"); strings.Contains(res.stdout, seedComment) {
		t.Errorf("the output quotes the seed file as something to write beside a legacy shape, where it governs only the floors:\n%s", res.stdout)
	}
	if all, own := strings.Count(combined, ".toml"), strings.Count(combined, ".trellis/rules.toml"); all != own {
		t.Errorf("the output names a .toml file other than .trellis/rules.toml:\n%s", combined)
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

	// THE COUNT NOW (TRL-97): install.sh reads THREE project files' contents —
	// the managed-block opening marker, the top-level `governed` key of an
	// existing .trellis/rules.toml, and CLAUDE.md's standalone @AGENTS.md import
	// line. The strictness read retired with the posture headers: one header
	// ships, so that key has nothing to select. What follows is how the count got
	// here, kept because each read's argument lives in it (decision-0090 D1).
	//
	// Before TRL-44 install.sh read EXACTLY THREE project files' contents: the
	// managed-block opening marker, and two keys of an existing .trellis/rules.toml
	// — `strictness` and the top-level `governed`. An earlier revision of this
	// subtest asserted zero, which was achieved by deleting the marker check and
	// regressed inline consumers into silent double delivery; it then pinned
	// exactly one, with the note that "a second content read has to come here and
	// argue itself". TRL-37 argued the second: the rendered file carries a posture
	// sentence, and taking it from a constant while the rows a few lines below say
	// `firm` put the adaptive header over a firm project — the plugin's own hook
	// reads the same key to pick the same header, so the render reads it too, with
	// the hook's own parser (a pair guard that retired with the read, TRL-97).
	// decision-0088 D5 records that read; the demand that a third "still has to
	// come and argue itself" is in that record's CONSEQUENCES, under "What is not
	// claimed here" — not in D5, which is about the second read and names this
	// test as one of the two things that answered the demand for it. (TRL-42 and
	// an earlier draft of decision-0090 both put the sentence in D5; a reviewer
	// caught it, and this comment carried the same error until decision-0090
	// fixed the pair.) This is that argument (TRL-38): a rules.toml holding
	// `governed = false` is a project saying Trellis does not govern here
	// (decision-0070 D5), and the hook reads that key before every delivery path
	// and injects nothing on it. The installer, not reading it, rendered the full
	// rules body into always-loaded context for that project — inverting the
	// opt-out, since the host loads the rendered file before any hook runs and the
	// hook can then only tell the session to disregard it. The third read
	// implements a decision already on main rather than making one: it only ever
	// REFUSES, with the hook's own matcher (TestInstallScriptGovernedParserMatchesHook),
	// selects nothing, and writes nothing. It still selects nothing from an
	// instructions file, and still writes nothing under .trellis/ beyond the seed.
	// Asserting the exact count keeps the widening bounded — a fourth read has to
	// come here and argue itself, and "here" is now ruled rather than assumed:
	// decision-0090 D1 makes this comment the home of the COUNT, changed by
	// argument in the same commit that changes it. D2 keeps the corpus for a read
	// that DECIDES — one that selects, patches, writes, or removes a refusal —
	// which is why the strictness read was owed decision-0088 and the governed
	// read was not.
	//
	// THE FOURTH READ CAME (TRL-44), and this is its argument, made here because
	// decision-0090 D1 puts the count here and D2 keeps only a read that DECIDES
	// in the corpus. It is the standalone `@AGENTS.md` import line of CLAUDE.md,
	// copied from the hook's own `claude_imports_agents` gate. Why it is owed:
	// install.sh probed CLAUDE.md and AGENTS.md alike for the managed-block
	// marker, then told an opted-out project that "Claude Code loads those at
	// launch over the opt-out" — of a file this host reads ONLY through that
	// import line (decision-0057). Measured on an opt-out plus a block in
	// AGENTS.md and no CLAUDE.md: the installer printed that warning and
	// `rules: LEFT IN PLACE` while the hook, which gates the same probe, emitted
	// nothing at all. Two components contradicting each other about the reader's
	// state is what decision-0073 D2 exists to end.
	//
	// Why it does not go to the corpus under D2: it selects no output artifact,
	// patches nothing, writes nothing, and removes no refusal. The render refusal
	// still runs off the ungated $static_conflict, so an un-imported AGENTS.md
	// block refuses exactly as before; the read changes which WARNING/NOTE is
	// printed and whether the shape joins the `LEFT IN PLACE` summary, nothing
	// else. Gating the refusal on it WOULD remove one, and that is D2's fourth
	// clause — a corpus question, deliberately left open rather than taken here.
	// The next read still has to come here and argue itself.
	t.Run("exactly three project file content reads: the marker check, the governed key and the @AGENTS.md gate", func(t *testing.T) {
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
		// A scripted command takes its PATTERN first and its FILES after, so the
		// pattern has to be identified and dropped before anything is counted.
		// Every command that can read a file, not just grep. The earlier version
		// keyed on `grep`, and a reviewer walked past it with
		//   sed -n '/inv-directional-flow/p' "$git_root/README.md"
		// which reads project content and left the whole suite green.
		//
		// TRL-45. The version before this one tokenised only the QUOTED runs of a
		// line and then took the first of them as the pattern. Two idiomatic
		// forms walked past it, both reproduced against this suite with an extra
		// read planted in install.sh and the whole suite still green:
		//   grep -q zzz "$git_root/CLAUDE.md"                     (bare pattern:
		//     the only quoted token is the FILE, so "first quoted = pattern"
		//     dropped the file and left nothing to count)
		//   sed -n "/^strictness/p" $git_root/.trellis/rules.toml  (bare path:
		//     an unquoted operand was never tokenised at all)
		// Both are what the shell accepts and what a hurried edit writes. So the
		// unit is now a WORD — quoted runs and bare runs, concatenated, which is
		// what the shell itself splits on — and quoting decides nothing about
		// whether a word is counted. What decides it is position (the pattern
		// slot) and expansion.
		//
		// A word is a read of pre-existing state when it EXPANDS a variable
		// outside single quotes — that is what a path built from this script's
		// variables looks like, and it drops literals, heredoc delimiters,
		// /dev/tty, and single-quoted `sed` scripts like '$d' whose `$` is inert
		// — and is not rooted at the staged bundle, the render temp file, or the
		// vendor target. Command names are matched as whole words, since a
		// substring match makes `chmod` look like `od`.
		//
		// WHAT THIS GUARD IS AND IS NOT, stated here because decision-0090 D1
		// makes this comment the count's home and #267 leans on the count to keep
		// the read widening bounded. It is still LEXICAL, and lexical analysis of
		// shell cannot be made sound: a word is only a command because of runtime
		// state this test never has. TWENTY-FOUR shapes were planted in install.sh
		// and re-run against this subtest. SEVENTEEN hidden reads are caught: a
		// bare pattern, a bare path, a path aliased through a variable quoted or
		// bare, a non-grep reader, an operand on a continuation line, a literal
		// operand ahead of the real one absolute or relative, a flag whose value
		// is a separate word (`tail -n 1 "$p"`, `cut -d : -f 2 "$p"`,
		// `sort -k 2 "$p"`, `head -n 3 "$p"`), a pattern whose literal text is
		// itself a command name (`grep -q sed "$p"`), a second reader after `||`,
		// a read sharing a line with a `>` write, and two stdin redirects. FOUR
		// more are correctly NOT counted — `>`, `>>`, `1>` and a glued `>"$p"`
		// write target, which an earlier revision counted as reads (an OVERcount,
		// the one failure mode needing no eval or indirection). Eight of the
		// twenty-four were found by review of this PR, four of them defects this
		// scan's own revisions introduced. THREE ARE NOT CAUGHT, and no amount of
		// further regex will catch them:
		//   eval 'grep -q zzz "$git_root/CLAUDE.md"'   (the command is a string)
		//   rd=grep; $rd -q zzz "$git_root/CLAUDE.md"  (the command is a value)
		//   probe() { grep -q "$1" "$2"; }; probe zzz "$git_root/CLAUDE.md"
		//                                              (the call site names no
		//                                               command and the body no
		//                                               path)
		// So this subtest raises the cost of a laundered read and states its own
		// ceiling; it is not proof that install.sh reads exactly three files. The
		// instrument that WOULD prove it is behavioural, not lexical: run the
		// installer against a scratch repo under an open(2) audit (strace, an
		// LD_PRELOAD shim) and count the opens whose path is under the repo root.
		// Expansion, aliasing, eval and indirection are all resolved by the
		// kernel before the count is taken. That is a real piece of test
		// infrastructure with its own portability question in CI, so it is
		// proposed rather than smuggled in here (inv-minimal-first); the
		// behavioural fixtures elsewhere in this file are what actually establish
		// what the script reads.
		// A reader command counts only at a COMMAND POSITION: the start of the
		// line, or after `; | & ( ) { $(`, with any `VAR=value` prefixes between.
		// Review of this PR found the earlier `[\s;|(&]` separator matching
		// command words inside English message strings — `could not be read
		// (permissions?)`, `the header tail produced nothing` — which is why the
		// first version of this scan had to stop at the first plain literal
		// operand to avoid counting a `$var` that was being PRINTED. That
		// stopping rule was the wrong instrument and cost two shapes the old
		// guard caught (below); prose is excluded here instead, at the match.
		reader := regexp.MustCompile(`(^|[;|&(){]|\$\()[ \t]*([A-Za-z_][A-Za-z0-9_]*=[^ \t]*[ \t]+)*(grep|sed|awk|cat|head|tail|wc|cut|tr|sort|od|read)[ \t]|<\s*['"$]`)
		// Quoted runs and bare runs, concatenated: `"$git_root"/CLAUDE.md` is one
		// word to the shell and must be one word here too.
		word := regexp.MustCompile(`(?:'[^']*'|"[^"]*"|[^\s'"]+)+`)
		single := regexp.MustCompile(`'[^']*'`)
		expands := regexp.MustCompile(`\$\{?[A-Za-z_]`)
		assign := regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)
		// `<` but not `<<` (a heredoc, skipped above) and not the `<!--` inside
		// the marker pattern, which the operand test rejects as a literal anyway.
		redirect := regexp.MustCompile(`(^|[^<])<[ \t]*([^\s<>|;&()]+)`)
		// An output redirect, optionally with a leading fd and optionally with the
		// target glued on: `>`, `>>`, `1>`, `>"$path"`. Group 1 is the glued
		// target, empty when the target is the next word.
		outRedirect := regexp.MustCompile(`^[0-9]*>>?(.*)$`)
		takesPattern := map[string]bool{"grep": true, "sed": true, "awk": true}
		// One word: is it a path built from this script's variables that names
		// pre-existing project state? Used for operands and redirect targets
		// alike, since the shell reads both the same way.
		isProjectRead := func(w string) bool {
			if !expands.MatchString(single.ReplaceAllString(w, "")) {
				return false
			}
			p := strings.ReplaceAll(strings.ReplaceAll(w, `"`, ""), "'", "")
			// The staged bundle, the render temp file, and the vendor target are
			// this script's own outputs, not the project's pre-existing state.
			return !strings.HasPrefix(p, "$stage/") && p != "$rendered_tmp" && !strings.HasPrefix(p, "$target/")
		}
		isCommand := map[string]bool{"grep": true, "sed": true, "awk": true, "cat": true,
			"head": true, "tail": true, "wc": true, "cut": true, "tr": true,
			"sort": true, "od": true, "read": true}
		// A trailing `\` continues the command, so the operand can sit on the next
		// line — install.sh already writes reads that way. Join first, or the
		// scan below sees a command with no file and a file with no command, and
		// counts neither.
		// `first` is what a failure REPORTS and what the three assertions below
		// match on: the physical line the command starts on, unchanged by the
		// join, so a continuation does not drag the next line's prose into an
		// assertion about this one's grep pattern.
		var lines []struct{ joined, first string }
		for _, line := range strings.Split(readFileT(t, installScriptPath(t)), "\n") {
			l := strings.TrimSpace(line)
			// NEVER fold into a comment. The shell does not continue a `#` line —
			// the comment ends at the newline and the next line is its own
			// statement — so joining there diverges from what actually runs.
			// Review of this PR: a comment ending in `\` (wrapped prose) swallowed
			// the command line after it into a string still starting with `#`,
			// which the skip below then dropped whole. Reproduced by putting such
			// a comment above the `governed` read: the count fell from four to
			// three and the read vanished from the scan.
			if n := len(lines); n > 0 && strings.HasSuffix(lines[n-1].joined, `\`) &&
				!strings.HasPrefix(lines[n-1].joined, "#") {
				lines[n-1].joined = strings.TrimSuffix(lines[n-1].joined, `\`) + " " + l
				continue
			}
			lines = append(lines, struct{ joined, first string }{l, l})
		}
		var reads []string
		for _, line := range lines {
			trimmed, reported := line.joined, line.first
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
			// One pass over the words. A command name arms or disarms the pattern
			// slot — a pipeline can hold several, and each stage gets its own —
			// flags are skipped, the armed pattern slot eats exactly one word, and
			// whatever is left is an operand.
			patternDue, hit, inReader, writeDue := false, false, true, false
			for _, raw := range word.FindAllString(cmd, -1) {
				// An OUTPUT redirect names a write target, not a read. Review of
				// this PR: `>` was stripped as punctuation with nothing recorded,
				// so the target after it fell through to the operand test and any
				// project-rooted path there was counted as a read — an OVERcount,
				// and the one failure mode that needs no eval or indirection, just
				// an ordinary `>` on a reader line. Reproduced with
				// `cat "$stage/manifest" > "$git_root/leftover"`: five reads.
				// Both spellings, since `>"$path"` is one word here and two to the
				// shell, and a leading fd (`1>`) is still a redirect.
				if m := outRedirect.FindStringSubmatch(raw); m != nil {
					if m[1] == "" {
						writeDue = true // the target is the next word
					}
					continue // a bare `>`, or the glued target itself
				}
				// Remaining redirection and pipeline punctuation glued to a word.
				w := strings.TrimLeft(raw, "<|;()&")
				if w == "" || assign.MatchString(w) { // `|`, `(`, or `LC_ALL=C`
					continue
				}
				if writeDue { // the target of the `>` just seen
					writeDue = false
					continue
				}
				if strings.HasPrefix(w, "-") { // a flag, never a file
					continue
				}
				// BEFORE the command-name lookup, not after: an armed pattern slot
				// consumes the very next word whatever it looks like. Review of
				// this PR found `grep -q sed "$git_root/CLAUDE.md"` — searching for
				// the literal string `sed` — re-arming the slot on its own pattern,
				// so the file after it was eaten as "the pattern" and never tested.
				if patternDue {
					patternDue = false // the pattern, whether or not it is quoted
					continue
				}
				// `;`, `||`, `&&` end the reader's operand list — what follows is
				// a different command, and its arguments are not files this reader
				// opens. A `|` does NOT: the next pipeline stage may itself be a
				// reader with a file. Without this, the scan ran off the end of
				// `grep -q … "$stage/…" || fail "…$header…"` and counted the
				// diagnostic's own text.
				if strings.Contains(raw, ";") || strings.Contains(raw, "||") || strings.Contains(raw, "&&") {
					inReader, patternDue = false, false
					continue
				}
				if bare := strings.Trim(w, `"'`); isCommand[bare] {
					inReader, patternDue = true, takesPattern[bare]
					continue
				}
				if !inReader {
					continue
				}
				// A literal operand is skipped, NOT a reason to stop scanning the
				// line. Stopping cost the guard two shapes the version before it
				// caught, both found by review of this PR: `cat missing "$path"`,
				// and any flag that takes a separate value — `tail -n 1 "$path"`,
				// `cut -d : -f 2 "$path"`, `sort -k 2 "$path"` — where the bare
				// value word sits between the command and its file.
				if isProjectRead(w) {
					hit = true
					break
				}
			}
			// A `<` redirect is a read whatever command it is attached to, and it
			// does not sit in the operand list at all: `done < "$git_root/x"`
			// closes a `while read` loop whose own words stop the scan above. So
			// every redirect operand on the line is tested independently.
			for _, m := range redirect.FindAllStringSubmatch(cmd, -1) {
				if isProjectRead(m[2]) {
					hit = true
					break
				}
			}
			if hit {
				reads = append(reads, reported)
			}
		}
		if len(reads) != 3 {
			t.Fatalf("install.sh makes %d content read(s) of a project file; exactly three are argued (the managed-block marker, the governed key of .trellis/rules.toml, and CLAUDE.md's @AGENTS.md import gate):\n%s", len(reads), strings.Join(reads, "\n"))
		}
		var marker, governed, importGate string
		for _, r := range reads {
			switch {
			case strings.Contains(r, `<!-- trellis:begin`):
				marker = r
			case strings.HasPrefix(r, `governed_head=`):
				governed = r
			case strings.Contains(r, `@AGENTS`):
				importGate = r
			}
		}
		// The governed read is the hook's own head-of-file sed, over the file as
		// an operand — the form the pair guard compares byte for byte.
		if governed == "" || !strings.Contains(governed, `sed "1s/^$bom//" "$git_root/.trellis/rules.toml"`) {
			t.Errorf("one permitted content read must be the hook's governed_head sed over \"$git_root/.trellis/rules.toml\"; reads were:\n%s", strings.Join(reads, "\n"))
		}
		if marker == "" || !strings.Contains(marker, `"^\(`) {
			t.Errorf("one permitted content read must be the column-0-anchored opening marker (optionally BOM-prefixed); reads were:\n%s", strings.Join(reads, "\n"))
		}
		if strings.Contains(marker, "trellis:end") {
			t.Errorf("the CLOSING marker grep is $-anchored and broke on CRLF checkouts — it must not come back: %s", marker)
		}
		// The import gate reads CLAUDE.md and nothing else, and reads it for ONE
		// line. Two properties, both load-bearing since TRL-48 made this read
		// decide the RENDER and not only the wording:
		//   - ANCHORED. An unanchored form is the defect the hook already fixed
		//     once: documentation ABOUT the import — prose, or a fenced example —
		//     read as the import itself, which puts the installer back to
		//     claiming this host loads a file it does not.
		//   - BOM-TOLERANT, `($bom)?`, exactly as the marker grep four lines down
		//     already is. A CLAUDE.md whose `@AGENTS.md` line is line 1 and whose
		//     encoding an editor rewrote on a Windows-default checkout otherwise
		//     reads as NOT importing — and the render then proceeds over a block
		//     the host does load. Reproduced against this script before the fix
		//     (BOM'd import + AGENTS.md block: rendered, and the NOTE asserted
		//     the host "never reads that file"); the hook carries the identical
		//     probe, so nothing reported it. Found by an automated reviewer on
		//     the PR that made this read gate the refusal, not by this guard,
		//     which is why the property is now asserted rather than assumed.
		if importGate == "" || !strings.Contains(importGate, `"^($bom)?[[:space:]]*@AGENTS\.md[[:space:]]*$"`) ||
			!strings.Contains(importGate, `"$git_root/CLAUDE.md"`) {
			t.Errorf("the third permitted content read must be the ANCHORED, BOM-tolerant standalone @AGENTS.md import line of \"$git_root/CLAUDE.md\"; reads were:\n%s", strings.Join(reads, "\n"))
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

// TRL-44. install.sh carries a comment enumerating every read it makes of
// pre-existing project state, with a headline count. That comment has been wrong
// three times — twice for claiming too little ("the ONE place this script reads
// .trellis/"; "no contents are read anywhere"), then for a stale count that TRL-38
// updated by adjusting rather than counting and left short of the read it had just
// added. Each correction was a fix to the instance; this is the fix to the class
// (inv-self-improvement), and it is the reason the enumeration now counts LINES:
// a unit a scan can reproduce, so the next read fails here in the commit that
// adds it instead of two releases later in a review. TRL-97 took the count from
// sixteen to fifteen by retiring the strictness read.
//
// This is deliberately NOT the paranoid classifier the content-read subtest uses.
// The stake is different: a laundered EXISTENCE test leaves a comment stale, not a
// script that reads a declared posture, and the content reads — the ones AC2
// actually bounds — are already scanned by operand over there. What this pins is
// that the number in the comment and the number in the code are the same number.
func TestInstallScriptReadEnumerationIsCounted(t *testing.T) {
	src := readFileT(t, installScriptPath(t))
	// The enumeration's own unit, transcribed: a line, outside a comment, that
	// tests or reads one of the project paths as "$git_root/…" or "$rules_dir/…".
	// Writes are excluded by construction — cp/mv/mkdir name their target as an
	// operand of a command this pattern does not match.
	site := regexp.MustCompile(`(\[ *!? *-[a-z] +"\$(git_root|rules_dir)/|(grep|sed|awk|cat|head|tail) [^|]*"\$(git_root|rules_dir)/|< *"\$(git_root|rules_dir)/)`)
	var sites []string
	for _, line := range strings.Split(src, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		if site.MatchString(line) {
			sites = append(sites, strings.TrimSpace(line))
		}
	}
	// The headline the comment states, in words, so the two cannot be edited
	// past each other by a digit.
	const want, stated = 15, "FIFTEEN LINES"
	if !strings.Contains(src, "pre-existing project state — "+stated) {
		t.Fatalf("the read enumeration's headline has moved or been reworded; this guard reads it verbatim and cannot check a count it cannot find (looked for %q)", stated)
	}
	if len(sites) != want {
		t.Errorf("install.sh reads pre-existing project state on %d lines; its enumeration says %s. Re-count and update the comment in the same commit — that is the whole point of the unit being a line:\n%s", len(sites), stated, strings.Join(sites, "\n"))
	}
	// The per-path tallies have to sum to the headline, or the enumeration is
	// internally inconsistent even when the total happens to be right — the shape
	// of the "existence x2" defect, which was a per-path count that had gone stale
	// under a headline that had not.
	tally := regexp.MustCompile(`(?m)^  #   - .*? (\d+)  \(`)
	sum := 0
	rows := tally.FindAllStringSubmatch(src, -1)
	for _, m := range rows {
		n, err := strconv.Atoi(m[1])
		if err != nil {
			t.Fatal(err)
		}
		sum += n
	}
	if len(rows) == 0 {
		t.Fatal("the read enumeration's per-path rows no longer parse; this guard checks that they sum to the headline")
	}
	if sum != want {
		t.Errorf("the read enumeration's per-path counts sum to %d across %d rows, but its headline says %s", sum, len(rows), stated)
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
	// them; TRL-37 argued a second read, the strictness key; TRL-38 a third, the
	// governed key; TRL-44 a fourth, CLAUDE.md's @AGENTS.md import line; TRL-97
	// retired the strictness read, leaving three. The
	// bounded version of that guard is the exactly-N-reads subtest in
	// TestVendorGuardsAddedByReviewAreActuallyPinned, which is where
	// decision-0090 D1 puts the count — "two" was this sentence's number for two
	// reads longer than it was true, which is the drift TRL-44 is about.
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
	block := readFileT(t, filepath.Join(vendoredBundleAbs(t), "reference", "block-inline.md"))
	if !strings.HasPrefix(block, "<!-- trellis:begin") {
		t.Fatalf("the shipped inline block no longer opens with the marker at column 0; this fixture and install.sh's grep both assume it does:\n%.120s", block)
	}
	if strings.Contains(block, "\n@") {
		t.Fatalf("the shipped inline block now carries an @-import; if the inline form gained a .trellis/ dependency, install.sh's existence checks may cover it and this fixture's premise needs re-deriving")
	}
	return block
}

// TRL-97. One header ships, reference/trellis.md, so the rendered file opens with
// its bytes whatever the project's file says. A file that still carries
// `strictness = "firm"` is the case worth running: this repository's own file
// does, and before TRL-97 that key selected a different header here. The expected
// head is the shipped payload above the @rules.md placeholder, not a literal, so
// a reword of the header cannot leave this green by luck. The project's file is
// its own and comes out byte-identical, and stdout says nothing about the key.
func TestVendorRendersTheOneShippedHeaderWhateverStrictness(t *testing.T) {
	src := payloadFiles()["trellis.md"]
	i := strings.Index(src, "\n@rules.md\n")
	if i < 0 {
		t.Fatalf("trellis.md carries no @rules.md placeholder line; the render's premise has drifted:\n%s", src)
	}
	head := src[:i+1]

	repo := t.TempDir()
	initGitRepo(t, repo)
	tomlPath := filepath.Join(repo, ".trellis", "rules.toml")
	const toml = "strictness  = \"firm\"\n\n[rules]\ninv-directional-flow = { active = true }\ninv-minimal-first    = { active = false }\nfloor-transparency   = { active = true }\n"
	writeFileT(t, tomlPath, toml)

	res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
	if res.code != 0 {
		t.Fatalf("exit %d\nstdout: %s\nstderr: %s", res.code, res.stdout, res.stderr)
	}
	got := readFileT(t, filepath.Join(repo, ".claude", "rules", "trellis.md"))
	if want := "<!-- trellis:rendered-begin -->\n" + head; !strings.HasPrefix(got, want) {
		t.Errorf("the rendered file does not open with the shipped trellis.md head:\nwant prefix:\n%s\ngot:\n%.600s", want, got)
	}
	if after := readFileT(t, tomlPath); after != toml {
		t.Errorf(".trellis/rules.toml was modified by the install; an existing file is the project's own and is never rewritten:\nbefore:\n%s\nafter:\n%s", toml, after)
	}
	for _, forbidden := range []string{"strictness", "posture"} {
		if strings.Contains(res.stdout, forbidden) {
			t.Errorf("stdout mentions %q; strictness selects nothing since TRL-97, and the install says nothing about it:\n%s", forbidden, res.stdout)
		}
	}
}

// hookAcceptSeed returns the file the Claude hook's TRELLIS_NOT_YET_GOVERNING
// accept instruction tells the agent to write: two backticked spans, one per
// line, each ending in a newline (TRL-97, KTD9). The spans sit inside a
// double-quoted shell string, so a byte the shell would unescape there (a
// backslash, `$`, a double quote, a backtick) would make the source differ from
// what the hook emits. The helper refuses such a span rather than unescaping it.
func hookAcceptSeed(t *testing.T) string {
	t.Helper()
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	var accept []string
	for _, line := range strings.Split(readFileT(t, hook), "\n") {
		if strings.Contains(line, `emit "TRELLIS_NOT_YET_GOVERNING`) {
			accept = append(accept, line)
		}
	}
	if len(accept) != 1 {
		t.Fatalf("expected exactly one TRELLIS_NOT_YET_GOVERNING emit in the hook, found %d; this guard's premise moved", len(accept))
	}
	const bt = "\\`" // an escaped backtick, as the shell source spells it
	lead := "If they ACCEPT, write $root/.trellis/rules.toml containing exactly these two lines, each ending in a newline: " + bt
	_, rest, ok := strings.Cut(accept[0], lead)
	if !ok {
		t.Fatalf("the hook's accept instruction no longer reads %q; this guard cannot find the file it quotes:\n%s", lead, accept[0])
	}
	first, rest, ok1 := strings.Cut(rest, bt+" then "+bt)
	second, _, ok2 := strings.Cut(rest, bt)
	if !ok1 || !ok2 {
		t.Fatalf("the accept instruction's two backticked lines no longer parse:\n%s", accept[0])
	}
	for _, l := range []string{first, second} {
		if l == "" || strings.ContainsAny(l, "\\$\"`") {
			t.Fatalf("a quoted line of the accept instruction is empty or carries a byte the shell would unescape (%q); this guard compares source bytes and cannot", l)
		}
	}
	return first + "\n" + second + "\n"
}

// TRL-97, KTD9, decision-0028: a guard per pair. install.sh seeds a missing
// .trellis/rules.toml, and the hook's accept instruction tells an agent to write
// the same file when a project accepts a user-scope install. Two writers of one
// file drift unless something fails when they do. This runs the installer and
// compares what it actually wrote, byte for byte, with what the hook quotes.
func TestInstallScriptSeedMatchesTheHooksAcceptInstruction(t *testing.T) {
	want := hookAcceptSeed(t)
	repo := t.TempDir()
	initGitRepo(t, repo)
	res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
	if res.code != 0 {
		t.Fatalf("exit %d\nstdout: %s\nstderr: %s", res.code, res.stdout, res.stderr)
	}
	if got := readFileT(t, filepath.Join(repo, ".trellis", "rules.toml")); got != want {
		t.Errorf("install.sh's seed differs from the file the hook's accept instruction quotes:\nhook:       %q\ninstall.sh: %q", want, got)
	}
}

// TRL-44, with the "does not abort" half of an older test folded in (TRL-97).
//
// An unreadable .trellis/rules.toml must not abort the install. Before TRL-97 the
// strictness read redirected that file into awk, and under dash the failed
// redirect killed the script after the bundle was already in place, with no rules
// file, no seed, and only the shell's own "cannot open" as the diagnosis. That
// read is gone; the exit-code assertion stays so that a later read of this file
// cannot bring the abort back. decision-0088 D3 rules that the installer renders
// anyway and says so, and records that as a DIVERGENCE from the hook, which emits
// TRELLIS_RULES_NOT_LOADED and injects nothing on the same input.
//
// What the message must name is the opt-out. The `governed` read is guarded
// regular-and-readable, so its head comes back empty and $opted_out stays `no`: a
// file declaring `governed = false` at mode 000 is rendered over, and the project
// is then governed by the rules file it declined. This test does not contest that
// behaviour, since an install that stops halfway is the worse state. It pins the
// DISCLOSURE: no opt-out in the file could be honoured and every rule applies,
// said on stdout, because the install is the one moment anyone reads this output
// (floor-transparency). The message used to lead with a posture that could not be
// honoured; no posture is read any more, so it must not mention one.
//
// Skipped when the test process can read a mode-000 file anyway (root, and CI
// images that run as it) — the fixture cannot exist there.
func TestVendorUnreadableRulesTomlSaysTheOptOutMayHaveBeenMissed(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: a mode-000 file is still readable, so the fixture cannot be built")
	}
	repo := t.TempDir()
	initGitRepo(t, repo)
	toml := filepath.Join(repo, ".trellis", "rules.toml")
	writeFileT(t, toml, "strictness = \"firm\"\ngoverned = false\n")
	if err := os.Chmod(toml, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(toml, 0o644) })
	res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
	if res.code != 0 {
		t.Fatalf("an unreadable rules.toml aborted the install (exit %d) instead of rendering and saying so (decision-0088 D3)\nstdout: %s\nstderr: %s", res.code, res.stdout, res.stderr)
	}
	// The behaviour D3 chose, asserted so that a later "fix" to the wording
	// cannot quietly become a change to what happens.
	rendered := filepath.Join(repo, ".claude", "rules", "trellis.md")
	if !strings.Contains(readFileT(t, rendered), "inv-directional-flow") {
		t.Fatalf("decision-0088 D3 renders on an unreadable rules.toml rather than stopping halfway; that is not what happened, and this test is about the message, not a behaviour change")
	}
	// The opt-out file itself is never rewritten, unreadable or not — checked by
	// mode rather than by content, since reading it back would need the chmod
	// undone and would then not be testing this fixture.
	if fi, err := os.Stat(toml); err != nil || fi.Mode().Perm() != 0 {
		t.Fatalf("the fixture file was replaced or its mode changed by the install (err=%v); the project's own rules.toml is never written: %v", err, fi)
	}
	for _, want := range []string{"could not be read", "any opt-out", "governed = false", "governed by the rules file it declined", "every rule applies", "TRELLIS_RULES_NOT_LOADED"} {
		if !strings.Contains(res.stdout, want) {
			t.Errorf("the installer must say the file could not be read, that no opt-out in it was honoured and every rule applies, and what the hook does instead (looked for %q); got:\n%s", want, res.stdout)
		}
	}
	// The hook does not fall back this way (decision-0088 D3), and nothing about
	// this file selects a posture any more.
	for _, forbidden := range []string{"falls back the same way", "posture"} {
		if strings.Contains(res.stdout, forbidden) {
			t.Errorf("stdout carries %q on an unreadable rules.toml:\n%s", forbidden, res.stdout)
		}
	}
	// NOT a count. TRL-44's report said such a project is "governed 16/16". Under
	// TRL-97 a rule with no row applies, so an import that loads nothing leaves
	// every rule applying, and the message says exactly that; a number would go
	// stale with the next catalog change and adds nothing the sentence lacks.
	if strings.Contains(res.stdout, "sixteen rules") {
		t.Errorf("the message claims a rule count; say what governs, not how much:\n%s", res.stdout)
	}
}

// TRL-38. A .trellis/rules.toml holding `governed = false` is a project saying
// Trellis does not govern here (decision-0070 D5). The hook reads that key
// before every delivery path and injects nothing on it, floors included. The
// installer did not read it: it rendered .claude/rules/trellis.md anyway, the
// host loaded that file at launch with no hook involved, and the hook — the
// component that honours the opt-out — was left injecting an override telling
// the session to disregard what was already read. The opt-out was not ignored;
// the thing that respects it was silenced by the thing that ignored it.
//
// The two deliveries have to agree on such a project, and the only way to know
// is to run both: install, then run the hook from the bundle the install just
// vendored. Neither may deliver a rule. Verified by mutation: with the opt-out
// branch removed from install.sh, the first subtest fails on the rendered file.
func TestVendorRefusesToRenderOverGovernedFalseOptOut(t *testing.T) {
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	runHook := func(t *testing.T, repo string) string {
		t.Helper()
		cmd := exec.Command(hook)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+repo,
			"CLAUDE_PLUGIN_ROOT="+filepath.Join(repo, ".claude", "skills", "trellis"))
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("hook exited non-zero: %v: %s", err, out)
		}
		return string(out)
	}
	assertNoRuleDelivered := func(t *testing.T, what, text string) {
		t.Helper()
		for _, s := range []string{"inv-directional-flow", "floor-transparency", "**How strictly to follow them:**"} {
			if strings.Contains(text, s) {
				t.Errorf("%s delivers a rule (%q) to a project that declared governed = false:\n%s", what, s, text)
			}
		}
	}

	// Every shape the hook's matcher opts out on: bare, BOM'd, commented,
	// indented. Each one must refuse the render, leave the bundle vendored and
	// the opt-out untouched, and leave the hook with nothing to say — no rules
	// file, no overlay, no inline block, so the hook's own opt-out branch emits
	// nothing at all. Both deliveries: nothing.
	for _, tc := range []struct{ name, toml string }{
		{"bare", "governed = false\n"},
		{"BOM'd, commented and indented", "\ufeff  governed = false  # declined on 2026-09-03\n"},
		{"opt-out ahead of stray rows", "governed = false\n\n[rules]\ninv-directional-flow = { active = true }\n"},
	} {
		t.Run("refuses over "+tc.name, func(t *testing.T) {
			repo := t.TempDir()
			initGitRepo(t, repo)
			toml := filepath.Join(repo, ".trellis", "rules.toml")
			writeFileT(t, toml, tc.toml)
			res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
			if res.code != 0 {
				t.Fatalf("the opt-out must refuse the render, not the install (exit %d):\nstdout: %s\nstderr: %s", res.code, res.stdout, res.stderr)
			}
			if _, err := os.Stat(filepath.Join(repo, ".claude", "rules", "trellis.md")); !os.IsNotExist(err) {
				t.Errorf(".claude/rules/trellis.md was rendered over a governed = false opt-out (stat err=%v); the hook injects nothing for this project and the installer must deliver nothing too", err)
			}
			assertBundleVendored(t, filepath.Join(repo, ".claude", "skills", "trellis"))
			if got := readFileT(t, toml); got != tc.toml {
				t.Errorf(".trellis/rules.toml was modified by the install — the opt-out must survive the run verbatim:\nbefore:\n%q\nafter:\n%q", tc.toml, got)
			}
			// Said out loud (floor-transparency): what was refused, why, and the
			// remedy — which names the opt-out, not a file to delete.
			// The remedy names the file a project adopts with (TRL-97, KTD10): the
			// two lines the hook's accept instruction quotes, not a shipped preset.
			seedComment, _, _ := strings.Cut(hookAcceptSeed(t), "\n")
			for _, want := range []string{"NOT rendering .claude/rules/trellis.md", "governed = false", "decision-0070", seedComment, "active = false", "This project is NOT governed"} {
				if !strings.Contains(res.stdout, want) {
					t.Errorf("stdout must carry %q; got:\n%s", want, res.stdout)
				}
			}
			// Every .toml the output names is the project's own file. The presets
			// under the plugin's reference/ no longer ship, so naming one sends the
			// reader to a file that is not there.
			if all, own := strings.Count(res.stdout, ".toml"), strings.Count(res.stdout, ".trellis/rules.toml"); all != own {
				t.Errorf("stdout names a .toml file other than .trellis/rules.toml (%d of %d mentions):\n%s", all-own, all, res.stdout)
			}
			if strings.Contains(res.stdout, "posture") {
				t.Errorf("stdout talks about a posture on an opted-out project; nothing selects one:\n%s", res.stdout)
			}
			out := runHook(t, repo)
			assertNoRuleDelivered(t, "the hook", out)
			if strings.TrimSpace(out) != "" {
				t.Errorf("with nothing rendered and nothing vendored under .trellis/, the hook's opt-out branch has nothing to override and must emit nothing; got:\n%s", out)
			}
		})
	}

	// The narrowed divergence the hook records: `governed = false` UNDER [rules]
	// is not a top-level key, so the hook governs such a project normally — and
	// therefore so must the installer. `governed = true` is not an opt-out
	// either. Both render, and the hook (path B, with the rendered file moved
	// aside) injects rules.
	//
	// TRL-44. These are also inputs that carry the words `governed = false` past
	// the opt-out refusal, and a note the installer printed for them used to
	// answer with advice written for a project it could no longer see: "the hook
	// stands down entirely while this rendered file still governs — delete
	// .claude/rules/trellis.md". Both halves are false here (measured: the hook
	// injects the full rule set on the first fixture), and deleting the rendered
	// file would leave the project MORE governed, not less. That note retired
	// with the posture notes (TRL-97); the forbidden claims below stay pinned so
	// no message brings them back.
	// TRL-45. The THIRD fixture is the exactly-one-key condition itself, and it
	// was the gap: TestInstallScriptGovernedParserMatchesHook pins three lines of
	// the matcher — the head sed, the count, and the false test — but NOT the
	// `[ "${governed_n:-0}" -eq 1 ]` that consumes the count, and no fixture fed
	// the installer a file with two top-level `governed` keys. Mutating that test
	// to `-ge 1` left the whole suite green while the installer began opting out
	// on `governed = false\ngoverned = true` and the hook — whose own count is
	// still `-eq 1` — went on rendering: the two-hosts-disagree class that the
	// comment above staleness.sh's own `governed_n` says the count exists to
	// prevent (cited by content, not by line: TRL-45's report cited the line
	// numbers and they had already drifted). Reproduced on a scratch repo,
	// installer and hook both run, before this fixture was written. A malformed
	// file is not an opt-out on either side; whichever key came first must not
	// decide it. It is also the shape the note above names but could not show:
	// the second top-level key that carries the words `governed = false` without
	// opting out.
	for _, tc := range []struct{ name, toml string }{
		{"misplaced under [rules]", "[rules]\ngoverned = false\ninv-directional-flow = { active = true }\n"},
		{"governed = true", "governed = true\n\n[rules]\ninv-directional-flow = { active = true }\ninv-minimal-first = { active = false }\n"},
		{"two top-level governed keys", "governed = false\ngoverned = true\n"},
	} {
		t.Run("renders for "+tc.name, func(t *testing.T) {
			repo := t.TempDir()
			initGitRepo(t, repo)
			writeFileT(t, filepath.Join(repo, ".trellis", "rules.toml"), tc.toml)
			res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
			if res.code != 0 {
				t.Fatalf("install failed: %s", res.stderr)
			}
			rendered := filepath.Join(repo, ".claude", "rules", "trellis.md")
			if !strings.Contains(readFileT(t, rendered), "inv-directional-flow") {
				t.Fatalf("the render was refused or truncated for a file that is not an opt-out")
			}
			if err := os.Remove(rendered); err != nil {
				t.Fatal(err)
			}
			if out := runHook(t, repo); !strings.Contains(out, "inv-directional-flow") {
				t.Errorf("the hook governs this project (its matcher does not opt out on this file) but the installer's disposition was checked above as a render — the two disagree if the hook injects nothing:\n%s", out)
			}
			// No message may describe the hook as standing down, nor advise
			// deleting the rendered file: the hook just injected the rule set for
			// this project, so both would be false, and the deletion would leave
			// it governed by the plugin instead of by a file it can edit. Matched
			// on the claim, not on the old sentence, so a reword cannot bring it
			// back under different words.
			for _, forbidden := range []string{"stands down entirely", "delete .claude/rules/trellis.md if that is not what you want"} {
				if strings.Contains(res.stdout, forbidden) {
					t.Errorf("stdout tells a project the hook GOVERNS that the hook stands down / that it should delete the rendered file (%q); measured on this same repo, the hook injects the full rule set:\n%s", forbidden, res.stdout)
				}
			}
		})
	}

	// An earlier run rendered the file and the opt-out came afterwards. The
	// installer cannot un-render any more than the hook can un-load: it leaves
	// the file, says the double state out loud, and names the remedy the hook
	// names (delete it, or /trellis:remove). The hook, for its part, emits its
	// disregard override — which is what the rendered file reduces it to.
	t.Run("a rendered file already present is left in place and warned about", func(t *testing.T) {
		repo := t.TempDir()
		initGitRepo(t, repo)
		writeFileT(t, filepath.Join(repo, ".trellis", "rules.toml"), "governed = false\n")
		rendered := filepath.Join(repo, ".claude", "rules", "trellis.md")
		const prior = "a previous render, contents irrelevant\n"
		writeFileT(t, rendered, prior)
		res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
		if res.code != 0 {
			t.Fatalf("install failed: %s", res.stderr)
		}
		if got := readFileT(t, rendered); got != prior {
			t.Errorf("the pre-existing rendered file was rewritten or removed over an opt-out; the installer neither created it nor may delete it:\n%s", got)
		}
		for _, want := range []string{"ALREADY EXISTS", "/trellis:remove", "rules: LEFT IN PLACE"} {
			if !strings.Contains(res.stdout, want) {
				t.Errorf("stdout must carry %q; got:\n%s", want, res.stdout)
			}
		}
		out := runHook(t, repo)
		assertNoRuleDelivered(t, "the hook", out)
		if !strings.Contains(out, "TRELLIS_NOT_GOVERNING") {
			t.Errorf("the hook must tell the session to disregard the file the host already loaded; got:\n%s", out)
		}
	})

	// Review of #267: the first version of the opt-out branch warned about a
	// pre-existing rendered file only. A legacy project carrying a static
	// overlay or an inline managed block beside its opt-out was told it was not
	// governed while the host loaded those rules at launch regardless, and got
	// no migration remedy — the shapes the hook's own TRELLIS_NOT_GOVERNING
	// override names. The installer must name what is loaded anyway, with the
	// paths to delete, and the hook must emit its override for the same shape.
	for _, tc := range []struct {
		name, wantShape, wantPath string
		plant                     func(t *testing.T, repo string)
	}{
		{"a vendored overlay", ".trellis/internal/ overlay", ".trellis/internal/",
			func(t *testing.T, repo string) {
				writeFileT(t, filepath.Join(repo, ".trellis", "internal", "version"), "payload@000000000000\n")
			}},
		{"an inline managed block", "managed block in CLAUDE.md", "the managed block in CLAUDE.md",
			func(t *testing.T, repo string) {
				writeFileT(t, filepath.Join(repo, "CLAUDE.md"), inlineBlockFixture(t))
			}},
		// TRL-44, the positive control for the import gate: a block in AGENTS.md
		// that CLAUDE.md DOES import is loaded at launch exactly like one in
		// CLAUDE.md, so the warning is true here and must still fire. Its twin —
		// the same block with no import — is the subtest below.
		{"an inline managed block in an IMPORTED AGENTS.md", "managed block in AGENTS.md", "the managed block in AGENTS.md",
			func(t *testing.T, repo string) {
				writeFileT(t, filepath.Join(repo, "AGENTS.md"), inlineBlockFixture(t))
				writeFileT(t, filepath.Join(repo, "CLAUDE.md"), "# Project\n\n@AGENTS.md\n")
			}},
	} {
		t.Run(tc.name+" already present is left in place and warned about", func(t *testing.T) {
			repo := t.TempDir()
			initGitRepo(t, repo)
			writeFileT(t, filepath.Join(repo, ".trellis", "rules.toml"), "governed = false\n")
			tc.plant(t, repo)
			res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
			if res.code != 0 {
				t.Fatalf("install failed: %s", res.stderr)
			}
			if _, err := os.Stat(filepath.Join(repo, ".claude", "rules", "trellis.md")); !os.IsNotExist(err) {
				t.Errorf("rendered over an opt-out (stat err=%v)", err)
			}
			for _, want := range []string{"STATICALLY (" + tc.wantShape + ")", tc.wantPath, "/trellis:remove", "rules: LEFT IN PLACE"} {
				if !strings.Contains(res.stdout, want) {
					t.Errorf("stdout must carry %q; got:\n%s", want, res.stdout)
				}
			}
			out := runHook(t, repo)
			assertNoRuleDelivered(t, "the hook", out)
			if !strings.Contains(out, "TRELLIS_NOT_GOVERNING") {
				t.Errorf("the hook must emit its disregard override for this shape; got:\n%s", out)
			}
		})
	}

	// TRL-44, the shape the warning above was wrong about and the one #267's own
	// fixture could not see, because it planted the block in CLAUDE.md only.
	// Claude Code reaches AGENTS.md ONLY through a standalone @AGENTS.md import
	// in CLAUDE.md (decision-0057), which is why the hook gates its probe on that
	// line. The installer did not, so on this repo it warned that "Claude Code
	// loads those at launch over the opt-out", reported rules: LEFT IN PLACE, and
	// offered a remedy that changes nothing for Claude — while the hook, on the
	// same repo, emitted nothing at all. Measured before the fix; both halves are
	// asserted here so neither can drift back alone.
	t.Run("a managed block in an UNIMPORTED AGENTS.md is not claimed to be loaded", func(t *testing.T) {
		repo := t.TempDir()
		initGitRepo(t, repo)
		writeFileT(t, filepath.Join(repo, ".trellis", "rules.toml"), "governed = false\n")
		writeFileT(t, filepath.Join(repo, "AGENTS.md"), inlineBlockFixture(t))
		res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
		if res.code != 0 {
			t.Fatalf("install failed: %s", res.stderr)
		}
		if _, err := os.Stat(filepath.Join(repo, ".claude", "rules", "trellis.md")); !os.IsNotExist(err) {
			t.Errorf("rendered over an opt-out (stat err=%v)", err)
		}
		// The false claim, and the summary line that followed from it. Nothing
		// this host reads is loaded here, so neither may appear.
		for _, forbidden := range []string{"Claude Code loads those at launch over the opt-out", "rules: LEFT IN PLACE"} {
			if strings.Contains(res.stdout, forbidden) {
				t.Errorf("stdout claims %q for a block in an AGENTS.md that no CLAUDE.md imports; this host never reads that file:\n%s", forbidden, res.stdout)
			}
		}
		// Still SAID, because the block is real for a host that reads AGENTS.md
		// directly, and because an opted-out project is owed the fact that its
		// opt-out does not travel. Silence would trade one wrong statement for a
		// missing one.
		for _, want := range []string{"managed block in AGENTS.md", "no standalone @AGENTS.md line", "Codex", "/trellis:remove"} {
			if !strings.Contains(res.stdout, want) {
				t.Errorf("stdout must still name the block and why it is inert for this host (%q); got:\n%s", want, res.stdout)
			}
		}
		// And the two components must now agree on this repo: the hook's gate
		// leaves it with nothing to override, so it says nothing.
		out := runHook(t, repo)
		assertNoRuleDelivered(t, "the hook", out)
		if strings.TrimSpace(out) != "" {
			t.Errorf("the hook's gate should leave nothing to override on this repo; got:\n%s", out)
		}
	})

	// Found by review OF THIS BRANCH, and it is the same defect one branch over:
	// the NOTE above is independent of the rendered-file warning, so both fire in
	// one run, and its first version generalised from its own condition to the
	// whole project — "nothing is loaded over the opt-out here", "the hook stays
	// silent on this project". With a pre-existing .claude/rules/trellis.md both
	// are false: that file IS loaded (the WARNING three lines up says so, and the
	// summary prints LEFT IN PLACE) and the hook emits TRELLIS_NOT_GOVERNING.
	// Reproduced before the reword. The subtest above could not see it because it
	// plants no rendered file, so this one does.
	t.Run("the unimported-AGENTS.md note claims nothing about a rendered file that IS loaded", func(t *testing.T) {
		repo := t.TempDir()
		initGitRepo(t, repo)
		writeFileT(t, filepath.Join(repo, ".trellis", "rules.toml"), "governed = false\n")
		writeFileT(t, filepath.Join(repo, "AGENTS.md"), inlineBlockFixture(t))
		writeFileT(t, filepath.Join(repo, ".claude", "rules", "trellis.md"), "a previous render, contents irrelevant\n")
		res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
		if res.code != 0 {
			t.Fatalf("install failed: %s", res.stderr)
		}
		// Both branches fire, which is the premise: without the first this test
		// is the one above with extra steps.
		for _, want := range []string{"ALREADY EXISTS", "no standalone @AGENTS.md line", "rules: LEFT IN PLACE"} {
			if !strings.Contains(res.stdout, want) {
				t.Fatalf("premise drifted — this run must print the rendered-file warning AND the AGENTS.md note (looked for %q):\n%s", want, res.stdout)
			}
		}
		// No claim in either may be about the PROJECT's state. These two were,
		// and the hook contradicts both on this repo.
		for _, forbidden := range []string{"nothing is loaded over the opt-out", "stays silent on this project"} {
			if strings.Contains(res.stdout, forbidden) {
				t.Errorf("the AGENTS.md note generalises %q to a project whose rendered file is loaded and whose hook emits TRELLIS_NOT_GOVERNING; every claim in that branch must be about the block:\n%s", forbidden, res.stdout)
			}
		}
		out := runHook(t, repo)
		assertNoRuleDelivered(t, "the hook", out)
		if !strings.Contains(out, "TRELLIS_NOT_GOVERNING") {
			t.Fatalf("premise drifted — the hook must override the rendered file here, which is what makes the withdrawn wording false:\n%s", out)
		}
	})

	// Review of #267: the governed read opened .trellis/rules.toml before any
	// `-f` check, so a FIFO at that path blocked the sed forever, ahead of the
	// non-regular handling the seed step already has. The read is now guarded
	// regular-and-readable. A FIFO is not an opt-out
	// and not a rules file: the installer must finish, render, and report the
	// seed as failed rather than hang. (The hook opens the same path unguarded;
	// that is the hook's own defect, outside this change.)
	t.Run("a FIFO at .trellis/rules.toml does not hang the opt-out read", func(t *testing.T) {
		repo := t.TempDir()
		initGitRepo(t, repo)
		mustMkdirAll(t, filepath.Join(repo, ".trellis"))
		if err := syscall.Mkfifo(filepath.Join(repo, ".trellis", "rules.toml"), 0o644); err != nil {
			t.Skipf("cannot create a FIFO here: %v", err)
		}
		cmd := exec.Command("/bin/sh", installScriptPath(t), "--non-interactive", "--scope", "project")
		cmd.Dir = repo
		cmd.Env = append(os.Environ(), "TRELLIS_BUNDLE_SOURCE="+vendoredBundleAbs(t))
		var so, se bytes.Buffer
		cmd.Stdout, cmd.Stderr = &so, &se
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("the installer failed on a FIFO rules.toml: %v\nstdout: %s\nstderr: %s", err, so.String(), se.String())
			}
		case <-time.After(30 * time.Second):
			_ = cmd.Process.Kill()
			t.Fatalf("the installer hung on a FIFO at .trellis/rules.toml — the opt-out read opened it before checking it is a regular file\nstdout so far: %s", so.String())
		}
		if !strings.Contains(so.String(), "Could not write .trellis/rules.toml") {
			t.Errorf("a non-regular rules.toml must be reported as a failed seed, not treated as rows or as an opt-out; got:\n%s", so.String())
		}
	})
}

// decision-0028: a guard per pair. The `governed` matcher — the regular-and-
// readable guard around the read (TRL-43: the empty init and the `-f && -r`
// test, since a copy that drops it reopens the FIFO hang), the head-of-file
// sed, the exactly-one count and the false test — is copied from the hook into
// install.sh because the hook ships inside the bundle the script vendors, so
// the two cannot share a file, and a copy
// drifts unless something fails when it does. The hook's root variable is
// `$root`; the script's is `$git_root`; nothing else may differ.
func TestInstallScriptGovernedParserMatchesHook(t *testing.T) {
	extract := func(path string) []string {
		t.Helper()
		var got []string
		for _, line := range strings.Split(readFileT(t, path), "\n") {
			l := strings.TrimSpace(line)
			switch {
			case l == `governed_head=""`,
				strings.HasPrefix(l, `if [ -f "$root/.trellis/rules.toml" ] && [ -r `),
				strings.HasPrefix(l, `if [ -f "$git_root/.trellis/rules.toml" ] && [ -r `),
				strings.HasPrefix(l, `governed_head="$(sed `),
				strings.HasPrefix(l, `governed_n="$(printf `),
				strings.HasPrefix(l, `printf '%s\n' "$governed_head" | LC_ALL=C grep -qE`):
				got = append(got, strings.ReplaceAll(l, `"$root/`, `"$git_root/`))
			}
		}
		if len(got) != 5 {
			t.Fatalf("%s: expected the five governed-matcher lines (empty init, regular-and-readable guard, head sed, count, false test), found %d:\n%s", path, len(got), strings.Join(got, "\n"))
		}
		return got
	}
	hookPath, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	h, s := extract(hookPath), extract(installScriptPath(t))
	for i := range h {
		if h[i] != s[i] {
			t.Errorf("install.sh's governed matcher has drifted from the hook's (plugins/trellis/hooks/staleness.sh); the two deliveries must read the opt-out identically.\nhook:      %s\ninstall.sh: %s", h[i], s[i])
		}
	}
}

// TRL-44, decision-0028 again: the third copied matcher gets the third guard.
// Whether Claude Code reads AGENTS.md is decided by one line in each script, and
// they must decide it the same way — the whole point of the read is that the
// installer stops asserting what the hook denies. As with the other two, the
// hook's root variable is `$root` and the script's is `$git_root`; nothing else
// may differ.
func TestInstallScriptAgentsImportGateMatchesHook(t *testing.T) {
	extract := func(path string) string {
		t.Helper()
		var got []string
		for _, line := range strings.Split(readFileT(t, path), "\n") {
			l := strings.TrimSpace(line)
			if strings.HasPrefix(l, `grep -qE "^($bom)?[[:space:]]*@AGENTS\.md[[:space:]]*$"`) {
				got = append(got, strings.ReplaceAll(l, `"$root/`, `"$git_root/`))
			}
		}
		if len(got) != 1 {
			t.Fatalf("%s: expected exactly one @AGENTS.md import-gate line, found %d:\n%s", path, len(got), strings.Join(got, "\n"))
		}
		return got[0]
	}
	hookPath, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	if h, s := extract(hookPath), extract(installScriptPath(t)); h != s {
		t.Errorf("install.sh's @AGENTS.md import gate has drifted from the hook's (plugins/trellis/hooks/staleness.sh); the two must agree on whether this host reads AGENTS.md at all.\nhook:      %s\ninstall.sh: %s", h, s)
	}
}

// TRL-48, decision-0091 D1. The gate TRL-44 wired to the MESSAGE now also gates
// the RENDER REFUSAL, and this is the test that pins the two deliveries to each
// other on the shape where they disagreed. Both are run on the SAME scratch repo,
// the way #273's fixtures do, because the defect was never visible in either one
// alone: install.sh refused with "this project already delivers the rules
// statically (managed block in AGENTS.md)" while staleness.sh on that same repo
// injected the full rule set. Measured on 700ca30 before the change.
//
// Both directions, because a gate that never fires and a gate that always fires
// are equally wrong: the un-imported block must RENDER, the imported one must
// still REFUSE, and a block in CLAUDE.md must be untouched by any of it.
func TestVendorAgentsBlockRefusalFollowsTheImportGate(t *testing.T) {
	hook, err := filepath.Abs("../plugins/trellis/hooks/staleness.sh")
	if err != nil {
		t.Fatal(err)
	}
	runHook := func(t *testing.T, repo string) string {
		t.Helper()
		cmd := exec.Command(hook)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(), "CLAUDE_PROJECT_DIR="+repo,
			"CLAUDE_PLUGIN_ROOT="+filepath.Join(repo, ".claude", "skills", "trellis"))
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("hook exited non-zero: %v: %s", err, out)
		}
		return string(out)
	}
	// The rule body actually reaching a session, by a string from the payload
	// rather than a landmark this test invents.
	const aRule = "inv-directional-flow"

	t.Run("an UNIMPORTED AGENTS.md block does not refuse the render", func(t *testing.T) {
		repo := t.TempDir()
		initGitRepo(t, repo)
		writeFileT(t, filepath.Join(repo, "AGENTS.md"), inlineBlockFixture(t))
		writeFileT(t, filepath.Join(repo, ".trellis", "rules.toml"), "# rows only\n")
		res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
		if res.code != 0 {
			t.Fatalf("exit %d: %s", res.code, res.stderr)
		}
		// The refusal, gone. Claude Code reaches AGENTS.md only through a
		// standalone @AGENTS.md line in CLAUDE.md (decision-0057), there is none
		// here, and .claude/rules/trellis.md is a file only Claude reads — so the
		// two cannot be one delivery doubled, which is what the refusal claimed.
		if _, err := os.Stat(filepath.Join(repo, ".claude", "rules", "trellis.md")); err != nil {
			t.Fatalf("the render was refused over a block in an AGENTS.md that no CLAUDE.md imports (stat err=%v); this host never reads that file, so there is no second delivery to double against — and the hook injects on this repo:\n%s", err, res.stdout)
		}
		if strings.Contains(res.stdout, "rules statically (managed block in AGENTS.md)") {
			t.Errorf("stdout still claims static delivery for a file this host does not read:\n%s", res.stdout)
		}
		// Removing a refusal silently is a silent decision. The block is real for
		// Codex, and the remedy the old refusal offered — delete it — would have
		// ungoverned that host, so the NOTE has to say which way round it is.
		for _, want := range []string{
			"NOTE: this project carries a Trellis managed block in AGENTS.md",
			"no standalone @AGENTS.md line",
			"Codex",
			"deleting the block would ungovern them",
			"TRELLIS_STATIC_SHAPES_CONFLICT",
			// The NOTE may not assert what neither component can know. An
			// @-import chain (CLAUDE.md -> foo.md -> @AGENTS.md) IS followed by
			// the host and is seen by neither anchored one-line probe, so on that
			// repo the block and the rendered file are the same rules twice —
			// and headless, where the plugin does not load (decision-0068), no
			// watcher reports it. Found by an automated reviewer on #276 and
			// measured: run 1 renders, run 2 renders again in silence.
			"CHECK THIS IF YOUR CLAUDE.md IMPORTS OTHER FILES",
		} {
			if !strings.Contains(res.stdout, want) {
				t.Errorf("the render path must SAY what it rendered over (%q); got:\n%s", want, res.stdout)
			}
		}
		// The withdrawn sentence, by name. It claimed a fact about the host that
		// this installer cannot establish from one anchored line.
		if strings.Contains(res.stdout, "so Claude Code never reads that file") {
			t.Errorf("the NOTE asserts the host never reads AGENTS.md; one anchored line in CLAUDE.md cannot establish that — an import chain reaching it is invisible to this probe:\n%s", res.stdout)
		}
		// And the two components now agree on this repo, which is the whole
		// point: the installer delivers, the hook stands down to what it
		// delivered. Neither is silent and neither contradicts the other.
		out := runHook(t, repo)
		if strings.Contains(out, aRule) {
			t.Errorf("the hook injected the rule set alongside the file the installer just rendered — that is double delivery, the state the refusal existed to prevent:\n%s", out)
		}
		if !strings.Contains(out, "already loaded from .claude/rules/trellis.md") {
			t.Errorf("the hook must stand down to the rendered file on this repo; got:\n%s", out)
		}
	})

	// The positive control, and the direction that fails if the gate is widened
	// into "never refuse over AGENTS.md". With the import line the host DOES load
	// that file, both deliveries would be live, and no hook can un-load either.
	t.Run("an IMPORTED AGENTS.md block still refuses the render", func(t *testing.T) {
		repo := t.TempDir()
		initGitRepo(t, repo)
		writeFileT(t, filepath.Join(repo, "AGENTS.md"), inlineBlockFixture(t))
		writeFileT(t, filepath.Join(repo, "CLAUDE.md"), "# Project\n\n@AGENTS.md\n")
		writeFileT(t, filepath.Join(repo, ".trellis", "rules.toml"), "# rows only\n")
		res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
		if res.code != 0 {
			t.Fatalf("exit %d: %s", res.code, res.stderr)
		}
		if _, err := os.Stat(filepath.Join(repo, ".claude", "rules", "trellis.md")); !os.IsNotExist(err) {
			t.Fatalf("rendered over a block in an AGENTS.md that CLAUDE.md DOES import (stat err=%v); the host loads both files, so this is live double delivery:\n%s", err, res.stdout)
		}
		if !strings.Contains(res.stdout, "rules statically (managed block in AGENTS.md)") {
			t.Errorf("the refusal must still name the shape it found; got:\n%s", res.stdout)
		}
		// The render-path NOTE belongs to the render path only. Printing it here
		// would tell an author their block was rendered over when nothing was.
		if strings.Contains(res.stdout, "file above was rendered anyway") {
			t.Errorf("the rendered-over NOTE fired on a run that rendered nothing:\n%s", res.stdout)
		}
		out := runHook(t, repo)
		if !strings.Contains(out, "TRELLIS_INLINE_BLOCK") {
			t.Errorf("the hook must refuse over the same block on the same repo; got:\n%s", out)
		}
	})

	// The gate reads CLAUDE.md and only CLAUDE.md. A block THERE is loaded by the
	// host whatever any import line says, so it must be unaffected — this is what
	// fails if the gate is applied to $static_conflict as a whole rather than to
	// the one shape the host may not read.
	t.Run("a block in CLAUDE.md is refused with or without the import line", func(t *testing.T) {
		for _, tc := range []struct{ name, claude string }{
			{"no import line", ""},
			{"with an import line", "\n@AGENTS.md\n"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				repo := t.TempDir()
				initGitRepo(t, repo)
				writeFileT(t, filepath.Join(repo, "CLAUDE.md"), inlineBlockFixture(t)+tc.claude)
				writeFileT(t, filepath.Join(repo, ".trellis", "rules.toml"), "# rows only\n")
				res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
				if res.code != 0 {
					t.Fatalf("exit %d: %s", res.code, res.stderr)
				}
				if _, err := os.Stat(filepath.Join(repo, ".claude", "rules", "trellis.md")); !os.IsNotExist(err) {
					t.Fatalf("rendered over a managed block in CLAUDE.md (stat err=%v); the host loads that file unconditionally:\n%s", err, res.stdout)
				}
			})
		}
	})

	// The two NOTEs for this shape are mutually exclusive and neither branch knows
	// about the other. An opted-out project renders nothing, so the render-path
	// NOTE would be telling its author about a file that does not exist, beside
	// the opt-out branch's own NOTE about the same block. TRL-44 built that
	// second one; this is the guard that stops the first from joining it.
	t.Run("an opted-out project gets the opt-out NOTE and not the rendered-over one", func(t *testing.T) {
		repo := t.TempDir()
		initGitRepo(t, repo)
		writeFileT(t, filepath.Join(repo, "AGENTS.md"), inlineBlockFixture(t))
		writeFileT(t, filepath.Join(repo, ".trellis", "rules.toml"), "governed = false\n")
		res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
		if res.code != 0 {
			t.Fatalf("exit %d: %s", res.code, res.stderr)
		}
		if _, err := os.Stat(filepath.Join(repo, ".claude", "rules", "trellis.md")); !os.IsNotExist(err) {
			t.Fatalf("premise drifted — the opt-out must still refuse the render (stat err=%v):\n%s", err, res.stdout)
		}
		if strings.Contains(res.stdout, "file above was rendered anyway") {
			t.Errorf("the rendered-over NOTE fired on an opted-out run, which renders nothing; that sentence is false the moment it is printed:\n%s", res.stdout)
		}
		if !strings.Contains(res.stdout, "no standalone @AGENTS.md line") {
			t.Errorf("the opt-out branch's own NOTE for this shape (TRL-44) must still fire; got:\n%s", res.stdout)
		}
	})

	// Found by an automated reviewer on the PR that made this read gate the
	// refusal, and it is the worst shape either component can be in: the import
	// grep was anchored `^[[:space:]]*@` with NO BOM tolerance, while the marker
	// grep four lines below it has been BOM-tolerant since #273 for the documented
	// reason — an editor on a Windows-default checkout rewrites the encoding of a
	// file whose first line is the thing being matched.
	//
	// Before TRL-48 that mis-selected a MESSAGE. After it, it selects the RENDER:
	// a CLAUDE.md that really does import AGENTS.md reads as not importing, the
	// refusal is skipped, and the host loads both the block and the rendered file.
	// Reproduced on this branch before the fix — rendered, with the NOTE asserting
	// "Claude Code never reads that file", which was false on that repo. And the
	// hook carries the byte-identical probe, so it stood down quietly and
	// TRELLIS_STATIC_SHAPES_CONFLICT never fired: silent and permanent, not the
	// transient false negative decision-0091 D1's argument is built on. Fixed in
	// both files in the same commit (decision-0028) and asserted here on the
	// installer and the hook together, because a fix in one is worth nothing.
	t.Run("a BOM'd @AGENTS.md import line is still an import line", func(t *testing.T) {
		repo := t.TempDir()
		initGitRepo(t, repo)
		writeFileT(t, filepath.Join(repo, "AGENTS.md"), inlineBlockFixture(t))
		writeFileT(t, filepath.Join(repo, "CLAUDE.md"), "\ufeff@AGENTS.md\n\n# Project\n")
		writeFileT(t, filepath.Join(repo, ".trellis", "rules.toml"), "# rows only\n")
		res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
		if res.code != 0 {
			t.Fatalf("exit %d: %s", res.code, res.stderr)
		}
		if _, err := os.Stat(filepath.Join(repo, ".claude", "rules", "trellis.md")); !os.IsNotExist(err) {
			t.Errorf("rendered over an AGENTS.md block whose import line is real and merely BOM'd (stat err=%v); the host loads both, and nothing downstream reports it:\n%s", err, res.stdout)
		}
		if strings.Contains(res.stdout, "never reads that file") {
			t.Errorf("the NOTE asserts this host does not read AGENTS.md, on a repo that imports it:\n%s", res.stdout)
		}
		// The hook's copy of the probe, on the same repo. A fix in one script and
		// not the other puts the two back to disagreeing, which is the whole
		// defect TRL-48 exists to close.
		out := runHook(t, repo)
		if !strings.Contains(out, "TRELLIS_INLINE_BLOCK") {
			t.Errorf("the hook must see the BOM'd import too and refuse over the block; got:\n%s", out)
		}
	})

	// The asymmetry that kept the refusal ungated, pinned as a fixture rather than
	// argued: this script WRITES, so its false negative is durable, while the hook
	// only injects. What answers it is that the durable state has a named watcher.
	// Render over an un-imported block, let the import line arrive afterwards, and
	// the hook — which re-reads that line every session — reports the conflict at
	// the next session start. If that ever stops being true, decision-0091 D1
	// loses the argument it was decided on, and this subtest is where that shows.
	t.Run("the import line arriving after the render is caught by the hook", func(t *testing.T) {
		repo := t.TempDir()
		initGitRepo(t, repo)
		writeFileT(t, filepath.Join(repo, "AGENTS.md"), inlineBlockFixture(t))
		writeFileT(t, filepath.Join(repo, ".trellis", "rules.toml"), "# rows only\n")
		res := runVendor(t, repo, "", vendoredBundleAbs(t), "--scope", "project")
		if res.code != 0 {
			t.Fatalf("exit %d: %s", res.code, res.stderr)
		}
		if _, err := os.Stat(filepath.Join(repo, ".claude", "rules", "trellis.md")); err != nil {
			t.Fatalf("premise drifted — this run must render (stat err=%v):\n%s", err, res.stdout)
		}
		writeFileT(t, filepath.Join(repo, "CLAUDE.md"), "# Project\n\n@AGENTS.md\n")
		out := runHook(t, repo)
		if !strings.Contains(out, "TRELLIS_STATIC_SHAPES_CONFLICT") {
			t.Fatalf("nothing reported the double delivery this render made possible once the import line arrived; that watcher is decision-0091 D1's answer to the write/inject asymmetry:\n%s", out)
		}
		if !strings.Contains(out, "AGENTS.md") {
			t.Errorf("the hook's report must name the file whose block is now loaded; got:\n%s", out)
		}
	})
}
