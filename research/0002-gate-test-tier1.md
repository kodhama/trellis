---
id: research-0002
type: research-note
status: draft
depends_on: [research-0001, invariants-v0, decision-0006]
owner: gundi
date: 2026-06-29
---

# Research 0002 — Gate-test of the Tier 1 frameworks (Step 1)

> **Method.** Deep-research pass (6 angles, 22 sources, 109 claims, top 25 verified 3-vote
> adversarial → 22 confirmed, **3 killed**) covered **Spec Kit** and **Kiro** at
> `confidence: high`. **BMAD** and **OpenSpec** dropped out of that run's synthesis, so they
> were scored by targeted primary-source fetch (`confidence: high` for OpenSpec — explicit
> docs; `medium` for BMAD — its user-guide returned 404, pieced from deepwiki + Step 0).
> Transcript: workflow `wf_3b7e60be-0ce`. *Operational caveats:* several Kiro agents stalled
> and retried; three "hard-enforcement" claims about Spec Kit were **refuted** (see below).

## The question

Do the Tier 1 frameworks satisfy the four **admission-gate** (`methodology`) invariants —
`{1 directional-flow, 2 gate-at-handover, 4a humans-own-intent, 5 independent-verification}`?
And where do the invariant *definitions* break against real frameworks?

## Matrix (strict "is it enforced?" reading)

| Framework | 1 Directional handover | 2 Gate at every handover | 4a Humans own intent | 5 Independent verification |
|---|---|---|---|---|
| **Spec Kit** | PARTIAL — ordered, but *recommended* not enforced; "lean path" skips gates | PARTIAL — Phase-(-1) prereqs mandatory; `clarify`/`checklist`/`analyze` skippable | PARTIAL — user approves tests; editable template, not CLI-enforced | **ABSENT** — same AI self-checks; no distinct verifier |
| **Kiro** | PARTIAL — standard enforces req→design→tasks; **Quick Plan bypasses**; Design-First reverses | PARTIAL — gates **skippable via Quick Plan** | PARTIAL — standard approval opt-in; Quick Plan reviews *after the fact* | **ABSENT** — no builder-distinct verifier evidenced |
| **OpenSpec** | PARTIAL→weak — *explicitly rejects phase-lock* ("dependencies are enablers… what's possible, not what's required next") | **ABSENT** — no mandatory gates; `verify` runs after code and "won't block archive" | PARTIAL — humans initiate; no formal approval step | **ABSENT** — `/opsx:verify` is the *same* AI, optional |
| **BMAD** | PARTIAL — ordered phases via workflows (PRD→arch→epics→stories); not hard-enforced | PARTIAL — planning artifacts gate implementation; "code reviews" in build phase | PARTIAL→PASS — human planning collaboration is core (mandatory-ness undocumented) | **PARTIAL→PASS** — **distinct roles**: Developer (Amelia) vs reviewer/QA, optional Test Architect (Murat) |
| **SpecSwarm** *(Tier 2)* | PARTIAL — `spec→plan→tasks→implement`; recommended, not enforced | PARTIAL→PASS — 80/100 quality gate **mandatory before merge**; `--quick`/`--minimal` skip sub-gates | PARTIAL — upfront decisions touchpoint (`/ss:decisions`) + clarification Qs; no formal sign-off | **PASS** — **fresh-context adversarial `spec-mentor`** (spec-vs-code) + per-task verifier, both **distinct from the implementer** |

Secondary (`3 provenance`, a `bonsai-design` invariant, noted for interest): **OpenSpec is
strongest** — versioned spec library + change deltas (ADDED/MODIFIED/REMOVED) + dated archive
preserving "the proposal explaining *why*, the design explaining *how*." Kiro/BMAD medium;
Spec Kit weak.

## The headline finding

**No Tier 1 framework cleanly passes the strict gate — and they fail the *same* way.** All
four *express* directional flow and gates as **documentation/agent-instruction**, but none
**enforce** them (the three refuted claims confirm Spec Kit's ordering is prose, not CLI-hard
blocking). Optional/skippable fast-paths (Kiro "Quick Plan", Spec Kit "lean path", OpenSpec's
fluid actions) puncture the gates. **Independent verification is the weakest property** —
ABSENT in all three spec-driven tools (the builder grades itself with the same AI).

Two cross-cutting patterns worth keeping:
- **Different archetypes are strong on different invariants.** BMAD's *role decomposition*
  gives it the independent-verification property the spec-driven tools lack; OpenSpec's
  *change/archive model* gives it the best provenance. This directly supports the
  "which-invariant-subset" framing (decision `0006`).
- **Enforcement is universally missing.** That is not a coincidence — it is the gap.

## The reframe that resolves "nothing passes" (and validates the taxonomy)

The strict reading makes everything PARTIAL/ABSENT — which would suggest the gate is too
harsh. It isn't, once we read it through the invariant set's own
`methodology`-vs-`bonsai-design` split:

