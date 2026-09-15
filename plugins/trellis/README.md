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
is absent **at project scope**, where the bundle vendored inside the repository is itself the
adoption act (`decision-0070` D3). A **user-scope** install has no such act to read, so **on
Claude** it announces instead and governs nothing until someone writes the file; an ignored
announcement leaves the project ungoverned and returns next session (`decision-0077`). Codex has
no announcement branch — see the Codex note below. No `CLAUDE.md` block, no `AGENTS.md` receipt,
no vendored overlay.

**Where a vendored `.trellis/internal/` still exists it remains authoritative** — though the two
hosts deliver it differently, and saying they behave alike was wrong. **On Claude** the hook detects
the overlay and injects nothing, so the rules arrive exactly once through the static chain. **On
Codex** the hook reads the overlay and injects what it finds — from the overlay's own copy, never
the plugin's. For those
projects the plugin's `reference/` files stay installation sources rather than runtime substitutes, which
is what the previous contract said of every project. One narrow exception, and it moves a pointer
rather than a payload: an overlay carrying no `invariants.md` has that *consulted* pointer
repointed at the plugin's copy, because the alternative is naming a file that is not there
(`decision-0093`). That substitution needs the plugin's copy to be **usable**, not merely present
(`decision-0094`) — otherwise it would trade one dead pointer for another. Where it is not usable
the pointer stays at the overlay's own address, and the **Codex** session is told which file is
unusable and how, and why no copy could stand in, on a channel that costs the injected context
nothing (`decision-0095`). Where the overlay's **own** copy is the unusable one — present but empty or
unreadable — nothing is substituted at all: that file is the project's, at the address the project
chose, so the pointer stays on it and the **Codex** session is told which file is unusable and how,
with the one repair that works on that branch (`decision-0096`). The plugin's copy is not offered
there, because it may not stand in whatever state it is in, and reinstalling would not move the
pointer. **On Claude neither report has a counterpart** — `staleness.sh` exits on any project with
a `.trellis/internal/`, long before either check, while the static import chain still carries the
pointer into the session. A Claude session in either cell is not told, and closing that gap is
tracked separately.

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
boundaries, desktop, IDE, and CI runners — a bare subagent is not a session, so
`SessionStart` never fires for one. **Claude Code cloud sessions are checked** (`TRL-35`, the
Trellis team's Linear issue holding the evidence; Claude Code 2.1.258, fresh containers):
`SessionStart` fires in a cloud container, and the plugin's hook injected the rules on turn one
— **when the environment installed the plugin before Claude Code launched**, through an
environment setup script, and not otherwise. The **automatic project-scope plugin path** —
`enabledPlugins` plus `extraKnownMarketplaces` in the repo's `.claude/settings.json` — does
**not** reach cloud: the container never trusts the folder, so the marketplace is never
registered, and a headless startup never installs project-enabled plugins, while the host's
reconcile reports zero failures. That is host behaviour, not the plugin's. `install.sh
--scope project`'s rendered `.claude/rules/trellis.md` was also loaded on turn one there, with
no plugin present. The recommended cloud path is the opt-in script below. `install.sh`
itself registers no hook — instead it renders
`.claude/rules/trellis.md`, which Claude Code loads at launch with no hook and no trust dialog
(`decision-0068`; the hook stands down when that file is present, so the rules never arrive
twice). **That covers Claude Code and project scope only**: a `--scope personal` install
delivers no rules, and neither does the curl path on **Codex CLI** and other hosts, which
**get nothing from it** — `.claude/rules/` is a Claude mechanism and the installer registers
nothing for any other host (`decision-0068` D7). On those hosts the rules arrive through the
plugin, or not at all. A project-scope run prints both limits. Separately from the host
question, `install.sh` runs only on **macOS, Linux and WSL** — it is a POSIX `sh` script, so
native Windows (cmd, PowerShell) cannot run the curl path at all, whichever host is being
targeted (`decision-0068` §8). **Trellis claims no support.** "Known to work" names
a check that ran; support is not claimed for any host, surface, or version, and nothing here
undertakes to keep any of them working or to repair them if they stop. And **the marketplace
listing has been exercised once**: in that cloud investigation, `claude plugin marketplace add
kodhama/stewards` followed by `claude plugin install trellis@kodhama` installed 0.7.0 in a
2.1.258 cloud container, and the installed plugin's hook then fired. No other recorded check
has exercised installing this package from a marketplace listing, on either host, so a catalog
entry means Trellis is listed — not that the listed install path has been shown to work beyond
that one run.

**Recommended for cloud — the environment setup script, opt-in shape** (`TRL-35`). Install
the plugin, then disable it at **user** scope: the environment then holds the plugin switched
off, and a repo's own `enabledPlugins: { "trellis@kodhama": true }` in `.claude/settings.json`
overrides that and opts in — a repo that declares nothing gets nothing. Verified on turn one in
a fresh, untrusted, headless container:

```bash
#!/bin/bash
C=/opt/claude-code/bin/claude
$C plugin marketplace add kodhama/stewards || true
$C plugin install trellis@kodhama || true
$C plugin disable trellis@kodhama --scope user   # not `|| true` — see below
```

**Always pass `--scope user` to `claude plugin disable`.** Without a scope it defaulted to
project scope and wrote `false` into the repo's tracked `.claude/settings.json` (observed on
2.1.258). **And do not mask that line's exit status.** The two commands above it fail closed —
the worst case is no plugin — but the `disable` is the whole opt-in boundary: masked, a failure
leaves the plugin **enabled** at user scope, so every repo run in that environment gets Trellis,
including one that declares nothing, and the snapshot keeps that state until it rebuilds. In the
run TRL-35 observed the command succeeded, so this is the untested branch rather than a failure
anyone has seen. Two limits: the environment snapshot keeps what the script wrote until it rebuilds,
so the plugin version is pinned between rebuilds, and a session gets this only when it runs in
an environment carrying the script.

There is no per-host disable: `/trellis:remove` removes both host blocks and the shared overlay.

## Install

From the kodhama family marketplace — the single front door for the whole family
(`kodhama-0002`; trellis's own in-repo marketplace is retired, `kodhama-0002` open question,
resolved):

```
/plugin marketplace add kodhama/stewards
/plugin install trellis@kodhama
```

That is the whole install **on Claude Code**, where installing at project scope is the adoption
act — every rule applies immediately, with no further command and no file required
(`decision-0070` D3).

**Codex is not supported yet.** It remains a delivery target (`kodhama-0013`) with **no date
attached**. The machinery is already here — `hooks/codex-context.mjs`, a `.codex-plugin/`
manifest, a catalog entry — and none of it is claimed as a supported path (`kodhama-0021` §2).
Until it is, Codex is carried rather than maintained: its behaviour is not kept in step with the
Claude path, and a difference between them is expected rather than a defect to file. **Reading
`.trellis/rules.toml` is not one of those differences:** on the plugin path, given the same file,
both hooks deliver the same rules text, switch off the same rules and give the same warnings,
pinned by a cross-host test. That closes one named gap; it does not make Codex a supported path.
What a
supported Codex distribution would require is tracked in Linear. Its adoption signal also differs —
`codex-context.mjs` walks up for `.trellis/rules.toml` and reports `project-root-not-found`
when there is none, so both the project-scope default and the user-scope announcement above are
Claude-only (`decision-0070` D7). On Codex the config file is the only adoption signal there is,
which is where `decision-0077` leaves the Claude path too.

**Configuring it is the same on both hosts.** `.trellis/rules.toml` keeps a `[rules]` table, and
only a row set `active = false` has any effect: `<slug> = { active = false }` switches that rule
off. A rule with no row applies, and so does a row that says `active = true`; `strictness` and
`seeded_from` do nothing. Floor rules cannot be switched off. A bad entry — a floor or an unknown
slug set to `active = false`, a row that does not parse, a key the file does not define — is ignored
with a warning in the session, and the file's other rows keep their effect. A new file holds exactly
these two lines, the same two `install.sh` seeds and the hook asks for when a user accepts its
announcement:

```toml
# Every Trellis rule applies. To switch one off, add a row: <slug> = { active = false }
[rules]
```

**With the one-line `governed = false` opt-out** — one top-level line, above any table — no rule
applies, floor rules included (`decision-0070` D5). Deleting that line turns governance back on with
every rule applying, so confirm the intent first. **An existing file keeps working as it is:** its
`strictness`, `seeded_from` and `active = true` rows now do nothing, and nothing asks you to remove
them. Removing rows is safe only once every install that opens the repository runs 0.24.0 or later,
because installed plugin versions are pinned per checkout and 0.6.0 and older refuse a file with
missing rows. **A legacy delivery shape keeps every row:** a vendored `.trellis/internal/` overlay,
an inline managed block, and a `.claude/rules/trellis.md` rendered by an older installer carry rules
text that applies a rule only when its row says `active = true`, so those projects keep a full row
set until they migrate.

Older projects still carry an **overlay**, split by who owns what (`decision-0051`):

- **`.trellis/` root — yours.** `rules.toml` alone (the machine-read config: one row per rule,
  `active = true|false`), **never rewritten**; editing a row *is* the configuration act, and it
  takes effect **immediately** — the readout ships complete with an authority header, and your rows
  govern which rules apply at read time (`decision-0053`); each rule in the readout ends with its
  row's slug, so the two are matchable. Keep a row for every rule on this shape: the overlay's
  authority header applies a rule only when its row says `active = true`.
  The two floors (`floor-transparency`, `floor-intent-gate`) have rows too, but the
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
managed block from your instructions file, keeping `.trellis/rules.toml`. An inline block with no
`.trellis/rules.toml` beside it holds its rows inside the block: before deleting it, write the
two-line file shown under Install, followed by each row the block sets to `active = false`. The
plugin then delivers the rules and the hook stops nudging. Three cases the retired refresh used to handle,
and what to do about each yourself:

- **Flat-layout overlays** (generated files directly in `.trellis/`, from before `decision-0051`):
  delete the old-path copies. If there is no `.trellis/rules.toml`, write the two-line file shown
  under Install rather than recovering rows from the legacy `profile:` key in `expression.md`.
- **A leftover `expression.md`** (seeded before the amendment retired it): move any hand-written
  body into your own instructions file, outside the managed block, then delete the file. Nothing
  reads it any more, so leaving it in place is harmless but inert.
- **Hand-authored content in the generated readout** (the clobber target of
  the historical clobber target): moot on the plugin path since
  `decision-0065` — the setup skill no longer wrote generated files, so there was nothing to
  rewrite whole.
  It survives as a concern for `install.sh`, which does vendor them.

## What it bundles

- **`skills/remove`** — `/trellis:remove`: cleanly reverse every delivery state in
  `decision-0073`'s closed set — the managed blocks in the documented instruction files, the
  rendered `.claude/rules/trellis.md`, a vendored bundle at `.claude/skills/trellis/` (behind
  confirmation), then `.trellis/` — and point a morphed project at its git rollback.
- **`reference/`** — the pre-rendered payload (`kodhama-0007`): `invariants.md` (the full signature
  catalog: every invariant with its *why* and a with/without example), the complete rules readout
  (`rules.md`, opened by the live-rows authority header), the rules header (`trellis.md`), the
  managed blocks (`block-claude.md`, `block-codex.md`, and `block-inline.md` with its head and tail
  parts), and the checksum manifest `install.sh` verifies against.
- **`hooks/`** — host-isolated hooks: Claude's `SessionStart` staleness hook stays quiet until the installed plugin's payload differs
  from the overlay in your project (`decision-0039` rule 1, mechanics per `decision-0043`), then
  nudges you once, with the manual migration steps. Binary-free and network-free:
  it compares your project's `.trellis/internal/version` stamp to the installed plugin's
  `reference/version` — file to file — so it can tell you the overlay is *behind the installed
  plugin*, not how far behind the marketplace. (A stamp still at the legacy flat path
  `.trellis/version` draws the migration nudge. Codex separately registers a startup-only context
  hook that validates and transports the installed overlay.)

## Removing it

Run `/trellis:remove` — it removes the rendered `.claude/rules/trellis.md` first (the skill's
load-bearing order: no interruption may strand that always-loaded file without its rows), then
strips the managed blocks from the documented instruction files, deletes a project-scope
vendored bundle at `.claude/skills/trellis/` behind your confirmation, then deletes `.trellis/`,
leaving your own content intact; for an M2-morphed project it points you at the recorded git
rollback (`trellis-pre-morph` / `.trellis/rollback`).

## Local QA

From the repository root, `npm ci && npm run test:plugin-smoke` runs both SessionStart hooks against
a temporary healthy project and verifies that each delivers a complete rules payload. The smoke
path uses no live credentials or network services and removes its temporary project at exit.
`npm run quality` adds the repository's formatter, JavaScript, ShellCheck, Go formatting, module,
and vet checks.

## Plugin vs manual copy

This plugin covers Claude Code and the Phase 1 trusted-local Codex startup boundary described
above. Other surfaces use the **manual copy path** (repo README, Get started): the payload in
[`reference/`](reference/) is plain files — copy them, paste the pre-rendered block, verify with
`shasum -c`. Same artifact, multiple mechanical carriers (`kodhama-0007`).
