package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestRepoOverlayIsCurrent is the self-application sync-guard (decision-0035),
// reworked by #117 (kodhama-0007 slice 1) and again by decision-0051 (the authority
// split): the repo's committed .trellis/ overlay is diffed against the payload file
// set — a fresh in-process render (payloadFiles()), the same render
// TestVendoredPayloadIsCurrent pins the vendored plugins/trellis/reference/ copy to.
// The two guards together give generator == vendored payload == repo overlay: drift
// stays impossible, and the repo's own overlay is exactly a mechanical copy of the
// shipped artifact (kodhama-0007 rule 2). The repo's posture is a/conductor (all
// rows active, so internal/rules.md is the shipped all-active assembly); the
// per-install `version` stamp stays excluded (gitignored — it's not behavior,
// decision-0035).
//
// Regenerate on failure — the manual copy path, applied to ourselves (from the repo
// root, after `go run . payload --out ../plugins/trellis/reference` in cli/):
//
//	mkdir -p .trellis/internal
//	cp plugins/trellis/reference/trellis-a.md  .trellis/internal/trellis.md
//	cp plugins/trellis/reference/rules.md      .trellis/internal/rules.md
//	cp plugins/trellis/reference/invariants.md .trellis/internal/invariants.md
func TestRepoOverlayIsCurrent(t *testing.T) {
	repoTrellis := filepath.Join("..", ".trellis")
	if _, err := os.Stat(repoTrellis); err != nil {
		t.Skip("no repo .trellis/ overlay yet — nothing to guard")
	}

	payload := payloadFiles()
	for overlayName, payloadName := range map[string]string{
		"internal/trellis.md":    "trellis-a.md", // conductor — the repo holds all invariants firmly
		"internal/rules.md":      "rules.md",     // all rows active — the shipped all-active assembly
		"internal/invariants.md": "invariants.md",
	} {
		got, err := os.ReadFile(filepath.Join(repoTrellis, overlayName))
		if err != nil {
			t.Fatalf("repo overlay missing .trellis/%s — copy it from the payload (see this test's doc comment): %v", overlayName, err)
		}
		if string(got) != payload[payloadName] {
			t.Errorf(".trellis/%s is stale vs the payload's %s — re-copy it from the payload (see this test's doc comment)", overlayName, payloadName)
		}
	}

	// The flat pre-decision-0051 layout is migrated, not left beside the new one:
	// the old generated paths must be gone (their content lives under internal/).
	for _, legacy := range []string{"trellis.md", "profile.md", "invariants.md", "version"} {
		if _, err := os.Stat(filepath.Join(repoTrellis, legacy)); !os.IsNotExist(err) {
			t.Errorf(".trellis/%s still exists — the flat layout retired with decision-0051; delete the old-path copy", legacy)
		}
	}

	// The managed block in the repo's CLAUDE.md is the payload's block-claude.md,
	// verbatim — the copier contract, applied to ourselves.
	c, err := os.ReadFile(filepath.Join("..", "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	i := strings.Index(string(c), trellisBegin)
	j := strings.Index(string(c), trellisEnd)
	if i < 0 || j <= i {
		t.Fatal("repo CLAUDE.md has no trellis:begin/end block — paste the payload's block-claude.md")
	}
	if block := string(c)[i : j+len(trellisEnd)]; block != payload["block-claude.md"] {
		t.Error("repo CLAUDE.md's managed block is stale vs the payload's block-claude.md — re-paste it between the markers")
	}
}

// TestRepoDeclaresRulesConfig: decision-0051 rules 1+2 — the repo's own overlay
// carries the consumer-authoritative config file, .trellis/rules.toml, seeded from
// the conductor posture (the posture TestRepoOverlayIsCurrent's trellis-a.md mapping
// pins) with every assessable row active (which is what pins internal/rules.md to
// the all-active assembly). Rows are the machine-read truth, so the machine-read
// keys are asserted; nothing else in .trellis/ root is pinned (it is
// consumer-authoritative).
func TestRepoDeclaresRulesConfig(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", ".trellis", "rules.toml"))
	if err != nil {
		t.Fatalf("repo overlay has no .trellis/rules.toml — the consumer-authoritative config (decision-0051 rule 1): %v", err)
	}
	content := string(b)
	if !strings.Contains(content, `strictness  = "firm"`) {
		t.Errorf(".trellis/rules.toml must declare strictness \"firm\" (the a/conductor posture the repo overlay is pinned to), got: %q", content)
	}
	for _, slug := range assessableSlugs {
		rowRe := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(slug) + `\s+= \{ active = true \}`)
		if !rowRe.MatchString(content) {
			t.Errorf(".trellis/rules.toml must carry an active row for %s — the repo holds every invariant firmly (its internal/rules.md is pinned to the all-active assembly)", slug)
		}
	}
}

// TestRepoExpressionIsPureProse: decision-0051 rule 5 — the profile: frontmatter key
// retired with the machine-read config's move to rules.toml, so the repo's own
// hand-owned .trellis/expression.md must carry no YAML frontmatter at all. Only that
// structural fact is asserted: the body is hand-owned and no test may pin it (the
// ownership rule — 100% generated or 100% hand-owned, never mixed).
func TestRepoExpressionIsPureProse(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", ".trellis", "expression.md"))
	if err != nil {
		t.Fatalf("repo overlay has no .trellis/expression.md — the hand-owned declaration file (kodhama-0007 rule 4): %v", err)
	}
	content := string(b)
	if strings.HasPrefix(content, "---") {
		t.Errorf(".trellis/expression.md must not open with YAML frontmatter — the machine-read profile: key retired (decision-0051 rule 5); strip the frontmatter, leave the body. Got: %q", content[:min(len(content), 60)])
	}
}
