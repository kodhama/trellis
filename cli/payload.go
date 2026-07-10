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

// payloadFiles renders the complete pre-rendered M1 payload: the verbatim catalog,
// both posture variants of the header / profile / inline block / expression seed
// skeleton, the constant CLAUDE.md block, a content-derived version stamp, and the
// checksums manifest. profile-a/profile-b are byte-identical today (renderProfile is
// posture-invariant, #117's verified evidence) but are named per posture so the
// payload layout survives a posture whose profile diverges. The expression skeletons
// (#119, kodhama-0007 rule 4) are manifest-covered like any payload file; only the
// *installed* .trellis/expression.md sits outside install-time verification, because
// it is hand-owned from the moment it is seeded.
func payloadFiles() map[string]string {
	files := map[string]string{
		"invariants.md":   invariantsRef, // the catalog, verbatim (decision-0028 single source)
		"block-claude.md": renderClaudeBlock(),
	}
	for _, p := range allProfiles {
		plan := Plan{Profile: p}
		files["trellis-"+p.Key+".md"] = renderHeader(plan)
		files["profile-"+p.Key+".md"] = renderProfile(plan)
		files["block-inline-"+p.Key+".md"] = renderInlineBlock(plan)
		files["expression-"+p.Key+".md"] = renderExpressionSkeleton(plan)
	}

	// The payload's version stamp is derived from its own content: a vendored file
	// cannot carry the commit sha that will contain it, and the generator's build
	// version would make local regeneration nondeterministic. A content hash changes
	// exactly when the payload changes — the "versioned payload" identity of
	// kodhama-0007 rule 1. (The install-time .trellis/version stamp — plugin@<sha>,
	// decision-0036/0039 — is the copier's job at install, not part of the payload.)
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
		if err := os.WriteFile(filepath.Join(*out, name), []byte(files[name]), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", name, err)
		}
	}
	fmt.Fprintf(w, "rendered payload (%d files) into %s\n  %s\n  verify: shasum -a 256 -c checksums  (from that dir)\n",
		len(files), *out, strings.Join(names, " "))
	return nil
}
