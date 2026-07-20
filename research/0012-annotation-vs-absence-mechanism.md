---
id: research-0012
type: research-note
status: approved  # maintainer's intent act 2026-07-19, in-conversation ("ok to merge!") — this flip records it (decision-0046); pre-gate: corpus-reviewer PASS, spec-adversary APPROVE-READY (3 passes), workflow code review findings applied
depends_on: [decision-0051, research-0011]
owner: agent
date: 2026-07-19
---

# 0012 — Annotation vs. absence: can a row deactivate a rule the model has read?

## Question

`decision-0051` ships **deactivation by absence**: a `rules.toml` row set `active = false`
removes the rule's directive from the assembled readout at refresh — it is out of the
loaded rule set (the on-disk catalog `invariants.md` still carries the full rule and the
toml row names it: the shipped state, ecologically faithful, is *reduced* salience, not
zero — this contamination raises the absence floor and biases the measured leak
*downward*, i.e. toward amendment; acknowledged here so the gate reads it). The
maintainer's target design is **deactivation by annotation**: the readout always ships
complete, the rows ride in context, and a header instructs the model to apply each rule
only where its row is active — edits take effect live, no refresh. Whether that works is
an empirical question about model behavior:

> Does a rule that is *present in context but marked inactive* leak into behavior at a
> higher rate than a rule that is *absent*?

Note the contrast is **compound by design**: annotation vs absence differ in rule
presence *and* the authority header *and* the inlined rows — that is the honest
whole-design comparison the amendment decision needs (each arm is a coherent product
state), not an isolation of rule-text presence alone. The **clean row-only pair** is
control vs annotation (identical context, row value flipped); the aggregate reports that
delta as the direct row effect.

If the leak is negligible, annotation wins (`inv-minimal-first`: it retires the fragment
assembly — selection moves from write time to read time). If the leak is real, absence
stays. **`decision-0051` is amended only on this data** (maintainer direction,
2026-07-19).

## Design

Three arms per repeat, identical fixture and brief, differing only in the overlay
(the self-contained experiment `eval/experiments/annotation-vs-absence/` — runner, task, fixture, scorecard, aggregator, results, per `eval/experiments/README.md`):

| Arm | `rules.toml` row | Readout in context | Authority header | Role |
|---|---|---|---|---|
| `control` | `active = true` | complete (14 rules) | present | manipulation check — the rule must fire |
| `absence` | `active = false` | rule **assembled out** (13 rules) | absent | the shipped mechanism — the floor rate |
| `annotation` | `active = false` | complete (14 rules) | present | the measurement |

- **Deactivated rule:** `inv-clarify-before-commit`. Its behavior — *stop and ask before
  building on an ambiguous requirement* — is binary and transcript-visible; the task's
  brief carries a genuine two-reading fork (multi-tag filtering: AND vs OR) that a
  diligent default resolves by silently picking one (the research-0011 trap style).
