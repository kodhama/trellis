---
id: decision-0022
type: decision
status: ratified
depends_on: [invariants-v1, spec-0001, decision-0005, brief-§7]
owner: gundi
date: 2026-07-04
ratified: 2026-07-04
---

# 0022 — Merge = ratify: the ratified *state* is core, the ratification *workflow* is instance-specific

**Raised by:** the maintainer — the `draft → ratified` two-step tripped us up (drafts merged to `main`
and ratified in a later pass, leaving limbo — the catalog/profile OPEN LOOP, the #34 rework). *"I count
the merge as the ratification."*

## Context

The friction was blamed on the ratified state, but the split is cleaner than that:

- **The ratified *state* is core — it cannot be dropped.** A4 (an upstream reaches an *approved* state
  downstream builds against), D2 (a human ratifies at the intent gate), and B3 (the producer does not
  self-ratify, so an agent's output lands as `draft` until the human accepts it) all *require* it.
- **The *friction* was workflow, not the state.** Nothing in core says "merge a draft to `main`, then
  ratify later." Making ratification a step *separate* from merge is precisely a violation of
  **ride-existing-rituals** (`inv-ride-existing-rituals` / B1's "rides existing rituals";
  `decision-0007` already ties a gate to the PR). The lag is the cost of that separation, not of the
  lifecycle.

## Decision

1. **The ratified state is core; the ratification workflow is instance-specific** (`decision-0005`
   layering). Core requires only: an approvable state (A4), human ratification at the intent gate (D2),
   producer ≠ ratifier (B3). *How and when* the flip happens is the instance's call — a dial, not a
   mandate.
2. **Trellis-self convention — *merge = ratify.*** The `status: draft → ratified` flip is carried **in
   the PR's reviewed diff**; **merging the PR is the ratification** (the maintainer's merge = the D2
   approval). Ratification rides the merge, not a separate ceremony. **No draft is left un-ratified on
   `main` past the PR that introduced it** — either flip it (ratify at merge) or keep it clearly WIP.
3. **The "flip needs a model call somewhere" → put it in the PR, not at merge.** The agent proposes the
   flip when authoring the PR; the merge needs no automation. Heavier machinery (a merge hook that
   flips, or a check that flags a draft-merge that should have ratified) is **deferrable** — over-
   engineering at current scale (B7). Manual commit stays available for direct, agent-advised cases.

## Consequences

- **The convention goes in `CLAUDE.md`** (the instance operating method), under Gates.
- **The draft-limbo is retired:** `signature-catalog-v1` and `profile-trellis-self` get ratified **at
  their next merge** (increment 2, when the catalog reaches its final `why`/`honored`/`violated`
  form), not left hanging.
- **`spec-0001`'s open question** — *"two consumable states or one?"* — is answered *keep two*
  (`draft`/`ratified`), with the flip tied to merge; the collapse-to-one option is not taken.

## Open questions

- **Automated flip-on-merge** vs. the convention — revisit only if the manual flip becomes a burden.
- Does this default extend to a *supervised project's own* per-instance artifacts (their merge = their
  ratify), or is it Trellis-self only? Likely a sensible default, set per instance.

## Supersedes / superseded by

— (none; instance workflow under `spec-0001`'s lifecycle)
