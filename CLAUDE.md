# Bonsai — operating method (seed)

> **What Bonsai is.** A shippable, portable pack that supervises an *agentic software-
> development process* — it fits, teaches, adapts, and guards whatever methodology a
> project uses, while enforcing a small set of invariants. It is **not** a process; it is
> the layer above the steps. See `agentic-dev-meta-layer-brief.md` (in this repo) for the
> full thesis and `invariants/bonsai-invariants-v0.md` for the load-bearing core.

> **We build Bonsai with Bonsai.** This file is a deliberately tiny instance of the seed
> operating method (brief §12). It dogfoods our own invariants from commit one. Friction
> we hit while following it *is product research* — record it, don't route around it.

> **Which layer is this? (under review — decision `0005`).** Bonsai self-hosts, so two
> layers must not be conflated: **Bonsai-core** (the shippable product — invariants, spine,
> ingestion engine, gates) vs **the methodology used to build Bonsai** (this file). The
> proposed resolution: *this file is the build methodology — instance #1, the first
> methodology Bonsai supervises — not Bonsai's product agent-instructions.* Bonsai-core
> lives in its own namespace. Pending ratification of `0005`, treat this file as Layer B.

## The iron rule (most important design constraint)

Bonsai must **always ground out in concrete, project-specific artifacts** — a real
instructions file, real gates, real sub-agents. If it ever just *describes* process
instead of *producing and enforcing* it, it has failed. Same rule applies to this repo:
prefer producing a checkable artifact over writing about one.

## Operating method

- **Artifacts.** Every non-code document opens with frontmatter
  (`id / type / status / depends_on / owner`). Statuses: `draft → ratified → approved`.
  **Downstream consumes only ratified/approved upstream, never drafts.** Every artifact
  carries `## Acceptance criteria` and `## Open questions`.
- **Decisions.** Significant choices get an **append-only** record in `decisions/`. You
  *supersede* (with a forward pointer), never edit, a ratified decision. The four strategic
  forks (brief §9) are records `0001–0004`.
- **Gates.** Human approval at the **intent** layer (vision, decisions, the invariant
  set). **Independent verification** at the **execution** layer (a conformance check
  against the approved upstream before merge — *the builder does not grade itself*).
- **Work.** One logical change per PR; descriptive, linear history; diffs small enough to
  review on a phone.
- **Self-improvement.** Triggers, not vigilance (invariant 8): when friction reveals a
  missing rule, add it *where it fires*, **prefer retiring to adding**, keep it subordinate
  to the work. This file is the first trigger home.
- **Loud failure.** If a required tool, source, or verification is missing, stop *loudly*.
  Never emit plausible-but-unverified output.
- **Epistemic integrity (no sycophancy).** Assessments track the evidence, not the
  maintainer's preferences: push back plainly when analysis warrants, affirm only when the
  data genuinely supports it (and then say so — withholding real positive signal is also a
  distortion), and never perform disagreement to look rigorous. Surface risks, counter-
  arguments, and improvements to proposals by default. The human-facing twin of independent
  verification (invariant B3); it is what makes the intent gate (D2) real. *(Encoded as
  invariant B11; applies to how this project is built.)*

## Naming guardrail (research discipline, applied to ourselves)

If we ever name the invariant set authoritatively, attribute it clearly as **our
synthesis** — never imply pre-existing provenance. For now it is exactly *"Bonsai's
invariants — our synthesis, v0."* Eponymous framing is a deliberate *later* decision, made
only once the set's durability is proven across multiple instances.

## Current state

- **Intent layer (in review):** invariant set v0 + decisions `0001–0005` are `draft`.
  The invariant set is **held for research, not ratified on assertion** (decision `0006`).
- **Active work:** research (decision `0006`) to validate/refine the invariants before
  ratification — chiefly testing real methodologies against the admission gate.
- **Next (deferred until invariants ratified):** the spine — portable artifact contract +
  lifecycle (brief §8.1) — the first machinery; it `depends_on` the ratified invariant set.

## Acceptance criteria

- A newcomer (human or agent) can read this file and the invariant set and know how to
  make a change that will pass the gates.
- Every claim of "done" in this repo traces to a concrete artifact, not a description.

## Open questions

- When does a second project (instance #2) get to test these invariants, given we have
  chosen to validate by dogfooding our own project first (decision `0001`)?
- What is the smallest enforcement that makes "downstream consumes only ratified" real
  here — convention, a check, or a gate sub-agent? (Resolve when the spine is built.)
