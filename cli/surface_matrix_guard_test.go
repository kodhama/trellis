package main

// The reintroduction guard for decision-0066: nothing brings the surface matrix
// back.
//
// Two parts, because a single repo-wide content grep is not implementable. The
// matrix's names legitimately survive in the append-only archive — decisions 0059,
// 0061, 0063, 0064, specs/0005 and specs/0008 all carry them, and must keep them.
// A whole-repo content guard would be red on its first run against this
// repository's own history and would degenerate into an exclusion list
// (decision-0066 AC6).

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const repoRoot = ".."

// TestNoSurfaceMatrixFile is AC6 part 1: a repository walk, excluding .git, that
// fails on any file whose basename is surfaces.json, case-insensitively. It
// cannot self-match — this guard's own basename is not surfaces.json — so no
// carve-out is needed and none exists.
func TestNoSurfaceMatrixFile(t *testing.T) {
	err := filepath.WalkDir(repoRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return fs.SkipDir
			}
			return nil
		}
		if strings.EqualFold(entry.Name(), "surfaces.json") {
			t.Errorf("%s reintroduces the retired surface matrix — decision-0066 §1: Trellis carries no exact surface rows, "+
				"no behavior state, and no marketplace-observation record. Where Trellis is known to work is stated in "+
				"plugins/trellis/README.md, under TestPluginReadmeStatesHostSupportClaim.", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking the repository: %v", err)
	}
}

// retiredFieldNames are the three field names the matrix carried, assembled from
// fragments so that this file contains no literal occurrence of any of them.
//
// decision-0066 AC6 specifies this rather than leaving it to the implementer,
// because an earlier draft's guard got it wrong: cli/ is the repository's only Go
// package, so the guard necessarily lives inside a directory it scans, and
// written with the literals inline it fails on itself. The criterion allows one
// alternative — excluding exactly this file's own path — which is not taken here,
// because then there is an exclusion list, and the next person to hit a match has
// somewhere to put it.
func retiredFieldNames() []string {
	return []string{
		"surface" + "_id",
		"behavior" + "_state",
		"marketplace_test" + "_observations",
	}
}

// TestNoMatrixFieldsInCode is AC6 part 2: code-scoped — cli/, plugins/ and
// install.sh only. decisions/ and specs/ are deliberately out of scope; the
// archive keeps its words.
func TestNoMatrixFieldsInCode(t *testing.T) {
	needles := retiredFieldNames()

	for _, scope := range []string{".", filepath.Join(repoRoot, "plugins"), filepath.Join(repoRoot, "install.sh")} {
		err := filepath.WalkDir(scope, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			for _, needle := range needles {
				if strings.Contains(string(content), needle) {
					t.Errorf("%s: retired surface-matrix field %q survives in code (decision-0066 §1, AC6)", path, needle)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walking %s: %v", scope, err)
		}
	}
}
