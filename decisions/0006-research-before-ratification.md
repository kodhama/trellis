---
id: decision-0006
type: decision
status: draft
depends_on: [invariants-v0, decision-0004]
owner: gundi
date: 2026-06-29
---

# 0006 — Research before ratifying the invariants

**Raised by:** the maintainer — *we are not ready to ratify the invariants directly; we
need more research first.*

## Context

The invariant set v0 is our synthesis lifted from the brief. Ratifying it on assertion
would violate our own research discipline (every load-bearing claim needs a source +
confidence tag; loud failure over plausible-but-unverified output — invariants 3, 7). The
central unproven claim (brief §7) is **generalization beyond N=1** — exactly what assertion
cannot settle. Two contested calls (provenance's class; invariant 1's gate/lifecycle split)
are also explicitly "needs data" per the maintainer.

## Decision

The invariant set stays `draft` and is **validated by research before ratification**, not
ratified by assertion. The research runs in **three steps**:

- **Step 0 — Target identification** *(done → `research/0001-target-landscape.md`)*. Deep
  research to find the AI-agentic-dev methodologies worth testing and rank a shortlist.
  Seeds incl. spec-kit, BMAD, Agent OS, Kiro, OpenSpec, spec-swarm; discover others. Non-AI
  methods (Shape Up, Scrum, RUP) included as **controls only** — first to cut if scope must
  shrink.
- **Step 1 — Invariant validation.** Gate-test each *AI* target against the four
  `methodology` invariants `{1-flow, 2, 4-intent, 5}` (pass/partial/fail + evidence): does
  the gate hold? which breaks first? For *non-AI* controls, instead map **which invariant
  subset they conform to** (that subset map signals which invariants are AI-specific vs
  general). Plus: prior-art/novelty honesty (does a named framework already articulate the
  set? — §4 guardrail), data for the two contested calls, and instance #1 (our own build
  methodology, decision `0005`) against the gate.
- **Step 2 — Invariant refinement/discovery** *(follow-up)*. Use the findings to **tweak
  existing invariants and propose new ones** that better reflect how real methodologies
  behave — bidirectional, not just pass/fail. Likely output: a draft `invariants-v1`.

## Consequences

- The spine (brief §8.1) remains blocked on a ratified invariant set — but **research is
  not blocked**, so this is the active workstream.
- Findings feed back as draft-set revisions (revised in place; this is the consolidated
  current-truth layer), with sources + confidence tags.
- Scope/depth of the research pass is the next thing to agree with the maintainer.

## Supersedes / superseded by

— (none)
