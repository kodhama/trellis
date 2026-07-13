---
id: research-0008
type: research-note
status: ratified
ratified: 2026-07-03
depends_on: [invariants-v1, decision-0012]
informed_by: [research-0005, research-0006, research-0007, brief-§4]
owner: gundi
---

# Research 0008 — Design changes the research implies + the naming question (deliverable 2.3)

> **Method & honesty.** Two parts: (1) what `research-0005/0006/0007` imply the product's design
> should *become* (a proposal, ratified 2026-07-03 together with `research-0005/0006/0007`), and (2) the
> name. **The naming call is intent-layer (D2) — the maintainer's, not mine;** I gave an evaluated
> recommendation (stage), the maintainer chose to rename now (option B). The outcome is recorded in
> `decision-0015`. This note preserves the analysis, including the path *not* taken.
> Load-bearing convention: in **Part 2**, **"Bonsai"** denotes the **old** product name and
> **"Trellis"** the **new** one — they are deliberately kept distinct here (this note is *about* the
> choice between them, so it is the one place the global rename does not flatten them).
>
> *Amended in place 2026-07-13 (`decision-0047` + `grove/adr-0011`; consumer-audit
> marking-class). WHAT: `research-0005`, `research-0006`, `research-0007`, and `brief-§4`
> moved out of frontmatter `depends_on` into a new `informed_by` list — they informed this
> note's design-implication and naming analysis without this note's own conclusions being
> contingent on them; provenance, not coupling. No `version` counter on this artifact to
> bump. POINTER: `decision-0047` Consequence 4, `grove/adr-0011`.*

## Part 1 — Design changes the research implies (proposal)

The lenses did not just decorate the invariants; they surfaced concrete design moves. Each is
already argued in a note — this consolidates them as the design proposal (adopt on ratification):

1. **Promote the *expression profile* to a first-class, typed artifact** — the per-instance control
   map: which invariants are active × C1 strength × C2 gatekeeper × delivery axes (A/B). It unifies
   #22 (a minimal profile), #23 (Assess produces it), #24 (one filled-in instance), #28 (diff across
   instances). Types are open (`research-0003`) so this needs only a recorded decision. *([[research-0005]], [[research-0006]] convergence.)*
2. **Ship the *invariant-signature catalog* as `trellis-core`** — the genome annotation (what each
   invariant looks like when honored). `decision-0009` already anticipates "catalog." Feeds #23 + #27.
3. **Compute the default gate-set** — formalize the mechanizable fragment (A1/A4/B1-flow/B2) as a
   spec K and compute the default enforcement as its supremal controllable sublanguage; feed it to
   the conformance sub-agent. *([[research-0006]] Proposal — the most concrete, iron-rule-friendly move.)*
4. **Extend delivery to two axes** — supervisor/advisor × payload-depth, with the "converge at
   +mechanism" caveat. *([[research-0007]]; a decision extending `decision-0012` when the fork is called.)*
5. **Bounded context needs an explicit observer** — B5 operations act on a *projection*; the
   `depends_on` + provenance graph *is* the state-estimator, and must be good enough to decide on.
   *([[research-0006]] Result 2.)*
6. **Add a gate-conflict check** — modular gates can be individually correct yet jointly blocking.
   *([[research-0006]] Result 4.)*
7. **Keep D1/D2 non-configurable — now with a control-theoretic rationale, not just a stipulation**
   (D2 = a fault class the supervisor cannot disable). No change to the floors; a firmer *why*.

None of 1–7 required a rename. That separation is why the name could be decided independently.

## Part 2 — The name (resolved: option B — renamed to Trellis; see `decision-0015`)

### Was the instinct right? Yes — and it was checkable.

The old brief defined **Bonsai** as *"keep a process minimal, deliberately pruned, and shaped to
fit… start small, prune relentlessly, **shape to the specimen**."* Two halves:

- **The prune/minimal/adaptive half was load-bearing and correct** — it maps cleanly to B7
  (minimal-first) and B1's retire-bias. It is *kept* (Trellis is still a garden object).
- **The "shape to the specimen" half is *topiary* — the wrong control philosophy.** Bonsai imposes
  an external *form* on a whole organism. The research says the essence is the **opposite**: a
  *maximally-permissive supervisor* that **constrains a space and permits maximal freedom within**
  ([[research-0006]] Result 3). Shaping-to-a-form ≠ constraining-a-space-and-permitting-freedom. So
  the dominant metaphor pointed the wrong way. The instinct held.

### The recommendation was to *stage*; the maintainer chose to rename *now*.

For the record (the path not taken): staging (option A) was recommended on (i) the project's own
naming guardrail — naming is a deliberate-later act; (ii) a name needs an *audience*, and there is
none yet; (iii) this lens is genealogically N=1. The maintainer overrode this (option B) on the
decisive counter-argument that **churn is cheapest now** (small, internal corpus) and every day on
the wrong name spreads it further. Both readings were legitimate; the call was the maintainer's (D2).

### The shortlist (as evaluated)

| Direction | Candidate | Captures | Miss / risk |
|---|---|---|---|
| Keep | **Bonsai** (old) | minimal/prune/adaptive (real) | topiary metaphor = shaping-to-a-form (wrong philosophy) |
| Horticultural, *corrected* | **Trellis** ← **chosen** | constrain **and support**; the plant grows *freely within bounds* = maximally-permissive supervisor; keeps the garden lineage; **fixes** the topiary flaw | still a metaphor; rename churn (accepted) |
| Control | Envelope | the permissive safe boundary; freedom within | abstract; loses the living warmth |
| Teaching / removable | Scaffold | educational scaffolding (Pillar III) + removable overlay | connotes *temporary* — wrong for the floors (D never removes) |
| Genetics | Regulon / Epigenome / Morphogen | the regulation/expression layer | obscure jargon; names the *weaker* lens; false-provenance risk — **kept as internal vocabulary, not the product name** |
| Invariant-core | ~~Keel~~ | the stable floor | **literally the project's own false-provenance canary** (`brief-§4`) — avoided |

**Why Trellis won:** the one candidate that keeps the garden lineage *while correcting the exact
flaw* — a trellis is **structure that enables growth rather than dictating form**, the permissive
supervisor rendered as a garden object.

### What did *not* change

- The **invariant-set** framing stays *"our synthesis, v1"* (only the possessive updates to
  "Trellis's invariants") — the guardrail's actual target is untouched.
- Genetics/DES terms remain **internal vocabulary** (expression profile, supervisor/advisor,
  observer), not the product name.

## Acceptance criteria

*(research-note: Open questions + sources/confidence; no acceptance-criteria gate.)*

## Open questions

- ~~Ratify `decision-0015`~~ — **done** (ratified 2026-07-03 by the maintainer).
- Which Part-1 changes graduate to decisions first? (Candidate order: expression-profile artifact →
  delivery-axes extension of `decision-0012` → compute-default-gate-set into the spine.)
- Follow-ups from the rename (`decision-0015`): local working-directory name; auto-memory files;
  open-issue bodies.
- Sources: old brief §1 (Bonsai definition), `brief-§4` (naming guardrail); [[research-0005]]/[[research-0006]]/[[research-0007]]. `verified` (corpus); naming judgments `inferred`.
