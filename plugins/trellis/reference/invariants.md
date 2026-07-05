---
id: signature-catalog-v1
type: signature-catalog
status: ratified
depends_on: [invariants-v1, spec-0002]
owner: gundi
scope: trellis-product
ratified: 2026-07-04
---

> **Ratified via merge (`decision-0022`).** The agent authored this; the maintainer's **merge of this
> PR is the ratification** (D2) — the `draft → ratified` flip rides the reviewed diff. Coverage is
> independently checked (AC1); the `signature` / `why` / example / dial calls embody judgment the
> maintainer accepts by merging.

# Signature catalog — v1 (the genome annotation)

> **What this is.** The one shipped **dictionary** of Trellis's invariants: per invariant — what it
> *is*, **why** it earns its place (the goal, agents-first), the observable **signature** by which a
> project is seen to honor it, **≥2 contrastive `honored` / `violated` cases each — drawn from
> different layers** (CI / spec / research / code / UI / ops …) so the principle reads as general, not
> domain-specific — and its **default dials**. Schema + lifecycle: `spec-0002`. Slugs: the
> `invariants-v1` registry. **The benefits page derives from the `why` + honored/violated here** — no
> claim on the page without a rule behind it (`decision-0020`). Consumed by Assess (#23) and tutoring
> (#27). `trellis-product` scope — one, shipped.

> **Coverage (spec-0002 §1, AC1).** Covers the **14 assessable invariants** — A structural, B operating
> (incl. **B6 `inv-self-improvement`**, `decision-0018`), D floors. `inv-reference-relationship` (B8)
> was **collapsed into D1 + the adopt/adapt dial** (`decision-0021`) — its "divergence from a framework"
> case lives in D1's example below. Excludes the two C dials (they are the *axes* entries are set
> along, not rows).

> **On `mechanizable`.** `true` marks the SCT-computable fragment — structurally checkable. `false`
> marks a **behavioral gene** whose signature is a judgment tell, not a regex. Every invariant carries
> a `honored`/`violated` pair, and **a change that edits an invariant without updating its examples is
> a conformance failure** (`decision-0020` meta-rule — the iron rule applied to the rule-set itself).

## Entries

### A — structural (admission gate) · class `methodology`

- **`inv-directional-flow`** (A1)
  - what: one-way stages of decreasing ambiguity (research → decisions → contracts → implementation
    → validation); downstream never consumes a draft.
  - why: agents always build on **settled** ground — ambiguity only decreases, so no one codes against
    a spec that is still moving.
  - signature: ordered stage folders or a defined pipeline; artifacts carry a stage/`status`; **no
    ratified artifact cites a draft upstream**.
  - honored:
    - *(spec)* implementation reads an **approved** spec; a ratified doc never depends on a draft.
    - *(research)* a synthesis note cites only ratified findings, not draft ones still under review.
  - violated:
    - *(code)* an agent builds against a spec still being edited; it shifts, and the work is now built
      on a version that no longer exists.
    - *(UI)* a screen is implemented from a mockup still in review; the mockup changes and the build is
      silently wrong.
  - class: `methodology`  ·  mechanizable: `true`  ·  intent_locus: `false`
  - default_C1: `enforced`  ·  default_C2: `independent-agent`

- **`inv-handover-points`** (A2)
  - what: defined transitions between stages, each a place a gate *can* attach.
  - why: development moves in **discrete** steps with boundaries, not one fluid blur — the seams are
    where work can be stopped, inspected, and handed on. (A1 is *which way* work moves; this is *that
    it moves in steps*.)
  - signature: named handoffs (PR boundaries, stage transitions, review checkpoints), each with a
    before/after artifact.
  - honored:
    - *(dev)* a plan is "done," a spec "approved," a change "ready" — defined seams you can pause at.
    - *(CI)* each pipeline stage is a gate a check can attach to (lint → test → build → deploy).
  - violated:
    - *(dev)* vibe-coding melts prompt → code → prompt into one stream with no seam to inspect or gate.
    - *(CI)* one monolithic script does build+test+deploy with no checkpoint to stop or roll back at.
  - class: `methodology`  ·  mechanizable: `true`  ·  intent_locus: `false`
  - default_C1: `enforced`  ·  default_C2: `independent-agent`

- **`inv-intent-locus`** (A3)
  - what: humans own intent/values *somewhere identifiable* — a process with no human intent locus is
    not targetable for accountable development.
  - why: an **accountable human owns the goal**, so a wrong *direction* gets caught before it is built.
  - signature: an accountable human `owner` on artifacts; a human sign-off/approval point (CODEOWNERS,
    a required review, a ratification step).
  - honored:
    - *(product)* every feature traces to an accountable human `owner`; the "why" is a recorded decision.
    - *(research)* a research direction has a named human sponsor who can say "that's not what we're after."
  - violated:
    - *(product)* agents optimize a proxy metric no human owns, and ship the wrong thing efficiently.
    - *(ops)* an automated system changes behavior with no human accountable for the intent behind it.
  - class: `methodology`  ·  mechanizable: `false` (an `owner` field is checkable; *that it is a
    genuine intent locus* is judgment)  ·  **intent_locus: `true`**
  - default_C1: `enforced`  ·  default_C2: `human` (never `none` — D2)

- **`inv-ratifiable-artifacts`** (A4)
  - what: upstream can reach an **approved** state downstream consumes; outputs are **checkable
    against** it.
  - why: you build against a **stable, approved target with a real pass/fail criterion** — not a
    moving one, and not "looks done to me."
  - signature: a `status` lifecycle (draft → approved/ratified); artifacts with acceptance criteria a
    result can be graded against.
  - honored:
    - *(spec)* a spec reaches `ratified`, carries acceptance criteria, and work is graded against it.
    - *(data)* a schema is versioned; downstream validates against the **approved** version, not HEAD.
  - violated:
    - *(spec)* nothing is ever "final," so implementation chases a spec that keeps moving under it.
    - *(design)* a design system has no "released" state, so teams build against inconsistent snapshots.
  - class: `methodology`  ·  mechanizable: `true`  ·  intent_locus: `false`
  - default_C1: `enforced`  ·  default_C2: `independent-agent`

### B — operating (what Trellis supplies) · class `trellis-design`

- **`inv-graph-maintenance`** (B1) *(neighbor of B6; shares `inv-prune-bias`)*
  - what: the dependency graph of artifacts **and rules** kept consistent and minimal, information
    flowing one way; trigger-driven; append-only records superseded, never edited-in-substance.
  - why: the knowledge base **stays coherent for the agents reading it**, and a discovery deep in the
    code **repairs the decision that should have known it** (backprop) — which happens more than people
    admit.
  - signature: a `depends_on` graph; supersede/retire records; dependents re-reviewed on upstream
    change; no silent downstream patches; a bias to retire rules over adding.
  - honored:
    - *(docs)* a repaired ADR re-reviews its specs → plans → code, in turn.
    - *(code)* a build uncovering a missing architectural decision **creates the ADR** instead of
      patching around it (backprop).
    - *(research)* a downstream finding that contradicts an upstream note updates the *note*, not just
      the finding.
  - violated:
    - *(docs)* a decision changes but its dependent specs are never updated — they silently diverge.
    - *(data)* a schema migration isn't propagated to the docs, so the docs now describe a DB that no
      longer exists.
    - *(rules)* a stale lint rule keeps firing long after the pattern it guarded was removed.
  - class: `trellis-design`  ·  mechanizable: `true` (the **flow** facet; forward/backward/prune are
    judgment)  ·  intent_locus: `false`
  - default_C1: `enforced`  ·  default_C2: `independent-agent`

- **`inv-self-improvement`** (B6) *(restored first-class, `decision-0018`; neighbor of B1)*
  - what: the process learns from friction and gets better — improvement signals are surfaced and
    acted on, deliberately, so a glitch does not happen twice.
  - why: **a process glitch never happens twice** — friction becomes a fix, not a recurring tax.
  - signature: a trigger format (`condition → action`) stored where it fires; improvement signals
    surfaced through the project's **chosen channel** (asked/inferred, never assumed); retirement in
    the same change; prune-bias (the trigger set does not grow monotonically).
  - honored:
    - *(CI)* a flaky test recurs → a trigger is filed, the root cause fixed, the trigger retired in the
      same change.
    - *(process)* a repeated review miss becomes a checklist item that rides the PR you already write.
    - *(research)* a recurring dead-end is recorded so the next inquiry skips it instead of re-walking it.
  - violated:
    - *(CI)* the same pipeline step fails weekly and everyone just re-runs it, forever.
    - *(process)* a PR raises open questions with no follow-up and they rot, unowned.
    - *(onboarding)* the same confusion hits every new contributor because no one wrote the lesson down.
  - class: `trellis-design`  ·  mechanizable: `false` (the SI-1 surfacing floor is checkable against the
    declared channel; the proactive-notice disposition is not)  ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `human`

- **`inv-gate-at-handover`** (B2)
  - what: apply the verification gate at every A2 handover; any skip is **surfaced** (D1).
  - why: the review **actually fires** (not quietly skipped under deadline) — and if it is skipped,
    you can *see* it was.
  - signature: a check/review fires at each handoff (CI gate, required review); skips are logged, not
    silent.
  - honored:
    - *(CI)* a conformance check + review fire on every PR; a deliberate skip is *recorded*, not hidden.
    - *(release)* a promotion gate blocks an artifact that didn't pass the prior stage.
  - violated:
    - *(dev)* the review is "optional," so under deadline it silently doesn't happen and a defect ships.
    - *(spec)* a spec goes straight to implementation with no sign-off, and no record it was skipped.
  - class: `trellis-design`  ·  mechanizable: `true`  ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `independent-agent`

- **`inv-independent-judgment`** (B3) *(two faces: conformance + intent)*
  - what: the assessor is independent of what it assesses — the builder does not grade itself
    (conformance face); the agent does not flatter the human (intent face).
  - why: **the builder doesn't grade its own homework**, and the agent **names the risk** instead of
    flattering the plan — so verification and the intent gate are real, not decorative.
  - signature: a verifier **distinct from the producer** (fresh-context review agent); reviews record
    dissent/risks, not reflexive assent; the verifier derives its checklist from the approved upstream.
  - honored:
    - *(review)* a read-only reviewer, distinct from the author, derives its own checklist and reports
      what's wrong — even when inconvenient.
    - *(research)* findings are adversarially verified by a separate pass, not self-certified.
  - violated:
    - *(code)* the agent that wrote the code decides the code is good.
    - *(product)* the agent agrees with a flawed plan to please the stakeholder, and it sails through.
  - class: `trellis-design`  ·  mechanizable: `false` (the intent face lives in system prompts, weakly
    checkable)  ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `independent-agent`

- **`inv-auditable-archive`** (B4)
  - what: provenance + immutable decision history + consolidated current-truth.
  - why: you can always answer **"why is it this way?"** — decisions are not lost or quietly rewritten.
  - signature: append-only decision records; retained change history (git); a current-truth doc kept
    separate from its change log.
  - honored:
    - *(ADR)* decisions are append-only and link their rationale; superseding writes a *new* record.
    - *(code)* git history + a current-truth doc kept separate from its changelog.
  - violated:
    - *(ADR)* a decision is edited in place and its rationale lost; months later it is re-litigated
      from scratch.
    - *(infra)* an undocumented prod change, so no one can say why it is configured this way.
  - class: `trellis-design`  ·  mechanizable: `true` (presence of archive/provenance is structural)  ·
    intent_locus: `false`
  - default_C1: `enforced`  ·  default_C2: `independent-agent`

- **`inv-bounded-context`** (B5)
  - what: each operation reads only its declared inputs, never the whole archive.
  - why: agents decide on **sharp, relevant context** instead of drowning in everything — better calls,
    and it scales as the archive grows.
  - signature: operations/agents declare their inputs (`depends_on`, scoped context); sub-agents with
    narrow context/tool scope; an explicit observer (the dep-graph) over project state.
  - honored:
    - *(agent)* a sub-agent gets exactly its declared inputs and decides crisply within them.
    - *(data)* a query reads a scoped view, not the whole warehouse.
  - violated:
    - *(agent)* an op dumps the entire repo into context, dilutes the signal, and decides on noise.
    - *(code)* a function reaches into global state instead of its parameters, and breaks when the
      state moves.
  - class: `trellis-design`  ·  mechanizable: `false` (is the context *genuinely* bounded? — judgment)
    ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `independent-agent`

- **`inv-minimal-first`** (B7)
  - what: the smallest process that works; add a step only when friction reveals the boundary; bias to
    retire over add.
  - why: **no ceremony for its own sake** — every step has earned its place, so the process stays light.
  - signature: a deliberately small rule/process set; steps added with recorded justification; retired
    rules pruned rather than accumulated.
  - honored:
    - *(process)* a deliberately tiny rule set; a new step lands only with a recorded reason.
    - *(tooling)* a build config with no unused steps; a dependency pulled in only when needed.
  - violated:
    - *(process)* a heavyweight methodology copied wholesale, most steps cargo-culted.
    - *(code)* a whole framework pulled in to use one function.
  - class: `trellis-design`  ·  mechanizable: `false` (a disposition)  ·  intent_locus: `false`
  - default_C1: `expressed`  ·  default_C2: `human`

- **`inv-clarify-before-commit`** (B9)
  - what: ambiguity in an upstream is actively surfaced and resolved (usually by asking the human)
    before downstream consumes it; never silently guessed.
  - why: **agents ask instead of guessing** — you don't discover three files later that they took the
    wrong reading of a vague spec.
  - signature: open-questions sections; clarifying exchanges recorded before build starts; a
    `/clarify`-like step ahead of implementation.
  - honored:
    - *(spec)* an agent flags a vague requirement and resolves it *before* coding.
    - *(UI)* a designer confirms an ambiguous interaction before building it.
  - violated:
    - *(code)* an agent silently picks one reading of a vague spec, builds it, and it's the wrong one.
    - *(data)* an ambiguous metric definition is guessed, and the dashboard is subtly, confidently wrong.
  - class: `trellis-design`  ·  mechanizable: `false` (behavioral gene)  ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `human`

### D — floors (never configurable to "off") · class `floor`

- **`floor-transparency`** (D1) *(absorbs the framework-divergence case from retired B8, `decision-0021`)*
  - what: every consequential choice is **surfaced** — a skipped gate, a missing capability, a degraded
    result, a relaxed setting, a divergence from a framework you claim to follow; a failed verification
    is escalated visibly, never silently abandoned. **Drift is allowed, but never silent.**
  - why: **nothing consequential happens silently** — you learn about the shortcut *when it is taken*,
    not when it breaks in production.
  - signature: skips/degradations logged and visible; loud-failure on a missing tool/source; no silent
    fallbacks; divergence from a reference captured as a recorded decision.
  - honored:
    - *(CI)* a skipped gate is logged; a missing tool halts loudly, not silently.
    - *(framework)* "we diverge from Spec Kit here" is a recorded decision, so an agent knows where you
      follow the book and where you don't.
    - *(ops)* a degraded fallback (cache miss, retry, downgrade) is surfaced, not swallowed.
  - violated:
    - *(code)* an agent silently falls back to a degraded path; you learn when it breaks in prod.
    - *(framework)* the team quietly drifts from the methodology it *claims* to follow.
    - *(infra)* an error is swallowed and resurfaces later as a mystery outage.
  - class: `floor`  ·  mechanizable: `false` (a disposition; partially checkable)  ·  intent_locus:
    `false`
  - default_C1: `enforced` (**floor — non-configurable to off**)  ·  default_C2: `human`

- **`floor-intent-gate`** (D2)
  - what: the intent gate never fully opens — at the intent locus (A3), `C2` can never be `none`; a
    human (or, by ratchet, a human-authorized independent check) is mandatory. The one place an
    upstream that is itself *wrong* gets caught.
  - why: **"is this the right thing to build?" always has a human behind it** — the one call you never
    hand fully to the machine.
  - signature: a mandatory human approval at the intent/ratification point; no fully-automated intent
    approval; ratification recorded as a human act.
  - honored:
    - *(product)* a human ratifies at the intent gate (here: the maintainer's merge).
    - *(release)* no deploy ships a feature no human approved, however green the pipeline.
  - violated:
    - *(product)* a fully-automated pipeline ships something *technically* correct that no human
      confirmed was the *right* thing.
    - *(research)* an agent auto-adopts a conclusion no human sponsored.
  - class: `floor`  ·  mechanizable: `false`  ·  **intent_locus: `true`**
  - default_C1: `enforced` (**floor — non-configurable to off**)  ·  default_C2: `human` (never `none`)

## Acceptance criteria

- Covers all **14 assessable** slugs (A1–A4, B1–B7, B9, D1–D2 — B8 collapsed into D1, `decision-0021`);
  the two C dials are excluded by design.
- Every entry carries `what` · **`why`** · `signature` · **`honored`** · **`violated`** · `class` ·
  `mechanizable` · `default_C1` · `default_C2` (+ `intent_locus` where `true`), and `honored`/`violated`
  each carry **≥2 examples from different layers**. A missing `why` / `honored` / `violated`, or fewer
  than 2 examples, is a conformance failure (`decision-0020` meta-rule).
- Every `default_C2` on an `intent_locus: true` entry is **not** `none` (D2).
- Each `signature` is concrete enough for Assess to point at a real project tell; each `honored`/
  `violated` example reads as a plain with/without a newcomer would recognize, and the set spans layers
  (CI / spec / research / code / UI …) so the invariant reads as general (the benefits page renders a
  subset).

## Open questions

- **Structured signatures.** Signatures are prose; Assess may need them split into *mechanizable tells*
  (checkable) vs *judgment tells* (read-and-decide). Owed to the Assess build (cluster 1).
- **A-invariant default dials.** A1–A4 are admission properties checked at *ingestion*, not per-gate;
  expressing their strength as C1/C2 is a slight stretch. Fold in when the ingestion check
  (`decision-0003`) is built.
- **`why` audience.** One line serving agents-first *and* humans; revisit if the two registers pull
  apart.
- **`mechanizable` breadth (a D2 judgment call).** This catalog marks `inv-handover-points` (A2) and
  `inv-auditable-archive` (B4) `mechanizable: true` — *broader* than `spec-0002` §1's illustrative
  fragment. The call: *presence* of handover points / an append-only archive is structurally
  detectable. Flagged, not silently conformed.
