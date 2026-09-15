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
#      own payload instead of read from vendored copies: the header and the
#      rules, then the project's rule activation section, where this hook
#      classifies .trellis/rules.toml and names the rules it switches off
#      (TRL-97). The tested wording stays the shipped wording (decision-0053).
#      The one edit is
#      repointing the invariants path at the plugin, which is where the file
#      actually is in this mode, and which therefore cannot go stale. It can
#      still be ABSENT from a half-installed payload, and the pointer moves
#      anyway: what a missing target earns is a report at the end of the
#      injected payload, not a refusal (TRL-70, decision-0093 rule 1).
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
# project adopted Trellis, so every rule applies; anywhere else means the
# project is told — every session until it answers — and governed by nothing
# meanwhile. The never-BY-SURPRISE half stands. The never-governed half fails
# only for the vendored-bundle case above, where the bundle IS the adoption act;
# under a user-scope install it holds, because an unanswered announcement never
# adopts (decision-0077). An earlier version of this comment said "told once"
# and that the never-governed half does not stand — both were the 0070 D4
# reading this code never implemented. Path C is
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

# The installed plugin's own version stamp is read further down, after the
# payload gateway is defined — a shell function called above its definition
# simply is not there, and that read now goes through the gateway like every
# other payload read in this file.

# Escape stdin as a JSON string body (no surrounding quotes). Newlines become \n,
# so the whole payload rides on the single line the output contract wants. UTF-8
# passes through untouched — it is valid inside a JSON string.
json_escape() {
  # Neutralise C0 control characters first. Anything below 0x20 other than tab
  # (escaped below), newline (the line separator) and carriage return (escaped
  # below) would make the JSON string invalid — a form feed in rules.toml did
  # exactly that. A config file has no business carrying one, so replace rather
  # than emit nothing: a stray byte degrades the payload, not the session.
  # Both tools run in the C locale so they work on bytes. In a UTF-8 locale a
  # byte that is not valid UTF-8, such as a Latin-1 comment in rules.toml, made
  # tr stop mid-stream, and everything after it (the rest of the file, every
  # warning, the footer) was silently cut from otherwise valid JSON.
  LC_ALL=C tr '\001-\010\013\014\016-\037' '    ' | LC_ALL=C awk '
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

# ======================================================== the payload gateway
# THE ONE PLACE THIS FILE OPENS A TRELLIS PAYLOAD FILE.
#
# A PAYLOAD file is one that ships in the Trellis bundle — read from the
# installed plugin, or from a vendored copy inside the consuming repository.
# Its absence or corruption is always a broken install, never a legitimate
# project state. That is what separates it from a PROJECT file
# (.trellis/rules.toml, CLAUDE.md, .claude/rules/trellis.md), where absent and
# empty are legitimate states with defined meanings. The two classes need
# opposite defaults, which is why one gateway cannot serve both and why this
# one is named for the class it serves.
#
# Fifteen defects, counted 2026-09-03, shared one shape (decision-0087 states
# the count as its own measurement: decision-0083 and decision-0084 record eight
# and several more without totalling them, and four were open when that count was
# taken). One shape: an
# absent, empty, truncated or unreadable payload input reached downstream
# logic, and the session ran ungoverned at exit 0 with nothing signalling a
# problem. Each was fixed where it was found. Almost none was found by the
# test suite or by reading the code — every one was found by a reviewer
# RUNNING the hook against a deliberately broken input. That recurrence is the
# finding (TRL-33), and this function is the answer to it: a payload read
# added to this file later is guarded by WHERE IT IS WRITTEN, not by whoever
# remembers. Two guards hold that: TestNoPayloadReadBypassesTheGateway, which
# fails if anything opens a payload path behind this function's back, and
# TestBrokenPayloadIsNeverSilent, which breaks every file the bundle ships,
# four ways, on both delivering paths, and refuses to accept silence.
#
# It CLASSIFIES; it does not judge. Some of those were the INVERSE defect — a
# guard that refused a HEALTHY payload: an unreadable comparison preset reported
# as payload incoherence, and a CRLF-terminated rules.md reported as truncated
# (that second one is recorded HERE, at the sentinel gate a few hundred lines
# down, and in TestTruncatedRulesMdIsRefusedByItsOwnTerminator -- not in either
# decision, which a corpus review of decision-0087 caught it being attributed
# to) — and a
# consumer who sees TRELLIS_RULES_NOT_LOADED with nothing wrong to fix is as
# badly served as one governed by a broken payload. So the four outcomes are
# reported and the CALL SITE decides what each one costs:
#
#   missing     the path is not there, or its symlink target is gone
#   unreadable  it exists and could not be opened — a permission mode, a stale
#               ACL, or a directory where a file must be
#   empty       it opened and yielded nothing
#   ok          content is in $payload_text
#
# missing and unreadable are told apart because their remedies differ (reinstall
# vs. fix the mode) — but NEITHER is silent anywhere, which is TRL-33's whole
# finding: on the vendored-defaults path the two were handled differently, an
# absent preset exiting silently while its unreadable sibling refused loudly,
# and nothing chose that.
#
# Returns 0 only for ok, so the shortest thing a caller can write —
# `payload_read "$f" || { emit "..."; exit 0; }` — is also the safe thing.
#
# The result comes back in GLOBALS rather than on stdout on purpose: a
# `$(payload_read ...)` capture would run the function in a SUBSHELL, and every
# status it set would be discarded at the closing paren.
payload_status=""
payload_why=""
payload_text=""
payload_read() {
  payload_text=""
  if [ ! -e "$1" ]; then
    payload_status=missing
    payload_why="is missing"
    return 1
  fi
  if [ ! -f "$1" ]; then
    payload_status=unreadable
    payload_why="is not a readable file — a directory or a device sits at that path"
    return 1
  fi
  # `-f` proves a file EXISTS, never that it can be READ, and the gap between
  # those two is where several of the eleven lived. Opening it is the only test
  # that settles it, so the read below is the check rather than a step after it.
  if ! payload_text="$(cat "$1" 2>/dev/null)"; then
    payload_text=""
    payload_status=unreadable
    payload_why="exists but could not be read — a permission mode, a stale ACL, or a symlink whose target is gone"
    return 1
  fi
  # Command substitution strips trailing newlines, so a file holding nothing but
  # blank lines lands here too — which is right: it is exactly as unusable as a
  # zero-byte one, and both shipped as separate defects.
  if [ -z "$payload_text" ]; then
    payload_status=empty
    payload_why="is empty"
    return 1
  fi
  payload_status=ok
  payload_why=""
  return 0
}

# ------------------------------------------ the installed payload's own stamp
# TRL-34. This was `current="$(head -n1 "$ref" 2>/dev/null | tr -d ...)"`, and
# an unreadable reference/version therefore yielded "" — after which three
# separate call sites exited silently or skipped their comparison, and the
# staleness warning this hook exists to produce was withheld with no signal at
# all. Measured on main: mode 000 on that file, a vendored overlay present,
# zero bytes of stdout at exit 0.
#
# $stamp_defect carries the reason forward so each of those three sites can say
# WHY it could not compare, rather than vanishing. Empty means the stamp is
# good.
ref="$plugin/reference/version"
current=""
stamp_defect=""
if payload_read "$ref"; then
  # THE WHOLE FILE IS VALIDATED, not just its first line. An earlier version of
  # this block read `head -n1 "$ref" | tr -d '[:space:]'`, which threw away
  # every byte after line 1 AND every space inside the stamp before any check
  # ran. Both shapes then reduced to a valid-looking stamp and were accepted as
  # authoritative — measured on a vendored-overlay project whose stamp matched,
  # zero bytes of stdout at exit 0 for each:
  #
  #   payload@<12 hex>\nGARBAGE     the garbage line never reached a check
  #   payload@046c 9109c663         the internal space was deleted, not rejected
  #
  # It was also a host divergence: codex-context.mjs validates the whole file
  # (/^payload@[0-9a-f]{12}\n?$/) and rejects both. Reported by review on
  # kodhama/trellis#262 and reproduced before this was written.
  #
  # WHAT IS FORGIVEN, and only this: carriage returns, trailing whitespace on a
  # line, and blank lines. A reference/version checked out on
  # core.autocrlf=true is a HEALTHY file, and refusing it would be the
  # over-correction this change set is as concerned with as the silence — the
  # same tolerance, for the same reason, that the rendered-file validator below
  # applies with `sub(/[ \t\r]+$/, "", line)`. Nothing else is forgiven: not
  # internal whitespace, not a second line of content, not a decoy stamp.
  #
  # More than one content line is reported separately from a malformed one,
  # because they are different corruptions and the message says which.
  current="$(printf '%s\n' "$payload_text" | awk '
    { line = $0; sub(/[ \t\r]+$/, "", line); if (line == "") next; n++; keep = line }
    END { if (n == 1) print keep; else if (n > 1) print "#multi" }
  ')"
  # Twelve `?` is an exact length test without needing a counting tool; the
  # nested case is the hex test. Both are shell built-ins, so this stays
  # binary-free. A TRUNCATED stamp is not a different version — it is an
  # unreadable one — and comparing it reports a perfectly healthy overlay as
  # STALE, which is a consumer told to migrate something already current.
  case "$current" in
    "#multi") stamp_defect="carries more than one line of content, where a payload stamp is exactly one"; current="" ;;
    payload@????????????)
      case "${current#payload@}" in
        *[!0-9a-f]*) stamp_defect="is not a Trellis payload stamp"; current="" ;;
      esac
      ;;
    *) stamp_defect="is not a Trellis payload stamp"; current="" ;;
  esac
