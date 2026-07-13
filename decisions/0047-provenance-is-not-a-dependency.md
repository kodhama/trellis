---
id: decision-0047
type: decision
status: draft
depends_on: [invariants-v1, signature-catalog-v1, spec-0001, decision-0037]
owner: agent
date: 2026-07-13
---

> Shaped interactively with the maintainer (2026-07-13), resolving the
> principle half of trellis#148. The maintainer chose to **mandate the
> split** (provenance is categorically not a dependency) over staying
> permissive or grandfathering — after the layering re-scope established
> that only the mechanism-free principle is trellis's; the concrete
> relation + grammar is grove's operating model (grove/adr-0010's
> versioning precedent). Two independent adversary rounds preceded the
> intent gate: R1 (NEEDS-REVISION — check-7 misuse, grounding overclaim),
> R2 (NEEDS-REVISION — the dependency definition must be coupling-based
> not drift-based, to admit check-7's frozen pin as a third category; and
> the contract needs a marking-class citation, not zero touch). Both
> folded below.

# 0047 — provenance is not a dependency; the dependency edge means genuine coupling

## Context

Two of trellis's enforced invariants act on the **dependency edge** and do
their drift-catching job cleanly only if that edge denotes genuine coupling:

- **`inv-directional-flow`** (enforced) — *"no ratified artifact cites a
  draft upstream"* / *"a ratified doc never depends on a draft."*
- **`inv-graph-maintenance`** (enforced) — *"when you change something,
  update everything that depends on it"*; a change to an upstream
  **surfaces its dependents**.

Both treat the dependency edge as **coupling**: the consumer's correctness
is contingent on the upstream. But the `spec-0001` §1 schema row types it
only as shape — *"list of `id`s and/or declared external refs"* — and says
nothing about what an edge *means*. So a source that merely **informed** an
artifact, without its correctness being contingent on that source, has
nowhere honest to sit: it goes into `depends_on` (which couples it) or a
succession pointer (which records history), neither of which fits "evidence
used at a point in time that can age without the consumer drifting."

