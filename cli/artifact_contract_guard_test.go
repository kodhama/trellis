package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// decision-0028 clause 2: a guard per source/derivative pair. The rubric
// core/rubrics/artifact-contract.md is the artifact contract; the corpus
// conformance check in cli/corpus_conformance_test.go is the code that applies
// it (TRL-88). This guard holds the two together.
//
// Why it exists. Until TRL-88 an agent charter applied the rubric, and TRL-60
// review found charter and rubric disagreeing in eight places. Check 4's
// `<repo>/<id>` form, which decision-0044:161-165 obliged the charter to gain,
// never reached it, and neither did the `@version` pin rule. Applied literally,
// that was 17 false FAILs across 11 files on `main`: a gate rejecting valid
// artifacts, which is the failure that makes a gate stop being believed. Check
// 5's `changes:` rule and the `informed_by` rules were stated in one file only,
// check 6 omitted three types, check 2's scopes went unnamed, and the corpus
// list went stale in both files at once. The charter is retired, and the same
// drift can open between the rubric and the code that replaced it, so the pins
// now run from the rubric to the check.
//
// WHAT THIS GUARD PINS:
//
//   - the corpus: the paths the rubric's **Corpus:** paragraph names equal
//     corpusRoots plus the fixture root the rubric excludes, in both directions;
//   - check 4: each of its five accepted-form rows maps, by its leading text, to
//     one form resolveRef implements, and each implemented form to one row;
//   - check 6: the conformance check consumes exactly the rubric's ten type
//     rows, each parsed to a type with its sections or `exempt`;
//   - the numbered checks: every top-level ordered item in the rubric's check
//     sections, bold or italic, has rows in contractOutcomeTable, there are
//     twelve, and no rows name a check the rubric lacks;
//   - each numbered check's text: its rows carry a digest of the check's
//     normalized rubric text, so an edit anywhere in a check fails here until
//     that check's rows and rule code are re-reviewed (KTD8). The pins this
//     guard carried before TRL-88 compared rows only, and the `changes:` and
//     `informed_by` divergences above were both prose;
//   - rule 10a's confidence tags: confidenceTags equals the values the
//     `confidence` row of schema-typed-artifacts §2 lists, in both directions.
//     Check 10's text does not list them, so its digest cannot catch a change.
//
// Deliberately NOT pinned, so a later reader knows the real bound:
//
//   - what a digest proves: that a check's text is the text its rows were last
//     reviewed against, not that the rows or the rule code are right. Recording
//     a new digest without re-reading the check defeats it, and nothing here can
//     tell;
//   - rubric text outside the numbered checks' windows: the header blockquote
//     beyond the corpus paths, a section's lead-in before its first item (the
//     typed-artifacts section's "Apply only when…"), and the unnumbered sections
//     from `## Honesty clause` on, whose outcome rows are keyed by name;
//   - an ordered item under an H2 that does not begin with `Check`, which is not
//     read as a check;
//   - whether the fixture root is named as excluded: the corpus comparison is a
//     set of paths, and the sentence around them is prose.

