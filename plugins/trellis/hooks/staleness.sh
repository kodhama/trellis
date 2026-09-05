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
# So the hook reads this file for exactly one thing — this key — and injects
# nothing when it is set. It never reads the rows; those are live, editable, and
# read by the model on demand, which is the behaviour this design is for.
bom="$(printf '\357\273\277')"
# KNOWN, NARROWED DIVERGENCE. A misplaced `governed = false` UNDER `[rules]` is
# not a top-level key, so neither host opts out — but they then differ: this hook
# ignores the stray line and governs normally (the full rule set), while codex-context.mjs
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
  for f in version trellis.md rules.md; do
    if ! payload_read "$internal/$f"; then
      emit "TRELLIS_RULES_NOT_LOADED — this project's vendored overlay is incomplete: .trellis/internal/$f $payload_why, so the managed block's imports cannot load the rules and this hook cannot tell which rules the surviving files represent. The hook will not inject over a broken overlay. To migrate onto plugin-delivered rules, delete .trellis/internal/ and the managed block from this project's instructions file, keeping .trellis/rules.toml. Show the user the exact paths you would delete and get explicit confirmation before deleting anything (floor-intent-gate): this hook advises, it never authorises a deletion, and the files are tracked. Tell the user before doing substantive work."
      exit 0
    fi
    [ "$f" = version ] && overlay="$(printf '%s\n' "$payload_text" | head -n1 | tr -d '[:space:]')"
  done
  # A non-empty file can still hold nothing but whitespace on its first line.
  [ -n "$overlay" ] || {
    emit "TRELLIS_RULES_NOT_LOADED — this project's vendored overlay is incomplete: .trellis/internal/version carries no stamp on its first line, so this hook cannot tell which rules the surviving files represent and will not inject over a broken overlay. To migrate onto plugin-delivered rules, delete .trellis/internal/ and the managed block from this project's instructions file, keeping .trellis/rules.toml. Show the user the exact paths you would delete and get explicit confirmation before deleting anything (floor-intent-gate): this hook advises, it never authorises a deletion, and the files are tracked. Tell the user before doing substantive work."
    exit 0
  }
  # TRL-34. This was `[ -n "$current" ] || exit 0` — silent. The overlay is
  # intact and the session IS governed by it, so this is NOT a
  # TRELLIS_RULES_NOT_LOADED: reusing the blackout marker here would be the
  # over-correction this change is as concerned with as the silence. What is
  # withheld is a WARNING, and the fix is to say so.
  if [ -z "$current" ]; then
    emit "TRELLIS_STALENESS_UNKNOWN — this session is governed by the vendored overlay at .trellis/internal/, and that is intact. What this hook could NOT do is check whether the overlay is stale: the installed Trellis plugin's own version stamp ($ref) $stamp_defect, so there is nothing to compare this project's stamp ($overlay) against. Nothing is wrong with this project and no rules are missing. Reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the likely fix."
    exit 0
  fi
  if [ "$overlay" != "$current" ]; then
    emit "Trellis overlay may be stale: this project's .trellis/internal/version stamp is $overlay, but the installed Trellis plugin ships $current. This project still carries a vendored overlay, which the plugin no longer writes or refreshes. To move it onto plugin-delivered rules, delete .trellis/internal/ and the managed block from this project's instructions file, keeping .trellis/rules.toml rows. Show the user the exact paths you would delete and get explicit confirmation before deleting anything (floor-intent-gate): this hook advises, it never authorises a deletion, and the files are tracked. Until then this session is governed by the vendored copy."
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
  payload_read "$legacy" && overlay="$(printf '%s\n' "$payload_text" | head -n1 | tr -d '[:space:]')"
  if [ -z "$overlay" ] || [ -z "$current" ]; then
    emit "Trellis overlay predates the .trellis/internal/ layout (decision-0051): its stamp sits at the legacy path .trellis/version. This hook could not read both stamps, so it cannot say how far behind this overlay is — but the LAYOUT itself is the stale part and the migration below is correct regardless. To migrate, delete the legacy overlay — .trellis/version, .trellis/trellis.md and .trellis/internal/ if present, plus the managed block from this project's instructions file — keeping your .trellis/rules.toml rows. An overlay this old may predate .trellis/rules.toml entirely; if there is none, copy $plugin/reference/rules-b.toml to $root/.trellis/rules.toml. Show the user the exact paths you would delete and get explicit confirmation before deleting anything (floor-intent-gate): this hook advises, it never authorises a deletion, and the files are tracked."
    exit 0
  fi
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
  emit "TRELLIS_INLINE_BLOCK — $inline_files carries a Trellis managed block (its trellis:begin marker at column 0), so this hook injected nothing. This project is in one of two states and the hook cannot tell which: if the rules readout is written out between the block's markers, the host already loaded those rules at launch and injecting here would put them in context twice; if the block holds only @-import lines whose .trellis/internal/ overlay was deleted, no rules are loaded and this session is ungoverned. Read each block named above to tell which. To move onto plugin-delivered rules either way, delete the managed block from EACH of $inline_files — everything from its trellis:begin marker through its trellis:end marker, in every file named; leaving one behind leaves this project in the same refused state — keeping .trellis/rules.toml rows if that file exists. Without it, read each block's own strictness value BEFORE deleting anything and copy the preset that matches — $plugin/reference/rules-a.toml for firm, rules-b.toml for adaptive — to $root/.trellis/rules.toml, so the project keeps the posture it had instead of silently becoming adaptive. If two blocks disagree on strictness, there is no posture to preserve: say so, show the user both values, and let them choose — never pick one silently. Or run /trellis:remove to take Trellis out of this project entirely — the opposite endpoint, not a migration. Show the user the exact lines you would delete and get explicit confirmation before deleting anything (floor-intent-gate): this hook advises, it never authorises a deletion, and the file is tracked. Tell the user before doing substantive work."
  exit 0
