---
id: decision-0010
type: decision
status: ratified
depends_on: [decision-0001, invariants-v1]
owner: gundi
date: 2026-06-29
ratified: 2026-06-29
---

# 0010 — Trellis imposes no runtime; it ships as agent instructions (CLI is optional support)

**Raised by:** the maintainer, catching a category error while scoping the spine — I had
framed Trellis's conformance check as a Python/Node *script*, treating Trellis as a code tool.
It isn't.

## Context

While scoping the spine, the conformance check was being designed as a runnable *script* —
which presumes Trellis is a program with a runtime. Trellis is a supervisor **pack**
(guardrails, rubrics, sub-agents) interpreted by whatever agentic surface a project uses, so a
runtime requirement would contradict both its nature and its portability goal (`decision-0001`).

*(`## Context` added as a mechanical conformance fix — the conformance check flagged its
absence. Substance unchanged; permitted under the append-only rule, which protects a decision's
*substance*, not its section scaffolding.)*

## Decision

- Trellis's **resources** — rules, sub-agents, skills, rubrics, conventions (including the
  artifact contract and its conformance check) — are **agent instructions that require no
  runtime**. They are interpreted by whatever agentic surface the project already uses (Claude
  Code, Cursor, …) and **composed into that surface** (`CLAUDE.md` / `.claude/` / `AGENTS.md`),
  coexisting with and tweaking what's there.
- The artifact-contract "validator" is a **conformance sub-agent applying a rubric**, failing
  loudly (B3 / D1) — **not** a program.
- A **support CLI / installer is permitted** for install, scaffolding, and ops — e.g. adapt
  the pack into the project's surface, run conformance in CI, help file the consent-based
  feedback issue (`decision-0009`). It is **support only, never a runtime dependency**: the
  methodology runs without it.
- Any deterministic helper a project wants for hard CI gating is written in **the target
  project's own stack** — never a runtime Trellis imposes.

## Consequences

- Resolves portability (`decision-0001`): the pack is portable *because* it is instructions,
  not code. The "Python vs Node" stack question was malformed and is closed.
- Surface-agnosticism (brief §1) is preserved — an installer's job is to *adapt to* the
  surface, not impose one.
- This had been implicit; making it explicit prevents exactly the error made here.

## Open questions

- Which agentic surfaces to support first (→ the delivery-mechanism decision, still open).
- How the conformance sub-agent is invoked in CI without a runtime (a thin language-native
  shim, or the agent itself inside an action, as with the existing review workflow).

## Supersedes / superseded by

— (none)
