package main

import (
	"encoding/json"
	"html"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// docSurfaces returns the user-facing files whose claims must match the shipped
// product (decision-0025) — anything a consumer reads, or that is injected into
// their session, discovered by walking the tree.
//
// It used to be a hand-maintained list, and that list was found SHORT TWICE:
// docs/lp-content.md on 2026-07-31 (the LP's own source, the file most likely to
// be edited), then plugins/trellis/README.md, skills/remove/SKILL.md and
// cli/README.md on 2026-08-02 while decision-0072 retired /trellis:setup — where
// fixing only the listed files would have turned this guard GREEN with three
// surfaces still teaching a command the plugin no longer has. Two misses with
// the same shape make the list the defect, so it is discovered rather than
// written down. The highest-exposure surface was among the missing:
// hooks/staleness.sh emits slash commands straight into the consumer's session.
//
// The governance corpus is excluded on purpose: decisions, specs and research
// records legitimately name artifacts that later retired, and they are
// append-only, so a live-command check there would be permanently red.
func docSurfaces(t *testing.T) []string {
	t.Helper()
	out, err := docSurfacesIn(repoRoot)
	if err != nil {
		t.Fatalf("walking the repo for doc surfaces: %v", err)
	}
	if len(out) < 20 {
		t.Fatalf("doc-surface discovery found only %d files (%v) — the walk is broken, "+
			"and a guard that checks nothing passes silently", len(out), out)
	}
	return out
}

// docSurfacesIn is docSurfaces's walk with its root passed in rather than fixed
// at the repository, so TestGuardWalksDoNotReadOtherCheckouts can aim it at a
// fixture. The bug that test pins cannot be reproduced against the real tree
// from a worktree, which is where an agent stands (TRL-59).
func docSurfacesIn(root string) ([]string, error) {
	// EDITORIAL skips: "this file is not a doc surface whose claims must be
	// live." They answer *this* walk's question, which is not the question
	// marketplaceCommands asks, so the two walks do not share them — see
	// structuralSkip for the ones they do.
	skipDirs := map[string]bool{
		// .grove dropped by decision-0076, which deleted the directory.
		// .git moved to structuralSkip (TRL-59): it was never an editorial call.
		".github": true, ".claude": true,
		"decisions": true, "specs": true, "research": true, "eval": true,
		"fixtures": true, "testdata": true,
		// docs/superpowers/ (committed specs and plans) and .superpowers/
		// (working SDD state) are exempt on the SAME ground as the three
		// above, not by inheritance from specs/: a planning record
		// legitimately names artifacts that later retired — e.g.
		// /trellis:setup before decision-0072 retired it — and a retirement
		// must not force an edit to the record of the work that preceded it.
		// An earlier version of this comment argued succession from specs/
		// via decision-0079. decision-0085 shows that argument assumed 0079
		// permitted retention, which as written it did not; the exemption
		// stands, its stated ground does not.
		"superpowers": true, ".superpowers": true,
	}
	exts := map[string]bool{".md": true, ".html": true, ".sh": true, ".mjs": true}
	var out []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if structuralSkip(root, path) || skipDirs[d.Name()] {
				return fs.SkipDir
			}
			return nil
		}
		if exts[filepath.Ext(d.Name())] {
			out = append(out, path)
		}
		return nil
	})
	return out, err
}

// structuralSkip reports whether a walk rooted at root must not descend into the
// directory at path. STRUCTURAL, as against the editorial skips above: it asks
// "is what is in here part of this checkout at all?", and that is the same
// question for every guard walk in this package, so they all ask it here — the
// two in this file, and TestNoSurfaceMatrixFile's in surface_matrix_guard_test.go
// (TRL-63), whose own comment argues why decision-0066 AC6 permits it. Two walks
// disagreeing about it was the TRL-59 defect; two walks legitimately disagreeing
// about EDITORIAL scope is not, which is why only this half is shared.
//
// The nested-checkout rule is TRL-59 itself, and it is wider than the bug
// reported: skipping the literal string ".claude" would have fixed the finding
// and left a plain clone parked anywhere in the tree just as readable.
func structuralSkip(root, path string) bool {
	name := filepath.Base(path)
	return name == ".git" || name == "node_modules" || isSeparateCheckout(root, path)
}

