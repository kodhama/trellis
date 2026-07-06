#!/usr/bin/env bash
# Trellis eval harness — one blind A/B run (research-0011).
# Scaffold a framework, run a dev task in two arms (baseline vs +Trellis), score each
# transcript with an independent reviewer against two rubrics. Pluggable via env.
#
#   FRAMEWORK=spec-kit TASK=eval/tasks/01-ambiguous-feature.md REPEATS=3 ./eval/run.sh
#
# Requires the framework installer (uv for spec-kit, npx for bmad) and a worker/reviewer
# agent on PATH. Fails loudly if a required tool is missing (invariant D1).
set -euo pipefail

FRAMEWORK="${FRAMEWORK:-spec-kit}"
TASK="${TASK:?set TASK=eval/tasks/NN-....md}"
REPEATS="${REPEATS:-1}"
# Worker edits files + runs commands (needs write perms); reviewer only reads + prints.
# For a fully-autonomous worker in a trusted sandbox, set WORKER_AGENT to use
# `--permission-mode bypassPermissions`. Run this yourself in an env where you've authorized
# headless agents — it deliberately runs unsupervised coding agents per arm.
WORKER_AGENT="${WORKER_AGENT:-claude -p --permission-mode acceptEdits}"
REVIEWER_AGENT="${REVIEWER_AGENT:-claude -p}"
OUTDIR="${OUTDIR:-runs}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
need() { command -v "$1" >/dev/null || { echo "FATAL: '$1' not on PATH — $2" >&2; exit 1; }; }

scaffold() {  # $1 = target dir. Non-interactive per research-0011.
  case "$FRAMEWORK" in
    spec-kit) need uv "install uv for Spec Kit"
      uvx --from git+https://github.com/github/spec-kit.git specify init "$1" \
        --integration claude --script sh --ignore-agent-tools ;;
    openspec) need npx "install node/npx for OpenSpec"
      mkdir -p "$1" && npx --yes @fission-ai/openspec@latest init "$1" --tools claude --force ;;
    cc-sdd) need npx "install node/npx for cc-sdd"
      mkdir -p "$1" && (cd "$1" && npx --yes cc-sdd@latest --claude-skills) ;;
    bmad) need npx "install node/npx for BMAD"
      mkdir -p "$1" && (cd "$1" && npx --yes bmad-method install --yes --tools claude-code --modules bmm) ;;
    spec-kit-lite)  # Spec Kit's rules as a plain AGENTS.md — no CLI, works with subagent workers
      mkdir -p "$1" && cp "$ROOT/eval/fixtures/spec-kit-lite.md" "$1/AGENTS.md" ;;
    # spec-swarm is intentionally absent — it installs only as an interactive Claude Code plugin
    # (research-0011), so it can't be scaffolded headlessly here.
    *) echo "FATAL: unknown FRAMEWORK '$FRAMEWORK' (spec-swarm is not scriptable — see research-0011)" >&2; exit 1 ;;
  esac
}

score() {  # $1 arm  $2 idx  $3 transcript  — one reviewer per rubric (extend to a panel via REPEATS)
  local arm="$1" i="$2" transcript="$3" base="$4"
  for rubric in invariants "${FRAMEWORK%-lite}"; do
    local card="$ROOT/eval/scorecards/$rubric.md"
    [ -f "$card" ] || continue
    local rp; rp="$(mktemp)"
    python3 "$ROOT/eval/fill.py" "$ROOT/eval/prompts/reviewer.md" \
      "TASK=$TASK" "TRANSCRIPT=$transcript" "RUBRIC=$card" > "$rp"
    $REVIEWER_AGENT "$(cat "$rp")" > "$base.$rubric.score.md" 2>&1 || true
    rm -f "$rp"
  done
}

run_arm() {  # $1 arm (baseline|trellis)  $2 idx
  local arm="$1" i="$2" dir; dir="$(mktemp -d)"
  scaffold "$dir"
  # seed the project-under-test this task needs (falls back to the shared base app)
  local fix="$ROOT/eval/fixtures/$(basename "$TASK" .md)"
  [ -d "$fix" ] || fix="$ROOT/eval/fixtures/sample-app"
  cp -R "$fix"/. "$dir"/
  # +Trellis arm only: apply the overlay, inlined into AGENTS.md so both subagent and claude -p
  # workers see the directives (an @import wouldn't resolve for a bare subagent worker).
  [ "$arm" = "trellis" ] && (cd "$ROOT/cli" && go run . setup --dir "$dir" \
      --profile a --mode m1 --target AGENTS.md --apply) >/dev/null
  local base="$OUTDIR/$FRAMEWORK/$(basename "$TASK" .md)/$arm-$i"
  mkdir -p "$(dirname "$base")"
  local wp; wp="$(mktemp)"
  python3 "$ROOT/eval/fill.py" "$ROOT/eval/prompts/worker.md" "TASK=$TASK" > "$wp"
  local transcript="$base.transcript.md"
  (cd "$dir" && $WORKER_AGENT "$(cat "$wp")") > "$transcript" 2>&1 || true
  rm -f "$wp"; rm -rf "$dir"
  score "$arm" "$i" "$transcript" "$base"
}

mkdir -p "$OUTDIR"
for i in $(seq 1 "$REPEATS"); do
  echo "[$FRAMEWORK] $(basename "$TASK") — repeat $i/$REPEATS" >&2
  run_arm baseline "$i"
  run_arm trellis "$i"
done
echo "done → $OUTDIR ; aggregate with: python3 eval/aggregate.py $OUTDIR" >&2
