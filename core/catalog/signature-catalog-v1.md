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
> project is seen to honor it, a contrastive **honored / violated** pair (what it looks like *with* vs
> the failure *without*), and its **default dials**. Schema + lifecycle: `spec-0002`. Slugs: the
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
  - honored: implementation reads an *approved* spec; a ratified doc never depends on a draft.
  - violated: an agent codes against a spec that is still being edited; the spec shifts; the work is
    now built on a version that no longer exists.
  - class: `methodology`  ·  mechanizable: `true`  ·  intent_locus: `false`
  - default_C1: `enforced`  ·  default_C2: `independent-agent`

- **`inv-handover-points`** (A2)
  - what: defined transitions between stages, each a place a gate *can* attach.
  - why: development moves in **discrete** steps with boundaries, not one fluid blur — the seams are
    where work can be stopped, inspected, and handed on. (A1 is *which way* work moves; this is *that
    it moves in steps*.)
  - signature: named handoffs (PR boundaries, stage transitions, review checkpoints), each with a
    before/after artifact.
  - honored: a plan is "done," a spec "approved," a change "ready" — defined seams you can pause at.
  - violated: vibe-coding melts prompt → code → prompt into a continuous stream with **no seam** to
    inspect or gate; nothing can be checked because nothing is ever "handed on."
  - class: `methodology`  ·  mechanizable: `true`  ·  intent_locus: `false`
  - default_C1: `enforced`  ·  default_C2: `independent-agent`

- **`inv-intent-locus`** (A3)
  - what: humans own intent/values *somewhere identifiable* — a process with no human intent locus is
    not targetable for accountable development.
  - why: an **accountable human owns the goal**, so a wrong *direction* gets caught before it is built.
  - signature: an accountable human `owner` on artifacts; a human sign-off/approval point (CODEOWNERS,
    a required review, a ratification step).
  - honored: every artifact names an `owner`; direction is a recorded human decision.
  - violated: no one owns the intent; agents optimize a proxy and no human catches that the *goal
    itself* was wrong until it ships.
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
  - honored: a spec reaches `ratified`, carries acceptance criteria, and work is checked against it.
  - violated: nothing is ever "approved"; downstream builds against a moving target and there is no
    criterion to verify the result against.
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
  - honored: a repaired ADR re-reviews its specs → plans → code; a build uncovering a missing decision
    *creates the ADR* instead of patching around it.
  - violated: a downstream discovery contradicts an upstream doc, but the doc is never updated —
    knowledge forks, and agents keep reading a truth that is no longer true.
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
  - honored: a repeated failure surfaces a trigger → the rule is fixed and the trigger retired in the
    same change; the agent notices a signal and offers *both* the fix and a standing guard.
  - violated: the same glitch recurs every few weeks because nothing captured the lesson; a PR raises
    open questions with no follow-up and they rot.
  - class: `trellis-design`  ·  mechanizable: `false` (the SI-1 surfacing floor is checkable against the
    declared channel; the proactive-notice disposition is not)  ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `human`

- **`inv-gate-at-handover`** (B2)
  - what: apply the verification gate at every A2 handover; any skip is **surfaced** (D1).
  - why: the review **actually fires** (not quietly skipped under deadline) — and if it is skipped,
    you can *see* it was.
  - signature: a check/review fires at each handoff (CI gate, required review); skips are logged, not
    silent.
  - honored: a conformance check + review run on every PR; a deliberate skip is recorded, not hidden.
  - violated: the review is "optional," so under deadline it silently does not happen and a defect
    ships with no record that the gate was bypassed.
  - class: `trellis-design`  ·  mechanizable: `true`  ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `independent-agent`

- **`inv-independent-judgment`** (B3) *(two faces: conformance + intent)*
  - what: the assessor is independent of what it assesses — the builder does not grade itself
    (conformance face); the agent does not flatter the human (intent face).
  - why: **the builder doesn't grade its own homework**, and the agent **names the risk** instead of
    flattering the plan — so verification and the intent gate are real, not decorative.
  - signature: a verifier **distinct from the producer** (fresh-context review agent); reviews record
    dissent/risks, not reflexive assent; the verifier derives its checklist from the approved upstream.
  - honored: a read-only reviewer, separate from the producer, derives its own checklist and reports
    what is wrong — even when that is inconvenient.
  - violated: the agent that wrote the code decides the code is good; or it agrees with a flawed plan
    to please you, and the mistake sails through.
  - class: `trellis-design`  ·  mechanizable: `false` (the intent face lives in system prompts, weakly
    checkable)  ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `independent-agent`

