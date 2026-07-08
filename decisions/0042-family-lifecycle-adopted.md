---
id: decision-0042
type: decision
status: gated
depends_on: [decision-0022, decision-0037]
owner: agent
date: 2026-07-08
---

# 0042 — adopt the family lifecycle for trellis-self

## Context

`decision-0037` made statuses methodology-defined and, for this repo's
own corpus, kept the native two-state `draft → ratified` — explicitly
considering and declining a `gated` ratchet for trellis-self.
`decision-0022` rides ratification on the merge, with the
`draft → ratified` flip carried in the PR's own diff. The family's
2026-07-07 status-vocabulary consolidation preserved that carve-out
deliberately.

The maintainer has since decided uniformity outweighs the variability
(`kodhama-0004-uniform-lifecycle`, ratified 2026-07-08, in the org meta
repo): same states, same mechanic, every family repo — including this
one. This decision is an *application* of `decision-0037`'s own
principle (the methodology declares the vocabulary; the methodology
here is the family's), not a reversal of it.

## Decision

1. **Forward artifacts** in `decisions/` and `specs/` use the family
   enum: `draft → gated → approved (→ superseded)`. `gated` =
   self-checked, agent-consumable. Consuming a `draft` remains
   forbidden.
2. **Mechanic:** a PR carries the `draft → gated` flip; **merging is
   the ratification act**; a post-merge bump commit records `approved`.
   Nobody writes `approved` into a PR's own diff for a new artifact.
   (Supersedes-in-part `decision-0022`'s in-diff `ratified` flip; its
   core — merge *is* ratification — stands.)
3. **History stands** (append-only): artifacts ratified before this
   decision keep `status: ratified`, which reads as `approved` under
   `decision-0037`'s declared equivalence. No relabel — that would
   forge history for zero information gain.
4. **Bootstrap note:** this decision itself lands `gated` in its own
   PR — the first artifact under the mechanic it establishes — and is
   bumped to `approved` after merge.

## Consequences

- `ratify-guard` updated: still fails a ready PR touching a
  `status: draft` decision/spec (message now says flip to `gated`);
  additionally fails a PR that **adds** a new decision/spec already
  `status: approved` (in-diff self-approval). Post-merge bump PRs —
  which *modify* an existing artifact to `approved` — pass.
- CLAUDE.md §Operating method: the Statuses line and the
  ratification-rides-the-merge bullet rewritten to the family enum and
  mechanic.
- `.trellis/profile.md` gains the family's "Lifecycle mapping" section —
  this was the one family repo without it (found during the grove
  install).
- `decision-0022` and `decision-0037` carry
  `superseded_in_part_by: [decision-0042]` with forward notes; their
  untouched parts remain in force.
- The corpus-reviewer instance (grove install, in flight) reads the
  enum from the profile's mapping, which post-0042 is identical across
  the family.
