# Trellis

> A trellis is structure that *enables* growth rather than dictating form.

**Trellis is the governing layer for agentic software development.** It sits *above* whatever
methodology your project already uses — Spec Kit, BMAD, your own — learns its shape, and governs a
small set of **invariants**: the handful of properties that must stay true no matter how fast
autonomous agents move. It doesn't replace your process; it keeps it honest.

It exists for one failure mode: **fast agents drift.** They skip reviews, resolve ambiguity by
guessing, grade their own work, and leave no trace of where a decision came from. Trellis makes a few
things non-negotiable — and surfaces every time something bends — without dictating how you build.

## Get started

**Claude Code (the primary path)** — install the plugin from the kodhama family marketplace.
That is the whole install:

```
/plugin marketplace add kodhama/stewards
/plugin install trellis@kodhama
```

Installing the plugin at **project scope is the adoption act** — no further command, no file
required (`decision-0070`). The shipped defaults apply: all rules, adaptive posture.
To change that, the recipe depends on whether you already have a `.trellis/rules.toml`:

- **No file yet** — copy a complete preset from the installed plugin: `reference/rules-a.toml`
  for the firm posture, `rules-b.toml` for adaptive. Then set `active = false` on any row you
  want off (rows govern at read time, `decision-0053`). Copying the whole file is still the clean
  start, but a partial one is no longer fatal on **either host**: the hook validates your rows
  against the shipped set and, on a mismatch, **reconciles** instead of refusing — a missing slug
  is delivered `active = true`, the session is governed from the reconciled set, and the agent
  writes that set back to your file and tells you what it changed (`decision-0083`,
  `decision-0084`). A file with no `strictness` at all is fine on both: the posture falls to
  adaptive rather than the file being rejected.
- **File already there** — edit it **in place**: change `strictness` (and `seeded_from` to
  match), flip individual rows. **Do not copy a preset over it** — both presets set every row
  active, so a posture-only change made by copying silently re-enables every rule you turned
  off.
- **The file is the one-line `governed = false` opt-out** — Trellis is switched off for this
  project (`decision-0070` D5). Editing `strictness` beside the opt-out leaves it in force and the
  hook stays silent, so that is not re-enabling. Deleting the line alone **does** re-enable
  governance — but the file is then empty, so it reconciles to **all rows active at the
  adaptive posture**, whatever this project ran before, on either host (`decision-0083`,
  `decision-0084`). Confirm that turning governance back on is what you want, then write a
  complete preset over it if you want a posture or a row set of your own. The plugin vendors
nothing. The rules themselves arrive at session start, injected by the plugin's own hook from the
plugin's payload, so there is no copy in your repo to install, refresh or let drift
(`decision-0065`). A project set up before that change still carries a vendored `.trellis/`
bundle and a managed block in your `CLAUDE.md`, and keeps working — the hook detects the overlay
and steps aside, so the rules still arrive exactly once. To migrate it, delete the overlay and the
managed block, keeping `.trellis/rules.toml` — **and the overlay's path depends on its age**:
`.trellis/internal/` for a `decision-0051` layout, `.trellis/trellis.md` plus `.trellis/version`
for a flat one from before it. Deleting only `internal/` on a flat-layout project removes the
managed block and leaves the legacy stamp, which the hook then detects — still ungoverned, one
repair later.
That earlier behaviour — the managed block, the vendored bundle, and an optional **M2 morph**
rewriting your own instructions on a fresh git branch — is retired for new installs; only existing
consumers still carry it. The plugin lives in [`plugins/trellis`](plugins/trellis).

