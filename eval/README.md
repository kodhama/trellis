# Trellis eval harness

Does laying Trellis over a project measurably change how an agent behaves? This harness answers that
**empirically**, per `research-0011`: a **blind A/B** where the only difference is the Trellis overlay,
scored by an **independent reviewer** against two rubrics.

```
scaffold framework X ─┬─ arm A (baseline)      ─┐
                      └─ arm B (+ trellis setup) ┘
        run a dev task with a worker agent  →  transcript per arm
        independent reviewer scores each transcript (blind to arm) against:
          · eval/scorecards/<framework>.md   — the framework's own rules
          · eval/scorecards/invariants.md     — the invariants (✗-failure rubric)
        Δ = B − A on each scorecard, over repeats × tasks
```

**Win condition:** +Trellis lifts the invariant score **without hurting** framework-adherence.

## Layout

| path | what |
|---|---|
| `tasks/*.md` | dev tasks chosen to create invariant-testing moments |
| `scorecards/invariants.md` | the invariant rubric — **auto-derived** from the catalog (`gen-invariant-scorecard.py`) |
| `scorecards/<framework>.md` | the framework's own declared rules |
| `prompts/worker.md` | the worker prompt (do the task, following the project's instructions) |
| `prompts/reviewer.md` | the reviewer prompt (score a transcript against a rubric; blind to arm) |
| `prompts/diff.md` | the behavioral-diff prompt (compare the two arms: *did behavior differ, and how?*) |
| `run.sh` | the orchestrator (scaffold → apply → worker → reviewer → record) |
| `aggregate.py` | roll per-run scores into the Δ |

Two outputs per run: the **Δ** (`aggregate.py` — *whether* behavior differed, in counts) and the
**behavioral diff** (`prompts/diff.md` — *how* it differed, in evidence). See `runs/…/RESULT.md` +
`BEHAVIORAL-DIFF.md` for a worked example.

**Frameworks** (`FRAMEWORK=`): `spec-kit` (scaffold verified), `openspec`, `cc-sdd`, `bmad`. All install
non-interactively. `spec-swarm` is **not** supported — it installs only as an interactive Claude Code
plugin (`research-0011`).

## Run

Pluggable via env; defaults to Spec Kit + headless Claude. Needs the framework's installer on PATH
(`uv` for Spec Kit, `npx` for BMAD) and a worker/reviewer agent (`claude`).

```sh
# one (framework, task) cell:
FRAMEWORK=spec-kit TASK=eval/tasks/01-ambiguous-feature.md REPEATS=3 ./eval/run.sh

# the whole suite, one framework:
for t in eval/tasks/*.md; do FRAMEWORK=spec-kit TASK="$t" REPEATS=3 ./eval/run.sh; done
python3 eval/aggregate.py runs/           # Δ across the recorded runs
```

Each run **auto-seeds the task's project-under-test** from `fixtures/` (see `fixtures/README.md`); both
arms get the same fixture, so the only difference is the overlay. Use `FRAMEWORK=spec-kit-lite` for the
no-CLI path (Spec Kit's rules as a plain `AGENTS.md`) — the fidelity ceiling for a bare subagent worker;
full skill-driven fidelity needs a real scaffold + a `claude -p` worker.

## Honest limits (see `research-0011` open questions)

- **Blinding is imperfect** — the overlay is in the worker's instructions; we blind the *reviewer* to
  instructions (it scores behavior), not the world.
- **Stochastic** — one run proves nothing; use `REPEATS` + the judge panel, and trust the **Δ** more than
  any absolute score.
- **Worker = headless Claude** by default; a different agent shifts the baseline (the Δ should be steadier).
- First runs are a **proof-of-concept**, not a powered study.
