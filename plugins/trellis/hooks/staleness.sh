#!/usr/bin/env bash
# Trellis SessionStart hook (decision-0039 rule 1, mechanics reworked by
# decision-0043 / kodhama-0007 slice 4, kodhama/trellis#120; compared path moved
# by decision-0051's authority split).
#
# Three paths, selected by what the project actually has:
#
#   A. Vendored overlay (.trellis/internal/version present) — the staleness
#      surface, unchanged. Compares the project's stamp against the installed
#      plugin's reference/version payload stamp, a file-to-file comparison. Both
#      sides speak payload@<content-hash> (the stamp changes exactly when the
#      payload content changes), so the nudge fires only when the overlay
#      genuinely differs from what the installed plugin would write. A stamp
#      found only at the legacy flat path (.trellis/version — pre-decision-0051
#      layouts, and before them the plugin@<sha> / bare-semver stamps of pre-#120
#      installs) always draws the nudge: the layout itself is stale, and
#      /trellis:setup's refresh is the migration vehicle. With `trellis status`
#      retired, this hook is the only drift surface (decision-0035: drift is made
#      visible, not silent).
#
#   B. Config only (.trellis/rules.toml present, no .trellis/internal/ directory) —
#      plugin-native delivery. The rules are injected from the installed plugin's
#      own payload instead of read from vendored copies. Same always-loaded chain
#      the import channel delivers — posture header, rules, live rows — so the
#      tested wording stays the shipped wording (decision-0053). The one edit is
#      repointing the invariants path at the plugin, which is where the file
#      actually is in this mode, and which therefore cannot go stale.
#
#   C. Curl install (.claude/rules/trellis.md present) — the install path
#      rendered that file and Claude Code loads it at launch on its own, so this
#      hook injects nothing and says which artifact it deferred to
#      (decision-0068 D10). Checked with -s, not -f: a truncated or zero-byte
#      file must not silence this hook while governing nothing.
#
# `.trellis/rules.toml` is the opt-in signal for path B. A project with none of
# the three gets nothing: this plugin may be installed user-wide, and a project
# that never adopted Trellis must not be governed by surprise. Path C is the one
# exception — it fires on its own artifact, with or without rules.toml, because
# that file is itself proof the project adopted Trellis.
#
# The three paths are mutually exclusive by construction, so a project that has
# vendored the overlay, or installed by curl, never receives the rules twice.
# Order matters: A before C, so a project MIGRATING off a vendored overlay still
# gets its staleness nudge (decision-0035: drift is made visible, not silent).
#
# Binary-free, git-free and node-free: bash plus head/tr/awk (decision-0010 —
# Trellis resources are agent instructions that require no runtime).
#
# Output contract (SessionStart): exit 0; a single-line JSON object on stdout
# injects context; empty stdout injects nothing. Never exit non-zero — a hook
# failure must not disrupt the session.
#
# The envelope MUST be nested:
#   {"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"..."}}
# A bare top-level {"additionalContext": "..."} is accepted by the host and then
# silently discarded — measured 2026-07-27 against Claude Code, flat vs nested,
# with file tools disabled so context was the only possible source: the nested
# codeword came back, the flat one was absent. This file previously emitted the
# flat form, which is why the staleness nudge had never actually reached a model
# and overlays drifted unnoticed (decision-0035 expected drift to be visible).

root="${CLAUDE_PROJECT_DIR:-.}"
plugin="${CLAUDE_PLUGIN_ROOT:-/nonexistent}"
ref="$plugin/reference/version"
current="$(head -n1 "$ref" 2>/dev/null | tr -d '[:space:]')"

# Escape stdin as a JSON string body (no surrounding quotes). Newlines become \n,
# so the whole payload rides on the single line the output contract wants. UTF-8
# passes through untouched — it is valid inside a JSON string.
json_escape() {
  # Neutralise C0 control characters first. Anything below 0x20 other than tab
  # (escaped below), newline (the line separator) and carriage return (escaped
  # below) would make the JSON string invalid — a form feed in rules.toml did
  # exactly that. A config file has no business carrying one, so replace rather
  # than emit nothing: a stray byte degrades the payload, not the session.
  tr '\001-\010\013\014\016-\037' '    ' | awk '
    BEGIN { ORS = "" }
    {
      gsub(/\\/, "\\\\")
      gsub(/"/, "\\\"")
      gsub(/\t/, "\\t")
      gsub(/\r/, "\\r")
      if (NR > 1) printf "\\n"
      printf "%s", $0
    }
  '
}

# Emit one nested SessionStart envelope, escaping the message body. Every
# emission goes through here: interpolating untrusted text into the JSON
# directly is what this file's own headline bug was.
emit() {
  printf '{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"%s"}}\n' \
    "$(printf '%s' "$1" | json_escape)"
}