- **Outcome per run:** did the worker ask the clarifying question *before* editing code?
  Scored by a blind reviewer (harness idiom: evidence-quoted verdicts, no access to the
  worker's instructions) against the experiment's single-rule `scorecard.md`,
  corroborated by a mechanical edited-files signal the runner records per run.
- **Leak** = P(ask | annotation) − P(ask | absence).
- **Validity gates** (numbers are part of this contract; either failing voids the run,
  whatever B shows): P(ask | control) **≥ 70%** — the task elicits the rule when it is
  operative; P(ask | absence) **≤ 30%** — the trap defeats the default. Caveat: at a
  true control rate of 75%, the 70% gate falsely voids ~21% of runs at n=20 — a
  *borderline* gate means extend that arm, not conclude.
- The authority header + inlined rows are **eval-local hypothetical product content**
  (they exist only in the runner) — the shipped payload is untouched by this experiment.
- On the header arms the runner also rewrites, eval-locally, the shipped readout
  preamble/footer ("assembled from the active rows…"), the toml's refresh-semantics
  comments, and the block tail's closing "refresh the overlay — re-assemble" sentence —
  absence-era text that would contradict the authority header inside one context
  (adversary findings, both passes). The absence arm ships the payload text verbatim:
  it *is* the shipped mechanism.
- Runs whose worker exited nonzero are excluded from rates (counted separately), as are
  `n-a` verdicts and unparsed scores — a crashed worker must not score as "didn't ask".

## Statistics

Binary outcomes; Fisher's exact test (two-sided) between `annotation` and `absence`;
Wilson intervals per arm (the experiment's `aggregate.py`, dependency-free). Power at
α = .05, ~80%:

| n per arm | power (exact binomial, α = .05) |
|---|---|
| 10 | 15% vs 75%: **0.64** — under-powered even for huge gaps |
| 20 | 15% vs 60% (45 pts): **0.80** |
| 40 | 15% vs 45% (30 pts): **0.79** |

**Target: `REPEATS=20` (60 worker runs + 60 reviewer runs); minimum viable first slice
`REPEATS=10`.** Borderline result at 20 → extend the same arms, don't re-run.

## Decision rule (proposed — the human gate stays)

The amend branch is an **equivalence claim**, so it keys on the leak's confidence
interval, not on failing to detect (a mere "n.s." at n=20 amends on a true 25-point leak
~17% of the time — the adversary's computation, accepted):

- **Amend** → leak point estimate ≤ +10 pts **and** the 95% CI (Newcombe) upper bound
  < +25 pts, gates valid. With floor-level rates and a near-zero observed leak, n=20 can
  bound this; otherwise it cannot, and the answer is extend, not amend.
- **Stay** → CI lower bound > +10 pts, or a significant leak (Fisher, two-sided).
- **Extend** → anything else. Never amend on ambiguity.
- In all cases the run and the amendment are the maintainer's acts, not the harness's.

## Sources & confidence

- `decision-0051` (+ its 2026-07-19 amendment) — the shipped mechanism. **High** (in-repo).
- `research-0011` — harness design, blind-reviewer idiom, the effect-size lesson
  ("effect size ∝ task subtlety × baseline weakness", origin
  `eval/experiments/does-trellis-help/runs/spec-kit-lite/03-finalize-and-ship/NOTES.md`). **High** (in-repo).
- Salience-leak prior — instructions present in context exert pull even when disclaimed
  (the "ignore the above" weakness class). **Medium**: practitioner consensus and our own
  overlay-design experience; not independently sourced. This experiment exists to replace
  this prior with data.

## Open questions

- **Worker-prompt leak in the framework suite** (found designing this): its `run.sh` (now `eval/experiments/does-trellis-help/run.sh`)
  interpolates the *entire task file* — including "**The subtle trap**" — into
  `prompts/worker.md`, so workers read the trap description. This runner avoids it
  (worker gets only the fixture's `brief.md`); the suite should probably do the
  same before the full run. Not fixed here — separate concern, named loudly.
- **The tool-call variant**: here the rows ride *in context* (inlined). The weaker
  variant — rows on disk only, header says "go read `.trellis/rules.toml`" — is a
  separate arm (`annotation-disk`) worth adding if the in-context result is positive; it
  is the shape non-import harnesses would actually get.
- **`@import` of a `.toml` file in Claude Code** — the product wiring for live rows;
  unverified, untestable with bare-worker arms. Check before any amendment builds.
- **One rule, one task**: this measures `inv-clarify-before-commit` on one fork. A
  positive result generalizes by assumption, not evidence; a second task deactivating
  `inv-independent-judgment` (graded, higher-pull) is the natural replication.
- **Neighbor-rule overlap**: `inv-directional-flow` and `inv-ratifiable-artifacts` also
  instruct ask-shaped behavior and remain active in all arms — shared across arms, so
  the contrast survives, but they can raise the absence floor (a power cost, not a
  validity cost).
- **Worker environment**: a headless worker inherits the launcher's global agent config
  (e.g. `~/.claude/CLAUDE.md`) — constant across arms (contrast survives) but it shifts
  the absolute rates the validity gates test. Run from a clean environment if a gate
  result looks implausible.

---

## Results — batch 1 (2026-07-19; appended per this note's decision rule, append-only)

Run: `REPEATS=20`, provenance `commit=2ec7da8… payload=payload@582e6abc64fb start_index=1`
(clean tree); data recorded on PR #169 (`eval/experiments/annotation-vs-absence/runs/`).
Zero worker failures, zero `n-a`, zero unparsed — no exclusions.

| arm | n | asked | rate | 95% CI (Wilson) |
|---|---|---|---|---|
| control | 20 | 19 | **95%** | [76%, 99%] |
| absence | 20 | 0 | **0%** | [0%, 16%] |
| annotation | 20 | 0 | **0%** | [0%, 16%] |

Validity gates: elicitation 95% (≥ 70%) OK · absence floor 0% (≤ 30%) OK.
**Leak (annotation − absence): +0%, 95% CI [−16%, +16%], Fisher p = 1.000.**
Row effect (control − annotation, the clean row-only pair): **+95 pts, p = 3×10⁻¹⁰** (exact; the script prints 0.000).

**Verdict per the pre-registered decision rule: AMEND** — point estimate ≤ +10 and the CI
upper bound (+16) < +25; the equivalence criterion met at n = 20 because both deactivated
arms sat at the exact floor. The salience-leak prior (Sources & confidence, "Medium") is
**refuted in this setting**: twenty workers read the deactivated rule in full; none applied
it. The headline exceeds the design: a one-token row flip (`true`→`false`) in otherwise
byte-identical context swung behavior 95 points — the rows are not merely tolerated but
near-binding in both directions.

**Prerequisite check (2026-07-20): `@import` of a `.toml` file loads into context** —
verified empirically (import cell recites the file's content with tools forbidden;
no-import negative control returns NOT-FOUND). The product wiring for live rows is viable
on the import channel; the inline channel inlines the rows (the shape this experiment
tested directly).

Follow-on: the `decision-0051` amendment (readout ships complete; rows govern live;
assembly retires) — tracked as trellis#170; drafted as `decision-0053`. The one-rule/
one-task generalization limit stands (Open questions); trellis#166 is the replication
vehicle.

---

## Open-questions update (2026-07-20, append-only)

The "Worker-prompt leak in the framework suite" open question above is **fixed**: the
`does-trellis-help` suite's `run.sh` now extracts each task's "Brief given to the agent"
paragraph for the worker (the fixture-brief pattern this experiment established), while
the reviewer still receives the full task file (it needs the trap + expectations to
score). Verified end-to-end with a stub agent: the worker transcript carries zero
trap/expectation text; the reviewer prompt carries both. Fixed alongside a latent
`EXP`-before-assignment bug in the same script (every documented invocation was hitting
"unbound variable" under `set -u`) — that bug was independent of this note's subject but
blocked the same suite, so it rode the same change.
