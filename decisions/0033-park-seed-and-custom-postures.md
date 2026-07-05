---
id: decision-0033
type: decision
status: ratified
depends_on: [spec-0003, invariants-v1]
owner: gundi
ratified: 2026-07-05
---

# 0033 — Offer two postures (conductor / author-adapt); park seed and custom

## Context

Setup offered four profile presets — **A conductor · B author-adapt · seed · Custom**. In practice:

- **`custom` was a literal duplicate of `seed`** (same `C1Lean`, same `Active` set) — its intended
  "choose every dial/knob" flow was never built, so it did nothing distinct.
- **`seed`'s value as a picker option was unclear** to the maintainer — it silently activated a 5-slug
  subset, with no visible effect a user could reason about.
- **A and B both activate all 14 invariants**, so they differ *only* in the `C1` enforcement lean — and
  in advisor mode that lean is **descriptive, not enforced** (nothing acts on it yet). The one concrete
  lever a posture has today is *which invariants are active*, and only `seed`/`custom` used it.

## Decision

1. **Offer exactly two postures — `conductor` (A) and `author-adapt` (B)** — with honest copy that names
   the difference as a **stance** ("hold firmly, by-the-book" vs "same invariants, adapt as you go"),
   not a mechanical one. Default stays **B**.
2. **Park `seed` and `custom`.** `custom` returns only with a real **per-dial dialog** (choosing each C1
   strength / knob), not a stub. `seed` (minimal-start) returns when the **active-subset lever earns its
   keep** — e.g. once `C1` is enforced (supervisor mode) or a "start minimal, ratchet up" onboarding is
   built.
3. **Keep the `Active`-subset mechanism in code** (`Profile.Active` + `renderProfile`), inert, so `seed`
   can come back as data without re-plumbing.

## Consequences

- Fewer, clearer choices; the honest limit is stated: **A vs B is a stated stance until `C1` is actually
  enforced.** When supervisor mode enforces it, A and B become mechanically different — revisit the copy
  then.
- The `--profile` flag accepts `a|b`; tests updated off the parked keys.

## Open questions

- The per-dial `custom` dialog — worth building once the dials are individually meaningful (post
  supervisor mode).
- Whether `author-adapt` should *also* differ in active set (not just lean), to give B a concrete effect
  before enforcement lands.
