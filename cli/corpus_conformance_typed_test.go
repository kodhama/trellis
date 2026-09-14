package main

// Checks 8–11 of the artifact contract: the two typed artifacts, a
// `signature-catalog` and an `expression-profile` (schema-typed-artifacts).
// The entry point, the finding format, the halt policy and the outcome table
// are in corpus_conformance_test.go. Rules 8a and 8d are not here:
// TestRowSetDerivativesFollowThePin already fails when the catalog's entries
// differ from the pinned slug set, dials and collapsed slugs included.

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// catalogFields are the ten fields rule 8b requires of every entry.
var catalogFields = []string{"what", "directive", "why", "signature", "honored", "violated", "class", "mechanizable", "default_C1", "default_C2"}

// catalogKnownKeys are the keys a `·`-joined line splits on: the ten, and the
// optional `intent_locus`. A middle dot followed by anything else is prose.
var catalogKnownKeys = setOf(append([]string{"intent_locus"}, catalogFields...))

var (
	catalogEntryShape = regexp.MustCompile("^- \\*\\*`([^`]+)`\\*\\*")
	catalogKeyShape   = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9_]*):(?:\s+(.*))?$`)
	catalogTagShape   = regexp.MustCompile(`^\*\(([^)]*)\)\*`)
)

// catalogExample is one honored or violated item: its `*(tag)*`, empty when it
// carries none, and its line.
type catalogExample struct {
	tag  string
	line int
}

// parsedCatalogEntry is one parsed catalog entry.
type parsedCatalogEntry struct {
	slug     string
	line     int
	fields   map[string][]string         // key → the value of each occurrence
	examples map[string][]catalogExample // honored / violated → items in order
}

// intentLocus reports whether the entry sets `intent_locus: true`. The field
// is optional and defaults to false (schema-typed-artifacts §1).
func (e *parsedCatalogEntry) intentLocus() bool {
	vs := e.fields["intent_locus"]
	if len(vs) == 0 {
		return false
	}
	words := strings.Fields(strings.ReplaceAll(vs[0], "`", ""))
	return len(words) > 0 && words[0] == "true"
}

// catalogStray is a line under `## Entries` that belongs to no entry or field.
type catalogStray struct {
	line int
	text string
}

// splitCatalogField splits a field's joined text on U+00B7 where the next part
// opens with a known key, and rejoins a part that does not, so a middle dot in
// prose or in a parenthetical stays in its field.
func splitCatalogField(text string) []string {
	var parts []string
	for _, p := range strings.Split(text, "·") {
		p = strings.TrimSpace(p)
		if m := catalogKeyShape.FindStringSubmatch(p); len(parts) == 0 || (m != nil && catalogKnownKeys[m[1]]) {
			parts = append(parts, p)
			continue
		}
		parts[len(parts)-1] += " · " + p
	}
	return parts
}

