#!/usr/bin/env bash
# Trellis staleness surface — SessionStart hook (decision-0039 rule 1, mechanics
# reworked by decision-0043 / kodhama-0007 slice 4, kodhama/trellis#120; compared
# path moved by decision-0051's authority split).
#
# Binary-free and git-free: compares the project's .trellis/internal/version stamp
# against the installed plugin's reference/version payload stamp — a file-to-file
# comparison. Both sides speak payload@<content-hash> (the stamp changes exactly when
# the payload content changes), so the nudge fires only when the overlay genuinely
# differs from what the installed plugin would write. A stamp found only at the
# legacy flat path (.trellis/version — pre-decision-0051 layouts, and before them the
# plugin@<sha> / bare-semver stamps of pre-#120 installs) always draws the nudge: the
# layout itself is stale, and /trellis:setup's refresh is the migration vehicle. With
# `trellis status` retired, this hook is the only drift surface (decision-0035: drift
# is made visible, not silent).
#
# Output contract (SessionStart): exit 0; a single-line JSON {"additionalContext": "..."}
# on stdout injects context; empty stdout injects nothing. Never exit non-zero — a hook
# failure must not disrupt the session.

root="${CLAUDE_PROJECT_DIR:-.}"
ref="${CLAUDE_PLUGIN_ROOT:-/nonexistent}/reference/version"
current="$(head -n1 "$ref" 2>/dev/null | tr -d '[:space:]')"

ver="$root/.trellis/internal/version"
if [ -f "$ver" ]; then
  overlay="$(head -n1 "$ver" 2>/dev/null | tr -d '[:space:]')"
  [ -n "$overlay" ] || exit 0                     # empty stamp → nothing to compare
  [ -n "$current" ] || exit 0                     # can't read the installed payload → silent
  if [ "$overlay" != "$current" ]; then
    printf '{"additionalContext": "Trellis overlay may be stale: this project'"'"'s .trellis/internal/version stamp is %s, but the installed Trellis plugin ships payload %s. The invariants may have moved on; run /trellis:setup to refresh the overlay (or re-copy it from the plugin'"'"'s reference/ payload)."}\n' \
      "$overlay" "$current"
  fi
  exit 0
fi

legacy="$root/.trellis/version"
if [ -f "$legacy" ]; then
  overlay="$(head -n1 "$legacy" 2>/dev/null | tr -d '[:space:]')"
  [ -n "$overlay" ] || exit 0                     # empty stamp → nothing to compare
  [ -n "$current" ] || exit 0                     # can't read the installed payload → silent
  printf '{"additionalContext": "Trellis overlay predates the .trellis/internal/ layout (decision-0051): its stamp sits at the legacy path .trellis/version (%s; the installed plugin ships payload %s). Run /trellis:setup to refresh — the refresh migrates the overlay to the new layout."}\n' \
    "$overlay" "$current"
fi
exit 0
