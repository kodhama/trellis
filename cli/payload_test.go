package main

// Tests for the release-render pipeline — kodhama/trellis#117 (kodhama-0007 slice 1,
// "one render, many copiers"), reshaped by decision-0051 (the overlay splits by
// authority: consumer-owned rules.toml at .trellis/, generated files under
// .trellis/internal/), again by decision-0053 (live rows: the readout ships
// complete with an authority header; rules.toml rows govern at read time; fragment
// assembly retires, and the fragments leave the shipped payload — no consumer
// remains), and again by TRL-97 (only a row set false has effect: the posture
// variants and the rules.toml seeds retire, one header and one inline block ship).
// Upstream anchors:
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
//     is head + readout + tail, its rows section retired with TRL-97) →
//     TestPayloadBlockCarriesBothImports, TestPayloadInlineBlockIsHeadReadoutTail.
//   - decision-0053 point 2 (the readout ships complete and carries the authority
//     header, research-0012's tested wording reworded by TRL-97 to the opt-out-only
//     activation rule) → TestPayloadReadoutIsCompleteWithAuthorityHeader,
//     TestPayloadRulesReadoutIsOrderedConcatenation.
//   - decision-0053 point 4 (no shipped text claims refresh-time semantics for rows;
//     the absence-era preamble/footer/tail/toml comments retired) →
//     TestPayloadShipsNoRefreshTimeRowClaims.
//   - decision-0053 Consequences (SKILL.md step 4's selection cat → a plain copy of
//     rules.md) → TestSetupSkillCopiesCompleteReadout.
//   - TRL-97 (every project receives the "By default" posture sentence; a file
//     written from now on holds only what has effect, so nothing is seeded; floors
//     are held by the authority header's floor sentence) →
//     TestPayloadFilesAreTheirRenders, TestPayloadShipsNoRulesTomlSeed,
//     TestPayloadReadoutIsCompleteWithAuthorityHeader.
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

// assessableSlugs is the pinned catalog slug set (signature-catalog-v1: the 16
// assessable invariants), alphabetical — the complete readout must carry one
// slug-tagged rule per slug (decision-0053 point 1).
//
// This literal is the ONE pin for the row set. Every other count in the suite
// derives from len(assessableSlugs) rather than repeating the number, because
// decision-0074's self-check named the repeated literal as the root cause of the
// misses on the previous row addition — "a sweep that matched only some of the
// shapes a count takes" (decision-0078).
var assessableSlugs = []string{
	"floor-intent-gate",
	"floor-transparency",
	"inv-auditable-archive",
	"inv-bounded-context",
	"inv-clarify-before-commit",
	"inv-deliberate-succession",
	"inv-directional-flow",
	"inv-gate-at-handover",
	"inv-graph-maintenance",
	"inv-handover-points",
	"inv-independent-judgment",
	"inv-intent-locus",
	"inv-minimal-first",
	"inv-no-orphan-followups",
	"inv-ratifiable-artifacts",
	"inv-self-improvement",
}

// defaultPostureSentence is the one posture sentence every project receives
// (TRL-97, R14), whatever `strictness` its rules.toml still carries.
const defaultPostureSentence = "**How strictly to follow them:** **By default**"

// retiredAuthoritySentence is the activation rule research-0012 tested and TRL-97
// retired: under it a rule with no row did not apply. No payload file may carry it.
const retiredAuthoritySentence = "ONLY if its row says `active = true`"

