package main

// Tests for the release-render pipeline — kodhama/trellis#117 (kodhama-0007 slice 1,
// "one render, many copiers"), reshaped by decision-0051 (the overlay splits by
// authority: consumer-owned rules.toml at .trellis/, generated files under
// .trellis/internal/) and again by decision-0053 (live rows: the readout ships
// complete with an authority header; rules.toml rows govern at read time; fragment
// assembly retires, and the fragments leave the shipped payload — no consumer
// remains). Upstream anchors:
//   - kodhama-0007 rule 1 (render once, at release: the full enumerable variant space
//     is pre-rendered into the vendored payload) → TestPayloadFileSet,
//     TestVendoredPayloadIsCurrent.
//   - kodhama-0007 rule 3 (verification is data: a checksum manifest anything can
//     check with standard tools) → TestPayloadManifestVerifies,
//     TestVendoredPayloadManifestVerifies.
//   - decision-0051 rule 1 (authority split; internal/trellis.md imports only its
//     sibling rules.md — @import paths resolve relative to the importing file, so no
//     ../ traversal) → TestPayloadHeaderImportsSiblingRules.
//   - decision-0053 point 2 (import channel: the managed block imports both
//     .trellis/internal/trellis.md and .trellis/rules.toml; inline channel: the block
//     is the rows-inlined sandwich) → TestPayloadBlockCarriesBothImports,
//     TestPayloadInlineBlockIsRowsInlinedSandwich.
//   - decision-0053 point 2 (the readout ships complete and carries the eval-tested
//     authority header, research-0012) → TestPayloadReadoutIsCompleteWithAuthorityHeader,
//     TestPayloadRulesReadoutIsOrderedConcatenation.
//   - decision-0053 point 4 (no shipped text claims refresh-time semantics for rows;
//     the absence-era preamble/footer/tail/toml comments retired) →
//     TestPayloadShipsNoRefreshTimeRowClaims, TestPayloadRulesTomlSeeds.
//   - decision-0053 Consequences (SKILL.md step 4's selection cat → a plain copy of
//     rules.md) → TestSetupSkillCopiesCompleteReadout.
//   - decision-0051 rules 2+3 (rules.toml is posture-as-seed, rows-as-truth; floors
//     are not rows a consumer can turn off — held by the authority header's floor
//     sentence since decision-0053) → TestPayloadRulesTomlSeeds.
//   - kodhama-0007 rider (the invariants pointer stays a trigger in the always-on
//     templates) → TestPayloadCarriesInvariantsTrigger.
//   - #117 scope (the generator is the existing Go render code, runnable in CI) →
//     TestPayloadCommandWritesPayload.

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// vendoredPayloadDir is the payload home named by #117: plugins/trellis/reference/.
const vendoredPayloadDir = "../plugins/trellis/reference"

// assessableSlugs is the pinned catalog slug set (signature-catalog-v1: the 14
// assessable invariants), alphabetical — the complete readout must carry one
// slug-tagged rule per slug (decision-0053 point 1).
var assessableSlugs = []string{
	"floor-intent-gate",
	"floor-transparency",
	"inv-auditable-archive",
	"inv-bounded-context",
	"inv-clarify-before-commit",
	"inv-directional-flow",
	"inv-gate-at-handover",
	"inv-graph-maintenance",
	"inv-handover-points",
	"inv-independent-judgment",
	"inv-intent-locus",
	"inv-minimal-first",
	"inv-ratifiable-artifacts",
	"inv-self-improvement",
}