fi

# ---------------------------------------------------------------------- path B
toml="$root/.trellis/rules.toml"

# A path that EXISTS but is not a regular file — a directory, a FIFO, a socket,
# a device node — is neither a rules file nor a missing one, and every read
# below is guarded by `-f`, so without this it fell into the missing-file
# branch: a vendored bundle governed with the shipped defaults under a heading
# saying the project "has no .trellis/rules.toml", and a user-scope install
# announced the same and told the user to WRITE that file — over a FIFO, a
# write that blocks exactly as the read did. Both were wrong about the
# reader's state (decision-0073's class). Measured on the guarded hook before
# this check: `mkdir .trellis/rules.toml` drew TRELLIS_NOT_YET_GOVERNING and
# "has no .trellis/rules.toml". So it is refused, loudly, like the mode-000
# file the slug check catches further down — the same disposition as an
# unreadable regular file, said before anything tries to read it.
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
  emit "TRELLIS_RULES_NOT_LOADED — this project's .trellis/rules.toml exists but is not a regular file (a directory, a FIFO, a socket or a device node; \`ls -l $toml\` says which), so the Trellis plugin hook did not open it. This project is configured for Trellis: something sits at .trellis/rules.toml, but it is not a rules file and it is not a governed = false opt-out, so the session is running ungoverned and NO rules and NO rows were injected. Nothing on disk was changed. To govern this project, replace it with a regular file holding the rows (copy $plugin/reference/rules-b.toml there) or the single line governed = false to opt out — show the user the exact path you would replace and get explicit confirmation before removing anything (floor-intent-gate): this hook advises, it never authorises a deletion. Tell the user before doing substantive work."
  exit 0
fi

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
    # visibly, greppably, and revocably by deleting it. Absent rows therefore
    # mean the standard set, not none. Rather than invent a second activation
    # semantics, point at the shipped preset and let every check below run
    # unchanged: same slugs, every row active, strictness "adaptive" (posture B).
    toml="$plugin/reference/rules-b.toml"
    # TRL-33. This was `[ -f "$toml" ] || exit 0`. Measured on main: an ABSENT
    # rules-b.toml on this path produced completely empty stdout, exit 0, zero
    # bytes of stderr — a session governed by nothing, with no signal of any
    # kind — while the UNREADABLE sibling of the same file was caught, loudly,
    # two hundred lines downstream by a message that names .trellis/rules.toml
    # to a project that HAS NO .trellis/rules.toml. Absent-vs-unreadable was an
    # inconsistency, not a choice, and the loud half also named the wrong file.
    #
    # Reachability is the payload-header case's: rules-b.toml is a line of
    # install.sh's bundle manifest, so an interrupted install leaves exactly
    # this state.
    #
    # THE WORDING IS THE `default_rows` REFUSAL'S, deliberately. That check
    # still stands further down and catches the shape this one cannot — a file
    # that reads fine and parses to no rows — so both doors lead to the same
    # room, and a consumer who hits either is told the same true thing.
    if ! payload_read "$toml"; then
      emit "TRELLIS_RULES_NOT_LOADED — this project has no .trellis/rules.toml and is governed by the rule rows the Trellis plugin ships ($toml), and that file $payload_why. This project adopted Trellis (the plugin is vendored in this repository), but the session is running ungoverned and NO rules and NO rows were injected. The hook refused rather than treat a broken payload file as if it were this project's own settings. NOTHING is wrong with this project and there is nothing here to correct: reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the fix. Tell the user before doing substantive work."
      exit 0
    fi
    rows_are_default=yes
  else
    # D4, as corrected by decision-0077. A user-wide install is a broad choice,
    # and this says so in the project it is about to affect rather than assuming
    # consent it never asked for. Announce, inject NO rules on this turn ("will
    # be", not "is"), and name both answers.
    #
    # SILENCE IS NOT ONE OF THEM. 0070 D4 said an ignored prompt seeds the preset
    # ("accept, or no objection -> seed"); this code has never done that, in any
    # version since #218 built the record. An unanswered announcement leaves the
    # project ungoverned and recurs next session — which the message below states
    # in as many words. decision-0077 corrected the record to match the code
    # rather than the reverse, so that nothing is governed by silence.
    emit "TRELLIS_NOT_YET_GOVERNING — the Trellis plugin is installed outside this project (user scope, or a location this hook cannot place), so it applies to every project opened here, and $root has no .trellis/rules.toml. Tell the user, in your own words and before doing substantive work: \"Trellis is installed for your user account, so this repo will be governed by it — 16 rules, followed by default and deviations said out loud. Do you want to disable that for this repo?\" If they want it DISABLED, write .trellis/rules.toml containing exactly the line: governed = false — and nothing else. If they ACCEPT, copy $plugin/reference/rules-b.toml to $root/.trellis/rules.toml so the choice persists — without that file this same announcement repeats every session and the project is never governed. (That file is theirs to edit afterwards: strictness = \"firm\" for the by-the-book posture, active = false on a row to turn a rule off.) Inject and follow no Trellis rules this turn: none are active yet."
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
#
# THROUGH THE GATEWAY, both of them. This was a bare `[ ! -f ]` pair, and `-f`
# proves existence and never readability — the gap the posture header fell
# through four separate ways (mode 000, zero-byte, truncated, and the firm
# trellis-a.md), each measured, each shipping sixteen activation rows with zero
# rules prose under them at exit 0. The header is read ONCE, here, where a
# failure can still be reported; $header_prose is what the assembly below uses,
# so the fatal positional open that produced that damage is gone rather than
# merely guarded.
if ! payload_read "$rules"; then
  emit "TRELLIS_RULES_NOT_LOADED — the Trellis plugin hook ran but could not read its own rules payload: $rules $payload_why. This project is configured for Trellis: .trellis/rules.toml is present, but the session is running ungoverned and NO rules and NO rows were injected. This is a broken or half-written plugin install, not a problem with your rows: reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the likely fix. Tell the user before doing substantive work."
  exit 0
