# Trellis — operating method (seed)

> **What Trellis is.** A shippable, portable pack that supervises an *agentic software-
> development process* — it fits, teaches, adapts, and guards whatever methodology a
> project uses, while enforcing a small set of invariants. It is **not** a process; it is
> the layer above the steps. See `agentic-dev-meta-layer-brief.md` (in this repo) for the
> full thesis and `core/invariants/trellis-invariants-v1.md` for the load-bearing core (ratified).

> **We build Trellis with Trellis.** This file is a deliberately tiny instance of the seed
> operating method (brief §12). It dogfoods our own invariants from commit one. Friction
> we hit while following it *is product research* — record it, don't route around it.

> **Which layer is this? (decision `0005`, ratified; reorg underway).** Trellis self-hosts, so
> two layers must not be conflated: **Trellis-core** (the shippable product — invariants, spine,
> gates) now lives in **`core/`**; **the methodology used to build Trellis** is the repo root
> (this file, `decisions/`, `research/`, `specs/`). *This file is Layer B — instance #1, the
> first methodology Trellis supervises — not Trellis's product agent-instructions.* The `core/`
> migration is incremental: `invariants/` moved in first; the corpus-reviewer sub-agent's product
> home (`core/agents/`) waits on the delivery slice (`0012`).

## The iron rule (most important design constraint)

Trellis must **always ground out in concrete, project-specific artifacts** — a real
instructions file, real gates, real sub-agents. If it ever just *describes* process
instead of *producing and enforcing* it, it has failed. Same rule applies to this repo:
prefer producing a checkable artifact over writing about one. **And it applies to our own
rules:** every invariant or abstract instruction carries ≥1 concrete example (few-shot) — *a
rule you can't exemplify is probably vaporware.*

## Operating method

