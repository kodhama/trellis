#!/bin/sh
#
# decision-id-guard.sh — refuse a pull request that claims an already-taken
# `decisions/NNNN-*.md` id. TRL-40 / decision-0089.
#
# WHY THIS EXISTS. `decisions/` on `main` is not the allocation authority; the
# merge queue is. Every branch does the right thing in isolation — list
# `decisions/`, take the next free number — and two branches cut before the race
# take the same one. It happened three times (decision-0077, then trellis#252 on
# an already-taken decision-0076, then decision-0086 on both trellis#262 and
# trellis#263). Nothing failed on any of them; a human noticed.
#
# THE RULE, in three parts:
#   1. An id already present on the base branch is taken. Red.
#   2. An id also claimed by another OPEN pull request is a collision on both,
#      but only the HIGHER-numbered PR goes red — the lower-numbered claim was
#      there first. Failing both would leave neither able to merge without
#      talking to the other. The lower-numbered PR gets a notice naming the
#      rival, and stays green.
#   3. Two files in ONE diff claiming one id is red on that PR alone. A branch
#      that collides with itself needs no tie-break.
#
# WHY A SCRIPT AND NOT INLINE YAML: so `cli/decision_id_guard_test.go` can
# execute it against fixtures. Every input is injectable, which is also why the
# status filter below lives here rather than in the `gh --jq` expression — what
# counts as a claim is the subtlest rule in the file, and it has to be in the
# tested half.
#
# INPUTS (each one overridden by an env var; unset means "go and fetch it"):
#
#   GUARD_PR_NUMBER   required. This pull request's number.
#   GUARD_BASE_REF    base branch name. Default `main`.
#   GUARD_REPO        `owner/name`, for the live `gh` path only.
#   GUARD_MAIN_FILES  newline-separated paths on the base branch.
#                     Unset -> `git ls-tree` against a shallow fetch of it.
#   GUARD_PR_FILES    newline-separated `<pr-number> <status> <path> [old-path]`
#                     rows, covering every open PR INCLUDING this one. The
#                     fourth field is GitHub's `previous_filename` and is
#                     present only on a rename. Paths must be space-free, which
#                     every `decisions/NNNN-slug.md` is; a row with more fields
#                     than its status allows is REFUSED (exit 2), never treated
#                     as "no claim" — a space would otherwise split the path
#                     across fields, match no decision shape, and pass green
#                     against a taken id.
#                     Unset -> `gh api` over every open PR, paginated.
#
# EXIT: 0 clean (a notice is still exit 0), 1 collision, 2 could not run.
#
# EXIT 2 IS LOAD-BEARING. This script is `sh`, not bash, so `pipefail` is not
# portably available (ubuntu-latest's /bin/sh is dash). Rather than rely on it,
# every processing step writes to a file and its status is checked: a failing
# `awk` must not leave an empty result that reads as "no id claimed, all clear".

set -u

work=$(mktemp -d) || exit 2
trap 'rm -rf "$work"' EXIT

base=${GUARD_BASE_REF:-main}
repo=${GUARD_REPO:-}
pr=${GUARD_PR_NUMBER:-}
fail=0

if [ -z "$pr" ]; then
	echo "decision-id-guard: GUARD_PR_NUMBER is required (this pull request's number)." >&2
	exit 2
fi

# A processing step failed. Never degrade to a pass.
cannot_run() {
	echo "decision-id-guard: $1" >&2
	exit 2
}

# decisions/0089-a-slug.md -> 0089 ; anything else -> nothing.
# Four digits and a dash is the whole shape: `decisions/README.md` is not a
# claim, and neither is `decisions/0089.md`.
decision_id() {
	case "$1" in
	decisions/[0-9][0-9][0-9][0-9]-*.md)
		d=${1#decisions/}
		printf '%s\n' "${d%%-*}"
		;;
	*) ;;
	esac
}