// How many rows each enumeration holds. See the count check in contractRows for
// why this is pinned rather than inferred: it is the "compared enough" tripwire
// beside the "compared something" one.
const (
	contractRefForms      = 5  // check 4's accepted depends_on forms
	contractTypeRows      = 10 // check 6's per-type section rules
	contractNumberedCount = 12 // the rubric's numbered checks, retired check 12 included
)

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
// this read check 4's rows under the check-6 label — wrong rows, confident
// message. A duplicated END marker makes the window swallow the rest of the
// document. A missing marker halts too: a rename must fail loudly here rather
// than reduce the read to an empty row set, which would let the guard pass by
// finding nothing.
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
// it, so the read covered a subset and stayed green. Any bullet in the window
// is therefore part of the enumeration. The cost of that choice is that an
// unrelated bulleted list added inside a window changes the row count, and the
// count pin fails loudly, which is the right direction to err.
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
// re-wrapping a row is correctly not a change.
//
// A sub-bullet therefore becomes part of its parent row's text. On a check 6 row
// that makes the row fail to parse; on a check 4 row the leading text still maps
// the row, and the check 4 digest fails instead.
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
			t.Errorf("%s: %q is an ordered-list item. This guard collects the enumeration by bullet, so an ordered row would not be read at all — write it as a bullet", label, line)
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
			// the file no longer would. A row belongs in the list.
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
			t.Errorf("%s: %q is indented deeper than the enumeration (%d spaces vs %d) and follows a blank line, so it belongs to no row and would not be read. Align it with the other rows", label, line, indent, rowIndent)
		default:
			inRow = false
		}
	}
	if len(rows) != want {
		t.Errorf("%s: found %d enumeration rows, expected %d. This count is itself part of the pin: without it a row that this scan cannot see — indented too deep, inside an HTML comment, hidden behind a marker form — silently shrinks what gets read while the rest still matches. If the enumeration genuinely gained or lost a row, update the expected count here in the same change", label, len(rows), want)
	}
	if len(rows) == 0 {
		t.Fatalf("%s: found no enumeration rows in the window — an empty row set would let this guard pass by reading nothing", label)
	}
	for i, r := range rows {
		rows[i] = strings.TrimSpace(contractSpaceRun.ReplaceAllString(r, " "))
	}
	seen := map[string]bool{}
	for _, r := range rows {
		if seen[r] {
			t.Errorf("%s: the row %q appears twice, and a duplicate can stand in for a row that went missing while the count still matches", label, r)
		}
		seen[r] = true
	}
	return rows
}

// contractRefFormRows reads rubric check 4's accepted-form rows. The conformance
// check reads its cross-repo registry from them and this guard maps them to the
// implemented forms, through this one read.
func contractRefFormRows(t *testing.T, rubric string) []string {
	t.Helper()
	return contractRows(t, "rubric check 4", contractWindow(t, "rubric", rubric,
		"4. **`depends_on` resolves.**", "5. **Directional flow"), contractRefForms)
}

// contractSectionRuleRows reads rubric check 6's type rows, for the conformance
// check and this guard alike.
func contractSectionRuleRows(t *testing.T, rubric string) []string {
	t.Helper()
	return contractRows(t, "rubric check 6", contractWindow(t, "rubric", rubric,
		"6. **Required body sections per type**", "7. **Supersede integrity.**"), contractTypeRows)
}

// A corpus path: slash-joined segments, optionally ending in a filename with an
// extension. Deliberately narrow — `[A-Za-z0-9_-]` per segment excludes the dot,
// so a sentence-ending "`profiles/`." yields "profiles/" rather than "profiles/."
// and needs no trimming afterwards.
var contractPathShape = regexp.MustCompile(`(?:[A-Za-z0-9_-]+/)+(?:[A-Za-z0-9_-]+\.[A-Za-z0-9]+)?`)

var contractQuotedToken = regexp.MustCompile("`([^`]+)`")

// contractCorpusPaths returns the distinct paths the rubric's corpus paragraph
// names, and a problem for anything it quotes that looks like a path and cannot
// be read as one.
func contractCorpusPaths(paragraph string) (paths, problems []string) {
	w := strings.NewReplacer("`", "", "*", "").Replace(paragraph)
	seen := map[string]bool{}
	for _, loc := range contractPathShape.FindAllStringIndex(w, -1) {
		m := w[loc[0]:loc[1]]
		// Reject a slash-joined prose token — "read/write" matches the shape as
		// far as "read/", but the character after it continues the word. A real
		// path in this sentence is followed by punctuation or space.
		if loc[1] < len(w) && strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-", rune(w[loc[1]])) {
			continue
		}
		if !seen[m] {
			seen[m] = true
			paths = append(paths, m)
		}
	}
	// A corpus path this shape cannot match would shrink the comparison
	// silently, and there is no natural row count to pin. So anything quoted in
	// the paragraph that looks like a path but did not match is called out
	// rather than dropped: "`core/plans`" without its trailing slash matched
	// only as far as "core/" and vanished.
	for _, tok := range captureAll(contractQuotedToken, paragraph) {
		if strings.Contains(tok, "/") && !seen[tok] && !seen[tok+"/"] {
			problems = append(problems, fmt.Sprintf("corpus: the rubric's corpus paragraph quotes %q, which contains a path separator but is not in the form this guard compares. Write a corpus directory with its trailing slash so it is pinned", tok))
		}
	}
	if len(paths) == 0 {
		problems = append(problems, "corpus: the rubric's corpus paragraph names no path, so there is nothing to hold corpusRoots to")
	}
	return paths, problems
}

