# Trellis (Claude Code and local Codex plugin)

A governance layer for agentic software development — the native Claude Code way to install it, no
binary required. This is the **primary install path** (`kodhama-0002`; the Homebrew/curl binary
channel retired per `kodhama-0007` rule 5).

## Phase 1 host support

Claude keeps its existing `CLAUDE.md` imports and staleness hook. The Codex branch supports
setup/refresh, **product-wide** remove, and **fresh startup** in a **trusted local Codex**
repository. It installs the generated `block-codex.md` receipt and **best-effort** fallback in
`AGENTS.md` and registers only `SessionStart(startup)`. Installed project files are authoritative;
the plugin's `reference/` files are setup sources, never runtime substitutes.

Native Codex delivery requires local **Node.js 20** or newer. Without it, setup reports
bootstrap-only degradation and leaves the installed-file fallback usable. Trellis requires no
project runtime, daemon, or network service. Row edits take effect at the next supported host
context-loading boundary without refresh, never in a context already in flight.

Phase 1 excludes Codex resume, clear, compact, subagent boundaries, desktop, IDE,
headless/automation, and cloud surfaces. There is no per-host disable: `/trellis:remove` removes
both host blocks and the shared overlay. Ordinary refresh preserves rows, strictness, and
`seeded_from`; it is not a confirmed preset reset. A Claude-hook replacement, every other
host-native transport, and revival of the parked `seed` or `custom` presets are also excluded.

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

- **`.trellis/` root — yours.** `rules.toml` alone (the machine-read config: one row per rule,
  `active = true|false`, plus a `strictness` key), seeded once from the payload and **never
  rewritten**; editing a row *is* the configuration act, and it takes effect **immediately** —
  the readout ships complete with an authority header, and your rows govern which rules apply at
  read time (`decision-0053`); each rule in the readout ends with its row's slug, so the two are
  matchable. The two floors (`floor-transparency`, `floor-intent-gate`) have rows too, but the
  floor rules apply regardless of their value — setup says so out loud if you try to turn one
  off, never silently honoring the row. (There is no `expression.md`: it retired with the
  `decision-0051` amendment — your governance prose belongs in your own instructions file, which
  every harness already loads.)
- **`.trellis/internal/` — trellis's.** The generated files (`trellis.md`, `rules.md` — the
  complete rules readout, `invariants.md`, the `version` stamp), rewritten verbatim on every
  refresh and verified byte-for-byte against the shipped checksum manifest.

All content is pre-rendered at release (`kodhama-0007`: the skill copies and verifies, it never
composes). One managed block in your `CLAUDE.md` imports `.trellis/internal/trellis.md` **and**
`.trellis/rules.toml`, so the rules and your rows stay always-loaded and a row edit governs the
very next session. Augment-never-clobber; nothing else is touched, and it's idempotent.

## Migrating an older install

`/trellis:setup`'s refresh **is** the migration vehicle — no flag-day:

- **Flat-layout overlays** (generated files directly in `.trellis/`, from before `decision-0051`):
  a refresh writes the new layout, deletes the old-path copies, and seeds `rules.toml` from the
  legacy `profile:` frontmatter key in `expression.md`.
- **A leftover `expression.md`** (seeded before the amendment retired it): **never silently
  deleted** — a refresh preserves any hand-written body and *offers* to move it into your own
  instructions file (outside the managed block), or to leave the file in place; a pure seed stub
  may be offered for deletion.
- **Hand-authored content in the generated readout** (the clobber target of
  [#112](https://github.com/kodhama/trellis/issues/112) — a refresh rewrites generated files
  whole): setup compares generated files against the payload — and, in a legacy readout, detects
  anything after its closing "(Generated from your …" line (the retired footer that older
  installs still carry) — and offers to move hand-authored content into your own instructions
  file before overwriting.

## What it bundles

- **`skills/setup`** — `/trellis:setup`: install or refresh the overlay (done natively, no binary),
  and — only on explicit request — the **M2 morph**: a model-driven rewrite of the project's own
  instructions on a `trellis/morph` git branch, with a recorded rollback point, for the human to
  review (`kodhama-0007` rule 5 moved M2 hosting here from the retired binary).
- **`skills/remove`** — `/trellis:remove`: cleanly reverse the overlay (strip the Claude and Codex
  blocks, then delete `.trellis/`, touching nothing else), and point a morphed project at its git
  rollback.
- **`reference/`** — the pre-rendered payload (`kodhama-0007`): `invariants.md` (the full signature
  catalog: every invariant with its *why* and a with/without example), the complete rules readout
  (`rules.md`, opened by the live-rows authority header), the `rules-<p>.toml` posture seeds,
  every posture variant of the header and managed blocks, and the checksum manifest the setup
  skill verifies against.
- **`hooks/`** — host-isolated hooks: Claude's `SessionStart` staleness hook stays quiet until the installed plugin's payload differs
  from the overlay in your project (`decision-0039` rule 1, mechanics per `decision-0043`), then
  nudges you once: *"the overlay may be stale — run `/trellis:setup`."* Binary-free and network-free:
  it compares your project's `.trellis/internal/version` stamp to the installed plugin's
  `reference/version` — file to file — so it can tell you the overlay is *behind the installed
  plugin*, not how far behind the marketplace. (A stamp still at the legacy flat path
  `.trellis/version` draws the migration nudge. Codex separately registers a startup-only context
  hook that validates and transports the installed overlay.)

## Removing it

Run `/trellis:remove` — it strips the managed blocks in `CLAUDE.md` and `AGENTS.md`, then deletes
`.trellis/`, leaving your own content intact; for an M2-morphed project it points you at the recorded git rollback
(`trellis-pre-morph` / `.trellis/rollback`).

## Plugin vs manual copy

This plugin covers Claude Code natively. Every other harness uses the **manual copy path** (repo
README, Get started): the payload in [`reference/`](reference/) is plain files — copy them, paste
the pre-rendered block, verify with `shasum -c`. Same artifact, two mechanical copiers
(`kodhama-0007`).
