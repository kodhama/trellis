#!/bin/sh
# install.sh — vendor the Trellis Claude Code plugin onto disk as a skills-directory
# plugin (kodhama/trellis#124, corrected design per spec-0005; supersedes the closed
# #128 attempt — see #128's own closing comment). This is NOT the retired end-user
# binary installer (kodhama-0007 rule 5, decision-0043 §4 — see the note appended
# there): it downloads no binary and, more importantly, makes exactly ONE decision
# (where to put the plugin) and composes NOTHING else. Every other decision —
# posture, which instructions file to patch, block style, hand-authored-content
# guarding — stays entirely inside plugins/trellis/skills/setup/SKILL.md, unmodified
# and identical whether the plugin arrived via marketplace, a pre-committed
# skills-dir vendor (this script), or the manual copy path. A second independent
# writer of that skill's *decision logic* is exactly the drift-risk class
# kodhama-0007 exists to close; this script is a mechanical copier of the plugin
# bundle only, same shape as the setup skill's own "copy, paste, verify"
# (kodhama-0007 rule 2) but one layer further out — it vends the *plugin*, not the
# *overlay* the plugin's skill later writes.
#
# MECHANISM (code.claude.com/docs/en/plugins-reference, "Skills-directory plugins" —
# fetch that doc yourself to confirm; summarized here for the header, not restated as
# a second source of truth). Any folder under a skills directory containing a
# .claude-plugin/plugin.json manifest loads as <name>@skills-dir on Claude Code's next
# session — no marketplace, no install step, discovered in place. Two scopes:
#   project  (default) — <repo-root>/.claude/skills/trellis/   checked into git,
#            reaches every collaborator on clone; gated by Claude Code's own
#            workspace trust dialog on first launch (unavoidable — this script just
#            tells you it's coming). Project-scope skills-directory plugins do NOT
#            walk up to the repo root the way plain skills/commands do, so this
#            script resolves the target via `git rev-parse --show-toplevel` from the
#            invocation directory, never $PWD — landing anywhere else would make
#            Claude Code silently fail to find the plugin when launched from root.
#   personal — ~/.claude/skills/trellis/   available in every project, no trust
#            dialog, no repo required, and (opt-in only, via --scope/env) never
#            even shells out to git.
#
#   curl -fsSL https://raw.githubusercontent.com/kodhama/trellis/main/install.sh | sh
#
# Inspect first, or pass flags:
#
#   curl -fsSLO https://raw.githubusercontent.com/kodhama/trellis/main/install.sh
#   sh install.sh --scope project
#
# WHAT THIS SCRIPT DOES, AND NOTHING MORE: resolves a scope (the one decision it
# makes), fetches the whole plugins/trellis/ tree, verifies every byte against the
# manifest baked in below, and writes it to the resolved scope directory
# (overwriting the plugin's own prior files on a re-run — same idempotent-artifact
# principle as the rest of this family). On PROJECT scope it additionally renders
# one file it wholly owns, .claude/rules/trellis.md, from bundle bytes — that is
# how the rules actually reach a session, since a vendored bundle alone delivers
# none (decision-0068; issue #201). It touches a project's .trellis/ at exactly ONE
# point — seeding .trellis/rules.toml from the shipped preset when none exists
# (decision-0070 D2), never overwriting; otherwise .trellis/ is the consumer's
# to edit (decision-0072 retired the setup skill). It DOES therefore pick a
# posture, the adaptive one, by copying
# rules-b.toml: that is a shipped constant, not a decision this script makes, but
# the header used to claim "no posture choice" full stop and that reads as false
# next to the seed. It reads no project file to decide anything, and it NEVER runs
# a git
# command that mutates anything (no add, no commit): it prints a suggested next
# command for project scope and leaves the commit to you.
#
# SCOPE RESOLUTION IS FAIL-CLOSED, NEVER A SILENT SUBSTITUTION (spec-0005 AC5).
# Outside a git repo, with no --scope/$TRELLIS_SKILLS_SCOPE given, project scope has
# no target: if a controlling tty is available, this script prompts once (offer
# personal scope, or abort); if none is available it exits non-zero immediately,
# naming exactly what's missing, and writes nothing. It never silently substitutes
# personal scope for an unresolvable project default — that would be exactly the
# "surprising, unrequested target" failure mode this family's discipline argues
# against everywhere else. (This corrects an earlier, wrong reading of the original
# issue brief, which asked for a silent fallback here; the spec's fail-closed
# requirement is the one this script implements.)
#
# BUNDLE INTEGRITY. TRELLIS_BUNDLE_MANIFEST below is a full sha256 manifest of every
# file under plugins/trellis/ as of this script's own commit, baked in literally.
# There is no existing shipped manifest that covers the whole bundle to lean on
# instead: plugins/trellis/reference/checksums covers only the 14 rendered M1 payload
# files (kodhama-0007 rule 1/3), not .claude-plugin/plugin.json, hooks/, skills/, or
# README.md — extending that manifest would mean teaching the release-time payload
# generator (cli/payload.go) about a second, non-rendered content class it has no
# other reason to know about, a bigger and more invasive change than this issue's
# scope. So this script carries its own manifest, generated once from the actual
# bundle and guarded for staleness the same way the payload pin was guarded in the
# retired binary's install.sh (regenerate-and-diff in CI, not by hand).
#   Fetch transport is raw @ main (a moving ref) rather than a pinned commit sha,
# deliberately: a sha pin would have to name a commit that does not exist yet at the
# time this very commit is authored (this script ships IN that commit). Pinning the
# manifest content instead — verified regardless of transport — sidesteps that
# chicken-and-egg problem while still giving the same guarantee: a bundle that has
# moved past what this copy of the script expects fails closed instead of installing
# something unverified, with a clear message to re-download. (A specific pinned
# commit sha fetched over HTTPS would also be a valid content-integrity mechanism —
# GitHub's TLS cert plus git's own content-addressing already guarantee those exact
# bytes — but it doesn't solve the chicken-and-egg problem above without a follow-up
# commit, so this script does not rely on it alone; the explicit per-file manifest
# below is the belt-and-suspenders check that also makes the corrupted-fetch case
# mechanically testable offline.)
#   HOW THE MANIFEST ADVANCES: cli/install_script_test.go
# TestInstallScriptBundleManifestIsCurrent regenerates it from plugins/trellis/ on
# disk and fails whenever this script's copy differs in content OR file set — script
# and bundle move atomically on main.
#
# Dependencies: POSIX sh, awk, grep, cp, mkdir, mktemp, dirname; curl for the default
# remote source (irrelevant if $TRELLIS_BUNDLE_SOURCE points at a local directory);
# shasum or sha256sum. git only to resolve project scope's target directory, or to
# detect whether one is available at all when scope is otherwise ambiguous — an
# explicit `--scope personal` (or $TRELLIS_SKILLS_SCOPE=personal) never shells out to
# git at all. No binary, no network beyond the bundle fetch.

