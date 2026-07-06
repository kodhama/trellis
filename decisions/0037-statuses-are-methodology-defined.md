---
id: decision-0037
type: decision
status: ratified
depends_on: [spec-0001, invariants-v1, decision-0003, research-0003]
owner: gundi
ratified: 2026-07-06
---

# 0037 — Statuses are methodology-defined; the contract requires a ratifiable *shape*, not an enum

## Context

`spec-0001` §1 makes `type` an **open, methodology-defined field** (`research-0003`: a
methodology brings its own type names; Trellis ships a soft seed) — but the very next row
hard-codes the `status` lifecycle as a closed enum: *"v0: `draft` → `ratified`
(+ `superseded`)"*. That is an asymmetry with no principle behind it: the invariants don't
need *these* status names any more than they need *our* type names. What
`inv-ratifiable-artifacts` actually requires (quoting `invariants-v1`): *"upstream can reach
an **approved** state that downstream consumes, and outputs can be **checked against** it"* —
a **shape**, not a vocabulary.

**Instance #1 already runs a different enum, and it works.** math-quest — the source instance
the invariants were extracted from (`decision-0009`) — runs `draft → gated → approved
(→ superseded)`, defined in its artifact contract (its `CLAUDE.md`, verbatim):

> "**Never consume a `draft` artifact downstream.** Only `gated` (self-checked against its
> rubric) or `approved` (human-merged) artifacts are valid inputs."
>
> "To promote `draft → gated`, run the artifact through its rubric and append a
> `## Rubric check` section with the result."
>
> "`approved` happens only by PR merge. Do not set it yourself."

`gated` is a rubric-self-checked, **agent-consumable** state that sits *before* human merge —
legitimate under a **recorded ratchet** (its ADR-0009: the maintainer is *"an adjudicator, not
a gatekeeper"*; execution gates run agent-autonomous, and the human gate *"loosens toward full
autonomy only on a recorded track record — a ratchet, reversible on any miss"*). That is
exactly the ratchet `floor-intent-gate` already allows (*"a human, or, by ratchet, an
independent check the human authorized"*). And on the question `spec-0001` §2 explicitly
deferred — an execution-layer `approved`: *"a third document status, or a gate-outcome on a
change rather than a status?"* — math-quest's answer in practice is **a gate-outcome**: its
conformance gate lands as a PR verdict (the `conformance-reviewer` run), not as a document
status. A working third status *and* a working gate-outcome answer, in one instance neither
of which our closed enum can describe.

## Decision

1. **The concrete status enum is methodology-defined, like types.** The artifact contract
   (`spec-0001` §2) requires of any methodology's lifecycle a **shape**, not names:
   - a **working state** downstream may not consume;
   - at least one **ratifiable state** — consumable by downstream, reachable only via
     **defined promotions** (this is what `inv-ratifiable-artifacts` needs to act on);
   - **the intent gate holds**: some ratified state is a human act (or a human-authorized
     ratchet, recorded) — opening the enum never opens `floor-intent-gate`;
   - **supersession is expressible**;
   - the methodology **declares** its enum and promotion rules, so the conformance check
     verifies against the declaration — an undeclared status is a conformance failure, and
     a lifecycle *without* the shape above fails the admission gate loudly.
2. **Trellis's own enum stays as the default / reference expression:** `draft → ratified
   (→ superseded)` — what this repo runs and what setup composes onto a project that brings
   no lifecycle of its own.
3. **`owner:` opens the same way, but the *role* is contract.** The contract keeps an
   accountable-human role per artifact — that is `inv-intent-locus` (A3), not vocabulary —
   and `spec-0001`'s `owner` field carries it by default, with **`author` optional** for
   authorship. A methodology may **map fields**: math-quest's corpus carries `owner: agent`
   meaning *authorship* (its accountable human is the maintainer, holding the merge gate) —
   legitimate, provided the mapping is **declared** (which field or mechanism carries the
   accountable human). An undeclared divergence is silent drift (`floor-transparency`).

## Consequences

- `spec-0001` §1 (`status`, `owner` rows) and §2 amended **in place** (it is revise-in-place
  current-truth; this decision is its rationale). The §2 deferred-`approved` note now cites
  the math-quest evidence instead of standing purely open.
- `core/rubrics/artifact-contract.md` check 2 and the `conformance-reviewer` check 2 verify
  `status` against the **methodology's declared lifecycle** (for this repo:
  `{draft, ratified, superseded}`) rather than a universal enum.
- Assess/Apply (cluster 1) gain a job: capture the host's lifecycle declaration (enum +
  promotions + the accountable-human mapping) in the expression profile — a candidate
  `spec-0002` field, owed to that build.

## Open questions

- Where the lifecycle declaration lives for a supervised project — an expression-profile
  field (leaning), or a standalone declaration artifact? Decide when Assess is specified.
- Does the promotion *evidence* (math-quest's `## Rubric check` section) deserve contract
  status — must a promotion be checkable, or only defined? Deferred until a second instance
  shows whether self-check-evidence generalizes.

## Supersedes / superseded by

— (none; amends `spec-0001` in place — the spec is revise-in-place current-truth,
`decision-0014` pattern)