// TestPayloadFileSet: the generator emits exactly the decision-0053 file set — the
// verbatim catalog, the complete readout, both posture variants of the header /
// inline block / rules.toml seed, the constant CLAUDE.md block, the version stamp,
// and the manifest. No expression seed (retired, decision-0051 amendment) and no
// rules/ fragment files (they left the shipped payload with decision-0053 point 1:
// assembly retired and no consumer remains — the setup skill copies rules.md whole,
// the manual copy path needs no rebuild, and both eval runners read only the files
// listed here; the exact-match list is the guard that both stay gone).
func TestPayloadFileSet(t *testing.T) {
	want := []string{
		"block-claude.md",
		"block-inline-a-head.md",
		"block-inline-a.md",
		"block-inline-b-head.md",
		"block-inline-b.md",
		"block-inline-tail.md",
		"checksums",
		"invariants.md",
		"rules-a.toml",
		"rules-b.toml",
		"rules.md",
		"trellis-a.md",
		"trellis-b.md",
		"version",
	}
	sort.Strings(want)

	files := payloadFiles()
	got := make([]string, 0, len(files))
	for name := range files {
		got = append(got, name)
	}
	sort.Strings(got)
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("payload file set mismatch\n got:  %v\n want: %v", got, want)
	}
	for name, content := range files {
		if content == "" {
			t.Errorf("payload file %s rendered empty", name)
		}
	}
}

// TestPayloadVariantsAreTheRenderedPostures: each payload file is the single render
// path's output for its posture — the catalog with its leading frontmatter block
// stripped (decision-0054 point 1: invariantsRef itself is untouched; only the
// payload-write site strips it), the posture-a/b strictness split in the
// always-on templates, the constant CLAUDE.md block (#117's verified
// variant-space evidence; the readout itself is posture-independent, decision-0051
// open question "fragment granularity vs. posture").
func TestPayloadVariantsAreTheRenderedPostures(t *testing.T) {
	files := payloadFiles()

	if files["invariants.md"] != stripFrontmatter(invariantsRef) {
		t.Error("payload invariants.md must be the bundled catalog with its leading frontmatter stripped (decision-0054)")
	}
	if files["invariants.md"] == invariantsRef {
		t.Error("payload invariants.md must not be the bundled catalog verbatim — frontmatter should be stripped (decision-0054)")
	}
	if files["block-claude.md"] != renderClaudeBlock() {
		t.Error("payload block-claude.md must be the constant CLAUDE.md block")
	}
	if !strings.Contains(files["trellis-a.md"], "**Firmly**") {
		t.Error("trellis-a.md missing posture A's enforced strictness line")
	}
	if !strings.Contains(files["trellis-b.md"], "**By default**") {
		t.Error("trellis-b.md missing posture B's default-on strictness line")
	}
	if !strings.Contains(files["block-inline-a.md"], "**Firmly**") ||
		!strings.Contains(files["block-inline-b.md"], "**By default**") {
		t.Error("inline blocks missing their posture strictness lines")
	}
	for _, name := range []string{"block-inline-a.md", "block-inline-b.md"} {
		if !strings.Contains(files[name], trellisBegin) || !strings.Contains(files[name], trellisEnd) {
			t.Errorf("%s must be a complete managed block (begin/end markers)", name)
		}
	}
	if !strings.Contains(files["rules.md"], "settled ground") {
		t.Error("rules.md missing the active rule directives")
	}
}

// TestPayloadBlockCarriesBothImports: decision-0053 point 2 (import channel) — the
// managed CLAUDE.md block imports both @.trellis/internal/trellis.md and
// @.trellis/rules.toml, as block-level imports (the empirically tested shape,
// research-0012's prerequisite check: a .toml @import loads into context; no
// nested-import dependency). The rules import comes after the header import so the
// rows land below the rules, matching the authority header's claim. expression.md
// stays retired (decision-0051 amendment). The inline blocks stay import-free —
// they exist precisely for files without @import support.
func TestPayloadBlockCarriesBothImports(t *testing.T) {
	files := payloadFiles()
	block := files["block-claude.md"]
	i := strings.Index(block, "@.trellis/internal/trellis.md")
	j := strings.Index(block, "@.trellis/rules.toml")
	if i < 0 {
		t.Fatalf("block-claude.md must import @.trellis/internal/trellis.md (decision-0051 rule 1): %q", block)
	}
	if j < 0 {
		t.Fatalf("block-claude.md must import @.trellis/rules.toml — the live-rows delivery (decision-0053 point 2): %q", block)
	}
	if j < i {
		t.Errorf("block-claude.md must import the rows after the header, so they land below the rules: %q", block)
	}
	if strings.Contains(block, "expression") {
		t.Errorf("block-claude.md must not reference expression.md — retired from the bundle (decision-0051 amendment): %q", block)
	}
	if strings.Contains(block, "@.trellis/trellis.md\n") {
		t.Errorf("block-claude.md still imports the retired flat-layout path .trellis/trellis.md: %q", block)
	}
	for _, name := range []string{"block-inline-a.md", "block-inline-b.md", "block-inline-a-head.md", "block-inline-b-head.md", "block-inline-tail.md"} {
		if strings.Contains(files[name], "@.trellis/") || strings.Contains(files[name], "@expression.md") || strings.Contains(files[name], "@rules.md") {
			t.Errorf("%s is (part of) the no-@import variant and must not carry an import line", name)
		}
	}
}