// TestPayloadFileSet: the generator emits exactly this file set — the catalog's
// entries section, the complete readout, the one header, the one inline block (whole,
// head and tail), the constant CLAUDE.md block, the Codex bootstrap, the version
// stamp, and the manifest. The posture variants and the rules.toml seeds retired with
// TRL-97: every project receives the "By default" sentence, and a file written from
// now on holds only what has effect, so there is nothing to seed. No expression seed
// (retired, decision-0051 amendment) and no rules/ fragment files (they left the
// shipped payload with decision-0053 point 1). The exact-match list is the guard that
// all of them stay gone.
func TestPayloadFileSet(t *testing.T) {
	want := []string{
		"block-claude.md",
		"block-codex.md",
		"block-inline-head.md",
		"block-inline-tail.md",
		"block-inline.md",
		"checksums",
		"invariants.md",
		"rules.md",
		"trellis.md",
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

// TestPayloadFilesAreTheirRenders: each payload file is the single render path's
// output — the catalog with only its entries section kept (decision-0055 point 1,
// widening decision-0054 point 1: invariantsRef itself is untouched; only the
// payload-write site extracts it), the constant CLAUDE.md block, and the one posture
// sentence in every always-on template. Since TRL-97 no posture split remains: the
// header and the inline block both carry "By default", and no payload file carries
// the retired "Firmly" or "As guidance" lines.
func TestPayloadFilesAreTheirRenders(t *testing.T) {
	files := payloadFiles()

	if files["invariants.md"] != extractEntriesSection(invariantsRef) {
		t.Error("payload invariants.md must be the bundled catalog's entries section only — preamble and tail excluded (decision-0055)")
	}
	if files["invariants.md"] == invariantsRef {
		t.Error("payload invariants.md must not be the bundled catalog verbatim — preamble/frontmatter/tail should be excluded (decision-0055)")
	}
	if strings.Contains(files["invariants.md"], "## Acceptance criteria") {
		t.Error("payload invariants.md must not carry the catalog's own Acceptance criteria section (decision-0055 point 1, exclusive boundary)")
	}
	if strings.Contains(files["invariants.md"], "Ratified via merge") {
		t.Error("payload invariants.md must not carry the catalog's ratification preamble (decision-0055 point 1)")
	}
	if files["block-claude.md"] != renderClaudeBlock() {
		t.Error("payload block-claude.md must be the constant CLAUDE.md block")
	}
	for _, name := range []string{"trellis.md", "block-inline-head.md", "block-inline.md"} {
		if n := strings.Count(files[name], defaultPostureSentence); n != 1 {
			t.Errorf("%s must carry the posture sentence every project receives, %q, exactly once (TRL-97); got %d", name, defaultPostureSentence, n)
		}
	}
	for name, content := range files {
		for _, retired := range []string{"**Firmly**", "**As guidance**"} {
			if strings.Contains(content, retired) {
				t.Errorf("%s carries the retired posture line %q — every project receives \"By default\" since TRL-97", name, retired)
			}
		}
	}
	if !strings.Contains(files["block-inline.md"], trellisBegin) || !strings.Contains(files["block-inline.md"], trellisEnd) {
		t.Error("block-inline.md must be a complete managed block (begin/end markers)")
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
// stays retired (decision-0051 amendment). The inline block parts stay import-free —
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
	for _, name := range []string{"block-inline.md", "block-inline-head.md", "block-inline-tail.md"} {
		if strings.Contains(files[name], "@.trellis/") || strings.Contains(files[name], "@expression.md") || strings.Contains(files[name], "@rules.md") {
			t.Errorf("%s is (part of) the no-@import variant and must not carry an import line", name)
		}
	}
}

// TestPayloadInlineBlockIsHeadReadoutTail: decision-0053 point 2 (inline channel) as
// narrowed by TRL-97 — the block is head + the complete readout + the live-rows tail,
// and nothing else. Its embedded "## Active rows" section retired: it carried a
// posture seed's rows, nothing needs seeding now that a rule with no row applies, and
// the managed inline shape is retired for new installs. The tail still says rule
// activation follows the rows in `.trellis/rules.toml`, and the tail is still the
// live-rows one, not the retired re-assembly sentence (decision-0053 point 4).
func TestPayloadInlineBlockIsHeadReadoutTail(t *testing.T) {
	files := payloadFiles()
	head, tail, block := files["block-inline-head.md"], files["block-inline-tail.md"], files["block-inline.md"]
	if !strings.HasSuffix(tail, trellisEnd) {
		t.Errorf("block-inline-tail.md must close the managed block (end marker): %q", tail)
	}
	if !strings.Contains(tail, "before deviating") {
		t.Errorf("block-inline-tail.md must carry the invariants trigger: %q", tail)
	}
	if !strings.Contains(tail, "Rule activation follows the rows in `.trellis/rules.toml`") {
		t.Errorf("block-inline-tail.md must close on the live-rows sentence (research-0012's header_arm_tail wording, decision-0053 point 4): %q", tail)
	}
	if !strings.HasPrefix(head, trellisBegin) {
		t.Errorf("block-inline-head.md must open the managed block (begin marker): %q", head)
	}
	if strings.Contains(head, "✗") || strings.Contains(tail, "✗") {
		t.Error("head/tail parts must carry no rule lines — the rules ride the readout between them")
	}
	if want := head + files["rules.md"] + tail; block != want {
		t.Errorf("block-inline.md must be exactly head + complete readout + tail\n got:\n%s\n want:\n%s", block, want)
	}
	// The retired rows section, named by each of its parts, so a seed cannot return
	// under a new heading or without its fence.
	for _, retired := range []string{"## Active rows", "```toml", "strictness", "seeded_from"} {
		if strings.Contains(block, retired) {
			t.Errorf("block-inline.md carries %q from the retired rows section (TRL-97)", retired)
		}
	}
	if m := seedRowLineRe.FindString(block); m != "" {
		t.Errorf("block-inline.md carries a rules.toml row line %q — the inline block embeds no rows since TRL-97", m)
	}
}

// TestPayloadHeaderImportsSiblingRules: decision-0051 rule 1 — the header installed
// at .trellis/internal/trellis.md imports only its sibling rules.md (resolved
// relative to the importing file), never ../-traversal and never an expression
// import (expression.md is retired from the bundle — decision-0051 amendment). Its
// invariants trigger points at the internal/ home of the reference.
func TestPayloadHeaderImportsSiblingRules(t *testing.T) {
	files := payloadFiles()
	content := files["trellis.md"]
	if !strings.Contains(content, "@rules.md") {
		t.Errorf("trellis.md must import its sibling @rules.md (decision-0051 rule 1): %q", content)
	}
	if strings.Contains(content, "@expression.md") {
		t.Errorf("trellis.md must not import @expression.md — the expression import rides the managed block (decision-0051 rule 1): %q", content)
	}
	if strings.Contains(content, "@profile.md") {
		t.Errorf("trellis.md still imports the retired profile.md readout (decision-0051 rule 5): %q", content)
	}
	if strings.Contains(content, "../") {
		t.Errorf("trellis.md carries a ../ traversal — imports resolve relative to the importing file: %q", content)
	}
	if !strings.Contains(content, ".trellis/internal/invariants.md") {
		t.Errorf("trellis.md must point its trigger at .trellis/internal/invariants.md: %q", content)
	}
	// INVERTED by decision-0073 (Consequence 2). This check used to require the
	// inline block to carry the .trellis/internal/invariants.md pointer —
	// which is exactly the defect 0073 names: clean S4 has no overlay, so the
	// shipped inline payload pointed readers at a file that need not exist and
	// the documented recipe could only produce the S2-plus-S4 conflict. The
	// inline trigger is forked (inlineInvariantsTrigger); the header above
	// keeps the internal pointer, which is correct for S2 and repointed by the
	// hook for path B. TestPayloadInlineTailNeedsNoOverlay carries the
	// positive half of the new property.
	if strings.Contains(files["block-inline.md"], ".trellis/internal/invariants.md") {
		t.Errorf("block-inline.md points its trigger at .trellis/internal/, which clean S4 does not have (decision-0073): %q", files["block-inline.md"])
	}
}

// TestPayloadReadoutIsCompleteWithAuthorityHeader: decision-0053 point 2 — the
// readout ships complete (every assessable rule, every install) and opens with the
// authority header. TRL-97 rewords research-0012's tested header to the opt-out-only
// activation rule (KTD2): a rule applies unless a row for it says `active = false`, a rule
// with no row applies, the floor rules always apply while the project is governed,
// and nothing else in the file changes which rules apply — the last clause covers an
// inert `strictness` without naming it. The tested "do not follow it" sentence, the
// one that carried the switched-off rule's effect, stays verbatim. The rule is stated
// once, here: the readout's own preamble points back at it rather than restating it.
// Rows-as-truth legibility survives: each rule's first line still ends with its
// catalog slug in backticks (row ↔ rule ↔ entry matchability), which is also what
// research-0012's runner keys its subset transform on.
func TestPayloadReadoutIsCompleteWithAuthorityHeader(t *testing.T) {
	files := payloadFiles()
	r := files["rules.md"]
	if !strings.HasPrefix(r, "**Rule activation is governed by `.trellis/rules.toml`") {
		t.Fatalf("rules.md must open with the authority header (decision-0053 point 2): %q", r)
	}
	for _, want := range []string{
		"apply each rule below unless a row for it says `active = false`; a rule with no row applies",
		"A rule whose row is `active = false` does not apply in this project — do not follow it",
		"The two `floor-` rules always apply while the project is governed, whatever their row says",
		"Nothing else in `.trellis/rules.toml` changes which rules apply",
		"## The rules — do these",
		"Each rule below ends with its row's slug",
		"Whether a rule applies follows the authority note above",
	} {
		if !strings.Contains(r, want) {
			t.Errorf("rules.md missing the activation wording %q (TRL-97, KTD2)", want)
		}
	}
	// The retired sentences: the old rule itself, its floor clause, and the readout
	// preamble's unconditional claim that the rows sit below the rules, which stopped
	// holding when the inline block dropped its rows section.
	for _, retired := range []string{
		retiredAuthoritySentence,
		"The two `floor-` rows apply regardless of their row value",
		"the rows are loaded below the rules",
	} {
		if strings.Contains(r, retired) {
			t.Errorf("rules.md still carries the retired activation wording %q (TRL-97)", retired)
		}
	}
	if n := strings.Count(r, "`active = false`"); n != 2 {
		t.Errorf("rules.md must state the activation rule once, in the authority header — expected its two `active = false` mentions and no restatement, got %d", n)
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
// preamble) followed by every rule's fragment render in catalog order,
// followed by the stable spec-0007 completion sentinel and no assembly footer
// (the "(Generated from your
// `rules.toml` …)" closing line retired with decision-0053 points 4+5). The inline
// block carries the identical complete readout.
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
	if !strings.HasSuffix(files["rules.md"], ruleFragment(order[len(order)-1])+trellisRulesLoadedSentinel+"\n") {
		t.Errorf("rules.md must end on the last rule's fragment plus the exact spec-0007 completion sentinel: %q", files["rules.md"])
	}
	if !strings.Contains(files["block-inline.md"], files["rules.md"]) {
		t.Error("block-inline.md must inline the complete readout verbatim")
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

// seedKeyLineRe and seedRowLineRe match the two line shapes a rules.toml seed was
// made of: a `strictness` or `seeded_from` assignment, and a `[rules]` row.
var (
	seedKeyLineRe = regexp.MustCompile(`(?m)^[ \t]*(strictness|seeded_from)[ \t]*=`)
	seedRowLineRe = regexp.MustCompile(`(?m)^[ \t]*(inv|floor)-[a-z-]+[ \t]*=[ \t]*\{`)
)

// TestPayloadShipsNoRulesTomlSeed: TRL-97 retires the rules.toml seeds with the
// posture presets. A file written from now on holds only what has effect, so no
// payload file may carry a seed's lines under any name. TestPayloadFileSet catches a
// seed that returns as its own file; this catches one folded into another file, the
// way the inline block embedded one until this change. The retired activation
// sentence is pinned out of every file here too: rules.md states the rule once, and
// no header or block may restate the old one.
func TestPayloadShipsNoRulesTomlSeed(t *testing.T) {
	for name, content := range payloadFiles() {
		if m := seedKeyLineRe.FindString(content); m != "" {
			t.Errorf("%s carries a rules.toml seed key line %q — the seeds retired with TRL-97", name, m)
		}
		if m := seedRowLineRe.FindString(content); m != "" {
			t.Errorf("%s carries a rules.toml row line %q — the seeds retired with TRL-97", name, m)
		}
		if strings.Contains(content, retiredAuthoritySentence) {
			t.Errorf("%s carries the retired activation sentence %q — a rule with no row applies since TRL-97", name, retiredAuthoritySentence)
		}
	}
}

// TestSetupSkillWritesOnlyTheConfig was DELETED by decision-0072, which retired
// /trellis:setup. It guarded the skill's one-write contract; there is no skill
// left to guard. It is not replaced by a narrower assertion on the same file
// because the file is gone — the guard that matters now is
// TestOnlyTheRemoveSkillShips below, which fails if a setup skill reappears
// without a decision reversing 0072.

// TestOnlyTheRemoveSkillShips: decision-0072 retired /trellis:setup, and the
// plugin ships exactly one skill. A skill directory is user-visible surface —
// it becomes a slash command the moment it exists — so its return must be a
// deliberate act, not a file someone restores from a stale branch.
func TestOnlyTheRemoveSkillShips(t *testing.T) {
	entries, err := os.ReadDir("../plugins/trellis/skills")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, e := range entries {
		if e.IsDir() {
			got = append(got, e.Name())
		}
	}
	if len(got) != 1 || got[0] != "remove" {
		t.Errorf("the plugin must ship exactly the remove skill (decision-0072 retired setup); got %v", got)
	}
}

// TestPayloadCarriesInvariantsTrigger: the kodhama-0007 rider — the always-on
// templates' invariants pointer is a trigger ("read the entry before deviating"),
// not a description.
func TestPayloadCarriesInvariantsTrigger(t *testing.T) {
	files := payloadFiles()
	for _, name := range []string{"trellis.md", "block-inline.md"} {
		if !strings.Contains(files[name], "before deviating") {
			t.Errorf("%s missing the invariants trigger (kodhama-0007 rider): %q", name, files[name])
		}
	}
}

// TestPayloadManifestVerifies: kodhama-0007 rule 3 — the manifest is shasum-format
// sha256 lines covering every payload file except itself (only the installed
// consumer-root .trellis/rules.toml sits outside verification, decision-0051 rule
// 1), and every line verifies against the rendered content. The version stamp is
// content-derived, so the whole payload (stamp included) regenerates
// deterministically.
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
// The generator never deletes, so a file the payload stops rendering is removed by hand.
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

// TestPayloadInlineTailNeedsNoOverlay guards decision-0073 (Consequence 2:
// clean S4 must be producible). The inline block is the delivery shape for a
// project with NO vendored overlay — the block plus .trellis/rules.toml is the
// whole install — yet its shipped tail pointed readers at
// `.trellis/internal/invariants.md`, a file that need not exist there, so the
// documented inline recipe could only produce a conflicting S2-plus-S4 layout.
// The tail's trigger is forked from the header's (renderHeader keeps the
// shared invariantsTrigger const: in S2 the vendored copy is real, and the
// hook repoints it for path B), and points at the shipped reference instead.
func TestPayloadInlineTailNeedsNoOverlay(t *testing.T) {
	files := payloadFiles()
	for _, name := range []string{"block-inline-tail.md", "block-inline.md"} {
		if strings.Contains(files[name], ".trellis/internal/") {
			t.Errorf("%s points its reader at .trellis/internal/, which clean S4 does not have (decision-0073): %q", name, files[name])
		}
	}
	tail := files["block-inline-tail.md"]
	// The affordance survives the repoint: an ambiguity/deviating trigger with
	// a target that exists wherever the block came from.
	if !strings.Contains(tail, "before deviating") {
		t.Errorf("block-inline-tail.md lost the invariants trigger — the repoint must move the target, not delete the affordance: %q", tail)
	}
	if !strings.Contains(tail, "invariants.md") {
		t.Errorf("block-inline-tail.md must still name an invariants reference the reader can actually open: %q", tail)
	}
}
