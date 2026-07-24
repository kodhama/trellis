package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestRepoOverlayIsCurrent is the self-application sync-guard (decision-0035),
// reworked by #117 (kodhama-0007 slice 1) and again by decision-0051 (the authority
// split): the repo's committed .trellis/ overlay is diffed against the payload file
// set — a fresh in-process render (payloadFiles()), the same render
// TestVendoredPayloadIsCurrent pins the vendored plugins/trellis/reference/ copy to.
// The two guards together give generator == vendored payload == repo overlay: drift
// stays impossible, and the repo's own overlay is exactly a mechanical copy of the
// shipped artifact (kodhama-0007 rule 2). The repo's posture is a/conductor (all
// rows active, so internal/rules.md is the shipped all-active assembly); the
// installed `version` stamp is tracked too: the self checkout is a real supported
// Codex startup fixture, so its four authoritative inputs must all exist.
//
// Regenerate on failure — the manual copy path, applied to ourselves (from the repo
// root, after `go run . payload --out ../plugins/trellis/reference` in cli/):
//
//	mkdir -p .trellis/internal
//	cp plugins/trellis/reference/trellis-a.md  .trellis/internal/trellis.md
//	cp plugins/trellis/reference/rules.md      .trellis/internal/rules.md
//	cp plugins/trellis/reference/invariants.md .trellis/internal/invariants.md
//	cp plugins/trellis/reference/version       .trellis/internal/version
func TestRepoOverlayIsCurrent(t *testing.T) {
	repoTrellis := filepath.Join("..", ".trellis")
	if _, err := os.Stat(repoTrellis); err != nil {
		t.Skip("no repo .trellis/ overlay yet — nothing to guard")
	}

	payload := payloadFiles()
	gitignore, err := os.ReadFile(filepath.Join("..", ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(gitignore), ".trellis/internal/version") {
		t.Error("repo .gitignore must not exclude the installed version input required by the self Codex startup fixture")
	}
	for overlayName, payloadName := range map[string]string{
		"internal/trellis.md":    "trellis-a.md", // conductor — the repo holds all invariants firmly
		"internal/rules.md":      "rules.md",     // all rows active — the shipped all-active assembly
		"internal/invariants.md": "invariants.md",
		"internal/version":       "version", // diagnostic provenance required by spec-0007
	} {
		got, err := os.ReadFile(filepath.Join(repoTrellis, overlayName))
		if err != nil {
			t.Fatalf("repo overlay missing .trellis/%s — copy it from the payload (see this test's doc comment): %v", overlayName, err)
		}
		if string(got) != payload[payloadName] {
			t.Errorf(".trellis/%s is stale vs the payload's %s — re-copy it from the payload (see this test's doc comment)", overlayName, payloadName)
		}
	}

	// The flat pre-decision-0051 layout is migrated, not left beside the new one:
	// the old generated paths must be gone (their content lives under internal/).
	for _, legacy := range []string{"trellis.md", "profile.md", "invariants.md", "version"} {
		if _, err := os.Stat(filepath.Join(repoTrellis, legacy)); !os.IsNotExist(err) {
			t.Errorf(".trellis/%s still exists — the flat layout retired with decision-0051; delete the old-path copy", legacy)
		}
	}

	// The managed block in the repo's CLAUDE.md is the payload's block-claude.md,
	// verbatim — the copier contract, applied to ourselves.
	c, err := os.ReadFile(filepath.Join("..", "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	i := strings.Index(string(c), trellisBegin)
	j := strings.Index(string(c), trellisEnd)
	if i < 0 || j <= i {
		t.Fatal("repo CLAUDE.md has no trellis:begin/end block — paste the payload's block-claude.md")
	}
	if block := string(c)[i : j+len(trellisEnd)]; block != payload["block-claude.md"] {
		t.Error("repo CLAUDE.md's managed block is stale vs the payload's block-claude.md — re-paste it between the markers")
	}
}

// TestRepoDeclaresRulesConfig: decision-0051 rules 1+2 — the repo's own overlay
// carries the consumer-authoritative config file, .trellis/rules.toml, seeded from
// the conductor posture (the posture TestRepoOverlayIsCurrent's trellis-a.md mapping
// pins) with every assessable row active (which is what pins internal/rules.md to
// the all-active assembly). Rows are the machine-read truth, so the machine-read
// keys are asserted; nothing else in .trellis/ root is pinned (it is
// consumer-authoritative).
func TestRepoDeclaresRulesConfig(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", ".trellis", "rules.toml"))
	if err != nil {
		t.Fatalf("repo overlay has no .trellis/rules.toml — the consumer-authoritative config (decision-0051 rule 1): %v", err)
	}
	content := string(b)
	if !strings.Contains(content, `strictness  = "firm"`) {
		t.Errorf(".trellis/rules.toml must declare strictness \"firm\" (the a/conductor posture the repo overlay is pinned to), got: %q", content)
	}
	for _, slug := range assessableSlugs {
		rowRe := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(slug) + `\s+= \{ active = true \}`)
		if !rowRe.MatchString(content) {
			t.Errorf(".trellis/rules.toml must carry an active row for %s — the repo holds every invariant firmly (its internal/rules.md is pinned to the all-active assembly)", slug)
		}
	}
}

// TestRepoOverlayCarriesNoExpressionFile: decision-0051 amendment (2026-07-19,
// append-only foot of the record) — expression.md is retired from the bundle; the
// consumer root is rules.toml alone, and a project's governance prose belongs in
// its own instructions file. The repo's own file was deleted under the amendment
// (its body was a pointer to CLAUDE.md §Operating method, already present there —
// the maintainer ratified the deletion), so self-application parity
// (decision-0035) means the file stays gone, same idiom as the flat-layout
// absence checks in TestRepoOverlayIsCurrent. (Replaces the retired
// TestRepoExpressionIsPureProse.)
func TestRepoOverlayCarriesNoExpressionFile(t *testing.T) {
	if _, err := os.Stat(filepath.Join("..", ".trellis", "expression.md")); !os.IsNotExist(err) {
		t.Error(".trellis/expression.md still exists — expression.md retired from the bundle (decision-0051 amendment); governance prose belongs in the project's own instructions file")
	}
}

// TestSharedProjectInstructionEntrypoints guards spec-0006 AC1–AC6 and AC8:
// AGENTS.md is the shared project-instruction authority, CLAUDE.md is the exact
// Claude adapter plus the existing Trellis import block, and the bounded
// current-truth surfaces point shared-method references at AGENTS.md.
func TestSharedProjectInstructionEntrypoints(t *testing.T) {
	readRepoFile := func(name string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join("..", name))
		if err != nil {
			t.Fatalf("read repo %s: %v", name, err)
		}
		return string(b)
	}

	agents := readRepoFile("AGENTS.md")
	claude := readRepoFile("CLAUDE.md")
	referenceBlock := readRepoFile("plugins/trellis/reference/block-claude.md")
	referenceCodexBlock := readRepoFile("plugins/trellis/reference/block-codex.md")
	normalizedAgents := strings.Join(strings.Fields(agents), " ")

	wantClaude := "@AGENTS.md\n\n" + referenceBlock + "\n"
	if claude != wantClaude {
		t.Errorf("CLAUDE.md must be exactly @AGENTS.md, one blank separator, the byte-identical block-claude.md payload, and one final newline")
	}
	if strings.Count(claude, "@AGENTS.md") != 1 {
		t.Errorf("CLAUDE.md must contain exactly one @AGENTS.md adapter, got %d", strings.Count(claude, "@AGENTS.md"))
	}
	if strings.Count(claude, trellisBegin) != 1 || strings.Count(claude, trellisEnd) != 1 {
		t.Errorf("CLAUDE.md must contain exactly one matched Trellis managed block")
	}
	for _, duplicate := range []string{"# Trellis — operating method", "<!-- grove:begin", "## Operating method"} {
		if strings.Contains(claude, duplicate) {
			t.Errorf("CLAUDE.md duplicates shared Layer-B/Grove prose %q; it must remain an adapter only", duplicate)
		}
	}

	for label, statement := range map[string]string{
		"canonical authority": "`AGENTS.md` is the canonical home for shared project instructions",
		"shared-rule edits":   "Edit new shared rules here, outside managed blocks",
		"Claude adapter":      "`CLAUDE.md` is the Claude adapter, not a shared-rule edit surface",
		"Claude-only rules":   "Genuinely Claude-only rules belong in `.claude/rules/`",
		"project choices":     "Grove and Trellis project choices remain in `.grove/` and `.trellis/` configuration files",
		"managed-block edits": "Do not hand-edit managed blocks",
	} {
		if !strings.Contains(normalizedAgents, statement) {
			t.Errorf("AGENTS.md is missing the %s routing statement %q", label, statement)
		}
	}
	maintainingSection := "## Maintaining project instructions"
	if strings.Count(agents, maintainingSection) != 1 || strings.Index(agents, maintainingSection) > strings.Index(agents, "<!-- grove:begin") {
		t.Error("AGENTS.md must contain one Maintaining project instructions section before its managed blocks")
	}
	for _, sharedContent := range []string{"# Trellis — operating method", "## Operating method", "<!-- grove:begin", "<!-- grove:end -->"} {
		if !strings.Contains(agents, sharedContent) {
			t.Errorf("AGENTS.md is missing moved Layer-B/Grove content %q", sharedContent)
		}
	}
	if strings.Count(agents, codexBootstrapBegin) != 1 || strings.Count(agents, codexBootstrapEnd) != 1 {
		t.Error("AGENTS.md must contain exactly one generated Codex bootstrap marker pair (spec-0007@v1)")
	}
	if strings.Contains(agents, "@.trellis/") {
		t.Error("AGENTS.md Codex bootstrap must contain no Claude @.trellis imports")
	}
	if blockStart := strings.Index(agents, codexBootstrapBegin); blockStart < 0 ||
		agents[blockStart:strings.Index(agents[blockStart:], codexBootstrapEnd)+blockStart+len(codexBootstrapEnd)] != referenceCodexBlock {
		t.Error("AGENTS.md's Codex bootstrap must be byte-identical to block-codex.md")
	}
	if strings.Contains(agents, rulesAuthorityHeader) {
		t.Error("AGENTS.md must not embed the generated Trellis rule readout")
	}

	boundedReferences := map[string]struct {
		wantAGENTS  bool
		allowClaude bool
	}{
		"README.md":                         {wantAGENTS: true, allowClaude: true},
		"profiles/trellis-self.md":          {wantAGENTS: true},
		".grove/config.toml":                {wantAGENTS: true},
		".grove/README.md":                  {wantAGENTS: true},
		".claude/agents/corpus-reviewer.md": {wantAGENTS: true},
	}
	for name, expectation := range boundedReferences {
		content := readRepoFile(name)
		if expectation.wantAGENTS && !strings.Contains(content, "AGENTS.md") {
			t.Errorf("%s must name AGENTS.md as the shared project-instruction authority", name)
		}
		if !expectation.allowClaude && strings.Contains(content, "CLAUDE.md") {
			t.Errorf("%s retains a stale CLAUDE.md shared-method reference", name)
		}
	}

	readme := readRepoFile("README.md")
	for _, adapterReference := range []string{
		"managed block in your `CLAUDE.md`",
		`block-claude.md >> CLAUDE.md`,
	} {
		if !strings.Contains(readme, adapterReference) {
			t.Errorf("README.md must retain Claude-adapter-specific reference %q", adapterReference)
		}
	}

	workflow := readRepoFile(".github/workflows/cli-ci.yml")
	pullRequestStart := strings.Index(workflow, "  pull_request:\n")
	pushStart := strings.Index(workflow, "  push:\n")
	jobsStart := strings.Index(workflow, "\njobs:\n")
	if pullRequestStart < 0 || pushStart <= pullRequestStart || jobsStart <= pushStart {
		t.Fatal("cli-ci must retain distinct pull_request and push trigger sections before jobs")
	}
	pullRequestTrigger := workflow[pullRequestStart:pushStart]
	pushTrigger := workflow[pushStart:jobsStart]
	for _, path := range []string{`"AGENTS.md"`, `"CLAUDE.md"`, `".trellis/**"`} {
		if !strings.Contains(pullRequestTrigger, path) {
			t.Errorf("cli-ci pull-request path filter is missing %s", path)
		}
		if !strings.Contains(pushTrigger, path) {
			t.Errorf("cli-ci main-push path filter is missing %s", path)
		}
	}
}
