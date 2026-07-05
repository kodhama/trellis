# Trellis (Claude Code plugin)

A governance layer for agentic software development — the native Claude Code way to install it, no
binary required.

## Install

```
/plugin marketplace add gundisalwa/trellis
/plugin install trellis@trellis
```

Then run the setup skill in any project:

```
/trellis:setup
```

It asks for a **posture** (conductor / author-adapt / seed) and composes Trellis onto your project as
an **overlay** — a `.trellis/` bundle (your profile + the full invariant reference) plus a one-line
`@import` in your `CLAUDE.md`. Augment-never-clobber; nothing else is touched, and it's idempotent.

## What it bundles

- **`skills/setup`** — `/trellis:setup`: install the overlay (done natively, no binary).
- **`skills/remove`** — `/trellis:remove`: cleanly reverse it (delete `.trellis/`, strip the `CLAUDE.md`
  block, touch nothing else).
- **`reference/invariants.md`** — the full signature catalog: every invariant with its *why* and a
  with/without example.

## Removing it

Run `/trellis:remove` — it deletes `.trellis/` and strips the managed `CLAUDE.md` block, leaving your
own content intact. (The [Trellis CLI](https://github.com/gundisalwa/trellis)'s `trellis remove` does
the same, and additionally handles the git rollback for an M2 morph.)

## Plugin vs CLI

This plugin covers the **M1 overlay** natively inside Claude Code. The [CLI](https://github.com/gundisalwa/trellis)
additionally offers **M2 (a model-driven morph on a git branch)** and works with harnesses beyond
Claude Code. Same invariants, two delivery routes.
