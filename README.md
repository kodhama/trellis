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
required (`decision-0070`). Every rule applies. `.trellis/rules.toml` is where a project changes
that, and a new one holds exactly these two lines:

```toml
# Every Trellis rule applies. To switch one off, add a row: <slug> = { active = false }
[rules]
```

- **To switch a rule off**, add its row under `[rules]`: `<slug> = { active = false }`. Only a row
  like that has any effect, and one is enough. **To switch the rule back on, delete that row**: a
  row that says `active = true` does not override it. A rule with no false row applies, and
  `strictness` and `seeded_from` do nothing. Floor rules cannot be switched off. A bad entry — a
  floor or an unknown slug set to `active = false`, a row that does not parse, a key the file does
  not define — is ignored with a warning in the session, and the rest of the file still counts.
- **To opt out**, the file holds `governed = false` on one top-level line, above any table. No rule
  applies then, floor rules included (`decision-0070` D5). Deleting that line turns governance back
  on with every rule applying, so confirm that is what you want.
- **An existing file keeps working as it is.** An older file usually carries `strictness`,
  `seeded_from` and an `active = true` row for every rule. Those lines now do nothing, and nothing
  asks you to remove them. Removing rows is safe only once every install that opens the repository
  runs plugin 0.24.0 or later: installed versions are pinned per checkout, and 0.6.0 and older
  refuse a file with missing rows.
- **A legacy delivery shape keeps every row.** A vendored `.trellis/internal/` overlay, an inline
  managed block, and a `.claude/rules/trellis.md` rendered by an older installer carry rules text
  that applies a rule only when its row says `active = true`. Those projects keep a full row set
  until they migrate.

The plugin vendors nothing. The rules themselves arrive at session start, injected by the plugin's own hook from the
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
launch in a project still applies — see the project-scope bullet below) — and seeds the two-line
`.trellis/rules.toml` shown above when none exists, so the project is governed at 16/16 the moment
the script exits (`decision-0070` D2). A project whose `.trellis/rules.toml` declares `governed = false`
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

Neither path needs a command to become governed — the curl path seeds the file and a
project-scoped plugin applies the shipped defaults without any file at all (`decision-0070`).
To switch individual rules off, add `active = false` rows as described at the top of this section,
and delete a rule's false row to switch it back on; the curl path never rewrites a file you already
have.

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
host blocks and the shared overlay. Also excluded is any other host-native transport.

