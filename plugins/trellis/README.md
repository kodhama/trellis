# Trellis plugin for Claude Code and local Codex

A governance layer for agentic software development, delivered as one dual-host plugin package
with no binary required. This is the **primary install path** (`kodhama-0002`; the Homebrew/curl
binary channel retired per `kodhama-0007` rule 5).

## Package identity

[`VERSION`](VERSION) is the sole plugin-package SemVer authority. Both host manifests carry that
same value, guarded by the repository's `TestPluginPackageParity` Go test. The generated overlay
keeps its separate content identity in
[`reference/version`](reference/version): package SemVer identifies the plugin package, while the
`payload@…` stamp identifies the exact generated rule bytes.

## Phase 1 host support

Since `decision-0065` both hosts deliver the same way: a `SessionStart` hook injects the rules
from the plugin's `reference/` payload plus the project's `.trellis/rules.toml`. On the plugin
path nothing writes that config file — the hook reads it, and applies the shipped defaults when it
is absent (`decision-0070` D3). No `CLAUDE.md` block, no `AGENTS.md` receipt, no vendored overlay.

**Where a vendored `.trellis/internal/` still exists it remains authoritative**, on both hosts:
the hooks detect it, read from it, and inject nothing, so the rules arrive exactly once. For those
projects the plugin's `reference/` files stay setup sources rather than runtime substitutes, which
is what the previous contract said of every project.

Native Codex delivery requires local **Node.js 20** or newer, and is unsupported either way —
see below. Trellis requires no
project runtime, daemon, or network service. Row edits take effect at the next supported host
context-loading boundary without refresh, never in a context already in flight.

**Where Trellis is known to work, and what that does not mean.** The hosts Trellis is known to
work on are **Claude Code** and the trusted-local **Codex CLI** startup boundary — no others.
What establishes that is a check, not an assurance: `decision-0065` moved rule delivery to the
plugin's `SessionStart` hooks on both hosts, and measured against a real session with file tools
disabled, `startup` and `resume` fire and the injected rules reach the model, including under
headless `claude -p`. **Not verified on any surface:** `compact`, `clear`, `fork`, subagent
boundaries, desktop, IDE, cloud, and CI runners — a bare subagent is not a session, so
`SessionStart` never fires for one. `install.sh` registers no hook either — instead it renders
`.claude/rules/trellis.md`, which Claude Code loads at launch with no hook and no trust dialog
(`decision-0068`; the hook stands down when that file is present, so the rules never arrive
twice). **That covers Claude Code and project scope only**: a `--scope personal` install
delivers no rules, and neither does the curl path on **Codex CLI** and other hosts, which
**get nothing from it** — `.claude/rules/` is a Claude mechanism and the installer registers
nothing for any other host (`decision-0068` D7). On those hosts the rules arrive through the
plugin, or not at all. A project-scope run prints both limits. **Trellis claims no support.** "Known to work" names
a check that ran; support is not claimed for any host, surface, or version, and nothing here
undertakes to keep any of them working or to repair them if they stop. And **no marketplace
install has been evidenced**: no recorded check has exercised installing this package from a
marketplace listing, on either host, so a catalog entry means Trellis is listed — not that the
listed install path has been shown to work.

Applying a preset **replaces** rows, strictness and `seeded_from`, so replacing a file you
already have shows up as a git diff — provided the file is committed, which nothing does for you. There is no per-host disable: `/trellis:remove` removes both
host blocks and the shared overlay. The parked `seed` and `custom` presets stay parked.

## Install

From the kodhama family marketplace — the single front door for the whole family
(`kodhama-0002`; trellis's own in-repo marketplace is retired, `kodhama-0002` open question,
resolved):

```
/plugin marketplace add kodhama/kodhama
/plugin install trellis@kodhama
```

That is the whole install **on Claude Code**, where installing at project scope is the adoption
act — the shipped defaults apply immediately, all fourteen rules at the adaptive posture, with no
further command and no file required (`decision-0070` D3).

**Codex is not supported yet.** It remains a delivery target (`kodhama-0013`) with **no date
attached**. The machinery is already here — `hooks/codex-context.mjs`, a `.codex-plugin/`
manifest, a catalog entry — and none of it is claimed as a supported path (`kodhama-0021` §2).
Until it is, Codex is carried rather than maintained: its behaviour is not kept in step with the
Claude path, and a difference between them is expected rather than a defect to file. `#220` holds
what a supported Codex distribution would require. Its adoption signal also differs —
`codex-context.mjs` walks up for `.trellis/rules.toml` and reports `project-root-not-found`
when there is none, so the project-scope default above is Claude-only (`decision-0070` D7).

