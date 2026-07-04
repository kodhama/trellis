---
id: signature-catalog-v1
type: signature-catalog
status: draft
depends_on: [invariants-v1, spec-0002]
owner: gundi
scope: trellis-product
---

> **Status `draft` — awaiting your ratification (D2).** I (the agent) authored this catalog, so I do
> not self-ratify it — the builder does not grade itself (`inv-independent-judgment` / D2). Its
> coverage is independently checked (AC1, the conformance run), but the `signature` prose and the
> `default_C1/C2` calls embody judgment that is the maintainer's to ratify.

# Signature catalog — v1 (the genome annotation)

> **What this is.** The one shipped **dictionary** of Trellis's invariants: per invariant, what it
> *is*, the observable **signature** by which a project can be seen to honor it *implicitly*, and its
> **default dials**. Schema + lifecycle: `spec-0002`. Slugs: the `invariants-v1` registry
> (`decision-0013`). This is the artifact **Assess** (#23) detects against and **tutoring** (#27)
> generates explain-cards from. `trellis-product` scope — one, shipped; not per-instance.

> **Coverage (spec-0002 §1, AC1).** Covers the **15 assessable invariants** — A structural, B
> operating (incl. **B6 `inv-self-improvement`**, restored first-class per `decision-0018`), D floors.
> It **excludes the two C dials** (`dial-enforcement-strength`, `dial-gatekeeper`): a project does not
> "honor a dial implicitly"; the dials are the *axes each entry is set along* — the **columns of a
> profile**, not rows of this catalog.

> **On `mechanizable`.** `true` marks the SCT-computable fragment (`research-0006` §Proposal) —
> structurally checkable. `false` marks a **behavioral gene** (`research-0005`) whose signature is a
> judgment tell, not a regex. Assess must treat the two differently: a `false` entry is detected by
> reading, and its `honored-implicitly` claim leans harder on `confidence` + `evidence` (`spec-0002`
> §2 evidence floor).

## Entries

### A — structural (admission gate) · class `methodology`

- **`inv-directional-flow`** (A1)
  - what: one-way stages of decreasing ambiguity (research → decisions → contracts → implementation
    → validation); downstream never consumes a draft.
  - signature: ordered stage folders or a defined pipeline; artifacts carry a stage/`status`; **no
    ratified artifact cites a draft upstream**.
  - class: `methodology`  ·  mechanizable: `true`  ·  intent_locus: `false`
  - default_C1: `enforced`  ·  default_C2: `independent-agent`

- **`inv-handover-points`** (A2)
  - what: defined transitions between stages, each a place a gate *can* attach.
  - signature: named handoffs (PR boundaries, stage transitions, review checkpoints), each with a
    before/after artifact.
  - class: `methodology`  ·  mechanizable: `true`  ·  intent_locus: `false`
  - default_C1: `enforced`  ·  default_C2: `independent-agent`

- **`inv-intent-locus`** (A3)
  - what: humans own intent/values *somewhere identifiable* — a process with no human intent locus is
    not targetable for accountable development.
  - signature: an accountable human `owner` on artifacts; a human sign-off/approval point (CODEOWNERS,
    a required human review, a ratification step).
  - class: `methodology`  ·  mechanizable: `false` (an `owner` field is checkable; *that it is a
    genuine intent locus* is judgment)  ·  **intent_locus: `true`**
  - default_C1: `enforced`  ·  default_C2: `human` (never `none` — D2)

- **`inv-ratifiable-artifacts`** (A4)
  - what: upstream can reach an **approved** state downstream consumes; outputs are **checkable
    against** it.
  - signature: a `status` lifecycle (draft → approved/ratified); artifacts with acceptance criteria a
    result can be graded against.
  - class: `methodology`  ·  mechanizable: `true`  ·  intent_locus: `false`
  - default_C1: `enforced`  ·  default_C2: `independent-agent`

### B — operating (what Trellis supplies) · class `trellis-design`

- **`inv-graph-maintenance`** (B1) *(neighbor of B6; shares `inv-prune-bias`)*
  - what: the dependency graph of artifacts **and rules** kept consistent and minimal, information
    flowing one way; trigger-driven; append-only records superseded, never edited-in-substance.
  - signature: a `depends_on` graph; supersede/retire records; dependents re-reviewed on upstream
    change; no silent downstream patches; a bias to retire rules over adding.
  - class: `trellis-design`  ·  mechanizable: `true` (the **flow** facet — no ratified consumes a
    draft; forward/backward/prune are judgment)  ·  intent_locus: `false`
  - default_C1: `enforced`  ·  default_C2: `independent-agent`

- **`inv-self-improvement`** (B6) *(restored first-class, `decision-0018`; neighbor of B1)*
  - what: the process learns from friction and gets better — improvement signals are surfaced and
    acted on, deliberately, so a glitch does not happen twice.
  - signature: a trigger format (`condition → action`) stored where it fires; improvement signals
    surfaced through the project's **chosen channel** (asked/inferred, never assumed); retirement in
    the same change; prune-bias (the trigger set does not grow monotonically).
  - class: `trellis-design`  ·  mechanizable: `false` (behavioral gene; the SI-1 surfacing floor is
    checkable against the declared channel, the proactive-notice disposition is not)  ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `human`

- **`inv-gate-at-handover`** (B2)
  - what: apply the verification gate at every A2 handover; any skip is **surfaced** (D1).
  - signature: a check/review fires at each handoff (CI gate, required review); skips are logged, not
    silent.
  - class: `trellis-design`  ·  mechanizable: `true`  ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `independent-agent`

- **`inv-independent-judgment`** (B3) *(two faces: conformance + intent)*
  - what: the assessor is independent of what it assesses — the builder does not grade itself
    (conformance face); the agent does not flatter the human (intent face).
  - signature: a verifier **distinct from the producer** (fresh-context review agent); reviews record
    dissent/risks, not reflexive assent; the verifier derives its checklist from the approved upstream.
  - class: `trellis-design`  ·  mechanizable: `false` (behavioral gene; the intent face lives in
    system prompts, weakly checkable)  ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `independent-agent`

- **`inv-auditable-archive`** (B4)
  - what: provenance + immutable decision history + consolidated current-truth.
  - signature: append-only decision records; retained change history (git); a current-truth doc kept
    separate from its change log.
  - class: `trellis-design`  ·  mechanizable: `true` (presence of archive/provenance is structural)
    ·  intent_locus: `false`
  - default_C1: `enforced`  ·  default_C2: `independent-agent`

- **`inv-bounded-context`** (B5)
  - what: each operation reads only its declared inputs, never the whole archive.
  - signature: operations/agents declare their inputs (`depends_on`, scoped context); sub-agents with
    narrow context/tool scope; an explicit observer (the dep-graph) over project state.
  - class: `trellis-design`  ·  mechanizable: `false` (is the context *genuinely* bounded? — judgment)
    ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `independent-agent`

- **`inv-minimal-first`** (B7)
  - what: the smallest process that works; add a step only when friction reveals the boundary; bias to
    retire over add.
  - signature: a deliberately small rule/process set; steps added with recorded justification; retired
    rules pruned rather than accumulated.
  - class: `trellis-design`  ·  mechanizable: `false` (a disposition)  ·  intent_locus: `false`
  - default_C1: `expressed`  ·  default_C2: `human`

- **`inv-reference-relationship`** (B8)
  - what: how much the model **adopts** one framework vs **adapts** from several is a free choice —
    but the choice, and every divergence from a reference, is recorded and surfaced.
  - signature: a recorded decision on framework sourcing; divergences from a reference captured as
    decisions, not silent drift.
  - class: `trellis-design`  ·  mechanizable: `false`  ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `human`

- **`inv-clarify-before-commit`** (B9)
  - what: ambiguity in an upstream is actively surfaced and resolved (usually by asking the human)
    before downstream consumes it; never silently guessed.
  - signature: open-questions sections; clarifying exchanges recorded before build starts; a
    `/clarify`-like step ahead of implementation.
  - class: `trellis-design`  ·  mechanizable: `false` (behavioral gene)  ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `human`

### D — floors (never configurable to "off") · class `floor`

- **`floor-transparency`** (D1)
  - what: every consequential choice is **surfaced** — a skipped gate, a missing capability, a degraded
    result, a relaxed setting; a failed verification is escalated visibly, never silently abandoned.
  - signature: skips/degradations logged and visible; loud-failure on a missing tool/source; no silent
    fallbacks or quietly-swallowed errors.
  - class: `floor`  ·  mechanizable: `false` (a disposition; partially checkable)  ·  intent_locus:
    `false`
  - default_C1: `enforced` (**floor — non-configurable to off**)  ·  default_C2: `human`

- **`floor-intent-gate`** (D2)
  - what: the intent gate never fully opens — at the intent locus (A3), `C2` can never be `none`; a
    human (or, by ratchet, a human-authorized independent check) is mandatory. The one place an
    upstream that is itself *wrong* gets caught.
  - signature: a mandatory human approval at the intent/ratification point; no fully-automated intent
    approval; ratification recorded as a human act.
  - class: `floor`  ·  mechanizable: `false`  ·  **intent_locus: `true`**
  - default_C1: `enforced` (**floor — non-configurable to off**)  ·  default_C2: `human` (never `none`)

## Acceptance criteria

- Covers all **15 assessable** slugs (A1–A4, B1–B9, D1–D2); the two C dials are excluded by design.
- Every entry carries `what` · `signature` · `class` · `mechanizable` · `default_C1` · `default_C2`
  (+ `intent_locus` where `true`).
- Every `default_C2` on an `intent_locus: true` entry is **not** `none` (D2).
- Each `signature` is concrete enough for Assess to point at a real project tell (iron rule).

## Open questions

- **Structured signatures.** Signatures are prose; Assess may need them split into *mechanizable
  tells* (checkable) vs *judgment tells* (read-and-decide), tracking the `mechanizable` flag. Owed to
  the Assess build (cluster 1).
- **A-invariant default dials.** A1–A4 are admission properties checked at *ingestion*, not per-gate;
  expressing their strength as C1/C2 is a slight stretch (they fail admission loudly if absent). Fold
  in when the ingestion check (`decision-0003`) is built.
- **B3 two-face gatekeeper.** The conformance face gates by `independent-agent`; the intent face lives
  in system prompts (`human`, weakly checkable). One entry carries one `default_C2` — revisit if a
  gene ever needs a per-face gatekeeper.
- **`mechanizable` breadth (a D2 judgment call).** This catalog marks `inv-handover-points` (A2) and
  `inv-auditable-archive` (B4) `mechanizable: true` — *broader* than `spec-0002` §1's illustrative
  fragment (A1/A4/B1-flow/B2). The call: *presence* of handover points and of an append-only archive
  is structurally detectable, so they belong in the computable fragment. This diverges from
  `research-0006` §Limits' narrower "roughly A1/A4/B1-flow/B2" — flagged as a judgment call for
  ratification, not silently conformed.
