# Bonsai

A shippable, portable pack that **supervises an agentic software-development process** — it
fits, teaches, adapts, and guards whatever methodology a project uses, while enforcing a
small set of invariants. It is **not** a process; it is the layer *above* the steps.

> Keep a process minimal, deliberately pruned, and shaped to fit, yet alive and adapting —
> the art of bonsai.

## Status

The **intent layer is ratified**: `invariants-v1` (the load-bearing invariant set) and
decisions `0001–0008` are ratified, after a three-step research pass (`research/`) that
gate-tested real frameworks (Spec Kit, Kiro, BMAD, OpenSpec, SpecSwarm). **Next:** the
*spine* — the portable artifact contract + lifecycle — the first machinery. The N=1
generalization risk (one project so far) remains the central open question.

## Where things are

- [`agentic-dev-meta-layer-brief.md`](agentic-dev-meta-layer-brief.md) — the full thesis
  (read §10 verdict, §11 start-here, §12 operating method first).
- [`core/invariants/bonsai-invariants-v1.md`](core/invariants/bonsai-invariants-v1.md) — the
  load-bearing core (ratified): *Bonsai's invariants — our synthesis, v1* (structural gate ·
  operating layer · dials · floors). v0 kept as a superseded diff.
- [`decisions/`](decisions/) — append-only decision records (`0001–0008`, ratified).
- [`research/`](research/) — the framework gate-test research notes.
- [`CLAUDE.md`](CLAUDE.md) — the methodology we use to build Bonsai (Layer B / instance #1).

## How we work

We build Bonsai with Bonsai — dogfooding our own invariants from commit one. See
[`CLAUDE.md`](CLAUDE.md) for the operating method (artifacts with frontmatter, append-only
decisions, intent-gate vs conformance-gate, minimal-first, loud failure).