// isSeparateCheckout reports whether dir holds a checkout of its own rather than
// more of the tree the walk started in. The test is the presence of a `.git`
// entry, not its kind: a worktree's is a FILE (`gitdir: …`) and a clone's is a
// DIRECTORY, and a rule that knew only one of them would miss the other.
//
// It is not exhaustive and does not try to be. A BARE repository carries no
// `.git` entry at all — and is conventionally named `foo.git`, which the name
// test above does not match — and neither does a checkout whose git directory
// was relocated with GIT_DIR. Neither shape has appeared in this tree, and
// neither holds the checked-out documentation these walks are looking for.
//
// An Lstat error that is not "no such file" — EACCES, EIO — is read as absence,
// so the walk descends. That is the safe direction here, and it was checked
// rather than assumed: a directory that is readable but not searchable fails the
// Lstat AND then fails the read of the content it did not skip, ending the walk
// LOUDLY (`walk error: open …/README.md: permission denied`). Resolving the
// other way — skipping whatever cannot be stat'd — would leave the guard quietly
// checking less and still passing, the one outcome this file refuses everywhere.
//
// The walk's own root is exempt, because it carries a `.git` entry too — without
// that exemption the very first callback skips the entire repository and every
// guard in this file passes while reading nothing.
func isSeparateCheckout(root, dir string) bool {
	if filepath.Clean(dir) == filepath.Clean(root) {
		return false
	}
	_, err := os.Lstat(filepath.Join(dir, ".git"))
	return err == nil
}

// proseAfterTrellis are lowercase words that legitimately follow "trellis" in prose
// (e.g. "a trellis is structure that enables growth") and are NOT commands. Extend
// this if the docs add more prose — that's the intended, low-friction escape hatch.
var proseAfterTrellis = map[string]bool{
	"is": true, "a": true, "an": true, "the": true, "and": true, "or": true,
	"on": true, "in": true, "to": true, "as": true, "for": true, "with": true,
	"that": true, "governance": true, "mark": true,
	// Added 2026-08-02 when docSurfaces became a walk of the whole tree: prose in
	// AGENTS.md, core/lexicon.md and the hook's own comments, not commands.
	"keeps": true, "expresses": true, "delivery": true,
}

// TestDocsClaimOnlyRealCommands enforces decision-0025: the docs must not advertise a
// `trellis <command>` the CLI does not have, nor a `/trellis:<skill>` the plugin lacks.
// This is decision-0020's "no claim without a rule behind it", generalized to the
// product surface — the guard that stops the docs drifting ahead of the code.
func TestDocsClaimOnlyRealCommands(t *testing.T) {
	cmds := commandNames()
	skills := pluginSkills(t)
	cmdRe := regexp.MustCompile(`trellis ([a-z][a-z-]+)`)
	skillRe := regexp.MustCompile(`/trellis:([a-z-]+)`)

	for _, f := range docSurfaces(t) {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("reading %s: %v", f, err)
		}
		text := string(b)

		for _, m := range cmdRe.FindAllStringSubmatch(text, -1) {
			word := m[1]
			if cmds[word] || proseAfterTrellis[word] {
				continue
			}
			t.Errorf("%s references `trellis %s`, which is not a real command. "+
				"Add the command, fix the docs, or (if it is prose) add %q to proseAfterTrellis.",
				f, word, word)
		}

		for _, m := range skillRe.FindAllStringSubmatch(text, -1) {
			if !skills[m[1]] {
				t.Errorf("%s references `/trellis:%s`, but the plugin has no such skill", f, m[1])
			}
		}
	}
}