// parseCatalogEntries reads the entries under a catalog's `## Entries` H2, up
// to the next H2. An entry opens with a top-level bullet naming a backticked
// slug in bold; its fields are two-space bullets, each joined with the lines
// that wrap it; honored and violated items are four-space bullets. Bolding is
// stripped before a field is read. A `###` heading or a blank line closes an
// entry. Anything else is a stray, returned so the caller reports it rather
// than dropping it. ok is false when the section is absent.
func parseCatalogEntries(a *corpusArtifact) (entries []*parsedCatalogEntry, strays []catalogStray, ok bool) {
	start, end, ok := a.section("Entries")
	if !ok {
		return nil, nil, false
	}
	type rawField struct {
		text     string
		examples []catalogExample
	}
	var cur *parsedCatalogEntry
	var raws []*rawField
	inExample := false
	finish := func() {
		if cur == nil {
			return
		}
		for _, rf := range raws {
			for k, p := range splitCatalogField(rf.text) {
				m := catalogKeyShape.FindStringSubmatch(p)
				if m == nil {
					continue
				}
				cur.fields[m[1]] = append(cur.fields[m[1]], strings.TrimSpace(m[2]))
				if k == 0 {
					cur.examples[m[1]] = append(cur.examples[m[1]], rf.examples...)
				}
			}
		}
		entries = append(entries, cur)
		cur, raws, inExample = nil, nil, false
	}
	a.eachBodyLine(func(i int, line string) {
		if i < start || i >= end {
			return
		}
		trimmed := strings.TrimSpace(line)
		indent := len(line) - len(strings.TrimLeft(line, " "))
		switch {
		case trimmed == "", indent == 0 && strings.HasPrefix(line, "#"):
			finish()
		case indent == 0 && catalogEntryShape.MatchString(line):
			finish()
			cur = &parsedCatalogEntry{slug: catalogEntryShape.FindStringSubmatch(line)[1], line: i + 1,
				fields: map[string][]string{}, examples: map[string][]catalogExample{}}
		case cur == nil:
			strays = append(strays, catalogStray{i + 1, trimmed})
		case indent == 2 && strings.HasPrefix(trimmed, "- "):
			inExample = false
			text := strings.ReplaceAll(strings.TrimSpace(trimmed[2:]), "**", "")
			if catalogKeyShape.FindStringSubmatch(text) == nil {
				strays = append(strays, catalogStray{i + 1, trimmed})
				return
			}
			raws = append(raws, &rawField{text: text})
		case indent >= 4 && strings.HasPrefix(trimmed, "- ") && len(raws) > 0:
			tag := ""
			if m := catalogTagShape.FindStringSubmatch(strings.TrimSpace(trimmed[2:])); m != nil {
				tag = m[1]
			}
			last := raws[len(raws)-1]
			last.examples = append(last.examples, catalogExample{tag, i + 1})
			inExample = true
		case indent >= 2 && inExample:
			// An item's wrapped line; only an item's tag is read.
		case indent >= 2 && len(raws) > 0:
			last := raws[len(raws)-1]
			last.text += " " + strings.ReplaceAll(trimmed, "**", "")
		case indent >= 2:
			// The entry heading's wrapped parenthetical.
		default:
			strays = append(strays, catalogStray{i + 1, trimmed})
		}
	})
	finish()
	return entries, strays, true
}

// canonicalCatalogID is the one signature-catalog the corpus may hold
// (schema-typed-artifacts §1: one, shipped). The id is versioned, so a v1 and a
// v2 side by side would revisit rule 8e.
const canonicalCatalogID = "signature-catalog-v1"

// checkTypedArtifacts applies checks 8–11. They apply only when a
// signature-catalog or an expression-profile is in the corpus. Rule 8e: the
// corpus holds one signature-catalog, canonicalCatalogID. Any other is a
// finding, and profiles resolve through the canonical catalog alone, so a
// second catalog cannot supply the entry that hides a rule 11 breach.
func (c *corpusCheck) checkTypedArtifacts() error {
	var profiles, catalogs []*corpusArtifact
	for _, a := range c.arts {
		switch a.scalarValue("type") {
		case "expression-profile":
			profiles = append(profiles, a)
		case "signature-catalog":
			catalogs = append(catalogs, a)
		}
	}
	canonical := c.artifactByID(canonicalCatalogID)
	if canonical != nil && canonical.scalarValue("type") != "signature-catalog" {
		canonical = nil
	}
	if canonical == nil && len(profiles)+len(catalogs) > 0 {
		// The canonical catalog is a shared input: checks 8–11 read it alone.
		return fmt.Errorf("no signature-catalog declares `%s`, and checks 8–11 read that catalog alone: a profile resolved through any other would pass on the wrong entries (rule 8e)", canonicalCatalogID)
	}
	entries := map[string]*parsedCatalogEntry{}
	for _, a := range catalogs {
		if a != canonical {
			line := a.typeLine()
			if f, ok := a.first("id"); ok {
				line = f.line
			}
			c.report(a, line, "8e", a.scalarValue("id"), "the corpus holds %d signature catalogs, and exactly one, `%s`, is allowed; no profile resolves through this one", len(catalogs), canonicalCatalogID)
		}
		es, strays, ok := parseCatalogEntries(a)
		if !ok && a == canonical && len(profiles) > 0 {
			// A shared input: without the entries no gene resolves and no
			// intent_locus is read, so checks 9 and 11 would pass on nothing.
			return fmt.Errorf("%s is a signature-catalog with no `## Entries` section, and %s needs its entries to resolve gene slugs and read intent_locus (checks 9 and 11)", a.path, profiles[0].path)
		}
		if !ok {
			continue // no profile needs it; rule 6a reports the missing section
		}
		for _, s := range strays {
			c.report(a, s.line, "8b", s.text, "%q under `## Entries` belongs to no entry or field, so nothing it carries is read", s.text)
		}
		for _, e := range es {
			c.checkCatalogFields(a, e)
			c.checkCatalogPairs(a, e)
			if a == canonical && entries[e.slug] == nil {
				entries[e.slug] = e
			}
		}
	}
	if len(profiles) > 0 && len(entries) == 0 {
		return fmt.Errorf("%s needs signature-catalog entries to resolve its gene slugs and read intent_locus (checks 9 and 11), and the corpus holds none: no signature-catalog, or one whose `## Entries` yields no entry", profiles[0].path)
	}
	for _, a := range profiles {
		c.checkProfile(a, entries)
	}
	return nil
}

