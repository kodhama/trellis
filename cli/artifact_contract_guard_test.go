package main

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// decision-0028 clause 2: a guard per source/derivative pair. The rubric
// `core/rubrics/artifact-contract.md` is the artifact contract; the charter
// `.claude/agents/corpus-reviewer.md` is the agent that applies it, and the
// rubric names it as its derived resource. Nothing held them together, and by
// TRL-60 review had found them disagreeing in eight places — five where the
// charter lagged the rubric, two where the charter enforced what no artifact
// stated, and one where both went stale together:
//
//   - check 4's accepted `depends_on` forms: the charter never gained the
//     `<repo>/<id>` cross-repo branch that decision-0044:161-165 obliged it to
//     gain ("corpus-reviewer's check 4, concretely: gains one new branch").
//     Applied literally that is 17 false FAILs across 11 files on `main` — a
//     conformance gate rejecting valid artifacts, which is the failure mode
//     that makes a gate stop being believed.
//   - check 4's `@version` pin rule, stated only in the rubric, while three live
//     `depends_on` entries carry a pin (`spec-0007@v1` and two others). Same
//     false-FAIL shape, same check.
//   - check 5's `changes:` rule: stated in the rubric, absent from the charter,
//     while nine decisions carry a `changes:` relation. Same false-FAIL shape as
//     the above, in the check next door.
//   - check 6's per-type sections: the charter omitted `signature-catalog`,
//     `expression-profile` and `lexicon`, and the rubric's `research-note` row
//     carried a `(+ sources)` parenthetical the charter never had (divergent at
//     birth in 4d64f23, never once present under `.claude/agents/`).
//   - check 2's `scope` enumeration: named in the rubric, unnamed in the charter.
//   - the `informed_by` resolution rules, and separately the `informed_by`
//     honesty rule (a coupling relabeled as provenance): both stated only in the
//     CHARTER — an agent applying rules its contract never stated.
//   - the corpus list: went stale in both files together in c75eace.
//
// WHAT THIS GUARD PINS, precisely: the three enumerations below, as rows. It
// does NOT pin the prose around them, and prose in these files can carry a
// normative rule — the `changes:` and `informed_by` divergences above were both
// prose. So this is a guard against the drift that has actually happened, not a
// proof that the two files agree. The reconciliation of the prose was done by
// hand in the same change, and the next prose divergence will need the same.
// Byte-identity is not the goal and would not survive a week: the rubric states
// a contract, the charter instructs an agent applying it, and they are
// deliberately different voices.
//
// Deliberately NOT pinned, so a later reader knows the real bound rather than
// trusting this further than it goes:
//
//   - every check's surrounding explanation, per the paragraph above;
//   - check 8's ten-field list for a `signature-catalog` entry, verified
//     identical across the pair when this guard was written; and the "Recognized
//     typed artifacts" list, where the two name the same three types but only the
//     rubric annotates each with its `scope`. Follow-up work, not a claim that
//     they cannot drift;
//   - anything about a RUNNING session. This pins file to file. An agent whose
//     charter was resolved before an edit still applies the old rules, and that
//     is invisible from here — it happened during this change's own review, and
//     the reviewer reported it would have produced false FAILs on nine files.

// How many rows each pinned enumeration holds. See the count check in
// contractRows for why this is pinned rather than inferred: it is the
// "compared enough" tripwire beside the "compared something" one.
const (
	contractRefForms = 5  // check 4's accepted depends_on forms
	contractTypeRows = 10 // check 6's per-type section rules
)

func contractRubricPath(t *testing.T) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join("..", "core", "rubrics", "artifact-contract.md"))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func contractCharterPath(t *testing.T) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join("..", ".claude", "agents", "corpus-reviewer.md"))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// U+00A0 is collapsed along with ordinary whitespace. Go's `\s` does not cover
// it while strings.TrimSpace does, and that inconsistency turned a pasted
// non-breaking space into a failure between two rows that render identically.
var contractSpaceRun = regexp.MustCompile(`[\s\x{00A0}]+`)

