---
id: invariants-v1
type: invariant-set
status: draft
depends_on: [invariants-v0, research-0002, decision-0008]
owner: gundi
supersedes: invariants-v0
---

# Bonsai's Invariants — our synthesis, v1 (draft)

> **Provenance & honesty (load-bearing).** Still **our synthesis** — *not* externally
> attributed; do not name it authoritatively (the §4 "Keel's invariants" canary). v1 revises
> v0 using **Step 1 evidence** (`research-0002` — gate-testing Spec Kit, Kiro, BMAD, OpenSpec,
> SpecSwarm) and the **enforcement reframe** (`decision-0008`). It is **draft**, pending
> ratification. v0 is kept (marked superseded) so the two can be diffed.

> **What v1 changes, in one breath:** v0 was a flat list of 9. v1 splits it into a small
> **structural gate** (what a methodology must *have the shape for*), a **configurable
> operating layer** (what Bonsai supplies on top), **two dials** (how strict, who enforces),
> and **two floors** (the two things never configurable to "off"). The driving evidence:
> real frameworks have the *structure* but not the *enforcement* — so enforcement is a dial,
> and the true floor is **surfacing**, not enforcing.

---

## The model

- **A. Structural invariants — the admission gate (`methodology`).** Properties a target
  methodology must be *shaped to allow*. Checked at ingestion (`decision-0003`); if absent,
  Bonsai fails loudly — out of contract.
- **B. Operating invariants — what Bonsai supplies (`bonsai-design`).** Guarantees a
  Bonsai-assisted project gets *because* it adopted Bonsai. Not admission criteria.
- **C. Two configuration dials.** Per gate: *how strict* (enforcement strength) and *who
  enforces* (gatekeeper identity). This is what keeps Bonsai buyer-neutral (`decision-0004`).
- **D. Two floors.** The only things that can never be configured off.

Durability tags carry forward (`durable` / `strong, less settled`); tags are claims to be
falsified. The **C/D additions are tagged `provisional`** — they come from a single round of
evidence (`research-0002`) plus a fresh design hypothesis (`0008`), and need a second
instance to settle.

---

## A. Structural invariants — the admission gate (`methodology`)

*Small by design. A methodology that has these shapes can be supervised; Bonsai supplies the
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

## B. Operating invariants — what Bonsai supplies (`bonsai-design`)

- **B1. Enforce directional flow** — *durable.* Downstream consumes only **ratified** upstream
  (never drafts). *The frameworks express A1 but rarely enforce this — Bonsai does.* Strength
  set by dial C1.
- **B2. Enforce a gate at each handover** — *durable.* Apply the verification gate at every
  A2 handover. *Real frameworks leave gates skippable (Kiro Quick Plan, Spec Kit lean path);
  Bonsai makes the gate real* — at the strictness of dial C1, and **any skip is surfaced**
  (floor D1).
- **B3. Independent verification — "the builder does not grade itself"** — *durable.*
  **Reclassified from a `methodology` gate (v0-5) to `bonsai-design`**, because Step 1 showed
  the spec-driven tools *lack* it (it is not something to admit *for*; it is something Bonsai
  *supplies*). When active, the verifier ≠ the producer and derives its own checklist from the
  approved upstream. *Proven implementable in an AI-native framework — SpecSwarm's
  fresh-context adversarial `spec-mentor` (`research-0002`).* Gatekeeper set by dial C2.
- **B4. Auditable archive** — *durable.* Provenance + immutable decision history + consolidated
  current-truth (the v0 3+7 merge, retained). *OpenSpec's change/delta/archive model is the
  field's best instance.*
- **B5. Bounded context** — *durable.* Each operation reads only its declared inputs, never
  the whole archive.
- **B6. Self-improvement is a disciplined loop** — *strong, less settled.* Trigger-driven,
  rides existing rituals, biased to retiring rules.
