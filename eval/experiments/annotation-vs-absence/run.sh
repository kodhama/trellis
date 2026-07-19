#!/usr/bin/env bash
# Mechanism runner — annotation vs absence (research-0012).
#
# Three arms per repeat, same fixture and brief, differing ONLY in the trellis overlay:
#   control    — rule row active=true,  full readout, authority header  (manipulation check)
#   absence    — rule row active=false, rule assembled OUT of the readout (shipped mechanism)
#   annotation — rule row active=false, full readout + authority header   (the measurement)
#
# Deliberately separate from the harness's run.sh: that runner's two arms
# (baseline|trellis) and its aggregate are hardwired for the framework A/B
# (research-0011); this experiment varies the OVERLAY, not the framework, and adds no
# scaffold (a framework would only add noise to a single-moment task). The worker
# receives ONLY the fixture's brief.md — never the task file — and the blind reviewer
# receives task.md WITHOUT its "What a strong run does" section (it names per-arm
# expectations; stripping it preserves blinding). Layout per eval/experiments/README.md.
#
#   REPEATS=20 ./eval/experiments/annotation-vs-absence/run.sh
#
# EXTEND semantics (the decision rule's "borderline → extend" branch): a new invocation
# continues numbering after the highest existing index in $OUTDIR — results accumulate,
# nothing is overwritten, and provenance appends one line per invocation.
#
# On the HEADER ARMS (control, annotation) the readout preamble/footer, the toml's
# refresh-semantics comments, AND the block tail's closing refresh sentence are rewritten
# eval-locally: the shipped payload text says "assembled from the active rows" / "no
# effect until refresh" / "refresh the overlay — re-assemble it", which is absence-era
# truth and would contradict the live-rows authority header inside the same context
# (adversary finding 3; code-review finding 5). The absence arm ships verbatim — it IS
# the shipped mechanism. All such rewrites are EVAL-LOCAL hypothetical product content;
# the payload itself is untouched.
set -euo pipefail

EXP="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$EXP/../../.." && pwd)"
TASK="${TASK:-$EXP/task.md}"
REPEATS="${REPEATS:-1}"
RULE_SLUG="${RULE_SLUG:-inv-clarify-before-commit}"
WORKER_AGENT="${WORKER_AGENT:-claude -p --permission-mode acceptEdits}"
REVIEWER_AGENT="${REVIEWER_AGENT:-claude -p}"
OUTDIR="${OUTDIR:-$EXP/runs}"
REF="$ROOT/plugins/trellis/reference"
FIX="$EXP/fixture"
[ -d "$FIX" ] || { echo "FATAL: fixture dir $FIX missing" >&2; exit 1; }
[ -f "$FIX/brief.md" ] || { echo "FATAL: $FIX/brief.md missing — the worker brief is fixture-local" >&2; exit 1; }
grep -q "\`$RULE_SLUG\`" "$REF/rules.md" || { echo "FATAL: slug $RULE_SLUG not tagged in payload rules.md" >&2; exit 1; }

# The authority header (eval-local; the live-rows mechanism under test).
AUTHORITY_HEADER='**Rule activation is governed by `.trellis/rules.toml` (its rows are inlined below the rules):** apply each rule below ONLY if its row says `active = true`. A rule whose row is `active = false` does not apply in this project — do not follow it. The two `floor-` rows apply regardless of their row value.'

# Readout with the rule removed — mechanical subset keyed on the slug tag decision-0051
# put on each rule's first line (the bullet line carrying `<slug>` plus its ✗ line).
subset_readout() {  # stdout = rules.md minus RULE_SLUG's two lines
  awk -v tag="\`$RULE_SLUG\`" '
    skip_next { skip_next=0; next }
    index($0, tag) { skip_next=1; next }
    { print }' "$REF/rules.md"
}