// An ordered-list marker inside a window: the one CommonMark list form these
// enumerations do not use. Rows are collected by bullet, so a row written as
// "1. `type` → …" would be silently invisible; this makes it loud instead.
var contractOrderedItem = regexp.MustCompile(`^\d+[.)]\s`)

// A blockquote marker, stripped BEFORE the indent is measured. Measuring first
// and stripping second was a silent-pass hole of its own: "   > - row" scored an
// indent of 5 against a 3-space enumeration, so it read as a nested sub-bullet
// and vanished into the row above — or, after a blank line, was dropped entirely.
var contractQuotePrefix = regexp.MustCompile(`^[ \t]*>[ \t]?`)

// A closed HTML comment, removed from a line before the line is classified.
// Skipping the whole line instead let "<!-- --> - `row`" hide a bullet.
var contractClosedComment = regexp.MustCompile(`<!--.*?-->`)

// contractWindow returns the slice of doc between start and end. Each marker
// must appear EXACTLY ONCE. Requiring uniqueness rather than taking the first
// match is deliberate: a marker that reappears elsewhere in the file silently
// moves the window, and the two ways that goes wrong are both bad. A duplicated
// START marker (a cross-reference to "check 6" written above check 1, say) makes
// this compare check 4's rows under the check-6 label — wrong rows, confident
// message. A duplicated END marker makes the window swallow the rest of the
// document. A missing marker halts too: a rename must fail loudly here rather
// than reduce the comparison to two empty row sets, which would let the guard
// pass by finding nothing.
func contractWindow(t *testing.T, label, doc, start, end string) string {
	t.Helper()
	if n := strings.Count(doc, start); n != 1 {
		t.Fatalf("%s: the marker %q occurs %d times, and this guard needs exactly one to locate the enumeration unambiguously — reword the duplicate, or update this test to a marker that is unique", label, start, n)
	}
	rest := doc[strings.Index(doc, start)+len(start):]
	if n := strings.Count(rest, end); n != 1 {
		t.Fatalf("%s: the closing marker %q occurs %d times after %q, and this guard needs exactly one to know where the enumeration ends", label, end, n, start)
	}
	return rest[:strings.Index(rest, end)]
}