set -eu

say()  { printf 'trellis: %s\n' "$*"; }
fail() { printf 'trellis: FAIL: %s\n' "$*" >&2; exit 1; }

usage() {
  cat <<'EOF'
install.sh — vendor the Trellis Claude Code plugin onto disk (skills-directory install).

  curl -fsSL https://raw.githubusercontent.com/kodhama/trellis/main/install.sh | sh
  sh install.sh [--scope personal|project] [--non-interactive]

This is the ONLY decision this script makes. The posture ships as a constant
(the adaptive preset) and is yours to change afterwards by editing
.trellis/rules.toml; see the "next steps" this script prints when it finishes.

Flags:
  --scope personal|project   where to vendor the plugin. Also settable via
                              $TRELLIS_SKILLS_SCOPE (the flag wins if both are given).
                                project  — <repo-root>/.claude/skills/trellis
                                           (checked into git, reaches collaborators
                                           on clone; the default when run inside a
                                           git repo)
                                personal — ~/.claude/skills/trellis
                                           (every project on this machine; never
                                           requires git at all when passed explicitly)
  --non-interactive           never prompt, even if a terminal is available
                              (automatic already when none is). Outside a git repo
                              with no scope given, this makes an ambiguous scope a
                              hard failure instead of a prompt — see below.
  --help                      this text.

Scope resolution when nothing is given explicitly:
  - Inside a git repo: defaults to project scope, no prompt.
  - Outside a git repo: project scope has no target. If a terminal is available,
    you are prompted once (offered personal scope, or the chance to abort). If not
    (CI, a plain curl|sh pipe with no controlling tty, or --non-interactive), this
    is a hard failure — nothing is written, and the exact missing input is named.
    Pass --scope personal (or $TRELLIS_SKILLS_SCOPE=personal) to avoid the prompt
    or the failure and go straight to personal scope.

Environment:
  TRELLIS_SKILLS_SCOPE   same as --scope; the flag takes precedence if both are set.
  TRELLIS_BUNDLE_SOURCE  alternate bundle source (an https:// URL or a local
                         directory laid out like plugins/trellis/) — verification
                         stays rooted in the manifest baked into this script
                         regardless of source.
EOF
}

SCOPE_FLAG=""
SCOPE_GIVEN=0   # "was the flag PRESENT", separate from "is it non-empty"
NONINTERACTIVE=0
while [ $# -gt 0 ]; do
  case "$1" in
    --scope)     [ $# -ge 2 ] || fail "--scope needs a value (personal or project)"; SCOPE_FLAG="$2"; SCOPE_GIVEN=1; shift ;;
    --scope=*)   SCOPE_FLAG="${1#--scope=}"; SCOPE_GIVEN=1 ;;
    --non-interactive) NONINTERACTIVE=1 ;;
    --help|-h)   usage; exit 0 ;;
    *)           fail "unknown flag: $1 (see --help)" ;;
  esac
  shift
done

# Resolve + validate the *requested* scope (if any) up front — a pure local check,
# so a bad --scope/env value fails instantly, before any network fetch or git call.
requested=""
requested_origin=""
if [ "$SCOPE_GIVEN" -eq 1 ]; then
  # Presence, not emptiness. `--scope ""` used to fall straight through to the
  # default, silently ignoring a flag the user explicitly passed. The load-bearing
  # check is the validator below, which also gates on SCOPE_GIVEN — mutation
  # showed this branch alone is not what enforces it.
  requested="$SCOPE_FLAG"; requested_origin="--scope"
elif [ -n "${TRELLIS_SKILLS_SCOPE:-}" ]; then
  requested="$TRELLIS_SKILLS_SCOPE"; requested_origin="\$TRELLIS_SKILLS_SCOPE"
fi
if [ "$SCOPE_GIVEN" -eq 1 ] || [ -n "$requested" ]; then
  case "$requested" in
    personal|project) ;;
    *) fail "scope must be personal or project, got: $requested (from $requested_origin)" ;;
  esac
fi

can_prompt() {
  [ "$NONINTERACTIVE" -eq 0 ] || return 1
  ( : </dev/tty ) 2>/dev/null || return 1
}

# --- 1. Scope — the one decision this script makes, resolved before any fetch ----
#         (so an unresolvable scope fails, or a decline-to-prompt aborts, before
#         doing any network or filesystem work at all)

if [ -n "$requested" ]; then
  scope="$requested"
  scope_origin="from $requested_origin"
  if [ "$scope" = "project" ]; then
    git_root="$(git rev-parse --show-toplevel 2>/dev/null)" \
      || fail "project scope was requested ($scope_origin), but the current directory is not inside a git repository. Re-run from inside a git repo, or pass --scope personal (or TRELLIS_SKILLS_SCOPE=personal)."
  fi
  # explicit personal scope: no git invocation at all, by design (see header).