// TestPayloadInlineBlockIsRowsInlinedSandwich: decision-0053 point 2 (inline
// channel) — the block inlines the rows below the rules, exactly the experiment's
// annotation/control-arm sandwich (research-0012 run.sh's overlay build), now the
// shipped shape: head + the complete readout + an "## Active rows" section carrying
// the rules.toml in a toml fence + the live-rows tail. The shipped
// block-inline-<p>.md is the seed-state instance (the seed toml's rows); on refresh
// an inline install rebuilds the rows section from the consumer's actual
// rules.toml (decision-0053 point 3). The tail is posture-independent and carries
// the live-rows closing sentence, not the retired re-assembly one (point 4).
func TestPayloadInlineBlockIsRowsInlinedSandwich(t *testing.T) {
	files := payloadFiles()
	tail := files["block-inline-tail.md"]
	if !strings.HasSuffix(tail, trellisEnd) {
		t.Errorf("block-inline-tail.md must close the managed block (end marker): %q", tail)
	}
	if !strings.Contains(tail, "before deviating") {
		t.Errorf("block-inline-tail.md must carry the invariants trigger: %q", tail)
	}
	if !strings.Contains(tail, "Rule activation follows the rows in `.trellis/rules.toml` directly") {
		t.Errorf("block-inline-tail.md must close on the live-rows sentence (research-0012's header_arm_tail wording, decision-0053 point 4): %q", tail)
	}
	for _, p := range []string{"a", "b"} {
		head := files["block-inline-"+p+"-head.md"]
		if !strings.HasPrefix(head, trellisBegin) {
			t.Errorf("block-inline-%s-head.md must open the managed block (begin marker): %q", p, head)
		}
		if strings.Contains(head, "✗") || strings.Contains(tail, "✗") {
			t.Error("head/tail parts must carry no rule lines — the rules ride the readout between them")
		}
		rows := "\n## Active rows (`.trellis/rules.toml`)\n\n```toml\n" + files["rules-"+p+".toml"] + "```\n"
		if want := head + files["rules.md"] + rows + tail; files["block-inline-"+p+".md"] != want {
			t.Errorf("block-inline-%s.md must be exactly head + complete readout + rows section + tail (the tested sandwich)\n got:\n%s\n want:\n%s", p, files["block-inline-"+p+".md"], want)
		}
	}
}

// TestPayloadHeaderImportsSiblingRules: decision-0051 rule 1 — the header installed
// at .trellis/internal/trellis.md imports only its sibling rules.md (resolved
// relative to the importing file), never ../-traversal and never an expression
// import (expression.md is retired from the bundle — decision-0051 amendment). Its
// invariants trigger points at the internal/ home of the reference.
func TestPayloadHeaderImportsSiblingRules(t *testing.T) {
	files := payloadFiles()
	for _, name := range []string{"trellis-a.md", "trellis-b.md"} {
		content := files[name]
		if !strings.Contains(content, "@rules.md") {
			t.Errorf("%s must import its sibling @rules.md (decision-0051 rule 1): %q", name, content)
		}
		if strings.Contains(content, "@expression.md") {
			t.Errorf("%s must not import @expression.md — the expression import rides the managed block (decision-0051 rule 1): %q", name, content)
		}
		if strings.Contains(content, "@profile.md") {
			t.Errorf("%s still imports the retired profile.md readout (decision-0051 rule 5): %q", name, content)
		}
		if strings.Contains(content, "../") {
			t.Errorf("%s carries a ../ traversal — imports resolve relative to the importing file: %q", name, content)
		}
		if !strings.Contains(content, ".trellis/internal/invariants.md") {
			t.Errorf("%s must point its trigger at .trellis/internal/invariants.md: %q", name, content)
		}
	}
	for _, name := range []string{"block-inline-a.md", "block-inline-b.md"} {
		if !strings.Contains(files[name], ".trellis/internal/invariants.md") {
			t.Errorf("%s must point its trigger at .trellis/internal/invariants.md: %q", name, files[name])
		}
	}
}

