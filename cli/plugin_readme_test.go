package main

// The shipped plugin README carries Trellis's host-support claim, under a test
// (decision-0066 §2, AC4/AC5).
//
// Before this file, nothing in the suite read plugins/trellis/README.md at all:
// docs_consistency_test.go's docSurfaces list covers ../README.md, ../docs/*.html
// and ../install.sh, and its only assertions are that docs name no nonexistent
// `trellis <cmd>` or `/trellis:<skill>`. Deleting the entire host-support section
// left `go test ./...` green, and so did shipping a live relative link into a
// file this change deletes — both are simulated failures decision-0066 records,
// not hypotheticals. These two tests are the deliverable that closes that gap.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const pluginReadmePath = "../plugins/trellis/README.md"

// hostSupportMarker opens the single paragraph decision-0066 §2 requires. The
// guard is paragraph-scoped deliberately: "Claude Code" also appears in the
// README's own title and "marketplace" in its Install section, so a whole-file
// assertion would stay green after the claim itself had been deleted. Scoping to
// one block also makes AC4's "one paragraph" mechanically true rather than a
// matter of layout.
const hostSupportMarker = "**Where Trellis is known to work, and what that does not mean.**"

// hostSupportParagraph returns the one blank-line-delimited block of
// plugins/trellis/README.md that opens with hostSupportMarker, with internal
// whitespace runs collapsed to single spaces: the README hard-wraps its prose, so
// a required phrase routinely straddles a line break. Line wrapping is not
// semantic, and a guard that broke on rewrapping would train people to weaken it.
func hostSupportParagraph(t *testing.T, readme string) string {
	t.Helper()
	for _, block := range strings.Split(readme, "\n\n") {
		if strings.HasPrefix(strings.TrimSpace(block), hostSupportMarker) {
			return strings.Join(strings.Fields(block), " ")
		}
	}
	t.Fatalf("%s must carry one paragraph opening with %q — the host-support claim decision-0066 §2 moves here from the retired surfaces.json",
		pluginReadmePath, hostSupportMarker)
	return ""
}

// TestPluginReadmeStatesHostSupportClaim is decision-0066 AC4/AC5: one paragraph,
// naming the hosts Trellis is known to work on, which check establishes that,
// that support is not claimed, and the marketplace hedge the retired rows
// carried. Each element is asserted separately so that deleting any one of the
// four turns this test red — an assertion scoped to a single sentence would leave
// the other three deletable with the suite green.
func TestPluginReadmeStatesHostSupportClaim(t *testing.T) {
	paragraph := hostSupportParagraph(t, readFileT(t, pluginReadmePath))

	for _, required := range []struct{ element, needle string }{
		{"the hosts it is known to work on — Claude", "Claude Code"},
		{"the hosts it is known to work on — Codex", "Codex CLI"},
		{"which check establishes them", "measured against a real session with file tools disabled"},
		{"that support is not claimed", "support is not claimed"},
		{"the marketplace hedge the retired surface rows carried", "not that the listed install path"},
	} {
		if !strings.Contains(paragraph, required.needle) {
			t.Errorf("%s: the host-support paragraph must state %s, and does not — no %q in it. "+
				"decision-0066 AC4 requires all of these in this one paragraph.",
				pluginReadmePath, required.element, required.needle)
		}
	}
}

var markdownLinkRe = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)

// TestPluginReadmeLinksResolve is the guard for the failure decision-0066's
// "risk this decision accepts" section reproduces by simulation: the coordinated
// deletion passed `go vet`, `go test ./...` and all eight vendor integration
// tests while `[surfaces.json](surfaces.json)` was still live in the shipped
// README. Every repository-relative link target must exist on disk, because this
// file ships to every consumer's disk.
func TestPluginReadmeLinksResolve(t *testing.T) {
	readme := readFileT(t, pluginReadmePath)
	dir := filepath.Dir(pluginReadmePath)

	for _, match := range markdownLinkRe.FindAllStringSubmatch(readme, -1) {
		target := match[1]
		if strings.Contains(target, "://") || strings.HasPrefix(target, "#") || strings.HasPrefix(target, "mailto:") {
			continue
		}
		if cut := strings.IndexAny(target, "#?"); cut >= 0 {
			target = target[:cut]
		}
		if target == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(target))); err != nil {
			t.Errorf("%s links to %q, which does not exist in the shipped bundle — a dead relative link ships to every consumer's disk: %v",
				pluginReadmePath, match[1], err)
		}
	}
}

// decision-0068 D11. The install path renders `.claude/rules/trellis.md`, a
// Trellis-authored instructions file the remove skill did not know about.
// Unchanged, `/trellis:remove` would delete `.trellis/` — the rows — while
// leaving the governing file behind with a dangling import, still loaded into
// every session, with nothing left to activate any rule.
//
// spec-0005 §1 names the standard itself: "a vendored install with no
// /trellis:remove is a governance tool with no clean exit, which spec-0004
// already treats as a trust defect."
func TestRemoveSkillEnumeratesTheRenderedRulesFile(t *testing.T) {
	body := readFileT(t, "../plugins/trellis/skills/remove/SKILL.md")

	if !strings.Contains(body, ".claude/rules/trellis.md") {
		t.Fatalf("/trellis:remove does not enumerate .claude/rules/trellis.md — it would delete the rows and leave the governing file behind")
	}

	// Ordering is the substantive half. An interrupted removal must never leave
	// the governing file present with its rows already gone: rules that cannot
	// be activated, in always-loaded context, is a worse state than either end.
	iRendered := strings.Index(body, ".claude/rules/trellis.md")
	// The needle must survive the prose's own emphasis markers. An earlier
	// version searched for "then removes `.trellis/`" and matched ZERO times,
	// because SKILL.md writes "and **then** removes" — so this assertion never
	// ran. Anchor on the unadorned phrase instead, and fail loudly if even that
	// stops matching rather than silently skipping the check.
	iOverlay := strings.LastIndex(body, "removes `.trellis/`")
	if iOverlay < 0 {
		t.Fatalf("cannot locate the .trellis/ removal step — the ordering assertion would silently not run")
	}
	if iOverlay >= 0 && iRendered > iOverlay {
		t.Errorf("the rendered rules file must be removed BEFORE .trellis/, so an interrupted removal never strands a governing file without its rows")
	}
}

// decision-0068 makes the README's install-path sentence false, and its own
// host-support test would keep passing while it shipped to every consumer's
// disk. The claim is load-bearing: it is the paragraph a reader consults to
// decide whether the curl path governs anything.
func TestPluginReadmeInstallPathClaimIsCurrent(t *testing.T) {
	body := readFileT(t, "../plugins/trellis/README.md")

	if strings.Contains(body, "install.sh` registers no hook at all, so a vendored") ||
		strings.Contains(body, "install delivers no rules") {
		t.Fatalf("the README still says a vendored install delivers no rules; decision-0068 D1 renders .claude/rules/trellis.md, so this ships false to every consumer")
	}
	if !strings.Contains(body, ".claude/rules/trellis.md") {
		t.Errorf("the README should name the delivery mechanism the install path now uses")
	}
	// The honesty of the surrounding paragraph must survive the correction: the
	// change fixes project scope only, and personal scope still delivers nothing.
	if !strings.Contains(body, "personal") {
		t.Errorf("the corrected claim must still name the scope that delivers no rules, or it over-claims")
	}
}
