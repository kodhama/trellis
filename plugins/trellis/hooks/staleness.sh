#!/usr/bin/env bash
# Trellis staleness surface — SessionStart hook (decision-0039 rule 1, mechanics
# reworked by decision-0043 / kodhama-0007 slice 4, kodhama/trellis#120).
#
# Binary-free and git-free: compares the project's .trellis/version stamp against the
# installed plugin's reference/version payload stamp — a file-to-file comparison. Both
# sides speak payload@<content-hash> (the stamp changes exactly when the payload
# content changes), so the nudge fires only when the overlay genuinely differs from
# what the installed plugin would write. Legacy stamps (plugin@<sha> from pre-#120
# skill installs, bare semver from the retired CLI) differ by construction and draw a
# one-time migration nudge — with `trellis status` retired, this hook is the only
# drift surface (decision-0035: drift is made visible, not silent).
#
# Output contract (SessionStart): exit 0; a single-line JSON {"additionalContext": "..."}
# on stdout injects context; empty stdout injects nothing. Never exit non-zero — a hook
# failure must not disrupt the session.

ver="${CLAUDE_PROJECT_DIR:-.}/.trellis/version"
[ -f "$ver" ] || exit 0                          # no overlay here → nothing to say

overlay="$(head -n1 "$ver" 2>/dev/null | tr -d '[:space:]')"
[ -n "$overlay" ] || exit 0                       # empty stamp → nothing to compare

ref="${CLAUDE_PLUGIN_ROOT:-/nonexistent}/reference/version"
current="$(head -n1 "$ref" 2>/dev/null | tr -d '[:space:]')"
[ -n "$current" ] || exit 0                       # can't read the installed payload → silent

if [ "$overlay" != "$current" ]; then
  printf '{"additionalContext": "Trellis overlay may be stale: this project'"'"'s .trellis/version stamp is %s, but the installed Trellis plugin ships payload %s. The invariants may have moved on; run /trellis:setup to refresh the overlay (or re-copy it from the plugin'"'"'s reference/ payload)."}\n' \
    "$overlay" "$current"
fi
exit 0
