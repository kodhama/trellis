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
#      the nudge carries the manual migration steps (decision-0072 retired
#      the setup skill, which used to be the vehicle). With the status command
#      retired (decision-0043), this hook is the only drift surface (decision-0035: drift is made
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
#      (decision-0068 D10). The guard is `-f` plus an ORDERED validation of the
#      file's generated structure, done inside the branch — not a size check.
#      Two earlier designs (-f alone, then -s) each let an incomplete file
#      silence this hook while governing nothing.
#
# `.trellis/rules.toml` was the opt-in signal for path B, and a project with none
# of the three used to get NOTHING. decision-0070 changed that, and this comment
# said the opposite until it was corrected. What such a project gets now depends
# on where the plugin lives: vendored under <repo>/.claude/skills/ means this
# project adopted Trellis, so the shipped defaults apply; anywhere else means the
# project is told once and governed by nothing until it answers. The
# never-BY-SURPRISE half stands; the never-governed half does not. Path C is
# still the one exception — it fires on its own artifact, with or without rules.toml, because
# that file is itself proof the project adopted Trellis.
#
# The paths are mutually exclusive, and where they cannot be — a project holding
# BOTH a static overlay and a rendered file — the coexistence branch reports it
# rather than pretending otherwise. An earlier version of this note claimed
# exclusivity "by construction", which was false for exactly that state.
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

# decision-0070 D5, and it runs BEFORE every delivery path — but AFTER emit() is
# defined, since a shell function called above its definition simply is not there.
#
# An explicit refusal outranks every default AND every other branch. A file
# holding `governed = false` is a project saying, in its own diff, that Trellis
# does not govern here.
#
# NOT GOVERNED MEANS NOT GOVERNED — the two floor- rules go too. The floors are a
# floor on CONFIGURATION, not on adoption: they stop a row dialling a rule to
# zero while the project is governed. They are not a claim on a project that
# declined. An intermediate version of this branch delivered the floors anyway,
# reasoning from "the only settings that never dial to zero"; that read a
# within-governance guarantee as a without-governance one.
#
# So the hook reads this file for exactly one thing — this key — and injects
# nothing when it is set. It never reads the rows; those are live, editable, and
# read by the model on demand, which is the behaviour this design is for.
bom="$(printf '\357\273\277')"
# KNOWN, NARROWED DIVERGENCE. A misplaced `governed = false` UNDER `[rules]` is
# not a top-level key, so neither host opts out — but they then differ: this hook
# ignores the stray line and governs normally (14 rules), while codex-context.mjs
# rejects the file as invalid-rules, because its parser validates every row shape
# and this one does not. Both fail SAFE — neither silently disables anything,
# which was the defect — but Codex is louder. Aligning them means teaching this
# awk slug check to reject unknown row shapes, which is a larger change than the
# bug warrants; recorded here so the next person sees it as known rather than
# discovering it as new.
# Normalise ONCE, then ask two questions of the result. Doing it in one pass is
# the point: each of the previous four rounds fixed a matcher in one place and
# left the other host, or an earlier stage, unfixed. The BOM strip must precede
# the table-header scan — a file beginning "<BOM>[rules]" was not recognised as
# having a header at all, so the whole file was searched and a `governed = false`
# under [rules] opted the project out, restoring the exact defect the header scan
# was added to prevent.
governed_head="$(sed "1s/^$bom//" "$root/.trellis/rules.toml" 2>/dev/null | sed -n '/^[[:space:]]*\[/q;p')"
# Exactly ONE top-level assignment counts. Two — `governed = false` and
# `governed = true` — is a malformed file, and opting out on whichever came first
# would honour a config the parser would reject.
# LC_ALL=C, so [[:space:]] is ASCII and nothing else. Without it the class is
# locale-dependent AND disagrees with itself across machines: Codex measured the
# shell REJECTING an NBSP-indented opt-out under C.UTF-8, while on macOS the same
# expression MATCHES it and silently opts the project out. Either way the two
# hosts disagreed about whether a project was governed, and which way round
# depended on where you ran it. Pinning the JS class alone could not fix that;
# the shell had to stop asking the locale.
governed_n="$(printf '%s\n' "$governed_head" | LC_ALL=C grep -cE '^[[:space:]]*governed[[:space:]]*=' 2>/dev/null || true)"
if [ -f "$root/.trellis/rules.toml" ] && [ "${governed_n:-0}" -eq 1 ] &&
   printf '%s\n' "$governed_head" | LC_ALL=C grep -qE '^[[:space:]]*governed[[:space:]]*=[[:space:]]*false[[:space:]]*(#.*)?$' 2>/dev/null; then
  # One thing the hook cannot do is UN-load. On the curl path the host reads
  # .claude/rules/trellis.md at launch, before any hook runs, so by now those
  # rules are already in context and no amount of silence removes them. Injecting
  # nothing would leave the project governed by a file it has declined.
  #
  # The only lever left is to override what was already read. This is second-best
  # and says so: deleting the file (or /trellis:remove) is the real fix, and the
  # message names it.
  # Every shape the host may ALREADY have loaded, not just the rendered file. A
  # vendored overlay reaches context through the managed block's imports exactly
  # as .claude/rules/trellis.md reaches it through the rules loader, and this
  # branch used to check only the latter — so a declining project carrying an
  # overlay got total silence while its rules were live. Same defect, other
  # transport.
  if [ -f "$root/.claude/rules/trellis.md" ] || [ -d "$root/.trellis/internal" ] || [ -f "$root/.trellis/trellis.md" ]; then
    emit "TRELLIS_NOT_GOVERNING — this project declares governed = false in .trellis/rules.toml, so Trellis does not govern here: no rule applies, including the two floor- rules. DISREGARD any Trellis rules already loaded this session — from .claude/rules/trellis.md, or from a managed block importing .trellis/internal/. Those are read by the host at launch, before any hook runs, so they could not be withheld. To stop them being loaded at all, run /trellis:remove."
  fi
  exit 0
