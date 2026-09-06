package main

// The reintroduction guard for decision-0066: nothing brings the surface matrix
// back.
//
// Two parts, because a single repo-wide content grep is not implementable. The
// matrix's names legitimately survive in the append-only archive — decisions 0059,
// 0061, 0063 and 0064 all carry them, and must keep them (specs/0005 and
// specs/0008 did too, until decision-0079 deleted specs/).
// A whole-repo content guard would be red on its first run against this
// repository's own history and would degenerate into an exclusion list
// (decision-0066 AC6).

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const repoRoot = ".."

// TestNoSurfaceMatrixFile is AC6 part 1: a repository walk, excluding .git, that
// fails on any file whose basename is surfaces.json, case-insensitively. It
// cannot self-match — this guard's own basename is not surfaces.json — so no
// carve-out is needed and none exists.
//
// AC6's "excluding .git" is read as naming the STRUCTURAL exclusion — what is
// not this checkout's content — rather than as an exhaustive list of one name,
// so this walk asks structuralSkip rather than testing that name (TRL-63). That
// serves the criterion's INTENT; where it reaches past AC6's enumeration it says
// so, rather than claiming the letter already covered it:
//
//   - AC6 states part 1's verification as "Verified to return empty against a
//     simulated post-deletion tree." Read narrowly that is past-tense narration
//     of the author's own feasibility check, and the tree they simulated held no
//     sibling worktrees — so this is not a letter argument and is not offered as
//     one. Read as the property the criterion was after — on a tree where the
//     matrix is deleted, this guard returns empty — the one-name version stops
//     delivering it as soon as a sibling holds a stale surfaces.json, which the
//     main checkout's 17 worktrees make live rather than theoretical.
//   - AC6's objection to exclusions is that they give real drift somewhere to
//     hide — a whole-repo content grep "degenerates into an exclusion list".
//     (Its sharper line, "A second exclusion entry means the guard is wrong", is
//     part 2's, about the self-reference carve-out, and does not bind part 1.)
//     That objection lands on one residual hole, named here rather than waved
//     off: an UNTRACKED .git entry inside a tracked directory hides that
//     directory, while its siblings stay perfectly committable. `mkdir core/.git`
//     alongside a core/surfaces.json is invisible to this walk while `git status`
//     offers the file for commit — verified, not reasoned. It takes a stray
//     `git init` or a copied repo to arise, CI's fresh checkout cannot have one,
//     and the two walks in docs_consistency_test.go have carried the same hole
//     since #287. What is NOT hidden is the reintroduction this guard exists
//     for: plugins/trellis/surfaces.json is still reported, pinned by
//     TestSurfaceMatrixGuardDoesNotReadOtherCheckouts. Eleven vendor and
//     bundle-manifest tests also fail on such a file, but only while install.sh's
//     manifest still disagrees with the tree — restore the manifest line too and
//     every one of them goes green again, leaving this guard the only thing that
//     reports it. The backstop covers a half-reintroduction, not the whole one.
//   - A nested checkout is another repository's content, marked by the .git
//     entry it carries. Not "under .git": a worktree's .git is a FILE and the
//     tree is its sibling, which is why isSeparateCheckout Lstats dir/.git
//     instead of matching the name.
//
// node_modules comes along with structuralSkip and is the one part not forced by
// the argument above, so it is stated rather than smuggled: no node_modules
// exists in this tree and .gitignore does not mention one, and a dependency's
// file named surfaces.json would not be Trellis carrying surface rows. The
// alternative — a bespoke structural rule here — is the shape TRL-59 was: two
// guard walks disagreeing about what counts as this checkout.
//
// No marking on decision-0066 is owed, and the ground is NOT that its letter
// stretches to cover this. It is that nothing in 0066's substance moves: §1's
// retirement stands, part 2 stands, and part 1 still fails on any surfaces.json
// in this repository's content, outside the one residual hole named above — a
// hole the one-name version did not have, and the honest price of the fix. What
// changed beyond that is one clause of one AC's implementation, whose cost of
// reversal is a one-line edit. decision-0081 argues supersession authority
// should scale with exactly that cost, on a test of "consequential AND
// irreversible" — cited here as reasoning, not as authority, because 0081 is a
// proposal on record rather than an applied rule ("Proposed — not taken", and
// its catalog edit still owed). The ground above does not rest on it.
func TestNoSurfaceMatrixFile(t *testing.T) {
	scan, err := scanForSurfaceMatrixFiles(repoRoot)
	if err != nil {
		t.Fatalf("walking the repository: %v", err)
	}
	if scan.examined < surfaceScanFloor {
		t.Fatalf("the walk examined %d files, under the floor of %d — this repository tracked 438 files when "+
			"that floor was set, so a broken walk is a likelier explanation than a clean tree. This guard's "+
			"pass condition is ZERO hits, which is also what a walk that reads nothing returns; without this "+
			"floor the two are indistinguishable and the guard passes silently forever.", scan.examined, surfaceScanFloor)
	}
	for _, path := range scan.hits {
		t.Errorf("%s reintroduces the retired surface matrix — decision-0066 §1: Trellis carries no exact surface rows, "+
			"no behavior state, and no marketplace-observation record. Where Trellis is known to work is stated in "+
			"plugins/trellis/README.md, under TestPluginReadmeStatesHostSupportClaim.", path)
	}
}

