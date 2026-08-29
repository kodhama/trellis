---
id: signature-catalog-v1
type: signature-catalog
status: ratified
depends_on: [invariants-v1, schema-typed-artifacts]
owner: gundi
scope: trellis-product
ratified: 2026-07-04
---

> **Ratified via merge (`decision-0022`).** The agent authored this; the maintainer's **merge of this
> PR is the ratification** (`floor-intent-gate`). *(As of `decision-0082` the merge is the whole
> mechanic — there is no `draft → ratified` flip; this banner formerly said the flip rode the
> reviewed diff.)* Coverage is
> independently checked (AC1); the `signature` / `why` / example / dial calls embody judgment the
> maintainer accepts by merging.
>
> *Amended in place 2026-07-12 (kodhama-0008 Lane A, kodhama/kodhama#35): `floor-intent-gate`'s
> honored `(product)` example de-merged — the propagated principle names no VCS mechanic; every
> mechanic mapping is the operational layer's (grove). One example sentence; no slug, schema, or
> dial change. `status` unchanged. Scope ruling (maintainer, 2026-07-12, trellis#149): the de-merge
> targets mechanic **mappings** — statements defining the approval act as a VCS event. The
> directive's "finalize, ship, or merge" verb list gates delivery acts and does not define the
> approval act; it stands unchanged.*

# Signature catalog — v1 (the genome annotation)

> **What this is.** The one shipped **dictionary** of Trellis's invariants: per invariant — what it
> *is*, **why** it earns its place (the goal, agents-first), the observable **signature** by which a
> project is seen to honor it, **≥2 matched `without → with` pairs** — the same use case shown failing
> then fixed (`violated[i]` and `honored[i]` are one pair, same layer tag; `decision-0027`), spanning
> different layers (CI / spec / research / code / ops …) — and its **default dials**. Schema +
> lifecycle: `schema-typed-artifacts`. Slugs: the
> `invariants-v1` registry. **The benefits page derives from the `why` + honored/violated here** — no
> claim on the page without a rule behind it (`decision-0020`). `trellis-product` scope — one, shipped.

> **Coverage (AC1).** Covers the **16 assessable invariants** — the structural set, the
> operating set (incl. **`inv-self-improvement`**, `decision-0018`, **`inv-deliberate-succession`**,
> `decision-0074`, and **`inv-no-orphan-followups`**, `decision-0078`), the floors. `inv-reference-relationship`
> was **collapsed into `floor-transparency` + the adopt/adapt dial** (`decision-0021`) — its "divergence from a framework"
> case lives in `floor-transparency`'s example below. Excludes the two dials (they are the *axes* entries are set
> along, not rows).

> **On `mechanizable`.** `true` marks the SCT-computable fragment — structurally checkable. `false`
> marks a **behavioral gene** whose signature is a judgment tell, not a regex. Every invariant carries
> a `honored`/`violated` pair, and **a change that edits an invariant without updating its examples is
> a conformance failure** (`decision-0020` meta-rule — the iron rule applied to the rule-set itself).

> **Derived resources — sync them on any change (`decision-0028`).** This catalog is the single source.
> It is **rendered** to [`docs/invariants.html`](../../docs/invariants.html) (the pairs as cards) and
> **copied verbatim** to `cli/assets/invariants.md` + `plugins/trellis/reference/invariants.md` (the
> bundled reference). Change an example here → regenerate all three. A CI check enforces it
> (`cli/sync_test.go`) — but this note is here so you see the dependents *before* the check does.
>
> **Adding or removing a row costs more than an example edit does**, and this note named only the
> three above until `decision-0078` — `decision-0074`'s review caught four obligations its author's
> sweep missed, three of them the same root cause. A **row-set change** additionally touches: the
> `invariants-v1` registry (a slug is a set amendment); the activation rows in
> `plugins/trellis/reference/rules-a.toml` / `rules-b.toml` and this repo's own `.trellis/rules.toml`
> (**without a row the rule ships but is inactive**); `plugins/trellis/hooks/codex-context.mjs`'s
> hardcoded `SLUGS`; the contract layer that states the count — this catalog's own Coverage note
> and AC1, `core/rubrics/artifact-contract.md`, the `corpus-reviewer` checklist; `profiles/trellis-self.md`;
> the prose counts in `README.md`, `plugins/trellis/README.md`, `install.sh`, `docs/index.html`,
> `docs/lp-content.md` and `docs/invariants.html` (which also needs a new card);
> `plugins/trellis/skills/remove/SKILL.md` and the needle pinning it; and the release stamp `plugins/trellis/VERSION` (**unguarded — trellis#245 is still open** — and
> without it every cached consumer keeps the old rule set, `d4a2c7b`). Prose counts go stale in more
> shapes than one grep matches: the digits, the word spelled out, the `N/N` form. Positional slug
> indices are gone from `cli/rules_test.go` — every count there now derives from one pinned list.

## Entries

### Structural — the admission gate · class `methodology`

- **`inv-directional-flow`**
  - what: one-way stages of decreasing ambiguity (research → decisions → contracts → implementation
    → validation); downstream never consumes a draft.
  - directive: Build only on settled ground — an approved spec or a made decision, never a draft that's still changing under you. If your input isn't settled, or you can't tell whether it is, ask before you build on it.
  - why: agents always build on **settled** ground — ambiguity only decreases, so no one codes against
    a spec that is still moving.
  - signature: ordered stage folders or a defined pipeline; artifacts carry a stage/`status`; **no
    ratified artifact cites a draft upstream**.
  - honored:
    - *(spec)* implementation reads only the **approved** spec; a ratified doc never depends on a draft.
    - *(research)* a synthesis cites only ratified findings, not drafts still under review.
  - violated:
    - *(spec)* an agent builds against a spec still being edited; it shifts, and the work is built on a
      version that no longer exists.
    - *(research)* a decision cites a draft finding that is later refuted — it stood on sand.
  - class: `methodology`  ·  mechanizable: `true`  ·  intent_locus: `false`
  - default_C1: `enforced`  ·  default_C2: `independent-agent`

- **`inv-handover-points`**
  - what: defined transitions between stages, each a place a gate *can* attach.
  - directive: Work in reviewable steps with clear stopping points — a plan, a spec, a PR — not one unbroken stream. Leave seams where the work can be paused and checked.
  - why: development moves in **discrete** steps with boundaries, not one fluid blur — the seams are
    where work can be stopped, inspected, and handed on. (Directional flow is *which way* work moves;
    this is *that it moves in steps*.)
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

- **`inv-intent-locus`**
  - what: humans own intent/values *somewhere identifiable* — a process with no human intent locus is
    not targetable for accountable development.
  - directive: Make sure a human owns the goal of what you're doing. Don't chase a proxy metric or ship something no human has confirmed is the right thing to build.
  - why: an **accountable human owns the goal**, so a wrong *direction* gets caught before it is built.
  - signature: an accountable human `owner` on artifacts; a human sign-off/approval point (CODEOWNERS,
    a required review, a ratification step).
  - honored:
    - *(product)* every feature traces to an accountable human `owner`; the "why" is a recorded decision.
    - *(research)* a research direction has a named human sponsor who can say "that's not what we're after."
  - violated:
    - *(product)* agents optimize a proxy metric no human owns, and ship the wrong thing efficiently.
    - *(research)* a research direction runs with no human sponsor and drifts somewhere no one intended.
  - class: `methodology`  ·  mechanizable: `false` (an `owner` field is checkable; *that it is a
    genuine intent locus* is judgment)  ·  **intent_locus: `true`**
  - default_C1: `enforced`  ·  default_C2: `human` (never `none` — `floor-intent-gate`)

- **`inv-ratifiable-artifacts`**
  - what: upstream can reach an **approved** state downstream consumes; outputs are **checkable
    against** it.
  - directive: Build against a fixed, approved target with a clear pass/fail for "done" — not a vague or moving one. If there's no agreed definition of done, get one first.
  - why: you build against a **stable, approved target with a real pass/fail criterion** — not a
    moving one, and not "looks done to me."
  - signature: a `status` lifecycle (draft → approved/ratified); artifacts with acceptance criteria a
    result can be graded against.
  - honored:
    - *(spec)* a spec reaches `ratified`, carries acceptance criteria, and work is graded against it.
    - *(data)* a schema is versioned; downstream validates against the **approved** version, not HEAD.
  - violated:
    - *(spec)* nothing is ever "final," so implementation chases a spec that keeps moving under it.
    - *(data)* there is no released schema version, so services validate against inconsistent snapshots.
  - class: `methodology`  ·  mechanizable: `true`  ·  intent_locus: `false`
  - default_C1: `enforced`  ·  default_C2: `independent-agent`

### Operating — what Trellis supplies · class `trellis-design`

- **`inv-graph-maintenance`** *(neighbor of `inv-self-improvement`; shares its prune-bias signature — the trigger/rule set does not grow monotonically)*
  - what: the dependency graph of artifacts **and rules** kept consistent and minimal, information
    flowing one way; trigger-driven; append-only records superseded, never edited-in-substance.
  - directive: When you change something, update everything that depends on it — and if you can't tell what depends on it, say so rather than assume nothing does. If you find a past decision is wrong or missing, fix the decision — don't just patch around it.
  - why: the knowledge base **stays coherent for the agents reading it**, and a discovery deep in the
    code **repairs the decision that should have known it** (backprop) — which happens more than people
    admit.
  - signature: a `depends_on` graph — better, **forward-edges too** (a source names what derives from
    it, so a change surfaces its dependents; `decision-0028`); supersede/retire records; dependents
    re-reviewed on upstream change; no silent downstream patches; a **sync guard per source→derivative
    pair**; a bias to retire rules over adding; **one home per kind of information**, placed by which
    consumer must trip over it — a copy elsewhere points at the home, never carries the truth
    (`decision-0040`); **tests name their upstream** (a spec anchor or a defect id), and a test↔spec
    conflict is repaired deliberately, never silently.
  - honored:
    - *(docs)* a repaired decision re-reviews its specs → plans → code, in turn.
    - *(research)* a downstream finding that contradicts an upstream note updates the *note*, not just
      the finding (backprop).
    - *(ops)* each kind of information has exactly **one home**, chosen by which consumer must trip
      over it; recorded elsewhere first (a chat thread, a meeting note), it lands in its home before
      anything downstream consumes it, and the other copy just points.
    - *(code)* every test names the upstream it guards (a spec anchor or a defect id); when a test and
      the spec disagree, the conflict is surfaced and resolved deliberately — the spec gains its
      missing invariant, or the over-pinning test is retired, citing why. A regression test is never
      weakened just to make a reading of the spec pass.
  - violated:
    - *(docs)* a decision changes but its dependent specs are never updated — they silently diverge.
    - *(research)* a finding contradicts an upstream note, but only the finding is recorded — the note
      stays wrong and agents keep reading it.
    - *(ops)* the same plan lives in a tracker *and* a wiki — the copies diverge, and the agents that
      should trip over the current state are reading the stale one; a parked item sits in a channel
      its executor never reads, trips nothing, and rots.
    - *(code)* a regression test blocks a convenient reading of the spec, so it is quietly deleted —
      the defect it pinned ships again; no one can say which requirement any test guards.
  - class: `trellis-design`  ·  mechanizable: `true` (the **flow** facet; forward/backward/prune are
    judgment)  ·  intent_locus: `false`
  - default_C1: `enforced`  ·  default_C2: `independent-agent`

- **`inv-self-improvement`** *(restored first-class, `decision-0018`; neighbor of `inv-graph-maintenance`; the entropy lean it briefly carried now homes in `inv-deliberate-succession`, `decision-0074`)*
  - what: the process learns from friction and gets better — improvement signals are surfaced and
    acted on, deliberately, so a glitch does not happen twice.
  - directive: When something breaks or causes friction, fix the root cause so it can't happen twice — don't just re-run it and move on.
  - why: **a process glitch never happens twice** — friction becomes a fix, not a recurring tax.
  - signature: a trigger format (`condition → action`) stored where it fires; improvement signals
    surfaced through the project's **chosen channel** (asked/inferred, never assumed); retirement in
    the same change; prune-bias (the trigger set does not grow monotonically). *(The entropy lean —
    signals thrown off where new work meets existing stock — moved to `inv-deliberate-succession`,
    `decision-0074`; SI-1 still governs how any such signal reaches the channel.)*
  - honored:
    - *(CI)* a flaky test recurs → a trigger is filed, the root cause fixed, the trigger retired in the
      same change.
    - *(process)* a repeated review miss becomes a checklist item that rides the PR you already write.
  - violated:
    - *(CI)* the same pipeline step fails weekly and everyone just re-runs it, forever.
    - *(process)* a PR raises the same open question every time, with no follow-up, and it rots unowned.
  - class: `trellis-design`  ·  mechanizable: `false` (the surfacing floor — improvement signals reach the
    declared channel — is checkable; the proactive-notice disposition is not)  ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `human`

- **`inv-deliberate-succession`** *(new, `decision-0074`; neighbor of `inv-self-improvement` — that one adapts the process to friction, this one decides the boundary between new work and what came before)*
  - what: wherever new work meets an existing base — code, conventions, decisions, data, docs — the
    boundary is **decided out loud in both directions**: what the new pattern leaves outside it, and
    what the old base quietly supplies to a gap in the new design. The prior version is evidence to
    weigh, never gravity to drift with and never debris to step over.
  - directive: When you introduce a new pattern, say what the existing stock outside it should become — migrate it, or name the exemption and ask; never resolve it silently in prose. And when your new design needs a value, shape, or default that the old one already supplies, say why it fits the new model or choose differently — reuse is often right, reuse by default is not.
  - why: **the new thing is actually new, and the old thing is actually decided** — otherwise a
    superseded convention lingers beside its replacement, and an inherited constant smuggles in the
    model it was calibrated for, each exempted by prose nobody approved.
  - signature: pattern-introducing changes throw off a retrofit-scope signal (what now sits outside
    the pattern?), surfaced through the project's channel — asked, not assumed (SI-1 at
    change-introduction time); gap-filling reuse stated with its justification against the new model
    rather than its provenance; inherited values named as parameters with the old calibration's
    context recorded; the superseded version reachable and marked superseded, never edited in
    substance or silently dropped.
  - honored:
    - *(structure)* a change that introduces a pattern ships with its retrofit question — "the old
      suite now sits outside this convention: migrate or exempt?" — and the human rules on it.
    - *(design)* a new model reuses a threshold from the one it replaces, and names the property of
      the new model that keeps the old value apt.
    - *(docs)* a rewritten guide states which claims of the previous version it carries forward and
      which it drops, so a reader who knew the old one can tell what changed.
    - *(metrics)* a success signal is defined against the outcome the new model is meant to produce,
      so the current implementation's behaviour can fail it.
  - violated:
    - *(structure)* a new directory convention lands while the old stock stays loose beside it,
      exempted by confident boundary prose nobody approved — two conventions in one tree.
    - *(design)* a new two-threshold model needs a lower bound; the constant from the retired model
      fills it unexamined, carrying a *streak* shape into a design whose other threshold is
      *coverage* — two thresholds, silently two different measures.
    - *(docs)* a rewritten spec silently keeps a constraint whose original rationale no longer
      applies; nobody can tell whether it was re-decided or merely retyped.
    - *(metrics)* a success signal is defined against what the current implementation already does,
      so the experiment passes on the untreated baseline and cannot fail.
  - class: `trellis-design`  ·  mechanizable: `false` (the surfacing floor is checkable per SI-1; noticing that a moment is one of succession is not)  ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `human`