fi

# ---------------------------------------------------------------------- path A
# The `.trellis/internal/` DIRECTORY decides the mode, not the stamp inside it.
# A half-deleted overlay is a broken vendored install, not a config-only project,
# and must not silently become one — path B would then inject alongside whatever
# vendored prose survived.
internal="$root/.trellis/internal"
ver="$internal/version"

# BOTH static paths at once. install.sh refuses to create this state, but it can
# arrive the other way round — a branch checkout or a collaborator's commit
# landing an overlay into a project that already had the rendered file. Path A
# would then exit first and, with a CURRENT stamp, emit nothing at all: the
# session receives the rules twice in silence, while the installer warns loudly
# about the very same state. Checked before path A for that reason.
static_overlay=""
overlay_paths=""
[ -d "$internal" ] && { static_overlay=".trellis/internal/ overlay"; overlay_paths=".trellis/internal/"; }
# The remedy must name the shape that is actually present. It used to be
# hard-coded to .trellis/internal/, so a flat-layout project was told to delete
# a directory it does not have — the alarm then fired every session forever,
# because this branch keys on file existence and following the advice removed
# nothing.
[ -z "$static_overlay" ] && [ -f "$root/.trellis/trellis.md" ] && { static_overlay="legacy flat .trellis/ overlay"; overlay_paths=".trellis/trellis.md (and .trellis/version if present)"; }
if [ -n "$static_overlay" ] && [ -f "$root/.claude/rules/trellis.md" ]; then
  emit "TRELLIS_RULES_LOADED_TWICE — this project has BOTH a vendored $static_overlay (imported by its managed block) and a rendered .claude/rules/trellis.md. Both are loaded by the host before any hook runs, so the rules are in context TWICE right now and no hook can undo it. Remove one: delete .claude/rules/trellis.md to keep the overlay, or delete $overlay_paths and the managed block from this project's instructions file, keeping .trellis/rules.toml, to keep the rendered file. Show the user the exact paths you would delete and get explicit confirmation before deleting anything (floor-intent-gate): this hook advises, it never authorises a deletion, and the files are tracked. Tell the user before doing substantive work."
  exit 0
fi

