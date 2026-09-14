---
id: decision-0029
type: decision
status: superseded
superseded_by: [decision-0043]
depends_on: [spec-0003, decision-0023, decision-0024]
owner: gundi
ratified: 2026-07-05
---

# 0029 — Setup asks the mode first; detection is per-mode (M1 needs no harness)

> **Superseded by `decision-0043` (2026-07-10, #120; text below preserved as written).** The
> interactive `setup` CLI this decision ordered — ask the install mode first, then detect only
> what that mode needs — is deleted along with the rest of the binary's interactive flow
> (`decision-0043` rule 1: "the interactive `setup` TUI … the harness detection, and the
> binary's M2 path are deleted, not merely undocumented"); the M1/M2 mode split now lives only
> as the plugin skill's own logic (`/trellis:setup`, incl. the M2 morph). This decision's entire
> subject matter — the CLI's mode-first ordering of harness detection — is retired; nothing
> below describes the current setup flow. Retirement of this decision's entire subject matter by
> `decision-0043` confirmed during the 2026-07-10 consistency-sweep review — not originally
> enumerated in `decision-0043`'s own Consequences section, but `decision-0043` rule 1's scope
> covers it entirely.

## Context

`spec-0003 §2b` had setup **detect the harness first** and **exit if the `claude` binary isn't on
PATH** — before the mode was even chosen (`setup.go` called `detectHarness` unconditionally; `harness.go`
is purely `lookPath("claude")`). But only **M2 (morph)** uses the binary: it invokes `claude` to rewrite
the project (`applym2.go`). **M1 (overlay)** is deterministic file editing that never calls it
(`apply.go` — *"plain file editing. No model"*). So the CLI **refused to do an M1 overlay unless you'd
installed the tool M1 never runs** — friction we hit dogfooding, recorded rather than routed around.

## Decision

**Ask the install mode first, then detect only what that mode needs.**

- **M1 (overlay):** no harness binary required — it augments an instruction file deterministically
  (v0: `CLAUDE.md`; detecting/choosing among instruction files — `AGENTS.md`, etc. — is the stacked
  follow-up, with an inline fallback for files that lack `@import`).
- **M2 (morph):** detects and **requires** the harness binary that drives the rewrite; exits loudly if
  none is found (D1).
- Profile follows; a model is asked only for M2.

## Consequences

- `setup.go` reordered; the binary is no longer a global gate. `spec-0003 §2b` updated to match, and
  `main.go`'s help. A new test covers M1 succeeding with no `claude` on PATH.
- Sets up the follow-up (B): M1 detects instruction files and lets you pick the target.
- The morph path is unchanged — the harness *is* the signal there, and it rewrites whatever file it
  loads.

## Open questions

- Whether M2 should offer a harness *choice* once a second harness is supported (today there is one, so
  there is nothing to choose).