else
  git_root=""
  repo=0
  if git_root="$(git rev-parse --show-toplevel 2>/dev/null)"; then repo=1; fi

  if [ "$repo" -eq 1 ] && can_prompt; then
    {
      printf '\nVendor the Trellis plugin at which scope?\n'
      printf '  1) project  — %s/.claude/skills/trellis (checked into this repo, reaches collaborators; default)\n' "$git_root"
      printf '  2) personal — %s/.claude/skills/trellis (every project on this machine)\n' "$HOME"
      printf 'Scope [1/2, Enter=project]: '
    } >/dev/tty
    read -r ans </dev/tty || ans=""
    case "$ans" in
      2) scope=personal ;;
      1|"") scope=project ;;
      *) fail "unrecognized scope answer: $ans (expected 1 or 2)" ;;
    esac
    scope_origin="prompted"
  elif [ "$repo" -eq 1 ]; then
    scope=project
    scope_origin="default (non-interactive, run inside a git repo)"
  elif can_prompt; then
    # Not inside a git repo: project scope has no target. Ask, rather than assume.
    {
      printf '\nNot inside a git repository — project scope needs one and has no target here.\n'
      printf 'Vendor the Trellis plugin at personal scope (%s/.claude/skills/trellis) instead? [Y/n]: ' "$HOME"
    } >/dev/tty
    read -r ans </dev/tty || ans=""
    case "$ans" in
      y|Y|"") scope=personal; scope_origin="prompted (not inside a git repository; personal scope confirmed)" ;;
      n|N)    fail "aborted at your request: not inside a git repository, and you declined personal scope. Nothing was written. Pass --scope personal (or TRELLIS_SKILLS_SCOPE=personal) to vendor globally without asking, or re-run inside a git repo for project scope." ;;
      *)      fail "unrecognized answer: $ans (expected y or n)" ;;
    esac
  else
    # No repo, no explicit scope, no controlling tty: fail closed rather than
    # silently picking a scope the invocation never asked for (spec-0005 AC5).
    fail "cannot resolve a scope: not inside a git repository (project scope needs one), and no controlling terminal is available to ask (--scope/\$TRELLIS_SKILLS_SCOPE was not given either). Nothing was written. Pass --scope personal (or TRELLIS_SKILLS_SCOPE=personal) to vendor the plugin globally, or re-run inside a git repo for project scope."
  fi
fi
say "scope: $scope ($scope_origin)"

if [ "$scope" = "project" ]; then
  target="$git_root/.claude/skills/trellis"
else
  target="$HOME/.claude/skills/trellis"
fi

# --- 2. Fetch the bundle into a staging dir and verify it — nothing in the target -
#         directory is touched until every staged byte checks out against the
#         manifest below. This is the pin-then-verify-before-write shape (adapted
#         from #128's install.sh:140-153), scoped to the whole plugin bundle.

# Reads a shasum-style manifest on stdin, checks it inside directory $1.
manifest_check() {
  if command -v shasum >/dev/null 2>&1; then (cd "$1" && shasum -a 256 -c -)
  else (cd "$1" && sha256sum -c -)
  fi
}

BUNDLE_SOURCE="${TRELLIS_BUNDLE_SOURCE:-https://raw.githubusercontent.com/kodhama/trellis/main/plugins/trellis}"

stage="$(mktemp -d "${TMPDIR:-/tmp}/trellis-vendor.XXXXXX")"
# $rendered_tmp joins the cleanup once it exists. Reproduced: SIGINT during the
# render left a partial .claude/rules/.trellis.md.<pid> behind AND the stage dir,
# so EXIT alone did not cover signals either.
#
# Two things this must get right, both measured the hard way:
#
#   1. A signal handler that only CLEANS UP does not stop the script. POSIX
#      resumes execution after the handler returns, so a `trap ... INT` that
#      omits `exit` left the script cat-ing files it had just deleted. Each
#      signal handler exits with the conventional 128+signo.
#   2. An EXIT trap becomes the shell's last command, and bash then reports ITS
#      status. That turned a fatal expansion error into `exit 0` — the installer
#      announcing success for an install it had refused. $rc is captured first
#      and re-raised, so the trap can no longer launder a failure into a pass.
rendered_tmp=""
target_new=""
target_old=""
cleanup() {
  rm -rf "$stage"
  [ -z "$rendered_tmp" ] || rm -f "$rendered_tmp"
  # A swap that died between the two moves leaves the target missing and the
  # previous install parked at $target_old. Put it back rather than leaving the
  # consumer with no plugin at all.
  [ -n "$target_new" ] && rm -rf "$target_new"
  if [ -n "$target_old" ] && [ -d "$target_old" ]; then
    [ -d "$target" ] || mv "$target_old" "$target" 2>/dev/null || true
    rm -rf "$target_old"
  fi
}
trap 'rc=$?; cleanup; exit $rc' EXIT
trap 'cleanup; exit 130' INT
trap 'cleanup; exit 129' HUP
trap 'cleanup; exit 143' TERM

