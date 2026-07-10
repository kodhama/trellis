# Trellis (Claude Code plugin)

A governance layer for agentic software development — the native Claude Code way to install it, no
binary required. This is the **primary install path** (`kodhama-0002`; the Homebrew/curl binary
channel retired per `kodhama-0007` rule 5).

## Install

From the kodhama family marketplace (the in-repo marketplace `kodhama/trellis` → `trellis@trellis`
still works as an alias):

```
/plugin marketplace add kodhama/kodhama
/plugin install trellis@kodhama
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

`.trellis/expression.md` is the bundle's one **hand-owned** file (`kodhama-0007` rule 4): its
frontmatter declares the posture (the only machine-read line), and its body is where the project
records how it expresses the invariants — dials, mappings, gate tables. Setup seeds it once and
never rewrites it; the overlay header imports it, so what you write there stays always-loaded.

## Migrating an older install (hand-authored content in `profile.md`)

Overlays installed before `expression.md` existed sometimes carry the project's own expression
appended to the generated `.trellis/profile.md` — the clobber target of
[#112](https://github.com/kodhama/trellis/issues/112): a refresh rewrites that file whole. To
migrate, once:

1. Open `.trellis/profile.md`. Everything below its closing "(Generated from your profile …)"
   line is yours — cut it.
2. Create `.trellis/expression.md`: YAML frontmatter declaring your posture (`profile: a` or
   `profile: b`), then your content as the body. (`/trellis:setup` offers this move itself when
   it detects such content on a refresh.)
3. Run `/trellis:setup` to refresh — it reads the posture from the frontmatter, asks nothing,
   and rewrites the generated files.

Done when (each checkable): `.trellis/profile.md` ends at its "(Generated from your profile …)"
line; the skill's manifest check passes; `.trellis/expression.md` says exactly what you wrote
(refreshes never touch it — it is excluded from the manifest); and `.trellis/trellis.md` carries
`@expression.md`, so your expression stays always-loaded.

## What it bundles

- **`skills/setup`** — `/trellis:setup`: install or refresh the overlay (done natively, no binary),
  and — only on explicit request — the **M2 morph**: a model-driven rewrite of the project's own
  instructions on a `trellis/morph` git branch, with a recorded rollback point, for the human to
  review (`kodhama-0007` rule 5 moved M2 hosting here from the retired binary).
- **`skills/remove`** — `/trellis:remove`: cleanly reverse the overlay (delete `.trellis/`, strip the
  `CLAUDE.md` block, touch nothing else), and point a morphed project at its git rollback.
- **`reference/`** — the pre-rendered payload (`kodhama-0007`): `invariants.md` (the full signature
  catalog: every invariant with its *why* and a with/without example), every posture variant of the
  overlay files, managed blocks, and `expression.md` seed skeletons, and the checksum manifest the
  setup skill verifies against.
- **`hooks/`** — a `SessionStart` hook that stays quiet until the installed plugin's payload differs
  from the overlay in your project (`decision-0039` rule 1, mechanics per `decision-0043`), then
  nudges you once: *"the overlay may be stale — run `/trellis:setup`."* Binary-free and network-free:
  it compares your project's `.trellis/version` stamp to the installed plugin's `reference/version` —
  file to file — so it can tell you the overlay is *behind the installed plugin*, not how far behind
  the marketplace.

## Removing it

Run `/trellis:remove` — it deletes `.trellis/` and strips the managed `CLAUDE.md` block, leaving your
own content intact; for an M2-morphed project it points you at the recorded git rollback
(`trellis-pre-morph` / `.trellis/rollback`).

## Plugin vs manual copy

This plugin covers Claude Code natively. Every other harness uses the **manual copy path** (repo
README, Get started): the payload in [`reference/`](reference/) is plain files — copy them, paste
the pre-rendered block, verify with `shasum -c`. Same artifact, two mechanical copiers
(`kodhama-0007`).