- **B7. Minimal-first** — *strong, less settled.* Smallest process that works; add a step
  only when friction reveals the boundary. *(v0's "reference-not-adoption" split out to B8 —
  strict single-framework adoption is legitimate, so "never inherit wholesale" was too strong.)*
- **B8. The reference relationship is explicit, recorded, and surfaced** — *provisional.* How
  much the operating model **adopts** one framework vs **adapts**/synthesizes from several
  (`decision-0002`) is a free choice — but the choice, and every *divergence* from a
  reference, is a recorded decision (B4) and surfaced (D1), never silent drift. *(Replaces
  v0's "never adopt wholesale," which wrongly foreclosed the conductor end — the Spec Kit case.)*
- **B9. Clarify before commit** — *strong (new, from framework analysis).* Ambiguity in an
  upstream artifact is actively **surfaced and resolved** — usually by asking the human —
  before downstream consumes it; ambiguity is never silently resolved by guessing.
  *Near-universal: Spec Kit `/clarify`, Kiro ambiguity/gap analysis pre-code, SpecSwarm
  clarification + `/ss:decisions`.* Arguably the most central uncertainty-reduction act, and
  absent from v0.
- **B10. Bounded correction, then loud escalation** — *provisional (new, agentic).* On
  verification failure, attempt **bounded** auto-correction (retry/repair to a limit), then
  **escalate visibly** — never infinite-loop, never silently abandon. *SpecSwarm `/ss:fix`
  auto-retry + watchdog; Agent SDK verify-loop.* (Most operational of the set — candidate to
  collapse into D1; see open questions.)
- **B11. Epistemic integrity — calibrated, independent judgment** — *provisional (new;
  agent-conduct, a different kind).* The agent's assessments **track the evidence, not the
  human's preferences**: it states disagreement and risks plainly when analysis warrants, and
  affirms only when the data genuinely supports it — *withholding warranted positive signal is
  as much a distortion as manufacturing praise*, and **performed disagreement is sycophancy
  inverted**. The **human-facing twin of B3** (the builder doesn't grade itself; the agent
  doesn't flatter the human) and the **precondition that makes the intent gate D2 real** — a
  sycophantic gatekeeper cannot catch a wrong upstream. *Weakly checkable: realized in
  sub-agent design (adversarial, self-derived checklists) and system prompts, not a gate.
  Candidate to merge into B3.*

---

## C. The two configuration dials (`decision-0008`) — *provisional*

Per gate, two settings — the mechanism that lets one invariant structure serve both a
speed-first startup and an assurance-first enterprise (`decision-0004`):

- **C1. Enforcement strength:** `expressed` (documented only) → `default-on-but-skippable`
  → `enforced`. Bonsai can move a methodology's expressed structure toward enforced; that
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
  Bonsai may *allow* skipping, but never *silently*. This is plausibly **the sharpest one-line
  statement of Bonsai's value — "surface the choice," not "enforce the choice."**
- **D2. The intent gate never fully opens** — *durable* (v0-4 core). At the intent locus (A3),
  C2 can never be `none`: a human (or, by ratchet, an independent check the human authorized)
  is mandatory. It is the only place an upstream that is itself *wrong* gets caught. The one
  gate strictness can never dial to zero.

---

## Changes from v0 → v1 (with rationale + source)

| # | Change | Why | Source |
|---|---|---|---|
| 1 | Split flat list into **A/B + dials + floors** | Step 1: frameworks have structure, not enforcement | `research-0002` |
| 2 | **Reclassified Independent Verification** (v0-5) `methodology` → `bonsai-design` | spec-driven tools lack it; Bonsai supplies it; SpecSwarm proves it implementable | `research-0002` |
| 3 | Split "gate at every handover" (v0-2) into **A2 (points exist)** + **B2 (enforced)** | enforced ≠ merely defined; skippable gates puncture it | `research-0002` |
| 4 | Added **A4 ratifiable/checkable artifacts** as explicit structural prerequisite | it's what B1/B3/B4 act on | synthesis |
| 5 | **Enforcement demoted to a dial (C1)**, not a mandate | strictness was a hypothesis; keep speed-first users | `decision-0008` |
| 6 | Added **gatekeeper dial (C2) `{agent\|human\|none}`** | who enforces is configurable | `decision-0008` |
| 7 | **Elevated loud-failure (v0-7) → transparency floor (D1)** incl. conscious skip | the true floor is surfacing, not enforcing | `decision-0008` |
| 8 | Named the **intent-gate floor (D2)** explicitly | the one non-configurable gate | v0-4 + `0008` |
| 9 | **Split B7**: minimal-first kept; reference-not-adoption → **B8** (recorded reference relationship) | strict single-framework adopt is legitimate (conductor); v0 foreclosed it | `decision-0002` + Spec Kit case |
| 10 | Added **B9 clarify-before-commit** | near-universal framework feature; central uncertainty-reduction act, absent from v0 | `research-0001/0002` |
| 11 | Added **B10 bounded-correction-then-escalation** | agentic reliability; prevents infinite-loop / silent abandonment | SpecSwarm, Agent SDK |
| 12 | Added **B11 epistemic integrity** (calibrated, anti-sycophancy) | human-facing twin of B3; precondition for D2 to function | maintainer + B3/D2 logic |
| — | Kept: v0 3+7 merge (B4); bounded context (B5); self-improvement (B6); minimal-first (B7); directional flow (A1/B1); intent locus (A3) | — | — |

---

## Acceptance criteria

- The **admission gate (A) is small** (4 structural properties) and is the *only* set
  `decision-0003`'s ingestion check uses.
- Each operating invariant (B) is something Bonsai *supplies*, not something the methodology
  must already have.
- Strictness and gatekeeper are **dials with surfaced defaults**, not hard-coded — so the
  same set serves startup and enterprise (`decision-0004`).
- The two floors are stated as non-configurable, with D2 the recognized exception to C2.

## Open questions

- **Is D1 (transparency) its own floor, or still part of loud-failure (B-layer)?** v1 elevated
  it on the strength of `0008`; ratification should confirm or fold it back.
- **Does A4 over-constrain pattern-level methodologies** (e.g. Agent SDK) that have no explicit
  "approved" state — or is a loose/implicit ratification enough to clear the gate?
- **Provenance (B4): bonsai-design or sometimes structural?** OpenSpec shows it can be a
  framework's *strength* — the v0 contested call, still open.
- **The dials (C) need a second instance.** They are `provisional` from one evidence round;
  validate against instance #2 (still the N=1 risk, `decision-0001`).
- **Tier-2 boundary confirmations** to fold in fully: Devin (open-ended; merge-time human gate
  = partial D2/A3, fails A1?), Cursor (fails A1/A2 — confirm as the clean negative control),
  Claude Agent SDK (passes A as pattern-with-guardrails).
- **Does any operating invariant (B) collapse or graduate to a dial?** Minimal-first (B7)
  applied to v1 itself.
- **Minimal-first tension:** this round grew the operating layer 7 → 11. Which of B8–B11 are
  real invariants vs. practices? **B10** (candidate: collapse into D1) and **B11** (candidate:
  merge into B3) are the likeliest to retire. *B10 was added on the maintainer's instruction
  over my recorded reservation that it is more operational than invariant — noted here, not
  buried (which is B11 in action).*
- **Is B11 a new *class* (agent-conduct/epistemic) or part of B3?** If kept separate, the set
  recognizes two faces of independence — conformance (B3) and intent (B11).
- **Adherence (`decision-0002`):** working stance is two coarse modes — *adopt* (one
  framework) vs *adapt* (synthesize from several); deliberately **not** formalized into the
  invariant model until instance #2 can test it. v1 encodes only the durable part (B8).