# The bundle manifest — baked in, covers the whole plugins/trellis/ tree. Advance-
# guarded by cli/install_script_test.go:TestInstallScriptBundleManifestIsCurrent.
bundle_manifest() {
  cat <<'TRELLIS_BUNDLE_MANIFEST'
89e04f3cf9a24f29b1bcc01daf5c3c795189a171d10100890dad836681a57779  .claude-plugin/plugin.json
600d207e6f4ea8dc73b54880d4def72947b25d3a054136f1c32446aa186d4a9b  .codex-plugin/plugin.json
17860ddbffa58a413af0e6164af559da695ef812900efb5f177aea97806ec3c2  README.md
40b8eb4000a913a7791090535f291d3d369874162a89ef3c9e3d4e887a1b9e79  VERSION
3da7f2cf8765fe95d1936a36d3341736f16b438353f2130368af58897dad20c4  hooks/codex-context.mjs
33bd291e8cab52f2b6f3d08eff19ca8e685c5357266f1960c31543076612f986  hooks/codex-hooks.json
a289f0cd911c4392a89f3339d03feead7a2735dacfb893ff886ccb625bd2c809  hooks/hooks.json
06c1693e9d29dfecdb4356309f362cff35a3c2dae50443483907b067ac854bc9  hooks/staleness.sh
a224cdcb7a0e2cb1b47c267a3d662d49f840aa49bc9390e21a5f04d451a6cd5c  reference/block-claude.md
3a676709b23fd12f730695c71b46f7a6f485ec5d363739c40f52fb902f86f842  reference/block-codex.md
c277d931c9f8512e948b8d79e50d7c60859b1f875f4f5e682ba07a228890a0a7  reference/block-inline-a-head.md
f15315d1df95ca5df6bc59901e65691ffb77187e45f4df0a5438365e80149342  reference/block-inline-a.md
32d15b7d14c252c97a08e1a900e01ebef31a954738fb5f888e8b47f9512bcaa6  reference/block-inline-b-head.md
f8efb5a0cd6f636164d8283c738a264b775671b10cb1e984613bb96f9ddef933  reference/block-inline-b.md
a33f9904a063986caab0ccc156a229b959c232829c1f1b90c252377d3c028795  reference/block-inline-tail.md
4318ba1db3eb5f270415e1288bfb179acc6229c27a276611ddd4e0fc9d76b214  reference/checksums
68b803d9f4a45fc1af07c327a35c4a4b58aac878b233a4196079dac419bd28b3  reference/invariants.md
a675233ee08c0c41b5c0490a163f4d6ff4e95c6bbf9964eac59e4772f6597454  reference/rules-a.toml
534c9178b4c5173f6dd51f382a48b970cc263a83d410c0e9fff6c41a7c937386  reference/rules-b.toml
238fca2fb2c9e91af22764fc403da9b0dc8625291ef8fc4530d52c1a2618df0e  reference/rules.md
d447439d5f393f8bbe2af31fea3f426c0e752f621b64b4262da0866bded15251  reference/trellis-a.md
df6bfd11ce981c821eff612b6dfb0c95313edbf4222b9c01ace2fd2cd08baae4  reference/trellis-b.md
f63c4d15f8ce3cf4932ed3412e141e3e47b886daed15223c8402b1c3718049c3  reference/version
dc981bb8def6820bc8ff4c6b82785569d50016fd41e52adf1dc209a561234a44  skills/remove/SKILL.md
TRELLIS_BUNDLE_MANIFEST
}

bundle_manifest >"$stage/manifest"
bundle_files="$(awk '{print $2}' "$stage/manifest")"

fetch() {
  rel="$1"
  dst="$stage/bundle/$rel"
  mkdir -p "$(dirname "$dst")"
  case "$BUNDLE_SOURCE" in
    http://*|https://*)
      command -v curl >/dev/null 2>&1 || fail "curl is required to fetch the bundle from $BUNDLE_SOURCE"
      curl -fsSL "$BUNDLE_SOURCE/$rel" -o "$dst" || fail "fetching $BUNDLE_SOURCE/$rel failed"
      ;;
    *)
      cp "$BUNDLE_SOURCE/$rel" "$dst" 2>/dev/null || fail "copying $BUNDLE_SOURCE/$rel failed"
      ;;
  esac
}

for f in $bundle_files; do fetch "$f"; done
out="$(manifest_check "$stage/bundle" <"$stage/manifest" 2>&1)" || fail "bundle checksum verify failed — the fetched files do not match this script's baked-in manifest. Nothing was installed. This means either the fetch was corrupted or tampered in transit, or the bundle at $BUNDLE_SOURCE has moved past what this copy of install.sh expects — re-download install.sh from https://raw.githubusercontent.com/kodhama/trellis/main/install.sh and re-run. shasum said:
$out"

# --- 3. Write — overwrite the plugin's own files; .trellis/ is never WRITTEN ----
#         (the setup skill owns .trellis/ entirely; this script reads only whether
#         a few paths EXIST there, never any file's contents,
#         and this script never runs a git command that mutates anything)

# The bundle REPLACES the target rather than being copied over it. Copying in
# place only ever creates and overwrites, so a file that LEAVES the bundle
# survives every future upgrade: when decision-0072 retired skills/setup/, an
# existing curl install kept the directory on disk and Claude Code kept
# discovering the retired setup skill from it — the retirement reached new
# installs and missed exactly the consumers who could not be told. That is a
# class, not one skill, so the fix is the swap and not a delete-list.
target_new="$target.new.$$"
target_old="$target.old.$$"
rm -rf "$target_new"
mkdir -p "$target_new" || fail "could not create $target_new (permissions?). Nothing was changed."
for f in $bundle_files; do
  mkdir -p "$target_new/$(dirname "$f")"
  cp "$stage/bundle/$f" "$target_new/$f"
done
chmod +x "$target_new/hooks/staleness.sh"
if [ -e "$target" ]; then
  mv "$target" "$target_old" || fail "could not move the existing install aside ($target). Nothing was changed."
fi
if mv "$target_new" "$target"; then
  target_new=""
  rm -rf "$target_old"
  target_old=""
else
  [ -d "$target_old" ] && mv "$target_old" "$target" 2>/dev/null
  target_old=""
  fail "could not put the new bundle in place at $target. The previous install was restored."
fi

