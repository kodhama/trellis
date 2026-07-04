---
id: invariants-v1
type: invariant-set
status: ratified
depends_on: [research-0002, decision-0008]
owner: gundi
supersedes: invariants-v0
ratified: 2026-06-29
---

# Trellis's Invariants — our synthesis, v1 (ratified 2026-06-29)

> **Provenance & honesty (load-bearing).** Still **our synthesis** — *not* externally
> attributed; do not name it authoritatively (the §4 "Keel's invariants" canary). v1 revises
> v0 using **Step 1 evidence** (`research-0002` — gate-testing Spec Kit, Kiro, BMAD, OpenSpec,
> SpecSwarm) and the **enforcement reframe** (`decision-0008`). **Ratified 2026-06-29 by the
> maintainer** — this is now the current-truth invariant set; the spine and all machinery
> consume *this*. It is the **compiled current-truth spec** (revise-in-place); its change
> history and rationale live in ADRs (`decision-0014`), not inline. `invariants-v0` is superseded.

> **What v1 changes, in one breath:** v0 was a flat list of 9. v1 splits it into a small
> **structural gate** (what a methodology must *have the shape for*), a **configurable
> operating layer** (what Trellis supplies on top), **two dials** (how strict, who enforces),
> and **two floors** (the two things never configurable to "off"). The driving evidence:
> real frameworks have the *structure* but not the *enforcement* — so enforcement is a dial,
> and the true floor is **surfacing**, not enforcing.

---

## The model

- **A. Structural invariants — the admission gate (`methodology`).** Properties a target
  methodology must be *shaped to allow*. Checked at ingestion (`decision-0003`); if absent,
  Trellis fails loudly — out of contract.
- **B. Operating invariants — what Trellis supplies (`trellis-design`).** Guarantees a
  Trellis-assisted project gets *because* it adopted Trellis. Not admission criteria.
- **C. Two configuration dials.** Per gate: *how strict* (enforcement strength) and *who
  enforces* (gatekeeper identity). This is what keeps Trellis buyer-neutral (`decision-0004`).
- **D. Two floors.** The only things that can never be configured off.

**Cross-cutting theme — *drift is allowed, but never silent* (`decision-0020`).** Several invariants
are one idea seen from different angles: directional-flow + graph-maintenance keep *artifacts* from
drifting, self-improvement makes evolution *deliberate rather than drift*, and **D1 (transparency)** is
the floor that makes all of it *never silent* — **including divergence from a framework you claim to
follow** (the case the retired B8 named, now a D1 example, `decision-0021`). It is D1 generalized
across every kind of drift — not a separate rule (minimal-first), a lens on the set. *(This theme is
why B8 collapsed: a "no silent drift from your framework" rule adds no mechanism D1 lacks.)*

Durability tags carry forward (`durable` / `strong, less settled`); tags are claims to be
falsified. The **C/D additions are tagged `provisional`** — they come from a single round of
evidence (`research-0002`) plus a fresh design hypothesis (`0008`), and need a second
instance to settle.

---

## Identifiers (stable slugs) — `decision-0013`

Each invariant has a **stable `slug`** — its *canonical identifier*. The `A/B/C/D`+number is a
**display label**: convenient to say, **frozen** for existing invariants (never renumbered),
but **not** a reference. **References use slugs.** When invariants merge, the absorbed slug is
`superseded_by` the survivor's, so any old reference — *including the ordinal* — resolves
through this registry (same historical-reference exemption as artifacts; append-only decisions
are never edited to chase a rename). This retires the tombstone hack: a merge is now a proper
slug-supersede, not an ordinal gap.

