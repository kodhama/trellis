---
id: decision-0024
type: decision
status: ratified
ratified: 2026-07-05
depends_on: [invariants-v1, spec-0002, spec-0003]
owner: gundi
date: 2026-07-05
---

# 0024 — Gatekeeping is detect-and-respect; Trellis surfaces silent human-gate bypasses

## Context

Scoping how the setup CLI handles the gatekeeper dial (C2), a real gap surfaced:

- **Real projects have mixed gates.** Some handovers are human-approved, some agent-approved
  (spec→plan human, plan→build agent, …). A single per-invariant or per-project C2 posture is too
  coarse, and recording "mixed" carries no information.
- **Trellis cannot know a project's gates in advance** — they belong to the methodology it lays on top
  of. SpecSwarm hands documents agent→agent through automatic steps (agent-gated); Spec Kit's commands
  are human-approved (human-gated).
- An earlier proposal — a Trellis-owned **per-gate map** the CLI enumerates and configures — is too
  cumbersome for the CLI to build for every underlying gate.

## Decision

1. **Detect-and-respect, not configure.** The host project owns its gate declarations (which handovers
   are human- vs agent-gated). Trellis **reads and respects** them; it does not build, choose, or impose
   a per-gate map. This is the Trellis thesis applied — fit the methodology, don't dictate it.
2. **Trellis's job is surfacing.** It installs the behavior of invariant **B2** — *"the gate is real at
   the strictness of dial C1, and any skip is surfaced"* — scoped by **C2** (which gates are human) and
   **D2** (intent gates are human by floor), made loud by **D1**. Not a new invariant; the operational
   composite of ones we already have.
3. **One-directional.** ONLY a human-gated handover executed **without its required human approval** is
   surfaced. Agent-gated handovers proceed silently. Trellis does **not** detect or flag human approval
   on agent gates — no value, and humans lack the bandwidth to touch agent gates.
4. **Recommend only for gaps.** Where a project declares no gate, Trellis may suggest the posture's
   default (coarse, configurable) — never gate-by-gate.
5. **Per-gate configuration by Trellis → v2.** Enumerating and choosing per gate is too cumbersome for
   v0.

## Consequences

- The profile's `C2` is a **detected, respected readout** (Assess populates it from the project), not
  an imposed map. **No schema change; no per-gate structure in v0.**
- The installed rules (M1 overlay / M2 morph) carry the **B2 surfacing behavior**, focused on silent
  human-gate bypasses.
- **Testable target behavior:** a project declares "spec→plan is human-gated"; an agent runs spec→plan
  with no recorded human approval → Trellis surfaces it. A clean e2e assertion (and the one worth a
  model-judge, since the phrasing is non-deterministic).
- Simplifies the CLI: no per-gate step; gatekeeping falls out of Assess (detect) + the installed rule
  (surface).

## Open questions

- **Reading gate declarations generically** — how Assess infers which handovers a project treats as
  human- vs agent-gated, across methodologies (the catalog `signature` tells seed it).
- **What counts as "human approval"** the agent checks for — a PR approval, a commit trailer, an
  explicit acknowledgement? How the agent knows approval did or did not happen.
- **Per-gate configuration** — deferred to v2.
