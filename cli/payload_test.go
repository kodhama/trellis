package main

// Tests for the release-render pipeline — kodhama/trellis#117 (kodhama-0007 slice 1,
// "one render, many copiers"), reshaped by decision-0051 (the overlay splits by
// authority: consumer-owned rules.toml at .trellis/, generated files under
// .trellis/internal/, the always-loaded readout assembled from per-rule fragments).
// Upstream anchors:
//   - kodhama-0007 rule 1 (render once, at release: the full enumerable variant space
//     is pre-rendered into the vendored payload) → TestPayloadFileSet,
//     TestVendoredPayloadIsCurrent.
//   - kodhama-0007 rule 3 (verification is data: a checksum manifest anything can
//     check with standard tools) → TestPayloadManifestVerifies,
//     TestVendoredPayloadManifestVerifies.
//   - decision-0051 rule 1 (authority split: the managed block imports
//     .trellis/internal/trellis.md and .trellis/expression.md; internal/trellis.md
//     imports only its sibling rules.md — @import paths resolve relative to the
//     importing file, so no ../ traversal) → TestPayloadBlockCarriesAuthoritySplitImports,
//     TestPayloadHeaderImportsSiblingRules.
//   - decision-0051 rules 2+3 (rules.toml is posture-as-seed, rows-as-truth; floors
//     are floor-held rows) → TestPayloadRulesTomlSeeds.
//   - decision-0051 rule 4 (the readout is assembled from manifest-covered fragments
//     in catalog order) → TestPayloadRuleFragmentsCoverCatalog,
//     TestPayloadRulesReadoutIsOrderedConcatenation,
//     TestSetupSkillAssemblyOrderMatchesCatalog.
//   - decision-0051 rule 5 (the name "profile" de-collided: the expression.md seed is
//     pure hand-owned prose with no profile: key; the readout closes on "Generated
//     from your rules.toml") → TestPayloadCarriesExpressionSeed,
//     TestPayloadRulesReadoutIsOrderedConcatenation.
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
// assessable invariants), alphabetical — the payload must carry one rule fragment
// per slug (decision-0051 rule 4).
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

