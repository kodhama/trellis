package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func planFor(profileKey string) Plan {
	p, _ := profileByKey(profileKey)
	m, _ := modeByKey("m1")
	mdl, _ := modelByKey("none")
	return Plan{Harness: Harness{Name: "Claude Code"}, Profile: p, Mode: m, Model: mdl}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestApplyM1WritesOverlay(t *testing.T) {
	dir := t.TempDir()
	if _, err := applyM1(dir, planFor("b")); err != nil {
		t.Fatalf("applyM1: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".trellis", "profile.md")); err != nil {
		t.Errorf(".trellis/profile.md not written: %v", err)
	}
	// CLAUDE.md: minimal — a human line + the header import, and no governance
	// content duplicated into the host's file.
	c := readFile(t, filepath.Join(dir, "CLAUDE.md"))
	for _, want := range []string{trellisBegin, trellisEnd, "@.trellis/trellis.md", "follows **Trellis**"} {
		if !strings.Contains(c, want) {
			t.Errorf("CLAUDE.md missing %q", want)
		}
	}
	if strings.Contains(c, "settled ground") {
		t.Error("the rules should live in .trellis/, not be inlined into CLAUDE.md")
	}
	// The header carries the imperative framing + strength and imports the profile.
	header := readFile(t, filepath.Join(dir, ".trellis", "trellis.md"))
	for _, want := range []string{"Follow the rules below", "@profile.md"} {
		if !strings.Contains(header, want) {
			t.Errorf(".trellis/trellis.md missing %q", want)
		}
	}
	// The profile carries the active directives; the bundled reference carries the catalog.
	if prof := readFile(t, filepath.Join(dir, ".trellis", "profile.md")); !strings.Contains(prof, "settled ground") {
		t.Errorf(".trellis/profile.md missing the active directives: %q", prof)
	}
	if inv := readFile(t, filepath.Join(dir, ".trellis", "invariants.md")); !strings.Contains(inv, "inv-directional-flow") {
		t.Error(".trellis/invariants.md should contain the bundled invariant reference")
	}
}

func TestApplyM1AugmentNotClobber(t *testing.T) {
	dir := t.TempDir()
	original := "# My Project\n\nSome existing house rules.\n"
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := applyM1(dir, planFor("a")); err != nil {
		t.Fatal(err)
	}
	c := readFile(t, filepath.Join(dir, "CLAUDE.md"))
	if !strings.Contains(c, "Some existing house rules.") {
		t.Error("existing content was clobbered")
	}
	if !strings.Contains(c, trellisBegin) {
		t.Error("trellis block not appended")
	}
}

func TestApplyM1Idempotent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# My Project\n\nHouse rules.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := applyM1(dir, planFor("a")); err != nil {
		t.Fatal(err)
	}
	if _, err := applyM1(dir, planFor("a")); err != nil {
		t.Fatal(err)
	}
	c := readFile(t, filepath.Join(dir, "CLAUDE.md"))
	if n := strings.Count(c, trellisBegin); n != 1 {
		t.Errorf("expected exactly one trellis block after re-run, got %d", n)
	}
	if !strings.Contains(c, "House rules.") {
		t.Error("re-run clobbered existing content")
	}
}

// kodhama/trellis#112: a hand-appended section below the generated block in
// .trellis/profile.md was silently destroyed on the next `trellis setup -apply`,
// with no warning. profile.md stays a pure generated snapshot (decision-0035) — this
// only makes the loss visible, it doesn't preserve the content.
func TestApplyM1WarnsBeforeOrphaningProfileContent(t *testing.T) {
	dir := t.TempDir()
	if _, err := applyM1(dir, planFor("b")); err != nil {
		t.Fatal(err)
	}
	profilePath := filepath.Join(dir, ".trellis", "profile.md")
	appended := readFile(t, profilePath) + "\n## project expression\n\nSome hand-authored, project-specific content.\n"
	if err := os.WriteFile(profilePath, []byte(appended), 0o644); err != nil {
		t.Fatal(err)
	}

	summary, err := applyM1(dir, planFor("b"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(summary, "hand-authored content") || !strings.Contains(summary, "project expression") {
		t.Errorf("expected a warning naming the lost content, got: %q", summary)
	}
	// The tool still fully regenerates the file — decision-0035 requires
	// byte-identical output; the warning informs, it does not preserve.
	if got := readFile(t, profilePath); strings.Contains(got, "project expression") {
		t.Error("profile.md should still be fully regenerated after warning")
	}
}

// kodhama/trellis#119 (kodhama-0007 rule 4): the warning's move-it-here pointer names
// `.trellis/expression.md` — the hand-owned home — not the instructions file. PR #114's
// original "move it into your instructions file (e.g. CLAUDE.md)" guidance is
// superseded; every home of that guidance must agree.
func TestWarnPointsAtExpressionHome(t *testing.T) {
	dir := t.TempDir()
	if _, err := applyM1(dir, planFor("b")); err != nil {
		t.Fatal(err)
	}
	tdir := filepath.Join(dir, ".trellis")
	profilePath := filepath.Join(tdir, "profile.md")
	appended := readFile(t, profilePath) + "\nHand-authored expression content.\n"
	if err := os.WriteFile(profilePath, []byte(appended), 0o644); err != nil {
		t.Fatal(err)
	}

	warning := warnOrphanedProfileContent(tdir)
	if !strings.Contains(warning, ".trellis/expression.md") {
		t.Errorf("warning must point at the hand-owned home .trellis/expression.md (kodhama-0007 rule 4, #119), got: %q", warning)
	}
	if strings.Contains(warning, "CLAUDE.md") {
		t.Errorf("warning still carries the superseded instructions-file guidance (PR #114 → #119): %q", warning)
	}
}

func TestApplyM1NoWarningOnPlainRerun(t *testing.T) {
	dir := t.TempDir()
	if _, err := applyM1(dir, planFor("b")); err != nil {
		t.Fatal(err)
	}
	summary, err := applyM1(dir, planFor("b"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(summary, "hand-authored content") {
		t.Errorf("unmodified re-run should not warn, got: %q", summary)
	}
}

func TestUpsertBlockReplaces(t *testing.T) {
	content := "top\n\n" + trellisBegin + "\nOLD\n" + trellisEnd + "\n\nbottom\n"
	out := upsertBlock(content, trellisBegin+"\nNEW\n"+trellisEnd)
	if strings.Contains(out, "OLD") {
		t.Error("old block not replaced")
	}
	for _, want := range []string{"NEW", "top", "bottom"} {
		if !strings.Contains(out, want) {
			t.Errorf("upsert lost %q: %q", want, out)
		}
	}
	if n := strings.Count(out, trellisBegin); n != 1 {
		t.Errorf("expected one block, got %d", n)
	}
}