# ---------------------------------------------------------------------- path A
# The `.trellis/internal/` DIRECTORY decides the mode, not the stamp inside it.
# A half-deleted overlay is a broken vendored install, not a config-only project,
# and must not silently become one — path B would then inject alongside whatever
# vendored prose survived.
internal="$root/.trellis/internal"
ver="$internal/version"
if [ -d "$internal" ]; then
  if [ ! -f "$ver" ]; then
    emit "TRELLIS_RULES_NOT_LOADED — this project has a .trellis/internal/ directory but no version stamp, so its vendored overlay is incomplete. The hook will not inject rules over a broken overlay, and cannot tell which rules the surviving files represent. Tell the user before doing substantive work; /trellis:setup can migrate this project onto plugin-delivered rules."
    exit 0
  fi
  # A current stamp is not proof the overlay can still load. If a payload file
  # has been deleted, the import transport is broken and the stamp says nothing
  # about it — checking the stamp alone left that project silently ungoverned.
  for f in trellis.md rules.md; do
    if [ ! -s "$internal/$f" ]; then
      emit "TRELLIS_RULES_NOT_LOADED — this project's vendored overlay is incomplete: .trellis/internal/$f is missing or empty, so the managed block's imports cannot load the rules. The stamp is intact, which is why nothing else flagged this. Run /trellis:setup to migrate onto plugin-delivered rules, and tell the user before doing substantive work."
      exit 0
    fi
  done

  overlay="$(head -n1 "$ver" 2>/dev/null | tr -d '[:space:]')"
  [ -n "$overlay" ] || exit 0                     # empty stamp → nothing to compare
  [ -n "$current" ] || exit 0                     # can't read the installed payload → silent
  if [ "$overlay" != "$current" ]; then
    emit "Trellis overlay may be stale: this project's .trellis/internal/version stamp is $overlay, but the installed Trellis plugin ships $current. This project still carries a vendored overlay, which the plugin no longer writes or refreshes. To move it onto plugin-delivered rules, run /trellis:setup and accept the migration — it removes .trellis/internal/ and the managed block and keeps your .trellis/rules.toml rows. Until then this session is governed by the vendored copy."
  fi
  exit 0
fi

legacy="$root/.trellis/version"
if [ -f "$legacy" ]; then
  overlay="$(head -n1 "$legacy" 2>/dev/null | tr -d '[:space:]')"
  [ -n "$overlay" ] || exit 0                     # empty stamp → nothing to compare
  [ -n "$current" ] || exit 0                     # can't read the installed payload → silent
  emit "Trellis overlay predates the .trellis/internal/ layout (decision-0051): its stamp sits at the legacy path .trellis/version ($overlay; the installed plugin ships $current). Run /trellis:setup and accept the migration — it removes the legacy overlay and keeps your .trellis/rules.toml rows."
  exit 0
fi

# ---------------------------------------------------------------------- path C
# The curl install path renders `.claude/rules/trellis.md`, which Claude Code
# loads at launch by itself (decision-0068 D1). If the plugin is also present,
# injecting here would deliver the same rules a SECOND time — measured, not
# predicted: both present puts the rule bodies in the project-instructions block
# and in additionalContext at once.
#
# The discriminator is the FILE, not the directory. `.claude/rules/` is a shared
# directory any project may fill with unrelated rules; only `trellis.md` inside
# it means Trellis is already delivered. This is the mirror of path A, where the
# `.trellis/internal/` DIRECTORY is the artifact.
#
# Placed AFTER path A on purpose. A project migrating off a vendored overlay can
# hold both artifacts at once, and path A's staleness nudge is the only signal it
# gets — decision-0035's floor is that drift is made visible, not silent. Move
# this block above path A and that consumer goes quiet.
rendered="$root/.claude/rules/trellis.md"
# -s, not -f. A zero-byte or truncated file would otherwise silence this
# hook while governing nothing — path A already carries this lesson at the
# completeness gate above ("checking the stamp alone left that project
# silently ungoverned"), and path C did not inherit it until review said so.
if [ -s "$rendered" ]; then
  emit "Trellis rules are already loaded from .claude/rules/trellis.md (the curl install path), so this hook injected nothing — delivering them here too would put the same rules in context twice. That file and .trellis/rules.toml govern this session. To move onto plugin-delivered rules instead, delete .claude/rules/trellis.md."
  exit 0
fi

# ---------------------------------------------------------------------- path B
toml="$root/.trellis/rules.toml"
[ -f "$toml" ] || exit 0                          # not a Trellis project → silent