// contractRootSpelling writes a corpus root the way the rubric names it:
// repository-relative, a directory with its trailing slash.
func contractRootSpelling(root string) string {
	p := corpusDisplayPath(path.Clean(root))
	if strings.HasSuffix(p, ".md") {
		return p
	}
	return p + "/"
}

// contractCorpusProblems compares the paths the rubric's corpus paragraph names
// with the roots the conformance check reads plus the fixture root the rubric
// excludes, in both directions.
func contractCorpusProblems(paragraph string, roots []string, fixtureRoot string) []string {
	named, problems := contractCorpusPaths(paragraph)
	read := map[string]bool{}
	var readOrder []string
	for _, r := range append(append([]string{}, roots...), fixtureRoot) {
		if p := contractRootSpelling(r); !read[p] {
			read[p] = true
			readOrder = append(readOrder, p)
		}
	}
	inRubric := map[string]bool{}
	for _, p := range named {
		inRubric[p] = true
		if !read[p] {
			problems = append(problems, fmt.Sprintf("corpus: the rubric's corpus paragraph names %q, which is neither a root in corpusRoots nor the fixture root the rubric excludes, so the contract claims a path the check does not read", p))
		}
	}
	for _, p := range readOrder {
		if !inRubric[p] {
			problems = append(problems, fmt.Sprintf("corpus: the conformance check reads %q (corpusRoots, or the fixture root under knownBadRoot), which the rubric's corpus paragraph does not name, so the check gates a path the contract does not claim", p))
		}
	}
	return problems
}

// contractRefFormLead maps one of rubric check 4's accepted-form rows to the
// form resolveRef returns for a reference of that form.
type contractRefFormLead struct {
	form string // an accepts rule id of contractOutcomeTable's check 4
	lead string // the row's leading text, whitespace collapsed, markdown as written
}

// contractRefFormLeads is the row-to-form map. A lead runs through the phrase
// that defines the form, so a row whose meaning moves stops matching; wording
// after the lead is pinned by the check 4 digest instead.
var contractRefFormLeads = []contractRefFormLead{
	{"4a", "an existing artifact `id` in this corpus"},
	{"4b", "a declared external-ref form — `brief-§…`"},
	{"4c", "a qualified `<repo>/<id>` cross-repo reference whose `<repo>` is a member of the recognized registry"},
	{"4d", "a **retired id** in the invariant-set's Identifiers registry"},
	{"4e", "a **retired artifact id** in `decision-0079`'s retired-artifacts registry"},
}

// contractRefPinForm is check 4's `@version` pin handling: an accepts row with
// no rubric row of its own. resolveRef strips a pin before it tries any form,
// and the rubric states the pin in check 4's prose, which the check 4 digest
// pins.
const contractRefPinForm = "4f"

