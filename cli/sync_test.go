package main

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// decision-0028: a source and its derived resources must stay in sync, and it must
// be checked, not remembered. The catalog (core/catalog/signature-catalog-v1.md) is
// the source; these are its derivatives.

// TestBundledCatalogInSync: assets/invariants.md stays byte-identical to the
// catalog source (decision-0028, unchanged — the //go:generate cp step is
// untouched); the payload copy (plugins/trellis/reference/invariants.md) is
// the same source with its leading frontmatter block stripped (decision-0054
// points 1+3 — the payload ships without frontmatter; the check is precise
// and mechanical via stripFrontmatter, not a loosened byte-equality check).
func TestBundledCatalogInSync(t *testing.T) {
	src, err := os.ReadFile("../core/catalog/signature-catalog-v1.md")
	if err != nil {
		t.Fatal(err)
	}

	assetsCopy, err := os.ReadFile("assets/invariants.md")
	if err != nil {
		t.Fatalf("reading assets/invariants.md: %v", err)
	}
	if string(assetsCopy) != string(src) {
		t.Errorf("assets/invariants.md is out of sync with the catalog. Regenerate it (`go generate ./...` in cli/, "+
			"or `cp core/catalog/signature-catalog-v1.md` to it). [decision-0028]")
	}

	payloadCopy, err := os.ReadFile("../plugins/trellis/reference/invariants.md")
	if err != nil {
		t.Fatalf("reading ../plugins/trellis/reference/invariants.md: %v", err)
	}
	want := stripFrontmatter(string(src))
	if string(payloadCopy) != want {
		t.Errorf("plugins/trellis/reference/invariants.md is out of sync with the catalog (frontmatter stripped). "+
			"Regenerate the payload (`go run . payload --out ../plugins/trellis/reference` in cli/). [decision-0028, decision-0054]")
	}
}

// TestInvariantsPageMatchesCatalog: every catalog example appears (stripped of
// markdown) on the rendered invariants page, so the page can't drift from the source.
func TestInvariantsPageMatchesCatalog(t *testing.T) {
	cat, err := os.ReadFile("../core/catalog/signature-catalog-v1.md")
	if err != nil {
		t.Fatal(err)
	}
	page, err := os.ReadFile("../docs/invariants.html")
	if err != nil {
		t.Fatal(err)
	}
	pageText := string(page)

	exRe := regexp.MustCompile(`(?m)^\s*- \*\(\w+\)\* (.+)$`)
	for _, m := range exRe.FindAllStringSubmatch(string(cat), -1) {
		want := stripMd(m[1])
		if want == "" {
			continue
		}
		if !strings.Contains(pageText, want) {
			t.Errorf("docs/invariants.html is out of sync with the catalog — this example is missing:\n"+
				"  %q\nRegenerate the page from the catalog. [decision-0028]", want)
		}
	}
}

func stripMd(s string) string {
	s = regexp.MustCompile(`\*\*(.+?)\*\*`).ReplaceAllString(s, "$1")
	s = regexp.MustCompile(`\*(.+?)\*`).ReplaceAllString(s, "$1")
	s = strings.ReplaceAll(s, "`", "")
	return strings.TrimSpace(s)
}