On Claude, change the posture by copying a **complete** preset and editing it —
`reference/rules-a.toml` for firm, `rules-b.toml` for adaptive — to `.trellis/rules.toml`,
then `active = false` on any row you want off. The hook validates the row set against what the
plugin ships and injects **nothing** when a slug is missing, so a hand-written partial file
leaves the project ungoverned rather than firm. Older projects still carry an
**overlay**, split by who owns what (`decision-0051`):

- **`.trellis/` root — yours.** `rules.toml` alone (the machine-read config: one row per rule,
  `active = true|false`, plus a `strictness` key), seeded once from the payload and **never
  rewritten**; editing a row *is* the configuration act, and it takes effect **immediately** —
  the readout ships complete with an authority header, and your rows govern which rules apply at
  read time (`decision-0053`); each rule in the readout ends with its row's slug, so the two are
  matchable. The two floors (`floor-transparency`, `floor-intent-gate`) have rows too, but the
  floor rules apply regardless of their value, and the injected readout says so rather than
  silently honoring a row set false. (There is no `expression.md`: it retired with the
  `decision-0051` amendment — your governance prose belongs in your own instructions file, which
  every harness already loads.)
- **`.trellis/internal/` — trellis's.** The generated files (`trellis.md`, `rules.md` — the
  complete rules readout, `invariants.md`, the `version` stamp), rewritten verbatim on every
  refresh and verified byte-for-byte against the shipped checksum manifest.

All content is pre-rendered at release (`kodhama-0007`: writers copy and verify, they never
compose). One managed block in your `CLAUDE.md` imports `.trellis/internal/trellis.md` **and**
`.trellis/rules.toml`, so the rules and your rows stay always-loaded and a row edit governs the
very next session. Augment-never-clobber; nothing else is touched, and it's idempotent.

## Migrating an older install

Migration is a manual edit since `decision-0072` retired the setup skill. Delete
`.trellis/internal/` (or the pre-`decision-0051` flat files directly in `.trellis/`) and the
managed block from your instructions file, keeping `.trellis/rules.toml`. The plugin then
delivers the rules and the hook stops nudging. Three cases the retired refresh used to handle,
and what to do about each yourself:

- **Flat-layout overlays** (generated files directly in `.trellis/`, from before `decision-0051`):
  delete the old-path copies. If there is no `.trellis/rules.toml`, copy the shipped preset
  (`reference/rules-b.toml`) rather than recovering rows from the legacy `profile:` key in
  `expression.md` — the preset is the current row set.
- **A leftover `expression.md`** (seeded before the amendment retired it): move any hand-written
  body into your own instructions file, outside the managed block, then delete the file. Nothing
  reads it any more, so leaving it in place is harmless but inert.
- **Hand-authored content in the generated readout** (the clobber target of
  [#112](https://github.com/kodhama/trellis/issues/112)): moot on the plugin path since
  `decision-0065` — setup no longer writes generated files, so there is nothing to rewrite whole.
  It survives as a concern for `install.sh`, which does vendor them.

## What it bundles

- **`skills/remove`** — `/trellis:remove`: cleanly reverse the overlay (strip the Claude and Codex
  blocks, then delete `.trellis/`, touching nothing else), and point a morphed project at its git
  rollback.
- **`reference/`** — the pre-rendered payload (`kodhama-0007`): `invariants.md` (the full signature
  catalog: every invariant with its *why* and a with/without example), the complete rules readout
  (`rules.md`, opened by the live-rows authority header), the `rules-<p>.toml` posture seeds,
  every posture variant of the header and managed blocks, and the checksum manifest
  `install.sh` verifies against.
- **`hooks/`** — host-isolated hooks: Claude's `SessionStart` staleness hook stays quiet until the installed plugin's payload differs
  from the overlay in your project (`decision-0039` rule 1, mechanics per `decision-0043`), then
  nudges you once, with the manual migration steps. Binary-free and network-free:
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

This plugin covers Claude Code and the Phase 1 trusted-local Codex startup boundary described
above. Other surfaces use the **manual copy path** (repo README, Get started): the payload in
[`reference/`](reference/) is plain files — copy them, paste the pre-rendered block, verify with
`shasum -c`. Same artifact, multiple mechanical carriers (`kodhama-0007`).