- **`inv-no-orphan-followups`** *(new, `decision-0078`; neighbor of `inv-self-improvement` — that one
  turns friction into a fix, this one governs the work you decide **not** to do now)*
  - what: a deferral is recorded only where a **named consumer** will re-present it — a queue an
    executor works, a tracker someone triages, a failing or skipped test, a scheduled sweep that is
    actually switched on. "Someone reading this file later" is not a consumer. Work with no such
    address is not deferred work: do it now, **drop it on the record**, or escalate it.
  - directive: When you defer something, put it where a named consumer will bring it back — a queue, a tracker, a failing test, a scheduled sweep. If you cannot name what will surface it again, do not record it: do it now, drop it and say you dropped it, or escalate. Writing it down is not tracking it.
  - why: **a follow-up you wrote down actually comes back** — agents produce deferrals far faster
    than humans consume them, and a plausible ledger is worse than none: it converts an unsolved
    problem into the *feeling* of a tracked one, and it is trusted precisely because it looks like
    bookkeeping.
  - signature: each deferral sink names its consumer and the cadence that drains it; a step that
    **appends** to a sink names the step that **reads** it, and that step is switched on; write-only
    sinks are detectable — entries far outnumbering resolutions, or a designed consumer never turned
    on; *"dropped, because X"* is a first-class recorded outcome beside *"deferred to Y"*, so the
    cheap escape is dropping rather than filing; sinks are pruned, not grown (the prune-bias hinge
    shared with `inv-self-improvement` / `inv-graph-maintenance`).
  - honored:
    - *(process)* a review's deferred findings go to the queue whose executor actually works it, and
      the ones nobody will work are dropped in the review itself, with the reason recorded.
    - *(code)* a known-broken case ships as a skipped test naming the defect — every run re-presents
      it, so it cannot be forgotten.
    - *(ops)* a workflow step that appends findings to a ledger names the step that drains it, and
      that step runs on a stated cadence; the write→read pair is guarded like any other.
  - violated:
    - *(process)* a review's findings are appended to a file nobody is scheduled to read; the real
      defects get fixed weeks later only because a fresh review re-derives them.
    - *(code)* a `TODO: fix before launch` sits in a comment that no test, lint rule or checklist
      ever surfaces, and ships.
    - *(ops)* a workflow appends every deferral to a ledger and is told not to check for duplicates;
      the sweep designed to drain it was never switched on — 61 entries, 3 resolutions, all three
      written the day someone audited it.
  - class: `trellis-design`  ·  mechanizable: `false` (that a sink *declares* a consumer is
    checkable; that anything actually drains it is judgment)  ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `human`

