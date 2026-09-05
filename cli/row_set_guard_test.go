package main

// The row-set guard (TRL-28). Minting or retiring an invariant slug used to
// oblige a hand-swept edit across a dozen surfaces, enforced only by a prose
// note in the catalog that decision-0074's and decision-0078's reviews each
// found short. Both tests here read the one pin — assessableSlugs in
// payload_test.go — so a miss is red instead of a review comment
// (decision-0028: a guard per source↔derivative pair).
//
// The surface was SHRUNK before the remainder was guarded (the maintainer's
// ruling on TRL-28, 2026-09-04). A first pass tabled 22 prose sites; review
// found most of them stated a count the sentence did not need, so those
// numerals were deleted ("all sixteen rules" → "all rules") rather than
// pinned. What is left below is only the prose where the NUMBER IS THE CLAIM:
// two runtime strings a user actually reads, one headline claim, and the
// catalog's own acceptance criterion. Do not grow this table by adding a count
// to prose — delete the count instead.
//
// Nothing here edits a surface; the tests read them. When one of these fails,
// the failure names the file and the exact expected text — follow it.

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// rowSetDiff returns the pin slugs a derivative lacks and the slugs it carries
// that the pin does not, both sorted, so a failure reads as an edit list.
func rowSetDiff(want, got []string) (missing, extra []string) {
	w := map[string]bool{}
	for _, s := range want {
		w[s] = true
	}
	g := map[string]bool{}
	for _, s := range got {
		g[s] = true
	}
	for s := range w {
		if !g[s] {
			missing = append(missing, s)
		}
	}
	for s := range g {
		if !w[s] {
			extra = append(extra, s)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	return missing, extra
}

// captureAll returns capture group 1 of every match of re in s.
func captureAll(re *regexp.Regexp, s string) []string {
	var out []string
	for _, m := range re.FindAllStringSubmatch(s, -1) {
		out = append(out, m[1])
	}
	return out
}

// TestRowSetDerivativesFollowThePin: every surface that carries the slug set
// carries exactly the pinned set — missing AND extra both fail, so a retire is
// caught as well as a mint. The rendered TOMLs and the catalog are pinned
// elsewhere too (TestPayloadRulesTomlSeeds, TestVendoredPayloadIsCurrent,
// rules_test.go); they are listed here so this one test's output is the
// complete map of what has not followed.
//
// Membership is the contract for every derivative but one: the profile also
// carries a per-row VALUE, and profiles/trellis-self.md claims in prose that
// every gene is active, so its pattern matches `| true |` only — a row flipped
// to false reads here as missing, which is what that claim means.
func TestRowSetDerivativesFollowThePin(t *testing.T) {
	tomlRowRe := regexp.MustCompile(`(?m)^([a-z][a-z-]*)\s*= \{`)
	derivatives := []struct {
		name string
		got  []string
		fix  string
	}{
		{
			// catalogSlugOrder parses the EMBEDDED copy (cli/assets/invariants.md,
			// generated from this file by `go generate`), not the file on disk.
			// Equivalent because TestBundledCatalogInSync fails whenever the two
			// differ by a byte; named for the source you would edit.
			name: "core/catalog/signature-catalog-v1.md (entries, via the embedded copy)",
			got:  catalogSlugOrder(),
			fix:  "add or remove the catalog entry, then `go generate ./...` and regenerate the payload",
		},
		{
			// Live entries head `- **`slug` — title**`; the collapsed
			// inv-reference-relationship heads `- **`slug` → collapsed**` and the
			// dials are not rows, so shape and prefix exclude both.
			name: "core/invariants/trellis-invariants-v1.md (live entries)",
			got: captureAll(regexp.MustCompile("(?m)^- \\*\\*`((?:inv|floor)-[a-z-]+)` — "),
				readFileT(t, "../core/invariants/trellis-invariants-v1.md")),
			fix: "a slug is a set amendment — add or retire the registry entry (and its legacy-map row if any)",
		},
		{
			name: "profiles/trellis-self.md (active table rows)",
			got: captureAll(regexp.MustCompile("(?m)^\\| `([a-z][a-z-]*)` \\| true \\|"),
				readFileT(t, "../profiles/trellis-self.md")),
			fix: "add the profile row, or set its `active` back to true — the reference organism assesses every gene AND holds it active (the Profile note says so in prose); a row present but false reads as missing here",
		},
		{
			name: "docs/invariants.html (cards)",
			got: captureAll(regexp.MustCompile(`<span class="code">([a-z-]+)</span>`),
				readFileT(t, "../docs/invariants.html")),
			fix: "add or remove the card; TestInvariantsPageMatchesCatalog checks its examples",
		},
		{
			name: ".trellis/rules.toml (this repo's rows)",
			got:  captureAll(tomlRowRe, readFileT(t, "../.trellis/rules.toml")),
			fix:  "add or remove the row — without a row the rule ships but is inactive",
		},
		{
			name: "plugins/trellis/reference/rules-a.toml",
			got:  captureAll(tomlRowRe, readFileT(t, vendoredPayloadDir+"/rules-a.toml")),
			fix:  "regenerate the payload (`go run . payload --out ../plugins/trellis/reference`)",
		},
		{
			name: "plugins/trellis/reference/rules-b.toml",
			got:  captureAll(tomlRowRe, readFileT(t, vendoredPayloadDir+"/rules-b.toml")),
			fix:  "regenerate the payload (`go run . payload --out ../plugins/trellis/reference`)",
		},
	}
	for _, d := range derivatives {
		missing, extra := rowSetDiff(assessableSlugs, d.got)
		if len(missing) == 0 && len(extra) == 0 {
			continue
		}
		t.Errorf("%s: row set differs from the pin (assessableSlugs, cli/payload_test.go)\n"+
			"  missing: %v\n  extra:   %v\n  fix: %s [decision-0028; TRL-28]",
			d.name, missing, extra, d.fix)
	}
}

// numberWord spells n out in lowercase English for the range the docs use.
// Outside 0–29 the guard itself would be wrong, not the docs — fail loudly.
func numberWord(n int) string {
	ones := []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine",
		"ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen"}
	switch {
	case n >= 0 && n < 20:
		return ones[n]
	case n >= 20 && n < 30:
		if n == 20 {
			return "twenty"
		}
		return "twenty-" + ones[n-20]
	}
	panic(fmt.Sprintf("numberWord: %d is outside the range this guard spells (0–29) — extend it", n))
}

// catalogClassCounts counts the catalog entries by `class:` — the AC1 breakdown
// ("the four structural, the ten remaining operating, the two floors") derives
// from these rather than being typed. Like catalogSlugOrder it reads the embedded
// copy while the AC1 site below is read from disk; TestBundledCatalogInSync holds
// the two byte-identical, so the check is against one catalog, not two.
func catalogClassCounts() (methodology, design, floor int) {
	re := regexp.MustCompile("(?m)^  - class: `(methodology|trellis-design|floor)`")
	for _, m := range re.FindAllStringSubmatch(invariantsRef, -1) {
		switch m[1] {
		case "methodology":
			methodology++
		case "trellis-design":
			design++
		case "floor":
			floor++
		}
	}
	return
}

// TestRowCountProseSitesFollowThePin: the few places where the row count IS the
// claim carry the pin's count, in the shape that site uses. A failure names the
// file and the exact text expected there.
//
// Almost everywhere else the count was deleted rather than pinned, so this table
// is deliberately short — six sites, not the twenty-two a first pass tabled. Two
// are runtime strings the user reads (the installer's closing line, the hook's
// not-yet-governing announcement, neither of which can say "all rules" without
// losing the point); one is the README's headline 16/16; one is the catalog's
// own acceptance criterion, where the class breakdown is the claim being
// accepted. The last two are the exception that proves the rule: sentences in
// plugins/trellis/README.md whose numeral SHOULD have been deleted, pinned
// instead because the deletion is blocked on a file another PR owns (see the
// note beside them). The remove skill's own 16/16 is already derived from the
// pin by remove_skill_test.go, so it is not repeated here.
//
// There is no sweep for sites this table does not know: at four sites the
// stopping rule is the review, and a pattern broad enough to find a new count
// also failed legitimate prose ("twelve of the sixteen are methodology rules")
// while still missing the noun-first shapes. Delete the count instead of
// adding a row.
//
// Template verbs: %[1]d count · %[2]s count spelled · %[3]s %[4]s %[5]s the
// methodology / trellis-design / floor class counts spelled.
func TestRowCountProseSitesFollowThePin(t *testing.T) {
	n := len(assessableSlugs)
	m, d, f := catalogClassCounts()
	if m+d+f != n {
		// Errorf, not Fatalf: when the pin grows before the catalog entry lands
		// this fires too, and the site failures below must still be listed.
		t.Errorf("catalog class counts %d+%d+%d do not sum to the pin's %d — the catalog entry has not followed the pin (see TestRowSetDerivativesFollowThePin), and the AC1 breakdown below is checked against the catalog as it stands", m, d, f, n)
	}
	args := []any{n, numberWord(n), numberWord(m), numberWord(d), numberWord(f)}

	sites := []struct{ path, template, why string }{
		{"../README.md", "governed at %[1]d/%[1]d",
			"the headline claim for what one install leaves you governed at"},
		{"../install.sh", "all %[2]s rules are active",
			"the installer says this to the user as it finishes"},
		{"../plugins/trellis/hooks/staleness.sh", "— %[1]d rules, followed by default",
			"the hook's announcement, quoted to a user who has not adopted yet"},
		{"../core/catalog/signature-catalog-v1.md",
			"Covers all **%[1]d assessable** slugs (the %[3]s structural, the %[4]s remaining operating, the %[5]s floors",
			"AC1 — the coverage claim the catalog is accepted against"},
		// Deferred deletions, pinned so they cannot go stale while they wait.
		// These two sentences do not need their numeral, but deleting it edits a
		// file inside the shipped bundle, which forces install.sh's baked manifest
		// to be re-hashed — and those manifest lines are owned by open trellis#262.
		// Pinned to the strings present verbatim today; when a change that already
		// touches the bundle deletes the numerals, delete these two rows with them.
		{"../plugins/trellis/README.md", "all %[2]s rules at the adaptive posture",
			"deferred deletion — re-baking install.sh's manifest is owned by trellis#262"},
		{"../plugins/trellis/README.md", "all %[2]s rows active at",
			"deferred deletion — re-baking install.sh's manifest is owned by trellis#262"},
	}
	contents := map[string]string{}
	for _, s := range sites {
		if _, ok := contents[s.path]; !ok {
			contents[s.path] = readFileT(t, s.path)
		}
		want := fmt.Sprintf(s.template, args...)
		if !strings.Contains(contents[s.path], want) {
			t.Errorf("%s does not carry %q — the row count is %d, and this site states it (%s); update the site to that count in this shape, or delete the numeral if the sentence no longer needs it [TRL-28]",
				strings.TrimPrefix(s.path, "../"), want, n, s.why)
		}
	}
}
