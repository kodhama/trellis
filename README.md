# Trellis

> A trellis is structure that *enables* growth rather than dictating form.

**Trellis is the governing layer for agentic software development.** It sits *above* whatever
methodology your project already uses — Spec Kit, BMAD, your own — learns its shape, and enforces a
small set of **invariants**: the handful of properties that must stay true no matter how fast
autonomous agents move. It doesn't replace your process; it keeps it honest.

It exists for one failure mode: **fast agents drift.** They skip reviews, resolve ambiguity by
guessing, grade their own work, and leave no trace of where a decision came from. Trellis makes a few
things non-negotiable — and surfaces every time something bends — without dictating how you build.

## Get started

**Claude Code (the primary path)** — install the plugin from the kodhama family marketplace, then
run the setup skill in any project:

```
/plugin marketplace add kodhama/kodhama
/plugin install trellis@kodhama
/trellis:setup
```

`/trellis:setup` asks one thing — a **posture** (`conductor` / `author-adapt`) — or reads it from
`.trellis/expression.md` if the project already declares one, then composes Trellis onto your
project as the **M1 "alongside" overlay**: a one-line `@import` in your `CLAUDE.md` plus a
`.trellis/` bundle (your profile + the invariant reference + your hand-owned `expression.md`).
Augment-never-clobber, idempotent, verified against a shipped checksum manifest. On explicit
request it also runs the **M2 morph** — a model-driven rewrite of your own instructions, on a
fresh git branch you review. The plugin lives in [`plugins/trellis`](plugins/trellis).

