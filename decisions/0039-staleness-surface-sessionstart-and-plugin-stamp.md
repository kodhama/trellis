---
id: decision-0039
type: decision
status: ratified
depends_on: [decision-0035, decision-0036]
owner: gundi
ratified: 2026-07-07
---

# 0039 — The staleness surface is a `SessionStart` hook; the plugin stamps `.trellis/version` too

## Context

`decision-0035` committed to *"drift is made visible, not silent"*: the overlay is stamped with
the version that generated it, and `trellis status` surfaces when it lags. Two mechanics facts,
verified against the Claude Code hooks documentation (checked 2026-07-06, re-verified
2026-07-07), correct that design before the update slice builds on wrong assumptions:

1. **`InstructionsLoaded` hooks cannot carry the staleness message.** The event exists (fires
   when a `CLAUDE.md`/rules file is loaded) but is **side-effect-only** — the docs list it
   under *"No decision control. Used for side effects like logging or cleanup"*; it cannot
   inject context. The event that **can** put words in front of the agent is **`SessionStart`**:
   *"Any text your hook script prints to stdout is added as context for Claude"*, plus the
   event-specific **`additionalContext`** field — *"string added to Claude's context at the
   start of the conversation, before the first prompt."* No current spec or decision names
   `InstructionsLoaded` (checked — repo-wide grep is clean); this decision pins the constraint
   so the pending update/supervisor slice is specced on the right event.
2. **The plugin path writes no stamp.** The CLI's M1 apply writes `.trellis/version`
   (`cli/apply.go` — *"the D1 staleness marker `trellis status` reads"*), but the plugin's
   setup skill (`plugins/trellis/skills/setup/SKILL.md`) wrote only the three bundle files. A
   plugin-installed overlay is therefore **invisible to every staleness check** — worse than
   silently stale: `trellis status` reports *"no Trellis overlay"* for a project that has one.
   That is exactly the `floor-transparency` violation `decision-0035` was written to close,
   reopened through the second install door.

## Decision

1. **Any agent-facing staleness surface SHALL be a `SessionStart` hook emitting
   `additionalContext`** (or plain stdout) — e.g. *"the Trellis overlay is behind what's
   installed — refresh it"* — never an `InstructionsLoaded` hook, which cannot inject context.
   Owed to the update slice when it is built; recorded now so it isn't specced wrong.
2. **The plugin setup skill writes `.trellis/version`, same as the CLI.** Stamp format
   `plugin@<short-sha>` — the commit SHA *is* the plugin version (`decision-0036`) — obtained
   from `git -C "${CLAUDE_PLUGIN_ROOT}" rev-parse --short HEAD`; fallback `plugin@unknown` if
   the plugin root is not a git checkout. An honest "unknown" stamp still beats an absent one:
   the overlay becomes *visible* to checks.

## Consequences

- The setup skill's bundle step and confirmation step now include the stamp (this change).
- `trellis status` on a plugin-installed overlay now reports *what generated it* and that it
  differs from the CLI binary — accurate (different generator), and it points at a refresh.
  Distinguishing `plugin@…` stamps with a plugin-appropriate refresh hint is a small follow-up
  (see Open questions).
- The future update surface ("overlay behind — refresh") inherits the `SessionStart` +
  `additionalContext` mechanism as a constraint, not a choice to re-litigate.

## Open questions

- Should `trellis status` special-case `plugin@…` stamps — suggesting `/plugin marketplace
  update` + `/trellis:setup` instead of *"re-run `trellis setup`"*? Small UX fix, owed when
  status is next touched.
- The staleness *comparison* for plugin overlays (is `plugin@abc1234` behind the marketplace
  HEAD?) needs a network check the no-runtime promise resists (`decision-0035` open question,
  unchanged) — the `SessionStart` hook may only be able to say *how old*, not *how far behind*.

## Supersedes / superseded by

— (none; refines `decision-0035`'s staleness surface with verified hook mechanics and closes
its plugin-side gap)