// surfaceScanFloor is the smallest number of files the AC6 part-1 walk may
// examine against the real repository before the walk itself is the likelier
// explanation than a clean tree. 438 tracked files when this was written, so it
// is a broken-walk detector with four-fold headroom, not a census. It exists
// because TRL-63 gave this walk a skip rule, and a guard whose pass condition is
// an empty result cannot otherwise tell "nothing to find" from "read nothing" —
// the same reason docSurfaces states a floor in docs_consistency_test.go.
//
// The mutation it uniquely catches is a NARROWED ROOT: scanning "." instead of
// repoRoot reaches only the files under cli/ — a couple of dozen, far under the
// floor — reports no hits, and leaves the fixture below green, because that
// fixture supplies its own root. A skip rule gone too wide
// is a different matter — losing isSeparateCheckout's root exemption zeroes this
// walk, but the fixture fails on it too, so the floor is not the only thing
// standing between that mutation and a silent pass. Both verified.
const surfaceScanFloor = 100

// surfaceMatrixScan is one walk's result: the files whose basename is
// surfaces.json, and how many files the walk actually looked at. The second
// number exists because the first is expected to be empty — see the floor above.
type surfaceMatrixScan struct {
	hits     []string
	examined int
}

// scanForSurfaceMatrixFiles is TestNoSurfaceMatrixFile's walk with its root
// passed in rather than fixed at the repository, so
// TestSurfaceMatrixGuardDoesNotReadOtherCheckouts can aim it at a fixture. The
// defect that test pins cannot be reproduced against the real tree from a
// worktree, which is where an agent stands (TRL-59, TRL-63).
func scanForSurfaceMatrixFiles(root string) (surfaceMatrixScan, error) {
	var scan surfaceMatrixScan
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if structuralSkip(root, path) {
				return fs.SkipDir
			}
			return nil
		}
		scan.examined++
		if strings.EqualFold(entry.Name(), "surfaces.json") {
			scan.hits = append(scan.hits, path)
		}
		return nil
	})
	return scan, err
}