# Posture selects the header, exactly as the import channel does.
# Both TOML string forms: "firm" and 'firm' are equally valid, and the Codex
# hook's parser accepts both, so matching only double quotes served the wrong
# posture to a firm project without saying so.
strictness="$(awk '
  /^[[:space:]]*strictness[[:space:]]*=/ {
    if (match($0, /"[^"]*"/) || match($0, /\x27[^\x27]*\x27/)) {
      print substr($0, RSTART + 1, RLENGTH - 2); exit
    }
  }' "$toml" 2>/dev/null)"
case "$strictness" in
  firm) header="$plugin/reference/trellis-a.md" ;;
  *)    header="$plugin/reference/trellis-b.md" ;;
esac
rules="$plugin/reference/rules.md"

# Fail loudly rather than govern silently on a partial payload. A hook cannot
# report that it never ran, but it can report that it ran and could not deliver.
if [ ! -f "$header" ] || [ ! -f "$rules" ]; then
  emit "TRELLIS_RULES_NOT_LOADED — the Trellis plugin hook ran but could not read its own rules payload (looked for $header and $rules). This project is configured for Trellis: .trellis/rules.toml is present, but the session is running ungoverned. Tell the user before doing substantive work."
  exit 0
fi

# Validate the rows before injecting them. The Codex hook has always done this
# (parseRulesToml against a known slug list); the Claude hook did not, so a
# truncated or hand-broken rules.toml was injected verbatim and the session ran
# on a config nobody checked. The slugs the payload actually ships are the
# authority: each must have exactly one row, and no row may name anything else.
slug_report="$(
  awk '
    FNR == NR {
      # Rule slugs as rules.md declares them: a trailing `slug` on a rule line.
      if (match($0, /`(inv|floor)-[a-z-]+`[[:space:]]*$/)) {
        s = substr($0, RSTART + 1, RLENGTH - 2)
        sub(/`[[:space:]]*$/, "", s)
        want[s] = 1
      }
      next
    }
    /^[[:space:]]*(inv|floor)-[a-z-]+[[:space:]]*=/ {
      row = $1
      sub(/[^a-z-].*$/, "", row)
      if (row in seen) { dup = dup " " row }
      seen[row] = 1
      if (!(row in want)) { unknown = unknown " " row }
    }
    END {
      for (s in want) if (!(s in seen)) missing = missing " " s
      if (length(want) == 0) print "no-slugs-in-payload"
      else if (missing != "") print "missing:" missing
      else if (unknown != "") print "unknown:" unknown
      else if (dup != "") print "duplicate:" dup
      else print "ok"
    }
  ' "$rules" "$toml"
)"
if [ "$slug_report" != "ok" ]; then
  emit "TRELLIS_RULES_NOT_LOADED — this project's .trellis/rules.toml does not match the rules the installed plugin ships ($slug_report). Nothing was injected, because a partial or unknown row set cannot be applied honestly. Run /trellis:setup to reapply a preset, and tell the user before doing substantive work."
  exit 0
fi

# The header carries `@rules.md`, an import the hook resolves itself, and a
# pointer at the vendored invariants path, which does not exist in this mode.
# Repointing it at the plugin keeps the trigger-read affordance and cannot go
# stale, because it names the payload this session is actually running.
payload="$(
  awk -v rules="$rules" -v inv="$plugin/reference/invariants.md" '
    BEGIN {
      # awk expands `&` in a gsub replacement to the matched text, and `\` escapes
      # there too, so a plugin root under e.g. R&D silently corrupted the pointer.
      gsub(/[\\&]/, "\\\\&", inv)
    }
    /^@rules\.md[[:space:]]*$/ { while ((getline line < rules) > 0) print line; next }
    { gsub(/`\.trellis\/internal\/invariants\.md`/, "`" inv "`"); print }
  ' "$header"
  printf '\n## Project rule activation\n\n'
  printf 'Rows from this project'"'"'s .trellis/rules.toml. Apply a rule only when its row says active = true; the two floor rules apply regardless of their row.\n\n'
  cat "$toml"
  printf '\nDelivered by the Trellis plugin (%s). No overlay is vendored in this project.\n' "$current"
)"

# A bounded payload, like the Codex hook's MAX_CONTEXT_BYTES. Without this a
# runaway rules.toml becomes a multi-megabyte injection: measured, a 5 MB file
# produced 4.8 MB of valid JSON and exit 0, which quietly consumes the session's
# context instead of failing. Refuse loudly instead.
limit=32768
size=$(printf '%s' "$payload" | wc -c | tr -d '[:space:]')
if [ "$size" -gt "$limit" ]; then
  emit "TRELLIS_RULES_NOT_LOADED — the assembled Trellis rules are ${size} bytes, over the ${limit}-byte injection budget, so nothing was injected. This usually means .trellis/rules.toml has grown far beyond a row list. Tell the user before doing substantive work."
  exit 0
fi

emit "$payload"
exit 0