// contractRows collects EVERY top-level bullet in a window, not just the first
// contiguous block. Collecting only the first block was a silent-pass hole: a
// row appended after the block's trailing paragraph — the most natural place for
// an editor to add one — was invisible here while a human reading the check saw
// it, so the guard compared a subset and stayed green. Any bullet in the window
// is therefore part of the pinned enumeration. The cost of that choice is that
// an unrelated bulleted list added inside a window must appear in both files;
// the guard fails loudly if it does not, which is the right direction to err.
//
// All three CommonMark bullet markers are accepted, after a tab or a space, and
// through a blockquote prefix. Every one of those was a way to write a row the
// guard could not see. Fenced blocks are skipped so an example's bullets are not
// mistaken for rules, and an ordered-list item — the one list form left — fails
// loudly rather than being dropped.
//
// A row's continuation is the non-blank line following it with no blank line
// between — a wrapped line, or a nested sub-bullet, both of which belong to the
// row they hang from rather than being rows of their own. A blank line ends the
// current row without ending collection, so prose paragraphs between blocks are
// skipped instead of being glued onto the row above. Whitespace is collapsed, so
// re-wrapping one side is correctly not a divergence.
//
// A sub-bullet therefore becomes part of its parent row's text, which means
// annotating a row in ONE file is a divergence and fails. That is intended: the
// two enumerations are pinned to be identical, so a clarification worth making
// is worth making on both sides. The failure names the row and both readings.
func contractRows(t *testing.T, label, window string, want int) []string {
	t.Helper()
	var rows []string
	rowIndent, inRow, fenced, commented := -1, false, false, false
	for _, quoted := range strings.Split(window, "\n") {
		raw := contractClosedComment.ReplaceAllString(quoted, "")
		quotedRow := raw != contractQuotePrefix.ReplaceAllString(raw, "")
		raw = contractQuotePrefix.ReplaceAllString(raw, "")
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~") {
			fenced = !fenced
			inRow = false
			continue
		}
		if fenced {
			continue
		}
		// An HTML comment renders as nothing, so a row inside one is in force for
		// no reader while still looking like a rule to a naive line scan.
		if strings.HasPrefix(line, "<!--") {
			commented = !strings.Contains(line, "-->")
			inRow = false
			continue
		}
		if commented {
			commented = !strings.Contains(line, "-->")
			continue
		}
		if line == "" {
			inRow = false
			continue
		}
		if contractOrderedItem.MatchString(line) {
			t.Errorf("%s: %q is an ordered-list item. This guard collects the enumeration by bullet, so an ordered row would not be compared at all — write it as a bullet", label, line)
			continue
		}
		indent := len(raw) - len(strings.TrimLeft(raw, " \t"))
		isBullet := len(line) > 2 &&
			strings.ContainsRune("-*+", rune(line[0])) &&
			(line[1] == ' ' || line[1] == '\t')
		switch {
		case isBullet && quotedRow:
			// Neither collected nor ignored. Ignoring it let a row hide inside a
			// blockquote; collecting it let a row be DEMOTED to a "historical
			// note" aside that still counted, so the scan saw a rule a reader of
			// the file no longer would. A pinned row belongs in the list.
			t.Errorf("%s: %q is a bullet inside a blockquote. An enumeration row must be in the list itself — a quoted aside reads as commentary to a human and as a rule to this scan", label, line)
		case isBullet && (rowIndent < 0 || indent <= rowIndent):
			if rowIndent < 0 {
				rowIndent = indent
			}
			rows = append(rows, strings.TrimSpace(line[2:]))
			inRow = true
		case inRow:
			rows[len(rows)-1] += " " + line
		case isBullet:
			// A bullet indented deeper than the enumeration with no row above it
			// to hang from. It is not a continuation and it is not a row, so it
			// would simply vanish — and a row one space too deep is an ordinary
			// authoring slip, not an adversarial construction.
			t.Errorf("%s: %q is indented deeper than the enumeration (%d spaces vs %d) and follows a blank line, so it belongs to no row and would not be compared. Align it with the other rows", label, line, indent, rowIndent)
		default:
			inRow = false
		}
	}
	if len(rows) != want {
		t.Errorf("%s: found %d enumeration rows, expected %d. This count is itself part of the pin: without it a row that this scan cannot see — indented too deep, inside an HTML comment, hidden behind a marker form — silently shrinks what gets compared while the rest still matches. If the enumeration genuinely gained or lost a row, update the expected count here in the same change", label, len(rows), want)
	}
	if len(rows) == 0 {
		t.Fatalf("%s: found no enumeration rows in the window — an empty row set on either side would let this guard pass by comparing nothing", label)
	}
	for i, r := range rows {
		rows[i] = strings.TrimSpace(contractSpaceRun.ReplaceAllString(r, " "))
	}
	seen := map[string]bool{}
	for _, r := range rows {
		if seen[r] {
			t.Errorf("%s: the row %q appears twice. This guard compares row sets, so a duplicate on one side can mask a missing row on the other", label, r)
		}
		seen[r] = true
	}
	return rows
}

