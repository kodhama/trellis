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

// TestBundledCatalogInSync: the two bundled copies are byte-identical to the catalog.
func TestBundledCatalogInSync(t *testing.T) {
	src, err := os.ReadFile("../core/catalog/signature-catalog-v1.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, cp := range []string{"assets/invariants.md", "../plugins/trellis/reference/invariants.md"} {
		b, err := os.ReadFile(cp)
		if err != nil {
			t.Fatalf("reading %s: %v", cp, err)
		}
		if string(b) != string(src) {
			t.Errorf("%s is out of sync with the catalog. Regenerate it (`go generate ./...` in cli/, "+
				"or `cp core/catalog/signature-catalog-v1.md` to both copies). [decision-0028]", cp)
		}
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
