---
id: decision-0005
type: decision
status: ratified
ratified: 2026-06-29
depends_on: [decision-0001, decision-0003, invariants-v0]
owner: gundi
date: 2026-06-29
---

# 0005 — Trellis self-hosts: separate Trellis-core from the build methodology

**Raised by:** the maintainer's question — *is `CLAUDE.md` Trellis's product agent-
instructions, or the instructions to build the Trellis pack?*

## Context

We dogfood: we build Trellis with Trellis. That creates two layers `CLAUDE.md` currently
conflates:

- **Layer A — Trellis-core (the product):** the shippable pack — invariants, the spine, the
  methodology-ingestion engine, gates, the conformance sub-agent. Ships into *other*
  projects. Contains *"how to supervise any methodology."*
- **Layer B — the build methodology:** our own stages, gates, decisions, operating method —
  *a specific methodology*.

The relationship is **stratified, not circular** (the compiler-compiling-itself shape):
Trellis-core is the supervisor; "how we build Trellis" is just the *first methodology that
Trellis-core supervises*. Core never contains "how to build Trellis" — it contains "how to
supervise any methodology," and our build methodology is *data fed to it* (instance #1).

## Decision

- **Trellis-core (Layer A)** and **the Trellis-build methodology (Layer B)** are distinct and
  must not leak into each other. The product must never ship our internal build cruft.
- **`CLAUDE.md` is Layer B** — the build methodology, **not** Trellis's product agent-
  instructions. It is reframed as **instance #1**: the first methodology Trellis supervises,
  to be expressed in the standard instruction-file format (decision `0003`) once it exists.
- **Trellis-core lives in its own namespace** (proposed `core/`; `invariants/` is its first
  content). Exact layout TBD.

## Consequences

- **Free dogfood:** our own build process becomes the first empirical test of the admission
  gate — does Layer B satisfy `{1-flow, 2, 4-intent, 5}`? (Feeds decision `0006`.)
- **First real consumer** for the standard instruction-file format (`0003`): our own
  methodology.
- **Requires a physical repo reorg** (move product content into `core/`, keep build
  methodology at root). **Deferred until this decision is ratified.**

## Open questions

- Exact namespace names (`core/` vs `pack/` vs `trellis/`); when to physically reorg (now vs
  after the spine).
- Does instance #1 (our own process) actually pass the admission gate? If not, that is
  either a bug in our process or a signal about the gate — both are findings.

## Supersedes / superseded by

— (none)
