package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRepoOverlayIsCurrent is the self-application sync-guard (decision-0035),
// reworked by #117 (kodhama-0007 slice 1): the repo's committed .trellis/ overlay is
// diffed against the payload file set — a fresh in-process render (payloadFiles()),
// the same render TestVendoredPayloadIsCurrent pins the vendored
// plugins/trellis/reference/ copy to. The two guards together give
// generator == vendored payload == repo overlay: drift stays impossible, and the
// repo's own overlay is exactly a mechanical copy of the shipped artifact
// (kodhama-0007 rule 2). The repo's posture is a/conductor; the per-install
// `version` stamp stays excluded (gitignored — it's not behavior, decision-0035).
//
// Regenerate on failure — the manual copy path, applied to ourselves (#120: the
// CLI's setup command retired; from the repo root, after `go run . payload --out
// ../plugins/trellis/reference` in cli/):
//
//	cp plugins/trellis/reference/trellis-a.md    .trellis/trellis.md
//	cp plugins/trellis/reference/profile-a.md    .trellis/profile.md
//	cp plugins/trellis/reference/invariants.md   .trellis/invariants.md
func TestRepoOverlayIsCurrent(t *testing.T) {
	repoTrellis := filepath.Join("..", ".trellis")
	if _, err := os.Stat(repoTrellis); err != nil {
		t.Skip("no repo .trellis/ overlay yet — nothing to guard")
	}

	payload := payloadFiles()
	for overlayName, payloadName := range map[string]string{
		"trellis.md":    "trellis-a.md", // conductor — the repo holds all invariants firmly
		"profile.md":    "profile-a.md",
		"invariants.md": "invariants.md",
	} {
		got, err := os.ReadFile(filepath.Join(repoTrellis, overlayName))
		if err != nil {
			t.Fatalf("repo overlay missing .trellis/%s — copy it from the payload (see this test's doc comment): %v", overlayName, err)
		}
		if string(got) != payload[payloadName] {
			t.Errorf(".trellis/%s is stale vs the payload's %s — re-copy it from the payload (see this test's doc comment)", overlayName, payloadName)
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

// TestRepoDeclaresExpression: kodhama/trellis#119 (kodhama-0007 rule 4) — the header
// this repo's own overlay imports now carries `@expression.md`, so self-hosting parity
// (decision-0035) requires the repo to hold its own hand-owned declaration file, with
// frontmatter that machine-reads to the same posture TestRepoOverlayIsCurrent pins
// (a/conductor). Only the frontmatter is asserted: the body is hand-owned and no test
// may pin it (the ownership rule — 100% generated or 100% hand-owned, never mixed).
func TestRepoDeclaresExpression(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", ".trellis", "expression.md"))
	if err != nil {
		t.Fatalf("repo overlay has no .trellis/expression.md — the header imports it (kodhama-0007 rule 4, #119): %v", err)
	}
	content := string(b)
	if !strings.HasPrefix(content, "---\n") {
		t.Fatalf(".trellis/expression.md must open with YAML frontmatter, got: %q", content[:min(len(content), 40)])
	}
	rest := content[len("---\n"):]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		t.Fatal(".trellis/expression.md frontmatter is unterminated")
	}
	if frontmatter := rest[:end]; !strings.Contains(frontmatter, "profile: a") {
		t.Errorf(".trellis/expression.md frontmatter must declare `profile: a` (the posture the repo overlay is pinned to), got: %q", frontmatter)
	}
}