// TestPayloadReadoutIsCompleteWithAuthorityHeader: decision-0053 point 2 — the
// readout ships complete (all 14 rules, every install) and opens with the authority
// header: research-0012's eval-tested AUTHORITY_HEADER wording, adapted in exactly
// one word ("inlined" → "loaded") so one shared readout is true on both channels
// (the inline block inlines the rows below the rules; the import block loads them
// below the rules via @.trellis/rules.toml). Rows-as-truth legibility survives
// assembly's retirement: each rule's first line still ends with its catalog slug in
// backticks (row ↔ rule ↔ entry matchability), which is also what research-0012's
// runner keys its subset transform on.
func TestPayloadReadoutIsCompleteWithAuthorityHeader(t *testing.T) {
	files := payloadFiles()
	r := files["rules.md"]
	if !strings.HasPrefix(r, "**Rule activation is governed by `.trellis/rules.toml`") {
		t.Fatalf("rules.md must open with the authority header (decision-0053 point 2): %q", r)
	}
	for _, want := range []string{
		"apply each rule below ONLY if its row says `active = true`",
		"A rule whose row is `active = false` does not apply in this project — do not follow it",
		"The two `floor-` rows apply regardless of their row value",
		"## The rules — do these",
		"Each rule below ends with its row's slug",
		"see the authority note above",
	} {
		if !strings.Contains(r, want) {
			t.Errorf("rules.md missing the tested live-rows wording %q (research-0012 / decision-0053)", want)
		}
	}
	if got := strings.Count(r, "\n    ✗ "); got != len(assessableSlugs) {
		t.Errorf("the readout ships complete — expected %d indented ✗ failure lines, got %d", len(assessableSlugs), got)
	}
	for _, slug := range assessableSlugs {
		if !strings.Contains(r, " `"+slug+"`\n") {
			t.Errorf("rules.md missing a rule line ending with its slug tag `%s`", slug)
		}
	}
}

// TestPayloadRulesReadoutIsOrderedConcatenation: the render contract after
// decision-0053 — rules.md is the readout header (authority note + heading +
// live-rows preamble) followed by every rule's fragment render in catalog order,
// with no assembly footer below the last rule (the "(Generated from your
// `rules.toml` …)" closing line retired with decision-0053 points 4+5). The inline
// blocks carry the identical complete readout.
func TestPayloadRulesReadoutIsOrderedConcatenation(t *testing.T) {
	files := payloadFiles()
	order := catalogSlugOrder()
	last := -1
	for _, slug := range order {
		i := strings.Index(files["rules.md"], " `"+slug+"`\n")
		if i < 0 {
			t.Errorf("rules.md missing the rule tagged `%s`", slug)
			continue
		}
		if i < last {
			t.Errorf("rules.md lists `%s` out of catalog order (decision-0051 rule 4's ordering survives in the render)", slug)
		}
		last = i
	}
	if !strings.HasSuffix(files["rules.md"], ruleFragment(order[len(order)-1])) {
		t.Errorf("rules.md must end on the last rule's fragment with nothing below it — the assembly footer retired (decision-0053 point 4): %q", files["rules.md"])
	}
	for _, name := range []string{"block-inline-a.md", "block-inline-b.md"} {
		if !strings.Contains(files[name], files["rules.md"]) {
			t.Errorf("%s must inline the complete readout verbatim", name)
		}
	}
}