The cost is real and observed (trellis#148, math-quest):

- **Diagnostic signature.** math-quest's ADRs are append-only/frozen; when a
  cited discovery doc is later revised, the ADR does **not** drift, and its
  correctness does not hinge on that doc. That proves those edges are
  provenance, not coupling — they are in `depends_on` only because there is
  no alternative. *(This is the evidence that the two relationships are
  distinct; it is not itself what this decision remedies — frozen artifacts
  cannot be edited. See Consequence 4.)*
- **The forward cost — the gating burden on live citations.** Because
  provenance sits in `depends_on`, *enforced* directional-flow forces every
  referenced research/feedback doc to be promoted to `gated`/`approved`
  **just to be citable** — machinery spent to keep a non-coupling from
  tripping a coupling rule. This is what the mandate removes going forward.
- **Polluted graph.** "What would force this artifact to change?" cannot be
  answered cleanly when evidence edges are mixed with real constraints.

## Decision

**The dependency relation means genuine coupling, and nothing else.** An
artifact declares as a dependency only a source its correctness **is or was
contingent on** — a source that, had it been different, would have made the
artifact different or wrong. A source that **informed construction without
its correctness being contingent on it** (research/discovery evidence, a
feedback artifact, an external reference used at a point in time) is
**provenance**, a categorically distinct relationship, and is **not**
modeled as a dependency.

Coupling comes in two forms, both genuine dependencies — the distinction is
*liveness*, not *kind*:

- a **live** coupling drift-bears: an upstream change obligates re-checking
  the consumer (`inv-graph-maintenance`);
- a **frozen** coupling is a genuine dependency recorded at ratification and
  then immutable — an append-only decision that built on `upstream@vN`. It
  no longer drifts (it is pinned and the consumer is frozen), but the
  consumer's correctness *was* contingent on it. This is exactly what
  `spec-0001` §3 check 7 already exempts as *"a historical fact, not
  current-truth consumption."*

**Provenance is neither** — the consumer's correctness was never contingent
on it. That is the line this decision draws, and check 7's frozen pin sits
firmly on the dependency side of it (a frozen coupling), never the
provenance side.

This is the enforced reading, family-wide: **using the dependency edge to
carry provenance is non-conformant** for new and live artifacts.

**This is a chosen reading on cost grounds, not a logical deduction.** The
invariants establish that the dependency edge is coupling; they do not by
themselves *forbid* overloading it with provenance — a permissive
`depends_on` could coexist with them, at the standing gating burden and a
polluted graph above. The maintainer chose to forbid the overload
(trellis#148, 2026-07-13). This decision records that choice and sharpens
the two invariants to match it.

**Mechanism-free by construction.** It names **no relation grammar** — no
field name for the provenance relation, no spelling, no schema row. Whether
provenance is spelled `cites` or otherwise, its frontmatter form, and how
the conformance agents check it, are **operating-model** concerns homed in
grove (`grove/adr-0010`: the *principle* is trellis's, the *grammar* is
grove's, reaching consumers via the installed operating model — never a new
relation row in trellis's spine contract). What trellis *does* record is a
**marking-class citation**, not a schema amendment — see Consequence 1.

## What this does and does not touch in `spec-0001`

- **Does NOT** add a provenance-relation row, field, or grammar to §1, and
  does **not** rewrite §3 check 7 or the §1 `depends_on` row's text (a new
  relation's grammar is grove's; the frozen-coupling reading makes check 7
  coherent as written).
- **DOES** require a **marking-class citation** (Consequence 1): the §1
  `depends_on` row gains an inline scope pointer to `decision-0047` (exactly
  as `status` cites `decision-0037` and `version` cites `grove/adr-0010` at
  their rows), and `spec-0001`'s frontmatter `depends_on` gains
  `decision-0047` (the `grove/adr-0010` de-reflection precedent, which added
  itself to `spec-0001`'s `depends_on`). This is the permitted marking-class
  touch on a ratified artifact — a citation, not an in-substance edit — and
  it is what keeps the narrowing visible at its point of use rather than
  living only in this decision.
- **Does NOT** rewrite history: already-frozen append-only artifacts keep
  their existing edges (Consequence 4).

## Consequences (execution — downstream, deferred)

1. **Trellis marking-class pass (`contract-author`).** Add the
   `decision-0047` scope citation to the §1 `depends_on` row ("genuine
   coupling; provenance is a distinct relationship") and `decision-0047` to
   `spec-0001`'s frontmatter `depends_on`. Marking-class, not a schema
   amendment — matches spec-0001's own per-term citation convention. Check 7
   needs no text edit: under the coupling definition its frozen pin is a
   frozen dependency, and its "historical fact" language is coherent.
2. **Optional catalog sharpening (maintainer's call, Open Q).** A
   mechanism-free clarifying line on `inv-graph-maintenance` /
   `inv-directional-flow` stating the converse (the dependency edge is
   coupling; provenance is distinct), naming no grammar.
3. **Grove (the encoding — triggers a grove shaping issue).** The concrete
   provenance relation: its name, its frontmatter grammar (a non-flow,
   non-drift forward-pointer — the exact class/spelling is grove's to decide,
   offered as shaping input, not fixed here), a relations entry on the
   `.grove/` companion axis, and the agent-duty edits (`shaper` records
   evidence as provenance not dependency; `corpus-reviewer` /
   `conformance-reviewer` type it non-flow). **Naming snag to hand grove:**
   `inv-directional-flow`'s own text uses "cites" to mean the *flow* edge —
   so the provenance relation likely should **not** be spelled `cites`.
4. **Consumers (triggers a consumer audit, downstream of grove's encoding).**
   *Live / revise-in-place* artifacts migrate provenance edges out of the
   dependency relation into the new one. *Already-frozen append-only*
   artifacts are **exempt — forced by append-only itself:** they cannot be
   edited to migrate an edge without violating the no-in-place-edit rule, so
   their existing edges stand as frozen history. (An append-only
   consequence, **not** the "grandfather to save effort" option the
   maintainer declined — the mandate binds every new and live artifact
   without exception.)

Sequencing (corrected home vs trellis#148): **this trellis decision → grove
shaping issue (the grammar) → consumer audit.** No grove charter change and
no migration until this ratifies.

## Open questions (parked, ≤3)

- **Marking form (for the intent gate):** confirm the Consequence-1
  marking-class approach — a `decision-0047` scope citation on the §1 row +
  frontmatter `depends_on`, check 7 left as coherent-in-place — vs. a fuller
  check-7 text rewrite. The marking-class reading is the recommendation and
  matches spec-0001's convention.
- **Frozen-artifact reach:** does the audit leave frozen append-only
  provenance edges silent-as-history, or give them a one-time marking? A
  reader distinguishes a legacy edge from a fresh violation only by the
  ratification-vs-`decision-0047` date, so a pure silent-as-history option
  must preserve that date discriminator, never obscure it. Deferred to the
  audit.
- **Catalog sharpening:** add the mechanism-free clarifying line to the
  invariant catalog now (Consequence 2), or let the decision record stand?

## Self-check (gate)

Mechanism-free on grammar: names no provenance-relation row/field/spelling —
that is grove's (`grove/adr-0010` layering). Distinguishes the **forbidden**
schema amendment (a new relation row — not done) from the **permitted**
marking-class citation (a scope pointer on the §1 row + frontmatter
`depends_on`, matching the `status`→`decision-0037` / `version`→`adr-0010`
convention — done, Consequence 1), correcting R2's finding that "zero touch"
overclaimed. The dependency definition is **coupling-based** (correctness
contingent), not drift-based, so check 7's frozen version-pin is a third-way
*frozen coupling* — a genuine dependency, not provenance — resolving R2's
crux without editing check 7. Grounded in two quoted, enforced invariants
(`signature-catalog-v1`), framed honestly as a **chosen cost-tradeoff, not a
deduction** (the "coherent only if" entailment language R1/R2 flagged is
gone). Depends only on ratified artifacts (invariants-v1,
signature-catalog-v1, spec-0001, decision-0037 — `depends_on` is a fixed
structural field, not one of 0037's open methodology-defined fields, so
mandating its semantics does not conflict), no draft consumed. Execution
scoped downstream (grammar and migration are not this decision's to
perform). Promote `draft → gated` on self-check; **`approved` is the
maintainer's intent act** (the procedural approval mechanic, decision-0046),
not this author's.