stamp="$(head -n1 "$stage/bundle/reference/version" 2>/dev/null | tr -d '[:space:]')"
nfiles="$(printf '%s\n' "$bundle_files" | wc -l | tr -d ' ')"

# --- 3b. Render the rules file (decision-0068 D1) -------------------------------
#
# Without this the vendored bundle delivers NO rules at all: measured, with user
# scope excluded, a bundle at .claude/skills/trellis/ produces nothing, because
# the skills-directory-plugin path needs a workspace-trust dialog a headless run
# cannot grant (issue #201). `.claude/rules/*.md` loads at launch with no hook,
# no settings write and no dialog.
#
# PROJECT SCOPE ONLY (D1, maintainer 2026-07-30). ~/.claude/rules/trellis.md
# would govern every repo on the machine and import ~/.trellis/rules.toml, which
# nothing writes.
#
# Still zero decision logic in the sense AC2's heading means — no posture chosen,
# no marker patched. AC2's "never reads" clause WAS amended for this branch (see
# the spec's frontmatter and AC2d): six reads over five paths, of which exactly
# ONE reads a file's contents — the managed-block opening marker. No .trellis/
# file's CONTENTS are read and no posture is chosen. It emits trellis-b's
# prose as a CONSTANT — staleness.sh already resolves absent strictness to `b`,
# so the install path inherits a ratified default rather than inventing one.
#
# Two edits, both of which staleness.sh already performs under decision-0065's
# "one edit" allowance:
#   1. resolve the @rules.md placeholder — left in place it resolves to
#      .claude/rules/rules.md, which does not exist, and the whole rules body
#      silently vanishes;
#   2. repoint the invariants pointer, which ships naming .trellis/internal/ —
#      a path this script never creates (D1). Repo-relative, never absolute: the
#      file is meant to be committed, and an absolute path would carry this
#      machine's layout to every collaborator.
rendered_note="no rules file (project scope only)"
static_conflict=""
if [ "$scope" = "project" ]; then
  # Every shape that already delivers the rules STATICALLY. Checking only
  # .trellis/internal/ missed two, both reported by review:
  #   - the pre-decision-0051 FLAT layout, whose managed block imports
  #     @.trellis/trellis.md with no internal/ directory at all; and
  #   - the managed block itself, which is the thing that actually does the
  #     importing.
  #
  #     (An earlier version of this comment justified the flat shape by claiming
  #     .trellis/version is gitignored, so a fresh clone would carry the import
  #     chain with no stamp. FALSE, verified: .trellis/internal/version is
  #     TRACKED in this tree, and decision-0043 scopes that exception to the
  #     repo's OWN self-hosted overlay. The shape is real; the reason given for
  #     it was not.)
  # $conflict_paths is what the remedy below tells the user to DELETE, and it has
  # to name the shape actually found. The remedy used to be hard-coded to
  # .trellis/internal/ while this variable took three values — so a flat-overlay
  # or managed-block project was told to delete something it does not have,
  # deleted nothing, and hit this same refusal on every subsequent run. The
  # sibling defect in staleness.sh was fixed earlier in this branch; this is the
  # same class in the other script, and it was missed there first.
  conflict_paths=""
  [ -d "$git_root/.trellis/internal" ] && { static_conflict=".trellis/internal/ overlay"; conflict_paths=".trellis/internal/ and the managed block in your instructions file"; }
  [ -z "$static_conflict" ] && [ -f "$git_root/.trellis/trellis.md" ] && { static_conflict="legacy flat .trellis/ overlay"; conflict_paths=".trellis/trellis.md, .trellis/version if present, and the managed block in your instructions file"; }
  # The INLINE managed block needs its own check, and this is the ONE content
  # read in the script. The existence tests above cannot reach it: the inline
  # form embeds the whole rules body in CLAUDE.md and needs no .trellis/internal/
  # at all — its own tail says "Rule activation follows the rows in
  # .trellis/rules.toml directly", so such a project has rules.toml and nothing
  # else under .trellis/. Both tests miss it and the render proceeds into live
  # double delivery. Measured, not reasoned: a consumer built from the shipped
  # block-inline-b.md got `inv-directional-flow` in CLAUDE.md AND in the rendered
  # file, with the hook emitting its quiet stand-down — both components
  # affirmatively misreporting.
  #
  # This check was deleted earlier in this branch on the claim that the existence
  # tests already covered every shape. That claim was false; the inline form is a
  # fourth shape, and deleting the check was a REGRESSION against this script's
  # own parent commit, which refused. Restored.
  #
  # Two prior findings killed the old version, and the fix for each is structural
  # rather than a patch:
  #   - CRLF (core.autocrlf=true is the Git-for-Windows default) broke the old
  #     CLOSING grep, which was $-anchored: `trellis:end -->\r` never matched
  #     `-->$`, so a REAL block went undetected. Dropping the closing grep
  #     removes the only $-anchor, so CR at end of line cannot matter.
  #   - Prose that merely NAMED the delimiters matched twice. Column-0 anchoring
  #     fixes that: documentation writes the marker mid-sentence.
  # Opening marker only, anchored at column 0, with an optional leading UTF-8
  # BOM. The BOM matters for the same reason it does in staleness.sh: setup wrote
  # the block at line 1 of a fresh CLAUDE.md, and an editor on a Windows-default
  # checkout rewrites the encoding. Without it a real block escapes the check and
  # renders into live double delivery -- the same fail-open direction, on the
  # writer side, as the reader-side bug fixed in the same change. Leaving one
  # half fixed would have been the worse outcome of the two.
  #
  # Only CLAUDE.md and AGENTS.md are checked, while /trellis:remove recognises
  # blocks in five instruction files. That is deliberate, not an oversight: the
  # other three (GEMINI.md, .github/copilot-instructions.md, .clinerules) are not
  # loaded by Claude Code, so a block in one of them cannot double-deliver
  # alongside .claude/rules/. Stated because D7 asks for stated, not implied.
  #
  # Verified: LF inline -> refuses; CRLF inline -> refuses; BOM'd inline ->
  # refuses; prose naming the marker mid-sentence -> renders.
  #
  # Known false positive, accepted deliberately: a CLAUDE.md that DOCUMENTS the
  # marker inside a fenced code block puts it at column 0, so this refuses and
  # tells the author their project already delivers the rules statically, which
  # is false. It fails CLOSED — nothing is written and the message names the file
  # — and the alternative is parsing markdown in POSIX sh to find fences. The
  # author can move the example or indent it. Chosen over the fail-OPEN
  # alternative of dropping the anchor, which reopens the prose case.
  if [ -z "$static_conflict" ]; then
    bom="$(printf '\357\273\277')"
    for f in CLAUDE.md AGENTS.md; do
      grep -q "^\($bom\)\{0,1\}<!-- trellis:begin" "$git_root/$f" 2>/dev/null \
        && { static_conflict="managed block in $f"; conflict_paths="the managed block in $f, from its trellis:begin marker to its trellis:end marker (there is no overlay directory in this shape)"; break; }
    done
  fi
