---
id: invariants-v0
type: invariant-set
status: superseded
depends_on: [brief-§4]
owner: gundi
superseded_by: invariants-v1
---

# Bonsai's Invariants — our synthesis, v0

> **⚠ Superseded (draft) by [`bonsai-invariants-v1.md`](bonsai-invariants-v1.md).** v1 splits
> this flat list into a structural admission gate + a configurable operating layer + dials +
> floors, using Step 1 evidence (`research-0002`) and `decision-0008`. v0 is kept only for
> diffing during ratification; **do not consume v0 as current truth.**

> **Provenance & honesty (load-bearing).** This is **our synthesis**, version 0. It is
> *not* an established or externally-attributed set. Do not name it authoritatively or
> imply prior provenance (see the §4 guardrail in the brief — the "Keel's invariants"
> canary). It is lifted from the brief's §4 so it can become a *first-class, checkable
> artifact* — the thing every methodology Bonsai supervises is checked against. Validating,
> pruning, and refining this set is **the project's first job** (decision `0004`), and it
> is **not yet ratified** — it awaits research (decision `0006`).

> **Durability tags** carry the brief's confidence read: invariants **1–7 clearly durable**;
> **8–9 strong but less settled**. Tags are claims to be falsified, not facts.

## Two kinds of invariant (the distinction that defines the gate)

Each invariant is classed as one of:

- **`methodology`** — a property a *target methodology must already exhibit* for Bonsai to
  supervise it. These form the **admission gate**: when Bonsai ingests a methodology
  (decision `0003`), it checks it against *these*; a methodology that lacks one is *out of
  contract* and Bonsai fails loudly (invariant 7). "What each target methodology already
  follows, and needs to follow."
- **`bonsai-design`** — a property Bonsai imposes on a Bonsai-*assisted* project **by our
  own volition**, regardless of methodology. The project gets it *because* it adopted
  Bonsai. **Not** admission criteria — Bonsai *supplies* them. "What a Bonsai-assisted
  project follows by our design choice."
- **`hybrid`** — both faces are load-bearing: a methodology-facing core (the gate) *plus* a
  Bonsai-supplied enforcement/extension. The two faces are stated separately, and **only
  the methodology-facing core is part of the admission gate.**

**Why this matters:** the gate for using Bonsai with a methodology is the `methodology`
kind, *never* the `bonsai-design` kind. Conflating them would either reject good
methodologies for lacking things Bonsai supplies, or fail to reject unsupervisable ones.

## The set

### 1. Directional handover — *durable* · `hybrid`
Information flows one way through stages of *decreasing ambiguity*
(research → decisions → contracts → implementation → validation). Downstream consumes only
*ratified* upstream, never drafts.
- **Gate (methodology must exhibit):** stages of decreasing ambiguity with one-way flow. A
  methodology with no directional structure is not supervisable — flag loudly.
- **Bonsai imposes:** the `draft → ratified → approved` status lifecycle that *enforces*
  "consume only ratified." *(Contested: does the methodology itself need a notion of
  ratification? — open question.)*
- **Checkable as:** at ingestion, the methodology's stages form a DAG of decreasing
  ambiguity; at runtime, no consumer reads an input below `ratified`.

### 2. A gate at every handover — *durable* · `methodology`
Each stage transition has a verification gate. Its *nature* varies (human judgment vs
automated check); its *existence* does not.
- **Gate (methodology must exhibit):** a defined verification at every stage transition (a
  review, a definition-of-done, a check).
- **Bonsai imposes:** enforcement only — it flags any handover the methodology leaves
  ungated; it does not invent the gate set.
- **Checkable as:** for every handover edge in the process graph, a named gate exists;
  missing → loud flag.

### 3. The auditable archive: provenance + immutable history — *durable* · `bonsai-design`
*(Merged from former invariants 3 (provenance/traceability) and 7 (immutable decision
history / consolidated current truth) per the simplify lean — invariant 9 applied to the
set. Re-split if friction demands.)*
Every artifact links to its upstream and rationale (sources, confidence, dependency edges).
The decision history is **append-only** (you *supersede*, never edit, a ratified decision);
current truth lives in **revised-in-place consolidated docs**. Traceable artifacts +
immutable decisions + consolidated current-truth together keep the archive auditable and
navigable as it grows.
- **Bonsai imposes:** the whole archive discipline — linking, confidence tags, append-only
  decision log, consolidated current-truth. A methodology needn't already do this; Bonsai
  supplies it (conforming to an existing decision log if one is present). *(Contested: a
  methodology producing no linkable artifacts at all may be un-auditable and belong at the
  gate — open question.)*
- **Checkable as:** every artifact resolves a non-empty `depends_on` (or is a declared
  root); load-bearing claims carry a source + confidence tag; the decision log is
  append-only (supersede via forward pointer); consolidated docs are the only ones revised
  in place.

### 4. Authority split: humans own intent, agents own conformance — *durable* · `hybrid`
Humans are mandatory where *values/intent* are set; agents run autonomously where work
merely *conforms* to an approved upstream. The boundary *ratchets* toward autonomy on a
track record. The intent gate **never fully opens** — it is the only place an upstream that
is itself *wrong* gets caught.
- **Gate (methodology must exhibit):** defined points where humans own intent/values and
  sign off. A methodology with no human intent locus is not targetable for accountable dev.
- **Bonsai imposes:** the agent-conformance side, the autonomy *ratchet*, and the "intent
  gate never fully opens" rule — how Bonsai runs *agents* safely.