// contractCompare fails naming every row that one side carries and the other
// does not, in both directions. Both directions matter: the charter lagging the
// rubric is the observed failure, but a charter that grows a rule the contract
// does not state is an agent enforcing something no artifact authorizes — which
// this pair had too, in its `informed_by` prose.
//
// Rows are printed with %q, not %s. A difference can be one invisible character,
// and two rows that render identically in a plain-text failure message send the
// reader hunting for a difference they cannot see.
func contractCompare(t *testing.T, what string, rubric, charter []string) {
	t.Helper()
	in := func(rows []string, row string) bool {
		for _, r := range rows {
			if r == row {
				return true
			}
		}
		return false
	}
	differ := false
	for _, r := range rubric {
		if !in(charter, r) {
			differ = true
			t.Errorf("%s: core/rubrics/artifact-contract.md states a rule that .claude/agents/corpus-reviewer.md does not, so the agent applies a contract narrower than the one on record:\n  rubric row:  %q", what, r)
		}
	}
	for _, c := range charter {
		if !in(rubric, c) {
			differ = true
			t.Errorf("%s: .claude/agents/corpus-reviewer.md states a rule that core/rubrics/artifact-contract.md does not, so the agent enforces something no artifact authorizes:\n  charter row: %q", what, c)
		}
	}
	// Only worth reporting an ordering problem once the two sides carry the same
	// rows. Reporting it alongside a content difference would print "both files
	// state the same rules but in a different order" when they plainly do not.
	// A duplicate row is reported, not fatal, so both sides can be named in one
	// run. That leaves the two slices possibly unequal in length here even when
	// every row of each appears in the other, and indexing the shorter one below
	// would panic and abort the whole package's tests.
	if differ || len(rubric) != len(charter) {
		return
	}
	// Order is pinned too. Two reasons: prose in either file may refer to a row
	// by position — the rubric's check 4 did exactly that ("That last clause…")
	// before this change named its referent outright — and an enumeration a human
	// is expected to read against its twin is easier to diff by eye when the two
	// agree line for line.
	for i := range rubric {
		if rubric[i] != charter[i] {
			t.Errorf("%s: both files state the same rules but in a different order, at position %d. Keep the two enumerations line-for-line identical — prose on either side may refer to a row by its position:\n  rubric row:  %q\n  charter row: %q", what, i+1, rubric[i], charter[i])
			break
		}
	}
}

// TestCorpusReviewerCharterMatchesRubricSectionRules pins check 6. This is the
// enumeration TRL-60 was filed about and the one #289 closed: the rubric states
// that its list "names every type the corpus holds", so a charter carrying
// fewer rows is an agent silently exempting a type the contract gates.
func TestCorpusReviewerCharterMatchesRubricSectionRules(t *testing.T) {
	rubric := readFileT(t, contractRubricPath(t))
	charter := readFileT(t, contractCharterPath(t))
	contractCompare(t, "check 6 (required body sections per type)",
		contractRows(t, "rubric check 6", contractWindow(t, "rubric", rubric,
			"6. **Required body sections per type**", "7. **Supersede integrity.**"), contractTypeRows),
		contractRows(t, "charter check 6", contractWindow(t, "charter", charter,
			"6. Required body sections per type", "7. Supersede integrity:"), contractTypeRows))
}

// TestCorpusReviewerCharterMatchesRubricDependsOnForms pins check 4 — the
// divergence with the worst blast radius, because a missing accepted form is a
// false FAIL rather than a false pass. decision-0044:161-165 named this check
// and obliged the branch; it landed in the rubric and never in the charter.
func TestCorpusReviewerCharterMatchesRubricDependsOnForms(t *testing.T) {
	rubric := readFileT(t, contractRubricPath(t))
	charter := readFileT(t, contractCharterPath(t))
	contractCompare(t, "check 4 (accepted depends_on forms)",
		contractRows(t, "rubric check 4", contractWindow(t, "rubric", rubric,
			"4. **`depends_on` resolves.**", "5. **Directional flow"), contractRefForms),
		contractRows(t, "charter check 4", contractWindow(t, "charter", charter,
			"4. Every `depends_on` resolves", "5. **Directional flow"), contractRefForms))
}