// TestNoUnqualifiedSetupClaims: a Codex P2 on #227, and the THIRD time a guard on
// this branch was found to recognise only one spelling of the thing it guards.
//
// TestDocsClaimOnlyRealCommands matches `/trellis:<skill>`. After decision-0072
// the docs still said "Setup installs no receipt and no fallback" and "without
// it, setup reports bootstrap-only degradation" — present-tense promises about a
// skill that no longer exists, invisible to a guard keyed on the slash spelling.
// (Compare TestEveryDestructiveInstructionIsGated, which matched only "delete"
// and missed "drop".)
//
// The word is not banned: the retired binary's `setup` TUI, and the remove
// skill's cleanup of artifacts a PAST setup left behind, are both legitimate and
// historical. What must not appear is the bare word as a live actor. Each
// allowed form below names why it is allowed, so the exemption list is a record
// rather than a mute.
var setupQualifiers = []string{
	"the retired setup skill",  // decision-0072's own name for it
	"a past setup",             // remove skill: cleaning artifacts it left
	"past setup",               // same, mid-sentence
	"setup skill retired",      // README pointer at the decision
	"the setup skill",          // qualified reference to the artifact
	"`setup` TUI",              // the retired v0 BINARY, not the skill
	"setup\n  CLI",             // same, across a line wrap
	"environment setup script", // Claude Code cloud's own name for the pre-launch script (TRL-35), not the skill
	// "grove:setup" was exempted here as "a different product's skill". Dropped by
	// decision-0076: grove is retired and no doc surface may name /grove:setup as a
	// live command, so keeping the exemption would let one back in unnoticed.
}

func TestNoUnqualifiedSetupClaims(t *testing.T) {
	re := regexp.MustCompile(`(?i)\bsetup\b`)
	checked := 0
	for _, f := range docSurfaces(t) {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("reading %s: %v", f, err)
		}
		for _, line := range strings.Split(string(b), "\n") {
			if !re.MatchString(line) {
				continue
			}
			checked++
			qualified := false
			for _, q := range setupQualifiers {
				if strings.Contains(line, strings.ReplaceAll(q, "\\n  ", "")) {
					qualified = true
					break
				}
			}
			if !qualified {
				t.Errorf("%s says \"setup\" as a live actor; decision-0072 retired it. Qualify it "+
					"(\"a past setup\", \"the retired setup skill\") or delete the claim:\n  %s", f, strings.TrimSpace(line))
			}
		}
	}
	if checked == 0 {
		t.Fatal("no line mentioning setup was found at all — the scan is broken, not the docs")
	}
	t.Logf("checked %d lines mentioning setup", checked)
}

func pluginSkills(t *testing.T) map[string]bool {
	t.Helper()
	entries, err := os.ReadDir("../plugins/trellis/skills")
	if err != nil {
		t.Fatalf("reading plugin skills: %v", err)
	}
	skills := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() {
			skills[e.Name()] = true
		}
	}
	return skills
}

func readDocSurface(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(b)
}

// docs/index.html claims in its own header to be "Generated per kodhama/design-system's
// lp-generator.md contract … composed against this repo's own docs/lp-content.md".
// No generator exists: no build script, no workflow step, and nothing outside
// lp-content.md even references it. The two files are maintained BY HAND IN
// PARALLEL and had silently diverged.
//
// This asserts the install terminal is CONSISTENT four ways at once. An earlier
// version of this guard checked a hardcoded list of command strings and was
// weaker than it looked: it could not see a `data-tab` renamed without its
// `data-panel`, could not see a commands VALUE changed to point anywhere, and —
// worst — disarmed itself, because renaming a URL correctly in both files left a
// list entry matching neither, after which that entry silently checked nothing
// forever. The commit that introduced it shipped a drift it could not see.
//
// Comparing extracted structure instead of hardcoded strings has no list to go
// stale.
func TestInstallTerminalIsConsistentAcrossSourceRenderAndScript(t *testing.T) {
	page := readDocSurface(t, "../docs/index.html")
	source := readDocSurface(t, "../docs/lp-content.md")

	buttons := setOf(submatches(regexp.MustCompile(`data-tab="([^"]+)"`), page))
	panels := parseRenderedPanels(t, page)
	keys := parseCommandsObject(t, page)
	brief := parseBriefTabs(t, source)

	// 1. Every rendered tab button has a panel, and a copy-command key. A button
	//    with no key makes the copy button write the string "undefined"; a button
	//    with no panel activates nothing at all. Both have shipped.
	assertSameKeys(t, "data-tab buttons", buttons, "data-panel panels", setOf(keysOf(panels)))
	assertSameKeys(t, "data-tab buttons", buttons, "commands object keys", setOf(keysOf(keys)))

	// 2. The copy button must copy what the panel shows. Key PRESENCE is not
	//    enough — a key pointing at the wrong command is silent and worse than a
	//    missing one.
	for tab, cmds := range panels {
		want := strings.Join(runnable(cmds), "\n")
		if got := keys[tab]; got != want {
			t.Errorf("tab %q: the copy button copies\n  %q\nbut the panel shows\n  %q\nthe copy button must hand over what the user is looking at", tab, got, want)
		}
	}

	// 3. The content brief and the rendered page must agree, command for command.
	//    These are hand-maintained in parallel with no generator; this is the only
	//    thing that makes a one-sided edit visible.
	assertSameKeys(t, "docs/index.html panels", setOf(keysOf(panels)), "docs/lp-content.md tabs", setOf(keysOf(brief)))
	for tab, pageCmds := range panels {
		briefCmds, ok := brief[tab]
		if !ok {
			continue // already reported by assertSameKeys
		}
		if strings.Join(pageCmds, "\n") != strings.Join(briefCmds, "\n") {
			t.Errorf("tab %q differs between the LP source and the rendered page:\n  docs/lp-content.md:\n    %s\n  docs/index.html:\n    %s\nboth are hand-maintained; a change to one must be made in the other", tab, strings.Join(briefCmds, "\n    "), strings.Join(pageCmds, "\n    "))
		}
	}
}