else
  stamp_defect="$payload_why"
fi

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
# So this read is for exactly one thing — this key — and nothing is injected
# when it is set. It runs before path B classifies the rest of the file, so an
# opted-out file is never classified, whatever else it holds.
bom="$(printf '\357\273\277')"
# A misplaced `governed = false` UNDER `[rules]` is not a top-level key, so
# neither host opts out on it, and both warn that it does not (TRL-97).
# Normalise ONCE, then ask two questions of the result. Doing it in one pass is
# the point: each of the previous four rounds fixed a matcher in one place and
# left the other host, or an earlier stage, unfixed. The BOM strip must precede
# the table-header scan — a file beginning "<BOM>[rules]" was not recognised as
# having a header at all, so the whole file was searched and a `governed = false`
# under [rules] opted the project out, restoring the exact defect the header scan
# was added to prevent.
#
# REGULAR AND READABLE BEFORE IT IS OPENED (TRL-43). This is the first thing in
# the hook to open .trellis/rules.toml, and it used to open it before anything
# asked what the path was. A FIFO there blocks open(2) until a writer appears,
# so the sed never returned and the SessionStart hook never did either: the
# host either hung on it or killed it on its timeout, and the project ran
# ungoverned with no message. Same guard, same shape, as install.sh's copy of
# this read (#267). An unreadable or non-regular file leaves $governed_head
# empty, so it is never an opt-out: a path the hook cannot read cannot be a
# project's refusal. What it IS is decided where the rows are read (path B).
# The `-f` and `-r` tests are stat(2) and access(2); neither opens the path.
governed_head=""
if [ -f "$root/.trellis/rules.toml" ] && [ -r "$root/.trellis/rules.toml" ]; then
  governed_head="$(sed "1s/^$bom//" "$root/.trellis/rules.toml" 2>/dev/null | sed -n '/^[[:space:]]*\[/q;p')"
fi
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

# ------------------------------------------------------------------ the S4 probe
# decision-0073 D2: the inline managed-block shape (S4 in decision-0073 D1's
# closed set) is a column-0 `<!-- trellis:begin` marker in an instruction file —
# the rules body embedded between the markers, OR a dangling import whose
# overlay was deleted. The probe cannot tell those two apart, so every message
# it feeds below is written for both states and asserts neither as fact. It
# feeds three decision sites: the governed = false disregard, the coexistence
# check, and a refusal before path B.
#
# DELIBERATE TWO-FILE SUBSET (decision-0073 D1's per-component relevance
# clause, stated here where it is done): only CLAUDE.md and AGENTS.md are
# probed, while /trellis:remove recognises blocks in five instruction files.
# These two are the files the Claude host loads; refusing delivery over a block
# in GEMINI.md, .github/copilot-instructions.md or .clinerules — files this
# host never reads — would ungovern a Claude session for content that was never
# in it, the exact wrong-about-the-reader's-state class decision-0073 exists to
# end. Same subset, same reason, as install.sh's render-time probe.
#
# Also deliberate, same clause: this hook does NOT probe the M2 morph markers
# (.trellis/rollback, the trellis-pre-morph tag — decision-0073 D1's S6).
# decision-0073 D2's change-set for this hook is the inline probe alone; a
# morphed project's delivery is its own rewritten files, its rules.toml still
# governs activation, and path B runs unchanged. The S6 fixture in
# cli/plugin_hook_test.go pins that by name, so any change to it is a
# decision, not a drive-by.
#
# Column-0 anchor with an optional UTF-8 BOM, mirroring install.sh's probe:
# prose that names the marker mid-sentence must not match, and a BOM'd block at
# line 1 must (an editor on a Windows-default checkout rewrites the encoding;
# the fail-open direction is a real block escaping the probe). The literal
# `trellis:begin` does not match `trellis:codex-bootstrap:begin` — different
# text after `trellis:` — so the Codex receipt alone never trips this.
# EVERY match, not the first. A block in CLAUDE.md AND one in AGENTS.md is a
# legitimate multi-file state — skills/remove/SKILL.md says so in terms ("a
# legitimate multi-file state — remove each; it is not a duplicate") — and the
# host loads both files, so both blocks are in context. Breaking at the first
# match named one file in every message below: the refusal's remedy then left
# the second block live, the project stayed in the refused state forever, and
# the message was wrong about the state the reader was actually in — the exact
# class decision-0073 exists to end, in the code that closed it. $inline_file
# holds the first match (kept for messages that name one file); $inline_files
# holds all of them, space-separated, and is what the remedies name.
# AGENTS.md reaches a CLAUDE session only through a CLAUDE.md import
# (decision-0057: "Claude Code can import AGENTS.md from CLAUDE.md; Codex
# discovers AGENTS.md directly"). Probing it unconditionally refused delivery
# over a block THIS host never read — a documented mixed-host layout, where a
# Codex-facing inline block in AGENTS.md left an otherwise plugin-governed
# Claude session ungoverned while the refusal claimed the block was loaded.
# That is the wrong-about-the-reader's-state class decision-0073 exists to end,
# and D2's own recorded reason for the two-file subset is exactly this test:
# refuse only over files this host actually loads.
# ANCHORED, not a substring: the adapter contract is "one standalone
# @AGENTS.md line" (decision-0057). An unanchored match set this to yes when
# CLAUDE.md merely MENTIONED the import in prose or inside a fenced example —
# documentation about the import read as the import itself, recreating the very
# mixed-host regression the gate exists to prevent. Same class as every other
# defect on this change: a check that matched text NEAR the thing instead of
# the thing.
claude_imports_agents=no
grep -qE "^($bom)?[[:space:]]*@AGENTS\.md[[:space:]]*$" "$root/CLAUDE.md" 2>/dev/null && claude_imports_agents=yes

inline_file=""
inline_files=""
for f in CLAUDE.md AGENTS.md; do
  if [ "$f" = AGENTS.md ] && [ "$claude_imports_agents" = no ]; then
    continue
  fi
  if grep -q "^\($bom\)\{0,1\}<!-- trellis:begin" "$root/$f" 2>/dev/null; then
    [ -n "$inline_file" ] || inline_file="$f"
    inline_files="${inline_files:+$inline_files and }$f"
  fi
done

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
  # The inline managed block belongs in this condition too (decision-0073 D2):
  # a block in CLAUDE.md/AGENTS.md is read by the host at launch exactly like
  # the rendered file and the overlay imports, so a declining project carrying
  # one used to get total silence while any rules embedded in it were live —
  # the same defect this branch's comment already records for the overlay,
  # on a transport the hook could not see at all.
  #
  # TWO COMPLETE LITERALS on purpose, not a base plus an interpolated note.
  # The destructive/deletion guards in cli/plugin_hook_test.go scan `emit "…"`
  # literals; prose assembled into a variable and spliced in is agent-facing
  # text those guards never see — a blind spot in a declared guard (this run's
  # review found it as such). Every string that reaches an agent lives inside
  # an emit literal, even at the cost of repeating the base message.
  if [ -n "$inline_file" ]; then
    emit "TRELLIS_NOT_GOVERNING — this project declares governed = false in .trellis/rules.toml, so Trellis does not govern here: no rule applies, including the two floor- rules. DISREGARD any Trellis rules already loaded this session — from .claude/rules/trellis.md, or from a managed block importing .trellis/internal/. Those are read by the host at launch, before any hook runs, so they could not be withheld. This project also carries a Trellis managed block in $inline_files (the inline shape, in each file named): any rules embedded between its markers were likewise loaded at launch and must be disregarded too — and if the block holds only @-import lines, it may be delivering nothing at all. To stop them being loaded at all, run /trellis:remove."
  elif [ -f "$root/.claude/rules/trellis.md" ] || [ -d "$root/.trellis/internal" ] || [ -f "$root/.trellis/trellis.md" ]; then
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

# The inline counterpart (decision-0073 D2/AC2): a managed block in an
# instruction file PLUS a rendered file, with no overlay to claim the arm
# above. The rendered file is loaded by the host unconditionally; the block may
# embed the rules (a live second copy) or be a dangling import delivering
# nothing — the probe cannot tell, so unlike the overlay arm this one does not
# assert "twice" as fact. What it does assert: two static delivery shapes
# coexist, and the project should keep at most one.
if [ -n "$inline_file" ] && [ -f "$root/.claude/rules/trellis.md" ]; then
  emit "TRELLIS_STATIC_SHAPES_CONFLICT — this project has BOTH a rendered .claude/rules/trellis.md and a Trellis managed block in $inline_files. The rendered file is loaded by the host before any hook runs. The block, if it embeds the rules readout between its markers, puts the same rules in context twice; if it holds only dangling @-import lines whose .trellis/ targets are gone, it delivers nothing — read each block named above to tell which. Either way, keep at most one static shape: delete the managed block from EACH of $inline_files (its trellis:begin marker through its trellis:end marker in every file named — leaving one behind leaves this conflict live) to keep the rendered file, or delete .claude/rules/trellis.md to keep the block — keeping .trellis/rules.toml either way. Show the user the exact paths you would delete and get explicit confirmation before deleting anything (floor-intent-gate): this hook advises, it never authorises a deletion, and the files are tracked. Tell the user before doing substantive work."
  exit 0
