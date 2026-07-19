# Trellis (Claude Code plugin)

A governance layer for agentic software development — the native Claude Code way to install it, no
binary required. This is the **primary install path** (`kodhama-0002`; the Homebrew/curl binary
channel retired per `kodhama-0007` rule 5).

## Install

From the kodhama family marketplace — the single front door for the whole family
(`kodhama-0002`; trellis's own in-repo marketplace is retired, `kodhama-0002` open question,
resolved):

```
/plugin marketplace add kodhama/kodhama
/plugin install trellis@kodhama
```

Then run the setup skill in any project:

```
/trellis:setup
```

It asks for a **posture** (conductor / author-adapt — or reads the config from
`.trellis/rules.toml` if the project already carries one, asking nothing) and copies Trellis onto
your project as an **overlay**, split by who owns what (`decision-0051`):

- **`.trellis/` root — yours.** `rules.toml` (the machine-read config: one row per rule, `active =
  true|false`, plus a `strictness` key) and `expression.md` (hand-owned prose: how your project
  expresses the invariants — dials, mappings, gate tables). Both are seeded once from the payload
  and **never rewritten**; editing a row in `rules.toml` *is* the configuration act, and the next
  refresh assembles exactly the rules your rows select. The two floors (`floor-transparency`,
  `floor-intent-gate`) are **floor-held**: their rows exist, but setup includes them regardless
  and says so out loud if you try to turn one off.
- **`.trellis/internal/` — trellis's.** The generated files (`trellis.md`, `rules.md` — the
  assembled active-rules readout, `invariants.md`, the `version` stamp), rewritten verbatim on
  every refresh and verified byte-for-byte against the shipped checksum manifest.

All content is pre-rendered at release; the readout is **assembled** from per-rule payload
fragments by mechanical concatenation, in catalog order (`kodhama-0007`: the skill copies and
concatenates, it never composes). One managed block in your `CLAUDE.md` imports
`.trellis/internal/trellis.md` and your `expression.md`, so both stay always-loaded.
Augment-never-clobber; nothing else is touched, and it's idempotent.

## Migrating an older install

`/trellis:setup`'s refresh **is** the migration vehicle — no flag-day:

- **Flat-layout overlays** (generated files directly in `.trellis/`, from before `decision-0051`):
  a refresh writes the new layout, deletes the old-path copies, and seeds `rules.toml` from the
  legacy `profile:` frontmatter key in `expression.md` (offering to strip the retired key — the
  file is yours, so it never edits without a yes).
- **Hand-authored content in the generated readout** (the clobber target of
  [#112](https://github.com/kodhama/trellis/issues/112) — a refresh rewrites generated files
  whole): setup detects anything after the readout's closing "(Generated from your …" line and
  offers to move it into `.trellis/expression.md`, its hand-owned home, before overwriting.

## What it bundles

- **`skills/setup`** — `/trellis:setup`: install or refresh the overlay (done natively, no binary),
  and — only on explicit request — the **M2 morph**: a model-driven rewrite of the project's own
  instructions on a `trellis/morph` git branch, with a recorded rollback point, for the human to
  review (`kodhama-0007` rule 5 moved M2 hosting here from the retired binary).
- **`skills/remove`** — `/trellis:remove`: cleanly reverse the overlay (delete `.trellis/`, strip the
  `CLAUDE.md` block, touch nothing else), and point a morphed project at its git rollback.
- **`reference/`** — the pre-rendered payload (`kodhama-0007`): `invariants.md` (the full signature
  catalog: every invariant with its *why* and a with/without example), the per-rule fragments in
  `rules/` plus their pre-assembled all-active readout (`rules.md`), the `rules-<p>.toml` posture
  seeds and the `expression.md` prose seed, every posture variant of the header and managed blocks,
  and the checksum manifest the setup skill verifies against.
- **`hooks/`** — a `SessionStart` hook that stays quiet until the installed plugin's payload differs
  from the overlay in your project (`decision-0039` rule 1, mechanics per `decision-0043`), then
  nudges you once: *"the overlay may be stale — run `/trellis:setup`."* Binary-free and network-free:
  it compares your project's `.trellis/internal/version` stamp to the installed plugin's
  `reference/version` — file to file — so it can tell you the overlay is *behind the installed
  plugin*, not how far behind the marketplace. (A stamp still at the legacy flat path
  `.trellis/version` draws the migration nudge.)

## Removing it

Run `/trellis:remove` — it deletes `.trellis/` and strips the managed `CLAUDE.md` block, leaving your
own content intact; for an M2-morphed project it points you at the recorded git rollback
(`trellis-pre-morph` / `.trellis/rollback`).

## Plugin vs manual copy

This plugin covers Claude Code natively. Every other harness uses the **manual copy path** (repo
README, Get started): the payload in [`reference/`](reference/) is plain files — copy them, paste
the pre-rendered block, verify with `shasum -c`. Same artifact, two mechanical copiers
(`kodhama-0007`).