// TestSurfaceMatrixGuardDoesNotReadOtherCheckouts is the regression test for
// TRL-63.
//
// This repository's parallel-work pattern puts full checkouts of OTHER branches
// inside the tree — `git worktree` under .claude/worktrees/, which .gitignore
// does not cover — so a walk that starts at the repository root and reads
// whatever it finds reads a neighbouring branch's content as this branch's. That
// is TRL-59's defect, fixed in #287 for the two walks in docs_consistency_test.go
// and left standing here, where it was latent rather than firing: this guard hits
// only on a file basenamed surfaces.json, and no sibling holds one today. It
// would fire on the first sibling branch that reintroduced the matrix — which is
// precisely the reintroduction this guard exists to catch — reporting a true
// positive against the WRONG branch, on a branch that is clean.
//
// The fixture is built rather than borrowed, because neither place this suite
// runs can demonstrate it: CI checks out a fresh tree with no worktrees, and an
// agent working IN a worktree cannot see its siblings. Both would stay green;
// only the main checkout could go red — and today not even that, the defect
// being latent, which is why the fixture supplies the surfaces.json no sibling
// currently holds.
//
// What it pins, beyond "skip .claude":
//
//   - a worktree's .git is a FILE and a clone's is a DIRECTORY, so the rule is
//     the entry's existence, not its kind;
//   - a checkout can sit anywhere, including under a directory this repo really
//     has, so a rule that matches DIRECTORY NAMES is not the fix — see the
//     placement note at the clone below, which is what makes this one bite;
//   - the root carries a .git entry too, so the rule must exempt it — otherwise
//     the first callback skips the whole repository and this guard, whose pass
//     condition is an empty result, passes forever while reading nothing;
//   - the walk still reads its OWN tree, case-insensitively, which is the half
//     of AC6 part 1 that a too-eager skip rule would quietly delete.
func TestSurfaceMatrixGuardDoesNotReadOtherCheckouts(t *testing.T) {
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

	// The fixture root is a checkout itself, exactly as the repository root is.
	write(".git/HEAD", "ref: refs/heads/main\n")
	// Inside .git, where this walk has always refused to look.
	write(".git/modules/x/surfaces.json", "{}\n")

	// This checkout's own content — the only thing the walk may report.
	write("plugins/trellis/README.md", "hosts Trellis is known to work on\n")
	write("plugins/trellis/surfaces.json", "{}\n")
	// Case-insensitivity is AC6 part 1's letter, so it is pinned here too.
	write("core/SURFACES.JSON", "{}\n")

	// A worktree: .git is a FILE.
	write(".claude/worktrees/agent-x/.git", "gitdir: /elsewhere/.git/worktrees/agent-x\n")
	write(".claude/worktrees/agent-x/plugins/trellis/surfaces.json", "{}\n")

	// A clone: .git is a DIRECTORY, and it sits under core/ — a directory this
	// repository really has and really scans. That placement is what makes the
	// bullet above true rather than merely asserted: parked under a name of its
	// own (vendor/, say), this clone is skipped just as well by a rule that
	// matches DIRECTORY NAMES and never looks for a .git entry at all, and both
	// this test and TestGuardWalksDoNotReadOtherCheckouts pass a structuralSkip
	// mutated to `name == ".git" || "node_modules" || ".claude" || "vendor"`
	// — verified. Under core/ the same mutation goes red, because it matches no
	// name here and the clone's own surfaces.json leak into the report; and
	// widening it to match `core` does not rescue it, because that loses
	// core/SURFACES.JSON from the want set above. Both halves verified.
	write("core/old-clone/.git/HEAD", "ref: refs/heads/other\n")
	write("core/old-clone/surfaces.json", "{}\n")
	write("core/old-clone/docs/nested/surfaces.json", "{}\n")

	// A vendored dependency: not this project's content either.
	write("node_modules/pkg/surfaces.json", "{}\n")

	scan, err := scanForSurfaceMatrixFiles(root)
	if err != nil {
		t.Fatalf("walking the fixture: %v", err)
	}

	var rels []string
	for _, p := range scan.hits {
		rel, err := filepath.Rel(root, p)
		if err != nil {
			t.Fatal(err)
		}
		rels = append(rels, filepath.ToSlash(rel))
	}
	sort.Strings(rels)
	if got, want := strings.Join(rels, ", "), "core/SURFACES.JSON, plugins/trellis/surfaces.json"; got != want {
		t.Errorf("the guard reported [%s]\nwant                 [%s]\n"+
			"an extra path is another checkout's content — a neighbouring branch's business, not this branch's "+
			"defect (TRL-63); a missing one means the walk stopped reading its own tree, and a guard whose pass "+
			"condition is an empty result passes silently when it reads nothing", got, want)
	}
	// plugins/trellis/README.md, plugins/trellis/surfaces.json and
	// core/SURFACES.JSON. Not a second spelling of the check above: that one
	// pins a NON-EMPTY list, so it already fails a walk that entered nothing.
	// What this adds is the count of files REACHED — a walk narrowed to look at
	// fewer files while still finding both hits satisfies the list and fails
	// here.
	if scan.examined != 3 {
		t.Errorf("the walk examined %d files, want 3 — this checkout's own content is three files, and every "+
			"other file in the fixture belongs to .git, to another checkout, or to a dependency", scan.examined)
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
// install.sh only. decisions/ is deliberately out of scope; the
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