fi
if ! payload_read "$header"; then
  emit "TRELLIS_RULES_NOT_LOADED — the Trellis plugin hook ran but could not read the posture header it was about to inject: $header $payload_why. This project is configured for Trellis: .trellis/rules.toml is present, but the session is running ungoverned and NO rules and NO rows were injected. This is a broken or half-written plugin install, not a problem with your rows: reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the likely fix. Tell the user before doing substantive work."
  exit 0
fi
header_prose="$payload_text"

# The payload's own terminator, checked BEFORE its slugs are trusted -- the
# order codex-context.mjs uses and for the reason its comment gives: derive
# first and a broken rules.md yields a broken slug set, after which every
# downstream verdict is about the consumer's rows when the defect is the
# plugin's. rules.md is 39 lines with the sentinel on line 39, so a truncation
# anywhere loses it, and this catches the shape DIRECTLY rather than inferring
# it from a slug count.
#
# It is also the one payload check with no second-file dependency. The
# coherence check further down needs reference/rules-b.toml and skips itself
# when that file is absent -- measured, that skip reopens the full
# quarantine-fourteen-rows hole. This gate has nothing to skip on.
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
# COMPLETE, CORRECT payload. This branch already fixed that blindness once, in
# the reconciler, which strips \r for the same reason and only a few hundred
# lines away; this guard reintroduced the assumption. Every other check here is
# already CRLF-safe by accident rather than intent: the slug scans anchor on
# `[[:space:]]*$`, and \r is in [[:space:]] under the C locale. Only an exact
# string compare could break, and it did. Stripped rather than tolerated in the
# regex, so the comparison stays a comparison.
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
    emit "TRELLIS_RULES_NOT_LOADED — the Trellis plugin's own rules payload ($rules) is not complete: it must carry exactly one \`<!-- trellis:rules-loaded -->\` terminator as its final line, and this hook found \"$sentinel_report\". A payload cut short of that line is a truncated or half-written install, and every rule below the cut is simply absent — so its slug list cannot be trusted to say what the rule set is. This project is configured for Trellis: .trellis/rules.toml is present, but the session is running ungoverned and NO rows were injected; your rows were not judged against it and nothing on disk was changed. Reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the likely fix. Tell the user before doing substantive work."
    exit 0
    ;;
esac

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
      if (length(want) == 0) { print "no-slugs-in-payload"; exit }
      # EVERY category, not the first one an else-if chain reaches. A plugin
      # update that RENAMES a slug produces a missing new row and an unknown old
      # row at the same time; reporting only `missing:` sent the agent to add the
      # new one, and validation failed again next session on the old one it was
      # never told about. The remedy text below promises the report names the
      # repair, so a partial report makes that promise false.
      report = ""
      if (missing != "") report = report "missing:" missing "; "
      if (unknown != "") report = report "unknown:" unknown "; "
      if (dup != "") report = report "duplicate:" dup "; "
      if (report == "") print "ok"
      else { sub(/; $/, "", report); print report }
    }
  ' "$rules" "$toml"
)"
# Reconcile rather than refuse. A mismatch used to inject nothing at all, so a
# single bad row cost all sixteen rules every session until a human edited the
# file (TRL-20). The rules the payload ships are still the authority; what
# changes is that an unmatched row is quarantined instead of blocking delivery.
#
# Quarantine — commenting the row out rather than deleting it — is what makes
# an ungated repair safe. `unknown:` has two causes with opposite repairs (the
# rule was retired, or the installed plugin is behind the project's config,
# TRL-27) and config-only mode carries no version stamp to tell them apart. A
# commented row is correct under both readings and loses nothing either way.
# It is also invisible to the validator above, which anchors rows at line
# start, so a repaired file draws no second notice.
# `no-slugs-in-payload` is a different failure than a project's rows not
# matching a valid payload: it means the validator above found NOTHING to
# check rows against (the payload's own rules.md is unreadable or malformed),
# not that this project's rows are wrong. Reconciling against an empty want
# set would quarantine every legitimate row and run the session ungoverned
# with exit 0 — silently inverting the fail-loud invariant stated above
# ("Fail loudly rather than govern silently on a partial payload"). This is
# the same broken-plugin shape the header/rules file-existence check already
# fails loudly on, just caught one step later.
if [ "$slug_report" = "no-slugs-in-payload" ]; then
  emit "TRELLIS_RULES_NOT_LOADED — the Trellis plugin's own rules payload ($rules) carries no rule slugs for this hook to validate .trellis/rules.toml against. This project is configured for Trellis: .trellis/rules.toml is present, but the session is running ungoverned. This is a broken or unrecognisable plugin payload, not a problem with your rows — reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the likely fix, not editing .trellis/rules.toml. Tell the user before doing substantive work."
  exit 0