fi

# TRL-101. True when $1 has the shape of a stamp some Trellis install wrote:
# payload@ or plugin@ followed by lowercase hex, the retired setup skill's
# fallback plugin@unknown (decision-0039), or a three-part version as the CLI wrote it
# (a release's v0.2.16, a dev build's 0.0.0-dev, a bare 0.2.16). A project's
# stamp file is the project's own, and a repository can commit it as a symbolic
# link to any file the user can read, so a first line of any other shape is never
# quoted back into the session's context. The character lists are spelled out
# rather than written as ranges: a range such as a-f follows the locale's
# collation, and under a UTF-8 locale it admitted accented letters. The Codex
# hook already refuses every stamp but payload@<12 hex> and names none.
stamp_shaped() {
  case "$1" in
    plugin@unknown) return 0 ;;
    payload@?* | plugin@?*)
      case "${1#*@}" in *[!0123456789abcdef]*) return 1 ;; esac
      [ "${#1}" -le 72 ]
      return
      ;;
  esac
  set -- "${1#v}"
  set -- "${1%-dev}"
  case "$1" in
    *.*.*.* | .* | *. | *..*) return 1 ;;
    ?*.?*.?*) ;;
    *) return 1 ;;
  esac
  case "$1" in *[!0123456789.]*) return 1 ;; esac
  [ "${#1}" -le 32 ]
}

if [ -d "$internal" ]; then
  # THE VENDORED OVERLAY IS PAYLOAD TOO — the same bundle files, copied into
  # the consumer's tree — so all three go through the same gateway, in one
  # loop, with one disposition. They did not before, and the three ways they
  # differed were each a live instance of the class TRL-33 names:
  #
  #   * a bare `-f` test on the stamp caught an ABSENT one loudly while an EMPTY
  #     fell through to `[ -n "$overlay" ] || exit 0` below and exited in total
  #     silence — the same absent-vs-empty split TRL-33 found on the plugin
  #     side, one path over.
  #   * `[ ! -s "$internal/$f" ]` catches missing-or-empty and NOT unreadable.
  #     Measured: mode 000 on .trellis/internal/rules.md passed this check and
  #     the hook then emitted "Trellis overlay may be stale ... Until then this
  #     session is governed by the vendored copy" — FALSE. The host's import of
  #     that file fails, so nothing governs, and the message asserted the
  #     opposite of the reader's actual state.
  #   * The stamp was read a second time with `head ... 2>/dev/null`, which is
  #     the swallow-and-continue shape the gateway exists to end.
  #
  # ONE message for all three files, naming $f. The two it replaces said
  # different things about the same broken overlay depending on which file was
  # broken; a reader gains nothing from that and the remedy is identical.
  overlay=""
  overlay_line=""
  for f in version trellis.md rules.md; do
    if ! payload_read "$internal/$f"; then
      emit "TRELLIS_RULES_NOT_LOADED — this project's vendored overlay is incomplete: .trellis/internal/$f $payload_why, so the managed block's imports cannot load the rules and this hook cannot tell which rules the surviving files represent. The hook will not inject over a broken overlay. To migrate onto plugin-delivered rules, delete .trellis/internal/ and the managed block from this project's instructions file, keeping .trellis/rules.toml. Show the user the exact paths you would delete and get explicit confirmation before deleting anything (floor-intent-gate): this hook advises, it never authorises a deletion, and the files are tracked. Tell the user before doing substantive work."
      exit 0
    fi
    # TRL-101: overlay_line is the first line with only surrounding whitespace
    # trimmed. The quote decision reads it, because overlay has every whitespace
    # character deleted, which once made a line such as "payload@ deadbeefcafe"
    # look like a stamp.
    if [ "$f" = version ]; then
      overlay="$(printf '%s\n' "$payload_text" | head -n1 | tr -d '[:space:]')"
      overlay_line="$(printf '%s\n' "$payload_text" | head -n1 | LC_ALL=C sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
    fi
  done
  # A non-empty file can still hold nothing but whitespace on its first line.
  [ -n "$overlay" ] || {
    emit "TRELLIS_RULES_NOT_LOADED — this project's vendored overlay is incomplete: .trellis/internal/version carries no stamp on its first line, so this hook cannot tell which rules the surviving files represent and will not inject over a broken overlay. To migrate onto plugin-delivered rules, delete .trellis/internal/ and the managed block from this project's instructions file, keeping .trellis/rules.toml. Show the user the exact paths you would delete and get explicit confirmation before deleting anything (floor-intent-gate): this hook advises, it never authorises a deletion, and the files are tracked. Tell the user before doing substantive work."
    exit 0
  }
  # TRL-101: the stamp is quoted only when it has a stamp's shape.
  overlay_said="stamp is $overlay"
  overlay_named=" ($overlay)"
  if ! stamp_shaped "$overlay_line"; then
    overlay_said="holds no recognisable version stamp (its first line is not quoted here)"
    overlay_named=""
  fi
  # TRL-34. This was `[ -n "$current" ] || exit 0` — silent. The overlay is
  # intact and the session IS governed by it, so this is NOT a
  # TRELLIS_RULES_NOT_LOADED: reusing the blackout marker here would be the
  # over-correction this change is as concerned with as the silence. What is
  # withheld is a WARNING, and the fix is to say so.
  if [ -z "$current" ]; then
    emit "TRELLIS_STALENESS_UNKNOWN — this session is governed by the vendored overlay at .trellis/internal/, and that is intact. What this hook could NOT do is check whether the overlay is stale: the installed Trellis plugin's own version stamp ($ref) $stamp_defect, so there is nothing to compare this project's stamp$overlay_named against. Nothing is wrong with this project and no rules are missing. Reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the likely fix."
    exit 0
  fi
  if [ "$overlay" != "$current" ]; then
    emit "Trellis overlay may be stale: this project's .trellis/internal/version $overlay_said, but the installed Trellis plugin ships $current. This project still carries a vendored overlay, which the plugin no longer writes or refreshes. To move it onto plugin-delivered rules, delete .trellis/internal/ and the managed block from this project's instructions file, keeping .trellis/rules.toml rows. Show the user the exact paths you would delete and get explicit confirmation before deleting anything (floor-intent-gate): this hook advises, it never authorises a deletion, and the files are tracked. Until then this session is governed by the vendored copy."
  fi
  exit 0
fi

legacy="$root/.trellis/version"
if [ -f "$legacy" ]; then
  # A legacy overlay is stale BY ITS LAYOUT — decision-0051 moved the stamp,
  # so the nudge below does not depend on comparing two stamps at all. Both
  # earlier guards exited silently when a stamp could not be read, which
  # withheld a migration nudge that was correct either way; the two literals
  # below say what could not be read instead of vanishing.
  overlay=""
  overlay_line=""
  if payload_read "$legacy"; then
    overlay="$(printf '%s\n' "$payload_text" | head -n1 | tr -d '[:space:]')"
    overlay_line="$(printf '%s\n' "$payload_text" | head -n1 | LC_ALL=C sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
  fi
  if [ -z "$overlay" ] || [ -z "$current" ]; then
    emit "Trellis overlay predates the .trellis/internal/ layout (decision-0051): its stamp sits at the legacy path .trellis/version. This hook could not read both stamps, so it cannot say how far behind this overlay is — but the LAYOUT itself is the stale part and the migration below is correct regardless. To migrate, delete the legacy overlay — .trellis/version, .trellis/trellis.md and .trellis/internal/ if present, plus the managed block from this project's instructions file — keeping your .trellis/rules.toml rows. An overlay this old may predate .trellis/rules.toml entirely; if there is none, write $root/.trellis/rules.toml containing exactly these two lines, each ending in a newline: \`# Every Trellis rule applies. To switch one off, add a row: <slug> = { active = false }\` then \`[rules]\`. Show the user the exact paths you would delete and get explicit confirmation before deleting anything (floor-intent-gate): this hook advises, it never authorises a deletion, and the files are tracked."
    exit 0
  fi
  # TRL-101: the stamp is quoted only when it has a stamp's shape.
  legacy_named="$overlay; "
  stamp_shaped "$overlay_line" || legacy_named="its first line is not a version stamp and is not quoted; "
  emit "Trellis overlay predates the .trellis/internal/ layout (decision-0051): its stamp sits at the legacy path .trellis/version (${legacy_named}the installed plugin ships $current). To migrate, delete the legacy overlay — .trellis/version, .trellis/trellis.md and .trellis/internal/ if present, plus the managed block from this project's instructions file — keeping your .trellis/rules.toml rows. An overlay this old may predate .trellis/rules.toml entirely; if there is none, write $root/.trellis/rules.toml containing exactly these two lines, each ending in a newline: \`# Every Trellis rule applies. To switch one off, add a row: <slug> = { active = false }\` then \`[rules]\`. Show the user the exact paths you would delete and get explicit confirmation before deleting anything (floor-intent-gate): this hook advises, it never authorises a deletion, and the files are tracked."
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
  # payload that ships the whole assessable set). A truncation severe enough to
  # matter cannot keep
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
      else if (n < 5) print "rule text (only " n " rule line(s) survived; the payload ships the full rule set)"
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
  # TRL-34, path C's arm. `[ -n "$current" ] && ...` skipped the comparison in
  # silence, so an unreadable plugin stamp meant this project was told its
  # rendered file is fine when the hook had in fact checked nothing. The
  # stand-down is still correct — the host loaded that file and it is complete
  # — so this is a SECOND STAND-DOWN LITERAL, not a refusal. Two complete
  # literals rather than a base plus an interpolated note, for the reason the
  # governed=false branch above gives: the destructive-instruction guards in
  # cli/plugin_hook_test.go scan `emit "…"` literals, and prose assembled into
  # a variable and spliced in is agent-facing text those guards never see.
  if [ -z "$current" ]; then
    emit "TRELLIS_STALENESS_UNKNOWN — Trellis rules are already loaded from .claude/rules/trellis.md (the curl install path), so this hook injected nothing; that file and .trellis/rules.toml govern this session, and the file is complete. What this hook could NOT do is check whether it is stale: the installed Trellis plugin's own version stamp ($ref) $stamp_defect, so there is nothing to compare the file's own stamp ($rendered_stamp) against. Reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the likely fix."
    exit 0
  fi
  if [ "$rendered_stamp" != "$current" ]; then
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

