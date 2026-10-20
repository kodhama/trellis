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
# The versions are the values package.json's "version" field has taken on the
# first-parent history of HEAD. The commit that set a version is the first commit
# at which the field took that value, so a release whose tag job failed is still
# tagged on its own commit after later pushes at the same version.
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
if ! printf '%s\n' "$floor" | grep -Eq '^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'; then
  echo "release-tags: the floor must be a version such as 0.25.0, got \"$floor\"" >&2
  exit 2
fi

# at_or_above <version>: true when the version sorts at or above the floor.
at_or_above() {
  [ "$(printf '%s\n%s\n' "$floor" "$1" | sort -V | head -n 1)" = "$floor" ]
}

# version_at <commit>: package.json's "version" at that commit, or nothing.
version_at() {
  git show "$1:package.json" 2>/dev/null |
    node -e 'let s = ""; process.stdin.on("data", (d) => (s += d)).on("end", () => { try { process.stdout.write(String(JSON.parse(s).version ?? "")); } catch { /* not JSON: no version */ } });'
}

# list_releases: each "<version> <commit>" pair, oldest first, for the commit
# that first set it.
list_releases() {
  prev=""
  seen=" "
  git rev-list --first-parent --reverse HEAD -- package.json | while read -r commit; do
    version="$(version_at "$commit")"
    if [ -n "$version" ] && [ "$version" != "$prev" ]; then
      case "$seen" in
        *" $version "*) ;;
        *)
          printf '%s %s\n' "$version" "$commit"
          seen="$seen$version "
          ;;
      esac
    fi
    prev="$version"
  done
}

releases="$(list_releases)"

failed=0
while read -r version commit; do
  [ -n "$version" ] || continue
  at_or_above "$version" || continue
  tag="v$version"
  tagged="$(git rev-parse -q --verify "refs/tags/$tag^{commit}" || true)"
  if [ "$mode" = plan ]; then
    [ -n "$tagged" ] || printf '%s %s\n' "$tag" "$commit"
    continue
  fi
  if [ -z "$tagged" ]; then
    echo "::error::$tag is missing: $commit set version $version on main, and no tag names it."
    failed=1
  elif [ "$tagged" != "$commit" ]; then
    echo "::error::$tag names $tagged, but $commit set version $version. Tags are never moved, so this version is burned until the maintainer decides."
    failed=1
  else
    echo "ok — $tag names $commit"
  fi
done <<EOF
$releases
EOF

exit "$failed"