- **`inv-gate-at-handover`**
  - what: apply the verification gate at every handover point; any skip is **surfaced** (`floor-transparency`).
  - directive: Don't skip the review or verification step before handing work on. If you have to skip it, say so out loud — never let it silently not happen.
  - why: the review **actually fires** (not quietly skipped under deadline) — and if it is skipped,
    you can *see* it was.
  - signature: a check/review fires at each handoff (CI gate, required review); skips are logged, not
    silent.
  - honored:
    - *(CI)* a conformance check + review fire on every PR; a deliberate skip is *recorded*, not hidden.
    - *(release)* a promotion gate blocks anything that didn't pass the prior stage's sign-off.
  - violated:
    - *(CI)* the review is "optional," so under deadline it silently doesn't happen and a defect ships.
    - *(release)* an artifact is promoted straight to prod with no gate, and no record the check was skipped.
  - class: `trellis-design`  ·  mechanizable: `true`  ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `independent-agent`

- **`inv-independent-judgment`** *(two faces: conformance + intent)*
  - what: the assessor is independent of what it assesses — the builder does not grade itself
    (conformance face); the agent does not flatter the human (intent face).
  - directive: Don't rule your own work correct — tell the human an independent review is needed and let someone (or something) other than the author check it. Don't just agree to please the human; say what you actually think, problems included. And before calling a thing right *or* wrong — especially when your verdict matches what the human just suggested — verify it against the source: quote it, run the obvious counter-checks, and separate what it says from what you infer.
  - why: **the builder doesn't grade its own homework**, and the agent **names the risk** instead of
    flattering the plan — so verification and the intent gate are real, not decorative.
  - signature: a verifier **distinct from the producer** (fresh-context review agent); reviews record
    dissent/risks, not reflexive assent; the verifier derives its checklist from the approved upstream;
    **verdicts cite the source they judged** (the quoted line, the counter-check run), with fact
    separated from inference (`decision-0040`).
  - honored:
    - *(review)* a read-only reviewer, distinct from the author, derives its own checklist and reports
      what's wrong — even when inconvenient.
    - *(research)* findings are adversarially verified by a separate pass, not self-certified.
    - *(collab)* before agreeing or disagreeing with the human's hunch — *especially* when the verdict
      would match it — the agent opens the file, quotes the actual lines, runs the obvious
      counter-checks, and labels what is fact and what is inference; "I can't confirm this" is said
      out loud when it can't.
  - violated:
    - *(review)* the agent that wrote the code reviews its own code and decides it's good.
    - *(research)* an agent certifies its own findings (or agrees with a flawed plan to please you), and
      it sails through.
    - *(collab)* the human suggests the old code is wrong; the agent answers "you're right, good catch"
      without opening it — the code was fine, and the plausible agreement ships a bug.
  - class: `trellis-design`  ·  mechanizable: `false` (the intent face lives in system prompts, weakly
    checkable)  ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `independent-agent`