# ---------------------------------------------------- S4: the inline managed block
# (decision-0073 D2/AC2.) Reaching here, the project has no overlay, no legacy
# stamp and no rendered file — the block the probe found is the only static
# shape, so this branch sees exactly embedded-or-dangling S4 (S2 took path A
# above, S3 the flat branches). Falling through to path B was the P1: the full
# payload injected on top of a block the host had already loaded. The probe
# cannot tell an EMBEDDED block (rules body between the markers — injecting
# would be double delivery) from a DANGLING import (bare @-import lines whose
# .trellis/internal/ overlay was deleted — nothing loaded, silently
# ungoverned), so the refusal names both states, says how to tell, and asserts
# neither as fact.
if [ -n "$inline_file" ]; then
  emit "TRELLIS_INLINE_BLOCK — $inline_files carries a Trellis managed block (its trellis:begin marker at column 0), so this hook injected nothing. This project is in one of two states and the hook cannot tell which: if the rules readout is written out between the block's markers, the host already loaded those rules at launch and injecting here would put them in context twice; if the block holds only @-import lines whose .trellis/internal/ overlay was deleted, no rules are loaded and this session is ungoverned. Read each block named above to tell which. To move onto plugin-delivered rules either way, delete the managed block from EACH of $inline_files — everything from its trellis:begin marker through its trellis:end marker, in every file named; leaving one behind leaves this project in the same refused state — keeping .trellis/rules.toml if that file exists. Without it, read each block for rows set to active = false BEFORE deleting anything, then write $root/.trellis/rules.toml containing exactly these two lines: \`# Every Trellis rule applies. To switch one off, add a row: <slug> = { active = false }\` then \`[rules]\`, followed by each row a block sets to active = false, so the project keeps the rules it switched off instead of silently gaining them back. If two blocks disagree about a row, show the user both and let them choose — never pick one silently. Or run /trellis:remove to take Trellis out of this project entirely — the opposite endpoint, not a migration. Show the user the exact lines you would delete and get explicit confirmation before deleting anything (floor-intent-gate): this hook advises, it never authorises a deletion, and the file is tracked. Tell the user before doing substantive work."
  exit 0
fi

# ---------------------------------------------------------------------- path B
toml="$root/.trellis/rules.toml"

# A path that EXISTS but is not a regular file — a directory, a FIFO, a socket,
# a device node — is neither a rules file nor a missing one, and every read
# below is guarded by `-f`, so without this it fell into the missing-file
# branch: a vendored bundle governed as though the project had no file, and a
# user-scope install announced the same and told the user to WRITE that file —
# over a FIFO, a write that blocks exactly as the read did. Both were wrong
# about the reader's state (decision-0073's class). Measured on the guarded
# hook before this check: `mkdir .trellis/rules.toml` drew
# TRELLIS_NOT_YET_GOVERNING and "has no .trellis/rules.toml". So it is refused,
# loudly, like the unreadable file further down, before anything tries to read
# it.
#
# HOSTS AGREE HERE, with different words. codex-context.mjs stat-checks
# `isFile()` while locating the overlay and never reaches its `unreadable-file`
# class on this input: measured, a FIFO or a directory at .trellis/rules.toml
# draws `project-root: project-root-not-found` from Codex. Neither host governs,
# neither is silent; only the label differs, and Codex's names the walk that
# skipped the path rather than the path itself. Recorded so the next reader
# does not go looking for the `unreadable-file` that TRL-43's text expected.
# `-e` follows symlinks, as `-f` does and as Codex's statSync does, so a
# dangling symlink is missing on both hosts and a symlink to a FIFO is this.
if [ -e "$toml" ] && [ ! -f "$toml" ]; then
  emit "TRELLIS_RULES_NOT_LOADED — this project's .trellis/rules.toml exists but is not a regular file (a directory, a FIFO, a socket or a device node; \`ls -l $toml\` says which), so the Trellis plugin hook did not open it. This project is configured for Trellis: something sits at .trellis/rules.toml, but it is not a rules file and it is not a governed = false opt-out, so the session is running ungoverned and NO rules were injected. Nothing on disk was changed. To govern this project, replace it with a regular file containing exactly these two lines: \`# Every Trellis rule applies. To switch one off, add a row: <slug> = { active = false }\` then \`[rules]\` — or with the single line governed = false to opt out. Show the user the exact path you would replace and get explicit confirmation before removing anything (floor-intent-gate): this hook advises, it never authorises a deletion. Tell the user before doing substantive work."
  exit 0
fi

# decision-0070. Adoption is the consent act, and every path has one; what a
# missing rules.toml means now depends on WHICH path installed this plugin.
#
# $rows_present says whether there is a project file to classify below. Only
# the vendored-bundle branch sets it to no.
rows_present=yes
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
  # reported project scope and delivered the whole rule set with no announcement.
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
    # visibly, greppably, and revocably by deleting it. No file therefore means
    # every rule applies. There is nothing to classify, so the delivery below
    # carries no activation section and never reads $toml (TRL-97). This branch
    # used to stand the shipped default rows in for the project file; a
    # leftover read of $toml here would now refuse every such project.
    rows_present=no
  else
    # D4, as corrected by decision-0077. A user-wide install is a broad choice,
    # and this says so in the project it is about to affect rather than assuming
    # consent it never asked for. Announce, inject NO rules on this turn ("will
    # be", not "is"), and name both answers.
    #
    # SILENCE IS NOT ONE OF THEM. 0070 D4 said an ignored prompt seeds the file
    # ("accept, or no objection -> seed"); this code has never done that, in any
    # version since #218 built the record. An unanswered announcement leaves the
    # project ungoverned and recurs next session — which the message below states
    # in as many words. decision-0077 corrected the record to match the code
    # rather than the reverse, so that nothing is governed by silence.
    #
    # The accept instruction quotes the new file byte for byte, one backticked
    # span per line (TRL-97, KTD9); install.sh seeds the same bytes, and a test
    # pins the pair.
    emit "TRELLIS_NOT_YET_GOVERNING — the Trellis plugin is installed outside this project (user scope, or a location this hook cannot place), so it applies to every project opened here, and $root has no .trellis/rules.toml. Tell the user, in your own words and before doing substantive work: \"Trellis is installed for your user account, so this repo will be governed by it — 16 rules, followed by default and deviations said out loud. Do you want to disable that for this repo?\" If they want it DISABLED, write .trellis/rules.toml containing exactly the line: governed = false — and nothing else. If they ACCEPT, write $root/.trellis/rules.toml containing exactly these two lines, each ending in a newline: \`# Every Trellis rule applies. To switch one off, add a row: <slug> = { active = false }\` then \`[rules]\` — so the choice persists; without that file this same announcement repeats every session and the project is never governed. (That file is theirs to edit afterwards: a row <slug> = { active = false } under [rules] switches that rule off, and the rule stays off until that row is gone, whatever any other row for it says.) Inject and follow no Trellis rules this turn: none are active yet."
    exit 0
  fi
fi

# One header ships (TRL-97). `strictness` selects nothing any more, and every
# project receives the By default posture sentence.
header="$plugin/reference/trellis.md"
rules="$plugin/reference/rules.md"