// contractRefFormProblems maps each check 4 row to the one implemented form
// whose lead it starts with, and reports a row that maps to no form or to
// several, a form with no row or with two, and a lead naming a form the outcome
// table does not mark accepts.
func contractRefFormProblems(rows []string, table []contractCheckOutcomes, leads []contractRefFormLead) []string {
	var problems []string
	accepts := map[string]bool{}
	var forms []string
	for _, c := range table {
		if c.check != 4 {
			continue
		}
		for _, r := range c.rules {
			if r.outcome == outcomeAccepts {
				accepts[r.rule] = true
				if r.rule != contractRefPinForm {
					forms = append(forms, r.rule)
				}
			}
		}
	}
	if !accepts[contractRefPinForm] {
		problems = append(problems, fmt.Sprintf("check 4: %s, the `@version` pin handling this guard exempts from the row mapping, is not an accepts row of contractOutcomeTable's check 4. Remove the exemption, or restore the row", contractRefPinForm))
	}
	for _, l := range leads {
		if !accepts[l.form] || l.form == contractRefPinForm {
			problems = append(problems, fmt.Sprintf("check 4: contractRefFormLeads maps the lead %q to %s, which is not an accepted form of contractOutcomeTable's check 4", l.lead, l.form))
		}
	}
	rowOf := map[string]string{}
	for _, row := range rows {
		var matched []string
		for _, l := range leads {
			if strings.HasPrefix(row, l.lead) {
				matched = append(matched, l.form)
			}
		}
		switch {
		case len(matched) == 0:
			problems = append(problems, fmt.Sprintf("check 4: the rubric row %q starts with the leading text of no implemented accepted form. If the form's meaning changed, change resolveRef and its outcome row to match; if only the wording did, update contractRefFormLeads", row))
		case len(matched) > 1:
			problems = append(problems, fmt.Sprintf("check 4: the rubric row %q starts with the leading text of %d forms (%s), so it maps to none. Make each lead in contractRefFormLeads specific to one row", row, len(matched), strings.Join(matched, ", ")))
		default:
			if prev, dup := rowOf[matched[0]]; dup {
				problems = append(problems, fmt.Sprintf("check 4: form %s is mapped by two rubric rows, %q and %q, so one of them states a form the check does not implement", matched[0], prev, row))
				continue
			}
			rowOf[matched[0]] = row
		}
	}
	for _, f := range forms {
		if _, ok := rowOf[f]; !ok {
			problems = append(problems, fmt.Sprintf("check 4: form %s is implemented and accepted by the conformance check, and no rubric row maps to it, so the check accepts references the contract does not state", f))
		}
	}
	return problems
}

// contractNumberedCheck is one top-level ordered item in the rubric's check
// sections.
type contractNumberedCheck struct {
	number int
	line   int    // 1-based
	text   string // from the item marker to the next top-level item or H2, whitespace collapsed
}

// contractNumberedChecks returns the top-level ordered items under every H2
// whose text begins with `Check`, in file order.
//
// A top-level item is an ordered-list marker indented less than the content of
// the item before it, which is how CommonMark tells a sibling from a nested
// item; fenced code is skipped when looking for markers and headings. An item
// this scan reads as nested stays inside its parent's text, so the parent's
// digest still changes.
func contractNumberedChecks(rubric string) []contractNumberedCheck {
	var checks []contractNumberedCheck
	var body []string
	inChecks, fence, contentIndent := false, "", -1
	closeItem := func() {
		if body != nil {
			checks[len(checks)-1].text = strings.TrimSpace(contractSpaceRun.ReplaceAllString(strings.Join(body, "\n"), " "))
		}
		body, contentIndent = nil, -1
	}
	for i, line := range strings.Split(strings.ReplaceAll(rubric, "\r\n", "\n"), "\n") {
		// Fenced code is never read as a marker or a heading, and stays in the
		// open item's text.
		var fenced bool
		if fence, fenced = nextFence(fence, line); !fenced {
			if m := h2Shape.FindStringSubmatch(line); m != nil {
				closeItem()
				inChecks = contractCheckSection.MatchString(h2Text(m[1]))
				continue
			} else if m := contractCheckMarker.FindStringSubmatch(line); inChecks && m != nil && (contentIndent < 0 || len(m[1]) < contentIndent) {
				closeItem()
				n, _ := strconv.Atoi(m[2])
				checks = append(checks, contractNumberedCheck{number: n, line: i + 1})
				gap := len(m[3])
				if gap == 0 || gap > 4 {
					gap = 1
				}
				body, contentIndent = []string{}, len(m[1])+len(m[2])+1+gap
			}
		}
		if body != nil {
			body = append(body, line)
		}
	}
	closeItem()
	return checks
}

// contractCheckSection is the H2 text of a section holding numbered checks:
// `Checks`, `Checks — the two typed artifacts (…)`, `Check — version cross-check
// (retired)`.
var contractCheckSection = regexp.MustCompile(`^Checks?(?:\s|$)`)

// contractCheckMarker is an ordered-list marker, bold or italic or plain text
// after it: the retired check 12 is italic.
var contractCheckMarker = regexp.MustCompile(`^( {0,3})(\d{1,9})[.)]([ \t]+|$)`)

// contractCheckDigest is the digest a check's outcome rows carry: the first 16
// hex digits of the SHA-256 of its normalized text.
func contractCheckDigest(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])[:16]
}