// TestPayloadShipsNoRefreshTimeRowClaims: decision-0053 point 4 — no shipped text
// may claim refresh-time semantics for rows. The absence-era phrases (the assembly
// preamble, the "no effect until refresh" toml comment, the "re-assemble" tail
// sentence, and the "(Generated from your …)" footer/sentinel) must appear in no
// payload file.
func TestPayloadShipsNoRefreshTimeRowClaims(t *testing.T) {
	for name, content := range payloadFiles() {
		for _, banned := range []string{
			"assembled from the active rows",
			"no effect until",
			"re-assemble",
			"(Generated from your",
		} {
			if strings.Contains(content, banned) {
				t.Errorf("%s carries the absence-era claim %q — retired by decision-0053 point 4", name, banned)
			}
		}
	}
}

// TestPayloadRulesTomlSeeds: decision-0051 rules 2+3, live-rows comments per
// decision-0053 point 4 — the posture seeds are explicit rows, one per assessable
// catalog slug, all active (posture-as-seed, rows-as-truth: seeded_from is
// provenance only, strictness the one instance-level key). The top comment is the
// tested header_arm_toml wording ("Rows govern rule activation live …"), never the
// retired "no effect until refresh" claim, and the floor rows carry the tested
// live-rows floor comment ("floor — applies regardless of this row"), not the
// retired assembly-speak one.
func TestPayloadRulesTomlSeeds(t *testing.T) {
	files := payloadFiles()
	for _, tc := range []struct {
		name, seededFrom, strictness string
	}{
		{"rules-a.toml", `seeded_from = "conductor"`, `strictness  = "firm"`},
		{"rules-b.toml", `seeded_from = "author-adapt"`, `strictness  = "adaptive"`},
	} {
		content := files[tc.name]
		if content == "" {
			t.Fatalf("payload missing %s", tc.name)
		}
		if !strings.HasPrefix(content, "# Rows govern rule activation live (see the authority note in the project instructions).\n") {
			t.Errorf("%s must open with the tested live-rows comment (research-0012's header_arm_toml, decision-0053 point 4): %q", tc.name, content)
		}
		if !strings.Contains(content, tc.seededFrom) {
			t.Errorf("%s missing %q (provenance-only seed key, decision-0051 rule 2): %q", tc.name, tc.seededFrom, content)
		}
		if !strings.Contains(content, tc.strictness) {
			t.Errorf("%s missing %q (the one instance-level key, decision-0051 rule 7): %q", tc.name, tc.strictness, content)
		}
		if !strings.Contains(content, "[rules]") {
			t.Errorf("%s missing the [rules] table: %q", tc.name, content)
		}
		for _, slug := range assessableSlugs {
			rowRe := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(slug) + `\s+= \{ active = true \}`)
			if !rowRe.MatchString(content) {
				t.Errorf("%s missing an active row for %s (one row per assessable slug, all rows active): %q", tc.name, slug, content)
			}
		}
		for _, floor := range []string{"floor-transparency", "floor-intent-gate"} {
			lineRe := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(floor) + `.*# floor — applies regardless of this row`)
			if !lineRe.MatchString(content) {
				t.Errorf("%s: the %s row must carry the live-rows floor comment (decision-0053 point 4): %q", tc.name, floor, content)
			}
		}
		if strings.Contains(content, "floor-held") || strings.Contains(content, "assembly") {
			t.Errorf("%s still carries assembly-era floor wording — retired with decision-0053: %q", tc.name, content)
		}
	}
}