fi
# A report that is neither `ok` nor one of the three defect shapes is not a
# mismatch to reconcile — it is the validator itself having failed, and the
# EMPTY STRING is how that arrives. The awk above reads two files positionally,
# and a positional file that exists but cannot be OPENED (mode 000, a stale ACL,
# a dangling symlink target) is a fatal awk error: it prints nothing, so the
# command substitution captures "". Both operands can produce it -- an
# unreadable payload rules.md, or an unreadable project rules.toml, each of
# which passes the `-f` existence checks above.
#
# Read as a mismatch, "" walked straight into the reconciler with an empty want
# set: every legitimate row failed `row in want`, all of them were quarantined,
# and the hook delivered `added 0 row(s); quarantined N row(s)` plus a mandate
# to write that file to disk -- the session ungoverned at exit 0. That is
# precisely the hazard the no-slugs-in-payload note above names, reached
# through a different door, so it exits through the same one.
case "$slug_report" in
  ok|missing:*|unknown:*|duplicate:*) ;;
  *)
    emit "TRELLIS_RULES_NOT_LOADED — the Trellis plugin hook could not validate this project's .trellis/rules.toml: its row check produced no usable report (\"$slug_report\"), which means one of the two files it reads — the payload's rules.md ($rules) or $toml — exists but could not be read. This project is configured for Trellis: .trellis/rules.toml is present, but the session is running ungoverned, and NO rows were injected or reconciled. Check that both files are readable; reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the likely fix if the payload is the unreadable side. Tell the user before doing substantive work."
    exit 0
    ;;
esac
# ON THE DEFAULTS PATH THE PAYLOAD **IS** THE ROWS, and that inverts what an
# unusable row file means. When rows_are_default=yes (decision-0070 D3: the
# bundle is vendored in this repository, so absent rows mean the standard set)
# $toml was repointed at the payload's own rules-b.toml. There is no consumer
# file at all -- and the reconciler, reached with zero rows parsed, reads all
# sixteen slugs as missing, adds them, and mandates writing .trellis/rules.toml
# into a project that never had one. Measured against a 0-byte rules-b.toml:
# exit 0, sixteen rows, and "Write .trellis/rules.toml with exactly the rows
# shown above" -- a broken payload driving a write into the consumer's tree,
# which is the persists-damage class the comment below this one exists for.
#
# It reached that path through the coherence gate skipping itself, correctly,
# on `length(rows) == 0`: that gate treats rules-b.toml as the COMPARISON file,
# and skipping an unusable comparison is right when the rows come from
# somewhere else. Here they do not. So the row count is checked on $toml
# directly rather than inferred from the gate below, which keeps this correct
# even if the defaults path is ever pointed at a different preset file.
#
# Only the defaults path. An empty PROJECT rules.toml is the supported
# hand-written-partial shape and must keep reconciling into the consumer's own
# file -- the difference is whose file the mandate names.
if [ "${rows_are_default:-no}" = yes ]; then
  default_rows="$(awk '/^[[:space:]]*(inv|floor)-[a-z-]+[[:space:]]*=/ { n++ } END { print n + 0 }' "$toml")"
  case "${default_rows:-0}" in
    0)
      emit "TRELLIS_RULES_NOT_LOADED — this project has no .trellis/rules.toml and is governed by the rule rows the Trellis plugin ships ($toml), and that file carries no rows at all — it is empty, unreadable, or truncated. This project adopted Trellis (the plugin is vendored in this repository), but the session is running ungoverned and NO rules and NO rows were injected. The hook refused rather than treat a broken payload file as if it were this project's own settings. NOTHING is wrong with this project and there is nothing here to correct: reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the fix. Tell the user before doing substantive work."
      exit 0
      ;;
  esac