if [ -d "$internal" ]; then
  if [ ! -f "$ver" ]; then
    emit "TRELLIS_RULES_NOT_LOADED — this project has a .trellis/internal/ directory but no version stamp, so its vendored overlay is incomplete. The hook will not inject rules over a broken overlay, and cannot tell which rules the surviving files represent. Tell the user before doing substantive work. To migrate this project onto plugin-delivered rules, delete .trellis/internal/ and the managed block from this project's instructions file, keeping .trellis/rules.toml. Show the user the exact paths you would delete and get explicit confirmation before deleting anything (floor-intent-gate): this hook advises, it never authorises a deletion, and the files are tracked."
    exit 0
  fi
  # A current stamp is not proof the overlay can still load. If a payload file
  # has been deleted, the import transport is broken and the stamp says nothing
  # about it — checking the stamp alone left that project silently ungoverned.
  for f in trellis.md rules.md; do
    if [ ! -s "$internal/$f" ]; then
      emit "TRELLIS_RULES_NOT_LOADED — this project's vendored overlay is incomplete: .trellis/internal/$f is missing or empty, so the managed block's imports cannot load the rules. The stamp is intact, which is why nothing else flagged this. To migrate onto plugin-delivered rules, delete .trellis/internal/ and the managed block from this project's instructions file, keeping .trellis/rules.toml. Show the user the exact paths you would delete and get explicit confirmation before deleting anything (floor-intent-gate): this hook advises, it never authorises a deletion, and the files are tracked. Tell the user before doing substantive work."
      exit 0
    fi
  done

  overlay="$(head -n1 "$ver" 2>/dev/null | tr -d '[:space:]')"
  [ -n "$overlay" ] || exit 0                     # empty stamp → nothing to compare
  [ -n "$current" ] || exit 0                     # can't read the installed payload → silent
  if [ "$overlay" != "$current" ]; then
    emit "Trellis overlay may be stale: this project's .trellis/internal/version stamp is $overlay, but the installed Trellis plugin ships $current. This project still carries a vendored overlay, which the plugin no longer writes or refreshes. To move it onto plugin-delivered rules, delete .trellis/internal/ and the managed block from this project's instructions file, keeping .trellis/rules.toml rows. Show the user the exact paths you would delete and get explicit confirmation before deleting anything (floor-intent-gate): this hook advises, it never authorises a deletion, and the files are tracked. Until then this session is governed by the vendored copy."
  fi
  exit 0
fi

