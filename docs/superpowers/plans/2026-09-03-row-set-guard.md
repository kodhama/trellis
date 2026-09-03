# Row-set guard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A row-set change (mint or retire an invariant slug) fails loudly, naming every derivative and every prose count that did not follow, instead of relying on a hand-kept list in the catalog.

**Architecture:** One new Go test file, `cli/row_set_guard_test.go`, with two tests that read `assessableSlugs` (`cli/payload_test.go`) as the single source: a set-equality guard over the structural derivatives, and a per-site phrase table over the 22 prose count sites plus a bounded negative sweep. The catalog's maintainer note is rewritten to point at the guard. No prose site, hook, `install.sh`, `VERSION` or `plugin.json` is edited.

**Tech Stack:** Go 1.x test suite in `cli/` (stdlib only: `os`, `regexp`, `strings`, `fmt`, `testing`).

**Spec:** `docs/superpowers/specs/2026-09-03-row-set-guard-design.md`

## Global Constraints

- Every test run is `go test -count=1` (`AGENTS.md`: the hook tests execute files Go's cache does not track).
- Do not edit `plugins/trellis/hooks/*`, `plugins/trellis/VERSION`, `install.sh`, `plugin.json` (owned by #261/#262/#263).
- `cli/assets/invariants.md` is generated — edit `core/catalog/signature-catalog-v1.md` and run `go generate ./...` in `cli/`.
- After the catalog edit, `git diff --stat -- plugins/trellis/reference/` must be empty; if `reference/version` moved, stop and comment on TRL-28.
- No decision record in this change; no `status:` field anywhere (`decision-0082`).
- Commit messages end with the Co-Authored-By / Claude-Session trailers used on this branch.

---

### Task 1: The structural guard — `TestRowSetDerivativesFollowThePin`

**Files:**
- Create: `cli/row_set_guard_test.go`
- Modify: `cli/selfapply_test.go:26-41` (`TestRepoDeclaresRulesConfig` — retire the per-slug loop)

**Interfaces:**
- Consumes: `assessableSlugs []string` (`cli/payload_test.go:62`), `catalogSlugOrder() []string` (`cli/apply.go:279`), `readFileT(t, path) string` (`cli/install_script_test.go:122`), `vendoredPayloadDir` (`cli/payload_test.go:51`).
- Produces: `rowSetDiff(want, got []string) (missing, extra []string)` and `captureAll(re *regexp.Regexp, s string) []string`; Task 2 appends to the same file and shares its imports.

- [ ] **Step 1: Write the test file with the structural guard**

Create `cli/row_set_guard_test.go`:

```go
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
	"os"
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
```

- [ ] **Step 2: Run it — it must pass on `main` as it stands**

Run: `cd cli && go test -count=1 -run 'TestRowSetDerivativesFollowThePin' -v ./...`
Expected: PASS. If any derivative fails here, the repo is already drifted — stop and record it on TRL-28 before going further.

- [ ] **Step 3: Mutation-prove — a fake slug in the pin only**

Temporarily append `"inv-zzz-fake",` to `assessableSlugs` in `cli/payload_test.go`, then:

Run: `cd cli && go test -count=1 -run 'TestRowSetDerivativesFollowThePin' ./... 2>&1 | grep -c 'missing: \[inv-zzz-fake\]'`
Expected: `7` — every one of the seven derivatives is named. Revert the pin (`git checkout cli/payload_test.go`).

- [ ] **Step 4: Mutation-prove — a real slug removed from one preset only**

Delete the `inv-minimal-first` row from `plugins/trellis/reference/rules-a.toml`, then:

Run: `cd cli && go test -count=1 -run 'TestRowSetDerivativesFollowThePin' ./... 2>&1 | grep 'rules-a.toml'`
Expected: one failure naming `rules-a.toml` with `missing: [inv-minimal-first]`; no other derivative named. Restore: `git checkout plugins/trellis/reference/rules-a.toml`.

- [ ] **Step 5: Retire the redundant forward loop in `TestRepoDeclaresRulesConfig`**

In `cli/selfapply_test.go`, replace the `for _, slug := range assessableSlugs { ... }` block (lines 35–40) with:

```go
	// The per-slug row check moved to TestRowSetDerivativesFollowThePin
	// (row_set_guard_test.go), which checks the set both ways — a stale row after
	// a retire failed nothing here. Strictness stays: it is this file's posture,
	// not the row set's.
```

If `regexp` is now unused in the file, drop it from the import block (`go vet` will say).

- [ ] **Step 6: Run the two tests together, plus vet**

Run: `cd cli && go test -count=1 -run 'TestRowSetDerivativesFollowThePin|TestRepoDeclaresRulesConfig' ./... && go vet ./... && gofmt -l .`
Expected: PASS, no vet output, no gofmt output.

- [ ] **Step 7: Commit**

```bash
git add cli/row_set_guard_test.go cli/selfapply_test.go
git commit -m "TRL-28: guard the row set — every derivative against the one pin, both ways"
```

---

### Task 2: The prose-count guard — `TestRowCountProseSitesFollowThePin`

**Files:**
- Modify: `cli/row_set_guard_test.go` (append)

**Interfaces:**
- Consumes: `assessableSlugs`, `readFileT`, `invariantsRef` (`cli/apply.go`, the embedded catalog), `vendoredPayloadDir`.
- Produces: `numberWord(n int) string` — lowercase English for 0–29.

- [ ] **Step 1: Append the prose guard to the test file**

```go
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
		// this fires too, and the 22 site failures below must still be listed.
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
```

- [ ] **Step 2: Run it — it must pass on `main` as it stands**

Run: `cd cli && go test -count=1 -run 'TestRowCountProseSitesFollowThePin' -v ./...`
Expected: PASS. If the negative sweep fires on a legitimate non-count (e.g. a year, a line number), tighten the regex window or noun list — do not add an exclusion list keyed on file content. If a positive template fails, re-read that site's exact line and fix the template, not the site.

- [ ] **Step 3: Mutation-prove — a wrong count at one site**

Change `16/16` to `15/15` at `README.md:70`, then:

Run: `cd cli && go test -count=1 -run 'TestRowCountProseSitesFollowThePin' ./... 2>&1 | grep 'README.md'`
Expected: two lines — the positive check (`does not carry "governed at 16/16"`) and the negative sweep (`"15/15" is not 16/16`). Restore: `git checkout README.md`.

- [ ] **Step 4: Mutation-prove — the pin grows**

Append `"inv-zzz-fake",` to `assessableSlugs`, then:

Run: `cd cli && go test -count=1 -run 'TestRowCountProseSitesFollowThePin' ./... 2>&1 | grep -c 'does not carry'`
Expected: `22` — every site named. The class-sum check also fires once (the fake slug has no catalog `class:` line, so 4+10+2 ≠ 17); that is the `Errorf` doing its job, not a masked failure. Also expect `grep -c 'is not the pin'` > 0 (every "sixteen" is now stale). Revert the pin (`git checkout cli/payload_test.go`).

- [ ] **Step 5: Run everything in the file, vet, fmt**

Run: `cd cli && go test -count=1 -run 'TestRowSet|TestRowCount' ./... && go vet ./... && gofmt -l .`
Expected: PASS, no output from vet or gofmt.

- [ ] **Step 6: Commit**

```bash
git add cli/row_set_guard_test.go
git commit -m "TRL-28: pin every prose row count to the one pin, in the shape each site uses"
```

---

### Task 3: The catalog note points at the guard

**Files:**
- Modify: `core/catalog/signature-catalog-v1.md:55-67` (the "Adding or removing a row" paragraph)
- Regenerate: `cli/assets/invariants.md` (`go generate ./...` in `cli/`)

**Interfaces:**
- Consumes: the test names from Tasks 1–2 (`TestRowSetDerivativesFollowThePin`, `TestRowCountProseSitesFollowThePin`) and the file `cli/row_set_guard_test.go`.

- [ ] **Step 1: Replace the paragraph**

Replace lines 55–67 of `core/catalog/signature-catalog-v1.md` (from `> **Adding or removing a row costs more than an example edit does**` through `every count there now derives from one pinned list.`) with:

```markdown
> **Adding or removing a row costs more than an example edit does — and the sweep is guarded,
> not listed.** `cli/row_set_guard_test.go` reads the pinned slug set (`assessableSlugs`,
> `cli/payload_test.go` — the one pin) and fails naming what has not followed: every derivative
> that carries the set — the `invariants-v1` registry, `profiles/trellis-self.md`, the
> `docs/invariants.html` cards, this repo's `.trellis/rules.toml`, the rendered `reference/rules-*.toml`
> (**without a row the rule ships but is inactive**) — and every prose site that states the count
> (the READMEs, `install.sh`, the hooks' announcements, `docs/`, this catalog's Coverage note and
> AC1, the contract, the reviewer charter, the remove skill), each in the shape it uses there:
> digits, the word spelled out, `N/N`, the class breakdown. Run `go test -count=1 ./...` in `cli/`
> and follow the failures. The list lives in the test because a list kept here was found short
> twice (`decision-0074`, `decision-0078`) and the sweep that rebuilt it found 22 sites where this
> note named six files (TRL-28). Two obligations the guard cannot see: a new card in
> `docs/invariants.html` needs its *examples* rendered (`cli/sync_test.go` catches those), and the
> release stamp `plugins/trellis/VERSION` (**unguarded — trellis#245 is still open**; without it
> every cached consumer keeps the old rule set, `d4a2c7b`). The Codex hook is not a surface: it
> derives its slug set from the generated `reference/rules.md` since `decision-0083`.
```

- [ ] **Step 2: Regenerate the bundled copy and check the payload did not move**

Run: `cd cli && go generate ./... && cd .. && git status --short && git diff --stat -- plugins/trellis/reference/`
Expected: `core/catalog/signature-catalog-v1.md` and `cli/assets/invariants.md` modified; the `reference/` diff-stat is **empty** (the note is preamble; `extractEntriesSection` excludes it). If `reference/version` shows, stop and comment on TRL-28 — do not bump `VERSION`.

- [ ] **Step 3: Run the sync tests and the new guards**

Run: `cd cli && go test -count=1 -run 'TestBundledCatalogInSync|TestRowSet|TestRowCount|TestPayload' ./...`
Expected: PASS (the Coverage note at `:37` is untouched, so the prose guard's two catalog sites still hold).

- [ ] **Step 4: Commit**

```bash
git add core/catalog/signature-catalog-v1.md cli/assets/invariants.md
git commit -m "TRL-28: the catalog's row-set note points at the guard instead of restating the list"
```

---

### Task 4: Proofs, full suite, corpus review, PR

**Files:**
- No new files. Reads everything; edits nothing except a dated note on the spec if a claim turned out wrong (`decision-0085` point 5).

- [ ] **Step 1: Not over-constrained — an example edit passes**

Edit one `- *(honored)*` example line in `core/catalog/signature-catalog-v1.md` (append ` — probe`), then:

Run: `cd cli && go test -count=1 -run 'TestRowSet|TestRowCount' ./...`
Expected: PASS — an example edit changes no slug and no count. (`TestBundledCatalogInSync` would fail on the un-regenerated copy; that is the existing guard doing its job, not this one over-firing.) Restore: `git checkout core/catalog/signature-catalog-v1.md`.

- [ ] **Step 2: Full suite, build, vet**

Run: `cd cli && go test -count=1 ./... && go build ./... && go vet ./... && gofmt -l .`
Expected: all PASS; no vet or gofmt output.

- [ ] **Step 3: Rebase on `origin/main`**

Run: `git fetch origin main && git rebase origin/main`
Expected: clean, or textual conflicts only in files this branch does not edit (none expected — this branch touches `cli/*_test.go`, the catalog, its generated copy, and `docs/superpowers/`). Re-run Step 2 after any rebase that brought changes.

- [ ] **Step 4: Corpus review (touched `core/`)**

Dispatch the repo-owned `corpus-reviewer` agent over `core/catalog/signature-catalog-v1.md`. It is read-only; record its verdict. Fix only conformance findings in the note itself.

- [ ] **Step 5: Push and open a draft PR**

```bash
git push -u origin feature/trl-28-row-set-guard
```

Open a **draft** PR titled `TRL-28: guard the row set from the one pin — 22 prose sites and 7 derivatives`. Body: the measured counts (22 sites / 12 files; the `git grep \b` finding), the mechanism, the four mutation results with the exact failure counts observed, the over-constraint proof, `reference/version` unmoved, and "no decision record — `decision-0028` applied (PR #245 precedent); ask on the ticket for `0089` if wanted". End with the generated-with footer used on this repo's PRs.

- [ ] **Step 6: Report on TRL-28 and mark the PR ready**

Comment on TRL-28 with the PR number and the same facts. Mark the PR ready for review. Do not merge.