fi
if [ "$scope" = "project" ] && [ -n "$static_conflict" ]; then
  # A pre-plugin-delivery consumer whose CLAUDE.md managed block imports
  # @.trellis/internal/trellis.md. Rendering here would put BOTH static chains
  # into context: Claude loads the managed block's imports AND .claude/rules/*.md
  # itself, before any hook runs. The hook's path-A-first ordering suppresses only
  # what the HOOK injects — it cannot un-load a file Claude already read. So this
  # is refused at install time, because there is no runtime fix for it.
  #
  # Every read this script makes of pre-existing project state — six sites, five
  # paths. An earlier version of this comment claimed "the ONE place this script
  # reads .trellis/", which was false; a later one claimed no contents are read
  # anywhere, which is false while the marker check exists. Both are corrected:
  #   - .trellis/internal/ and .trellis/trellis.md   (existence, above)
  #   - CLAUDE.md / AGENTS.md                        (CONTENT: opening marker)
  #   - .claude/rules/trellis.md                     (existence x2: the live
  #                                                   double-delivery warning
  #                                                   below, and the non-regular
  #                                                   refusal before the mv)
  #   - .trellis/rules.toml                          (existence, for the
  #                                                   floors-only guidance line)
  # Exactly ONE is a content read, and it reads one line-anchored string. None
  # selects a posture, patches a marker, or writes anything under .trellis/.
  # spec-0005 AC2's second amendment permits exactly this and no more.
  rendered_note="no rules file — $static_conflict present"
  say "NOT rendering .claude/rules/trellis.md: this project already delivers the"
  say "rules statically ($static_conflict). Adding the rendered file would deliver"
  say "them twice, and no hook can undo that — both are loaded before any hook runs."
  if [ -f "$git_root/.claude/rules/trellis.md" ]; then
    # Refusing does not help if the file is ALREADY there: an earlier run
    # rendered it and the overlay arrived afterwards (a collaborator's commit, a
    # reverted migration). Double delivery is live right now, and saying
    # "no rules file" would be false at the moment it is printed.
    rendered_note="LEFT IN PLACE — .claude/rules/trellis.md exists AND $static_conflict"
    say ""
    say "WARNING: .claude/rules/trellis.md ALREADY EXISTS in this project, so the"
    say "double delivery described above is live right now — this installer did not"
    say "create it and has not removed it. Delete that file, or migrate off the"
    say "static overlay, before relying on either."
  fi
  say "Migrate first: delete $conflict_paths,"
  say "keeping your .trellis/rules.toml rows — or /trellis:remove to take Trellis"
  say "out entirely. Then re-run this installer."