legacy="$root/.trellis/version"
if [ -f "$legacy" ]; then
  overlay="$(head -n1 "$legacy" 2>/dev/null | tr -d '[:space:]')"
  [ -n "$overlay" ] || exit 0                     # empty stamp → nothing to compare
  [ -n "$current" ] || exit 0                     # can't read the installed payload → silent
  emit "Trellis overlay predates the .trellis/internal/ layout (decision-0051): its stamp sits at the legacy path .trellis/version ($overlay; the installed plugin ships $current). To migrate, delete the legacy overlay — .trellis/version, .trellis/trellis.md and .trellis/internal/ if present, plus the managed block from this project's instructions file — keeping your .trellis/rules.toml rows. An overlay this old may predate .trellis/rules.toml entirely; if there is none, copy $plugin/reference/rules-b.toml to $root/.trellis/rules.toml. Show the user the exact paths you would delete and get explicit confirmation before deleting anything (floor-intent-gate): this hook advises, it never authorises a deletion, and the files are tracked."
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
# Placed after path A. NOTE: the coexistence branch above is what now protects a
# migrating consumer — moving this block between that branch and path A is
# behaviour-preserving. An earlier version of this comment claimed the ordering
# itself was load-bearing; it was, before the coexistence branch existed, and was
# not retracted when that stopped being true.
rendered="$root/.claude/rules/trellis.md"
# Existence is not delivery, and NON-EMPTY is not delivery either. `-f` let a
# zero-byte file silence this hook; `-s` still let a one-byte file do it —
# reproduced both times, and both leave a session ungoverned while the stand-down
# message below claims the rules are loaded.
#
# The guard keys on the same terminal sentinel `rules.md` ships and the Codex
# hook already validates (`codex-context.mjs` requires exactly one). A file
# carrying it has the whole rules body by construction, because the sentinel is
# the LAST line of that body — a truncation cannot keep the end and lose the
# middle. This is path A's completeness gate applied to path C's artifact:
# "checking the stamp alone left that project silently ungoverned".
# The FILE'S EXISTENCE claims this path — validation happens inside, never as
# part of the condition. An earlier version made completeness part of the guard,
# so an incomplete file fell through to path B: with no rules.toml path B exited
# SILENTLY, and once a .trellis/rules.toml later appeared, path B injected the full
# payload on top of the body the host had already loaded. Falling through was the
# double delivery this path exists to prevent, arriving one step later.
if [ -f "$rendered" ]; then
  # Standing down is not the end of this hook's duty. A rendered file written by
  # an OLDER installer, with a NEWER plugin now installed, would otherwise sit on
  # stale rule bytes forever: the newer plugin neither injects nor warns.
  # decision-0035's floor is that drift is made visible, not silent — path A has
  # carried that for the vendored overlay since decision-0043 rule 3, and path C
  # shipped without it until review said so.
  # Validated on MACHINE-OWNED markers plus actual content — never on prose.
  # Two earlier designs failed opposite ways: unordered substring greps accepted
  # a file whose footer was replaced by arbitrary lines containing the right
  # words; then ordered PROSE landmarks made any legitimate payload reword
  # produce a permanent false "not governed" warning on every fresh install,
  # with the suite green. Both were reported by review.
  #
  # FIVE landmarks, all owned by install.sh or the payload's generator, in order
  # — plus a content assertion that counts DISTINCT rule lines, not presence.
  #
  # Both refinements come from review finding the previous versions too weak: a
  # file with the whole footer deleted between the sentinel and the import passed
  # (so the footer got its own marker), and a 167-byte file carrying one line of
  # `inv-x` passed the "at least one slug" test (so the count is five, against a
  # payload that ships fourteen). A truncation severe enough to matter cannot keep
  # five distinct rule lines. Trailing CR and whitespace are
  # tolerated: this file is committed, and a collaborator on core.autocrlf=true
  # otherwise gets told a complete file is incomplete.
  incomplete="$(awk -v bom="$(printf '\357\273\277')" '
    { line = $0; sub(/[ \t\r]+$/, "", line) }
    # A UTF-8 BOM on line 1 made the opening marker compare unequal, and the hook
    # then told a fully-governed project it was NOT governed. Same harm, and the
    # same population, as the trailing-CR tolerance directly above: an editor on
    # a Windows-default checkout rewrites the encoding, and nothing in the
    # trellis delivery chain ever writes a BOM. The host loads it either way.
    #
    # The bytes arrive via -v rather than as a regex escape. Written as
    # /^\357\273\277/ the first attempt was INERT -- octal escapes in a regex
    # literal are not portable across awks, and it silently matched nothing while
    # looking correct. substr is exact and needs no escape rules.
    # (No apostrophes in here: this whole awk program is single-quoted.)
    NR == 1 && substr(line, 1, length(bom)) == bom { line = substr(line, length(bom) + 1) }
    stage == 0 && line == "<!-- trellis:rendered-begin -->"        { stage = 1; next }
    stage == 1 && line == "<!-- trellis:rules-loaded -->"          { stage = 2; next }
    stage == 2 && line == "<!-- trellis:rendered-footer -->"       { stage = 3; next }
    stage == 3 && line == "@../../.trellis/rules.toml"             { stage = 4; next }
    stage == 4 && line ~ /^<!-- trellis:rendered-from payload@[0-9a-f]+ -->$/ { stage = 5; next }
    stage >= 1 && line ~ /`(inv|floor)-[a-z-]+`/                   { rules[line] = 1 }
    END {
      n = 0; for (k in rules) n++
      if (stage == 0) print "opening marker"
      else if (stage == 1) print "rules body (no trellis:rules-loaded sentinel)"
      else if (n < 5) print "rule text (only " n " rule line(s) survived; the payload ships fourteen)"
      else if (stage == 2) print "fixed footer"
      else if (stage == 3) print "rule-activation import"
      else if (stage == 4) print "rendered-from stamp"
    }' "$rendered" 2>/dev/null)"
  if [ -n "$incomplete" ]; then
    emit "TRELLIS_RULES_NOT_LOADED — .claude/rules/trellis.md exists but is incomplete: its $incomplete is missing, so this project is NOT governed by the rules it appears to carry. This hook did not inject over it, because a half-written governing file and a full one are indistinguishable to a reader. Re-run install.sh, or delete the file to move onto plugin-delivered rules. Tell the user before doing substantive work. Show the user the exact paths you would delete and get explicit confirmation before deleting anything (floor-intent-gate)."
    exit 0
  fi
  # The awk above stops at the FIRST stamp line that follows the import, so this
  # must not take the first line in the FILE — a decoy stamp above the real
  # content would otherwise pass validation and then be reported STALE against
  # the wrong value. Anchored and hex-required to match the awk's own pattern
  # rather than being looser on both ends.
  rendered_stamp="$(sed -n 's/^<!-- trellis:rendered-from \(payload@[0-9a-f][0-9a-f]*\) -->$/\1/p' "$rendered" 2>/dev/null | tail -n1)"
  if [ -z "$rendered_stamp" ]; then
    emit "TRELLIS_RULES_NOT_LOADED — .claude/rules/trellis.md exists but is incomplete: it carries no trellis:rendered-from stamp, which install.sh writes as its last line, so the file was truncated and its rule activation rows are missing. This hook did not inject over it, because a half-written governing file and a full one are indistinguishable to the reader. Re-run install.sh, or delete the file to move onto plugin-delivered rules. Tell the user before doing substantive work. Show the user the exact paths you would delete and get explicit confirmation before deleting anything (floor-intent-gate)."
    exit 0
  fi
  if [ -n "$current" ] && [ "$rendered_stamp" != "$current" ]; then
    emit "Trellis rules come from .claude/rules/trellis.md (the curl install path), and that file is STALE: it was rendered from $rendered_stamp, but the installed plugin ships $current. This hook injected nothing — the rendered file governs this session and it is out of date. Re-run install.sh to refresh it, or delete it to move onto plugin-delivered rules. Show the user the exact paths you would delete and get explicit confirmation before deleting anything (floor-intent-gate)."
    exit 0
  fi
  emit "Trellis rules are already loaded from .claude/rules/trellis.md (the curl install path), so this hook injected nothing — delivering them here too would put the same rules in context twice. That file and .trellis/rules.toml govern this session. To move onto plugin-delivered rules instead, delete .claude/rules/trellis.md — show the user the exact paths you would delete and get explicit confirmation before deleting anything (floor-intent-gate)."
  exit 0