// TestSetupSkillCopiesCompleteReadout: decision-0053 Consequences — SKILL.md step
// 4's per-row selection cat became a plain copy of the shipped complete readout; no
// fragment-selection command remains anywhere in the skill; the skill states the
// live row semantics (edits take effect immediately) and names a floor row set
// false as overridden-by-floor, never silently honored. (Replaces
// TestSetupSkillAssemblyOrderMatchesCatalog — the catalog-order second home retired
// with the assembly command it pinned.)
func TestSetupSkillCopiesCompleteReadout(t *testing.T) {
	b, err := os.ReadFile("../plugins/trellis/skills/setup/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, `cp "${CLAUDE_PLUGIN_ROOT}/reference/rules.md" .trellis/internal/rules.md`) {
		t.Error("SKILL.md must install the readout as a plain copy of the shipped rules.md (decision-0053: assembly retires)")
	}
	if strings.Contains(s, `"$ref"/rules/`) {
		t.Error("SKILL.md still carries fragment-assembly commands — retired with decision-0053")
	}
	if !strings.Contains(s, "take effect immediately") {
		t.Error("SKILL.md must state the live row semantics: row edits take effect immediately (decision-0053 point 3)")
	}
	if !strings.Contains(s, "overridden-by-floor") {
		t.Error("SKILL.md must name a floor row set false as overridden-by-floor, loudly (decision-0053 point 3)")
	}
	// The SKILL's rows-section printf lines are the one non-payload byte source in the
	// inline rebuild; pin them to the generator's renderRowsSection so the two cannot
	// drift apart (decision-0028 sync-guard per source→derivative pair; conformance
	// finding on the decision-0053 build). Expected in step 7 (paste) and step 8(d)
	// (verify oracle) — exactly twice.
	openRaw := strings.TrimSuffix(renderRowsSection(""), "```\n")
	needle := "printf '" + strings.ReplaceAll(openRaw, "\n", `\n`) + "'"
	if n := strings.Count(s, needle); n != 2 {
		t.Errorf("SKILL.md rows-section printf must byte-match renderRowsSection in step 7 and step 8(d): want 2 occurrences of %q, found %d", needle, n)
	}
	for _, banned := range []string{"no effect until", "re-assemble", "next refresh — there is no per-session reader"} {
		if strings.Contains(s, banned) {
			t.Errorf("SKILL.md still carries the absence-era claim %q (decision-0053 point 4)", banned)
		}
	}
}

// TestPayloadCarriesInvariantsTrigger: the kodhama-0007 rider — the always-on
// templates' invariants pointer is a trigger ("read the entry before deviating"),
// not a description.
func TestPayloadCarriesInvariantsTrigger(t *testing.T) {
	files := payloadFiles()
	for _, name := range []string{"trellis-a.md", "trellis-b.md", "block-inline-a.md", "block-inline-b.md"} {
		if !strings.Contains(files[name], "before deviating") {
			t.Errorf("%s missing the invariants trigger (kodhama-0007 rider): %q", name, files[name])
		}
	}
}

// TestPayloadManifestVerifies: kodhama-0007 rule 3 — the manifest is shasum-format
// sha256 lines covering every payload file except itself (the rules.toml seeds and
// the fragments included: they are payload; only the installed consumer-root copies
// sit outside verification, decision-0051 rule 1), and every line verifies against
// the rendered content. The version stamp is content-derived, so the whole payload
// (stamp included) regenerates deterministically.
func TestPayloadManifestVerifies(t *testing.T) {
	files := payloadFiles()
	manifest := files["checksums"]

	lines := strings.Split(strings.TrimRight(manifest, "\n"), "\n")
	if len(lines) != len(files)-1 {
		t.Fatalf("manifest must cover every payload file except itself: %d lines for %d files", len(lines), len(files))
	}
	covered := map[string]bool{}
	for _, ln := range lines {
		// shasum -a 256 text-mode format: "<64 hex>  <name>" (two spaces).
		parts := strings.SplitN(ln, "  ", 2)
		if len(parts) != 2 || len(parts[0]) != 64 {
			t.Fatalf("manifest line is not shasum-compatible: %q", ln)
		}
		content, ok := files[parts[1]]
		if !ok {
			t.Errorf("manifest names a file the payload does not contain: %q", parts[1])
			continue
		}
		if got := fmt.Sprintf("%x", sha256.Sum256([]byte(content))); got != parts[0] {
			t.Errorf("manifest checksum for %s does not match the rendered content", parts[1])
		}
		covered[parts[1]] = true
	}
	for name := range files {
		if name != "checksums" && !covered[name] {
			t.Errorf("payload file %s missing from the manifest", name)
		}
	}

	if ok, _ := regexp.MatchString(`^payload@[0-9a-f]{12}\n$`, files["version"]); !ok {
		t.Errorf("version stamp must be payload@<12-hex content hash>, got %q", files["version"])
	}
	again := payloadFiles()
	if again["version"] != files["version"] || again["checksums"] != files["checksums"] {
		t.Error("payload render is not deterministic — version/checksums differ between runs")
	}
}