**Any other harness — the manual copy path.** This is for harnesses the plugin does **not** cover.
On **Claude Code** it is superseded — use the marketplace or curl paths above
(`decision-0069`). **On Codex CLI it is not**: there is no way to install Trellis there yet
(`/plugin` commands are Claude Code's, and the curl path is Claude-only), so the manual copy is
what Codex has until a real channel exists. It is also **mutually exclusive with the curl path**: a repo that already
carries a hand-built overlay or a managed block makes `install.sh` refuse to render the rules
file, on purpose, because both would load and deliver the rules twice.

Every bundle file is pre-rendered plain text in
[`plugins/trellis/reference/`](plugins/trellis/reference) (the payload, `kodhama-0007`: one
render, many copiers). Start with the rules file:

```sh
git clone --depth 1 https://github.com/kodhama/trellis /tmp/trellis
ref=/tmp/trellis/plugins/trellis/reference
mkdir -p .trellis
# first install only — the file is yours after that
[ -e .trellis/rules.toml ] || cat > .trellis/rules.toml <<'EOF'
# Every Trellis rule applies. To switch one off, add a row: <slug> = { active = false }
[rules]
EOF
```

Then pick **exactly one** delivery branch — the two blocks below are alternatives, never both
(both at once is the overlay-plus-inline conflict the installer refuses and the hook alarms on).

**@import-capable file (CLAUDE.md)** — vendor the overlay the block imports:

```sh
mkdir -p .trellis/internal
cp "$ref"/invariants.md .trellis/internal/invariants.md
cp "$ref"/rules.md      .trellis/internal/rules.md   # the complete rules readout
cp "$ref"/trellis.md    .trellis/internal/trellis.md
cp "$ref"/version       .trellis/internal/version
cat "$ref"/block-claude.md >> CLAUDE.md
sed -n -e 's|  invariants\.md$|  .trellis/internal/invariants.md|p' \
       -e 's|  rules\.md$|  .trellis/internal/rules.md|p' \
       -e 's|  trellis\.md$|  .trellis/internal/trellis.md|p' \
       -e 's|  version$|  .trellis/internal/version|p' \
       "$ref"/checksums | shasum -a 256 -c -           # verify: all four lines print OK
```

**No @import support (e.g. AGENTS.md)** — append the SELF-CONTAINED inline block instead. It
embeds the rules and takes NO `.trellis/internal/` copies: the block plus
`.trellis/rules.toml` is the whole install, and copying the overlay beside it would deliver the
rules twice. The shipped block switches no rule off: if your `.trellis/rules.toml` already has an
`active = false` row, build the block with those rows as described below instead of appending
this one.

```sh
# the block must start at column 0 of its own line: guard against a file
# whose last line has no trailing newline before appending
[ -s AGENTS.md ] && [ -n "$(tail -c1 AGENTS.md)" ] && echo >> AGENTS.md
(cd "$ref" && grep '  block-inline\.md$' checksums | shasum -a 256 -c -)  # verify: OK
cat "$ref"/block-inline.md >> AGENTS.md
```

To switch a rule off later, add its row to `.trellis/rules.toml` as described under Get started
(`decision-0053`): the readout ships complete and opens with an authority note, and a row edit
takes effect at the next host context-loading boundary. The two `floor-*` rules apply whatever
their row says. On an **import** install the block loads your current `rules.toml` every session.

**On an inline install, the block is the only copy of your opt-outs the host reads.** The shipped
`block-inline.md` carries no rows, so a project that switches a rule off builds the block instead:
`cat "$ref"/block-inline-head.md "$ref"/rules.md`, then its `active = false` rows inside a ```toml
fence, then `cat "$ref"/block-inline-tail.md`. After every row edit, replace everything between
the `trellis:begin` and `trellis:end` markers with a block built the same way.

No binary or project runtime — the assets are plain files, and anything can verify them with
`shasum -c` against the shipped manifest. The optional local Codex native transport uses the
host's Node.js runtime as documented above. (The Homebrew/curl binary channel retired in
`kodhama-0007` rule 5;
the Go code in [`cli/`](cli/) survives as the release-time payload generator only.)

## The model

Deliberately tiny — small enough that a newcomer, human or agent, can read it and know how to make a
change that will pass.

- **Structural gate** — an admission check (one-way flow, handover points, a human
  intent locus, checkable artifacts). If a process lacks the shape, Trellis says so *loudly*.
- **Operating layer** — what Trellis supplies: a gate at every handover, independent
  verification (*the builder does not grade itself*), an auditable archive, bounded context,
  clarify-before-commit.
- **Dials** — per gate: *how strict* (`expressed` → `default-on-but-skippable` → `enforced`,
  the `dial-enforcement-strength` values from `core/invariants/trellis-invariants-v1.md:234`; an
  earlier wording here invented "documented → default-on", which matched no surface) and *who checks*
  (an agent, a human, or nobody). The same core serves a weekend hack and a regulated pipeline.
- **Floors** — the only settings that never dial to zero: every consequential choice is
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
  **independent conformance check** (`docs/rubrics/artifact-contract.md`, running on this repo),
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
| [`core/`](core/) | The shippable product: invariants, the signature catalog, the lexicon. |
| [`cli/`](cli/) | The **payload generator** (Go) — `trellis payload` renders the pre-built bundle + manifest at release; its tests are the CI sync-guards. Generator-only since `decision-0043`. |
| [`plugins/trellis/`](plugins/trellis/) | The **Claude Code and local Codex plugin** — `/trellis:remove`, host-isolated hooks, and the vendored payload (`reference/`). |
| [`install.sh`](install.sh) | The **curl path** — vends the whole plugin bundle onto disk as a skills-directory plugin, and on project scope renders `.claude/rules/trellis.md`, the file that actually delivers the rules (`decision-0068`). Claude Code only. |
| [`docs/decisions/`](docs/decisions/) | Append-only decision records, written only when a wrong call would be expensive to undo, slow to notice, or seriously damaging while it stands; later changes mark them with dated notes. |
| [`docs/research/`](docs/research/) | Framework gate-tests + the genetics / control-theory lenses behind the design. |
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
`main` is the acceptance** (`decision-0082`) — there is no status field to maintain. Decision records
are written only when a wrong call would be expensive to undo, slow to notice, or seriously damaging
while it stands, and are never rewritten: a later change marks what it outdates with a dated note, or
retires a record with a forward pointer; **intent is human-gated and
execution is independently verified** (the builder never grades itself); friction we hit becomes
product research rather than something to route around. See
[`AGENTS.md`](AGENTS.md).

### Local quality checks

Use Node.js 20.19 or newer and Go 1.22 or newer, install the pinned repository tools once, and
install [ShellCheck](https://www.shellcheck.net/) on `PATH` (`brew install shellcheck` on macOS).
Then run the same static checks and plugin smoke path as CI:

```sh
npm ci
npm run quality
```

`npm run quality` checks Go formatting, module tidiness, and vet; runs
[staticcheck](https://staticcheck.dev/) over the CLI (pinned via `go run`, so no install step);
formats and lints the plugin's JavaScript and JSON; runs ShellCheck over the maintained shell
entrypoints; enforces a 500 KB ceiling on tracked files; requires every TODO/FIXME in source to
name its tracker (`TODO(TRL-123)`, `TODO(#45)`, or `TODO(decision-0042)`); and starts both host
hooks against a temporary healthy project. It does not use live credentials, network services, or
your current `.trellis/` state. On a machine without Go, leave the full gate to CI; it runs the same
command.

The `cli-ci` workflow additionally enforces a test-coverage floor on the CLI: the suite runs with a
coverprofile and fails below 90% of statements (measured 92.3% when the floor was set). The floor is
a ratchet — raise it as coverage climbs, lower it only with a reason recorded in the PR.

A [devcontainer](.devcontainer/devcontainer.json) provisions the whole toolchain — Go 1.22, Node 20,
the pinned Node tools, and ShellCheck — for GitHub Codespaces or any devcontainer-capable editor.

To enable the checked-in pre-commit gate for this clone:

```sh
git config core.hooksPath .githooks
```

The hook runs `npm run quality`. It fails with the relevant installation command when the pinned Node
tools, Go, or ShellCheck are missing. Use Git's standard `--no-verify` escape hatch only when you intend
CI to be the first full check; CI runs the same command and cannot be bypassed by the local flag.

## License

**[MIT](LICENSE)** — free and open (`decision-0019`). Read it, fork it, run it; that's the whole point
of Advisor mode. The Apache-2.0 upgrade path stays open should an enterprise / open-core future ever
make the patent + trademark grant worth it (cheap while single-owner). Any future monetization is
*services* (a managed supervisor, hosted conformance, compliance) — never a paywall on the invariants.
