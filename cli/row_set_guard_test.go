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
// from these rather than being typed.
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

// TestRowCountProseSitesFollowThePin: every prose site that states the row
// count carries the pin's count, in the shape that site uses — digits, the
// word spelled out, N/N, or the class breakdown. One table row per site; a
// failure names the file and the exact text expected there. The sweep that
// produced this table (TRL-28, 2026-09-03) found 22 sites where the catalog
// note named six files, because `git grep -E '\bsixteen\b'` matches nothing on
// macOS — a pattern is not a guard, a table is.
//
// Template verbs: %[1]d count · %[2]s count spelled · %[3]s count Spelled ·
// %[4]s %[5]s %[6]s the methodology / trellis-design / floor class counts spelled.
func TestRowCountProseSitesFollowThePin(t *testing.T) {
	n := len(assessableSlugs)
	word := numberWord(n)
	m, d, f := catalogClassCounts()
	if m+d+f != n {
		// Errorf, not Fatalf: when the pin grows before the catalog entry lands
		// this fires too, and the site failures below must still be listed.
		t.Errorf("catalog class counts %d+%d+%d do not sum to the pin's %d — the catalog entry has not followed the pin (see TestRowSetDerivativesFollowThePin), and the AC1 breakdown below is checked against the catalog as it stands", m, d, f, n)
	}
	args := []any{n, word, strings.ToUpper(word[:1]) + word[1:], numberWord(m), numberWord(d), numberWord(f)}

	sites := []struct{ path, template string }{
		{"../README.md", "all %[2]s rules, adaptive posture"},
		{"../README.md", "all %[2]s rows active at the"},
		{"../README.md", "governed at %[1]d/%[1]d"},
		{"../plugins/trellis/README.md", "all %[2]s rules at the adaptive posture"},
		{"../plugins/trellis/README.md", "all %[2]s rows active at"},
		{"../install.sh", "all %[2]s rules are active"},
		{"../install.sh", "two rules out of %[2]s"},
		{"../plugins/trellis/hooks/staleness.sh", "— %[1]d rules, followed by default"},
		{"../docs/index.html", "all %[1]d rules active, adaptive posture"},
		{"../docs/index.html", "%[3]s load-bearing invariants"},
		{"../docs/index.html", "See all %[2]s, with why + examples"},
		{"../docs/lp-content.md", "all %[1]d rules active, adaptive posture"},
		{"../docs/lp-content.md", "%[3]s load-bearing"},
		{"../docs/lp-content.md", "%[2]s, with why + examples"},
		{"../docs/invariants.html", "The %[2]s Trellis invariants"},
		{"../docs/invariants.html", "<h1>%[3]s invariants."},
		{"../profiles/trellis-self.md", "All %[1]d assessable genes"},
		{"../core/catalog/signature-catalog-v1.md", "Covers the **%[1]d assessable invariants**"},
		{"../core/catalog/signature-catalog-v1.md", "Covers all **%[1]d assessable** slugs (the %[4]s structural, the %[5]s remaining operating, the %[6]s floors"},
		{"../core/rubrics/artifact-contract.md", "all %[2]s, **excluding** the two dials"},
		{"../.claude/agents/corpus-reviewer.md", "all %[2]s, **excluding** the two dials"},
		{"../plugins/trellis/skills/remove/SKILL.md", "governs it at %[1]d/%[1]d"},
	}
	contents := map[string]string{}
	for _, s := range sites {
		if _, ok := contents[s.path]; !ok {
			contents[s.path] = readFileT(t, s.path)
		}
		want := fmt.Sprintf(s.template, args...)
		if !strings.Contains(contents[s.path], want) {
			t.Errorf("%s does not carry %q — the row count is %d; update this site to that count in this shape [TRL-28]",
				strings.TrimPrefix(s.path, "../"), want, n)
		}
	}

	// Negative sweep, pure-doc files only: any other count in the row-count range
	// next to the row nouns, or any N/N with N ≠ the pin, is a stale site — one the
	// table above does not know about yet. install.sh and the hooks are excluded:
	// they carry comments that narrate historical counts on purpose.
	docs := []string{
		"../README.md", "../plugins/trellis/README.md", "../docs/index.html", "../docs/lp-content.md",
		"../docs/invariants.html", "../profiles/trellis-self.md", "../core/catalog/signature-catalog-v1.md",
		"../core/rubrics/artifact-contract.md", "../.claude/agents/corpus-reviewer.md",
		"../plugins/trellis/skills/remove/SKILL.md",
	}
	numAlt := `(1[0-9]|2[0-9]|ten|eleven|twelve|thirteen|fourteen|fifteen|sixteen|seventeen|eighteen|nineteen|twenty(?:-(?:one|two|three|four|five|six|seven|eight|nine))?)`
	staleRe := regexp.MustCompile(`(?i)\b` + numAlt + `\b[^\n]{0,20}?\b(?:rules?|rows?|invariants?|genes?|slugs?)\b`)
	nnRe := regexp.MustCompile(`\b(1[0-9]|2[0-9])/(1[0-9]|2[0-9])\b`)
	isPin := func(tok string) bool {
		tok = strings.ToLower(tok)
		return tok == fmt.Sprint(n) || tok == word
	}
	for _, path := range docs {
		body, ok := contents[path]
		if !ok {
			body = readFileT(t, path)
		}
		for i, line := range strings.Split(body, "\n") {
			for _, mm := range staleRe.FindAllStringSubmatch(line, -1) {
				if !isPin(mm[1]) {
					t.Errorf("%s:%d: %q names a row count that is not the pin's (%d) — a stale count, or a new site missing from the table above [TRL-28]",
						strings.TrimPrefix(path, "../"), i+1, mm[0], n)
				}
			}
			for _, mm := range nnRe.FindAllStringSubmatch(line, -1) {
				if !isPin(mm[1]) || !isPin(mm[2]) {
					t.Errorf("%s:%d: %q is not %d/%d — a stale N/N count [TRL-28]",
						strings.TrimPrefix(path, "../"), i+1, mm[0], n, n)
				}
			}
		}
	}
}