**Same plugin, without the marketplace — the curl path
([#124](https://github.com/kodhama/trellis/issues/124)).** `install.sh` vends the whole
`plugins/trellis/` tree onto disk as a [skills-directory
plugin](https://code.claude.com/docs/en/plugins-reference#skills-directory-plugins): any folder
under a skills directory with its own `.claude-plugin/plugin.json` loads as `trellis@skills-dir` on
Claude Code's next session, no marketplace and no install step. This script makes exactly **one**
decision — where to put the plugin — and composes nothing else; every other decision (posture,
which file to patch, and so on) is still made entirely by `/trellis:setup`, unmodified, once the
plugin is on disk:

```sh
curl -fsSL https://raw.githubusercontent.com/kodhama/trellis/main/install.sh | sh
```

Two scopes, `--scope project` (default, inside a git repo) or `--scope personal`:

- **project** — `<repo-root>/.claude/skills/trellis/`, checked into git so it reaches every
  collaborator on clone. Resolved via `git rev-parse --show-toplevel` from wherever you run it, not
  `$PWD` — project-scope skills-directory plugins do **not** walk up to the repo root, so launch
  Claude Code from that root (or run `/reload-plugins` after `cd`'ing there), and expect its
  workspace **trust dialog** on first launch there (unavoidable — the content came from the repo,
  not from you). The script never runs `git add`/`git commit`; it prints the command and leaves the
  commit to you.
- **personal** — `~/.claude/skills/trellis/`, available in every project on the machine, no trust
  dialog, no repo required.

Every fetched byte is verified against a manifest baked into the script *before* anything is
written — a mismatch fails loudly and installs nothing. Inspect first, or pass flags, instead of
piping straight to `sh`:

```sh
curl -fsSLO https://raw.githubusercontent.com/kodhama/trellis/main/install.sh
less install.sh && sh install.sh --scope personal
```

Then run `/trellis:setup` as above — that skill is the one real interactive writer either path
leads to.

**Any other harness — the manual copy path.** Every bundle file is pre-rendered plain text in
[`plugins/trellis/reference/`](plugins/trellis/reference) (the payload, `kodhama-0007`: one
render, many copiers). Pick a posture key (`a` = conductor, `b` = author-adapt) and copy:

```sh
git clone --depth 1 https://github.com/kodhama/trellis /tmp/trellis
ref=/tmp/trellis/plugins/trellis/reference   # <p> below: a | b
mkdir -p .trellis
cp "$ref"/invariants.md    .trellis/invariants.md
cp "$ref"/profile-<p>.md   .trellis/profile.md
cp "$ref"/trellis-<p>.md   .trellis/trellis.md
cp "$ref"/version          .trellis/version
cp "$ref"/expression-<p>.md .trellis/expression.md   # first install only — hand-owned after that
cat "$ref"/block-claude.md >> CLAUDE.md              # @import-capable files
# no @import support (e.g. AGENTS.md)?  append block-inline-<p>.md instead
sed -n -e 's|  invariants\.md$|  .trellis/invariants.md|p' \
       -e 's|  profile-<p>\.md$|  .trellis/profile.md|p' \
       -e 's|  trellis-<p>\.md$|  .trellis/trellis.md|p' \
       -e 's|  version$|  .trellis/version|p' \
       "$ref"/checksums | shasum -a 256 -c -           # verify: all four lines print OK
```

No binary, no runtime — the assets are plain files, and anything can verify them with `shasum -c`
against the shipped manifest. (The Homebrew/curl binary channel retired in `kodhama-0007` rule 5;
the Go code in [`cli/`](cli/) survives as the release-time payload generator only.)

## The model

Deliberately tiny — small enough that a newcomer, human or agent, can read it and know how to make a
change that will pass.

- **Structural gate** — a four-point admission check (one-way flow, handover points, a human
  intent locus, checkable artifacts). If a process lacks the shape, Trellis says so *loudly*.
- **Operating layer** — what Trellis supplies: a gate at every handover, independent
  verification (*the builder does not grade itself*), an auditable archive, bounded context,
  clarify-before-commit.
- **Two dials** — per gate: *how strict* (documented → default-on → enforced) and *who checks*
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
  agents **consult**; nothing of Trellis runs at agent-time. This is what `/trellis:setup` (or the
  manual copy path) installs today: the M1 overlay, plus the M2 morph on request. Nothing to secure
  or remove at runtime.
- **Supervisor** *(installed, live — in progress)* — Trellis wired into your pipeline: gates fire on
  commit/PR events via hooks, it stays current through an update channel, and it comes off cleanly.
  The next delivery slice.

These are the two ends of the delivery relationship; the cross-lens vocabulary lives in
[`core/lexicon.md`](core/lexicon.md).

## Where it stands

Built in the open, dogfooded on itself from commit one. The honest state:

- **Ratified** — the invariant set (`invariants-v1`), 40+ decisions, 8+ research notes.
- **Shipped** — the **Claude Code plugin** (marketplace install, `/trellis:setup` /
  `/trellis:remove`, a bundled staleness hook) riding a **pre-rendered, checksum-manifested
  payload** (`kodhama-0007`: render once at release, writers only copy and verify), plus the
  documented **manual copy path** for any other harness. It stands on the *spine* + an
  **independent conformance check** (`spec-0001`, running on this repo), the expression-profile +
  catalog **schema** (`spec-0002`), the machinery design (`spec-0003`), the populated catalog and
  the first per-project **profile** (instance #1), and the cross-lens **lexicon**. The v0 **setup
  CLI** shipped first (`v0.1.0`–`v0.2.29`) and its end-user channel retired in favor of the above
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
| [`cli/`](cli/) | The **payload generator** (Go) — `trellis payload` renders the pre-built bundle + manifest at release; its tests are the CI sync-guards. Generator-only since `decision-0043` (#120). |
| [`plugins/trellis/`](plugins/trellis/) | The **Claude Code plugin** — `/trellis:setup`, `/trellis:remove`, the staleness hook, and the vendored payload (`reference/`). |
| [`install.sh`](install.sh) | The **curl path** (`#124`) — vends the whole plugin bundle onto disk as a skills-directory plugin; makes exactly one decision (scope) and composes nothing else. |
| [`specs/`](specs/) | The spine (`0001`), the profile / catalog schema (`0002`), the delivery machinery (`0003`). |
| [`decisions/`](decisions/) | Append-only decision records. |
| [`research/`](research/) | Framework gate-tests + the genetics / control-theory lenses behind the design. |
| [`profiles/`](profiles/) | Per-instance expression profiles (`trellis-self` = instance #1). |
| [`CLAUDE.md`](CLAUDE.md) | The methodology we use to build Trellis (Layer B / instance #1). |

## How we work

**This repo has two jobs, kept separate by the install boundary (`decision-0035`):** it *produces*
Trellis — the invariants, catalog, payload generator, and plugin, in `core/` — **and it is itself a
Trellis-governed project**, installing Trellis through the official path (the same mechanical copy of
the pre-rendered payload any consumer gets) to govern its own work. So the invariants land in
`.trellis/` via the same overlay any user gets (not hand-composed), and `CLAUDE.md` holds only the
project's own *method* (the how). That's self-application, not self-reference — a compiler built, then
run on itself. A CI guard keeps the committed overlay identical to what the product produces, so it
can't drift.

Every non-code artifact carries frontmatter and a lifecycle (`draft → ratified`); decisions are
append-only; **intent is human-gated and execution is independently verified** (the builder never grades
itself); friction we hit becomes product research rather than something to route around. See
[`CLAUDE.md`](CLAUDE.md).

## License

**[MIT](LICENSE)** — free and open (`decision-0019`). Read it, fork it, run it; that's the whole point
of Advisor mode. The Apache-2.0 upgrade path stays open should an enterprise / open-core future ever
make the patent + trademark grant worth it (cheap while single-owner). Any future monetization is
*services* (a managed supervisor, hosted conformance, compliance) — never a paywall on the invariants.