- **`inv-auditable-archive`**
  - what: provenance + immutable decision history + consolidated current-truth.
  - directive: Record why decisions are made and keep that history — don't edit past decisions in place and lose the reasoning. "Why is it this way?" should be answerable later.
  - why: you can always answer **"why is it this way?"** — decisions are not lost or quietly rewritten.
  - signature: append-only decision records; retained change history (git); a current-truth doc kept
    separate from its change log; **supersession can be partial** — a superseded-in-part pointer marks
    the outgrown half so the live remainder stays navigable and no reader lands on stale text without
    a forward link (`decision-0040`).
  - honored:
    - *(ADR)* decisions are append-only and link their rationale; superseding writes a *new* record.
    - *(infra)* every prod change carries provenance — git history + a current-truth doc — so "why" is
      always answerable.
  - violated:
    - *(ADR)* a decision is edited in place and its rationale lost; months later it is re-litigated
      from scratch.
    - *(infra)* an undocumented prod change, so no one can say why it is configured this way.
  - class: `trellis-design`  ·  mechanizable: `true` (presence of archive/provenance is structural)  ·
    intent_locus: `false`
  - default_C1: `enforced`  ·  default_C2: `independent-agent`

- **`inv-bounded-context`**
  - what: each operation reads only its declared inputs, never the whole archive.
  - directive: Pull in only the inputs the task actually needs — don't dump the whole codebase into context. If you're unsure what's relevant, ask rather than grabbing everything.
  - why: agents decide on **sharp, relevant context** instead of drowning in everything — better calls,
    and it scales as the archive grows.
  - signature: operations/agents declare their inputs (`depends_on`, scoped context); sub-agents with
    narrow context/tool scope; an explicit observer (the dep-graph) over project state.
  - honored:
    - *(agent)* a sub-agent gets exactly its declared inputs and decides crisply within them.
    - *(data)* a query reads a scoped view, not the whole warehouse.
  - violated:
    - *(agent)* an op dumps the entire repo into context, dilutes the signal, and decides on noise.
    - *(data)* a query scans the whole warehouse, drowns in irrelevant rows, and gets slower as it grows.
  - class: `trellis-design`  ·  mechanizable: `false` (is the context *genuinely* bounded? — judgment)
    ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `independent-agent`