# Header-arm readout: full list, with the absence-era assembly preamble replaced and the
# assembly footer dropped (they would contradict the authority header — see top comment).
header_arm_readout() {
  awk '
    /^This list is assembled from the active rows/ {
      print "Each rule below ends with its row'\''s slug. Whether a rule applies is governed by its row in `.trellis/rules.toml` (see the authority note above; the rows are inlined below the rules). Each is a rule to follow, then the ✗ failure it prevents:"; next }
    /^\(Generated from your `rules\.toml`/ { next }
    { print }' "$REF/rules.md"
}

# Header-arm tail: the shipped block tail's closing sentence says row edits need a
# refresh/re-assembly — the same absence-era truth, and it is the LAST thing the model
# reads. Replace that sentence with live-rows wording; keep the invariants pointer.
header_arm_tail() {
  awk '
    /^If a rule seems ambiguous/ {
      print "If a rule seems ambiguous, or in tension with this project'\''s own instructions, read its entry in `.trellis/internal/invariants.md` — the description and with/without examples — before deviating. Rule activation follows the rows in `.trellis/rules.toml` directly (see the authority note above)."; next }
    { print }' "$REF/block-inline-tail.md"
}

# Header-arm toml: the shipped seed's top comment says edits have no effect until a
# refresh — absence-era truth, contradicts the authority header. Replace it (and the
# floor rows' assembly-speak trailing comment) with live-rows wording.
header_arm_toml() {
  awk '
    NR==1,/^$/ { if (/^#|^$/) { if (!printed) { print "# Rows govern rule activation live (see the authority note in the project instructions)."; print ""; printed=1 }; next } }
    { gsub(/# floor-held — assembly includes it even if set false, and says so loudly/,
           "# floor — applies regardless of this row")
      print }' "$REF/rules-a.toml"
}

# Flip the rule's row in a rules.toml copy.  $1 = file, $2 = true|false
set_row() {
  sed -i.bak "s/^$RULE_SLUG *=.*/$RULE_SLUG = { active = $2 }/" "$1" && rm -f "$1.bak"
  grep -q "^$RULE_SLUG = { active = $2 }" "$1" || { echo "FATAL: row flip failed in $1" >&2; exit 1; }
}

# Build the arm's AGENTS.md block + .trellis/ state in $1 (the workdir).  $2 = arm.
overlay() {
  local dir="$1" arm="$2"
  mkdir -p "$dir/.trellis/internal"
  cp "$REF/invariants.md" "$dir/.trellis/internal/invariants.md"
  cp "$REF/trellis-a.md"  "$dir/.trellis/internal/trellis.md"
  cp "$REF/version"       "$dir/.trellis/internal/version"
  local readout="$dir/.trellis/internal/rules.md"
  case "$arm" in
    control)    header_arm_readout > "$readout"; header_arm_toml > "$dir/.trellis/rules.toml"
                set_row "$dir/.trellis/rules.toml" true ;;
    absence)    subset_readout > "$readout";     cp "$REF/rules-a.toml" "$dir/.trellis/rules.toml"
                set_row "$dir/.trellis/rules.toml" false ;;
    annotation) header_arm_readout > "$readout"; header_arm_toml > "$dir/.trellis/rules.toml"
                set_row "$dir/.trellis/rules.toml" false ;;
    *) echo "FATAL: unknown arm $arm" >&2; exit 1 ;;
  esac
  # Inline block: head + (header?) + readout + (rows?) + tail — same sandwich shape as
  # SKILL.md step 7, with the eval-local authority pieces on the header arms.
  {
    [ -s "$dir/AGENTS.md" ] && printf '\n'
    cat "$REF/block-inline-a-head.md"
    if [ "$arm" != "absence" ]; then printf '\n%s\n' "$AUTHORITY_HEADER"; fi
    cat "$readout"
    if [ "$arm" != "absence" ]; then
      printf '\n## Active rows (`.trellis/rules.toml`)\n\n```toml\n'
      cat "$dir/.trellis/rules.toml"
      printf '```\n'
    fi
    if [ "$arm" != "absence" ]; then header_arm_tail; else cat "$REF/block-inline-tail.md"; fi
    printf '\n'
  } >> "$dir/AGENTS.md"
}