// contractOutcomeCoverageProblems compares the rubric's numbered checks with
// contractOutcomeTable: the count, a number used twice on either side, a check
// with no rows, rows for a check the rubric lacks, and a check whose text no
// longer matches the digest its rows carry.
func contractOutcomeCoverageProblems(checks []contractNumberedCheck, table []contractCheckOutcomes, want int) []string {
	var problems []string
	if len(checks) != want {
		problems = append(problems, fmt.Sprintf("numbered checks: found %d top-level numbered items in the rubric's check sections, expected %d. The count is part of the pin: a check this scan cannot see would otherwise leave the rest matching. If the rubric genuinely gained or lost a check, record its outcome rows and update contractNumberedCount in the same change", len(checks), want))
	}
	rubric := map[int]contractNumberedCheck{}
	var unique []contractNumberedCheck
	for _, c := range checks {
		if prev, dup := rubric[c.number]; dup {
			problems = append(problems, fmt.Sprintf("numbered checks: rubric check %d is numbered twice, on lines %d and %d, so outcome rows cannot say which one they account for", c.number, prev.line, c.line))
			continue
		}
		rubric[c.number] = c
		unique = append(unique, c)
	}
	rowed := map[int]bool{}
	for _, row := range table {
		if rowed[row.check] {
			problems = append(problems, fmt.Sprintf("numbered checks: contractOutcomeTable carries check %d twice; keep one group of rows per check", row.check))
			continue
		}
		rowed[row.check] = true
		c, ok := rubric[row.check]
		if !ok {
			problems = append(problems, fmt.Sprintf("numbered checks: contractOutcomeTable carries rows for check %d, which the rubric's check sections do not number, so those rows account for a check that does not exist", row.check))
			continue
		}
		if got := contractCheckDigest(c.text); got != row.digest {
			problems = append(problems, fmt.Sprintf("numbered checks: rubric check %d (line %d) is not the text its outcome rows were reviewed against: contractOutcomeTable records digest %q, and the rubric text now digests to %q. Re-review check %d's outcome rows and rule code against the new text, then record the new digest, in the same change (KTD8)", c.number, c.line, row.digest, got, c.number))
		}
	}
	for _, c := range unique {
		if !rowed[c.number] {
			problems = append(problems, fmt.Sprintf("numbered checks: rubric check %d (line %d) has no rows in contractOutcomeTable, so no outcome is recorded for it (R5). Add a row, with an outcome and a reason, for each of its parts", c.number, c.line))
		}
	}
	return problems
}

// contractGuardReport fails the test once per problem.
func contractGuardReport(t *testing.T, problems []string) {
	t.Helper()
	for _, p := range problems {
		t.Error(p)
	}
}

// TestArtifactContractCorpusMatchesCorpusRoots pins the corpus: the rubric's
// **Corpus:** paragraph names exactly the roots the conformance check reads,
// plus the fixture root it excludes. TRL-54 found this list stale only because a
// human went looking.
//
// Paths are matched BY SHAPE, after stripping markdown emphasis, rather than by
// looking for backticks. Backtick extraction was a silent-pass hole: the rubric
// mixes `core/schemas/` with **`core/schemas/`**, so a path written
// bold-but-unquoted is a live authoring slip a backtick scan could not see.
//
// The window runs to the END of the corpus paragraph, not to the word "Exclude",
// because a path named in that tail is still a path the rubric names. Paths are
// deduplicated, because the paragraph names `core/schemas/` twice, once in the
// list and once in the prose defending it.
func TestArtifactContractCorpusMatchesCorpusRoots(t *testing.T) {
	paragraph := contractWindow(t, "rubric corpus paragraph", readFileT(t, artifactContractPath),
		"**Corpus:**", "**Derived resource")
	contractGuardReport(t, contractCorpusProblems(paragraph, corpusRoots, path.Dir(knownBadRoot)))
}

// TestArtifactContractRefFormRowsMapToImplementedForms pins check 4, the
// enumeration with the worst blast radius: a missing accepted form is a false
// FAIL rather than a false pass.
func TestArtifactContractRefFormRowsMapToImplementedForms(t *testing.T) {
	rows := contractRefFormRows(t, readFileT(t, artifactContractPath))
	contractGuardReport(t, contractRefFormProblems(rows, contractOutcomeTable, contractRefFormLeads))
}