func submatches(re *regexp.Regexp, s string) []string {
	var out []string
	for _, m := range re.FindAllStringSubmatch(s, -1) {
		out = append(out, m[1])
	}
	return out
}

func setOf(xs []string) map[string]bool {
	m := map[string]bool{}
	for _, x := range xs {
		m[x] = true
	}
	return m
}

func keysOf[V any](m map[string]V) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func assertSameKeys(t *testing.T, aName string, a map[string]bool, bName string, b map[string]bool) {
	t.Helper()
	for k := range a {
		if !b[k] {
			t.Errorf("%q appears in %s but not in %s", k, aName, bName)
		}
	}
	for k := range b {
		if !a[k] {
			t.Errorf("%q appears in %s but not in %s", k, bName, aName)
		}
	}
}

var tagRe = regexp.MustCompile(`<[^>]*>`)

// text strips markup and decodes entities, so a rendered command compares equal
// to the same command written as plain text in the brief.
func text(s string) string {
	return strings.TrimRight(html.UnescapeString(tagRe.ReplaceAllString(s, "")), " \t")
}

// runnable drops pure-comment lines. The copy button should hand over commands,
// not `#` annotations the user cannot execute.
func runnable(cmds []string) []string {
	var out []string
	for _, c := range cmds {
		body := strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(c), "$>"))
		if body == "" || strings.HasPrefix(body, "#") {
			continue // a pure annotation line — nothing to run
		}
		if i := strings.Index(body, "    #"); i >= 0 {
			body = strings.TrimRight(body[:i], " ") // trailing annotation
		}
		out = append(out, body)
	}
	return out
}

func parseRenderedPanels(t *testing.T, page string) map[string][]string {
	t.Helper()
	// `[^>]*` after data-panel, not `>`: the panels became real tabpanels in the
	// TRL-11 a11y pass and now carry role/id/aria-labelledby/tabindex after that
	// attribute. Pinning the `>` would have made this test fail closed on any
	// future attribute — which is the loud failure this file wants, but the fix
	// is to widen the match, not to drop the check. Everything it actually
	// asserts on (the data-panel key, the <code> bodies) is unchanged.
	re := regexp.MustCompile(`(?s)<div class="panel[^"]*" data-panel="([^"]+)"[^>]*>(.*?)</div>\s*(?:<div class="panel|</div>)`)
	ms := re.FindAllStringSubmatch(page, -1)
	if len(ms) == 0 {
		t.Fatal("no terminal panels found in docs/index.html — this test's premise has drifted")
	}
	out := map[string][]string{}
	code := regexp.MustCompile(`(?s)<code>(.*?)</code>`)
	for _, m := range ms {
		var cmds []string
		for _, c := range code.FindAllStringSubmatch(m[2], -1) {
			cmds = append(cmds, text(c[1]))
		}
		out[m[1]] = cmds
	}
	return out
}