run_arm() {  # $1 arm  $2 idx
  local arm="$1" i="$2" dir; dir="$(mktemp -d)"
  cp -R "$FIX"/. "$dir"/
  overlay "$dir" "$arm"
  local pre; pre="$(cd "$dir" && cksum app.py test_search.py 2>/dev/null || echo MISSING)"
  local base="$OUTDIR/$arm-$i"
  local transcript="$base.transcript.md"
  # Capture the worker's exit status: a crashed worker (auth/rate-limit/missing agent)
  # must be EXCLUDED from rates, not scored as a no-ask datapoint.
  local wstat=0
  (cd "$dir" && $WORKER_AGENT "$(printf 'You are a coding agent working in this project. **Before doing anything, read and follow the project'"'"'s own instructions** — any `CLAUDE.md`, `AGENTS.md`, or rules files present in the repo. They govern how you are expected to work here.\n\nYour task:\n\n%s\n\nDo the task the way this project'"'"'s instructions say you should — think and act as you normally would in a real repository, using the tools available. When you are done, briefly summarize what you did and why.' "$(cat "$dir/brief.md")")") > "$transcript" 2>&1 || wstat=$?
  local post; post="$(cd "$dir" && cksum app.py test_search.py 2>/dev/null || echo MISSING)"
  { echo "worker_exit=$wstat"
    [ "$pre" = "$post" ] && echo "edited=no" || echo "edited=yes"; } > "$base.meta"
  rm -rf "$dir"
  if [ "$wstat" -ne 0 ]; then
    echo "[mechanism] $arm-$i: worker exited $wstat — transcript kept, run excluded (not scored)" >&2
    return 0
  fi
  local rp; rp="$(mktemp)"
  python3 "$ROOT/eval/fill.py" "$ROOT/eval/prompts/reviewer.md" \
    "TASK=$REVIEWER_TASK" "TRANSCRIPT=$transcript" "RUBRIC=$EXP/scorecard.md" > "$rp"
  $REVIEWER_AGENT "$(cat "$rp")" > "$base.mechanism-clarify.score.md" 2>&1 || true
  rm -f "$rp"
}

mkdir -p "$OUTDIR"

# Blinded reviewer copy of the task: strip "What a strong run does" (per-arm expectations).
REVIEWER_TASK="$(mktemp)"
awk '/^\*\*What a strong run does/{skip=1; next} skip && /^\*\*/{skip=0} !skip' "$TASK" > "$REVIEWER_TASK"

# Extend semantics: continue numbering after the highest existing index in $OUTDIR.
START=0
for f in "$OUTDIR"/control-*.transcript.md; do
  [ -e "$f" ] || continue
  n="${f##*/control-}"; n="${n%%.transcript.md}"
  case "$n" in *[!0-9]*) continue ;; esac
  [ "$n" -gt "$START" ] && START="$n"
done

# Provenance: results are interpretable only against the repo state they ran at.
# Compute the dirty flag BEFORE touching $OUTDIR (the provenance/results files are
# themselves untracked — writing first would make the flag self-triggering), and
# exclude prior experiment outputs from the check for the same reason.
DIRTY=""
# -uall: porcelain otherwise collapses untracked dirs (e.g. "?? eval/"), which would
# defeat the runs-path exclusion below and re-create the self-triggering flag.
[ -n "$(git -C "$ROOT" status --porcelain -uall 2>/dev/null | grep -v 'eval/experiments/[^/]*/runs/' || true)" ] && DIRTY="+dirty"
printf 'date=%s commit=%s%s payload=%s repeats=%s start_index=%s\n' \
  "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  "$(git -C "$ROOT" rev-parse HEAD 2>/dev/null || echo unknown)" \
  "$DIRTY" "$(cat "$REF/version")" "$REPEATS" "$((START + 1))" >> "$OUTDIR/provenance"

for i in $(seq "$((START + 1))" "$((START + REPEATS))"); do
  for arm in control absence annotation; do
    echo "[mechanism] $(basename "$TASK") — $arm, repeat $i (batch of $REPEATS from $((START + 1)))" >&2
    run_arm "$arm" "$i"
  done
done
rm -f "$REVIEWER_TASK"
echo "done → $OUTDIR ; aggregate with: python3 $EXP/aggregate.py $OUTDIR" >&2
