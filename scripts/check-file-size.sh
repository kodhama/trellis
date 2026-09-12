#!/bin/sh
# Large-file detection: fail when a tracked file exceeds the byte ceiling. The
# ceiling catches accidentally committed binaries, vendored bundles, and
# generated artifacts — the failure mode that is invisible in a diff review and
# permanent in history. 500 KB is far above the largest legitimate file today
# (~120 KB), so it bites on accidents, not on growth.
set -eu

limit=512000

oversized=$(git ls-files | while IFS= read -r f; do
  size=$(wc -c < "$f" | tr -d ' ')
  [ "$size" -le "$limit" ] || printf '%s (%s bytes)\n' "$f" "$size"
done)

if [ -n "$oversized" ]; then
  echo "Tracked files over ${limit} bytes:" >&2
  echo "$oversized" >&2
  exit 1
fi
