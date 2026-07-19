#!/usr/bin/env bash
# Sync-guard: eval/experiments/does-trellis-help/scorecards/invariants.md is a derived resource of the signature catalog
# (decision-0028). Regenerate and fail loudly if the committed copy is stale.
set -euo pipefail
cd "$(dirname "$0")/../../.."
python3 eval/experiments/does-trellis-help/gen-invariant-scorecard.py
if ! git diff --quiet -- eval/experiments/does-trellis-help/scorecards/invariants.md; then
  echo "FATAL: eval/experiments/does-trellis-help/scorecards/invariants.md is stale vs the catalog." >&2
  echo "       re-run: python3 eval/experiments/does-trellis-help/gen-invariant-scorecard.py" >&2
  git --no-pager diff -- eval/experiments/does-trellis-help/scorecards/invariants.md >&2
  exit 1
fi
echo "eval invariant scorecard is in sync with the catalog"
