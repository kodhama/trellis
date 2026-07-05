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
> migration is incremental: `invariants/` moved in first; the conformance sub-agent's product
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
  (`id / type / status / depends_on / owner`). Statuses: `draft → ratified → approved`.
  **Downstream consumes only ratified/approved upstream, never drafts.** Required body
  sections are **per-type** (not a blanket rule — a strategic decision has no "acceptance
  criteria"; ratification *is* its acceptance):
  - `decision` → `## Context` / `## Decision` / `## Consequences`
  - `spec` / `invariant-set` → `## Acceptance criteria` / `## Open questions`
  - `research-note` → `## Open questions` (+ sources & confidence tags)
  - `feedback` → exempt (advisory rubric, never a gate)
- **Decisions.** Significant choices get an **append-only** record in `decisions/`. You
  *supersede* (with a forward pointer), never edit, a ratified decision. The four strategic
  forks (brief §9) are records `0001–0004`.
- **Gates.** Human approval at the **intent** layer (vision, decisions, the invariant
  set). **Independent verification** at the **execution** layer (a conformance check
  against the approved upstream before merge — *the builder does not grade itself*).
  **Ratification rides the merge (`decision-0022`):** a PR carries the `draft → ratified`
  flip in its diff, and **merging *is* the ratification** — no draft is left un-ratified on
  `main` past the PR that introduced it (flip it, or keep it clearly WIP).
- **Work.** One logical change per PR; descriptive, linear history; diffs small enough to
  review on a phone.
- **Self-improvement.** Triggers, not vigilance (invariant 8): when friction reveals a
  missing rule, add it *where it fires*, **prefer retiring to adding**, keep it subordinate
  to the work. This file is the first trigger home.
- **Derived resources stay in sync (`decision-0028`).** When you change a *source* — the
  catalog, a spec, the CLI's command set — update everything that **derives** from it in the
  same change. A source names its derivatives (so you see them at the edit), and a check
  guards each pair. If you edit something and can't name what derives from it, that's the
  question to ask. (This is B1 made salient: the graph, pointed forward.)
- **Loud failure.** If a required tool, source, or verification is missing, stop *loudly*.
  Never emit plausible-but-unverified output.
- **Epistemic integrity (no sycophancy).** Assessments track the evidence, not the
  maintainer's preferences: push back plainly when analysis warrants, affirm only when the
  data genuinely supports it (and then say so — withholding real positive signal is also a
  distortion), and never perform disagreement to look rigorous. Surface risks, counter-
  arguments, and improvements to proposals by default. This is the *intent face* of
  independent judgment (invariant B3); it is what makes the intent gate (D2) real.

## Naming guardrail (research discipline, applied to ourselves)

If we ever name the invariant set authoritatively, attribute it clearly as **our
synthesis** — never imply pre-existing provenance. For now it is exactly *"Trellis's
invariants — our synthesis, v1."* Eponymous framing is a deliberate *later* decision, made
only once the set's durability is proven across multiple instances.

## Current state

- **Intent layer: ratified.** `invariants-v1` is the ratified current-truth set (A1–A4 gate
  · B1–B9 operating · C dials · D floors); decisions `0001–0008` are ratified; v0 superseded.
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
