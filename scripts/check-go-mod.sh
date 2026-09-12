#!/bin/sh
set -eu

repo_root="$(cd -- "$(dirname -- "$0")/.." && pwd)"
work="$(mktemp -d "${TMPDIR:-/tmp}/trellis-go-mod.XXXXXX")"
trap 'rm -rf "$work"' EXIT HUP INT TERM

cp "$repo_root/cli/go.mod" "$work/go.mod"
cp "$repo_root"/cli/*.go "$work/"

(
  cd "$work"
  go mod tidy
)

if ! cmp -s "$repo_root/cli/go.mod" "$work/go.mod" || [ -e "$work/go.sum" ]; then
  echo "cli/go.mod is not tidy; run 'cd cli && go mod tidy' and commit the result." >&2
  exit 1
fi
