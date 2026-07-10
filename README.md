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

```sh
brew install kodhama/tap/trellis                                                    # Homebrew
# or, any Unix:  curl -fsSL https://raw.githubusercontent.com/kodhama/trellis/main/install.sh | sh
trellis setup
```

A single static binary — no package manager, no runtime. `trellis setup` rides your agent harness
(Claude Code today), asks a few things — a **posture** (`conductor` / `author-adapt` / `seed`), an
**install mode**, and, for a rewrite, a **model** — then, only with your go-ahead, composes Trellis
onto your project:

- **M1 · alongside** — a deterministic overlay: a one-line `@import` in your `CLAUDE.md` plus a
  `.trellis/` bundle (your profile + the invariant reference). Augment-never-clobber, idempotent.
- **M2 · morph** — a model-driven rewrite of your own instructions, on a fresh git branch to review.

Nothing is written without `--apply` (or your `y` at the prompt). Built from [`cli/`](cli/) (Go).

**Or install it as a Claude Code plugin** (no binary) — covers the M1 overlay natively:

```
/plugin marketplace add kodhama/kodhama
/plugin install trellis@kodhama
/trellis:setup
```

The CLI additionally offers the M2 morph and non-Claude harnesses; the plugin lives in
[`plugins/trellis`](plugins/trellis).

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
  agents **consult**; nothing of Trellis runs at agent-time. This is what `trellis setup` installs
  today (M1 overlay or M2 morph). Nothing to secure or remove at runtime.
- **Supervisor** *(installed, live — in progress)* — Trellis wired into your pipeline: gates fire on
  commit/PR events via hooks, it stays current through an update channel, and it comes off cleanly.
  The next delivery slice.

These are the two ends of the delivery relationship; the cross-lens vocabulary lives in
[`core/lexicon.md`](core/lexicon.md).

## Where it stands

Built in the open, dogfooded on itself from commit one. The honest state:

- **Ratified** — the invariant set (`invariants-v1`), 24 decisions, 8 research notes.
- **Shipped** — the **v0 setup CLI** (`trellis setup`, released `v0.1.0`): harness detection, the
  onboarding flow, and both install modes (M1 deterministic overlay, M2 model-driven morph). It stands
  on the *spine* + an **independent conformance check** (`spec-0001`, running on this repo), the
  expression-profile + catalog **schema** (`spec-0002`), the machinery design (`spec-0003`), the
  populated catalog and the first per-project **profile** (instance #1), and the cross-lens **lexicon**.
- **In progress** — **supervisor mode** (installed live gates) and a Claude-marketplace install route.
- **The open risk** — the invariants are validated on essentially *one* project. **Instance #2** — a
  second, different project — is the next real test of whether they generalize.

## Repo map

| Path | What |
|---|---|
| [`agentic-dev-meta-layer-brief.md`](agentic-dev-meta-layer-brief.md) | The full thesis (start at §10 verdict, §11 start-here, §12 operating method). |
| [`core/`](core/) | The shippable product: invariants, the conformance rubric, the signature catalog, the lexicon. |
| [`cli/`](cli/) | The **setup CLI** (Go) — `trellis setup`: detect the harness, onboard, and apply (M1/M2). |
| [`specs/`](specs/) | The spine (`0001`), the profile / catalog schema (`0002`), the delivery machinery (`0003`). |
| [`decisions/`](decisions/) | Append-only decision records (`0001–0024`). |
| [`research/`](research/) | Framework gate-tests + the genetics / control-theory lenses behind the design. |
| [`profiles/`](profiles/) | Per-instance expression profiles (`trellis-self` = instance #1). |
| [`CLAUDE.md`](CLAUDE.md) | The methodology we use to build Trellis (Layer B / instance #1). |

## How we work

**This repo has two jobs, kept separate by the install boundary (`decision-0035`):** it *produces*
Trellis — the invariants, catalog, CLI, and plugin, in `core/` — **and it is itself a Trellis-governed
project**, installing Trellis through the official path (`trellis setup`) to govern its own work. So the
invariants land in `.trellis/` via the same overlay any user gets (not hand-copied), and `CLAUDE.md`
holds only the project's own *method* (the how). That's self-application, not self-reference — a compiler
built, then run on itself. A CI guard keeps the committed overlay identical to what the product produces,
so it can't drift.

Every non-code artifact carries frontmatter and a lifecycle (`draft → ratified`); decisions are
append-only; **intent is human-gated and execution is independently verified** (the builder never grades
itself); friction we hit becomes product research rather than something to route around. See
[`CLAUDE.md`](CLAUDE.md).

## License

**[MIT](LICENSE)** — free and open (`decision-0019`). Read it, fork it, run it; that's the whole point
of Advisor mode. The Apache-2.0 upgrade path stays open should an enterprise / open-core future ever
make the patent + trademark grant worth it (cheap while single-owner). Any future monetization is
*services* (a managed supervisor, hosted conformance, compliance) — never a paywall on the invariants.
