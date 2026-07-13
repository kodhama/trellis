---
id: invariants-v1
type: invariant-set
status: ratified
depends_on: [decision-0008]
informed_by: [research-0002]
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

> *Amended in place 2026-07-13 (`decision-0047` + `grove/adr-0011`; consumer-audit
> marking-class). WHAT: `research-0002` moved out of frontmatter `depends_on` into a new
> `informed_by` list — the Step 1 gate-test evidence informed v1's revision without v1's own
> correctness being contingent on `research-0002` remaining unchanged; provenance, not
> coupling. `decision-0008` (the enforcement reframe) stays in `depends_on` — v1's structure
> is genuinely contingent on that reframe. No `version` counter on this artifact (it versions
> by id + supersession, `invariants-v0` → `invariants-v1`) to bump. POINTER: `decision-0047`
> Consequence 4, `grove/adr-0011`.*

> **What v1 changes, in one breath:** v0 was a flat list of 9. v1 splits it into a small
> **structural gate** (what a methodology must *have the shape for*), a **configurable
> operating layer** (what Trellis supplies on top), **two dials** (how strict, who enforces),
> and **two floors** (the two things never configurable to "off"). The driving evidence:
> real frameworks have the *structure* but not the *enforcement* — so enforcement is a dial,
> and the true floor is **surfacing**, not enforcing.

---

## The model

- **Structural invariants — the admission gate (`methodology`).** Properties a target
  methodology must be *shaped to allow*. Checked at ingestion (`decision-0003`); if absent,
  Trellis fails loudly — out of contract.
- **Operating invariants — what Trellis supplies (`trellis-design`).** Guarantees a
  Trellis-assisted project gets *because* it adopted Trellis. Not admission criteria.
- **The two configuration dials.** Per gate: *how strict* (enforcement strength) and *who
  enforces* (gatekeeper identity). This is what keeps Trellis buyer-neutral (`decision-0004`).
- **The two floors.** The only things that can never be configured off.