fi
# PAYLOAD-VS-PAYLOAD, not payload-vs-project. Everything above this line checks
# the PROJECT's rows against the payload; nothing checked the payload against
# ITSELF, and `length(want) == 0` is the only shape of broken rules.md the
# validator can see. A rules.md truncated BELOW its first slug is non-empty, so
# it passes that test and is then treated as authoritative -- measured with a
# 9-line, 2-slug payload, the hook reported `quarantined 14 row(s)`, commented
# out BOTH floor rules, and instructed the agent to write that file to
# .trellis/rules.toml. Exit 0, no loud marker.
#
# That is worse in kind than the blackouts above rather than another of them.
# Those WITHHELD governance for a session; this one PERSISTS DAMAGE: a broken
# payload drives a mandate to comment out fourteen rules in the consumer's own
# file, while the whole safety argument for reconciling without a gate is that
# a repair loses nothing.
#
# The check is possible because the payload ships two independent statements of
# the same set: rules.md tags sixteen slugs and reference/rules-b.toml carries
# sixteen rows, and they are IDENTICAL by construction. A payload whose own two
# halves disagree is provable internal corruption -- and it is exactly what
# quarantine cannot be allowed to act on, because it is distinguishable from
# the stale-plugin case quarantine legitimately exists to handle (there the
# payload is coherent and the PROJECT is out of step). So the disagreement is
# refused loudly and the reconciler is never reached.
#
# Placed AFTER the report classification deliberately: an unreadable rules.md
# already exits above with a message that names the read failure, which is a
# better diagnosis than "the payload disagrees with itself".
#
# Skipped, not failed, when the preset is unusable: a payload without it offers
# nothing to compare against, which is where this hook already stood. That skip
# is the ONE payload read in this file whose failure is deliberately silent, and
# it stays that way -- an over-correction here is what
# TestAnUnusablePresetSkipsTheCoherenceCheckRatherThanBlackingOut exists to
# catch, and nothing is wrong for the consumer when only the COMPARISON file is
# broken. The gateway is used anyway, so the silence is a stated disposition at
# a checked read rather than the by-product of a bare `-f`. (The other silent
# exit this comment used to point at -- rules-b.toml missing on the DEFAULTS
# path, where the same file is the rows rather than the comparison -- was
# TRL-33, and is now a loud refusal.)
preset="$plugin/reference/rules-b.toml"
if payload_read "$preset"; then
  coherence="$(
    awk '
      FNR == NR {
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
        rows[row] = 1
      }
      END {
        # NOTHING TO COMPARE is a third answer, and collapsing it into
        # "they disagree" was a false blackout on a healthy payload: an
        # EMPTY rules-b.toml yielded 16 want vs 0 rows and read as corruption.
        if (length(want) == 0 || length(rows) == 0) { print "incomparable"; exit }
        for (s in want) if (!(s in rows)) d++
        for (s in rows) if (!(s in want)) d++
        printf "%d %d %d\n", length(want), length(rows), d + 0
      }
    ' "$rules" "$preset"
  )"
  # `-f` proves the file EXISTS, never that it can be READ, and the difference
  # was inverted here: an ABSENT rules-b.toml skipped the check and governed
  # normally, while an UNREADABLE or EMPTY one produced a full
  # TRELLIS_RULES_NOT_LOADED blaming payload incoherence -- with rules.md and
  # the project rows both perfectly healthy. The more broken state was handled
  # better than the less broken one.
  #
  # A guard that cannot tell "I could not read this" from "this is corrupt" is
  # not a guard. So the two are separated: an empty capture means the awk died
  # on the positional read, `incomparable` means it ran and found no rows, and
  # both mean the same thing this check already does for an absent file -- skip
  # it. Silently, because nothing is wrong for the consumer: the terminator gate
  # above is the unconditional half of this pair and needs no second file, so
  # skipping here loses the narrower case only (a rules.md whose slug list is
  # wrong while its ending is intact).
  case "$coherence" in
    "" | incomparable) ;;
    *" 0") ;;
    *)
      coherence_rest="${coherence#* }"
      # NO SLUG NAMES in the message below. It is a payload defect, the reader
      # can do nothing with the list, and the loud paths are pinned by tests
      # that assert no rule slug appears anywhere in a refusal.
      #
      # And the word "preset" cannot appear in it either: the destructive-verb
      # scan in cli/plugin_hook_test.go matches SUBSTRINGS, so "p-reset" hits
      # `reset` and demands a confirmation gate on a message that instructs no
      # mutation at all. Erring safe is the right default for that guard, so the
      # wording moves rather than the guard.
      emit "TRELLIS_RULES_NOT_LOADED — the Trellis plugin's own payload is internally inconsistent: its rules.md and the rules-b.toml default row list it ships alongside do not describe the same rule set (rules.md slugs: ${coherence%% *}; rules-b.toml rows: ${coherence_rest%% *}; named in one but not the other: ${coherence_rest##* }). Those two files ship together and are identical by construction, so this is a truncated or corrupted plugin payload, not a problem with your rows. This project is configured for Trellis: .trellis/rules.toml is present, but the session is running ungoverned and NO rows were injected — your file was not reconciled against a payload that cannot be trusted to say what the rule set is, and nothing on disk was changed. Reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the likely fix. Tell the user before doing substantive work."
      exit 0
      ;;
  esac