func parseCommandsObject(t *testing.T, page string) map[string]string {
	t.Helper()
	body := regexp.MustCompile(`(?s)var commands = \{(.*?)\n\s*\};`).FindStringSubmatch(page)
	if body == nil {
		t.Fatal("could not find `var commands = {...};` in docs/index.html — if the copy-to-clipboard code was restructured, update this test in the same commit rather than letting it silently stop checking")
	}
	out := map[string]string{}
	for _, m := range regexp.MustCompile(`(?m)^\s*([A-Za-z0-9_-]+):\s*"((?:[^"\\]|\\.)*)"`).FindAllStringSubmatch(body[1], -1) {
		v := strings.ReplaceAll(m[2], `\n`, "\n")
		out[m[1]] = v
	}
	return out
}

func parseBriefTabs(t *testing.T, source string) map[string][]string {
	t.Helper()
	re := regexp.MustCompile("(?s)- `([a-z0-9_-]+)` \\(.*?\\):\n\\s*```\n(.*?)```")
	ms := re.FindAllStringSubmatch(source, -1)
	if len(ms) == 0 {
		t.Fatal("no install tabs found in docs/lp-content.md — this test's premise has drifted")
	}
	out := map[string][]string{}
	for _, m := range ms {
		var cmds []string
		for _, line := range strings.Split(m[2], "\n") {
			if s := strings.TrimSpace(line); s != "" {
				cmds = append(cmds, s)
			}
		}
		out[m[1]] = cmds
	}
	return out
}

// TestManualRecipeBranchesAreSeparatePastes guards decision-0073 Consequence 2
// (global review H1). The manual recipe's two delivery branches — the @import
// branch (overlay copies + block-claude.md) and the inline branch
// (block-inline-<p>.md, no overlay) — lived in ONE fenced sh block where both
// were live commands, so a wholesale paste produced overlay + import block +
// inline block: exactly the S2-plus-S4 conflict the record ordered the recipe
// to stop shipping, and the hook absorbs it silently (path A exits first on a
// current stamp — a reviewer built the state and ran the hook: silence). The
// branches must live in separate fenced blocks so no single paste can produce
// both.
func TestManualRecipeBranchesAreSeparatePastes(t *testing.T) {
	body, err := os.ReadFile("../README.md")
	if err != nil {
		t.Fatal(err)
	}
	// Column-0 fence delimiters only: the README also NAMES a ```toml fence
	// mid-sentence in its re-paste prose, which a naive split on backticks
	// would count as a boundary and silently misalign every block after it.
	var blocks []string
	var cur []string
	in := false
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(line, "```") {
			if in {
				blocks = append(blocks, strings.Join(cur, "\n"))
				cur = nil
			}
			in = !in
			continue
		}
		if in {
			cur = append(cur, line)
		}
	}
	var importBlocks, inlineBlocks []int
	for i, b := range blocks {
		if strings.Contains(b, "block-claude.md") {
			importBlocks = append(importBlocks, i)
		}
		if strings.Contains(b, "block-inline-<p>.md") {
			inlineBlocks = append(inlineBlocks, i)
		}
	}
	if len(importBlocks) == 0 || len(inlineBlocks) == 0 {
		t.Fatalf("cannot locate both recipe branches in fenced blocks (import: %v, inline: %v) — the guard would silently check nothing", importBlocks, inlineBlocks)
	}
	for _, a := range importBlocks {
		for _, b := range inlineBlocks {
			if a == b {
				t.Errorf("the @import and inline branches share fenced block %d — one wholesale paste produces the overlay AND both managed blocks, the S2-plus-S4 conflict decision-0073 Consequence 2 ordered this recipe to stop shipping", a)
			}
		}
	}
}

