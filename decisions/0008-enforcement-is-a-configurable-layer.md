---
id: decision-0008
type: decision
status: draft
depends_on: [decision-0002, decision-0004, invariants-v0, research-0002]
owner: gundi
date: 2026-06-29
---

# 0008 — Enforcement is a configurable layer; Bonsai surfaces, the org configures

**Raised by:** the maintainer, reflecting on Step 1's "frameworks express but don't enforce"
finding — flagging that **strict enforcement was a hypothesis, not a locked design choice.**

## Context

Step 1 (`research-0002`) framed Bonsai's wedge as "turn expressed → enforced." Taken as a
mandate, that over-indexes on strictness and would **alienate users who don't want it**
(speed-first startups), narrowing the market the project deliberately kept open (decision
`0004`, buyer-neutral). The frameworks themselves already show the better pattern (SpecSwarm:
a mandatory-by-default gate *with* surfaced `--quick` skip).

## Decision (direction — draft, for future reasoning)

- **Enforcement strength is a dial, not Bonsai's fixed stance.** Extends decision `0002`
  (adaptation as a dial) to the *enforcement* axis: `expressed` → `default-on-but-skippable`
  → `enforced`. Strictness/assurance is an **opt-in layer** aimed at B2B/enterprise; it must
  stay optional so speed-first users aren't lost.
- **The non-negotiable is *surfacing*, not enforcing.** Bonsai may allow skipping a gate,
  but the skip must be **surfaced** — a conscious, visible choice, never silent. This is the
  real floor (extends invariant 7, loud-failure / no silent degradation). The hard invariant
  is *transparency*; enforcement rides configurably on top.
- **Skipping leverages the underlying framework's own skip machinery** where it exists (Kiro
  Quick Plan, Spec Kit lean path, SpecSwarm `--quick`). Skipping may be a feature *of the
  framework* (surfaced by Bonsai) or offered *by Bonsai* even where the framework lacks it —
  but always consciously.
- **Target UX is conversational.** The user asks "what's the next step?"; Bonsai answers
  "it's X — and you may skip it," and choosing skip invokes the underlying skip. Bonsai is a
  conversational guide/tutor over the framework (Pillar III, onboarding, made concrete).
- **The gatekeeper is configurable.** For each gate, *who enforces it* is a setting:
  **independent agent | human | none**. Bonsai's *opinion* may be a recommendation on which
  gatekeeper fits the context (enterprise intent gate → human; conformance → independent
  agent; speed → none) — but the choice is the org's.

## Consequences

- Reframes the Step 1 wedge: Bonsai's value is **making the gate/skip choice explicit and
  configurable + surfacing it**, not forcing strictness. ("Surface the choice" > "enforce.")
- Invariant 5 gains a config dimension: gatekeeper ∈ {agent, human, none}; **"none" is a
  door we optionally open**, permitted only when the skip is surfaced.
- Stays buyer-neutral (`0004`): one invariant structure, different enforcement config per
  buyer — strictness for enterprise, lightness for startups, same machinery.

## Open questions

- **Is "surfacing" already invariant 7 (loud failure), or its own invariant** ("informed/
  conscious skip")? Step 2 should decide whether transparency is elevated to the hard floor.
- **If a gatekeeper can be "none," does the gate still exist (invariant 2)?** Reframes the
  skippable-gate question: a *skipped-but-surfaced* gate may be acceptable where a
  *silently-absent* one is not.
- **Where is the floor that is never configurable to "off"?** Candidate: the **intent gate**
  — invariant 4 already says it "never fully opens," so "none" may be disallowed *there*
  specifically, even as everything downstream is configurable. Surfacing/provenance is the
  other candidate floor.

## Supersedes / superseded by

— (none; refines the enforcement framing in `research-0002` and `0002`)