// TestPayloadFileSet: the generator emits exactly the decision-0051 file set — the
// verbatim catalog, the per-rule fragments plus their header/footer, the assembled
// all-active readout, both posture variants of the header / inline block / rules.toml
// seed, the single hand-owned expression seed, the constant CLAUDE.md block, the
// version stamp, and the manifest.
func TestPayloadFileSet(t *testing.T) {
	want := []string{
		"block-claude.md",
		"block-inline-a-head.md",
		"block-inline-a.md",
		"block-inline-b-head.md",
		"block-inline-b.md",
		"block-inline-tail.md",
		"checksums",
		"expression.md",
		"invariants.md",
		"rules-a.toml",
		"rules-b.toml",
		"rules.md",
	}
	for _, slug := range assessableSlugs {
		want = append(want, "rules/"+slug+".md")
	}
	want = append(want,
		"rules/_footer.md",
		"rules/_header.md",
		"trellis-a.md",
		"trellis-b.md",
		"version",
	)
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
// path's output for its posture — the catalog verbatim, the posture-a/b strictness
// split in the always-on templates, the constant CLAUDE.md block (#117's verified
// variant-space evidence; the readout itself is posture-independent, decision-0051
// open question "fragment granularity vs. posture").
func TestPayloadVariantsAreTheRenderedPostures(t *testing.T) {
	files := payloadFiles()

	if files["invariants.md"] != invariantsRef {
		t.Error("payload invariants.md must be the bundled catalog, verbatim")
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

// TestPayloadBlockCarriesAuthoritySplitImports: decision-0051 rule 1 — the managed
// CLAUDE.md block imports the trellis-authoritative header
// (.trellis/internal/trellis.md) and the consumer's hand-owned expression
// (.trellis/expression.md), rules before expression (kodhama-0007 rule 4 via #119's
// ordering). Both imports live in the block because @import paths resolve relative
// to the importing file: internal/trellis.md could not reach ../expression.md
// without traversal. The inline blocks stay import-free — they exist precisely for
// files without @import support.
func TestPayloadBlockCarriesAuthoritySplitImports(t *testing.T) {
	files := payloadFiles()
	block := files["block-claude.md"]
	i := strings.Index(block, "@.trellis/internal/trellis.md")
	j := strings.Index(block, "@.trellis/expression.md")
	if i < 0 || j < 0 {
		t.Fatalf("block-claude.md must import both @.trellis/internal/trellis.md and @.trellis/expression.md (decision-0051 rule 1): %q", block)
	}
	if j < i {
		t.Errorf("block-claude.md must import the rules header before the project's expression: %q", block)
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

// TestPayloadInlineBlockIsHeadReadoutTail: conformance-review remediation of
// decision-0051 rule 4's letter — "the managed block's @import (or the inline
// block) carries the assembled readout … so an edited row takes effect at the next
// refresh." The inline channel honors rows the same mechanical way the import
// channel does: on refresh the block is rebuilt as the manifest-covered head part +
// the assembled .trellis/internal/rules.md + the manifest-covered tail part,
// concatenated with no authored bytes (the same mechanical class as the rules.md
// assembly). The shipped block-inline-<p>.md is exactly the all-active instance of
// that sandwich; the tail is posture-independent (only the head's strictness line
// differs), so one tail file ships, not two.
func TestPayloadInlineBlockIsHeadReadoutTail(t *testing.T) {
	files := payloadFiles()
	tail := files["block-inline-tail.md"]
	if !strings.HasSuffix(tail, trellisEnd) {
		t.Errorf("block-inline-tail.md must close the managed block (end marker): %q", tail)
	}
	if !strings.Contains(tail, "before deviating") {
		t.Errorf("block-inline-tail.md must carry the invariants trigger: %q", tail)
	}
	for _, p := range []string{"a", "b"} {
		head := files["block-inline-"+p+"-head.md"]
		if !strings.HasPrefix(head, trellisBegin) {
			t.Errorf("block-inline-%s-head.md must open the managed block (begin marker): %q", p, head)
		}
		if strings.Contains(head, "✗") || strings.Contains(tail, "✗") {
			t.Error("head/tail parts must carry no rule lines — the rules ride the assembled readout between them")
		}
		if want := head + files["rules.md"] + tail; files["block-inline-"+p+".md"] != want {
			t.Errorf("block-inline-%s.md must be exactly head + all-active readout + tail\n got:\n%s\n want:\n%s", p, files["block-inline-"+p+".md"], want)
		}
	}
}

// TestPayloadHeaderImportsSiblingRules: decision-0051 rule 1 — the header installed
// at .trellis/internal/trellis.md imports only its sibling rules.md (resolved
// relative to the importing file), never ../-traversal and never the expression
// (that import rides the managed block). Its invariants trigger points at the new
// internal/ home of the reference.
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

// TestPayloadCarriesExpressionSeed: decision-0051 rule 5 — expression.md becomes pure
// hand-owned prose: the seed carries no YAML frontmatter and no machine-read
// profile: key (the config moved to rules.toml), declares its ownership rule, and
// cross-links the reserved meaning of "profile" (the expression-profile artifact,
// decision-0016). One seed, not per-posture: with the frontmatter gone the seed has
// no posture content left.
func TestPayloadCarriesExpressionSeed(t *testing.T) {
	files := payloadFiles()
	content := files["expression.md"]
	if strings.HasPrefix(content, "---") {
		t.Errorf("expression.md seed must not open with YAML frontmatter (the profile: key retired, decision-0051 rule 5): %q", content)
	}
	if strings.Contains(content, "profile: a") || strings.Contains(content, "profile: b") {
		t.Errorf("expression.md seed must not carry a machine-read profile: key: %q", content)
	}
	for _, want := range []string{"hand-owned", "rules.toml", "expression-profile"} {
		if !strings.Contains(content, want) {
			t.Errorf("expression.md seed missing %q — it must declare its ownership rule, point at the config's new home, and cross-link the reserved word (decision-0051 rule 5)", want)
		}
	}
	for _, stray := range []string{"<p>", "<project>"} {
		if strings.Contains(content, stray) {
			t.Errorf("expression.md seed carries an unfilled placeholder %q — seeds are copied verbatim, never filled", stray)
		}
	}
}

// TestPayloadRuleFragmentsCoverCatalog: decision-0051 rule 4 — one pre-rendered
// fragment per assessable catalog slug, each the rule's imperative directive plus
// the ✗ failure it prevents, and the two non-rule fragments (_header/_footer) that
// make the assembled readout pure concatenation with no byte authored at install
// time.
func TestPayloadRuleFragmentsCoverCatalog(t *testing.T) {
	files := payloadFiles()
	for _, slug := range assessableSlugs {
		name := "rules/" + slug + ".md"
		content := files[name]
		if content == "" {
			t.Errorf("payload missing fragment %s", name)
			continue
		}
		if !strings.HasPrefix(content, "- ") {
			t.Errorf("%s must open with its imperative directive as a list item: %q", name, content)
		}
		if !strings.Contains(content, "✗") {
			t.Errorf("%s must carry the ✗ failure line under its rule: %q", name, content)
		}
		if !strings.HasSuffix(content, "\n") {
			t.Errorf("%s must end with a newline so concatenation is seamless: %q", name, content)
		}
	}
	if !strings.Contains(files["rules/_header.md"], "## The rules — do these") {
		t.Errorf("rules/_header.md must carry the readout heading: %q", files["rules/_header.md"])
	}
	if !strings.Contains(files["rules/_footer.md"], "(Generated from your `rules.toml`") {
		t.Errorf("rules/_footer.md must carry the closing \"Generated from your rules.toml\" line (decision-0051 rule 5): %q", files["rules/_footer.md"])
	}
}

// TestPayloadRulesReadoutIsOrderedConcatenation: decision-0051 rule 4's verify
// contract — the assembled all-active readout the payload ships (rules.md, the
// common case's copy source) is byte-for-byte the ordered concatenation of
// _header + every rule fragment in catalog order + _footer, and the inline blocks
// carry the same all-active body.
func TestPayloadRulesReadoutIsOrderedConcatenation(t *testing.T) {
	files := payloadFiles()
	order := catalogSlugOrder()
	var b strings.Builder
	b.WriteString(files["rules/_header.md"])
	for _, slug := range order {
		b.WriteString(files["rules/"+slug+".md"])
	}
	b.WriteString(files["rules/_footer.md"])
	if files["rules.md"] != b.String() {
		t.Errorf("rules.md is not the ordered concatenation of its fragments\n got:\n%s\n want:\n%s", files["rules.md"], b.String())
	}
	body := strings.TrimSuffix(strings.TrimPrefix(files["rules.md"], files["rules/_header.md"]), files["rules/_footer.md"])
	for _, name := range []string{"block-inline-a.md", "block-inline-b.md"} {
		if !strings.Contains(files[name], body) {
			t.Errorf("%s must inline the same all-active rules body the fragments assemble to", name)
		}
	}
}

// TestPayloadRulesTomlSeeds: decision-0051 rules 2+3 — the posture seeds are
// explicit rows, one per assessable catalog slug, all active (posture-as-seed,
// rows-as-truth: seeded_from is provenance only, strictness the one instance-level
// key), and the floor rows are marked floor-held (a consumer cannot turn them off;
// assembly includes them regardless).
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
			lineRe := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(floor) + `.*floor-held`)
			if !lineRe.MatchString(content) {
				t.Errorf("%s: the %s row must be marked floor-held (decision-0051 rule 3): %q", tc.name, floor, content)
			}
		}
	}
}

// TestSetupSkillAssemblyOrderMatchesCatalog: decision-0028's sync-guard for the one
// place catalog order has a second home — the setup skill's step-4 assembly command
// lists the fragment files in the order setup concatenates them, and that sequence
// must be exactly _header, then the catalog's document order, then _footer
// (decision-0051 rule 4: "in catalog order"; the closing "(Generated from your
// rules.toml …)" line rides _footer — decision-0051 rule 5's required closing line,
// which the conformance review caught the command dropping). The `"$ref"/rules/…`
// form is unique to that command, so its occurrences in reading order pin the
// command block completely: presence of both non-rule fragments included, not just
// the slug order.
func TestSetupSkillAssemblyOrderMatchesCatalog(t *testing.T) {
	b, err := os.ReadFile("../plugins/trellis/skills/setup/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	fragRe := regexp.MustCompile(`"\$ref"/rules/([A-Za-z_-]+)\.md`)
	var got []string
	for _, m := range fragRe.FindAllStringSubmatch(string(b), -1) {
		got = append(got, m[1])
	}
	want := append(append([]string{"_header"}, catalogSlugOrder()...), "_footer")
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("SKILL.md's step-4 assembly command must be exactly _header + catalog order + _footer\n got:  %v\n want: %v", got, want)
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