- **What the admission gate should test (structure / `methodology`):** does the process
  *have* directional stages, defined handover points, an intent locus, and a place where
  verification can attach? On this reading **all four largely PASS** — the structure is
  there.
- **What the frameworks LACK (enforcement / `bonsai-design`):** consume-only-ratified,
  *mandatory non-skippable* gates, and *producer ≠ verifier* independent checking. This is
  exactly the layer the invariant set says **Bonsai supplies**.

So the Step 1 result is not "the frameworks fail" — it's "**the frameworks supply the
structure; Bonsai supplies the enforcement.**" That both *validates the methodology/bonsai-
design distinction* and *defines Bonsai's wedge precisely*: turn expressed→enforced, and add
the independent verifier none of them have.

## Tier 2 addendum — SpecSwarm: the independent verifier exists in the wild

SpecSwarm answers Step 1's biggest gap. Its **`spec-mentor`** is a *fresh-context,
adversarial* subagent doing **spec-vs-code** verification, and a per-task **verifier**
subagent confirms each task's acceptance criteria — both **distinct agents from the
implementer** (producer ≠ verifier). `/ss:ship` adds a parallel review panel (code-reviewer,
silent-failure-hunter, type-design-analyzer). `confidence: high` (repo README; `WORKFLOW.md`
404'd).

**It matters twice:**
1. **Invariant 5 is not aspirational** — an AI-native framework already embodies "the builder
   does not grade itself" via a fresh-context adversarial checker. It nearly mirrors the
   source project's `conformance-reviewer` (the brief's harvest target). So independent
   verification is *demonstrably implementable*, not just present in human-role models (BMAD).
2. **SpecSwarm is the closest existing model of the maintainer's target UX** (decision
   `0008`): a quality gate that is **mandatory-by-default** (80/100 before merge) yet has
   **surfaced skip fast-paths** (`--quick`, `--minimal`). Enforced-but-skippable, consciously
   — exactly the "allow the skip, but surface it" stance.

## Candidate refinements for Step 2 (the bidirectional payoff)

1. **Expressed vs Enforced is a first-class axis — but enforcement is a *dial*, not Bonsai's
   fixed stance (decision `0008`).** Re-score every gate property on *enforcement strength*:
   `expressed` (documented) / `default-on-but-skippable` / `enforced`. The admission gate
   (methodology) tests only that the structure is *expressed*. Bonsai can move it toward
   *enforced* — but that strictness is an **opt-in layer** (enterprise/assurance), never
   forced, so speed-first users aren't alienated. **The actual floor is surfacing, not
   enforcement:** a skip is allowed if it is made *visible/conscious* (extends invariant 7,
   loud failure). That — "surface the choice" — may be the sharpest statement of Bonsai's
   value, sharper than "enforce."
2. **Skippable gate ≠ gate (invariant 2).** A gate that an optional fast-path bypasses does
   not satisfy "a gate at every handover." Define mandatory-vs-optional explicitly.
3. **Intent sign-off must be mandatory and *pre-downstream* (invariant 4a).** Review-after-
   the-fact (Kiro Quick Plan) does not count — approval must precede downstream work.
4. **Verification must not be self-review (invariant 5).** The same AI checking its own
   output (Spec Kit, OpenSpec `/verify`) fails "the builder does not grade itself." Requires
   producer ≠ verifier — BMAD's role separation is the positive example to encode.
5. **Directionality without enforcement is PARTIAL, not PASS.** Suggested ordering ≠
   directional handover; the structure must at least *intend* one-way flow (OpenSpec's
   explicit anti-phase-lock stance is the edge case — does it even clear the structural
   gate?).

## Acceptance criteria

- Every framework × property cell has a verdict; unscored frameworks were filled by targeted
  fetch, not left blank or guessed (BMAD's gaps are flagged, not hidden).
- The analysis distinguishes *structure* (gate) from *enforcement* (Bonsai's layer) rather
  than collapsing them into one pass/fail.

## Open questions (carried into Step 2)

- **Does Spec Kit's CLI actually block on missing upstream, or only instruct the agent to?**
  (The refuted claims say instruct-only — worth a code-level check before relying on it.)
- **Does the "structural gate" reading hold for OpenSpec**, which *deliberately* rejects
  enforced ordering? If a framework rejects directional flow on principle, does it clear even
  the structural admission gate — or is it the first true gate failure?
- **Should provenance (invariant 3) move toward the gate** after all? OpenSpec shows it can
  be a framework's *strength*, not just something Bonsai retrofits (the contested call from
  the invariant set).
- **BMAD verification depth:** is its reviewer/QA role a *genuine* independent check (own
  checklist, distinct context) or nominal? Needs a closer look to confirm the PASS.
- **Tier 2 status:** SpecSwarm ✅ done (see addendum — it *is* the real independent-verifier).
  Still owed: Agent SDK, Devin, Cursor (negative control) — boundary cases, can fold into
  Step 2.