| Slug | Label | Note |
|---|---|---|
| `inv-directional-flow` | A1 | |
| `inv-handover-points` | A2 | |
| `inv-intent-locus` | A3 | |
| `inv-ratifiable-artifacts` | A4 | |
| `inv-graph-maintenance` | B1 | absorbed the backprop reflex; `inv-self-improvement` **restored to B6** (`decision-0018`) |
| `inv-gate-at-handover` | B2 | |
| `inv-independent-judgment` | B3 | absorbed `inv-epistemic-integrity` (intent face) |
| `inv-auditable-archive` | B4 | |
| `inv-bounded-context` | B5 | |
| `inv-self-improvement` | B6 | **restored, first-class** (`decision-0018`); neighbor to B1, shares `inv-prune-bias` |
| `inv-minimal-first` | B7 | |
| `inv-reference-relationship` | B8 | **collapsed → `floor-transparency` (D1) + `decision-0002` dial** (`decision-0021`); id resolves to D1 |
| `inv-clarify-before-commit` | B9 | |
| `dial-enforcement-strength` | C1 | |
| `dial-gatekeeper` | C2 | |
| `floor-transparency` | D1 | absorbed `inv-bounded-correction` (escalate-don't-abandon) + `inv-reference-relationship` (framework-divergence, `decision-0021`) |
| `floor-intent-gate` | D2 | |

**Retired artifact ids** (file deleted; the id resolves to its successor via this registry —
the historical-reference exemption lets append-only decisions keep the old id, `decision-0014`):

| Retired id | → Successor |
|---|---|
| `invariants-v0` | `invariants-v1` |

## A. Structural invariants — the admission gate (`methodology`)

*Small by design. A methodology that has these shapes can be supervised; Trellis supplies the
rest. Validated: Spec Kit, Kiro, BMAD, OpenSpec, SpecSwarm all clear A1–A3; Cursor Rules
fails A1/A2 (pure config, no flow) — the negative control that proves the gate discriminates;
pattern-level guidance (Claude Agent SDK) clears it if it carries the shape, not only if it
names stages.*

- **A1. Directional flow exists** — *durable.* One-way stages of **decreasing ambiguity**
  (research → decisions → contracts → implementation → validation). *Named or unnamed; rigid
  or pattern-level — what matters is the one-way shape, not fixed step names* (refinement from
  the Cursor↔Agent-SDK contrast, `research-0001`).
- **A2. Handover points exist** — *durable.* Defined transitions between stages, each a place
  where a gate **can** attach. (Whether a gate *is* enforced there is layer B + dial C.)
- **A3. Intent locus** — *durable.* Humans own intent/values *somewhere identifiable*. A
  process with no human intent locus is not targetable for accountable development.
- **A4. Ratifiable, checkable artifacts** — *strong, less settled.* Upstream can reach an
  **approved** state that downstream consumes, and outputs can be **checked against** it. This
  is the structural prerequisite that lets B1 (flow enforcement), B3 (verification) and B4
  (archive) have something to act on. *(Open: does this over-constrain pattern-level methods
  with no explicit "approved" state? — see open questions.)*

---

## B. Operating invariants — what Trellis supplies (`trellis-design`)

- **B1. Directional-graph maintenance** — *durable.* The dependency graph of **artifacts and
  rules** is kept **consistent and minimal**, information flowing one way (decreasing
  ambiguity). Maintenance is trigger-driven (not vigilance), rides existing rituals, stays
  subordinate to the work, and proceeds by **surfaced suggestions the human rules on** (D2);
  **append-only** records are superseded, never edited-in-substance or deleted (B4), while
  revise-in-place/derived artifacts change directly. Four operations:
  - **Flow:** downstream consumes only **ratified** upstream, never drafts (strength = dial
    C1). *(e.g. implementation reads an approved spec, not a draft.)*
  - **Forward — re-propagate on change:** when an upstream changes, its **direct dependents**
    are flagged for re-review; recursion emerges (each review may flag *its* dependents).
    *(e.g. a repaired ADR → re-review its specs → plans → tests → code, in turn.)* *(Full
    content-consistency enforcement is the deferred conformance-to-upstream check, `spec-0001`.)*
  - **Backward — repair (backprop):** downstream reveals an upstream is missing/incomplete/
    contradicted → suggest **updating or creating** it, never a silent downstream patch, so
    downstream never forks from upstream. *(e.g. build uncovers a missing architectural
    decision → create the ADR; undocumented infra → a retroactive ADR; a too-abstract rule →
    add a grounding example.)*
  - **Prune:** retire an upstream that outlives what it governs; **bias toward retiring over
    adding**. *(e.g. a rule whose trigger is gone; a spec/doc/config whose referent was removed.)*
  *Absorbs former B1 (flow) + B6 (self-improvement) + the backprop candidate + forward
  re-propagation — the four operations of keeping the graph true and minimal.*
- **B2. Enforce a gate at each handover** — *durable.* Apply the verification gate at every
  A2 handover. *Real frameworks leave gates skippable (Kiro Quick Plan, Spec Kit lean path);
  Trellis makes the gate real* — at the strictness of dial C1, and **any skip is surfaced**
  (floor D1).
- **B3. Independent judgment — the assessor is independent of what it assesses** — *durable.*
  One principle, two faces:
  - **Conformance face — "the builder does not grade itself":** whatever produces work is
    never its sole judge; the verifier ≠ the producer and derives its own checklist from the
    approved upstream. *Reclassified from a `methodology` gate (v0-5) to `trellis-design`* —
    Step 1 showed spec-driven tools *lack* it, so Trellis *supplies* it. *Proven implementable:
    SpecSwarm's fresh-context adversarial `spec-mentor` (`research-0002`).* Gatekeeper by C2.
  - **Intent face — the agent does not flatter the human** *(merged from B11):* the agent's
    assessments **track the evidence, not the human's preferences** — disagreement and risks
    stated plainly when analysis warrants, affirmation only when the data supports it
    (*withholding warranted positive signal distorts as much as manufacturing praise*), and
    no performed disagreement to look rigorous. This is what makes the **intent gate (D2)
    real** rather than decorative: a sycophantic gatekeeper cannot catch a wrong upstream.
    *Weakly checkable — lives in sub-agent design + system prompts, not a mechanical gate.*
- **B4. Auditable archive** — *durable.* Provenance + immutable decision history + consolidated
  current-truth (the v0 3+7 merge, retained). *OpenSpec's change/delta/archive model is the
  field's best instance.*
- **B5. Bounded context** — *durable.* Each operation reads only its declared inputs, never
  the whole archive.
- **B6. Self-improvement — the process learns from friction and gets better** — *provisional*
  (**restored to first-class**, `decision-0018`; the earlier B6→B1 merge kept the prune mechanic but
  lost the *"evolve"* pillar — brief invariant 9 / Pillar II). *One principle, two faces — like B3:*
  - **Checkable floor — `inv-propagation-surfaced` (SI-1):** the improvement signals a change or a
    work session throws off are **surfaced through the project's chosen channel** — asked (B9) or
    inferred (Assess), **never assumed** — never silently dropped, with any retirement/update in the
    same change. *Surfacing is the floor (an application of D1); silence is the only violation.*
  - **Dispositional face** *(weakly checkable, lives in sub-agent design + system prompts, as B3's
    intent face does):* the agent **proactively notices** a signal and proposes *both* the fix *and* a
    standing trigger, inferring the channel from context, **asking not acting**.
  - Carries **`inv-ride-existing-rituals` (SI-2)** — surfacing rides the work you already do, never a
    separate ceremony — and **`inv-prune-bias` (SI-3)** — retire over add; the trigger set never grows
    monotonically (the **shared hinge with B1**). *(e.g. a pipeline failing on the same step with no
    correction; a PR raising open questions with no follow-up; a rule that hasn't fired → retire it.)*
  **Neighbor to B1 (`inv-graph-maintenance`):** B1 keeps the graph *true* (referential integrity); B6
  makes the process *better* (adaptation). *Backported from math-quest, which kept these
  trellis-compatible; concrete engine in `decision-0018`.*
- **B7. Minimal-first** — *strong, less settled.* Smallest process that works; add a step
  only when friction reveals the boundary. *(v0's "reference-not-adoption" split out to B8 —
  strict single-framework adoption is legitimate, so "never inherit wholesale" was too strong.)*
- **B8 (`inv-reference-relationship`) → collapsed** (`decision-0021`). It had **no mechanism of its
  own**: the adopt-vs-adapt *free choice* is the **`decision-0002`** dial; *recorded* is B4; *surfaced,
  never silent drift* is **D1** / the no-silent-drift theme. So it was D1 + B4 + the dial applied to
  the object "external frameworks." Retired; the *divergence-from-a-framework* case is a **D1** example
  (above), the adopt/adapt choice stays **`decision-0002`**'s dial. Any reference to B8 resolves to D1
  via the registry — no decision is edited to chase the collapse.
- **B9. Clarify before commit** — *strong (new, from framework analysis).* Ambiguity in an
  upstream artifact is actively **surfaced and resolved** — usually by asking the human —
  before downstream consumes it; ambiguity is never silently resolved by guessing.
  *Near-universal: Spec Kit `/clarify`, Kiro ambiguity/gap analysis pre-code, SpecSwarm
  clarification + `/ss:decisions`.* Arguably the most central uncertainty-reduction act, and
  absent from v0.
*(Two earlier candidates were retired during v1 drafting — a bounded-correction invariant (its
durable half absorbed into D1) and an epistemic-integrity invariant (merged into B3's intent
face). Minimal-first correcting a one-session overshoot.)*

---

## C. The two configuration dials (`decision-0008`) — *provisional*

Per gate, two settings — the mechanism that lets one invariant structure serve both a
speed-first startup and an assurance-first enterprise (`decision-0004`):

- **C1. Enforcement strength:** `expressed` (documented only) → `default-on-but-skippable`
  → `enforced`. Trellis can move a methodology's expressed structure toward enforced; that
  strictness is **opt-in**, never forced.
- **C2. Gatekeeper identity:** `independent-agent | human | none`. Who applies B3's check at
  this gate. `none` is permitted **only when the skip is surfaced** (floor D1), and **never**
  at the intent gate (floor D2).

The dials are configuration, *not* invariants — but the *existence* of the dials (that
strictness and gatekeeper are choices, surfaced and recorded) is the on-thesis commitment.

---

## D. The floors — never configurable to "off" — *provisional*

- **D1. Transparency over silent action** — *the candidate hard floor.* Every consequential
  choice is **surfaced**: a skipped gate, a missing capability, a degraded result, a relaxed
  setting. **Generalizes v0-7 (loud failure) to also cover the conscious skip** (`0008`):
  Trellis may *allow* skipping, but never *silently*. Likewise a failed verification is
  **escalated visibly, never silently abandoned** (the durable half of the former B10;
  bounded retry-before-escalation is an operating *practice*, not an invariant). This is
  plausibly **the sharpest one-line statement of Trellis's value — "surface the choice," not
  "enforce the choice."**
- **D2. The intent gate never fully opens** — *durable* (v0-4 core). At the intent locus (A3),
  C2 can never be `none`: a human (or, by ratchet, an independent check the human authorized)
  is mandatory. It is the only place an upstream that is itself *wrong* gets caught. The one
  gate strictness can never dial to zero.

---

## Acceptance criteria

- The **admission gate (A) is small** (4 structural properties) and is the *only* set
  `decision-0003`'s ingestion check uses.
- Each operating invariant (B) is something Trellis *supplies*, not something the methodology
  must already have.
- Strictness and gatekeeper are **dials with surfaced defaults**, not hard-coded — so the
  same set serves startup and enterprise (`decision-0004`).
- The two floors are stated as non-configurable, with D2 the recognized exception to C2.
- **Every invariant carries ≥1 concrete example** (few-shot), especially abstract ones — the
  iron rule (brief §7) applied to the set itself: *a rule you can't exemplify is probably
  vaporware.* Name the application instance where useful.
- **This file is the compiled current-truth spec** (revise-in-place, `ratified`, shippable);
  *significant* invariant changes are recorded as ADRs (`decision-0014`), not inline — B4
  applied to the set itself. It carries rationale by *reference* (links to governing ADRs).

## Open questions

- **B3 positive control (friction-logged 2026-06-29, from our own CI-reviewer verification):**
  should independent verification require the verifier be *demonstrably able to fail* —
  validated against a known-bad **positive control** — not merely that a verifier *exists*? A
  checker that can't be shown to reject anything is a decorative gate (the D2/sycophancy
  disease one level down). Surfaced when our own verification loop read a null result from a
  test with no discriminating power. Decide when building the spine's verification machinery.
- **Is D1 (transparency) its own floor, or still part of loud-failure (B-layer)?** v1 elevated
  it on the strength of `0008`; ratification should confirm or fold it back.
- **Does A4 over-constrain pattern-level methodologies** (e.g. Agent SDK) that have no explicit
  "approved" state — or is a loose/implicit ratification enough to clear the gate?
- **Provenance (B4): trellis-design or sometimes structural?** OpenSpec shows it can be a
  framework's *strength* — the v0 contested call, still open.
- **The dials (C) need a second instance.** They are `provisional` from one evidence round;
  validate against instance #2 (still the N=1 risk, `decision-0001`).
- **Tier-2 boundary confirmations** to fold in fully: Devin (open-ended; merge-time human gate
  = partial D2/A3, fails A1?), Cursor (fails A1/A2 — confirm as the clean negative control),
  Claude Agent SDK (passes A as pattern-with-guardrails).
- **Does any operating invariant (B) collapse or graduate to a dial?** Minimal-first (B7)
  applied to v1 itself.
- **Minimal-first, applied to v1 itself (resolved this round):** B11 merged into B3 (intent
  face); B10 dropped (durable half into D1). Operating layer trimmed back **11 → 9**; the set
  is now A1–A4 + B1–B9 + dials + floors.
- **Adherence (`decision-0002`):** working stance is two coarse modes — *adopt* (one
  framework) vs *adapt* (synthesize from several); deliberately **not** formalized into the
  invariant model until instance #2 can test it. v1 encodes only the durable part (B8).
- *(Data-dependent, parked for instance #2 — not resolvable by wording):* does A4
  over-constrain pattern-level methods? is provenance (B4) a gate property? do the dials (C)
  hold across a real second project?