// TestArtifactContractSectionRulesAreTheRowsTheCheckReads pins check 6, the
// enumeration TRL-60 was filed about: the rubric states that its list "names
// every type the corpus holds", so a check reading fewer rows silently exempts
// a type the contract gates.
//
// It pins that the rubric holds contractTypeRows (ten) check 6 rows, and that
// each parses, through parseSectionRule, into a type with its sections or
// `exempt`, with no type named twice. contractSectionRuleRows pins the count,
// both in the direct call here and inside loadArtifactContract.
// loadArtifactContract parses the rows and halts on a row that does not parse
// or a repeated type. The check's section rules are built from that one read
// and parse, so no second parse exists that could diverge from it, and this
// test compares nothing against a second copy.
func TestArtifactContractSectionRulesAreTheRowsTheCheckReads(t *testing.T) {
	contractSectionRuleRows(t, readFileT(t, artifactContractPath))
	loadArtifactContract(t)
}

// TestArtifactContractNumberedChecksHaveReviewedOutcomes pins R5: every numbered
// rubric check has outcome rows, and its text is the text those rows were last
// reviewed against.
func TestArtifactContractNumberedChecksHaveReviewedOutcomes(t *testing.T) {
	checks := contractNumberedChecks(readFileT(t, artifactContractPath))
	contractGuardReport(t, contractOutcomeCoverageProblems(checks, contractOutcomeTable, contractNumberedCount))
}

// confidenceTagCell is a backticked value in the schema's `confidence` row.
var confidenceTagCell = regexp.MustCompile("`([^`]+)`")

// schemaConfidenceTags reads the values the `confidence` row of the schema's
// Field/Rule tables lists in its Rule cell, up to the dash that opens the gloss.
// A schema with no such row, two of them, or a row naming no value halts the
// test.
func schemaConfidenceTags(t *testing.T, schema *corpusArtifact) []string {
	t.Helper()
	var cells []string
	for _, tb := range schema.tables(0, len(schema.lines)) {
		field, rule := slices.Index(tb.header, "Field"), slices.Index(tb.header, "Rule")
		if field < 0 || rule < 0 {
			continue
		}
		for _, r := range tb.rows {
			if len(r.cells) == len(tb.header) && cellValue(r.cells[field]) == "confidence" {
				cells = append(cells, r.cells[rule])
			}
		}
	}
	if len(cells) != 1 {
		t.Fatalf("%s: found %d `confidence` rows in its Field/Rule tables, and this guard needs exactly one to read the tags rule 10a accepts", schema.path, len(cells))
	}
	values, _, _ := strings.Cut(cells[0], " — ")
	var tags []string
	for _, m := range confidenceTagCell.FindAllStringSubmatch(values, -1) {
		tags = append(tags, m[1])
	}
	if len(tags) == 0 {
		t.Fatalf("%s: the `confidence` row's Rule cell names no backticked value: %q", schema.path, cells[0])
	}
	return tags
}

// confidenceTagProblems compares the tags rule 10a accepts with the tags the
// schema lists, in both directions.
func confidenceTagProblems(accepted map[string]bool, listed []string) []string {
	var problems []string
	inSchema := map[string]bool{}
	for _, tag := range listed {
		inSchema[tag] = true
		if !accepted[tag] {
			problems = append(problems, fmt.Sprintf("schema-typed-artifacts lists confidence tag %q, and confidenceTags does not accept it", tag))
		}
	}
	var extra []string
	for tag := range accepted {
		if !inSchema[tag] {
			extra = append(extra, tag)
		}
	}
	slices.Sort(extra)
	for _, tag := range extra {
		problems = append(problems, fmt.Sprintf("confidenceTags accepts %q, and schema-typed-artifacts does not list it", tag))
	}
	return problems
}

// TestArtifactContractConfidenceTagsMatchTheSchema pins rule 10a's accepted
// confidence tags to the schema table they are copied from (decision-0028).
func TestArtifactContractConfidenceTagsMatchTheSchema(t *testing.T) {
	schema := liveArtifact(t, liveCorpusCheck(t), "schema-typed-artifacts")
	contractGuardReport(t, confidenceTagProblems(confidenceTags, schemaConfidenceTags(t, schema)))
}