// TestVendoredPayloadIsCurrent is the regenerate-and-diff guard (#117: decision-0035's
// guard continuity — drift stays impossible, the mechanism changes): the payload
// vendored in plugins/trellis/reference/ must be byte-identical to what the generator
// renders from HEAD, and the directory must hold exactly the payload — reference/ is
// 100% generated, never mixed (kodhama-0007 rule 4's ownership rule). Since
// decision-0051 the payload nests the rules/ fragment directory, so the stray check
// walks recursively.
//
// Regenerate on failure:  (from cli/)  go run . payload --out ../plugins/trellis/reference
func TestVendoredPayloadIsCurrent(t *testing.T) {
	files := payloadFiles()
	for name, want := range files {
		got, err := os.ReadFile(filepath.Join(vendoredPayloadDir, name))
		if err != nil {
			t.Errorf("vendored payload missing %s — regenerate it (see this test's doc comment): %v", name, err)
			continue
		}
		if string(got) != want {
			t.Errorf("vendored %s is stale vs the generator — regenerate the payload (see this test's doc comment)", name)
		}
	}
	for _, rel := range walkFiles(t, vendoredPayloadDir) {
		if _, ok := files[rel]; !ok {
			t.Errorf("stray file in the payload dir: %s — reference/ is 100%% generated (kodhama-0007 rule 4)", rel)
		}
	}
}

// TestVendoredPayloadManifestVerifies is `shasum -c` in Go (kodhama-0007 AC1
// groundwork): the vendored manifest verifies the vendored files as they sit on
// disk, with no reference to the generator — the check anything can run.
func TestVendoredPayloadManifestVerifies(t *testing.T) {
	manifest, err := os.ReadFile(filepath.Join(vendoredPayloadDir, "checksums"))
	if err != nil {
		t.Fatalf("vendored payload has no checksums manifest: %v", err)
	}
	for _, ln := range strings.Split(strings.TrimRight(string(manifest), "\n"), "\n") {
		parts := strings.SplitN(ln, "  ", 2)
		if len(parts) != 2 {
			t.Fatalf("manifest line is not shasum-compatible: %q", ln)
		}
		b, err := os.ReadFile(filepath.Join(vendoredPayloadDir, parts[1]))
		if err != nil {
			t.Errorf("manifest names %s but it is not on disk: %v", parts[1], err)
			continue
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(b)); got != parts[0] {
			t.Errorf("vendored %s fails its manifest checksum", parts[1])
		}
	}
}

// TestPayloadCommandWritesPayload: the generator is runnable as release tooling
// (#117: "the existing Go render code, run in CI") — `trellis payload --out <dir>`
// writes the complete payload, nested fragment directory included; --out is required
// so it never scatters files.
func TestPayloadCommandWritesPayload(t *testing.T) {
	tmp := t.TempDir()
	var out strings.Builder
	if err := run(strings.NewReader(""), &out, []string{"payload", "--out", tmp}); err != nil {
		t.Fatalf("payload command: %v", err)
	}
	for name, want := range payloadFiles() {
		got, err := os.ReadFile(filepath.Join(tmp, name))
		if err != nil {
			t.Errorf("payload command did not write %s: %v", name, err)
			continue
		}
		if string(got) != want {
			t.Errorf("payload command wrote a different %s than the generator renders", name)
		}
	}

	if err := run(strings.NewReader(""), &strings.Builder{}, []string{"payload"}); err == nil {
		t.Error("payload without --out should be a loud error, not a silent default")
	}
}
