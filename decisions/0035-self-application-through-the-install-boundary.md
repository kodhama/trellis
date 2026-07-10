---
id: decision-0035
type: decision
status: ratified
depends_on: [decision-0005, decision-0010, decision-0028]
owner: gundi
ratified: 2026-07-06
---

# 0035 — Trellis self-applies through its own install boundary; overlay↔product drift is never silent

> **Superseded in part by `decision-0043` (2026-07-10, #120; text below preserved as written).**
> Rule 3's user-facing surface (`trellis status`, "update the tool, then re-run `trellis setup`")
> retired with the binary channel; the surface is now the plugin's bundled SessionStart hook — a
> file-to-file compare of `.trellis/version` against the installed plugin's payload stamp. The
> floor itself ("drift is made visible, not silent") and the repo-facing CI sync-guard stand
> unchanged; the install-boundary/self-application model below is untouched.

## Context

Two things were blurred, and they are the same question from two sides — *is the overlay current with
the product?*

1. The repo **hand-authors its invariants** into `CLAUDE.md` — a parallel copy of the product's own
   output that can drift (self-*reference*).
2. The install writes a **snapshot** of the invariants into `.trellis/`, with **no version stamp and no
   update signal**, so a user's overlay goes **silently stale** as the product evolves — which violates
   our own transparency floor (**D1**).

## Decision

**1. The install boundary separates *producing* Trellis from *consuming* it.** The product is authored in
`core/`; it is consumed only via the official install path (`trellis setup` → `.trellis/`). That boundary
makes the repo's use of Trellis **self-application, not self-reference** — a compiler built, then run on
itself: bootstrapped, not circular. Three roles, kept distinct:

- **Build** — invariants / catalog / CLI / plugin, in `core/` (the output).
- **Govern** — the invariant directives, which **land only via the install** (`.trellis/`), never
  hand-restated. Single-sourced from Build, so they can't diverge.
- **Method** — the *how* (the iron rule, the artifact contract, the decision/gate mechanics),
  hand-authored in `CLAUDE.md` (Layer B, `decision-0005`). A generic overlay can't carry a specific
  project's method.

**2. The repo eats its own dogfood through the front door.** `CLAUDE.md` is stripped of invariant
*echoes* (the bullets that just restate D1/B3/B6…) — those land from the install now. The repo runs
`trellis setup` (conductor, M1, `CLAUDE.md`), commits `.trellis/` + the block, and a **CI sync-guard**
regenerates it and diffs — so the repo's overlay is always the current product: drift is *impossible*,
not merely visible.

**3. For users, drift is made visible, not silent (D1).** No package manager — the overlay stays
self-contained (`decision-0010`); the update path is *update the tool, then re-run `trellis setup`*. To
keep that honest, the overlay is **stamped with the version that generated it**, and `trellis status`
surfaces when the overlay is behind the installed binary. Snapshot-by-design, drift-by-surface.

**4. The dual role is stated out loud** (README + CLAUDE.md): *"This repo produces Trellis (`core/`, the
plugin, the CLI); it is also a Trellis-governed project — it installs Trellis via the official path to
govern its own work."*

## Consequences

- The overlay stamps a version; `trellis status` reports staleness (the user-facing D1 surface). The CI
  sync-guard (the repo-facing enforcement) normalizes the version line so a dev build doesn't spuriously
  fail the diff.
- `CLAUDE.md` shrinks to the project method + thesis; the invariants come from `.trellis/`.
- `decision-0028`'s derived-resource rule now covers the overlay itself, from both the repo and user
  sides.

## Open questions

- **The plugin path doesn't auto-pull latest either** (verified): third-party marketplaces have
  auto-update **off** by default, and our `plugin.json` version isn't bumped on release — so plugin users
  can be pinned to an old version. Fix in a follow-up: bump `plugin.json` in `auto-release`, or omit its
  `version` so each commit counts as a new version. Same "surface the staleness" principle.
- Whether `trellis status` should also compare against the *published* latest release (a network check)
  or only the installed binary — the latter keeps the no-runtime-dep promise; the former needs a
  deliberate opt-in.
- The exact `CLAUDE.md` cut line — which method bullets are genuine Layer-B *how* vs invariant *echo* —
  is worked out in the self-application slice, not here.
