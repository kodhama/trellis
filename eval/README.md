# eval/ — behavioral evidence for Trellis

Everything here measures *behavior* — does the overlay, or a mechanism inside it,
actually change what agents do? All experiments live under **`experiments/`**, one
self-contained directory each (convention + index: `experiments/README.md`):

- **`experiments/does-trellis-help/`** — the framework A/B suite (`research-0011`):
  baseline vs +overlay across methodology frameworks, 11 trap-tasks.
- **`experiments/annotation-vs-absence/`** — the deactivation-mechanism experiment
  (`research-0012`): can a `rules.toml` row switch off a rule the model has read?

The `eval/` root keeps only the **shared substrate**: `fill.py` (prompt-template filler)
and `prompts/reviewer.md` (the blind-reviewer idiom — the reviewer never sees the
worker's instructions and must quote evidence for every verdict). Runs are launched by a
human, per experiment, from each experiment's own `run.sh` — see its README.
