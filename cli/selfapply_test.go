package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestRepoOverlayIsCurrent was DELETED by decision-0071, not disabled.
//
// It compared this repo's committed .trellis/internal/ overlay against a fresh
// payload render, giving "generator == vendored payload == repo overlay". The
// repo no longer holds an overlay: it consumes the plugin's payload at session
// start like any other project, so the drift that guard prevented is now
// structurally impossible rather than tested — there is no second copy to drift.
//
// The other two links in that chain are untouched and still enforced:
// TestVendoredPayloadIsCurrent pins plugins/trellis/reference/ to the generator,
// and install.sh's baked manifest pins what a consumer receives. Only the
// overlay-to-payload comparison is gone, with its subject.
//
// decision-0035's self-application PRINCIPLE stands; its overlay sync-guard is
// superseded in part.
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

	// decision-0071: this repo self-applies through the PLUGIN, so CLAUDE.md no
	// longer carries a managed block. What still holds is the adapter shape — one
	// @AGENTS.md and nothing duplicated from Layer B — and, newly asserted, that
	// no block came back. A managed block here would mean the repo had silently
	// returned to the overlay mode decision-0065 retired for everyone else, and
	// would make install.sh refuse to render (decision-0068 D12).
	if strings.TrimSpace(claude) != "@AGENTS.md" {
		t.Errorf("CLAUDE.md must be exactly the @AGENTS.md adapter since decision-0071; got:\n%s", claude)
	}
	if strings.Count(claude, trellisBegin) != 0 || strings.Count(claude, trellisEnd) != 0 {
		t.Errorf("CLAUDE.md carries a Trellis managed block — decision-0071 removed it, and its return means this repo has drifted back to the overlay delivery it retired for consumers")
	}
	_ = referenceBlock
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