fi

# A legacy FLAT overlay reaches here: path A keys on the .trellis/internal/
# DIRECTORY, which this layout does not have, so nothing above catches it. Injecting would deliver the
# rules a second time on top of that chain. The installer already refuses this
# shape; the hook did not know it existed.
if [ -f "$root/.trellis/trellis.md" ]; then
  emit "TRELLIS_RULES_NOT_LOADED — this project carries a legacy flat .trellis/trellis.md overlay, which its managed block imports directly. This hook injected nothing: doing so would deliver the same rules twice. To move onto plugin-delivered rules, delete .trellis/trellis.md and the managed block from this project's instructions file, keeping .trellis/rules.toml. Show the user the exact paths you would delete and get explicit confirmation before deleting anything (floor-intent-gate): this hook advises, it never authorises a deletion, and the files are tracked. Tell the user before doing substantive work."
  exit 0
fi

# ---------------------------------------------------------------------- path B
toml="$root/.trellis/rules.toml"

# decision-0070. Adoption is the consent act, and every path has one; what a
# missing rules.toml means now depends on WHICH path installed this plugin.
#
if [ ! -f "$toml" ]; then
  # D6. Scope by containment: a project-scoped plugin is vendored INSIDE the
  # repository, a user-scoped one lives under the user's home. Resolved with pwd
  # -P so a symlinked checkout or a marketplace cache cannot fake either answer.
  # When it cannot tell, it falls through to the announcement — the failure mode
  # is one extra paragraph, never governing a project that did not expect it.
  plugin_real="$(cd "$plugin" 2>/dev/null && pwd -P)" || plugin_real=""
  root_real="$(cd "$root" 2>/dev/null && pwd -P)" || root_real=""
  scoped_to_project=no
  # NOT merely "inside the project" — that was wrong, and wrong in the direction
  # that governs. A dotfiles repo rooted at $HOME contains
  # ~/.claude/plugins/cache/..., the USER-scope location, so containment alone
  # reported project scope and delivered all fourteen rules with no announcement.
  # Measured before this fix: 12 slugs, zero announcements — exactly the shape
  # decision-0070 D6 promises never happens.
  #
  # Project scope has ONE location: the vendored bundle under
  # <repo>/.claude/skills/, where install.sh writes it and where a project-scope
  # marketplace install lands. Anything else is not this project's copy, whatever
  # directory it happens to sit beneath.
  #
  # And it cannot be the home directory. `install.sh --scope personal` vendors to
  # $HOME/.claude/skills/trellis, which is byte-identical to the project-scope
  # location when the project IS $HOME — a dotfiles repo, or simply an unset
  # CLAUDE_PROJECT_DIR while sitting in $HOME, since root falls back to `.`.
  # Containment can never separate those two, so the only safe answer is to stop
  # claiming it can: when the project root is $HOME, treat it as unadopted and
  # ask. Measured before this guard: a personal install delivered 12 rules with
  # no announcement. The first fix for the dotfiles case narrowed the path and
  # missed this, because the path is the same path.
  home_real="$(cd "${HOME:-/nonexistent}" 2>/dev/null && pwd -P)" || home_real=""
  case "$plugin_real/" in
    "$root_real"/.claude/skills/*)
      [ -n "$root_real" ] && [ "$root_real" != "$home_real" ] && scoped_to_project=yes
      ;;
  esac

  if [ "$scoped_to_project" = yes ]; then
    # D3. The bundle sits in this repository, so this project adopted Trellis —
    # visibly, greppably, and revocably by deleting it. Absent rows therefore
    # mean the standard set, not none. Rather than invent a second activation
    # semantics, point at the shipped preset and let every check below run
    # unchanged: same slugs, 14/14 active, strictness "adaptive" (posture B).
    toml="$plugin/reference/rules-b.toml"
    [ -f "$toml" ] || exit 0
    rows_are_default=yes
  else
    # D4. A user-wide install is a broad choice, and this says so in the project
    # it is about to affect rather than assuming consent it never asked for.
    # Announce, inject NO rules on this turn ("will be", not "is"), and ask for
    # the negative action explicitly so silence cannot read as refusal.
    emit "TRELLIS_NOT_YET_GOVERNING — the Trellis plugin is installed outside this project (user scope, or a location this hook cannot place), so it applies to every project opened here, and $root has no .trellis/rules.toml. Tell the user, in your own words and before doing substantive work: \"Trellis is installed for your user account, so this repo will be governed by it — 14 rules, followed by default and deviations said out loud. Do you want to disable that for this repo?\" If they want it DISABLED, write .trellis/rules.toml containing exactly the line: governed = false — and nothing else. If they ACCEPT, copy $plugin/reference/rules-b.toml to $root/.trellis/rules.toml so the choice persists — without that file this same announcement repeats every session and the project is never governed. (That file is theirs to edit afterwards: strictness = \"firm\" for the by-the-book posture, active = false on a row to turn a rule off.) Inject and follow no Trellis rules this turn: none are active yet."
    exit 0
  fi
fi

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
  emit "TRELLIS_RULES_NOT_LOADED — this project's .trellis/rules.toml does not match the rules the installed plugin ships ($slug_report). Nothing was injected, because a partial or unknown row set cannot be applied honestly. Repair the rows WITHOUT discarding the project's choices: add the missing slugs, drop the unknown ones, and leave strictness and every existing active value exactly as they are. If you reseed from a preset instead — $plugin/reference/rules-a.toml for strictness = \"firm\", rules-b.toml for adaptive — show the user a diff and get explicit confirmation FIRST: a reseed resets every row they chose. Tell the user before doing substantive work."
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
  if [ "${rows_are_default:-no}" = yes ]; then
    printf 'Rows: this project has no .trellis/rules.toml, so these are the shipped defaults — every rule active, adaptive posture (decision-0070). Write .trellis/rules.toml to change them. Apply a rule only when its row says active = true; the two floor rules apply regardless of their row.\n\n'
  else
    printf 'Rows from this project'"'"'s .trellis/rules.toml. Apply a rule only when its row says active = true; the two floor rules apply regardless of their row.\n\n'
  fi
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
