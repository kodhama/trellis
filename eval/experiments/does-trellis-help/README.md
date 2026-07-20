# does-trellis-help — the framework A/B suite (research-0011)

**Question** (`research-0011`, the design contract): does installing the Trellis overlay
measurably improve invariant-following, per methodology framework? Two arms per run —
`baseline` (framework scaffold alone) vs `trellis` (same scaffold + the overlay, inlined
into `AGENTS.md`) — over a pool of 11 tasks designed to trap *subtle* invariant moments
(effect size ∝ task subtlety × baseline weakness; the confounded task-03 run in
`runs/spec-kit-lite/03-finalize-and-ship/NOTES.md` is the lesson's origin).

**Suite-shaped, not single-task**: unlike the default experiment layout, this one carries
pools — `tasks/` (11), `fixtures/` (per-task + `sample-app` fallback), `scorecards/`
(per-framework cards + `invariants.md`, **generated** from the signature catalog by
`gen-invariant-scorecard.py`; `check-scorecard.sh` is the CI sync-guard,
`.github/workflows/eval-scorecard.yml`), and its own `prompts/` (`worker.md`,
`diff.md` — the shared blind-reviewer prompt stays at `eval/prompts/reviewer.md`).

**Run** (yours to launch — spawns unsupervised headless workers):

```sh
# one (framework, task) cell:
FRAMEWORK=spec-kit TASK=eval/experiments/does-trellis-help/tasks/01-ambiguous-feature.md \
  REPEATS=3 ./eval/experiments/does-trellis-help/run.sh

# the whole suite, one framework:
for t in eval/experiments/does-trellis-help/tasks/*.md; do
  FRAMEWORK=spec-kit TASK="$t" REPEATS=3 ./eval/experiments/does-trellis-help/run.sh
done
python3 eval/experiments/does-trellis-help/aggregate.py eval/experiments/does-trellis-help/runs
```

**Status: task suite ready (11 tasks, fixtures verified); first spec-kit-lite slices in
`runs/`; the full run is pending.** The worker-prompt leak is **fixed** (2026-07-20):
`run.sh` now extracts just each task's "Brief given to the agent" paragraph for the
worker — the trap description and "what a strong run does" reach only the reviewer,
mirroring the `annotation-vs-absence` fixture-brief pattern (verified end-to-end with a
stub agent: worker transcript carries zero trap/expectation text, reviewer prompt
carries both). A latent `EXP`-before-assignment bug in the same file (the default
`OUTDIR` referenced `$EXP` one line before it was set — every documented invocation
would have hit "unbound variable" under `set -u`) was fixed in the same change. Migrated
here from the `eval/` root (was the original harness) for one-convention consistency.
