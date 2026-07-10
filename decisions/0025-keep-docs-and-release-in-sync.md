---
id: decision-0025
type: decision
status: ratified
ratified: 2026-07-05
depends_on: [invariants-v1, decision-0020, decision-0023]
owner: gundi
date: 2026-07-05
---

# 0025 — Keep the landing/docs and the release in sync, automatically

> **Superseded in part by `decision-0043` (2026-07-10, #120; text below preserved as written).**
> Guard 2 (auto-release on merge) retired with the end-user binary channel — with no published
> binaries there is no release to keep in sync; the shipped artifact is the vendored payload at
> HEAD, guarded by regenerate-and-diff. Guard 1 (`docs_consistency_test.go` — no doc claim without
> a shipped feature behind it) stands, unchanged.

## Context

Friction, caught by the maintainer (a B6 self-improvement trigger): the landing page advertised
`trellis remove` / `trellis uninstall` while `releases/latest` was still `v0.1.0`, which lacked them —
so a fresh `curl … | sh` got a binary that did not match the page. The maintainer had to *remember* to
cut a new release.

This breaks two invariants:

- **D1 (transparency — no silent drift).** The landing drifted *ahead* of the shipped binary, and
  nothing surfaced it. "Drift is allowed, but never silent." This was silent.
- **B1 (graph-maintenance).** The landing is a **downstream dependent** of the shipped command surface;
  when that surface changed, the dependent was not re-checked. The dependency existed; nothing
  maintained it.

The fix must be a **checkable guard, not the maintainer's memory** (the iron rule; B6).

## Decision

Two guards keep **landing ↔ code ↔ release** locked:

1. **No doc claim without a shipped feature behind it** — a CI check (`cli/docs_consistency_test.go`)
   that scans the landing, README, and `install.sh` for every `trellis <command>` and `/trellis:<skill>`
   reference and **fails if any names something the CLI or plugin does not actually have**. This is
   `decision-0020`'s rule — *no claim on the page without a rule behind it* — generalized from the
   invariants to the whole product surface. It stops the docs running **ahead** of the code. The check
   runs whenever `cli/`, `plugins/`, `install.sh`, `docs/`, or `README.md` change.

2. **Auto-release on merge to main** — when a change to the **shipped surface** (`cli/**`,
   `install.sh`, `plugins/**`) lands on `main`, `auto-release.yml` bumps the patch version, builds the
   binaries, and publishes a GitHub Release. `releases/latest` therefore always reflects `main`, so the
   landing (which describes `main`) always matches the installer. It stops the release falling
   **behind** the code. Manual `v*` tags still work (`release.yml`) for a curated minor/major; the next
   auto-bump continues from whatever the latest tag is.

Guard 1 blocks *ahead*; guard 2 blocks *behind*.

## Consequences

- **No release is ever hand-remembered.** Every shippable change auto-publishes; doc-only and
  decision-only changes do not (the path filter is the shipped surface).
- The CLI's command set becomes a **single source** (`commands` map in `main.go`) the doc check reads —
  the switch, the usage text, and the docs cannot silently diverge.
- More releases (a patch per shipped change), each a valid state of `main`. Accepted for a tool where
  `main` is always green.

## Open questions

- **Version scheme.** Auto-bump is patch-only; minor/major stay a deliberate manual tag. Revisit if we
  want conventional-commit-driven bumps.
- **Doc-check breadth.** v1 checks command/skill *names*; it does not yet check that described
  *behaviour* matches (e.g. a flag's effect). Extend if a behaviour-claim ever drifts.