# --- input: what the base branch already holds --------------------------------
if [ -n "${GUARD_MAIN_FILES+x}" ]; then
	printf '%s\n' "$GUARD_MAIN_FILES" >"$work/base-files"
else
	git fetch --no-tags --depth=1 origin "$base" >/dev/null 2>&1 ||
		cannot_run "could not fetch origin/$base."
	git ls-tree -r --name-only FETCH_HEAD -- decisions/ >"$work/base-files" ||
		cannot_run "could not read decisions/ on origin/$base."
fi

: >"$work/base-ids" # `<id> <path>`
while IFS= read -r f; do
	[ -n "$f" ] || continue
	id=$(decision_id "$f")
	[ -n "$id" ] || continue
	printf '%s %s\n' "$id" "$f" >>"$work/base-ids"
done <"$work/base-files"

# --- input: every open PR's files, this one included --------------------------
if [ -n "${GUARD_PR_FILES+x}" ]; then
	printf '%s\n' "$GUARD_PR_FILES" >"$work/pr-files"
else
	[ -n "$repo" ] ||
		cannot_run "GUARD_REPO is required when GUARD_PR_FILES is not supplied."

	# EVERY open PR, and the comments above mean it. `gh pr list --limit N`
	# caps silently: past the cap it returns a short list and exits 0, so an
	# id held by an unlisted PR would read as free. --paginate has no cap.
	gh api "repos/$repo/pulls" -X GET -f state=open -f per_page=100 --paginate \
		--jq '.[].number' >"$work/open-prs" ||
		cannot_run "could not list open pull requests."

	# This PR must be in the list even if it is not open (a closed PR still
	# gets checks) — duplicates are removed below.
	printf '%s\n' "$pr" >>"$work/open-prs"
	sort -nu "$work/open-prs" >"$work/open-prs.sorted" ||
		cannot_run "could not sort the open pull request list."

	: >"$work/pr-files"
	while read -r n; do
		[ -n "$n" ] || continue
		# previous_filename is GitHub's OLD path, present only on a rename.
		# </dev/null: this loop's stdin IS the PR list. `gh api` does not read
		# stdin on a GET, so nothing is broken today; closing it costs nothing
		# and means a future flag that does read stdin cannot silently eat the
		# remaining PR numbers and shorten the sweep.
		gh api "repos/$repo/pulls/$n/files" --paginate \
			--jq '.[] | "\(.status) \(.filename) \(.previous_filename // "")"' \
			< /dev/null >"$work/one-pr" ||
			cannot_run "could not read the files of PR #$n."
		while IFS= read -r row; do
			[ -n "$row" ] || continue
			printf '%s %s\n' "$n" "$row" >>"$work/pr-files"
		done <"$work/one-pr"
	done <"$work/open-prs.sorted"
fi

# One shape for both sources. The live jq template emits the fourth field
# unconditionally — `previous_filename // ""` is empty on anything but a rename —
# so a production row carries a trailing space that an injected fixture does not.
# An anchored pattern written against the fixtures then silently never matches in
# CI, which is how the `removed`-source clause below was dead in production while
# green in tests. Normalising here means the tests exercise the rows CI feeds.
sed 's/[[:space:]][[:space:]]*$//' "$work/pr-files" >"$work/pr-files.trimmed" ||
	cannot_run "could not normalise the pull request file rows."
mv "$work/pr-files.trimmed" "$work/pr-files" ||
	cannot_run "could not replace the pull request file rows."

