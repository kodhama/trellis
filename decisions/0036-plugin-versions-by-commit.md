---
id: decision-0036
type: decision
status: ratified
depends_on: [decision-0035]
owner: gundi
ratified: 2026-07-06
---

# 0036 — The plugin versions by commit, not a frozen number

## Context

`decision-0035` flagged that the Claude Code plugin didn't auto-pull latest: `plugin.json` was frozen at
`0.1.0` while the CLI moved to `0.2.16`, so even `/plugin marketplace update` saw an unchanged version and
pulled nothing. It named two fixes — bump `plugin.json` on each release, or omit the version so each
commit counts as a new version.

## Decision

**Omit the `version` field from `plugin.json`.** Claude Code's version resolution falls back to the git
commit SHA when no version is set (verified), so **every merge to `main` is automatically a new plugin
version** — `/plugin marketplace update` always resolves to the latest.

Chosen over an auto-release bump because it is **zero-maintenance** (no CI step that commits back to
`main`, so no trigger-loop or branch-protection snag) and **always-fresh** — which fits an early,
fast-moving project where plugin users should be on the current invariants, not a pinned snapshot.

## Consequences

- The plugin version is the commit SHA, **not** semver-aligned with the CLI. Users can't pin to a semver
  plugin version. If stable, pinnable plugin releases are wanted later, reintroduce a `version` + an
  auto-bump (with a loop guard).
- The CLI path keeps its own drift surface — the `.trellis/version` stamp + `trellis status`
  (`decision-0035`). *(Superseded in part by `decision-0043` (#120): `trellis status` retired with
  the binary; `.trellis/version` now carries the payload's render stamp, compared file-to-file by
  the plugin's SessionStart hook.)*

## Open questions

- Auto-update stays **off by default** for third-party marketplaces (Claude Code default) — users still
  run `/plugin marketplace update` (or opt into auto-update). Nothing we control from our side; worth a
  line in the install docs.
