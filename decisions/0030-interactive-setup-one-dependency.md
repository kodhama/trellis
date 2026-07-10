---
id: decision-0030
type: decision
status: ratified
depends_on: [decision-0023, invariants-v1]
owner: gundi
ratified: 2026-07-05
---

# 0030 — The interactive setup takes one dependency (golang.org/x/term), not a TUI framework

> **Mooted by `decision-0043` (2026-07-10, #120; text below preserved as written).** The
> interactive setup TUI retired with the end-user binary channel, and the `x/term` dependency was
> removed with it — the generator-only CLI is dependency-free. The "smallest dependency that
> does the job, or none" reasoning below still governs any future dependency question.

## Context

The setup CLI is the product's front door (`spec-0003 §2b`) and it was bland — flat text, type-a-key
selection. Colour is free (ANSI), but **arrow-key navigation needs the terminal in raw mode**, which the
stdlib can't do cleanly. So it forces the CLI's **first dependency** — a `minimal-first` (B7) call. The
tool that preaches minimal-first should practise it: a Trellis CLI that pulls a whole TUI framework
(bubbletea/huh) to ask four questions would undercut its own thesis.

## Decision

- **One dependency: `golang.org/x/term`** — a thin, quasi-official package (only `x/sys` behind it). We
  **hand-roll a small single-select component** (↑/↓/enter, j/k, Ctrl-C) rather than adopt a TUI
  framework. Pinned to `v0.17.0` so the `go 1.22` toolchain stays put (no toolchain bump).
- **Colour reuses the landing's accent green** (`#1f9d68`), applied to **accents only** so it reads on
  light and dark terminals with no theme detection; respects `NO_COLOR`.
- **TTY-only.** The selector engages only when stdin+stdout are both terminals. Pipes, CI, flags, and
  tests keep the **deterministic line-based flow, byte-identical** — the whole suite is unchanged.
- **Single static binary preserved** (`decision-0023`): Go links the dep in; `curl | sh` is unaffected.

## Consequences

- `go.mod` gains its first dependency; CI module caching enabled.
- A `tui.go` selector + palette; `ask()` routes to it on a real terminal, else falls back.
- Rejected: a TUI framework (huh/bubbletea) — more polish, less code, but the first *framework* + a
  transitive tree, against B7 and posture-consistency; and colour-only — no dep, but no arrow-nav.

## Open questions

- Whether the y/N confirm and the Custom-profile dials later warrant richer widgets — revisit only if
  the hand-rolled surface grows beyond a single-select (minimal-first).