- **`inv-minimal-first`**
  - what: the smallest process that works; add a step only when friction reveals the boundary; bias to
    retire over add.
  - directive: Prefer the smallest thing that works. Don't add process, tooling, or abstraction until it's clearly needed — lean toward removing over adding.
  - why: **no ceremony for its own sake** — every step has earned its place, so the process stays light.
  - signature: a deliberately small rule/process set; steps added with recorded justification; retired
    rules pruned rather than accumulated.
  - honored:
    - *(process)* a deliberately tiny rule set; a new step lands only with a recorded reason.
    - *(tooling)* a build config with no unused steps; a dependency pulled in only when needed.
  - violated:
    - *(process)* a heavyweight methodology copied wholesale, most steps cargo-culted.
    - *(tooling)* a whole framework pulled in to use one function; the build carries steps no one remembers.
  - class: `trellis-design`  ·  mechanizable: `false` (a disposition)  ·  intent_locus: `false`
  - default_C1: `expressed`  ·  default_C2: `human`

- **`inv-clarify-before-commit`**
  - what: ambiguity in an upstream is actively surfaced and resolved (usually by asking the human)
    before downstream consumes it; never silently guessed.
  - directive: If a requirement, spec, or input is ambiguous, ask before you build — don't quietly pick one reading and risk building the wrong thing.
  - why: **agents ask instead of guessing** — you don't discover three files later that they took the
    wrong reading of a vague spec.
  - signature: open-questions sections; clarifying exchanges recorded before build starts; a
    `/clarify`-like step ahead of implementation.
  - honored:
    - *(spec)* an agent flags a vague requirement and resolves it *before* coding.
    - *(data)* an ambiguous metric definition is confirmed with a human before the dashboard is built.
  - violated:
    - *(spec)* an agent silently picks one reading of a vague spec, builds it, and it's the wrong one.
    - *(data)* an ambiguous metric definition is guessed, and the dashboard is subtly, confidently wrong.
  - class: `trellis-design`  ·  mechanizable: `false` (behavioral gene)  ·  intent_locus: `false`
  - default_C1: `default-on-but-skippable`  ·  default_C2: `human`

