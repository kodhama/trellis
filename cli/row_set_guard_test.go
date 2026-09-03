package main

// The row-set guard (TRL-28). Minting or retiring an invariant slug used to
// oblige a hand-swept edit across a dozen surfaces, enforced only by a prose
// note in the catalog that decision-0074's and decision-0078's reviews each
// found short. The note now points here. Both tests read the one pin —
// assessableSlugs in payload_test.go — so every derivative and every prose
// count is checked against the same list, and a miss is red instead of a
// review comment (decision-0028: a guard per source↔derivative pair).
//
// Nothing here edits a surface; the tests read them. When one of these fails,
// the failure names the file and the exact expected text — follow it.

import (
	"regexp"
	"sort"
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
func TestRowSetDerivativesFollowThePin(t *testing.T) {
	tomlRowRe := regexp.MustCompile(`(?m)^([a-z][a-z-]*)\s*= \{`)
	derivatives := []struct {
		name string
		got  []string
		fix  string
	}{
		{
			name: "core/catalog/signature-catalog-v1.md (entries)",
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
			name: "profiles/trellis-self.md (table rows)",
			got: captureAll(regexp.MustCompile("(?m)^\\| `([a-z][a-z-]*)` \\| (?:true|false) \\|"),
				readFileT(t, "../profiles/trellis-self.md")),
			fix: "add or remove the profile row — the reference organism assesses every gene",
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