// TestArtifactContractGuardComparisonsFail seeds each failure path of the
// comparisons above, so a comparison that stopped detecting its drift turns
// this red instead of leaving the live tests green by finding nothing. Every
// case runs on strings and slices; none edits the rubric.
func TestArtifactContractGuardComparisonsFail(t *testing.T) {
	rubric := readFileT(t, artifactContractPath)
	paragraph := contractWindow(t, "rubric corpus paragraph", rubric, "**Corpus:**", "**Derived resource")
	fixtureRoot := path.Dir(knownBadRoot)
	refRows := contractRefFormRows(t, rubric)
	checks := contractNumberedChecks(rubric)

	withRow := func(i int, row string) []string {
		rows := append([]string{}, refRows...)
		rows[i] = row
		return rows
	}
	withLeadForm := func(i int, form string) []contractRefFormLead {
		leads := append([]contractRefFormLead{}, contractRefFormLeads...)
		leads[i].form = form
		return leads
	}
	withoutRule := func(rule string) []contractCheckOutcomes {
		var table []contractCheckOutcomes
		for _, c := range contractOutcomeTable {
			kept := c
			kept.rules = nil
			for _, r := range c.rules {
				if r.rule != rule {
					kept.rules = append(kept.rules, r)
				}
			}
			table = append(table, kept)
		}
		return table
	}
	withoutCheck := func(n int) []contractCheckOutcomes {
		var table []contractCheckOutcomes
		for _, c := range contractOutcomeTable {
			if c.check != n {
				table = append(table, c)
			}
		}
		return table
	}
	withChecks := func(extra ...contractCheckOutcomes) []contractCheckOutcomes {
		return append(append([]contractCheckOutcomes{}, contractOutcomeTable...), extra...)
	}

	// One changed word inside check 4's prose, in a copy of the rubric text.
	const word, changed = "swallowed silently", "swallowed quietly"
	if n := strings.Count(rubric, word); n != 1 {
		t.Fatalf("the rubric carries %q %d times, and the check 4 digest case needs it exactly once", word, n)
	}
	mutated := strings.Replace(rubric, word, changed, 1)
	newDigest := "(check 4 not found)"
	for _, c := range contractNumberedChecks(mutated) {
		if c.number == 4 {
			if !strings.Contains(c.text, changed) {
				t.Errorf("check 4's text window does not hold %q, so the digest case changes a word outside check 4", changed)
			}
			newDigest = contractCheckDigest(c.text)
		}
	}

	// A synthetic rubric: a bold check, an italic retired check under its own
	// `Check` heading, and a numbered item under a heading that is not a check.
	const synthetic = "# Rubric\n\n## Checks\n\n1. **Bold.** A rule.\n\n## Check — retired\n\n2. *(Retired.)*\n\n## Honesty clause\n\n3. Not a check.\n"
	syntheticTable := []contractCheckOutcomes{{1, contractCheckDigest("1. **Bold.** A rule."), []contractRuleOutcome{{"1", outcomeImplemented, "seeded"}}}}

	cases := []struct {
		name     string
		problems []string
		n        int      // exactly this many problems
		want     []string // each appears in some problem
	}{
		{"corpus: a path the rubric names and corpusRoots does not read",
			contractCorpusProblems(paragraph+" `templates/`", corpusRoots, fixtureRoot), 1, []string{`"templates/"`}},
		// The seeded root carries no `../` prefix: cli/ci_paths_guard_test.go reads
		// a `../` literal as a repository read, and contractRootSpelling spells
		// both forms alike.
		{"corpus: a root corpusRoots reads and the rubric does not name",
			contractCorpusProblems(paragraph, append(append([]string{}, corpusRoots...), "templates"), fixtureRoot), 1, []string{`"templates/"`}},
		{"corpus: a quoted path the extraction cannot read",
			contractCorpusProblems(paragraph+" `core/plans`", corpusRoots, fixtureRoot), 1, []string{`"core/plans"`}},
		{"corpus: a paragraph that names no path",
			contractCorpusProblems("no paths here", corpusRoots, fixtureRoot), 2 + len(corpusRoots), []string{"names no path", `"docs/decisions/"`, `"core/fixtures/"`}},

		{"check 4: a row whose leading text changed",
			contractRefFormProblems(withRow(0, "an existing artifact `id` in any corpus"), contractOutcomeTable, contractRefFormLeads),
			2, []string{strconv.Quote("an existing artifact `id` in any corpus"), "form 4a "}},
		{"check 4: an implemented form with no row",
			contractRefFormProblems(refRows[1:], contractOutcomeTable, contractRefFormLeads), 1, []string{"form 4a "}},
		{"check 4: a row matching two forms' leading text",
			contractRefFormProblems(refRows, contractOutcomeTable, append(append([]contractRefFormLead{}, contractRefFormLeads...), contractRefFormLead{"4b", "an existing"})),
			2, []string{strconv.Quote(refRows[0]), "form 4a "}},
		{"check 4: two rows mapping to one form",
			contractRefFormProblems(withRow(1, refRows[0]+" again"), contractOutcomeTable, contractRefFormLeads), 2, []string{"form 4a ", "form 4b "}},
		{"check 4: a lead naming a form the outcome table does not mark accepts",
			contractRefFormProblems(refRows, contractOutcomeTable, withLeadForm(0, "4i")), 2, []string{"4i", "form 4a "}},
		{"check 4: the pin handling is not an accepts row",
			contractRefFormProblems(refRows, withoutRule(contractRefPinForm), contractRefFormLeads), 1, []string{contractRefPinForm}},

		{"numbered checks: a check with no outcome rows",
			contractOutcomeCoverageProblems(checks, withoutCheck(9), contractNumberedCount), 1, []string{"check 9 "}},
		{"numbered checks: outcome rows naming a check the rubric lacks",
			contractOutcomeCoverageProblems(checks, withChecks(contractCheckOutcomes{13, "", []contractRuleOutcome{{"13", outcomeRetired, "seeded"}}}), contractNumberedCount),
			1, []string{"check 13,"}},
		{"numbered checks: a check numbered twice in the rubric",
			contractOutcomeCoverageProblems(append(append([]contractNumberedCheck{}, checks...), contractNumberedCheck{3, 999, "3. again"}), contractOutcomeTable, contractNumberedCount),
			2, []string{"found 13", "check 3 is numbered twice"}},
		{"numbered checks: one check carrying two groups of rows",
			contractOutcomeCoverageProblems(checks, withChecks(contractOutcomeTable[1]), contractNumberedCount), 1, []string{"check 2 twice"}},
		{"numbered checks: the rubric's count moved",
			contractOutcomeCoverageProblems(checks, contractOutcomeTable, contractNumberedCount+1), 1, []string{"found 12", "expected 13"}},
		{"digest: one changed word inside check 4's prose",
			contractOutcomeCoverageProblems(contractNumberedChecks(mutated), contractOutcomeTable, contractNumberedCount), 1, []string{"check 4 ", strconv.Quote(newDigest)}},
		{"numbered checks: an italic item is a check, and an item outside the check sections is not",
			contractOutcomeCoverageProblems(contractNumberedChecks(synthetic), syntheticTable, 2), 1, []string{"check 2 "}},

		{"confidence tags: a tag the schema lists and confidenceTags does not accept",
			confidenceTagProblems(confidenceTags, []string{"verified", "inferred", "speculated", "likely"}), 1, []string{`"likely"`, "does not accept"}},
		{"confidence tags: a tag confidenceTags accepts and the schema does not list",
			confidenceTagProblems(confidenceTags, []string{"verified", "inferred"}), 1, []string{`"speculated"`, "does not list"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.problems) != tc.n {
				t.Errorf("want %d problems, got %d:\n  %s", tc.n, len(tc.problems), strings.Join(tc.problems, "\n  "))
			}
			for _, w := range tc.want {
				if !slices.ContainsFunc(tc.problems, func(p string) bool { return strings.Contains(p, w) }) {
					t.Errorf("no problem names %s:\n  %s", w, strings.Join(tc.problems, "\n  "))
				}
			}
		})
	}
}