- **`inv-auditable-archive`** (B4)
  - what: provenance + immutable decision history + consolidated current-truth.
  - why: you can always answer **"why is it this way?"** — decisions are not lost or quietly rewritten.
  - signature: append-only decision records; retained change history (git); a current-truth doc kept
    separate from its change log.
  - honored: decisions are append-only and link their rationale; superseding writes a new record, never
    edits the old one.
  - violated: a decision is edited in place and its rationale lost; months later no one can say why the
    architecture is the way it is, so it gets re-litigated from scratch.
  - class: `trellis-design`  ·  mechanizable: `true` (presence of archive/provenance is structural)  ·
    intent_locus: `false`
  - default_C1: `enforced`  ·  default_C2: `independent-agent`

- **`inv-bounded-context`** (B5)
  - what: each operation reads only its declared inputs, never the whole archive.
  - why: agents decide on **sharp, relevant context** instead of drowning in everything — better calls,
    and it scales as the archive grows.
  - signature: operations/agents declare their inputs (`depends_on`, scoped context); sub-agents with
    narrow context/tool scope; an explicit observer (the dep-graph) over project state.
  - honored: a sub-agent gets exactly its declared inputs and decides crisply within them.
  - violated: an operation dumps the whole repo into context, dilutes the signal, decides on noise (or
    blows the token budget), and gets *worse* as the project grows.
  - class: `trellis-design`  ·  mechanizable: `false` (is the context *genuinely* bounded? — judgment)
    ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `independent-agent`

- **`inv-minimal-first`** (B7)
  - what: the smallest process that works; add a step only when friction reveals the boundary; bias to
    retire over add.
  - why: **no ceremony for its own sake** — every step has earned its place, so the process stays light.
  - signature: a deliberately small rule/process set; steps added with recorded justification; retired
    rules pruned rather than accumulated.
  - honored: the rule set is deliberately tiny; a new step lands only with a recorded reason.
  - violated: a heavyweight process is copied wholesale; agents spend more time on ritual than work,
    and no one remembers why half the steps exist.
  - class: `trellis-design`  ·  mechanizable: `false` (a disposition)  ·  intent_locus: `false`
  - default_C1: `expressed`  ·  default_C2: `human`

- **`inv-clarify-before-commit`** (B9)
  - what: ambiguity in an upstream is actively surfaced and resolved (usually by asking the human)
    before downstream consumes it; never silently guessed.
  - why: **agents ask instead of guessing** — you don't discover three files later that they took the
    wrong reading of a vague spec.
  - signature: open-questions sections; clarifying exchanges recorded before build starts; a
    `/clarify`-like step ahead of implementation.
  - honored: an agent flags a vague requirement and resolves it before coding; open questions are
    captured, not glossed.
  - violated: an agent silently picks one interpretation of an ambiguous spec, builds it, it is the
    wrong one, and the work is thrown away.
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
  - honored: a skipped gate is logged; a missing tool halts loudly; "we diverge from Spec Kit here" is
    a recorded decision, so an agent knows where you follow the book and where you don't.
  - violated: an agent silently falls back to a degraded path, skips a check, or quietly drifts from the
    methodology you *say* you follow — and you find out only when it breaks.
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
  - honored: a human ratifies at the intent gate (here: the maintainer's merge); no pipeline ships
    intent end-to-end without that approval.
  - violated: the whole pipeline is automated end-to-end; agents ship something *technically* correct
    that no human ever confirmed was the *right* thing to build.
  - class: `floor`  ·  mechanizable: `false`  ·  **intent_locus: `true`**
  - default_C1: `enforced` (**floor — non-configurable to off**)  ·  default_C2: `human` (never `none`)

## Acceptance criteria

- Covers all **14 assessable** slugs (A1–A4, B1–B7, B9, D1–D2 — B8 collapsed into D1, `decision-0021`);
  the two C dials are excluded by design.
- Every entry carries `what` · **`why`** · `signature` · **`honored`** · **`violated`** · `class` ·
  `mechanizable` · `default_C1` · `default_C2` (+ `intent_locus` where `true`). A missing `why` /
  `honored` / `violated` is a conformance failure (`decision-0020` meta-rule).
- Every `default_C2` on an `intent_locus: true` entry is **not** `none` (D2).
- Each `signature` is concrete enough for Assess to point at a real project tell; each `honored`/
  `violated` reads as a plain with/without a newcomer would recognize (the benefits page renders these).

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
