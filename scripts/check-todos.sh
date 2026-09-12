#!/bin/sh
# Debt markers in source must name the tracker that will re-present them
# (decision-0078's consumer rule, applied to code comments): a bare TODO is a
# note nobody agreed to. Prose is excluded by scanning source files only — the
# invariants catalog and eval fixtures quote TODOs as examples on purpose.
set -eu

# Self-excluded: this script's patterns and prose necessarily name the tokens.
hits=$(git ls-files '*.go' '*.mjs' '*.js' '*.cjs' '*.sh' '*.yml' '*.yaml' |
  grep -v '^eval/' |
  grep -v '^scripts/check-todos.sh$' |
  xargs grep -nE 'TODO|FIXME' 2>/dev/null |
  grep -vE '(TODO|FIXME)\((TRL-[0-9]+|#[0-9]+|decision-[0-9]+)\)' || true)

if [ -n "$hits" ]; then
  echo "TODO/FIXME without a tracker reference (write TODO(TRL-123), TODO(#45), or TODO(decision-0042)):" >&2
  echo "$hits" >&2
  exit 1
fi