**Cross-cutting theme — *drift is allowed, but never silent* (`decision-0020`).** Several invariants
are one idea seen from different angles: directional-flow + graph-maintenance keep *artifacts* from
drifting, self-improvement makes evolution *deliberate rather than drift*, and **`floor-transparency`**
is the floor that makes all of it *never silent* — **including divergence from a framework you claim to
follow** (the case the retired `inv-reference-relationship` named, now a `floor-transparency` example,
`decision-0021`). It is transparency generalized across every kind of drift — not a separate rule
(minimal-first), a lens on the set. *(This theme is why `inv-reference-relationship` collapsed: a "no
silent drift from your framework" rule adds no mechanism `floor-transparency` lacks.)*

Durability tags carry forward (`durable` / `strong, less settled`); tags are claims to be
falsified. The **dial and floor additions are tagged `provisional`** — they come from a single round of
evidence (`research-0002`) plus a fresh design hypothesis (`0008`), and need a second
instance to settle.

---

## Identifiers (stable slugs) — `decision-0013`, `decision-0038`

Each invariant's **stable `slug`** is its **only name** — canonical for reference *and*
display (`decision-0038`). The `A/B/C/D`+number codes that used to serve as display labels
are **retired**: they were opaque to anyone outside the file (the maintainer included), and a
label no one can decode is not a label. They survive **only in the legacy map below**, so
append-only decisions and older artifacts that cite them still resolve — no citing artifact
is edited to chase the retirement (`decision-0013`'s historical-reference exemption stands).
When invariants merge, the absorbed slug is `superseded_by` the survivor's, so any old
reference — *including a legacy code* — resolves through this registry.

**Legacy code map (retired display labels → slugs):**

| Slug | Legacy code (retired) | Note |
|---|---|---|
| `inv-directional-flow` | A1 | |
| `inv-handover-points` | A2 | |
| `inv-intent-locus` | A3 | |
| `inv-ratifiable-artifacts` | A4 | |
| `inv-graph-maintenance` | B1 | absorbed the backprop reflex; `inv-self-improvement` **restored to first-class** (`decision-0018`) |
| `inv-gate-at-handover` | B2 | |
| `inv-independent-judgment` | B3 | absorbed `inv-epistemic-integrity` (intent face) |
| `inv-auditable-archive` | B4 | |
| `inv-bounded-context` | B5 | |
| `inv-self-improvement` | B6 | **restored, first-class** (`decision-0018`); neighbor to `inv-graph-maintenance`, shares `inv-prune-bias` |
| `inv-minimal-first` | B7 | |
| `inv-reference-relationship` | B8 | **collapsed → `floor-transparency` + `decision-0002` dial** (`decision-0021`); id resolves to `floor-transparency` |
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

## Structural invariants — the admission gate (`methodology`)

*Small by design. A methodology that has these shapes can be supervised; Trellis supplies the
rest. Validated: Spec Kit, Kiro, BMAD, OpenSpec, SpecSwarm all clear the flow/handover/intent
gate; Cursor Rules fails directional-flow and handover-points (pure config, no flow) — the
negative control that proves the gate discriminates; pattern-level guidance (Claude Agent SDK)
clears it if it carries the shape, not only if it names stages.*

- **`inv-directional-flow` — directional flow exists** — *durable.* One-way stages of
  **decreasing ambiguity** (research → decisions → contracts → implementation → validation).
  *Named or unnamed; rigid or pattern-level — what matters is the one-way shape, not fixed
  step names* (refinement from the Cursor↔Agent-SDK contrast, `research-0001`).
- **`inv-handover-points` — handover points exist** — *durable.* Defined transitions between
  stages, each a place where a gate **can** attach. (Whether a gate *is* enforced there is the
  operating layer + the dials.)
- **`inv-intent-locus` — intent locus** — *durable.* Humans own intent/values *somewhere
  identifiable*. A process with no human intent locus is not targetable for accountable
  development.
- **`inv-ratifiable-artifacts` — ratifiable, checkable artifacts** — *strong, less settled.*
  Upstream can reach an **approved** state that downstream consumes, and outputs can be
  **checked against** it. This is the structural prerequisite that lets graph-maintenance
  (flow enforcement), independent judgment (verification) and the auditable archive have
  something to act on. *(Open: does this over-constrain pattern-level methods with no explicit
  "approved" state? — see open questions.)*

---

## Operating invariants — what Trellis supplies (`trellis-design`)

- **`inv-graph-maintenance` — directional-graph maintenance** — *durable.* The dependency
  graph of **artifacts and rules** is kept **consistent and minimal**, information flowing one
  way (decreasing ambiguity). Maintenance is trigger-driven (not vigilance), rides existing
  rituals, stays subordinate to the work, and proceeds by **surfaced suggestions the human
  rules on** (`floor-intent-gate`); **append-only** records are superseded, never
  edited-in-substance or deleted (`inv-auditable-archive`), while revise-in-place/derived
  artifacts change directly. Four operations:
  - **Flow:** downstream consumes only **ratified** upstream, never drafts (strength =
    `dial-enforcement-strength`). *(e.g. implementation reads an approved spec, not a draft.)*
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
  *Absorbs v0's flow rule + self-improvement's prune mechanic + the backprop candidate +
  forward re-propagation — the four operations of keeping the graph true and minimal.*
- **`inv-gate-at-handover` — enforce a gate at each handover** — *durable.* Apply the
  verification gate at every handover point. *Real frameworks leave gates skippable (Kiro
  Quick Plan, Spec Kit lean path); Trellis makes the gate real* — at the strictness of
  `dial-enforcement-strength`, and **any skip is surfaced** (`floor-transparency`).
- **`inv-independent-judgment` — the assessor is independent of what it assesses** — *durable.*
  One principle, two faces:
  - **Conformance face — "the builder does not grade itself":** whatever produces work is
    never its sole judge; the verifier ≠ the producer and derives its own checklist from the
    approved upstream. *Reclassified from a `methodology` gate (v0-5) to `trellis-design`* —
    Step 1 showed spec-driven tools *lack* it, so Trellis *supplies* it. *Proven implementable:
    SpecSwarm's fresh-context adversarial `spec-mentor` (`research-0002`).* Gatekeeper by
    `dial-gatekeeper`.
  - **Intent face — the agent does not flatter the human** *(merged from the epistemic-
    integrity candidate):* the agent's assessments **track the evidence, not the human's
    preferences** — disagreement and risks stated plainly when analysis warrants, affirmation
    only when the data supports it (*withholding warranted positive signal distorts as much as
    manufacturing praise*), and no performed disagreement to look rigorous. This is what makes
    the **intent gate (`floor-intent-gate`) real** rather than decorative: a sycophantic
    gatekeeper cannot catch a wrong upstream. *Weakly checkable — lives in sub-agent design +
    system prompts, not a mechanical gate.*
- **`inv-auditable-archive` — auditable archive** — *durable.* Provenance + immutable decision
  history + consolidated current-truth (the v0 3+7 merge, retained). *OpenSpec's
  change/delta/archive model is the field's best instance.*
- **`inv-bounded-context` — bounded context** — *durable.* Each operation reads only its
  declared inputs, never the whole archive.
- **`inv-self-improvement` — the process learns from friction and gets better** — *provisional*
  (**restored to first-class**, `decision-0018`; the earlier merge into `inv-graph-maintenance`
  kept the prune mechanic but lost the *"evolve"* pillar — brief invariant 9 / Pillar II). *One
  principle, two faces — like `inv-independent-judgment`:*
  - **Checkable floor — `inv-propagation-surfaced` (SI-1):** the improvement signals a change or a
    work session throws off are **surfaced through the project's chosen channel** — asked
    (`inv-clarify-before-commit`) or inferred (Assess), **never assumed** — never silently
    dropped, with any retirement/update in the same change. *Surfacing is the floor (an
    application of `floor-transparency`); silence is the only violation.*
  - **Dispositional face** *(weakly checkable, lives in sub-agent design + system prompts, as
    `inv-independent-judgment`'s intent face does):* the agent **proactively notices** a signal
    and proposes *both* the fix *and* a standing trigger, inferring the channel from context,
    **asking not acting**.
  - Carries **`inv-ride-existing-rituals` (SI-2)** — surfacing rides the work you already do, never a
    separate ceremony — and **`inv-prune-bias` (SI-3)** — retire over add; the trigger set never grows
    monotonically (the **shared hinge with `inv-graph-maintenance`**). *(e.g. a pipeline failing on the
    same step with no correction; a PR raising open questions with no follow-up; a rule that hasn't
    fired → retire it.)*
  **Neighbor to `inv-graph-maintenance`:** graph-maintenance keeps the graph *true* (referential
  integrity); self-improvement makes the process *better* (adaptation). *Backported from math-quest,
  which kept these trellis-compatible; concrete engine in `decision-0018`.*
- **`inv-minimal-first` — minimal-first** — *strong, less settled.* Smallest process that
  works; add a step only when friction reveals the boundary. *(v0's "reference-not-adoption"
  split out to `inv-reference-relationship` — strict single-framework adoption is legitimate,
  so "never inherit wholesale" was too strong.)*
- **`inv-reference-relationship` → collapsed** (`decision-0021`). It had **no mechanism of its
  own**: the adopt-vs-adapt *free choice* is the **`decision-0002`** dial; *recorded* is
  `inv-auditable-archive`; *surfaced, never silent drift* is **`floor-transparency`** / the
  no-silent-drift theme. So it was transparency + archive + the dial applied to the object
  "external frameworks." Retired; the *divergence-from-a-framework* case is a
  **`floor-transparency`** example (above), the adopt/adapt choice stays **`decision-0002`**'s
  dial. Any reference to it resolves to `floor-transparency` via the registry — no decision is
  edited to chase the collapse.
- **`inv-clarify-before-commit` — clarify before commit** — *strong (new, from framework
  analysis).* Ambiguity in an upstream artifact is actively **surfaced and resolved** — usually
  by asking the human — before downstream consumes it; ambiguity is never silently resolved by
  guessing. *Near-universal: Spec Kit `/clarify`, Kiro ambiguity/gap analysis pre-code,
  SpecSwarm clarification + `/ss:decisions`.* Arguably the most central uncertainty-reduction
  act, and absent from v0.
*(Two earlier candidates were retired during v1 drafting — a bounded-correction invariant (its
durable half absorbed into `floor-transparency`) and an epistemic-integrity invariant (merged
into `inv-independent-judgment`'s intent face). Minimal-first correcting a one-session
overshoot.)*

---

## The two configuration dials (`decision-0008`) — *provisional*

Per gate, two settings — the mechanism that lets one invariant structure serve both a
speed-first startup and an assurance-first enterprise (`decision-0004`):

- **`dial-enforcement-strength` — enforcement strength:** `expressed` (documented only) →
  `default-on-but-skippable` → `enforced`. Trellis can move a methodology's expressed
  structure toward enforced; that strictness is **opt-in**, never forced.
- **`dial-gatekeeper` — gatekeeper identity:** `independent-agent | human | none`. Who applies
  `inv-independent-judgment`'s check at this gate. `none` is permitted **only when the skip is
  surfaced** (`floor-transparency`), and **never** at the intent gate (`floor-intent-gate`).

The dials are configuration, *not* invariants — but the *existence* of the dials (that
strictness and gatekeeper are choices, surfaced and recorded) is the on-thesis commitment.

---

## The floors — never configurable to "off" — *provisional*

- **`floor-transparency` — transparency over silent action** — *the candidate hard floor.*
  Every consequential choice is **surfaced**: a skipped gate, a missing capability, a degraded
  result, a relaxed setting. **Generalizes v0-7 (loud failure) to also cover the conscious
  skip** (`0008`): Trellis may *allow* skipping, but never *silently*. Likewise a failed
  verification is **escalated visibly, never silently abandoned** (the durable half of the
  former bounded-correction candidate; bounded retry-before-escalation is an operating
  *practice*, not an invariant). This is plausibly **the sharpest one-line statement of
  Trellis's value — "surface the choice," not "enforce the choice."**
- **`floor-intent-gate` — the intent gate never fully opens** — *durable* (v0-4 core). At the
  intent locus (`inv-intent-locus`), the gatekeeper can never be `none`: a human (or, by
  ratchet, an independent check the human authorized) is mandatory. It is the only place an
  upstream that is itself *wrong* gets caught. The one gate strictness can never dial to zero.

---

## Acceptance criteria

- The **admission gate is small** (4 structural properties) and is the *only* set
  `decision-0003`'s ingestion check uses.
- Each operating invariant is something Trellis *supplies*, not something the methodology
  must already have.
- Strictness and gatekeeper are **dials with surfaced defaults**, not hard-coded — so the
  same set serves startup and enterprise (`decision-0004`).
- The two floors are stated as non-configurable, with `floor-intent-gate` the recognized
  exception to `dial-gatekeeper`.
- **Every invariant carries ≥1 concrete example** (few-shot), especially abstract ones — the
  iron rule (brief §7) applied to the set itself: *a rule you can't exemplify is probably
  vaporware.* Name the application instance where useful.
- **This file is the compiled current-truth spec** (revise-in-place, `ratified`, shippable);
  *significant* invariant changes are recorded as ADRs (`decision-0014`), not inline —
  `inv-auditable-archive` applied to the set itself. It carries rationale by *reference*
  (links to governing ADRs).

## Open questions

- **Independent-judgment positive control (friction-logged 2026-06-29, from our own
  CI-reviewer verification):** should independent verification require the verifier be
  *demonstrably able to fail* — validated against a known-bad **positive control** — not
  merely that a verifier *exists*? A checker that can't be shown to reject anything is a
  decorative gate (the intent-gate/sycophancy disease one level down). Surfaced when our own
  verification loop read a null result from a test with no discriminating power. Decide when
  building the spine's verification machinery.
- **Is `floor-transparency` its own floor, or still part of loud-failure (operating layer)?**
  v1 elevated it on the strength of `0008`; ratification should confirm or fold it back.
- **Does `inv-ratifiable-artifacts` over-constrain pattern-level methodologies** (e.g. Agent
  SDK) that have no explicit "approved" state — or is a loose/implicit ratification enough to
  clear the gate?
- **Provenance (`inv-auditable-archive`): trellis-design or sometimes structural?** OpenSpec
  shows it can be a framework's *strength* — the v0 contested call, still open.
- **The dials need a second instance.** They are `provisional` from one evidence round;
  validate against instance #2 (still the N=1 risk, `decision-0001`).
- **Tier-2 boundary confirmations** to fold in fully: Devin (open-ended; merge-time human gate
  = partial intent-locus, fails directional-flow?), Cursor (fails directional-flow +
  handover-points — confirm as the clean negative control), Claude Agent SDK (passes the
  structural gate as pattern-with-guardrails).
- **Does any operating invariant collapse or graduate to a dial?** Minimal-first applied to
  v1 itself.
- **Minimal-first, applied to v1 itself (resolved this round):** the epistemic-integrity
  candidate merged into `inv-independent-judgment` (intent face); the bounded-correction
  candidate dropped (durable half into `floor-transparency`). Operating layer trimmed back
  **11 → 9**; the set is now 4 structural + 9 operating + dials + floors.
- **Adherence (`decision-0002`):** working stance is two coarse modes — *adopt* (one
  framework) vs *adapt* (synthesize from several); deliberately **not** formalized into the
  invariant model until instance #2 can test it. v1 encodes only the durable part
  (`inv-reference-relationship`, since collapsed).
- *(Data-dependent, parked for instance #2 — not resolvable by wording):* does
  `inv-ratifiable-artifacts` over-constrain pattern-level methods? is provenance a gate
  property? do the dials hold across a real second project?
