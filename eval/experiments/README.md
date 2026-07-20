# eval/experiments/ — self-contained behavioral experiments

**One directory per experiment, self-contained** — runner, analysis, tasks, fixtures,
scoring, committed results — so an experiment is findable on `main`, re-runnable later,
and liftable as a reference by other repos, with no commit-history digging. The `eval/`
root keeps only the shared substrate (`fill.py`, `prompts/reviewer.md`) and this
directory.

This is a **convention, not a framework** (`inv-minimal-first`): there is no shared
experiment runner, and none should be added until at least two experiments genuinely
duplicate machinery.

## What an experiment directory carries

```
eval/experiments/<name>/
  README.md      — the experiment card: question, arms, how to run, status, results pointer
  run.sh         — the runner (arms, overlays, worker + reviewer invocation)
  aggregate.py   — analysis: rates/deltas, validity gates, significance
  task.md        — the harness-style task description (reviewer- and human-facing)
  scorecard.md   — the rubric the blind reviewer scores against
  fixture/       — the project-under-test, incl. brief.md (the ONLY thing the worker reads
                   about the task — trap descriptions stay reviewer-only)
  runs/          — results, committed after a run (transcripts, scores, meta), incl.
                   runs/provenance: one line per invocation — date, commit (+dirty),
                   payload stamp, repeats. Results are read against THAT commit, not
                   HEAD — an experiment may stop making sense in a future repo state,
                   and the provenance line is what keeps its numbers interpretable.
```

The single-task shape above is the default; a **suite-shaped** experiment carries pools
instead (`tasks/`, `fixtures/`, `scorecards/`, its own `prompts/`) and says so in its
card — `does-trellis-help` is the instance. Each experiment pairs with a `research-` note
in `research/` carrying the design, statistics, and decision rule — the directory is the
machinery, the note is the contract.

## The shared substrate (use it, don't fork it)

Experiments may call `eval/fill.py` and `eval/prompts/reviewer.md` (the blind-reviewer
idiom: no access to worker instructions, evidence-quoted verdicts, the
`<rule-id> | followed | violated | n-a | "quote"` grammar). They must **not** reach into
another experiment's directory — each suite's task/fixture pools and runner belong to it
alone, so one experiment's loop can never schedule another's task by construction.

## Experiments

| directory | question | research note | status |
|---|---|---|---|
| `does-trellis-help/` | does installing the Trellis overlay measurably improve invariant-following, per framework? | `research-0011` | task suite ready; full run pending (fix the worker-prompt leak first — see its card) |
| `annotation-vs-absence/` | can a `rules.toml` row deactivate a rule the model has *read*, or is assembling it out the only reliable off? | `research-0012` | **ran 2026-07-19 (batch 1, REPEATS=20): verdict AMEND** — zero leak, +95pt row effect; results in research-0012's appendix + `runs/`; implemented by decision-0053 (live rows) |