# Fail loudly rather than govern silently on a partial payload. A hook cannot
# report that it never ran, but it can report that it ran and could not deliver.
#
# THROUGH THE GATEWAY, both of them. This was a bare `[ ! -f ]` pair, and `-f`
# proves existence and never readability — the gap the header fell through four
# separate ways (mode 000, zero-byte, truncated, and the second posture header
# that has since retired), each measured, each shipping activation rows with
# zero rules prose under them at exit 0. The header is read ONCE, here, where a
# failure can still be reported; $header_prose is what the assembly below uses,
# so the fatal positional open that produced that damage is gone rather than
# merely guarded.
if ! payload_read "$rules"; then
  emit "TRELLIS_RULES_NOT_LOADED — the Trellis plugin hook ran but could not read its own rules payload: $rules $payload_why. This project is configured for Trellis, but the session is running ungoverned and NO rules were injected. This is a broken or half-written plugin install, not a problem with this project: reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the likely fix. Tell the user before doing substantive work."
  exit 0
fi
if ! payload_read "$header"; then
  emit "TRELLIS_RULES_NOT_LOADED — the Trellis plugin hook ran but could not read the header it was about to inject: $header $payload_why. This project is configured for Trellis, but the session is running ungoverned and NO rules were injected. This is a broken or half-written plugin install, not a problem with this project: reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the likely fix. Tell the user before doing substantive work."
  exit 0
fi
header_prose="$payload_text"

# The payload's own terminator, checked BEFORE its slugs are trusted -- the
# order codex-context.mjs uses and for the reason its comment gives: derive
# first and a broken rules.md yields a broken slug set, after which every
# downstream verdict is about the consumer's file when the defect is the
# plugin's. rules.md ends with the sentinel on its last line, so a truncation
# anywhere loses it, and this catches the shape DIRECTLY rather than inferring
# it from a slug count.
#
# Exactly one occurrence, and it must be the last line: the same two conditions
# Codex enforces (`split(SENTINEL).length - 1 !== 1` and `endsWith`). One
# nuance is deliberately NOT matched: Codex also requires the trailing newline,
# which no portable awk can see, so a rules.md ending at the sentinel with no
# newline passes here and fails there. Stated rather than papered over.
#
# CRLF. `last` is the raw record awk read under RS="\n", so a rules.md checked
# out or packaged with CRLF normalization leaves a trailing \r on it and an
# exact ASCII comparison fails -- reporting `not-last` and blacking out a
# COMPLETE, CORRECT payload. Stripped rather than tolerated in the regex, so
# the comparison stays a comparison.
sentinel_report="$(
  awk '
    { last = $0; sub(/\r$/, "", last); n += gsub(/<!-- trellis:rules-loaded -->/, "&") }
    END { print n + 0, (last == "<!-- trellis:rules-loaded -->" ? "last" : "not-last") }
  ' "$rules"
)"
case "$sentinel_report" in
  "1 last") ;;
  *)
    # Empty when the awk died outright, which is the unreadable-file case; the
    # message covers both readings rather than asserting one.
    [ -n "$sentinel_report" ] || sentinel_report="unreadable"
    emit "TRELLIS_RULES_NOT_LOADED — the Trellis plugin's own rules payload ($rules) is not complete: it must carry exactly one \`<!-- trellis:rules-loaded -->\` terminator as its final line, and this hook found \"$sentinel_report\". A payload cut short of that line is a truncated or half-written install, and every rule below the cut is simply absent — so its slug list cannot be trusted to say what the rule set is. This project is configured for Trellis, but the session is running ungoverned and NO rules were injected; nothing on disk was changed. Reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the likely fix. Tell the user before doing substantive work."
    exit 0
    ;;
esac

