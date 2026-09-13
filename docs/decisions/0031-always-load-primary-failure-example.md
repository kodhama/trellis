---
id: decision-0031
type: decision
status: ratified
depends_on: [decision-0026, decision-0027]
owner: gundi
ratified: 2026-07-05
---

# 0031 — Always-load one primary failure example per active rule

## Context

`decision-0026` put the active invariants in context as one-line *rules* (always loaded), with the full
why + examples on-demand — but flagged its own risk: agents trip on abstract rules. `decision-0027` made
the catalog examples matched without→with pairs. The missing piece: the always-loaded rules had no
concrete grounding.

## Decision

Under each always-loaded rule, also load **one** example — the invariant's **primary `violated` case**
(the failure to avoid), a terse `✗` line pulled from the bundled catalog (`violated[0]`).

- **The failure, not a pair.** For an abstract rule the concrete *mistake to catch* is the highest-signal
  grounding; the rule already states the "do". So one violated example beats a balanced pair here.
- **One, to stay terse.** The cost is ~one line per active invariant (`decision-0026`'s leanness is the
  constraint). Use-case bias — a single example anchoring the rule to one domain — is mitigated by the
  abstract rule sitting above it and by spreading the examples across layers, **not** by loading a
  second. A second is available as a later knob if a specific invariant still skews.
- **Curation by ordering.** The example we want always-loaded is placed first in the catalog; the parser
  (`invariantPrimaryFailure`) pulls `violated[0]`. No new catalog syntax.

## Consequences

- `renderProfile` and the inline overlay block (shared `activeRuleLines`) emit `- **slug** — <rule>` then
  `    ✗ <primary failure>`. The always-loaded descriptions were updated to match (no drift). The catalog
  is unchanged — the examples already exist.
- Full pairs stay on-demand in `.trellis/invariants.md`.

## Open questions

- Whether some invariants warrant 2 (to de-bias further) — revisit if a rule reads as domain-specific in
  practice; the knob is a small change.