### Floors — never configurable to "off" · class `floor`

- **`floor-transparency`** *(absorbs the framework-divergence case from retired `inv-reference-relationship`, `decision-0021`)*
  - what: every consequential choice is **surfaced** — a skipped gate, a missing capability, a degraded
    result, a relaxed setting, a divergence from a framework you claim to follow; a failed verification
    is escalated visibly, never silently abandoned. **Drift is allowed, but never silent.**
  - directive: Say every consequential choice out loud — a skipped step, a missing capability, a degraded result, a workaround, a place you diverged from the plan. Never quietly work around a problem.
  - why: **nothing consequential happens silently** — you learn about the shortcut *when it is taken*,
    not when it breaks in production.
  - signature: skips/degradations logged and visible; loud-failure on a missing tool/source; no silent
    fallbacks; divergence from a reference captured as a recorded decision; **bounded runs checkpoint
    and resume** — resumable state left behind, a successor continuing rather than restarting,
    bounded auto-resumes, and the bound itself a loud demand for human attention (`decision-0040`).
  - honored:
    - *(framework)* "we diverge from Spec Kit here" is a recorded decision, so an agent knows where you
      follow the book and where you don't.
    - *(ops)* a degraded fallback (cache miss, retry, downgrade) is surfaced, not swallowed.
    - *(agent)* every bounded run leaves resumable state — the work pushed, a note of what's left — so
      a successor continues instead of restarting; auto-resumes are bounded, and hitting the bound
      produces a loud demand for human attention, never a quiet dead-end.
  - violated:
    - *(framework)* the team quietly drifts from the methodology it *claims* to follow.
    - *(ops)* an agent silently falls back to a degraded path or swallows an error — you learn when it
      breaks in prod.
    - *(agent)* a run dies at its turn cap mid-task and nothing marks where it stopped — the next run
      redoes the work from zero, or the half-finished state sits unnoticed until something breaks.
  - class: `floor`  ·  mechanizable: `false` (a disposition; partially checkable)  ·  intent_locus:
    `false`
  - default_C1: `enforced` (**floor — non-configurable to off**)  ·  default_C2: `human`

