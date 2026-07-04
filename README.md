# Trellis

> A trellis is structure that *enables* growth rather than dictating form.

**Trellis is the governing layer for agentic software development.** It sits *above* whatever
methodology your project already uses — Spec Kit, BMAD, your own — learns its shape, and enforces a
small set of **invariants**: the handful of properties that must stay true no matter how fast
autonomous agents move. It doesn't replace your process; it keeps it honest.

It exists for one failure mode: **fast agents drift.** They skip reviews, resolve ambiguity by
guessing, grade their own work, and leave no trace of where a decision came from. Trellis makes a few
things non-negotiable — and surfaces every time something bends — without dictating how you build.

## The model

Deliberately tiny — small enough that a newcomer, human or agent, can read it and know how to make a
change that will pass.

- **A · Structural gate** — a four-point admission check (one-way flow, handover points, a human
  intent locus, checkable artifacts). If a process lacks the shape, Trellis says so *loudly*.
- **B · Operating layer** — what Trellis supplies: a gate at every handover, independent
  verification (*the builder does not grade itself*), an auditable archive, bounded context,
  clarify-before-commit.
- **C · Two dials** — per gate: *how strict* (documented → default-on → enforced) and *who checks*
  (an agent, a human, or nobody). The same core serves a weekend hack and a regulated pipeline.
- **D · Two floors** — the only settings that never dial to zero: every consequential choice is
  **surfaced**, and the **human intent gate never fully opens**.

The full set: [`core/invariants/trellis-invariants-v1.md`](core/invariants/trellis-invariants-v1.md).
The thesis behind it: [`agentic-dev-meta-layer-brief.md`](agentic-dev-meta-layer-brief.md).

## Two ways to run it

- **Advisor** *(open, no runtime)* — Trellis is never installed; your agents **consult** this repo as
  a reference and internalize what applies. Today that means pointing your coding agent at the repo.
  The intended interface is a small **CLI your host invokes**, so the boundary is crisp rather than
  "read whatever you find" *(in progress)*. Nothing to secure or remove at runtime.
- **Supervisor** *(installed, live)* — Trellis is wired into your pipeline: gates fire on commit/PR
  events via hooks, it stays current through an update channel, and it comes off cleanly. *In
  progress — this is the machinery the delivery slice is building.*

These are the two ends of the delivery relationship; the cross-lens vocabulary lives in
[`core/lexicon.md`](core/lexicon.md).

## Where it stands

Built in the open, dogfooded on itself from commit one. The honest state:

- **Ratified** — the invariant set (`invariants-v1`), 17 decisions, 8 research notes.
- **Built** — the *spine*: the artifact contract, lifecycle, and an **independent conformance check**
  (`spec-0001`, running on this repo); the expression-profile + signature-catalog **schema**
  (`spec-0002`); the populated catalog and the first per-project **profile** (dogfooded as instance
  #1); the cross-lens **lexicon**.
- **In progress** — the **delivery machinery**: the Assess → Apply flow and the installed supervisor /
  CLI. This is what turns the design into a running product.
- **The open risk** — the invariants are validated on essentially *one* project. **Instance #2** — a
  second, different project — is the next real test of whether they generalize.

## Repo map

| Path | What |
|---|---|
| [`agentic-dev-meta-layer-brief.md`](agentic-dev-meta-layer-brief.md) | The full thesis (start at §10 verdict, §11 start-here, §12 operating method). |
| [`core/`](core/) | The shippable product: invariants, the conformance rubric, the signature catalog, the lexicon. |
| [`specs/`](specs/) | The spine (`0001`) and the expression-profile / catalog schema (`0002`). |
| [`decisions/`](decisions/) | Append-only decision records (`0001–0017`). |
| [`research/`](research/) | Framework gate-tests + the genetics / control-theory lenses behind the design. |
| [`profiles/`](profiles/) | Per-instance expression profiles (`trellis-self` = instance #1). |
| [`CLAUDE.md`](CLAUDE.md) | The methodology we use to build Trellis (Layer B / instance #1). |

## How we work

We build Trellis with Trellis. Every non-code artifact carries frontmatter and a lifecycle
(`draft → ratified`); decisions are append-only; **intent is human-gated and execution is
independently verified** (the builder never grades itself); friction we hit becomes product research
rather than something to route around. See [`CLAUDE.md`](CLAUDE.md).

## License

**[MIT](LICENSE)** — free and open (`decision-0019`). Read it, fork it, run it; that's the whole point
of Advisor mode. The Apache-2.0 upgrade path stays open should an enterprise / open-core future ever
make the patent + trademark grant worth it (cheap while single-owner). Any future monetization is
*services* (a managed supervisor, hosted conformance, compliance) — never a paywall on the invariants.
