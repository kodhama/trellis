package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRepoOverlayIsCurrent is the self-application sync-guard (decision-0035): the repo's
// committed .trellis/ overlay must equal what `trellis setup` produces from HEAD, so the
// repo's own governance can't drift from the product it ships. If the catalog or the
// render changes, this fails until the overlay is regenerated — drift is impossible, not
// merely visible. The per-build `version` stamp is excluded (it's not behavior).
//
// Regenerate on failure:  (from cli/)  go run . setup --dir .. --profile a --mode m1 --target CLAUDE.md --apply
func TestRepoOverlayIsCurrent(t *testing.T) {
	repoTrellis := filepath.Join("..", ".trellis")
	if _, err := os.Stat(repoTrellis); err != nil {
		t.Skip("no repo .trellis/ overlay yet — nothing to guard")
	}

	tmp := t.TempDir()
	if _, err := applyM1(tmp, planFor("a")); err != nil { // conductor — the repo holds all invariants firmly
		t.Fatal(err)
	}
	for _, f := range []string{"trellis.md", "profile.md", "invariants.md"} {
		want := readFile(t, filepath.Join(tmp, ".trellis", f))
		got, err := os.ReadFile(filepath.Join(repoTrellis, f))
		if err != nil {
			t.Fatalf("repo overlay missing .trellis/%s — run `trellis setup` on the repo: %v", f, err)
		}
		if string(got) != want {
			t.Errorf(".trellis/%s is stale vs the product — re-run `trellis setup` on the repo (see this test's doc comment)", f)
		}
	}
}