- **Checkable as:** methodology locates intent ownership in humans; every gate is typed
  `intent` (human-required) or `conformance` (agent-eligible); no `intent` gate is fully
  automated.

### 5. Independent verification — *the builder does not grade itself* — *durable* · `methodology`
Whatever produces work is never its sole judge. Some independent check (agent, panel,
human) verifies against the approved upstream with a checklist it derives itself.
- **Gate (methodology must exhibit):** production is separated from judgment — a check
  independent of the author. A methodology where the builder is sole grader → flag loudly.
- **Bonsai imposes:** the *self-derived checklist* discipline, applied to agent work.
- **Checkable as:** verifier identity ≠ producer identity; the checklist is derived from
  the upstream, not supplied by the producer.

### 6. Bounded context — *durable* · `bonsai-design`
Each operation reads only what it needs — its contract, the code, and the *live* decisions
it cites — never the whole archive. A scaling necessity for running agents.
- **Bonsai imposes:** this entirely — a methodology written for humans says nothing about
  context windows; it is an agentic-execution property.
- **Checkable as:** an operation's declared inputs are an enumerated strict subset of the
  archive, never "everything."

### 7. Loud failure over silent degradation — *durable* · `bonsai-design`
When a required capability is missing (a tool, a source, a verification), the system fails
*visibly* rather than producing plausible-but-unverified output.
- **Bonsai imposes:** this operating stance — it is *also the mechanism* by which an
  admission-gate failure (a missing `methodology` invariant) surfaces.
- **Checkable as:** a missing required input halts with a visible error; no degraded-mode
  output is emitted unlabeled.

### 8. Self-improvement is a disciplined loop, not vigilance — *strong, less settled* · `bonsai-design`
The process evolves by *triggers* (condition → action, recorded where the actor trips over
them), riding an existing ritual rather than adding ceremony, biased toward *retiring*
rules and staying *subordinate to the work*.
- **Bonsai imposes:** this evolution mechanism — how Bonsai changes a fitted process safely
  over time.
- **Checkable as:** each process rule has a recorded trigger and a home where it fires; net
  rule count trends flat-or-down absent justified additions.

### 9. Minimal-first, reference-not-adoption — *strong, less settled* · `bonsai-design`
Start with the smallest process that works; add a step only when *friction reveals the
boundary*; treat existing frameworks (BMAD, spec-kit) as a *parts catalog*, recorded as an
explicit decision — never a wholesale identity to inherit.
- **Bonsai imposes:** this is *meta to methodologies* — it governs how Bonsai itself fits,
  prunes, and borrows. The most purely-Bonsai invariant in the set.
- **Checkable as:** every added step cites the friction that justified it; every borrowed
  element cites the decision that adopted it.

## The admission gate (summary)

When Bonsai ingests a methodology (decision `0003`), it checks **only the methodology-facing
content**:

- **1** — directional flow exists
- **2** — every handover has a gate
- **4a** — humans own intent somewhere
- **5** — the builder does not grade itself

If any is absent, Bonsai fails loudly (invariant 7): *this methodology is out of contract.*
Everything else — **3, 4b, 6, 7, 8, 9** — Bonsai **supplies**; they are guarantees a
Bonsai-assisted project gets by adopting Bonsai, not demands on the methodology.

**Consequence:** the admission gate is **small (~4 structural properties)**. That (a) is
on-thesis (minimal-first, invariant 9), and (b) makes the central empirical question —
*do real methodologies satisfy the invariants?* — far more tractable: candidate
methodologies are tested against ~4 properties, not 9.

## Acceptance criteria

- Each invariant is classed `methodology` / `bonsai-design` / `hybrid`; for hybrids the two
  faces (gate vs Bonsai-supplied) are stated separately.
- The **admission-gate subset** (what a target methodology must exhibit) is explicit and
  small, and is the *only* set decision `0003`'s ingestion check keys on.
- Each invariant is stated independently of any specific methodology, domain, or stack.
- Each has at least a *seed* of how it could be checked (**Checkable as**).

## Open questions

*These are the research agenda (decision `0006`) — the set is held `draft` until they are
addressed, not ratified on assertion.*

- **Is the admission gate exactly {1-flow, 2, 4-intent, 5}?** Two contested calls:
  - **3 (provenance):** Bonsai-supplied (it can retrofit links) — *or* does a methodology
    producing no linkable artifacts fail the gate as un-auditable, moving it into the gate?
  - **1's split:** is the ratification *lifecycle* Bonsai's to add, or must the methodology
    already carry a notion of ratification (enlarging 1's gate face)?
  *(Maintainer: can't lean obviously yet — needs data. Deferred to research.)*
- **The central empirical test:** pointed at real methodologies (BMAD, spec-kit, others)
  and real projects across team sizes / domains / regulatory contexts, do they satisfy the
  four gate invariants? Which is the *first to break*? (The N=1 risk; instances #2/#3 — and
  our own build process as instance #1, decision `0005` — answer it.)
- **Are 8 and 9 invariants or merely good practice?** Tagged weaker on purpose; both
  `bonsai-design`, so neither ever gates a methodology. What evidence would demote/promote?
- **Did the 3+7 merge lose anything?** Re-split if a project needs provenance and
  decision-immutability gated or checked separately.
- **Completeness:** what property of agentic dev under human accountability is not yet
  captured that friction will later reveal — and would it be `methodology` or
  `bonsai-design`?
- **Are we reinventing a named set?** (Provenance honesty.) Does an existing framework
  already articulate these invariants? If so, cite it; do not claim false novelty.
