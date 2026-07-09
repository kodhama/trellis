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
// Regenerate on failure:  (from cli/)  go run . setup --dir .. --profile a --mode m1 --target CLAUDE.md --apply
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
			t.Fatalf("repo overlay missing .trellis/%s — run `trellis setup` on the repo: %v", overlayName, err)
		}
		if string(got) != payload[payloadName] {
			t.Errorf(".trellis/%s is stale vs the payload's %s — re-run `trellis setup` on the repo (see this test's doc comment)", overlayName, payloadName)
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
		t.Fatal("repo CLAUDE.md has no trellis:begin/end block — run `trellis setup` on the repo")
	}
	if block := string(c)[i : j+len(trellisEnd)]; block != payload["block-claude.md"] {
		t.Error("repo CLAUDE.md's managed block is stale vs the payload's block-claude.md — re-run `trellis setup` on the repo")
	}
}