**Same plugin, without the marketplace — the curl path.** `install.sh` vends the whole
`plugins/trellis/` tree onto disk as a [skills-directory
plugin](https://code.claude.com/docs/en/plugins-reference#skills-directory-plugins): any folder
under a skills directory with its own `.claude-plugin/plugin.json` loads as `trellis@skills-dir` on
Claude Code's next session, no marketplace and no install step. On **project scope** it also
renders one file it wholly owns, `.claude/rules/trellis.md` — the rules themselves, which Claude
loads at launch with no hook and no *plugin* trust prompt (the workspace-trust dialog on first
launch in a project still applies — see the project-scope bullet below) — and seeds `.trellis/rules.toml` from the shipped
preset when none exists, so the project is governed at 16/16 on the adaptive posture the moment the
script exits (`decision-0070` D2). A project whose `.trellis/rules.toml` declares `governed = false`
has opted out (`decision-0070` D5): it gets the bundle and **no** rules file, exactly as the plugin
hook injects nothing there (`TRL-38`). That is how the rules actually reach a session:
`decision-0068` measured that the vendored bundle alone delivered **none**, which is
why this paragraph no longer says the script "composes nothing else". It is Claude Code only, and
`--scope personal` delivers no rules at all; each run prints whichever limit applies to it.
Governance starts with adoption, not with a command — editing `.trellis/rules.toml` is how you
change it afterwards (`decision-0070`; the setup skill retired by `decision-0072`):

```sh
curl -fsSL https://raw.githubusercontent.com/kodhama/trellis/main/install.sh | sh
```

**Platform for this path — macOS, Linux, WSL** (`decision-0068` §8,
[#210](https://github.com/kodhama/trellis/issues/210)). `install.sh` is a POSIX `sh`
script, so cmd and PowerShell cannot run it at all; a Windows user runs it inside WSL.
A POSIX layer such as Git Bash or MSYS will also run it, but sits outside the boundary
named above and no check here covers it — WSL is the route this repo states. The shell
is the whole of the barrier: the script creates no symlink, so no filesystem or
developer-mode behaviour rides on top of it. Running natively on Windows would take a
separate installer — `decision-0068` §8 names "a future PowerShell or Go installer" —
not a flag here.

Two scopes, `--scope project` (default, inside a git repo) or `--scope personal`:

- **project** — `<repo-root>/.claude/skills/trellis/`, checked into git so it reaches every
  collaborator on clone. Resolved via `git rev-parse --show-toplevel` from wherever you run it, not
  `$PWD` — project-scope skills-directory plugins do **not** walk up to the repo root, so launch
  Claude Code from that root (or run `/reload-plugins` after `cd`'ing there), and expect its
  workspace **trust dialog** on first launch there (unavoidable — the content came from the repo,
  not from you). The script never runs `git add`/`git commit`; it prints the command and leaves the
  commit to you.
- **personal** — `~/.claude/skills/trellis/`, available in every project on the machine, no trust
  dialog, no repo required — and never shells out to git at all when passed explicitly.

Outside a git repo, with no `--scope`/`$TRELLIS_SKILLS_SCOPE` given, project scope has no target:
the script prompts once if a terminal is available (offering personal scope, or the chance to
abort), and otherwise **fails closed** — non-zero exit, nothing written, naming exactly what's
missing — rather than silently picking a scope you didn't ask for.

Every fetched byte is verified against a manifest baked into the script *before* anything is
written — a mismatch fails loudly and installs nothing. Inspect first, or pass flags, instead of
piping straight to `sh`:

```sh
curl -fsSLO https://raw.githubusercontent.com/kodhama/trellis/main/install.sh
less install.sh && sh install.sh --scope personal
```

Neither path needs a command to become governed — the curl path seeds the rows and a
project-scoped plugin applies the shipped defaults without any file at all (`decision-0070`).
When you want the **firm** posture instead of the default adaptive one, or to turn individual
rules off: copy the matching preset (`reference/rules-a.toml` / `rules-b.toml`) if you have no
`.trellis/rules.toml` yet, otherwise edit the one you have in place. The row set must stay
complete, and copying over an existing file discards every row you disabled.

### Local Codex — carried, not supported

**Codex is not supported yet.** It remains a delivery target (`kodhama-0013`) with **no date
attached**, and until it is claimed, nothing here is kept in step with the Claude path
(`kodhama-0028`; `kodhama-0021` §2). What a supported Codex distribution would require is
tracked in Linear.

What exists: since `decision-0065` the plugin path works the same shape on both hosts — a
`SessionStart(startup)` hook injects the rules from the plugin's own payload together with the
project's `.trellis/rules.toml`. The **curl** path is Claude-only, delivering through
`.claude/rules/`, which is a Claude mechanism (`decision-0068` D7). Adoption also differs: on
Codex the config file is the adoption signal, so there is no project-scope default
(`decision-0070` D7).

A project that still carries a vendored `.trellis/internal/` overlay is read from that overlay
instead, on both hosts, so nothing is delivered twice and no existing consumer breaks. The hook
stands down the same way for `.claude/rules/trellis.md` when the curl path has already delivered
the rules — and if it ever finds a project holding **both** static shapes, it says so rather than
choosing one. A valid row
edit is seen at the next startup without refresh, and does not change a context already in
flight.

Native Codex delivery requires local **Node.js 20** or newer; without it the hook cannot run and
nothing reports that, which is one of the things the tracked Codex-support work covers. Trellis adds no project runtime,
daemon, or network service. Native hook success is stronger than the fallback: fallback execution
remains model-directed rather than deterministic.

Phase 1 does not support Codex resume, clear, compact, subagent boundaries, desktop, IDE,
headless/automation, or cloud surfaces. It adds no per-host disable: `/trellis:remove` removes both
host blocks and the shared overlay. Applying a preset **replaces** the consumer's rows, strictness
and `seeded_from`, so replacing a file you already have shows up as a git diff — provided the file
is committed, which the installer suggests but never does for you (`decision-0072` §3). Also excluded are any other host-native
transport, and revival of the parked `seed` or `custom` presets.

**Any other harness — the manual copy path.** This is for harnesses the plugin does **not** cover.
On **Claude Code** it is superseded — use the marketplace or curl paths above
(`decision-0069`). **On Codex CLI it is not**: there is no way to install Trellis there yet
(`/plugin` commands are Claude Code's, and the curl path is Claude-only), so the manual copy is
what Codex has until a real channel exists. It is also **mutually exclusive with the curl path**: a repo that already
carries a hand-built overlay or a managed block makes `install.sh` refuse to render the rules
file, on purpose, because both would load and deliver the rules twice.

Every bundle file is pre-rendered plain text in
[`plugins/trellis/reference/`](plugins/trellis/reference) (the payload, `kodhama-0007`: one
render, many copiers). Pick a posture key (`a` = conductor, `b` = author-adapt) and copy:

```sh
git clone --depth 1 https://github.com/kodhama/trellis /tmp/trellis
ref=/tmp/trellis/plugins/trellis/reference   # <p> below: a (conductor) | b (author-adapt)
mkdir -p .trellis
cp "$ref"/rules-<p>.toml .trellis/rules.toml          # first install only — yours after that
```

Then pick **exactly one** delivery branch — the two blocks below are alternatives, never both
(both at once is the overlay-plus-inline conflict the installer refuses and the hook alarms on).

**@import-capable file (CLAUDE.md)** — vendor the overlay the block imports:

```sh
mkdir -p .trellis/internal
cp "$ref"/invariants.md  .trellis/internal/invariants.md
cp "$ref"/rules.md       .trellis/internal/rules.md   # the complete rules readout
cp "$ref"/trellis-<p>.md .trellis/internal/trellis.md
cp "$ref"/version        .trellis/internal/version
cat "$ref"/block-claude.md >> CLAUDE.md
sed -n -e 's|  invariants\.md$|  .trellis/internal/invariants.md|p' \
       -e 's|  rules\.md$|  .trellis/internal/rules.md|p' \
       -e 's|  trellis-<p>\.md$|  .trellis/internal/trellis.md|p' \
       -e 's|  version$|  .trellis/internal/version|p' \
       "$ref"/checksums | shasum -a 256 -c -           # verify: all four lines print OK
```

**No @import support (e.g. AGENTS.md)** — append the SELF-CONTAINED inline block instead. It
embeds the rules and the rows and takes NO `.trellis/internal/` copies: the block plus
`.trellis/rules.toml` is the whole install, and copying the overlay beside it would deliver the
rules twice.

```sh
# the block must start at column 0 of its own line: guard against a file
# whose last line has no trailing newline before appending
[ -s AGENTS.md ] && [ -n "$(tail -c1 AGENTS.md)" ] && echo >> AGENTS.md
(cd "$ref" && grep '  block-inline-<p>\.md$' checksums | shasum -a 256 -c -)  # verify: OK
cat "$ref"/block-inline-<p>.md >> AGENTS.md
```

To deactivate a rule later, set its row in `.trellis/rules.toml` to `active = false` — that's
it (`decision-0053`): the readout ships complete and opens with an authority header, so agents
apply a rule only where its row says `active = true`, and a row edit takes effect at the next
host context-loading boundary.
**A row starting with `#` is a quarantined row** — the hook found a slug the installed payload does
not ship and commented it out, with the date and the payload stamp, rather than deleting it
(`decision-0083`). It is inert and safe to leave. If a newer Trellis release ships that slug, update
the plugin and uncomment the row; if the rule was retired, delete the line whenever you like.
The two `floor-*` rows apply regardless of their value. On an **import** install the block loads
your current `rules.toml` every session. On an **inline** install the block carries a copy of
the rows inlined below the rules, so re-paste that copy after editing: replace everything
between the `trellis:begin`/`trellis:end` markers with the sandwich `cat
"$ref"/block-inline-<p>-head.md "$ref"/rules.md`, then an `## Active rows (`.trellis/rules.toml`)`
heading followed by your `rules.toml` inside a ```toml fence, then `cat
"$ref"/block-inline-tail.md` (the shipped `block-inline-<p>.md` is that sandwich with the seed
rows).

No binary or project runtime — the assets are plain files, and anything can verify them with
`shasum -c` against the shipped manifest. The optional local Codex native transport uses the
host's Node.js runtime as documented above. (The Homebrew/curl binary channel retired in
`kodhama-0007` rule 5;
the Go code in [`cli/`](cli/) survives as the release-time payload generator only.)

## The model

Deliberately tiny — small enough that a newcomer, human or agent, can read it and know how to make a
change that will pass.

- **Structural gate** — a four-point admission check (one-way flow, handover points, a human
  intent locus, checkable artifacts). If a process lacks the shape, Trellis says so *loudly*.
- **Operating layer** — what Trellis supplies: a gate at every handover, independent
  verification (*the builder does not grade itself*), an auditable archive, bounded context,
  clarify-before-commit.
- **Two dials** — per gate: *how strict* (`expressed` → `default-on-but-skippable` → `enforced`,
  the `dial-enforcement-strength` values from `core/invariants/trellis-invariants-v1.md:234`; an
  earlier wording here invented "documented → default-on", which matched no surface) and *who checks*
  (an agent, a human, or nobody). The same core serves a weekend hack and a regulated pipeline.
- **Two floors** — the only settings that never dial to zero: every consequential choice is
  **surfaced**, and the **human intent gate never fully opens**.

The full set: [`core/invariants/trellis-invariants-v1.md`](core/invariants/trellis-invariants-v1.md).
Every invariant with its **why** and a **with/without** example at several layers lives in the
[**signature catalog**](core/catalog/signature-catalog-v1.md) — the single source (rendered readable on
the [project site](https://kodhama.github.io/trellis/invariants.html)). The thesis behind it:
[`agentic-dev-meta-layer-brief.md`](agentic-dev-meta-layer-brief.md).

## Two ways to run it

- **Advisor** *(open, no runtime — shipped)* — Trellis composes onto your project as instructions your
  agents **consult**; nothing of Trellis runs at agent-time. This is what the plugin (or the
  manual copy path) installs today. On the plugin and curl paths that is `.trellis/rules.toml` plus
  plugin- or curl-delivered rules — neither writes an overlay. The **M1 overlay is now written only
  by the manual copy path**, which `decision-0069` retains for harnesses the plugin does not cover;
  it is not legacy. The **M2 morph is retired outright** and survives only where it already ran.
  Nothing to secure or remove at runtime.
- **Supervisor** *(installed, live — in progress)* — Trellis wired into your pipeline: gates fire on
  commit/PR events via hooks, it stays current through an update channel, and it comes off cleanly.
  The next delivery slice.

These are the two ends of the delivery relationship; the cross-lens vocabulary lives in
[`core/lexicon.md`](core/lexicon.md).

## Where it stands

Built in the open, dogfooded on itself from commit one. The honest state:

- **Ratified** — the invariant set (`invariants-v1`), 40+ decisions, 8+ research notes.
- **Shipped** — the **Claude Code plugin** (marketplace install, `/trellis:remove`, a bundled
  staleness hook) riding a **pre-rendered, checksum-manifested
  payload** (`kodhama-0007`: render once at release, writers only copy and verify), plus the
  documented **manual copy path** for any other harness. It stands on the *spine* + an
  **independent conformance check** (`core/rubrics/artifact-contract.md`, running on this repo),
  the expression-profile + catalog **schema** (`core/schemas/typed-artifacts.md`), the populated catalog and
  the first per-project **profile** (instance #1), and the cross-lens **lexicon**. The v0 setup **CLI** (the retired binary, unrelated to the retired setup skill) shipped first (`v0.1.0`–`v0.2.29`) and its end-user channel retired in favor of the above
  (`kodhama-0007` rule 5, `decision-0043`); the Go code survives as the release-time payload
  generator.
- **In progress** — **supervisor mode** (installed live gates).
- **The open risk** — the invariants are validated on essentially *one* project. **Instance #2** — a
  second, different project — is the next real test of whether they generalize.

## Repo map

| Path | What |
|---|---|
| [`agentic-dev-meta-layer-brief.md`](agentic-dev-meta-layer-brief.md) | The full thesis (start at §10 verdict, §11 start-here, §12 operating method). |
| [`core/`](core/) | The shippable product: invariants, the conformance rubric, the signature catalog, the lexicon. |
| [`cli/`](cli/) | The **payload generator** (Go) — `trellis payload` renders the pre-built bundle + manifest at release; its tests are the CI sync-guards. Generator-only since `decision-0043`. |
| [`plugins/trellis/`](plugins/trellis/) | The **Claude Code and local Codex plugin** — `/trellis:remove`, host-isolated hooks, and the vendored payload (`reference/`). |
| [`install.sh`](install.sh) | The **curl path** — vends the whole plugin bundle onto disk as a skills-directory plugin, and on project scope renders `.claude/rules/trellis.md`, the file that actually delivers the rules (`decision-0068`). Claude Code only. |
| [`decisions/`](decisions/) | Append-only decision records. |
| [`research/`](research/) | Framework gate-tests + the genetics / control-theory lenses behind the design. |
| [`profiles/`](profiles/) | Per-instance expression profiles (`trellis-self` = instance #1). |
| [`AGENTS.md`](AGENTS.md) | The methodology we use to build Trellis (Layer B / instance #1). |

## How we work

**This repo has two jobs, kept separate by the install boundary (`decision-0035`):** it *produces*
Trellis — the invariants, catalog, payload generator, and plugin, in `core/` — **and it is itself a
Trellis-governed project**, installing Trellis through the official path (the same mechanical copy of
the pre-rendered payload any consumer gets) to govern its own work. So the invariants land in
`.trellis/rules.toml` and are delivered by the plugin's `SessionStart` hook — the same way any
consumer receives them (`decision-0071`). **On Claude Code.** This repo is *not* governed on Codex
CLI: Codex has no plugin installation channel yet (see above), and `decision-0071` D5 removed the
`AGENTS.md` bootstrap along with the overlay it read from. That gap is accepted and tracked in
Linear. `CLAUDE.md` is the Claude import adapter and `AGENTS.md`
holds the project's own *method* (the how). That's self-application, not self-reference — a compiler
built, then run on itself.

This repo used to carry a committed `.trellis/internal/` overlay with a CI guard keeping it
byte-identical to the payload. Both are gone: the overlay was the delivery mode `decision-0065`
retired for consumers, and with no second copy there is nothing left to drift. What that costs is
recorded in `decision-0071` — a marketplace plugin is the last released version, so this repo now
dogfoods shipped Trellis rather than the working tree.

Every non-code artifact carries frontmatter, and its lifecycle is the repository's own: **merging to
`main` is the acceptance** (`decision-0082`) — there is no status field to maintain. Decisions are
append-only, superseded by a forward pointer rather than edited; **intent is human-gated and
execution is independently verified** (the builder never grades itself); friction we hit becomes
product research rather than something to route around. See
[`AGENTS.md`](AGENTS.md).

## License

**[MIT](LICENSE)** — free and open (`decision-0019`). Read it, fork it, run it; that's the whole point
of Advisor mode. The Apache-2.0 upgrade path stays open should an enterprise / open-core future ever
make the patent + trademark grant worth it (cheap while single-owner). Any future monetization is
*services* (a managed supervisor, hosted conformance, compliance) — never a paywall on the invariants.
