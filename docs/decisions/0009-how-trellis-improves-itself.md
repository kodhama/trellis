---
id: decision-0009
type: decision
status: ratified
ratified: 2026-06-30
depends_on: [decision-0001, decision-0002, invariants-v1]
owner: gundi
date: 2026-06-29
---

# 0009 — How Trellis improves itself (validation scenarios + feedback loop)

**Raised by:** the maintainer — *Trellis's self-improvement loop improves the project or the
methodology, not Trellis. So what improves Trellis, beyond our design sessions?*

## Context

Trellis's self-improvement loop (B6) lands on the project (adopt) or the methodology (adapt)
per the adherence axis (B8) — never on Trellis-core. Trellis is an improvement engine with no
input channel for its own improvement. Trellis-core can only improve from **cross-instance
signal**, which is structurally unavailable at N=1: a claim about an *invariant* is a
meta-claim, and the only non-circular evidence is **recurrence across independent instances**.
So "what improves Trellis" has two honest answers — near-term validation, and an ongoing loop.

## Decision

### A. Near-term validation scenarios (run after spine + delivery exist)

Three scenarios spanning the adherence axis (B8):

1. **Empty project, no references** — author-from-nothing. Tests whether a **built-in seed**
   methodology is needed.
2. **math-quest** — adapt onto an existing, *evolving* project. Tests non-redundant,
   non-contradictory **retrofit**.
3. **Project with spec-kit installed** — strict-adopt / conductor. Tests Trellis **enforcing
   an existing methodology** more strictly.

**Honest limit (load-bearing):** all three are run by us on projects we know, and **math-quest
is the *source* project — genealogically N=1** (same maintainer/assumptions, the project the
invariants were extracted from). They test **mechanism** (does author/adapt/conduct work; does
retrofit avoid contradiction), **not generalization** of the invariants. A
genealogically-independent instance (different domain, ideally different person) is still owed
and is the only true N≥2 signal. *(Positive-control discipline: a project we shaped cannot
disconfirm our own invariants — same failure mode as the CI-reviewer verification.)*

### B. The ongoing feedback loop

- **Channel:** a special GitHub issue type, raised **only with the user's explicit permission**
  to record the feedback. Consent is also an **enterprise-trust feature** — never exfiltrate a
  project's process friction silently (D1 surfacing, pointed at the vendor relationship).
- **Signal source:** D1-surfaced friction — skips, overrides, gate-failures, divergences. The
  **gold signal is the human overriding Trellis and being right** (a labeled correction).
- **Triage:** an agent (possibly an *independent* reviewer agent rather than Trellis's own
  voice — see the B3 positive-control note) watches the issues and, when **critical mass** is
  reached on a topic — **weighted by instance diversity, not raw volume** (N reports from one
  project ≠ signal; a few from different domains = signal) — proposes a change and **opens a PR
  for the maintainer to review**.
- **Human gate (B3 + D2):** Trellis **suggests, never merges**. Trellis-deciding-to-change-Trellis
  is the builder grading itself; the maintainer's review at the intent gate is the independent
  check. Core changes are never auto-applied.

## Consequences

- **Spine requirement:** the spine must capture D1-surfaced friction as **structured,
  exportable records**, so signal isn't lost before the loop's collection exists. Cheap
  insurance — build it into the spine's surfacing.
- **Delivery dependency (`decision-0001`):** the consent-based GitHub-issue channel fits the
  self-hosted / open-source-shaped delivery; a hosted tier could later add telemetry.
- **Adherence interaction (`decision-0002`/B8):** adds a **third routing tier** — improvements
  land on the *project* (adopt), the *methodology* (adapt), or **Trellis-core** (when the
  improvement is about *how processes are supervised* — an invariant/gate/dial/catalog — not a
  specific process). The Trellis-core tier has the highest bar (cross-instance recurrence) or
  B7/minimal-first dies.

## Open questions

- The **critical-mass threshold**, and how to measure **instance diversity** honestly.
- Is the triage agent being **Trellis itself** too circular even with the human gate? (Ties to
  the B3 positive-control note — may need an independent reviewer agent.)
- **Consent/privacy model** for enterprise: what exactly is recorded, redaction, retention.
- Run **math-quest** despite genealogical weakness? *Yes for retrofit mechanics — flag the
  limit; it is not instance #2.*

## Supersedes / superseded by

— (none)
