#!/bin/sh
# Enforce the coverage floor: check-coverage.sh <coverprofile> <floor-percent>.
# The floor is a ratchet set below the current measurement (92.3% at
# introduction); raise it as coverage climbs, and lower it only with a reason
# recorded in the PR.
set -eu

profile=$1
floor=$2

total=$(go tool cover -func="$profile" | awk '/^total:/ {gsub(/%/, "", $NF); print $NF}')
[ -n "$total" ] || { echo "no total coverage line in $profile" >&2; exit 1; }

if awk -v t="$total" -v f="$floor" 'BEGIN { exit !(t+0 >= f+0) }'; then
  echo "coverage ${total}% meets the ${floor}% floor"
else
  echo "coverage ${total}% is below the ${floor}% floor" >&2
  exit 1
fi
