# Trellis (Claude Code plugin)

A governance layer for agentic software development — the native Claude Code way to install it, no
binary required.

## Install

```
/plugin marketplace add kodhama/trellis
/plugin install trellis@trellis
```

Then run the setup skill in any project:

```
/trellis:setup
```

It asks for a **posture** (conductor / author-adapt — or reads it from `.trellis/expression.md` if
the project already declares one, asking nothing) and copies Trellis onto your project as an
**overlay** — a `.trellis/` bundle (your profile + the full invariant reference) plus a one-line
`@import` in your `CLAUDE.md`. All content is pre-rendered at release and verified against a shipped
checksum manifest (`kodhama-0007`: the skill copies, it never composes). Augment-never-clobber;
nothing else is touched, and it's idempotent.

## What it bundles

- **`skills/setup`** — `/trellis:setup`: install the overlay (done natively, no binary).
- **`skills/remove`** — `/trellis:remove`: cleanly reverse it (delete `.trellis/`, strip the `CLAUDE.md`
  block, touch nothing else).
- **`reference/`** — the pre-rendered payload (`kodhama-0007`): `invariants.md` (the full signature
  catalog: every invariant with its *why* and a with/without example), every posture variant of the
  overlay files and managed blocks, and the checksum manifest the setup skill verifies against.
- **`hooks/`** — a `SessionStart` hook that stays quiet until the plugin updates past the overlay it
  wrote (`decision-0039`), then nudges you once: *"the overlay may be stale — run `/trellis:setup`."*
  Network-free (it compares the overlay's stamped `plugin@<sha>` to the installed plugin's HEAD), so it
  can tell you the overlay is *behind the installed plugin*, not how far behind the marketplace.

## Removing it

Run `/trellis:remove` — it deletes `.trellis/` and strips the managed `CLAUDE.md` block, leaving your
own content intact. (The [Trellis CLI](https://github.com/kodhama/trellis)'s `trellis remove` does
the same, and additionally handles the git rollback for an M2 morph.)

## Plugin vs CLI

This plugin covers the **M1 overlay** natively inside Claude Code. The [CLI](https://github.com/kodhama/trellis)
additionally offers **M2 (a model-driven morph on a git branch)** and works with harnesses beyond
Claude Code. Same invariants, two delivery routes.