- **`floor-intent-gate`**
  - what: the intent gate never fully opens — at the intent locus (`inv-intent-locus`), `C2` can never be `none`; a
    human (or, by ratchet, a human-authorized independent check) is mandatory. The one place an
    upstream that is itself *wrong* gets caught.
  - directive: Never finalize, ship, or merge something a human is meant to approve without that approval. When you reach such a point, stop and get sign-off. Unsure whether a human must approve? Assume yes.
  - why: **"is this the right thing to build?" always has a human behind it** — the one call you never
    hand fully to the machine.
  - signature: a mandatory human approval at the intent/ratification point; no fully-automated intent
    approval; ratification recorded as a human act.
  - honored:
    - *(product)* a human ratifies at the intent gate — the approval recorded as a human intent act, whatever mechanism performs it.
    - *(release)* no deploy ships a feature no human approved, however green the pipeline.
  - violated:
    - *(product)* a fully-automated pipeline ships something *technically* correct that no human
      confirmed was the *right* thing.
    - *(release)* a deploy auto-promotes a change no human approved, on green tests alone.
  - class: `floor`  ·  mechanizable: `false`  ·  **intent_locus: `true`**
  - default_C1: `enforced` (**floor — non-configurable to off**)  ·  default_C2: `human` (never `none`)