// checkCatalogFields applies rule 8b to one entry: each of the ten fields once,
// with a value. `honored` and `violated` carry their items, not a value.
func (c *corpusCheck) checkCatalogFields(a *corpusArtifact, e *parsedCatalogEntry) {
	for _, f := range catalogFields {
		vs := e.fields[f]
		switch {
		case len(vs) == 0:
			c.report(a, e.line, "8b", e.slug+" "+f, "entry `%s` has no `%s` field", e.slug, f)
		case len(vs) > 1:
			c.report(a, e.line, "8b", e.slug+" "+f, "entry `%s` carries `%s` %d times", e.slug, f, len(vs))
		case vs[0] == "" && f != "honored" && f != "violated":
			c.report(a, e.line, "8b", e.slug+" "+f, "entry `%s` has an empty `%s` field", e.slug, f)
		}
	}
}

// checkCatalogPairs applies rule 8c to one entry: `honored` and `violated`
// form at least two pairs, equal in number, and the i-th of each carries the
// same `*(tag)*`, compared exactly (decision-0027).
func (c *corpusCheck) checkCatalogPairs(a *corpusArtifact, e *parsedCatalogEntry) {
	h, v := e.examples["honored"], e.examples["violated"]
	if len(h) < 2 || len(v) < 2 || len(h) != len(v) {
		c.report(a, e.line, "8c", e.slug+" pairs", "entry `%s` has %d honored and %d violated examples; they must form at least two matched pairs", e.slug, len(h), len(v))
	}
	for i := 0; i < len(h) && i < len(v); i++ {
		if h[i].tag == "" || h[i].tag != v[i].tag {
			c.report(a, v[i].line, "8c", fmt.Sprintf("%s pair %d", e.slug, i+1), "entry `%s` pair %d: the honored tag (%s) and the violated tag (%s) do not match", e.slug, i+1, h[i].tag, v[i].tag)
		}
	}
}

// confidenceTags are the values rule 10a accepts (schema-typed-artifacts §2).
var confidenceTags = map[string]bool{"verified": true, "inferred": true, "speculated": true}

// noEvidence reports whether an evidence cell holds no pointer: blank, or only
// a dash placeholder.
func noEvidence(cell string) bool {
	return strings.Trim(cell, " -—–") == ""
}