# The slugs the payload ships, in rules.md order and each once: a trailing
# backticked slug on a rule line, the anchor codex-context.mjs slugsFromRules
# uses. The classifier below finds unknown slugs and floors against this set,
# and the computed sentence names switched-off rules in its order.
rule_slugs="$(
  LC_ALL=C awk '
    { line = $0; sub(/\r$/, "", line) }
    match(line, /`(inv|floor)-[a-z-]+`[ \t]*$/) {
      s = substr(line, RSTART + 1)
      sub(/`.*$/, "", s)
      if (!(s in seen)) { seen[s] = 1; out = out (out == "" ? "" : " ") s }
    }
    END { print out }
  ' "$rules"
)"
# An EMPTY set is a payload whose rule lines lost their slugs, not a project
# fault: against it every row would name an unknown slug and nothing could be
# switched off, so the session would look governed with no way to tell which
# rules it names. Refused here, before anything consumes the set.
if [ -z "$rule_slugs" ]; then
  emit "TRELLIS_RULES_NOT_LOADED — the Trellis plugin's own rules payload ($rules) carries no rule slugs, so this hook cannot tell which rules it ships. This project is configured for Trellis, but the session is running ungoverned. This is a broken or unrecognisable plugin payload, not a problem with this project — reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the likely fix, not editing .trellis/rules.toml. Tell the user before doing substantive work."
  exit 0
fi

# The header carries `@rules.md`, an import the hook resolves itself, and a
# pointer at the vendored invariants path, which does not exist in this mode.
# Repointing it at the plugin keeps the trigger-read affordance and cannot go
# stale, because it names the payload this session is actually running.
#
# The Codex hook has always refused a header that yields nothing --
# readRequired reports unreadable-file/missing-file, and an explicit check
# rejects empty prose (codex-context.mjs) -- so the Claude-side gap was an
# oversight, not a design choice. Match it. The header is read ONCE, above,
# where a failure can still be reported, and the assembly reads that text from
# stdin, so no positional open of it is left to fail silently.
#
# Exactly one `@rules.md` import is required for the same reason Codex rejects
# invalid-placeholder-count: a header truncated ABOVE that line is non-empty and
# assembles into an activation section with no rules above it. Counted with awk
# over stdin -- never a positional read, which is the failure being closed here.
# $header_prose was read through payload_read at the top of path B, which is
# also where emptiness is refused; the `-z` test below is kept as the second
# lock on the same door rather than as the only one.
header_imports="$(printf '%s\n' "$header_prose" |
  awk '/^@rules\.md[[:space:]]*$/ { n++ } END { print n + 0 }')"
if [ -z "$header_prose" ] || [ "$header_imports" != "1" ]; then
  emit "TRELLIS_RULES_NOT_LOADED — the Trellis plugin hook could not assemble its own rules payload: the header it was about to inject ($header) read as empty, or carries ${header_imports} @rules.md imports where exactly one is required, so the rules themselves would have been missing from what was injected. This project is configured for Trellis, but the session is running ungoverned and NO rules were injected — the hook refused rather than deliver a rule activation section with no rules above it. This is a broken or half-written plugin payload, not a problem with this project: reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the likely fix. Tell the user before doing substantive work."
  exit 0
fi
# The `@rules.md` expansion reads through a redirected getline, whose -1 on a
# failed open is silent, so its return value is checked: a failed open used to
# print ZERO rules prose and carry on.
#
# What triggered that is the `-v` channel, not a permission: `awk -v`
# ESCAPE-PROCESSES its value (`awk -v v=/a\tb/c` yields length 6, not 7), so a
# CLAUDE_PLUGIN_ROOT containing a backslash reached this awk as a DIFFERENT path
# than the one every `-f` test and every positional read used -- all of which
# pass. Measured with a root named `plug\tools`: activation rows, 0 rules
# prose, exit 0, no marker, which is verbatim the damage shape the header guard
# above exists to stop.
#
# `inv` took the identical mangling in silence, so the invariants pointer this
# block exists to REPOINT was itself wrong on such a root. ENVIRON does no escape
# processing, so both values arrive verbatim and a backslash root now works
# rather than merely failing loudly.
#
# The substitution is index-based rather than gsub-based, and the BEGIN escaping
# that guarded it is GONE with it. That escaping was half right and half wrong,
# and only measuring said which half. Measured on BSD awk 20200816, the awk this
# ran on, with the fixtures passed through ENVIRON so the -v channel could not
# mangle them before the test began:
#
#   input        escaped      via escaped replacement   via raw replacement
#   plug\tools   plug\\tools  plug\\tools  (doubled)     plug\tools  (right)
#   x\\y         x\\\\y       x\\\\y        (doubled)     x\\y        (right)
#   R&D          R\&D         R&D          (right)       RTOKD       (wrong)
#
# So `&` IS special in a gsub replacement and guarding it was necessary, while a
# backslash is NOT -- and escaping it doubled every backslash in the delivered
# pointer. That settles it on its own: the escaping was wrong on the awk this
# hook runs on, for every backslash input, whatever any other awk does.
#
# A portability fork exists too, but it is narrower than an earlier version of
# this comment claimed and the example that version cited was the wrong one.
# Reported by review and measured there on gawk 5.4.1; NOT reproducible on this
# machine, which ships no gawk, so it is recorded as attributed measurement
# rather than as something checked here: gawk default AGREES with BSD awk on
# `plug\tools`, the two diverge on DOUBLED input (`x\\y` gives BSD `x\\\\y` and
# gawk `x\\y`), and `gawk --posix` alone round-trips the single-backslash case.
#
# None of that is load-bearing any more. Substituting by index invokes no
# replacement semantics at all, so there is nothing to escape and neither
# metacharacter has anything to do -- on any awk, without needing to know which.
#
# TRL-70. The substitution below hands the model an ADDRESS; until this read
# existed nothing asked whether anything is at it, so a plugin payload with no
# invariants.md shipped a dead pointer in silence. codex-context.mjs closed the
# same gap on its own branch (TRL-69, #294) and the two hosts then disagreed
# about a broken install with the address guard green.
#
# The pointer still moves whatever this read finds, for #294's reason: leaving
# the raw token ships `.trellis/internal/invariants.md` to the one mode DEFINED
# by not having that directory, which is a strictly worse dead pointer. And the
# session is not failed closed over it -- invariants.md is CONSULTED on demand,
# never injected, and decision-0093 rule 1 rules that failing a session over
# such a file trades a dead pointer for no governance at all. What the defect
# earns is the report at the end of the payload block below.
#
# Through payload_read rather than a bare `-f`, which is not a stylistic
# choice: TestNoPayloadReadBypassesTheGateway requires it of every payload path
# this file names, and the gateway is worth more than the rule here, because it
# CLASSIFIES. codex-context.mjs asks existingFile, a bare stat, and cannot tell
# a missing copy from an unreadable or an empty one; the remedies differ, so
# the reader is told which. This makes the Claude host the stricter of the two
# on the same broken install -- deliberate, and recorded in
# TestBothHostsReportAMissingInvariantsTarget rather than left to be found.
#
# Reads LAST of the payload reads on purpose. payload_read assigns $payload_text
# and $payload_why, and every earlier caller has already copied what it needed
# out of them ($header_prose, $stamp_defect, $current).
inv="$plugin/reference/invariants.md"
inv_defect=""
payload_read "$inv" || inv_defect="$payload_why"
rules_prose="$(
  printf '%s\n' "$header_prose" |
  TRELLIS_RULES="$rules" TRELLIS_INV="$inv" awk '
    BEGIN {
      rules = ENVIRON["TRELLIS_RULES"]
      inv = ENVIRON["TRELLIS_INV"]
      tok = "`.trellis/internal/invariants.md`"
    }
    /^@rules\.md[[:space:]]*$/ {
      rc = 0
      while ((rc = (getline line < rules)) > 0) { print line; imported++ }
      if (rc < 0 || imported == 0) { failed = 1; exit 1 }
      next
    }
    {
      line = $0
      out = ""
      while ((i = index(line, tok)) > 0) {
        out = out substr(line, 1, i - 1) "`" inv "`"
        line = substr(line, i + length(tok))
      }
      print out line
    }
    END { if (failed) print "#trellis-rules-import-failed" }
  '
)"
case "$rules_prose" in
  "" | *"#trellis-rules-import-failed")
    emit "TRELLIS_RULES_NOT_LOADED — the Trellis plugin hook assembled its header but could not import the rules themselves from $rules, so what it was about to inject framed the project's rule activation without saying what any rule is. This project is configured for Trellis, but the session is running ungoverned and NO rules were injected — the hook refused rather than deliver an activation section with nothing above it. This is a broken or half-written plugin payload: reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the likely fix. Tell the user before doing substantive work."
    exit 0
    ;;
esac

# ----------------------------------------------------- the project rules file
# TRL-97. Only a row set `active = false` has any effect, and one is enough: a
# rule is off when any row for it under [rules] says false, whatever its other
# rows say, so a repeated row is never an error. The contract is KTD3 of the
# TRL-97 plan as the maintainer amended it; codex-context.mjs classifyRules
# implements the same one, and TestBothHostsClassifyRulesRowsIdentically runs
# both hooks on one table, so a line class the two read differently is a red
# test rather than two hosts governing one project differently.
#
# Nothing here refuses the file for what it says: a bad entry costs that entry
# and is reported, never the session, because over-governance the session
# announces beats under-governance nobody sees (decision-0083 section 5). What
# is refused is a file this hook cannot read, because a governed = false it
# cannot read cannot be ruled out (KTD5), and a file past the read bound.
#
# The order is the one both hosts share: regular (above), then the governed =
# false read (at the top of this file, before any delivery path), then readable
# and bounded, then the NUL byte check, then classification.

# The read bound, codex-context.mjs MAX_PROJECT_CONFIG_BYTES. A rules file holds
# a comment and a few rows; this is a runaway guard about nine hundred times the
# largest consumer file known on 2026-09-14, and the two hosts refuse the same
# files.
rules_file_max=1048576
# B, codex-context.mjs RULES_ECHO_MAX_BYTES, where its measurement is recorded.
# The file is echoed verbatim under the computed sentence only while its bytes
# plus the bytes of its rendered warning block (every warning line with its
# newline, the count line included) fit this; otherwise one line in the payload
# block below stands in for it. The Codex context budget is the tighter of the
# two hosts, so it sets the value, and both hosts share it so the activation
# sections stay identical. They differ only for a byte that is not valid UTF-8:
# this hook charges the bytes read, and Codex the three bytes such a byte decodes
# to, so Codex can show the too-large line where this hook echoes (TRL-100).
rules_echo_max=1800
# The warning cap (KTD4), codex-context.mjs WARNINGS_NAMED and WARNING_NAME_MAX.
# It is applied in one place, at the end of the classifier below: five warnings
# are named in total, first the attempts (a false row for a shipped rule that
# takes no effect) and then the rest, each group in file line order; past the
# fifth, one count line says how many more entries were ignored and how many of
# them are attempts. A name taken from the file is cut at
# $rules_warning_name_max bytes, so a runaway file cannot turn its warnings into
# the context.
rules_warnings_named=5
rules_warning_name_max=40

rules_off=""
rules_warnings=""
rules_echo=yes
rules_symlink=no
toml_size=0
if [ "$rows_present" = yes ]; then
  # A rules file reached through a symbolic link (TRL-97). A repository can
  # commit .trellis/rules.toml, or the directory holding it, as a link to
  # any file the user can read, a credentials file included, and the echo would
  # put that file into the model context. It is still read and classified, so
  # its opt-outs apply and the sentence names them from the payload slugs, but
  # nothing read from it is shown: no echo and no key, slug or line number in a
  # warning, only counts. Only the two components this project owns are tested,
  # never the project root, whose own path often runs through a link. The
  # governed = false read, the read bound and the NUL byte check go through the
  # link unchanged.
  if [ -L "$toml" ] || [ -L "$root/.trellis" ]; then
    rules_symlink=yes
  fi
  classified=""
  if [ -r "$toml" ] && toml_size="$(wc -c 2>/dev/null < "$toml")"; then
    toml_size="$(printf '%s' "$toml_size" | tr -d '[:space:]')"
    if [ "$toml_size" -gt "$rules_file_max" ]; then
      emit "TRELLIS_RULES_NOT_LOADED — this project's .trellis/rules.toml is ${toml_size} bytes, over the ${rules_file_max}-byte bound this hook reads a rules file up to, so it was not read and NO rules were injected. A rules file holds a comment and a few rows, so this one is probably not the file the project meant to commit; show the user its size. Tell the user before doing substantive work."
      exit 0
    fi
    # A NUL byte (KTD5). Such a file is not text both hosts split into the same
    # lines, so none of its rows takes effect, it is not echoed, and one warning
    # says why. Counted by tr and wc, which see every byte; a shell variable
    # cannot hold a NUL at all. The governed = false read has already run, so an
    # opted-out file with a NUL byte still loads nothing.
    if [ "$(LC_ALL=C tr -cd '\000' 2>/dev/null < "$toml" | wc -c | tr -d '[:space:]')" != 0 ]; then
      classified="$(printf 'off \nwarn nul-byte 0 #\n')"
      rules_echo=no
    else
      # The classifier. It prints one `off <slugs>` line, the switched-off
      # rules in rules.md order, then one `warn <kind> <line> <name>` record per
      # named warning in the order it is shown, already bounded, then at most one
      # count record: `warn more <entries> <attempts>`, or for a symlinked file
      # `warn symlink-count <entries> <attempts>` and nothing else. `#` stands in
      # for a name the entry does not have, since no slug or key can hold one.
      # The words of each warning live in the payload block below, as printf
      # format literals the destructive-instruction scans read. A file it cannot
      # open prints `#unreadable`, which is refused below rather than read as a
      # file that switches nothing off.
      #
      # LC_ALL=C, so the character classes are ASCII and length is bytes, as on
      # the other host. Paths and the byte-order mark arrive through ENVIRON,
      # which does no escape processing (see the -v note above).
      classified="$(
        TRELLIS_TOML="$toml" TRELLIS_SLUGS="$rule_slugs" TRELLIS_BOM="$bom" TRELLIS_SYMLINK="$rules_symlink" \
        TRELLIS_NAMED="$rules_warnings_named" TRELLIS_NAME_MAX="$rules_warning_name_max" \
        LC_ALL=C awk '
          # A warning that names a slug (a floor row set false, an unknown slug
          # set false, a row above [rules]) is noted once per kind and slug, at
          # the first line that draws it, so a repeated row adds none. A later
          # row can only make that entry an attempt: a well-formed false row for
          # a shipped slug that takes no effect, which the cap names first.
          function note(ln, kind, name, attempt,    k) {
            if (kind == "floor-row" || kind == "unknown-slug" || kind == "row-above-rules") {
              k = kind " " name
              if (k in first) {
                if (attempt) e_attempt[first[k]] = 1
                return
              }
              first[k] = n + 1
            }
            n++
            e_line[n] = ln
            e_kind[n] = kind
            e_name[n] = name
            e_attempt[n] = attempt ? 1 : 0
          }
          # The value a governed line holds, as the regex on the other host
          # reads it: the longest run with no space or tab, unless what follows
          # it is neither blank nor a comment, when the match backtracks to the
          # last # inside that run. #null is no value at all.
          function governed_value(v,    tok, i) {
            tok = v
            sub(/[ \t].*$/, "", tok)
            if (substr(v, length(tok) + 1) ~ /^[ \t]*(#.*)?$/) return tok
            for (i = length(tok) - 1; i >= 0; i--)
              if (substr(tok, i + 1, 1) == "#") return substr(tok, 1, i)
            return "#null"
          }
          BEGIN {
            path = ENVIRON["TRELLIS_TOML"]
            bom = ENVIRON["TRELLIS_BOM"]
            named_max = ENVIRON["TRELLIS_NAMED"] + 0
            name_max = ENVIRON["TRELLIS_NAME_MAX"] + 0
            symlink = ENVIRON["TRELLIS_SYMLINK"]
            nslugs = split(ENVIRON["TRELLIS_SLUGS"], order, " ")
            for (i = 1; i <= nslugs; i++) shipped[order[i]] = 1
            section = "top"
            lineno = 0
            while ((rc = (getline raw < path)) > 0) {
              lineno++
              line = raw
              if (lineno == 1 && bom != "" && substr(line, 1, length(bom)) == bom) line = substr(line, length(bom) + 1)
              sub(/\r$/, "", line)
              sub(/^[ \t]+/, "", line)
              sub(/[ \t]+$/, "", line)
              if (line == "" || substr(line, 1, 1) == "#") continue

              if (substr(line, 1, 1) == "[") {
                if (line ~ /^\[[ \t]*rules[ \t]*\]([ \t]*#.*)?$/) {
                  if (rules_opened) note(lineno, "rules-again", "#", 0)
                  rules_opened = 1
                  section = "rules"
                } else {
                  section = "other"
                  note(lineno, "other-table", "#", 0)
                }
                continue
              }
              if (section == "other") continue

              is_row = (line ~ /^(inv|floor)-[a-z-]+[ \t]*=[ \t]*[{][ \t]*active[ \t]*=[ \t]*(true|false)[ \t]*[}]([ \t]*#.*)?$/)
              slug = ""
              value = ""
              if (is_row) {
                match(line, /^(inv|floor)-[a-z-]+/)
                slug = substr(line, 1, RLENGTH)
                value = line
                sub(/^[^{]*[{][ \t]*active[ \t]*=[ \t]*/, "", value)
                value = (substr(value, 1, 4) == "true") ? "true" : "false"
              }
              shipped_false = (is_row && value == "false" && (slug in shipped))
              key = "#"
              rest = ""
              if (match(line, /^[A-Za-z0-9_-]+[ \t]*=/)) {
                match(line, /^[A-Za-z0-9_-]+/)
                key = substr(line, 1, RLENGTH)
                rest = substr(line, RLENGTH + 1)
                sub(/^[ \t]*=[ \t]*/, "", rest)
              }

              if (section == "top") {
                if (line ~ /^(inv|floor)-[a-z-]+[ \t]*=/) {
                  match(line, /^(inv|floor)-[a-z-]+/)
                  note(lineno, "row-above-rules", substr(line, 1, RLENGTH), shipped_false)
                } else if (key == "#") {
                  note(lineno, "top-level-line", "#", 0)
                } else if (key == "governed") {
                  # The opt-out read at the top of this file already exited on
                  # the one shape that opts out, so a governed line reaching
                  # here did not, unless it is the first and says true.
                  if (governed_seen || governed_value(rest) != "true") note(lineno, "governed", "#", 0)
                  governed_seen = 1
                } else if (key != "strictness" && key != "seeded_from") {
                  note(lineno, "unknown-key", key, 0)
                }
                continue
              }

              if (!is_row) {
                if (key == "governed") note(lineno, "governed", "#", 0)
                else note(lineno, "malformed-row", key, 0)
                continue
              }
              if (value == "true") continue
              if (!(slug in shipped)) note(lineno, "unknown-slug", slug, 0)
              else if (substr(slug, 1, 6) == "floor-") note(lineno, "floor-row", slug, 1)
              else off[slug] = 1
            }
            if (rc < 0) {
              print "#unreadable"
              exit
            }

            out = ""
            for (i = 1; i <= nslugs; i++)
              if (order[i] in off) out = out (out == "" ? "" : ", ") order[i]
            print "off " out

            # THE WARNING CAP (KTD4), in the one place it is applied. A
            # symlinked file draws one count and nothing it holds. Otherwise the
            # attempts are named first (pass 1) and then the rest (pass 2), five
            # in all, and one count record covers every entry left.
            attempts = 0
            for (i = 1; i <= n; i++) if (e_attempt[i]) attempts++
            if (symlink == "yes") {
              if (n > 0) print "warn symlink-count " n " " attempts
              exit
            }
            said = 0
            for (pass = 1; pass <= 2; pass++) {
              for (i = 1; i <= n && said < named_max; i++) {
                if (e_attempt[i] != (pass == 1)) continue
                said++
                nm = e_name[i]
                if (nm != "#" && length(nm) > name_max) nm = substr(nm, 1, name_max) "..."
                print "warn " e_kind[i] " " e_line[i] " " nm
              }
            }
            if (n > said) print "warn more " (n - said) " " (attempts - (attempts < named_max ? attempts : named_max))
          }
        '
      )"
    fi
  fi
  case "$classified" in
    off*) ;;
    *)
      emit "TRELLIS_RULES_NOT_LOADED — this project's .trellis/rules.toml exists but could not be read ($toml), so the Trellis plugin hook cannot tell which rules it switches off or whether it declares governed = false, and NO rules were injected. Check the file's permissions. Tell the user before doing substantive work."
      exit 0
      ;;
  esac
  rules_off="$(printf '%s\n' "$classified" | sed -n 's/^off //p')"
  rules_warnings="$(printf '%s\n' "$classified" | sed -n 's/^warn //p')"
fi

payload="$(
  printf '%s\n' "$rules_prose"
  if [ "$rows_present" = yes ]; then
    # THE WARNINGS (KTD4), rendered first because their bytes decide below
    # whether the file is shown, and printed last. They arrive in the order and
    # bound the classifier already applied. Every template is a printf format
    # literal, so the destructive-instruction and deletion scans read it; the
    # line, slug and key arrive as arguments, and for the two count lines the
    # counts do. None quotes the raw line, and none asks for the file to change.
    # Case arms open with a parenthesis: bash 3.2 otherwise reads a bare pattern
    # paren as the end of this command substitution.
    rules_warning_block=""
    rules_warning_bytes=0
    if [ -n "$rules_warnings" ]; then
      rules_warning_block="$(printf '%s\n' "$rules_warnings" | while read -r kind line name; do
        case "$kind" in
          (nul-byte)
            printf 'Trellis warning: .trellis/rules.toml contains a NUL byte, so none of its rows take effect and it is not shown; every rule applies.\n'
            ;;
          (unknown-slug)
            printf 'Trellis warning: line %s of .trellis/rules.toml sets %s to active = false, but this plugin ships no rule by that name, so the row is ignored; another plugin version may ship it.\n' "$line" "$name"
            ;;
          (floor-row)
            printf 'Trellis warning: line %s of .trellis/rules.toml sets the floor rule %s to active = false, but floor rules cannot be switched off, so the row is ignored and the rule applies.\n' "$line" "$name"
            ;;
          (row-above-rules)
            printf 'Trellis warning: line %s of .trellis/rules.toml is a row for %s above the [rules] table, so it is ignored; rows count only under [rules].\n' "$line" "$name"
            ;;
          (rules-again)
            printf 'Trellis warning: line %s of .trellis/rules.toml opens [rules] a second time; the rows under it still count.\n' "$line"
            ;;
          (other-table)
            printf 'Trellis warning: line %s of .trellis/rules.toml opens a table other than [rules], so every line under it is ignored.\n' "$line"
            ;;
          (governed)
            printf 'Trellis warning: line %s of .trellis/rules.toml sets governed but does not opt out; only one governed = false line, above every table, opts a project out, so this project stays governed.\n' "$line"
            ;;
          (unknown-key)
            printf 'Trellis warning: line %s of .trellis/rules.toml sets %s, which is not a key this file defines, so it is ignored and switches no rule off; a row under [rules] such as slug = { active = false } switches a rule off.\n' "$line" "$name"
            ;;
          (top-level-line)
            printf 'Trellis warning: line %s of .trellis/rules.toml is not an entry this file defines, so it is ignored and switches no rule off.\n' "$line"
            ;;
          (malformed-row)
            if [ "$name" = "#" ]; then
              printf 'Trellis warning: line %s of .trellis/rules.toml is a malformed row, so it is ignored and switches no rule off.\n' "$line"
            else
              printf 'Trellis warning: line %s of .trellis/rules.toml is a malformed row naming %s, so it is ignored and switches no rule off; any rule it names still applies.\n' "$line" "$name"
            fi
            ;;
          (more)
            if [ "$line" = 1 ]; then
              if [ "$name" = 1 ]; then
                printf 'Trellis warning: .trellis/rules.toml has 1 more ignored entry, and it tries to switch a rule off.\n'
              else
                printf 'Trellis warning: .trellis/rules.toml has 1 more ignored entry, and it does not try to switch a rule off.\n'
              fi
            elif [ "$name" = 0 ]; then
              printf 'Trellis warning: .trellis/rules.toml has %s more ignored entries, none of which tries to switch a rule off.\n' "$line"
            elif [ "$name" = 1 ]; then
              printf 'Trellis warning: .trellis/rules.toml has %s more ignored entries, 1 of which tries to switch a rule off.\n' "$line"
            else
              printf 'Trellis warning: .trellis/rules.toml has %s more ignored entries, %s of which try to switch a rule off.\n' "$line" "$name"
            fi
            ;;
          (symlink-count)
            if [ "$line" = 1 ]; then
              if [ "$name" = 1 ]; then
                printf 'Trellis warning: .trellis/rules.toml is a symbolic link, so its 1 ignored entry is counted but not named; it tries to switch a rule off.\n'
              else
                printf 'Trellis warning: .trellis/rules.toml is a symbolic link, so its 1 ignored entry is counted but not named; it does not try to switch a rule off.\n'
              fi
            elif [ "$name" = 0 ]; then
              printf 'Trellis warning: .trellis/rules.toml is a symbolic link, so its %s ignored entries are counted but not named; none of them tries to switch a rule off.\n' "$line"
            elif [ "$name" = 1 ]; then
              printf 'Trellis warning: .trellis/rules.toml is a symbolic link, so its %s ignored entries are counted but not named; 1 of them tries to switch a rule off.\n' "$line"
            else
              printf 'Trellis warning: .trellis/rules.toml is a symbolic link, so its %s ignored entries are counted but not named; %s of them try to switch a rule off.\n' "$line" "$name"
            fi
            ;;
        esac
      done)"
      rules_warning_bytes="$(printf '%s\n' "$rules_warning_block" | wc -c | tr -d '[:space:]')"
    fi
    printf '\n## Project rule activation\n\n'
    # THE COMPUTED SENTENCE (KTD1). It names the rules the file switches off, in
    # rules.md order, so the model is told the effective result even where the
    # file below still shows a row this hook ignored.
    if [ -z "$rules_off" ]; then
      printf 'The project file .trellis/rules.toml switches no rule off, so every rule above applies.\n'
    else
      printf 'The project file .trellis/rules.toml switches these rules off: %s. Every other rule above applies.\n' "$rules_off"
    fi
    # THE FILE, OR THE LINE THAT STANDS IN FOR IT. A file reached through a
    # symbolic link is never shown, at any size. Any other file is echoed
    # verbatim, with a newline supplied when it has none, only while its bytes
    # plus the bytes of the warning block fit B; past that, one line. A file
    # holding a NUL byte is neither, and an empty file has nothing to show.
    if [ "$rules_echo" = yes ]; then
      if [ "$rules_symlink" = yes ]; then
        printf '\n'
        printf 'The project file is a symbolic link, so its contents are not shown here; the sentence above names every rule it switches off.\n'
      elif [ "$((toml_size + rules_warning_bytes))" -gt "$rules_echo_max" ]; then
        printf '\n'
        printf 'The project file and its warnings are too large to show here; the sentence above names every rule it switches off.\n'
      elif [ "$toml_size" -gt 0 ]; then
        printf '\n'
        cat "$toml"
        [ -z "$(tail -c 1 "$toml")" ] || printf '\n'
      fi
    fi
    if [ -n "$rules_warning_block" ]; then
      printf '\n%s\n' "$rules_warning_block"
    fi
  fi
  # The path-B arm of TRL-34. Delivery is unaffected -- the rules payload is
  # fine and the session IS governed -- so an unreadable version stamp degrades
  # the PROVENANCE line and nothing else. It used to degrade it silently,
  # printing an empty pair of parentheses.
  # (NO APOSTROPHES anywhere inside this payload="$( ... )" block, comments
  # included: bash 3.2 scans for the closing paren without re-entering comment
  # context, so one lone apostrophe in a comment here swallows the rest of the
  # file. Cost this change a syntax error before it cost a reader anything.)
  if [ -n "$current" ]; then
    printf '\nDelivered by the Trellis plugin (%s). No overlay is vendored in this project.\n' "$current"
  else
    printf '\nDelivered by the Trellis plugin. Its own version stamp could not be read (%s %s), so this readout cannot name which payload build it came from; the rules above are complete and govern this session. No overlay is vendored in this project.\n' "$ref" "$stamp_defect"
  fi
)"

# A bounded payload, like the Codex hook's MAX_CONTEXT_BYTES. It once caught a
# runaway rules.toml — measured, a 5 MB file produced 4.8 MB of valid JSON and
# exit 0 — and since TRL-97 no project file can reach it: the file is echoed
# only while it and its warnings fit $rules_echo_max bytes, and its warnings are
# bounded. What is left for it is a plugin payload that has outgrown the budget
# on its own.
limit=32768
size=$(printf '%s' "$payload" | wc -c | tr -d '[:space:]')
if [ "$size" -gt "$limit" ]; then
  emit "TRELLIS_RULES_NOT_LOADED — the assembled Trellis rules are ${size} bytes, over the ${limit}-byte injection budget, so nothing was injected. The project file is echoed only while it and its warnings fit ${rules_echo_max} bytes, and its warnings are bounded, so this is a plugin payload that has outgrown the budget: reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the likely fix. Tell the user before doing substantive work."
  exit 0
fi

# TRL-70, and the sibling of the provenance arm at the end of the payload
# block: delivery is unaffected -- the rules and the rows are complete and the
# session IS governed -- so a CONSULTED reference that cannot be read degrades
# this one line and nothing else, the same way an unreadable version stamp
# degrades the line above it.
#
# Said in the payload rather than on a channel of its own because emit writes
# exactly one key. Every TRELLIS_ marker this hook can print is a path that
# injects NOTHING, so borrowing that vocabulary for a complete, governing
# delivery would blur the one signal a consumer can rely on. The Codex hook has
# a systemMessage field beside its context and uses it; this one does not, and
# the two reports are pinned on their SUBSTANCE rather than their channel.
#
# APPENDED AFTER THE BUDGET CHECK, not inside the payload block, and that
# ordering is the whole point rather than a tidiness choice. Measured on the
# version that assembled it inside, counting UTF-8 BYTES throughout, which is
# what the wc -c above counts and the only measure this claim can honestly
# quote: a project whose rules.toml put the assembled payload at 32673 bytes
# of the 32768-byte budget was fully governed, and DELETING
# reference/invariants.md -- changing nothing else -- added the report and
# took it to 33149, turning the session into a TRELLIS_RULES_NOT_LOADED
# refusal with no rules and no rows. That is exactly the trade decision-0093
# rule 1 forbids: a dead pointer swapped for no governance at all, over a file
# that is consulted on demand. Codex is not exposed to it -- its budget bounds
# `context` alone (codex-context.mjs:1328) and the same warning rides
# systemMessage outside it -- so this ordering also stops the two hosts
# disagreeing about the one property the pair guard exists to protect.
#
# On those figures, since two rounds of review went at them. The report costs
# 484 bytes at a 70-character plugin root and grows with the path it names;
# 33149 is that run, whose wording was 8 bytes shorter than todays. An earlier
# version of this comment quoted 32573 for the healthy side, which was the
# CHARACTER count of the delivered context rather than its byte count, and
# review could not reproduce the pair. The two measures differ by 100 on that
# payload. Attributing all 100 to em dashes was the SECOND wrong version of
# this sentence: measured, 30 em dashes carry 60 of it, the 17 ballot Xs in
# the live-rows readout carry 34, and two arrows and two middle dots carry the
# last 6. Multi-byte characters generally, not one of them.
#
# Two consequences, both wanted. The refusal above now counts only what it
# blames, instead of charging the diagnostic bytes to rules.toml. And an
# over-budget session drops the report: it has already been told loudly that
# NOTHING was injected, which is the larger problem and names its own remedy.
if [ -n "$inv_defect" ]; then
  payload="$payload$(printf '\n\nThis plugin payload has no readable %s (it %s), so the invariants pointer in the rules above names a file that yields nothing to read. The rules and rows above are complete and govern this session; that reference is consulted on demand, so only a rule that turns out ambiguous needs it. Reinstalling or updating the plugin (`claude plugin update trellis@kodhama`) is the likely fix.' "$inv" "$inv_defect")"
fi

emit "$payload"
exit 0
