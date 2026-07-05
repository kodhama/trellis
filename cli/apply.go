package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// The invariant reference shipped in the overlay. Kept in sync from the single
// source in core/ by the generate step below (run `go generate ./...` in cli/).
//
//go:generate cp ../core/catalog/signature-catalog-v1.md assets/invariants.md
//
//go:embed assets/invariants.md
var invariantsRef string

// The M1 overlay writes into CLAUDE.md between these markers only. Everything
// outside them is the host's and is never touched (augment-never-clobber); a
// re-run replaces what is between them (idempotent).
const (
	trellisBegin = "<!-- trellis:begin (managed by trellis — edit .trellis/, not this block) -->"
	trellisEnd   = "<!-- trellis:end -->"
)

// applyM1 performs the deterministic M1 overlay: write the .trellis/ bundle
// (a header, the profile, the invariant reference) and add a minimal import of
// the header to CLAUDE.md. No model — plain file editing.
//
// Layout (imports resolve relative to the importing file — verified against the
// Claude Code docs):
//
//	CLAUDE.md      -> @.trellis/trellis.md      (header, auto-loaded)
//	.trellis/trellis.md -> @profile.md          (profile, auto-loaded)
//	                    -> `.trellis/invariants.md` (backticked = read on demand)
func applyM1(dir string, plan Plan) (string, error) {
	tdir := filepath.Join(dir, ".trellis")
	if err := os.MkdirAll(tdir, 0o755); err != nil {
		return "", fmt.Errorf("creating .trellis/: %w", err)
	}
	bundle := map[string]string{
		"trellis.md":    renderHeader(plan),
		"profile.md":    renderProfile(plan),
		"invariants.md": invariantsRef,
	}
	for name, content := range bundle {
		if err := os.WriteFile(filepath.Join(tdir, name), []byte(content), 0o644); err != nil {
			return "", fmt.Errorf("writing .trellis/%s: %w", name, err)
		}
	}

	claudePath := filepath.Join(dir, "CLAUDE.md")
	existing := ""
	if b, err := os.ReadFile(claudePath); err == nil {
		existing = string(b)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("reading CLAUDE.md: %w", err)
	}
	if err := os.WriteFile(claudePath, []byte(upsertBlock(existing, renderClaudeBlock())), 0o644); err != nil {
		return "", fmt.Errorf("writing CLAUDE.md: %w", err)
	}

	return "applied (M1 overlay):\n" +
		"  wrote .trellis/{trellis,profile,invariants}.md\n" +
		"  updated CLAUDE.md (imports .trellis/trellis.md)\n", nil
}

// upsertBlock replaces the delimited trellis block in content if present, else
// appends it. Content outside the markers is preserved exactly.
func upsertBlock(content, block string) string {
	i := strings.Index(content, trellisBegin)
	j := strings.Index(content, trellisEnd)
	if i >= 0 && j > i {
		return content[:i] + block + content[j+len(trellisEnd):]
	}
	if strings.TrimSpace(content) == "" {
		return block + "\n"
	}
	return strings.TrimRight(content, "\n") + "\n\n" + block + "\n"
}

// renderClaudeBlock is the minimal CLAUDE.md footprint: a human-readable line plus
// a native @import of the header. Everything else lives in .trellis/.
func renderClaudeBlock() string {
	return trellisBegin + "\n" +
		"This project is governed by **Trellis** (see the `.trellis/` folder). Its rules are imported here:\n" +
		"@.trellis/trellis.md\n" +
		trellisEnd
}

// renderHeader is the entry point CLAUDE.md imports: the intro + the governance
// behavior, then it pulls in the profile and points at the invariant reference.
func renderHeader(plan Plan) string {
	return "# Trellis governance\n\n" +
		"This project is supervised by **Trellis**, a governance layer over your existing process: it holds a small set of invariants at the strengths this project has adopted, and otherwise respects your methodology.\n\n" +
		"**Key behavior:** surface any **human-gated handover performed without its human approval** (invariant B2). Agent-gated handovers proceed silently. Gatekeepers are whatever this project already declares — respected, not imposed (decision-0024).\n\n" +
		"## This project's profile\n\n" +
		"@profile.md\n\n" +
		"## Invariant reference\n\n" +
		"Full definitions — what each invariant is, why it matters, and how it's honored vs violated — live in `.trellis/invariants.md`. Read it when you need the detail behind a rule.\n"
}

// renderProfile is the tunable readout: posture, active invariants, dials. The
// governance behavior lives in the header (single source), not here.
func renderProfile(plan Plan) string {
	return "# Trellis expression profile\n\n" +
		fmt.Sprintf("- posture: %s — %s\n", plan.Profile.Name, plan.Profile.Description) +
		fmt.Sprintf("- enforcement (C1) lean: `%s`\n", plan.Profile.C1Lean) +
		fmt.Sprintf("- active invariants: %s\n", activeList(plan)) +
		"- gatekeeper (C2): detected from this project, not preset (decision-0024)\n" +
		fmt.Sprintf("- install mode: %s\n", plan.Mode.Name) +
		"\nEdit this file to tune the profile; `CLAUDE.md` imports `.trellis/trellis.md`, which imports this.\n"
}

func activeList(plan Plan) string {
	if len(plan.Profile.Active) == 0 {
		return "all assessable invariants"
	}
	return strings.Join(plan.Profile.Active, ", ")
}
