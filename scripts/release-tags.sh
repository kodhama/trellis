#!/bin/sh
# The tag logic behind .github/workflows/release-tag.yml, kept in a script so
# cli/release_tags_test.go can drive it against throwaway repositories.
#
# decision-2026-10-20-trellis-delivers-rules-as-committed-files point 11: every
# CLI release is tagged v<version> on the commit that set that version, tags are
# never moved, and versions below a floor committed in the workflow file are
# skipped. The floor arrives as TRELLIS_TAG_FLOOR; while it is empty, both modes
# do nothing, so no plugin release before the CLI's first is tagged.
#
#   release-tags.sh plan    print "v<version> <commit>" for each version at or
#                           above the floor that has no tag yet
#   release-tags.sh verify  fail naming each version at or above the floor
#                           whose tag is missing or names another commit
#
# The versions are the values package.json's top-level "version" field has taken
# on the first-parent history of HEAD. The commit that set a version is the commit
# at which the field took that value, so a release whose tag job failed is still
# tagged on its own commit after later pushes at the same version. A version at or
# above the floor must be plain X.Y.Z, because `sort -V` (GNU and BSD, not POSIX)
# orders a prerelease after its release, and it must never come back once main
# has moved past it, because one tag can name only one of its commits.
set -eu

mode="${1:-}"
case "$mode" in
  plan | verify) ;;
  *)
    echo "usage: release-tags.sh plan|verify" >&2
    exit 2
    ;;
esac

floor="${TRELLIS_TAG_FLOOR:-}"
if [ -z "$floor" ]; then
  echo "release-tags: no floor is set, so no version is tagged or checked" >&2
  exit 0
fi
plain='^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'
if ! printf '%s\n' "$floor" | grep -Eq "$plain"; then
  echo "release-tags: the floor must be a version such as 0.25.0, got \"$floor\"" >&2
  exit 2
fi
if [ "$(git rev-parse --is-shallow-repository)" = true ]; then
  echo "release-tags: this clone is shallow, so its version history is incomplete; check out with fetch-depth: 0" >&2
  exit 2
fi

# at_or_above <version>: true when the version sorts at or above the floor.
at_or_above() {
  [ "$(printf '%s\n%s\n' "$floor" "$1" | sort -V | head -n 1)" = "$floor" ]
}

# version_at <commit>: package.json's top-level "version" at that commit, or
# nothing when the file or the field is absent.
version_at() {
  git show "$1:package.json" 2>/dev/null |
    node -e 'let s = ""; process.stdin.on("data", (d) => (s += d)).on("end", () => { try { process.stdout.write(String(JSON.parse(s).version ?? "")); } catch { /* not JSON: no version */ } });'
}

# list_releases: one "<version> <commit>" line, oldest first, for each commit at
# which the version changed. A commit with no version leaves it as it was.
list_releases() {
  prev=""
  git rev-list --first-parent --reverse HEAD -- package.json | while read -r commit; do
    version="$(version_at "$commit")"
    [ -n "$version" ] || continue
    [ "$version" = "$prev" ] || printf '%s %s\n' "$version" "$commit"
    prev="$version"
  done
}

# fail <message>: plan's stdout carries only tag lines, so its errors go to
# stderr; verify's become workflow annotations.
fail() {
  if [ "$mode" = plan ]; then echo "release-tags: $1" >&2; else echo "::error::$1"; fi
  failed=1
}

releases="$(list_releases)"

failed=0
while read -r version commit; do
  [ -n "$version" ] || continue
  at_or_above "$version" || continue
  if ! printf '%s\n' "$version" | grep -Eq "$plain"; then
    echo "release-tags: $commit sets version $version, which is at or above the floor but not plain X.Y.Z" >&2
    exit 2
  fi
  first="$(printf '%s\n' "$releases" | awk -v v="$version" '$1 == v { print $2; exit }')"
  if [ "$first" != "$commit" ]; then
    fail "version $version came back at $commit after $first had set it and main had moved past it. A tag names one commit, so this release needs a new version."
    continue
  fi
  tag="v$version"
  tagged="$(git rev-parse -q --verify "refs/tags/$tag^{commit}" || true)"
  if [ "$mode" = plan ]; then
    [ -n "$tagged" ] || printf '%s %s\n' "$tag" "$commit"
  elif [ -z "$tagged" ]; then
    fail "$tag is missing: $commit set version $version on main, and no tag names it."
  elif [ "$tagged" != "$commit" ]; then
    fail "$tag names $tagged, but $commit set version $version. Tags are never moved, so this version is burned until the maintainer decides."
  else
    echo "ok — $tag names $commit"
  fi
done <<EOF
$releases
EOF

exit "$failed"
