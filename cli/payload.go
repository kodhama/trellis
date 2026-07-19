package main

// The release-render pipeline (kodhama-0007 rule 1, "render once, at release";
// kodhama/trellis#117). The full M1 variant space is enumerable — 2 postures ×
// 2 block styles, where the profile is posture-invariant today and the CLAUDE.md
// block is a constant — so every bundle file every writer will ever need is
// pre-rendered here and vendored in plugins/trellis/reference/. Downstream
// writers only copy, paste between markers, and verify (rule 2); the payload
// ships a checksum manifest so anything can verify with standard tools —
// `shasum -a 256 -c checksums` — or the CI regenerate-and-diff guard in
// payload_test.go (rule 3).

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// payloadFiles renders the complete pre-rendered M1 payload (decision-0051 shape):
// the verbatim catalog, the per-rule fragments plus their header/footer
// (rules/<slug>.md — the assembly source setup concatenates in catalog order), the
// assembled all-active readout (rules.md — the common case's copy source and the
// concatenation oracle), both posture variants of the header / inline block /
// rules.toml seed, the single hand-owned expression seed, the constant CLAUDE.md
// block, a content-derived version stamp, and the checksums manifest. The rules.toml
// seeds and the expression seed are manifest-covered like any payload file; only the
// *installed* consumer-root copies (.trellis/rules.toml, .trellis/expression.md)
// sit outside verification, because the consumer owns them from the moment they are
// seeded (decision-0051 rule 1).
func payloadFiles() map[string]string {
	files := map[string]string{
		"invariants.md":        invariantsRef, // the catalog, verbatim (decision-0028 single source)
		"block-claude.md":      renderClaudeBlock(),
		"expression.md":        renderExpressionSeed(),
		"rules.md":             renderRulesReadout(),
		"block-inline-tail.md": renderInlineBlockTail(), // posture-independent — one tail, not two
	}
	for name, content := range ruleFragments() {
		files[name] = content
	}
	for _, p := range allProfiles {
		files["trellis-"+p.Key+".md"] = renderHeader(p)
		files["block-inline-"+p.Key+".md"] = renderInlineBlock(p)
		files["block-inline-"+p.Key+"-head.md"] = renderInlineBlockHead(p)
		files["rules-"+p.Key+".toml"] = renderRulesToml(p)
	}

	// The payload's version stamp is derived from its own content: a vendored file
	// cannot carry the commit sha that will contain it, and the generator's build
	// version would make local regeneration nondeterministic. A content hash changes
	// exactly when the payload changes — the "versioned payload" identity of
	// kodhama-0007 rule 1. Since #120 (decision-0043) this stamp IS the install
	// stamp: every writer copies it to .trellis/version, and the staleness hook
	// compares the two files. (The plugin@<sha> install stamp of decision-0039
	// rule 2 is superseded in part by decision-0043.)
	files["version"] = "payload@" + manifestHash(files)[:12] + "\n"
	files["checksums"] = manifestLines(files)
	return files
}

// manifestLines renders the shasum-compatible manifest: "<sha256>  <name>" per line
// (two spaces — text mode), sorted by name, covering every payload file passed in.
func manifestLines(files map[string]string) string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, name := range names {
		fmt.Fprintf(&b, "%x  %s\n", sha256.Sum256([]byte(files[name])), name)
	}
	return b.String()
}

// manifestHash is the payload's content identity: the sha256 of the manifest over
// the content files (everything rendered before the version stamp itself).
func manifestHash(files map[string]string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(manifestLines(files))))
}

// payload is the release-time generator command (#117): render the full payload
// into --out. Release tooling, not an end-user command — the vendored copy in
// plugins/trellis/reference/ is what ships; TestVendoredPayloadIsCurrent keeps it
// impossible to drift from this render.
func payload(in io.Reader, w io.Writer, args []string) error {
	fs := flag.NewFlagSet("payload", flag.ContinueOnError)
	fs.SetOutput(w)
	out := fs.String("out", "", "directory to render the payload into (the vendored home is plugins/trellis/reference)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *out == "" {
		return fmt.Errorf("payload needs --out <dir> — e.g. (from cli/) go run . payload --out ../plugins/trellis/reference")
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", *out, err)
	}

	files := payloadFiles()
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		target := filepath.Join(*out, name)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("creating dir for %s: %w", name, err)
		}
		if err := os.WriteFile(target, []byte(files[name]), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", name, err)
		}
	}
	fmt.Fprintf(w, "rendered payload (%d files) into %s\n  %s\n  verify: shasum -a 256 -c checksums  (from that dir)\n",
		len(files), *out, strings.Join(names, " "))
	return nil
}