elif [ "$scope" = "project" ]; then
  rules_dir="$git_root/.claude/rules"
  mkdir -p "$rules_dir" || fail "could not create $rules_dir (is .claude/rules present as a file?). The bundle is already vendored; re-run once the path is clear."

  # The placeholder must exist or the first sed silently emits the whole file
  # minus its last line and the second emits nothing — a truncated render that
  # every other check would pass. Unreachable through a verified fetch; this is
  # defence in depth against a coordinated payload edit.
  grep -q '^@rules\.md[[:space:]]*$' "$stage/bundle/reference/trellis-b.md" \
    || fail "reference/trellis-b.md carries no @rules.md placeholder line; refusing to render a truncated rules file"

  # Render to a sibling temp file and move it into place only on success.
  # `{ ...; } > target` truncates the target BEFORE the body runs and swallows a
  # redirect failure: measured, a read-only target produced exit 0, a success
  # banner, and the user's prior bytes untouched — the installer lying about what
  # it did. It is also shell-dependent (bash continues; dash aborts bare). A
  # partial render is worse still: the hook stands down on the leftover file and
  # the session runs ungoverned while both the installer and the hook claim rules
  # are loaded.
  # A non-regular target must be refused BEFORE the move. `mv file dir` moves the
  # file INTO the directory: measured, that produced exit 0, a success banner, no
  # rules file, and a stray temp file buried inside .claude/rules/trellis.md/.
  if [ -e "$rules_dir/trellis.md" ] && [ ! -f "$rules_dir/trellis.md" ]; then
    fail "$rules_dir/trellis.md exists and is not a regular file; refusing to render over it. The bundle is already vendored; clear that path and re-run."
  fi
  rendered_tmp="$rules_dir/.trellis.md.$$"
  # The three payload-derived parts are rendered SEPARATELY and each checked,
  # because neither `set -eu` nor the group's `|| {...}` can see them fail.
  # A `{ ...; } > f || {...}` group exits with its LAST command's status, and it
  # is the left operand of `||` so `set -e` is suppressed inside it; measured in
  # both dash and bash, a failing command mid-group yields group exit 0 and a
  # truncated file. Worse, these are PIPELINES, whose status is the last stage's:
  # `sed ... | sed '$d'` reports success when the first sed dies, and POSIX sh
  # has no pipefail. So the only reliable signal is the artifact itself.
  #
  # Deliberately NOT a prose check. Probing for the posture sentence or the
  # invariants pointer would re-create the exact defect this branch fixed — a
  # legitimate payload reword breaking every install. `-s` asks only "did this
  # step produce anything", which no reword can falsify.
  sed -n '1,/^@rules\.md[[:space:]]*$/p' "$stage/bundle/reference/trellis-b.md" | sed '$d' > "$stage/render.head"
  [ -s "$stage/render.head" ] || fail "rendering the posture prose produced nothing; the bundle's reference/trellis-b.md is damaged or has no @rules.md placeholder. Nothing was written."
  cat "$stage/bundle/reference/rules.md" > "$stage/render.body"
  [ -s "$stage/render.body" ] || fail "rendering the rules body produced nothing; the bundle's reference/rules.md is missing or empty. Nothing was written."
  sed -n '/^@rules\.md[[:space:]]*$/,$p' "$stage/bundle/reference/trellis-b.md" | sed '1d' \
    | sed 's|`\.trellis/internal/invariants\.md`|`.claude/skills/trellis/reference/invariants.md`|' > "$stage/render.tail"
  [ -s "$stage/render.tail" ] || fail "rendering the header tail produced nothing; the bundle's reference/trellis-b.md is damaged. Nothing was written."
  {
    # A MACHINE-OWNED opening marker. The hook used to validate this file by
    # matching prose landmarks — the invariants sentence, the posture note, the
    # activation heading. All of that is payload text that may legitimately be
    # reworded, and when it was, every freshly installed project got a permanent
    # false "not governed" warning while the whole suite stayed green. Markers
    # this script owns cannot drift out from under the reader.
    printf '<!-- trellis:rendered-begin -->\n'
    cat "$stage/render.head"
    cat "$stage/render.body"
    cat "$stage/render.tail"
    # D5's single sentence of new prose. The posture sentence above is frozen at
    # install time; the rows below are live and carry their own `strictness`.
    # They can disagree, and the reader is told which wins rather than left to
    # see a contradiction.
    printf '<!-- trellis:rendered-footer -->\n'
    printf '**If the posture sentence above and the rows below disagree, the rows win:**\n'
    printf 'the `strictness` key in `.trellis/rules.toml` is authoritative. The sentence\n'
    printf 'above was fixed when this file was written; the rows are read fresh every\n'
    printf 'session. Edit `strictness` in `.trellis/rules.toml` to change the posture.\n'
    printf '\n'
    printf '## Project rule activation\n'
    printf '\n'
    printf '@../../.trellis/rules.toml\n'
    # The drift surface. Without a stamp the hook can only stand down blindly,
    # and a file rendered by an older installer would govern forever with no
    # signal — decision-0035's floor applied to this artifact.
    printf '\n<!-- trellis:rendered-from %s -->\n' "$stamp"
  } > "$rendered_tmp" || {
    rm -f "$rendered_tmp"
    fail "could not write $rendered_tmp — the rules file was not rendered and nothing was replaced. The bundle is already vendored; fix the permission and re-run."
  }
  # `{ ...; } > file || {...}` catches REDIRECT failure only: the group's own
  # exit status is the LAST command's, and `set -eu` does not fire inside the
  # left operand of `||`. Measured in both dash and bash: a failing `cat` mid-
  # group yields group exit 0 and a truncated file. Two of the writes above are
  # sed pipelines whose failure leaves every marker and every rule line intact —
  # install.sh would report success, the hook would stand down quietly, and the
  # file could silently lose the activation sentence, which INVERTS the rule
  # semantics (every rule applies regardless of its row).
  #
  # So validate the artifact, not the exit status. This is WEAKER than the hook's
  # reader-side check and is not a mirror of it: the probes are matched
  # independently (no ordering), and `grep -c` counts matching LINES where the
  # hook counts DISTINCT ones. Three of the five probes also check strings this
  # script itself printf's, so they cannot catch a payload-derived step failing —
  # that is the -s checks' job, above. What it does guarantee is that a render
  # which would not satisfy the hook never reaches the mv.
  render_defect=""
  for probe in \
    '<!-- trellis:rendered-begin -->' \
    '<!-- trellis:rules-loaded -->' \
    '<!-- trellis:rendered-footer -->' \
    '@../../.trellis/rules.toml' \
    'the `strictness` key in `.trellis/rules.toml` is authoritative'
  do
    grep -qF "$probe" "$rendered_tmp" 2>/dev/null || { render_defect="$probe"; break; }
  done
  [ -n "$render_defect" ] || [ "$(grep -cE '`(inv|floor)-[a-z-]+`' "$rendered_tmp" 2>/dev/null)" -ge 5 ] || render_defect="the rule body (fewer than 5 rule lines survived)"
  [ -z "$render_defect" ] || {
    rm -f "$rendered_tmp"
    fail "the rendered rules file is incomplete — missing: $render_defect. Nothing was replaced. This means a step of the render failed without reporting it; re-run, and if it repeats the bundle at $BUNDLE_SOURCE is damaged."
  }
  mv -f "$rendered_tmp" "$rules_dir/trellis.md" || {
    rm -f "$rendered_tmp"
    fail "could not move the rendered rules file into place at $rules_dir/trellis.md"
  }
  rendered_note=".claude/rules/trellis.md"

  # decision-0070 D2. Seed the rows, so the path that renders the rules also
  # makes them apply. Without this the curl path shipped its full context cost —
  # all fourteen rules, always loaded — while only the two floor- rules turned,
  # and the rendered file's own @../../.trellis/rules.toml import resolved to
  # nothing, silently.
  #
  # Running an installer INSIDE a repository is the adoption act (D1), which is
  # what makes writing here legitimate where the plugin path must ask. Never
  # overwritten: an existing rules.toml is the project's own and outranks a seed.
  #
  # This amends decision-0065:26-29, the plugin/install split ("install.sh is
  # vendoring and never configures ... and continues to never touch .trellis/") —
  # NOT :18-19's "setup writes exactly one file", which binds the skill and is
  # untouched. That clause was named here, and in three other places, before the
  # boundary was got right. Openly,
  # in decision-0070 D2, not by routing around it. The clause's purpose, that no
  # path silently vendors an overlay, is untouched: one config file, in the repo
  # the user pointed this script at.
  if [ -e "$git_root/.trellis/rules.toml" ] && [ ! -f "$git_root/.trellis/rules.toml" ]; then
    # A non-regular target must be refused BEFORE the copy. `cp file dir` copies
    # INTO the directory and returns 0 — measured, that produced a success banner,
    # seeded_rows=yes, a stray rules-b.toml buried inside .trellis/rules.toml/,
    # and an import resolving to nothing. Exactly the defect already guarded on
    # the rendered file's move; this block reintroduced it.
    seeded_rows=failed
  elif [ ! -f "$git_root/.trellis/rules.toml" ]; then
    if mkdir -p "$git_root/.trellis" 2>/dev/null &&
       cp "$stage/bundle/reference/rules-b.toml" "$git_root/.trellis/rules.toml" 2>/dev/null; then
      seeded_rows=yes
    else
      # Not fatal: the rules file is already in place and governing at the floors.
      # Say so rather than exit, because the install itself succeeded.
      seeded_rows=failed
    fi
  fi