# --- what counts as a claim ---------------------------------------------------
#
# A claim is a file this PR puts at a NEW `decisions/NNNN-*.md` path:
#
#   added    the obvious case.
#   copied   the destination path is new and the source survives, so it claims.
#   renamed  GitHub reports the NEW path in `filename` and the old one in
#            `previous_filename`. Renaming `0088-old.md` to `0090-new.md`
#            therefore claims 0090 — judge the DESTINATION id. A pure slug
#            rename (`0090-a.md` -> `0090-b.md`) keeps its own id and claims
#            nothing.
#
# `modified`, `removed`, `changed` and anything else are never claims: editing
# decision-0078 does not take 0078.
#
# DELIBERATELY NOT DECIDED HERE: whether a rename RELEASES the source id. Once
# the PR merges the old path is gone and the id looks free — but until then the
# source record is still on the base branch, and handing that id to a second
# branch on the strength of an unmerged rename would cause exactly the collision
# this guard exists to prevent. So the source id stays taken. If that is ever
# wrong, it is wrong in the direction of one wasted number rather than a
# duplicate.
#
# KNOWN, AND NOT FIXED HERE: rename detection is GitHub's, not ours. If a
# re-slug changes enough of the file that the API reports it as `removed` +
# `added` rather than `renamed`, the `added` row claims an id that is on the
# base branch — its own — and the guard reds a legitimate retitle. Renumbering
# to dodge that would be wrong; the remedy is to merge anyway (the check is
# advisory) or to split the rename from the content edit. Pre-dates the rename
# rule above rather than being caused by it: under the old `added`-only rule
# the same pair failed the same way.
: >"$work/claims" # `<pr> <id> <path>`
while read -r p st path prev extra; do
	[ -n "${p:-}" ] || continue

	# A space in a path splits it across fields, so it matches no decision shape
	# and reads as "not a claim" — green against a taken id, which is the one
	# direction this guard exists to eliminate. A surplus field is therefore a
	# refusal rather than a pass.
	#
	# SCOPED TO decisions/ ON PURPOSE, and this is load-bearing. The check sees
	# every row of every open PR, so refusing on any spaced path anywhere would
	# let one `docs/my notes.md` in one unrelated PR abort the guard for every
	# PR in the repo — including PRs that touch no decision file at all. That is
	# a repo-wide CI outage triggered by a filename, traded for a hole nobody
	# has ever hit. A row whose path field does not begin with `decisions/`
	# cannot become a claim, so its shape is not this check's business.
	#
	# Known over-caution, and it fails closed: a rename INTO decisions/ whose
	# SOURCE path contains a space refuses too, though the destination parsed
	# fine. Loud and actionable, unlike the silent pass it replaces.
	case "${path:-}" in
	decisions/*)
		if [ -n "${extra:-}" ] || { [ -n "${prev:-}" ] && [ "${st:-}" != "renamed" ]; }; then
			cannot_run "unparseable row for PR #$p — a decisions/ path containing a space cannot be read as a claim: $p $st $path $prev ${extra:-}"
		fi
		;;
	esac

	id=$(decision_id "${path:-}")
	[ -n "$id" ] || continue
	case "${st:-}" in
	added | copied) ;;
	renamed)
		# Same id at both ends: the record kept its number, only its slug moved.
		if [ "$(decision_id "${prev:-}")" = "$id" ]; then
			continue
		fi
		;;
	*) continue ;;
	esac
	printf '%s %s %s\n' "$p" "$id" "$path" >>"$work/claims"
done <"$work/pr-files"

awk -v me="$pr" '$1 == me { print $2 " " $3 }' "$work/claims" >"$work/mine.raw" ||
	cannot_run "could not select this pull request's claims."
sort -u "$work/mine.raw" >"$work/mine" ||
	cannot_run "could not sort this pull request's claims."

if [ ! -s "$work/mine" ]; then
	echo "PR #$pr adds no decisions/NNNN-*.md file — no id claimed, nothing to check."
	exit 0
fi

cut -d' ' -f1 "$work/mine" >"$work/mine-ids" ||
	cannot_run "could not read the claimed ids."
sort "$work/mine-ids" >"$work/mine-ids.sorted" ||
	cannot_run "could not sort the claimed ids."

tr '\n' ' ' <"$work/mine-ids.sorted" >"$work/mine-ids.line" ||
	cannot_run "could not format the claimed ids for display."
echo "PR #$pr claims: $(cat "$work/mine-ids.line")"
echo "Checking against $base and every other open pull request."
echo

# --- rule 3: this PR colliding with itself ------------------------------------
# `mine` is deduplicated on (id, path), so one id under two different filenames
# survives that dedup and lands here. No tie-break applies: both files are on
# the same branch, and the author picks which one moves.
uniq -d "$work/mine-ids.sorted" >"$work/dups" ||
	cannot_run "could not scan this pull request's claims for duplicates."
while read -r dup; do
	[ -n "$dup" ] || continue
	files=$(awk -v i="$dup" '$1 == i { printf "%s ", $2 }' "$work/mine") ||
		cannot_run "could not list the files claiming decision-$dup."
	echo "::error::PR #$pr adds two files claiming decision-$dup: ${files}— one id, one record. Renumber one of them."
	fail=1
done <"$work/dups"

# --- rules 1 and 2: the base branch, then the other open PRs ------------------
while read -r id path; do
	[ -n "$id" ] || continue

	taken=$(awk -v i="$id" '$1 == i { print $2; exit }' "$work/base-ids") ||
		cannot_run "could not check decision-$id against $base."
	if [ -n "$taken" ]; then
		# When the file named as "already on $base" is the one THIS PR is
		# renaming away, the message reads as a contradiction. The behaviour is
		# right — the source id stays taken until the rename merges — but that
		# reasoning lives in a comment the reader never sees, so say it here.
		# Arises when one diff both moves a record away from an id and puts
		# another record on it — two renames, or a remove plus an add.
		same_pr_source=""
		if grep -q "^$pr renamed [^ ]* $taken\$" "$work/pr-files" 2>/dev/null ||
			grep -q "^$pr removed $taken\$" "$work/pr-files" 2>/dev/null; then
			same_pr_source=" NOTE: $taken is the very file this pull request moves off that id. It still counts, because the id is only free once this pull request merges, and handing it out before then would cause the collision this check exists to prevent — a branch cannot free an id for its own use."
		fi
		echo "::error file=$path::decision-$id is already on $base as $taken. Ids are allocated by whoever merges first, so a branch cut before that merge cannot see it. Renumber to an id free on $base AND on every open pull request.$same_pr_source"
		fail=1
	fi

	awk -v i="$id" -v me="$pr" '$2 == i && $1 != me { print $1 " " $3 }' \
		"$work/claims" >"$work/rivals.raw" ||
		cannot_run "could not check decision-$id against the other open pull requests."
	sort -n "$work/rivals.raw" >"$work/rivals" ||
		cannot_run "could not order the pull requests claiming decision-$id."

	while read -r rival rpath; do
		[ -n "$rival" ] || continue
		if [ "$rival" -lt "$pr" ]; then
			echo "::error file=$path::decision-$id is also claimed by open PR #$rival ($rpath). TIE-BREAK: the older claim wins — the lower-numbered open PR keeps the id, so #$rival keeps decision-$id and this PR (#$pr) renumbers."
			fail=1
		else
			echo "::notice file=$path::decision-$id is also claimed by open PR #$rival ($rpath). TIE-BREAK: the older claim wins — the lower-numbered open PR keeps the id, so this PR (#$pr) keeps decision-$id and #$rival renumbers. Reported here, red there."
		fi
	done <"$work/rivals"

	# "free" must not print for an id this PR claims twice — rule 3 already
	# failed it, and a line saying otherwise contradicts the error above it.
	if [ -z "$taken" ] && [ ! -s "$work/rivals" ] && ! grep -qx "$id" "$work/dups"; then
		echo "  - decision-$id ($path) — free."
	fi
done <"$work/mine"

if [ "$fail" -ne 0 ]; then
	echo
	echo "A decision id must be free on $base, on every open pull request numbered"
	echo "below this one, and within this pull request's own diff. Where two open"
	echo "PRs claim one id, the lower-numbered PR keeps it and the higher-numbered"
	echo "PR renumbers — failing both would leave neither able to merge without"
	echo "talking to the other."
	exit 1
fi

echo
echo "Every decision id claimed by PR #$pr is free."