// A corpus path: slash-joined segments, optionally ending in a filename with an
// extension. Deliberately narrow — `[A-Za-z0-9_-]` per segment excludes the dot,
// so a sentence-ending "`profiles/`." yields "profiles/" rather than "profiles/."
// and needs no trimming afterwards.
var contractPathShape = regexp.MustCompile(`(?:[A-Za-z0-9_-]+/)+(?:[A-Za-z0-9_-]+\.[A-Za-z0-9]+)?`)

// TestCorpusReviewerCharterMatchesRubricCorpus pins the third enumeration: which
// directories the gate reads. c75eace moved this line in both files at once, so
// it has never actually diverged — but it is the pair's third enumerated list
// and the cheapest of the three to hold, and TRL-54 only found it stale because
// a human went looking.
//
// Unlike the other two this compares paths in place rather than bulleted rows:
// neither side writes the corpus as a list, and rewriting two sentences to make
// them machine-comparable buys nothing the extraction does not already give.
//
// Paths are matched BY SHAPE, after stripping markdown emphasis, rather than by
// looking for backticks. Backtick extraction was a silent-pass hole: the rubric
// already mixes `core/schemas/` with **`core/schemas/`**, so a path written
// bold-but-unquoted is a live authoring slip the guard could not see.
//
// The window runs to the END of each file's corpus paragraph, not to the word
// "Exclude", because a path named in that tail is still a path this gate reads.
// Paths are deduplicated before comparison because both files legitimately name
// `core/schemas/` twice in that region, once in the list and once in the prose
// defending it.
//
// What this pins is the SET OF PATHS NAMED, and nothing else about the
// paragraph. The surrounding claims are prose and are not compared: flipping
// "checked, not merely consulted" to its opposite on one side leaves this green.
// That is the same disclosed bound as the other two tests, stated here because
// the corpus paragraph is unusually prose-heavy for an enumeration.
func TestCorpusReviewerCharterMatchesRubricCorpus(t *testing.T) {
	paths := func(label, doc, start, end string) []string {
		t.Helper()
		orig := contractWindow(t, label, doc, start, end)
		w := strings.NewReplacer("`", "", "*", "").Replace(orig)
		seen := map[string]bool{}
		var got []string
		for _, loc := range contractPathShape.FindAllStringIndex(w, -1) {
			m := w[loc[0]:loc[1]]
			// Reject a slash-joined prose token — "read/write" matches the shape
			// as far as "read/", but the character after it continues the word.
			// A real path in this sentence is followed by punctuation or space.
			if loc[1] < len(w) && strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-", rune(w[loc[1]])) {
				continue
			}
			if !seen[m] {
				seen[m] = true
				got = append(got, m)
			}
		}
		// A corpus path this shape cannot match would shrink the comparison
		// silently, and unlike checks 4 and 6 there is no natural row count to
		// pin. So anything quoted in this paragraph that looks like a path but
		// did not match is called out rather than dropped: "`core/plans`" without
		// its trailing slash matched only as far as "core/" and vanished.
		for _, m := range regexp.MustCompile("`([^`]+)`").FindAllStringSubmatch(orig, -1) {
			tok := m[1]
			if !strings.Contains(tok, "/") {
				continue
			}
			if !seen[tok] && !seen[tok+"/"] {
				t.Errorf("%s: %q is quoted in the corpus paragraph and contains a path separator, but is not in the form this guard compares. Write a corpus directory with its trailing slash so it is pinned", label, tok)
			}
		}
		if len(got) == 0 {
			t.Fatalf("%s: no corpus paths found between %q and %q", label, start, end)
		}
		return got
	}
	rubric := paths("rubric corpus paragraph", readFileT(t, contractRubricPath(t)),
		"**Corpus:**", "**Derived resource")
	charter := paths("charter corpus paragraph", readFileT(t, contractCharterPath(t)),
		"**Default corpus:**", "Recognized typed artifacts")
	contractCompare(t, "the corpus this gate reads", rubric, charter)
}