fi

# --- 4. Confirm — never a git mutation; the commit is yours -----------------------

say "vendored the Trellis plugin ($stamp) to $target"
say "  $nfiles files written; manifest verify OK on every byte before anything was written"
say "  rules: $rendered_note"
if [ "$scope" = "project" ]; then
  say ""
  say "Claude Code will show its workspace-trust dialog the next time you launch it"
  say "here (project-scope plugins load only after you accept it: see"
  say "code.claude.com/docs/en/settings)."
  say "Project-scope skills-directory plugins do NOT walk up to the repo root: launch"
  say "Claude Code from $git_root itself, or run /reload-plugins after cd'ing there —"
  say "starting from a subdirectory will silently miss the plugin."
  say ""
  say "Review the new files, then commit them yourself if you want collaborators to"
  say "get them on clone — this script never runs git:"
  if [ "$rendered_note" = ".claude/rules/trellis.md" ]; then
    add_paths=".claude/skills/trellis .claude/rules/trellis.md"
    # The seeded rows too, and only when this run actually wrote them. Without
    # this a collaborator cloning the repo gets the rules file and the bundle but
    # NO activation rows — which is the pre-decision-0070 state the seed exists to
    # end, reintroduced one `git clone` later. Same reason the else-branch omits
    # the rendered file: naming a path that was not written makes `git add` exit
    # 128 and the `&&` then swallows the commit.
    [ "${seeded_rows:-}" = yes ] && add_paths="$add_paths .trellis/rules.toml"
  else
    # On any refusal path the file was not written. Naming it would make the
    # printed command fail with `pathspec ... did not match any files` (exit
    # 128), and because of the `&&` the commit would never run either.
    add_paths=".claude/skills/trellis"
  fi
  say "  git -C \"$git_root\" add $add_paths && git -C \"$git_root\" commit -m 'chore: vendor the Trellis plugin'"
fi
say ""
if [ "$scope" = "project" ]; then
  # Gated on project scope: personal scope renders no rules file at all and
  # already says so, and spec-0005 AC10 caps its output at items 1 and 5.
  say "The rendered rules file is a Claude Code mechanism: Codex CLI and other hosts"
  say "get nothing from it. On those hosts the rules arrive through the plugin, or not"
  say "at all."
  say ""
fi
case "${seeded_rows:-}" in
  yes)
    say "This project is governed now: all fourteen rules are active, followed by"
    say "default with deviations said out loud. .trellis/rules.toml holds the rows —"
    say "it is yours to edit: strictness = \"firm\" for the by-the-book posture,"
    say "active = false on a row to turn that rule off."
    ;;
  failed)
    say "Could not write .trellis/rules.toml (permissions?). The rules file is in place,"
    say "but until those rows exist only floor-transparency and floor-intent-gate apply."
    say "Create .trellis/rules.toml yourself to activate the rest — copy the preset"
    say "from the installed plugin's reference/rules-b.toml."
    ;;
  *)
    # Nothing was seeded on this run — either the rows already existed, or the
    # render was refused (a static-delivery conflict), or this is personal scope.
    # Those are NOT the same state, and the old code only ever printed one line
    # for all of them. `seeded_rows` is unset here, so ask the disk instead.
    if [ "$scope" = "project" ] && [ ! -f "$git_root/.trellis/rules.toml" ]; then
      # The floors-only warning, restored. Dropping it was a regression: a
      # static-conflict repo with no rows is running on two rules out of fourteen
      # and was no longer told so.
      say "This project has no .trellis/rules.toml, so only floor-transparency and"
      say "floor-intent-gate apply — every other rule is gated on a row in that file."
      say "Write it yourself to activate the rest — copy the preset from the"
      say "installed plugin's reference/rules-b.toml."
    else
      say "Edit .trellis/rules.toml to change the posture or turn individual rules"
      say "off: strictness = \"firm\" for by-the-book, active = false on a row. That"
      say "file is yours and is the authority (decision-0053)."
    fi
    ;;
esac
