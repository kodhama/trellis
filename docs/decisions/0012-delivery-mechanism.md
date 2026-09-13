---
id: decision-0012
type: decision
status: ratified
depends_on: [decision-0001, decision-0003, decision-0010, invariants-v1]
owner: gundi
date: 2026-06-29
ratified: 2026-06-29
---

# 0012 — Delivery: own Claude-plugin marketplace (v0) → support CLI (v1) → git-copy always

**Raised by:** the maintainer — given Trellis ships as runtime-free instructions (`0010`),
*how* is it delivered, and how is it **activated** (hooked into host behavior, not merely
present)?

## Context

Verified from the Claude Code docs: a plugin **marketplace** is just a git repo with a
`marketplace.json` — **anyone can self-host, no Anthropic approval.** Enabling a plugin
**auto-activates** its hooks (fire on events), skills (model-invoked), and — via
`settings.json`'s `agent` key — can set the **default agent/system prompt**. So a plugin is
*activation*, not mere availability; git copy-in is availability only.

## Decision

- **v0 — host our own Claude-Code plugin marketplace** (`marketplace.json` in the Trellis
  repo). Bundles Trellis's resources (skills, sub-agents, hooks, optional default-agent). No
  gatekeeping; fastest to dogfooding; **the plugin repo = where `0009`'s feedback PRs land =
  the update channel** (push → users `/plugin marketplace update`).
- **v1 — a support CLI** (`0010`) that adapts/wires the pack into non-plugin surfaces
  (inspect-or-be-told, `0003`) and carries ops (CI conformance, feedback export). Support,
  never a runtime dependency.
- **Always — git copy-in** works, because resources are instructions.
- **Activation level *is* the C1 enforcement dial, surfaced** (`0008`): install offers
  *available + referenced* → *hooks fire* → *default agent*, as the user's **surfaced
  choice** — because mere availability ≠ used (expressed-vs-enforced at the delivery level).
- **`claude-community` submission** (review pipeline + safety screening) is a later growth
  move for discoverability, **not v0**.

## Consequences

- **Delivery is parallel to the spine, not a prerequisite** — self-hosting (building Trellis
  in a Claude Code repo) needs no install.
- **The spine spec must specify the activation/wiring contract** (which hooks/skills/default-
  agent per dial level; how it composes with the host `CLAUDE.md` without clobbering).

## Open questions

- Exact hooks / skills / default-agent at each activation-dial level.
- Composition with an existing host `CLAUDE.md` (coexist, not overwrite).
- When to submit to `claude-community`.

## Supersedes / superseded by

— (none)
