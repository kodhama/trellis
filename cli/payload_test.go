package main

// Tests for the release-render pipeline — kodhama/trellis#117 (kodhama-0007 slice 1,
// "one render, many copiers"). Upstream anchors:
//   - kodhama-0007 rule 1 (render once, at release: the full enumerable variant space
//     is pre-rendered into the vendored payload) → TestPayloadFileSet,
//     TestVendoredPayloadIsCurrent.
//   - kodhama-0007 rule 3 (verification is data: a checksum manifest anything can
//     check with standard tools) → TestPayloadManifestVerifies,
//     TestVendoredPayloadManifestVerifies.
//   - kodhama-0007 rider (the invariants pointer upgrades from description to
//     trigger in the always-on templates) → TestPayloadCarriesInvariantsTrigger.
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

// TestPayloadFileSet: the generator emits exactly the file set #117 enumerates —
// the verbatim catalog, both posture variants of the profile / header / inline
// block, the constant CLAUDE.md block, the version stamp, and the manifest —
// plus, since #119's conformance follow-up, both posture variants of the
// expression.md seed skeleton (kodhama-0007 rule 4).
func TestPayloadFileSet(t *testing.T) {
	want := []string{
		"block-claude.md",
		"block-inline-a.md",
		"block-inline-b.md",
		"checksums",
		"expression-a.md",
		"expression-b.md",
		"invariants.md",
		"profile-a.md",
		"profile-b.md",
		"trellis-a.md",
		"trellis-b.md",
		"version",
	}
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
// variant-space evidence).
func TestPayloadVariantsAreTheRenderedPostures(t *testing.T) {
	files := payloadFiles()

	if files["invariants.md"] != invariantsRef {
		t.Error("payload invariants.md must be the bundled catalog, verbatim")
	}
	if files["block-claude.md"] != renderClaudeBlock() {
		t.Error("payload block-claude.md must be the constant CLAUDE.md block")
	}
	// Postures differ in exactly one dimension: the strictness line (issue #117:
	// "only trellis.md's strictness line and the AGENTS.md inline block differ").
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
	if !strings.Contains(files["block-claude.md"], "@.trellis/trellis.md") {
		t.Error("block-claude.md must import .trellis/trellis.md")
	}
	for _, name := range []string{"profile-a.md", "profile-b.md"} {
		if !strings.Contains(files[name], "settled ground") {
			t.Errorf("%s missing the active rule directives", name)
		}
	}
}

// TestPayloadHeaderCarriesExpressionImport: kodhama-0007 rule 4 via kodhama/trellis#119
// — `.trellis/expression.md` is the project's always-on hand-owned declaration, so the
// header the payload ships must import it alongside `@profile.md`, rules first, then
// the project's expression (matching how projects actually used it: the first migrant's
// expression sat below the rules). The inline blocks stay import-free — they exist
// precisely for files without @import support.
func TestPayloadHeaderCarriesExpressionImport(t *testing.T) {
	files := payloadFiles()
	for _, name := range []string{"trellis-a.md", "trellis-b.md"} {
		content := files[name]
		i := strings.Index(content, "@profile.md")
		j := strings.Index(content, "@expression.md")
		if i < 0 || j < 0 {
			t.Errorf("%s must import both @profile.md and @expression.md (kodhama-0007 rule 4, #119): %q", name, content)
			continue
		}
		if j < i {
			t.Errorf("%s must import the rules (@profile.md) before the project's expression (@expression.md): %q", name, content)
		}
	}
	for _, name := range []string{"block-inline-a.md", "block-inline-b.md"} {
		if strings.Contains(files[name], "@expression.md") {
			t.Errorf("%s is the no-@import variant and must not carry an import line", name)
		}
	}
}

// TestPayloadCarriesExpressionSkeletons: kodhama-0007 rule 4 via #119 (conformance-gate
// follow-up) — the expression.md seed skeleton is payload content like everything else
// ("one render, many copiers", applied to the skeleton itself): per-posture files with
// the frontmatter pre-filled, so every writer copies verbatim with nothing left to fill
// and the skeleton has no second home in writer prose.
func TestPayloadCarriesExpressionSkeletons(t *testing.T) {
	files := payloadFiles()
	for _, p := range allProfiles {
		name := "expression-" + p.Key + ".md"
		content := files[name]
		if !strings.HasPrefix(content, "---\nprofile: "+p.Key+"\n---\n") {
			t.Errorf("%s must open with pre-filled machine-read frontmatter (profile: %s), got: %q", name, p.Key, content)
		}
		for _, want := range []string{"hand-owned", "kodhama-0007 rule 4"} {
			if !strings.Contains(content, want) {
				t.Errorf("%s missing %q — the skeleton must declare its ownership rule", name, want)
			}
		}
		for _, stray := range []string{"<p>", "<project>"} {
			if strings.Contains(content, stray) {
				t.Errorf("%s carries an unfilled placeholder %q — skeletons are copied verbatim, never filled", name, stray)
			}
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
// sha256 lines covering every payload file except itself, and every line verifies
// against the rendered content. The version stamp is content-derived, so the whole
// payload (stamp included) regenerates deterministically.
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
// 100% generated, never mixed (kodhama-0007 rule 4's ownership rule).
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
	entries, err := os.ReadDir(vendoredPayloadDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if _, ok := files[e.Name()]; !ok {
			t.Errorf("stray file in the payload dir: %s — reference/ is 100%% generated (kodhama-0007 rule 4)", e.Name())
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
// writes the complete payload; --out is required so it never scatters files.
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