## Acceptance criteria

- Covers all **16 assessable** slugs (the four structural, the ten remaining operating, the two floors — `inv-reference-relationship` collapsed into `floor-transparency`, `decision-0021`);
  the two dials are excluded by design.
- Every entry carries `what` · **`directive`** · **`why`** · `signature` · **`honored`** · **`violated`** · `class` ·
  `mechanizable` · `default_C1` · `default_C2` (+ `intent_locus` where `true`), and `honored`/`violated`
  are **≥2 matched pairs** — `violated[i]` and `honored[i]` share a use case (same layer tag, same
  order), forming one *without → with* pair (`decision-0027`). A missing `why` / `honored` / `violated`,
  fewer than 2, or an unaligned pair is a conformance failure (`decision-0020` meta-rule).
- Every `default_C2` on an `intent_locus: true` entry is **not** `none` (`floor-intent-gate`).
- Each `signature` is concrete enough for Assess to point at a real project tell; each pair reads as
  the *same situation broken then fixed* a newcomer would recognize, and the pairs span layers
  (CI / spec / research / code / ops …) so the invariant reads as general (the benefits page renders
  these as contrastive cards, `decision-0027`).

## Open questions

- **Structured signatures.** Signatures are prose; Assess may need them split into *mechanizable tells*
  (checkable) vs *judgment tells* (read-and-decide). Owed to the Assess build (cluster 1).
- **Structural-invariant default dials.** The four structural invariants are admission properties checked at *ingestion*, not per-gate;
  expressing their strength as per-gate dials is a slight stretch. Fold in when the ingestion check
  (`decision-0003`) is built.
- **`why` audience.** One line serving agents-first *and* humans; revisit if the two registers pull
  apart.
- **`mechanizable` breadth (an intent-gate judgment call).** This catalog marks `inv-handover-points` and
  `inv-auditable-archive` `mechanizable: true` — *broader* than the catalog's own illustrative
  fragment. The call: *presence* of handover points / an append-only archive is structurally
  detectable. Flagged, not silently conformed.