// TestMarketplaceAddNamesTheRepoThatServesIt (decision-0028 — a guard per pair).
// The documented install command, `/plugin marketplace add <owner/repo>`, is
// copied by hand into six live surfaces: both READMEs, the CLI's usage text,
// the LP source and twice in the rendered page. TRL-26 found all six naming a
// repository that does not resolve (`kodhama/kodhama`) while this repo's own
// `.claude/settings.json` resolved the `kodhama` marketplace to
// `kodhama/stewards` — the docs contradicted the config, and the config was
// right. TestInstallTerminalIsConsistentAcrossSourceRenderAndScript compares
// only the LP source with the page, so the other four could drift again with
// the suite green (review of #277).
//
// The canonical value is the one this repo resolves for itself, read from
// settings.json — not a literal repeated here, which would be a seventh copy.
// The surfaces are DISCOVERED by walking the tree for the command, the same
// lesson docSurfaces records above; the six known ones are a floor, so the
// command vanishing from one of them fails too. Corpus citations of
// kodhama-the-repo (`kodhama/kodhama-0007…`, `kodhama/kodhama#35`) never carry
// the words `marketplace add`, so the match is on the command, not the name.
// `\s+` between the command and the slug, not a literal space: the plugin
// README wraps one copy at exactly that point, and a literal space left that
// seventh copy invisible while the file's other two copies satisfied the floor
// (review of #277, verified by mutating the wrapped copy alone).
func TestMarketplaceAddNamesTheRepoThatServesIt(t *testing.T) {
	var settings struct {
		ExtraKnownMarketplaces map[string]struct {
			Source struct {
				Repo string `json:"repo"`
			} `json:"source"`
		} `json:"extraKnownMarketplaces"`
	}
	if err := json.Unmarshal([]byte(readDocSurface(t, "../.claude/settings.json")), &settings); err != nil {
		t.Fatalf("parsing .claude/settings.json: %v", err)
	}
	canonical := settings.ExtraKnownMarketplaces["kodhama"].Source.Repo
	if canonical == "" {
		t.Fatal(".claude/settings.json declares no extraKnownMarketplaces.kodhama.source.repo — the canonical marketplace repository this test checks every documented install command against")
	}

	found, err := marketplaceCommands(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, rel := range keysOf(found) {
		for _, slug := range found[rel] {
			if slug != canonical {
				t.Errorf("%s documents `marketplace add %s`, but this repo resolves the kodhama marketplace to %s (.claude/settings.json) — the six copies of the install command move together (decision-0028)", rel, slug, canonical)
			}
		}
	}
	for _, surface := range []string{"README.md", "plugins/trellis/README.md", "cli/main.go", "docs/index.html", "docs/lp-content.md"} {
		if len(found[surface]) == 0 {
			t.Errorf("%s no longer carries a `marketplace add` command — it is one of the live install surfaces this guard pins; if the command moved, update this list in the same change", surface)
		}
	}
}

// marketplaceCommands returns every marketplace slug the tree under root
// documents, keyed by path relative to root. Rooted where it is told, and
// separated from the assertions above, for the reason docSurfacesIn is.
func marketplaceCommands(root string) (map[string][]string, error) {
	command := regexp.MustCompile(`marketplace add\s+([A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+)`)
	textual := map[string]bool{".md": true, ".html": true, ".go": true, ".sh": true, ".json": true, ".toml": true, ".txt": true, ".yml": true, ".yaml": true, ".mjs": true}
	found := map[string][]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if structuralSkip(root, path) {
				return fs.SkipDir
			}
			return nil
		}
		if !textual[filepath.Ext(path)] {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		key := filepath.ToSlash(rel)
		for _, m := range command.FindAllStringSubmatch(string(b), -1) {
			found[key] = append(found[key], m[1])
		}
		return nil
	})
	return found, err
}