- **Artifacts.** Every non-code document opens with frontmatter
  (`id / type / status / depends_on / owner`). Statuses: `draft → gated → approved
  (→ superseded)` — the family lifecycle (`decision-0042`); artifacts ratified before
  2026-07-08 read `ratified`, the same state as `approved` under `decision-0037`'s
  equivalence. **Downstream consumes only gated/approved upstream, never drafts.** Required body
  sections are **per-type** (not a blanket rule — a strategic decision has no "acceptance
  criteria"; ratification *is* its acceptance):
  - `decision` → `## Context` / `## Decision` / `## Consequences`
  - `spec` / `invariant-set` → `## Acceptance criteria` / `## Open questions`
  - `research-note` → `## Open questions` (+ sources & confidence tags)
  - `feedback` → exempt (advisory rubric, never a gate)
- **`owner: agent` mapping (`decision-0037` point 3).** Where an artifact sets `owner: agent`
  (`decision-0042`, `spec-0005`), the field carries **authorship**, not the accountable human —
  that role stays the maintainer (gundi), held via the merge gate (`decision-0022`). Declared
  here because `decision-0037` permits the mapping only when a methodology declares it for
  itself.
- **Decisions.** Significant choices get an **append-only** record in `decisions/`. You
  *supersede* (with a forward pointer), never edit, a ratified decision. The four strategic
  forks (brief §9) are records `0001–0004`.
- **Gates.** Human approval at the **intent** layer (vision, decisions, the invariant
  set). **Independent verification** at the **execution** layer (a conformance check
  against the approved upstream before merge — *the builder does not grade itself*).
  **Ratification is a human intent act (`decision-0022`/`0042`, refined by `decision-0046`):**
  a human's approval — in conversation, review, or by merging — is the ratification act;
  flipping `draft → gated → approved` in the PR **records** it. An **in-PR `approved` flip is
  legitimate when it records a human act**; an agent writing `approved` with no human act is
  forbidden (`floor-intent-gate`). Merging is one way to perform ratification, not the only
  one. **No draft is left on `main` past the PR that introduced it** (gate it, or keep the PR
  clearly WIP) — the `ratify-guard` check still enforces this draft-landing rule.
- **Work.** One logical change per PR; descriptive, linear history; diffs small enough to
  review on a phone.
- **Self-improvement.** Triggers, not vigilance (invariant 8): when friction reveals a
  missing rule, add it *where it fires*, **prefer retiring to adding**, keep it subordinate
  to the work. This file is the first trigger home.
- **Derived resources stay in sync (`decision-0028`).** When you change a *source* — the
  catalog, a spec, the CLI's command set — update everything that **derives** from it in the
  same change. A source names its derivatives (so you see them at the edit), and a check
  guards each pair. If you edit something and can't name what derives from it, that's the
  question to ask. (This is `inv-graph-maintenance` made salient: the graph, pointed forward.)
> The invariants this section used to restate — **transparency** (surface everything; fail loudly;
> never emit plausible-but-unverified output) and **independent judgment** (no sycophancy; the builder
> doesn't grade itself) — now arrive through Trellis's self-applied overlay (`.trellis/`), not
> hand-written here. `CLAUDE.md` loads that overlay for Claude; Codex delivery is separate plugin
> work. This section is the project's *method* for holding the invariants, not a copy of them
> (`decision-0035`). If a behavior below reads like a bare invariant, it belongs in the overlay,
> not here.

## Naming guardrail (research discipline, applied to ourselves)

If we ever name the invariant set authoritatively, attribute it clearly as **our
synthesis** — never imply pre-existing provenance. For now it is exactly *"Trellis's
invariants — our synthesis, v1."* Eponymous framing is a deliberate *later* decision, made
only once the set's durability is proven across multiple instances.

## Current state

- **Intent layer: ratified.** `invariants-v1` is the ratified current-truth set (the
  structural admission gate · the operating set · the dials · the floors); decisions
  `0001–0008` are ratified; v0 superseded.
- **Research done:** Steps 0–2 (`research-0001` target landscape, `research-0002` gate-test
  of real frameworks); findings folded into v1.
- **Machinery:** automated PR review live (decision `0007`).
- **Next:** the **spine** — portable artifact contract + lifecycle (brief §8.1) — the first
  build, consuming ratified `invariants-v1`. Then find **instance #2** (the N=1 risk).

## Acceptance criteria

- A newcomer (human or agent) can read this file and the invariant set and know how to
  make a change that will pass the gates.
- Every claim of "done" in this repo traces to a concrete artifact, not a description.

## Open questions

- When does a second project (instance #2) get to test these invariants, given we have
  chosen to validate by dogfooding our own project first (decision `0001`)?
- What is the smallest enforcement that makes "downstream consumes only ratified" real
  here — convention, a check, or a gate sub-agent? (Resolve when the spine is built.)

## Maintaining project instructions

`AGENTS.md` is the canonical home for shared project instructions. Edit new shared rules
here, outside managed blocks. `CLAUDE.md` is the Claude adapter, not a shared-rule edit
surface. Genuinely Claude-only rules belong in `.claude/rules/`.

Grove and Trellis project choices remain in `.grove/` and `.trellis/` configuration files.
Do not hand-edit managed blocks.

<!-- grove:begin (managed by grove — dials live in .grove/, not this block) -->
trellis is a **grove consumer** ([grove](https://github.com/kodhama/grove) adr-0026, the
thin-vendor boundary): work items matching a grove workflow (W1–W6 — e.g. a bug report → the
bug pipeline, a research ask → divergent research) run as grove runs, sequenced through
grove's chartered agent roles, loaded from the grove plugin as `grove:<role>` subagents (all
thirteen — never vendored into `.claude/agents/`). Anything else — conversation, trivial
asks, out-of-scope questions — proceeds normally. This repo's dials live in `.grove/` (see
its README). trellis keeps its own `corpus-reviewer` in `.claude/agents/` — a repo-owned role
that coexists with the plugin's `grove:corpus-reviewer` by namespacing (adr-0026 D5; the
corpus-reviewer lineage originated in trellis, grove adr-0001). Telemetry (`grove-status`) is
not installed — wisp is not vendored here, and telemetry is optional by construction. Version
skew (adr-0026 D4): at role start, if the installed grove plugin's version differs from the
stamp below, disclose the divergence loudly in your report and continue — the stamp is the
in-repo ratified record, never a lock; grove never enforces it.
grove plugin@0.1.0
<!-- grove:end -->