// checkProfile applies the profile rules to one expression-profile. A table it
// cannot read is reported under the rule that needs the missing part. Rule 9:
// every gene slug resolves to a catalog entry. Rule 10a: an active,
// honored-implicitly gene carries a confidence tag and an evidence pointer. The
// row's scope is matched case-insensitively, so a capitalized `True` cannot
// exempt it; the confidence value itself is compared exactly.
func (c *corpusCheck) checkProfile(a *corpusArtifact, entries map[string]*parsedCatalogEntry) {
	rows, problems := parseProfileRows(a)
	for _, p := range problems {
		c.report(a, p.line, p.rule, p.subject, "%s", p.detail)
	}
	for _, r := range rows {
		slug := r.cells["slug"]
		if entries[slug] == nil {
			c.report(a, r.line, "9", slug, "gene slug `%s` resolves to no signature-catalog entry", slug)
		}
		if strings.EqualFold(r.cells["active"], "true") && strings.EqualFold(r.cells["basis"], "honored-implicitly") {
			if !confidenceTags[r.cells["confidence"]] {
				c.report(a, r.line, "10a", slug+" confidence", "gene `%s` is active and honored-implicitly, and its confidence %q is not `verified`, `inferred` or `speculated`", slug, r.cells["confidence"])
			}
			if noEvidence(r.cells["evidence"]) {
				c.report(a, r.line, "10a", slug+" evidence", "gene `%s` is active and honored-implicitly, and carries no evidence pointer: a bare \"honored\" claim", slug)
			}
		}
		// Rule 11, active or not: the intent gate never fully opens.
		if e := entries[slug]; e != nil && e.intentLocus() && strings.EqualFold(r.cells["c2"], "none") {
			c.report(a, r.line, "11", slug, "gene `%s` sets `C2: none`, and its catalog entry has `intent_locus: true` (floor-intent-gate)", slug)
		}
	}
}

// profileRow is one row of a profile's `## Profile` table: its line, and its
// cells by lower-cased header name, markers stripped.
type profileRow struct {
	line  int
	cells map[string]string
}

// profileProblem is a profile table the rules cannot read: a missing table, a
// missing column, or a row whose cells do not line up with the header.
type profileProblem struct {
	line                  int
	rule, subject, detail string
}

// profileColumns are the columns checks 9–11 read, located by header name, each
// with the rule that cannot be applied without it.
var profileColumns = []struct{ name, rule string }{
	{"slug", "9"}, {"active", "10a"}, {"basis", "10a"}, {"confidence", "10a"}, {"evidence", "10a"}, {"c2", "11"},
}

// parseProfileRows reads the first table under a profile's `## Profile` H2. A
// profile with no such section yields nothing here; rule 6a reports it.
func parseProfileRows(a *corpusArtifact) ([]profileRow, []profileProblem) {
	start, end, ok := a.section("Profile")
	if !ok {
		return nil, nil
	}
	tables := a.tables(start, end)
	if len(tables) == 0 {
		return nil, []profileProblem{{start, "9", "table", "the `## Profile` section holds no table, so no gene slug in it can be resolved"}}
	}
	t := tables[0]
	col, count := map[string]int{}, map[string]int{}
	for j, h := range t.header {
		name := strings.ToLower(cellValue(h))
		col[name] = j
		count[name]++
	}
	var problems []profileProblem
	for _, c := range profileColumns {
		switch n := count[c.name]; {
		case n == 0:
			problems = append(problems, profileProblem{t.index + 1, c.rule, "column " + c.name,
				"the `## Profile` table has no `" + c.name + "` column, so rule " + c.rule + " cannot be applied to any row"})
		case n > 1:
			// A repeated column would be read as its last copy, so a `none` in an
			// earlier `C2` or a dangling slug in an earlier `slug` would pass unseen.
			problems = append(problems, profileProblem{t.index + 1, c.rule, "duplicate column " + c.name,
				fmt.Sprintf("the `## Profile` table has %d `%s` columns, so rule %s cannot tell which one a row means", n, c.name, c.rule)})
		}
	}
	if len(problems) > 0 {
		return nil, problems
	}
	var rows []profileRow
	for _, r := range t.rows {
		if len(r.cells) != len(t.header) {
			problems = append(problems, profileProblem{r.index + 1, "9", cellValue(r.cells[0]),
				"the row's cells do not line up with the table header, so its columns cannot be read"})
			continue
		}
		cells := map[string]string{}
		for name, j := range col {
			cells[name] = cellValue(r.cells[j])
		}
		rows = append(rows, profileRow{r.index + 1, cells})
	}
	return rows, problems
}