// TestGuardWalksDoNotReadOtherCheckouts is the regression test for TRL-59.
//
// Both walks in this file start at the repository root and read whatever they
// find, and this repo's own parallel-work pattern puts full checkouts of OTHER
// branches inside the tree — `git worktree` under .claude/worktrees/. Those
// copies are stale by construction, being other branches, so a walk that reads
// one reports a neighbouring branch's content as this branch's defect.
//
// The fixture is built rather than borrowed, because the real tree cannot
// demonstrate this from anywhere the suite normally runs: CI checks out a fresh
// tree with no worktrees, and an agent working IN a worktree cannot see its
// siblings. Both are green while the main checkout is red — which is exactly how
// 36 findings, every one a .claude/worktrees/… path and not one from the real
// tree, survived CI and three agent sessions in a day.
//
// What the fixture pins, beyond "skip .claude":
//
//   - a worktree's .git is a FILE and a clone's is a DIRECTORY, so the rule is
//     the entry's existence, not its kind;
//   - a checkout can sit anywhere, not only under .claude/, so learning that one
//     string is not the fix;
//   - the root carries a .git entry too, so the rule must exempt it or the first
//     callback skips the repository and every guard passes checking nothing;
//   - the two walks keep DIFFERENT editorial scopes — decisions/ is in
//     marketplaceCommands's and out of docSurfacesIn's — so the tempting repair,
//     giving walker B walker A's whole skip set, is a coverage loss and fails
//     here. That scope is PRESERVED, not newly chosen, and it sits in tension
//     with the reason docSurfaces gives for excluding the same corpus: records
//     are append-only, so a live-command check over them can go permanently red.
//     It cannot today — no governance record matches the command WITH a slug,
//     decisions/0069's `/plugin marketplace add` being a bare mention — and if a
//     marketplace rename ever makes one match, the answer is to exempt it then,
//     deliberately, rather than to narrow the walk now on a guess.
func TestGuardWalksDoNotReadOtherCheckouts(t *testing.T) {
	root := t.TempDir()
	write := func(rel, body string) {
		t.Helper()
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Assembled from fragments so this file carries no literal install command.
	// marketplaceCommands reads .go files under cli/, so a written-out one is a
	// finding against this test — caught by running the guard over a tree that
	// contained this file. Every other mention of the command in this file is
	// already broken the same way, by punctuation rather than by intent
	// (`marketplace add\s+`, "marketplace add %s"), which is why the self-match
	// had never come up. surface_matrix_guard_test.go answers its own version of
	// this by assembling from fragments too, and states the reason to prefer
	// that over an exclusion list: an exclusion list gives the next real drift
	// somewhere to hide.
	const cmd = "/plugin marketplace " + "add "
	const own = cmd + "kodhama/stewards\n"
	const otherBranch = cmd + "kodhama/kodhama\n"

	// The fixture root is a checkout itself, exactly as the repository root is.
	write(".git/HEAD", "ref: refs/heads/main\n")

	// This checkout's own source — the only thing either walk may read.
	write("README.md", own)
	// In marketplaceCommands's scope and outside docSurfacesIn's.
	write("decisions/a-record.md", own)

	// A worktree: .git is a FILE.
	write(".claude/worktrees/agent-x/.git", "gitdir: /elsewhere/.git/worktrees/agent-x\n")
	write(".claude/worktrees/agent-x/README.md", otherBranch)

	// A clone: .git is a DIRECTORY, and it is nowhere near .claude/.
	write("vendor/old-clone/.git/HEAD", "ref: refs/heads/other\n")
	write("vendor/old-clone/README.md", otherBranch)
	write("vendor/old-clone/docs/nested.md", otherBranch)

	// A vendored dependency: not this project's source either.
	write("node_modules/pkg/README.md", otherBranch)

	cmds, err := marketplaceCommands(root)
	if err != nil {
		t.Fatalf("scanning the fixture for install commands: %v", err)
	}
	if got, want := strings.Join(keysOf(cmds), ", "), "README.md, decisions/a-record.md"; got != want {
		t.Errorf("marketplaceCommands read [%s]\nwant                        [%s]\n"+
			"an extra file is another checkout's content, which is a neighbouring branch's business and not this branch's defect (TRL-59); "+
			"a missing one means the walk stopped reading its own tree, and a guard that reads nothing passes silently", got, want)
	}

	surfaces, err := docSurfacesIn(root)
	if err != nil {
		t.Fatalf("walking the fixture for doc surfaces: %v", err)
	}
	var rels []string
	for _, p := range surfaces {
		rel, err := filepath.Rel(root, p)
		if err != nil {
			t.Fatal(err)
		}
		rels = append(rels, filepath.ToSlash(rel))
	}
	sort.Strings(rels)
	if got, want := strings.Join(rels, ", "), "README.md"; got != want {
		t.Errorf("docSurfacesIn read [%s]\nwant               [%s]\nsame rule, same reason (TRL-59)", got, want)
	}
}