fi
reconciled=""
repair_summary=""
# `no-slugs-in-payload` already exited above, and the case just proved the
# report is one of the four well-formed values, so `ok` is the only non-defect.
if [ "$slug_report" != "ok" ]; then
  today="$(date +%Y-%m-%d)"
  # The quarantine note is written INTO THE CONSUMER'S OWN FILE and reads
  # "not in <stamp>". With an unreadable reference/version, $current is the
  # empty string and that note used to read "not in ." — a broken sentence
  # persisted into a tracked file, silently. A named fallback keeps the line a
  # sentence and keeps it TRUE; the rules themselves are fine, so this
  # degrades the provenance rather than the delivery.
  note_stamp="$current"
  [ -n "$note_stamp" ] || note_stamp="the installed Trellis payload"
  reconciled="$(
    TRELLIS_WANT_SRC="$rules" awk -v stamp="$note_stamp" -v today="$today" '
      BEGIN {
        # ENVIRON, not -v, for the same reason as the assembly below: -v
        # escape-processes its value, so a plugin root containing a backslash
        # arrived here as a path that does not exist. The rc check under this
        # made that fail LOUDLY rather than silently, which was right for a
        # broken payload and wrong for a legitimate root that merely has a
        # backslash in it. Now it simply works, and rc guards real read failures.
        want_src = ENVIRON["TRELLIS_WANT_SRC"]
        # A REDIRECTED getline is silent where a positional read is fatal: it
        # returns -1 when the file cannot be opened and 0 at EOF, and the plain
        # `> 0` test could not tell those apart from a file that simply held no
        # slugs. So an unreadable want_src left want[] empty and quarantined
        # every row at exit 0. Keep the return value and refuse to reconcile
        # against nothing -- the guard above should already have caught this,
        # and this is the second lock on the same door.
        rc = 0
        while ((rc = (getline line < want_src)) > 0) {
          if (match(line, /`(inv|floor)-[a-z-]+`[[:space:]]*$/)) {
            s = substr(line, RSTART + 1, RLENGTH - 2)
            sub(/`[[:space:]]*$/, "", s)
            want[s] = 1
            order[++n] = s
          }
        }
        if (rc < 0 || n == 0) { no_want_set = 1; exit 1 }
        note = "  # quarantined " today ": not in " stamp ". If a newer Trellis" \
               " release ships this slug, update the Trellis plugin and uncomment this row."
      }
      # The default record separator here is "\n" alone, so a CRLF-terminated
      # line arrives with its "\r" still attached to $0 -- stripped BEFORE any
      # other rule reads $0, so every rule below (the header check, the row
      # check, the bare passthrough) sees the same CR-free line the Codex
      # hook reconciler produces (it splits on /\r?\n/, which consumes a CRLF
      # pair as one delimiter and never leaves a trailing \r on a line). Left
      # unstripped, `print "# " $0 note` on a quarantined CRLF row emitted a
      # bare CR MID-LINE, before the note -- measured directly against this
      # block. This is host parity, not a cosmetic fix: the two reconcilers
      # must produce the same bytes from the same input, and only one of them
      # was doing that.
      { sub(/\r$/, "") }
      # A row can be appended below with no `[rules]` table preceding it in the
      # file at all (the hand-written-partial shape: just strictness, no rows).
      # parseRulesToml in codex-context.mjs only accepts inv-/floor- keys INSIDE
      # `[rules]` — outside it, any key but seeded_from/strictness/governed is a
      # fatal invalid-rules on Codex while Claude governs normally from the same
      # file. Track whether the file already opens the table so the END block
      # can open one itself before appending, rather than assume it is there.
      #
      # Leading whitespace is tolerated here for the same host-parity reason,
      # and it was not at first: parseRulesToml trims each line before matching
      # its section regex (codex-context.mjs), so Codex reads an INDENTED
      # `  [rules]` as opening the table while an anchored match here did not.
      # A file with an indented table plus any missing row therefore had a
      # SECOND `[rules]` appended below -- and a second table header is
      # precisely what parseRulesToml rejects, so the repaired file read
      # invalid-rules on Codex. Nothing was lost (such a file was already
      # Codex-invalid before the repair) but the mandate promises the written
      # file matches what governs, and a file Codex refuses does not. Matching
      # the row regex one line down, which was already whitespace-tolerant.
      /^[[:space:]]*\[rules\][[:space:]]*(#.*)?$/ { has_rules = 1 }
      /^[[:space:]]*(inv|floor)-[a-z-]+[[:space:]]*=/ {
        row = $1
        sub(/[^a-z-].*$/, "", row)
        if (!(row in want) || (row in seen)) {
          print "# " $0 note
          quarantined++
          next
        }
        seen[row] = 1
        print
        next
      }
      { print }
      # ONE header comment above the whole appended block, not one per row
      # (Ruling 6, TRL-20 task 3). MAX_CONTEXT_BYTES on the Codex hook leaves
      # only about 176 bytes of headroom over the plain firm preset, and the
      # stamp-and-date note repeated on all sixteen rows -- the
      # hand-written-partial worst case -- overran it: a reconciled Codex
      # session hit its own budget and injected NOTHING, reintroducing on
      # Codex exactly the blackout this whole change removes on Claude. The
      # provenance (date, stamp, count) is unchanged; it is stated once
      # instead of sixteen times. Quarantine notes stay per-row (unchanged):
      # this is the shape review measured as the actual overrun.
      # NO APOSTROPHES ABOVE OR BELOW, to the end of this awk program: it is
      # single-quoted at the shell level, and one would close the quote early.
      END {
        # `exit` in BEGIN still runs END, so the marker is printed here rather
        # than there; the shell below turns it into a loud refusal.
        if (no_want_set) { print "#trellis-reconcile-no-want-set"; exit 1 }
        missing_n = 0
        for (i = 1; i <= n; i++) {
          s = order[i]
          if (!(s in seen)) { missing[++missing_n] = s }
        }
        if (missing_n > 0) {
          if (!has_rules) { print "[rules]"; has_rules = 1 }
          print "# added " missing_n " row(s) below on " today " (missing from " stamp ")"
          for (i = 1; i <= missing_n; i++) print missing[i] " = { active = true }"
        }
        # THIS run-s counts, stated by the code that did the work. Peeled off
        # by the shell below and never delivered. See the note there for why
        # the summary cannot be recovered from the text.
        print "#trellis-reconcile-counts " (missing_n + 0) " " (quarantined + 0)
      }
    ' "$toml"
  )"
  # No rows, or the marker the BEGIN block prints when it had nothing to
  # reconcile against. Either way there is no reconciled set to govern from,
  # and delivering the heading with an empty block under it is the blackout
  # this whole guard chain exists to prevent.
  case "$reconciled" in
    "" | "#trellis-reconcile-no-want-set")
      emit "TRELLIS_RULES_NOT_LOADED — the Trellis plugin hook tried to reconcile this project's .trellis/rules.toml against the rules the payload ships ($rules) and could not read that payload, so there was nothing to reconcile against. This project is configured for Trellis: .trellis/rules.toml is present, but the session is running ungoverned, and NO rows were injected — your rows were NOT quarantined and nothing was changed on disk. Reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the likely fix, not editing .trellis/rules.toml. Tell the user before doing substantive work."
      exit 0
      ;;
  esac
  # The summary must describe THIS session, so awk states its own counts on a
  # trailer line that is peeled off here. Counting the reconciled text instead
  # was one source of truth but the wrong one: quarantine notes and the
  # `# added N row(s)` header are PERSISTED provenance, so a partially repaired
  # file already carries earlier sessions' marks and a text count adds them to
  # this run's. Measured on the pre-fix hook: a file with one old quarantine
  # line and one old added-header, plus one further missing row, reported
  # "added 2 row(s); quarantined 1 row(s)" for a session that added 1 and
  # quarantined 0. The in-file provenance was right either way; the SPOKEN
  # summary was not — and that summary is what the agent reports to the user,
  # which is the whole channel this change is built to make trustworthy.
  counts="$(printf '%s\n' "$reconciled" | sed -n '$p')"
  case "$counts" in
    '#trellis-reconcile-counts '*)
      reconciled="$(printf '%s\n' "$reconciled" | sed '$d')"
      added="${counts#* }"
      added="${added%% *}"
      quarantined="${counts##* }"
      ;;
    *)
      # The trailer is unconditional in the END block above, so this is
      # unreachable short of awk failing outright. Never strip a line that is
      # not the trailer — a delivered row is worth more than a count — and say
      # the count is unknown rather than assert one that was not counted.
      added="unreported"
      quarantined="unreported"
      ;;
  esac
  repair_summary="added ${added} row(s); quarantined ${quarantined} row(s)"
fi

# Delivery reads $toml directly when nothing was reconciled, and that read sits
# inside the payload command substitution below, where a failing `cat` writes
# nothing and changes no exit status anybody looks at: the payload would ship
# the "Rows from this project's .trellis/rules.toml" heading with zero rows
# under it and no warning at all. Probe the read here instead, where a failure
# can still be reported. Belt-and-braces -- an unreadable $toml already fails
# the validator guard above -- but the silent-empty shape is exactly the defect
# class this change is closing, so it does not get to survive anywhere.
if [ -z "$reconciled" ] && ! cat "$toml" >/dev/null 2>&1; then
  emit "TRELLIS_RULES_NOT_LOADED — the Trellis plugin hook could not read the rule rows it was about to inject ($toml). This project is configured for Trellis: that file is present, but it could not be read, so the session is running ungoverned and NO rows were injected. Check the file's permissions. Tell the user before doing substantive work."
  exit 0
fi

# The header carries `@rules.md`, an import the hook resolves itself, and a
# pointer at the vendored invariants path, which does not exist in this mode.
# Repointing it at the plugin keeps the trigger-read affordance and cannot go
# stale, because it names the payload this session is actually running.
#
# The assembly awk below used to read $header POSITIONALLY -- the same construct
# as the validator above, behind the same bare `-f` existence check -- and that
# was the worst-looking member of this whole family. A $header that EXISTS but
# yields nothing dies fatally and prints nothing, while the printfs and the row
# block around it carry on: measured four ways (mode 000, zero-byte, truncated,
# and the firm-posture trellis-a.md), the hook emitted sixteen activation rows,
# ZERO rules prose, no loud marker, and exit 0. That is more dangerous than the
# two blackouts above rather than less, because the payload looks substantive
# and nothing signals a problem -- the agent is told which sixteen rules are
# active and handed none of them. It needs no permission trickery either: a
# header left truncated by an interrupted install.sh is enough.
#
# The Codex hook has always refused exactly this -- readRequired reports
# unreadable-file/missing-file, and an explicit check rejects empty prose
# (codex-context.mjs) -- so the Claude-side gap was an oversight, not a design
# choice. Match it. The header is read ONCE, here, where a failure can still be
# reported, and the assembly reads that text from stdin, so the fatal positional
# open is gone rather than merely guarded.
#
# Exactly one `@rules.md` import is required for the same reason Codex rejects
# invalid-placeholder-count: a header truncated ABOVE that line is non-empty and
# still assembles into rows with no rules under them, which is the identical
# blackout reached through a shorter truncation. Counted with awk over stdin --
# never a positional read, which is the failure being closed here.
# $header_prose was read through payload_read at the top of path B, which is
# also where emptiness is now refused; the `-z` test below is kept as the second
# lock on the same door rather than as the only one.
header_imports="$(printf '%s\n' "$header_prose" |
  awk '/^@rules\.md[[:space:]]*$/ { n++ } END { print n + 0 }')"
if [ -z "$header_prose" ] || [ "$header_imports" != "1" ]; then
  emit "TRELLIS_RULES_NOT_LOADED — the Trellis plugin hook could not assemble its own rules payload: the posture header it was about to inject ($header) read as empty, or carries ${header_imports} @rules.md imports where exactly one is required, so the rules themselves would have been missing from what was injected. This project is configured for Trellis: .trellis/rules.toml is present, but the session is running ungoverned and NO rules and NO rows were injected — the hook refused rather than deliver a rule ACTIVATION list with no rules under it. This is a broken or half-written plugin payload, not a problem with your rows: reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the likely fix. Tell the user before doing substantive work."
  exit 0
fi
# The `@rules.md` expansion is the TWIN of the reconciler getline this branch
# already fixed -- same redirected read, same silent -1 on a failed open -- left
# unguarded on the `ok` path while its sibling 186 lines up checks `rc`. A
# failed open printed ZERO rules prose and carried on.
#
# What triggers it is the `-v` channel, not a permission: `awk -v`
# ESCAPE-PROCESSES its value (`awk -v v=/a\tb/c` yields length 6, not 7), so a
# CLAUDE_PLUGIN_ROOT containing a backslash reaches this awk as a DIFFERENT path
# than the one every `-f` test and every positional read used -- all of which
# pass. Measured with a root named `plug\tools`: 16 activation rows, 0 rules
# prose, exit 0, no marker, which is verbatim the damage shape the posture-header
# guard above exists to stop. The same root with a mismatched rules.toml already
# refused loudly, from the sibling that checks `rc`.
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
rules_prose="$(
  printf '%s\n' "$header_prose" |
  TRELLIS_RULES="$rules" TRELLIS_INV="$plugin/reference/invariants.md" awk '
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
    emit "TRELLIS_RULES_NOT_LOADED — the Trellis plugin hook assembled its posture header but could not import the rules themselves from $rules, so what it was about to inject named which rules are active without saying what any of them are. This project is configured for Trellis: .trellis/rules.toml is present, but the session is running ungoverned and NO rules and NO rows were injected — the hook refused rather than deliver an activation list with nothing under it. This is a broken or half-written plugin payload: reinstalling or updating the plugin (\`claude plugin update trellis@kodhama\`) is the likely fix. Tell the user before doing substantive work."
    exit 0
    ;;
esac
payload="$(
  printf '%s\n' "$rules_prose"
  printf '\n## Project rule activation\n\n'
  if [ -n "$reconciled" ]; then
    printf 'Rows from this project'"'"'s .trellis/rules.toml, RECONCILED against the rules this payload ships (%s) — the file on disk still differs. Apply a rule only when its row says active = true; the two floor rules apply regardless of their row.\n\n' "$repair_summary"
  elif [ "${rows_are_default:-no}" = yes ]; then
    printf 'Rows: this project has no .trellis/rules.toml, so these are the shipped defaults — every rule active, adaptive posture (decision-0070). Write .trellis/rules.toml to change them. Apply a rule only when its row says active = true; the two floor rules apply regardless of their row.\n\n'
  else
    printf 'Rows from this project'"'"'s .trellis/rules.toml. Apply a rule only when its row says active = true; the two floor rules apply regardless of their row.\n\n'
  fi
  if [ -n "$reconciled" ]; then
    printf '%s\n' "$reconciled"
  else
    cat "$toml"
  fi
  if [ -n "$reconciled" ]; then
    printf '\n## Rule activation was reconciled this session\n\n'
    printf 'This project'"'"'s .trellis/rules.toml did not match the rules this payload ships (%s). The rows above are the reconciled set and are what governs this session; the file on disk still differs. Reconciliation: %s.\n\n' "$slug_report" "$repair_summary"
    printf 'Write .trellis/rules.toml with exactly the rows shown above, so the file matches what governs. Nothing is lost by this: a row the payload does not ship is commented out with the reason and the date, its line kept rather than taken out, and every value the project chose is preserved verbatim. Tell the user what you reconciled, row by row, before doing substantive work — a repair they did not see is the failure this reconciliation exists to prevent. If a quarantined slug is one a newer Trellis release added, the installed plugin is the stale side: `claude plugin update trellis@kodhama`, restart the session, and uncomment the row.\n'
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
    printf '\nDelivered by the Trellis plugin. Its own version stamp could not be read (%s %s), so this readout cannot name which payload build it came from; the rules and rows above are complete and govern this session. No overlay is vendored in this project.\n' "$ref" "$stamp_defect"
  fi
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