// TestCorpusConformanceReadsLiveCatalogAndProfile pins the typed parsers to the
// live catalog and profile, whose wrapped, bolded and `·`-joined lines are the
// shapes most likely to break them.
func TestCorpusConformanceReadsLiveCatalogAndProfile(t *testing.T) {
	c := liveCorpusCheck(t)
	cat := liveArtifact(t, c, "signature-catalog-v1")
	entries, strays, ok := parseCatalogEntries(cat)
	if !ok {
		t.Fatalf("%s has no `## Entries` section", cat.path)
	}
	for _, s := range strays {
		t.Errorf("%s:%d: %q reads as belonging to no entry", cat.path, s.line, s.text)
	}
	if len(entries) != len(assessableSlugs) {
		t.Errorf("%s: parsed %d entries, and the pinned slug set holds %d", cat.path, len(entries), len(assessableSlugs))
	}
	bySlug := map[string]*parsedCatalogEntry{}
	var intentLocus []string
	for _, e := range entries {
		bySlug[e.slug] = e
		for _, f := range append(append([]string{}, catalogFields...), "intent_locus") {
			switch vs := e.fields[f]; {
			case len(vs) != 1:
				t.Errorf("%s:%d: %s: field `%s` parsed %d times, want once", cat.path, e.line, e.slug, f, len(vs))
			case vs[0] == "" && f != "honored" && f != "violated":
				t.Errorf("%s:%d: %s: field `%s` parsed with no value", cat.path, e.line, e.slug, f)
			}
		}
		if e.intentLocus() {
			intentLocus = append(intentLocus, e.slug)
		}
	}
	if got, want := strings.Join(intentLocus, " "), "inv-intent-locus floor-intent-gate"; got != want {
		t.Errorf("entries with intent_locus true: %q, want %q — the wrapped and bold forms at the two intent-locus entries must both read", got, want)
	}
	tags := func(es []catalogExample) string {
		var out []string
		for _, e := range es {
			out = append(out, e.tag)
		}
		return strings.Join(out, " ")
	}
	for slug, want := range map[string]string{
		"inv-graph-maintenance": "docs research ops code",
		"inv-self-improvement":  "CI process",
		"inv-auditable-archive": "ADR infra",
	} {
		e := bySlug[slug]
		if e == nil {
			t.Errorf("no live catalog entry %s", slug)
			continue
		}
		for _, side := range []string{"honored", "violated"} {
			if got := tags(e.examples[side]); got != want {
				t.Errorf("%s: %s tags parse as %q, want %q", slug, side, got, want)
			}
		}
	}

	profile := liveArtifact(t, c, "profile-trellis-self")
	rows, problems := parseProfileRows(profile)
	for _, p := range problems {
		t.Errorf("%s:%d: %s", profile.path, p.line, p.detail)
	}
	if len(rows) != len(assessableSlugs) {
		t.Errorf("%s: parsed %d profile rows, and the pinned slug set holds %d", profile.path, len(rows), len(assessableSlugs))
	}
	for _, r := range rows {
		if bySlug[r.cells["slug"]] == nil {
			t.Errorf("%s:%d: slug %q resolves to no parsed catalog entry", profile.path, r.line, r.cells["slug"])
		}
	}
	for _, slug := range []string{"inv-intent-locus", "floor-intent-gate"} {
		found := false
		for _, r := range rows {
			if r.cells["slug"] != slug {
				continue
			}
			found = true
			if r.cells["c2"] != "human" {
				t.Errorf("%s:%d: %s reads C2 %q, want human", profile.path, r.line, slug, r.cells["c2"])
			}
			if e := bySlug[slug]; e == nil || !e.intentLocus() {
				t.Errorf("%s resolves to no catalog entry with intent_locus true", slug)
			}
		}
		if !found {
			t.Errorf("%s: no row for %s", profile.path, slug)
		}
	}
}
