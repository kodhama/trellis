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
# THE RULE, in two halves:
#   1. An id already present on the base branch is taken. Red.
#   2. An id also claimed by another OPEN pull request is a collision on both,
#      but only the HIGHER-numbered PR goes red — the lower-numbered claim was
#      there first. Failing both would leave neither able to merge without
#      talking to the other. The lower-numbered PR gets a notice naming the
#      rival, and stays green.
#
# WHY A SCRIPT AND NOT INLINE YAML: so `cli/decision_id_guard_test.go` can
# execute it against fixtures. Every input is injectable, which is also why the
# status filter below lives here rather than in the `gh --jq` expression — a
# renamed or modified decision file is not a claim, and that rule has to be in
# the tested half.
#
# INPUTS (each one overridden by an env var; unset means "go and fetch it"):
#
#   GUARD_PR_NUMBER   required. This pull request's number.
#   GUARD_BASE_REF    base branch name. Default `main`.
#   GUARD_REPO        `owner/name`, for the live `gh` path only.
#   GUARD_MAIN_FILES  newline-separated paths on the base branch.
#                     Unset -> `git ls-tree` against a shallow fetch of it.
#   GUARD_PR_FILES    newline-separated `<pr-number> <status> <path>` rows,
#                     covering every open PR INCLUDING this one.
#                     Unset -> `gh pr list` + `gh api .../files`.
#
# EXIT: 0 clean (a notice is still exit 0), 1 collision, 2 could not run.

set -u

work=$(mktemp -d) || exit 2
trap 'rm -rf "$work"' EXIT

base=${GUARD_BASE_REF:-main}
repo=${GUARD_REPO:-}
pr=${GUARD_PR_NUMBER:-}

if [ -z "$pr" ]; then
	echo "decision-id-guard: GUARD_PR_NUMBER is required (this pull request's number)." >&2
	exit 2
fi

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
	if ! git fetch --no-tags --depth=1 origin "$base" >/dev/null 2>&1; then
		echo "decision-id-guard: could not fetch origin/$base." >&2
		exit 2
	fi
	if ! git ls-tree -r --name-only FETCH_HEAD -- decisions/ >"$work/base-files"; then
		echo "decision-id-guard: could not read decisions/ on origin/$base." >&2
		exit 2
	fi
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
	if [ -z "$repo" ]; then
		echo "decision-id-guard: GUARD_REPO is required when GUARD_PR_FILES is not supplied." >&2
		exit 2
	fi
	if ! gh pr list --repo "$repo" --state open --limit 200 --json number --jq '.[].number' >"$work/open-prs"; then
		echo "decision-id-guard: could not list open pull requests." >&2
		exit 2
	fi
	# This PR must be in the list even if `gh pr list` paged it out.
	printf '%s\n' "$pr" >>"$work/open-prs"
	: >"$work/pr-files"
	sort -nu "$work/open-prs" >"$work/open-prs.sorted"
	while read -r n; do
		[ -n "$n" ] || continue
		if ! gh api "repos/$repo/pulls/$n/files" --paginate \
			--jq '.[] | "\(.status) \(.filename)"' >"$work/one-pr"; then
			echo "decision-id-guard: could not read the files of PR #$n." >&2
			exit 2
		fi
		while IFS= read -r row; do
			[ -n "$row" ] || continue
			printf '%s %s\n' "$n" "$row" >>"$work/pr-files"
		done <"$work/one-pr"
	done <"$work/open-prs.sorted"
fi

# A claim is an ADDED file whose name has the decision shape. A modified or
# renamed decision file is not a claim: modifying 0078 does not take 0078, and a
# rename is a file that already exists somewhere.
: >"$work/claims" # `<pr> <id> <path>`
while read -r p st path; do
	[ -n "${p:-}" ] || continue
	[ "${st:-}" = "added" ] || continue
	id=$(decision_id "${path:-}")
	[ -n "$id" ] || continue
	printf '%s %s %s\n' "$p" "$id" "$path" >>"$work/claims"
done <"$work/pr-files"

awk -v me="$pr" '$1 == me { print $2 " " $3 }' "$work/claims" | sort -u >"$work/mine"

if [ ! -s "$work/mine" ]; then
	echo "PR #$pr adds no decisions/NNNN-*.md file — no id claimed, nothing to check."
	exit 0
fi

echo "PR #$pr claims: $(cut -d' ' -f1 "$work/mine" | tr '\n' ' ')"
echo "Checking against $base and every other open pull request."
echo

fail=0
while read -r id path; do
	[ -n "$id" ] || continue

	taken=$(awk -v i="$id" '$1 == i { print $2; exit }' "$work/base-ids")
	if [ -n "$taken" ]; then
		echo "::error file=$path::decision-$id is already on $base as $taken. Ids are allocated by whoever merges first, so a branch cut before that merge cannot see it. Renumber to an id free on $base AND on every open pull request."
		fail=1
	fi

	awk -v i="$id" -v me="$pr" '$2 == i && $1 != me { print $1 " " $3 }' "$work/claims" |
		sort -n >"$work/rivals"

	while read -r rival rpath; do
		[ -n "$rival" ] || continue
		if [ "$rival" -lt "$pr" ]; then
			echo "::error file=$path::decision-$id is also claimed by open PR #$rival ($rpath). TIE-BREAK: the older claim wins — the lower-numbered open PR keeps the id, so #$rival keeps decision-$id and this PR (#$pr) renumbers."
			fail=1
		else
			echo "::notice file=$path::decision-$id is also claimed by open PR #$rival ($rpath). TIE-BREAK: the older claim wins — the lower-numbered open PR keeps the id, so this PR (#$pr) keeps decision-$id and #$rival renumbers. Reported here, red there."
		fi
	done <"$work/rivals"

	if [ -z "$taken" ] && [ ! -s "$work/rivals" ]; then
		echo "  - decision-$id ($path) — free."
	fi
done <"$work/mine"

if [ "$fail" -ne 0 ]; then
	echo
	echo "A decision id must be free on $base and on every open pull request numbered"
	echo "below this one. Where two open PRs claim one id, the lower-numbered PR keeps"
	echo "it and the higher-numbered PR renumbers — failing both would leave neither"
	echo "able to merge without talking to the other."
	exit 1
fi

echo
echo "Every decision id claimed by PR #$pr is free."
