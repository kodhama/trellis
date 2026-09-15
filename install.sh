#!/bin/sh
# install.sh — vendor the Trellis Claude Code plugin onto disk as a skills-directory
# plugin (corrected design per spec-0005 — that spec retired with `specs/` in
# decision-0079, so the `spec-0005 AC#` markers below name requirements whose only
# surviving statement is cli/install_script_test.go; the text is in git history.
# Supersedes the closed
# #128 attempt — see #128's own closing comment). This is NOT the retired end-user
# binary installer (kodhama-0007 rule 5, decision-0043 §4 — see the note appended
# there): it downloads no binary and, more importantly, makes exactly ONE decision
# (where to put the plugin) and composes NOTHING else. Every other decision —
# posture, which instructions file to patch, block style, hand-authored-content
# guarding — stayed entirely inside the retired setup skill's SKILL.md, unmodified
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
# point — seeding a missing .trellis/rules.toml with a fixed two-line file, a
# comment naming the opt-out row and an empty [rules] table (decision-0070 D2,
# as its dated note states it now), never overwriting; otherwise .trellis/ is
# the consumer's to edit (decision-0072 retired the setup skill). The seed
# switches no rule off, and nothing in it is a decision this script makes. One
# header ships, reference/trellis.md, so the rendered rules file is the same for
# every project whatever its rules.toml says (TRL-97). When that file declares
# `governed = false`, the project has opted
# out of Trellis (decision-0070 D5): the plugin hook reads that key before every
# delivery path and injects nothing there, so this script renders nothing there
# either — the bundle is vendored, no rules file is written, and the opt-out is
# left as it is (TRL-38). That key is the only project CONTENT that decides
# anything about delivery here; the managed-block marker and CLAUDE.md's
# @AGENTS.md import line are read too, but only to decide whether the render is
# refused. It NEVER runs a git
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
#
# PLATFORM: macOS, Linux, WSL (decision-0068 §8; #210). That dependency list is the
# whole of the boundary, and the first entry is the binding one: this is a POSIX `sh`
# script, so cmd and PowerShell cannot run it at all. A Windows user needs WSL — or
# some other POSIX environment; Git Bash/MSYS does run it, but sits outside the stated
# boundary and no check here covers it. Nothing narrows that slice further: the script
# creates no symlink, so no filesystem or developer-mode question rides on top of the
# shell one. Running natively on Windows would take a separate installer — §8 names "a
# future PowerShell or Go installer" — not a change to this file.

set -eu

say()  { printf 'trellis: %s\n' "$*"; }
fail() { printf 'trellis: FAIL: %s\n' "$*" >&2; exit 1; }

usage() {
  cat <<'EOF'
install.sh — vendor the Trellis Claude Code plugin onto disk (skills-directory install).

  curl -fsSL https://raw.githubusercontent.com/kodhama/trellis/main/install.sh | sh
  sh install.sh [--scope personal|project] [--non-interactive]

Platform: macOS, Linux, WSL. A POSIX sh script — cmd and PowerShell cannot run
it at all, so use WSL there. The dependency list at the top of this file is the
whole of the boundary, and POSIX sh is the binding entry in it.

This is the ONLY decision this script makes. A project with no .trellis/rules.toml
is seeded with a file that switches no rule off, so every rule applies; one that
already has the file keeps it exactly as it is. One whose file declares
`governed = false` has opted out: it gets the bundle and no rules file, exactly as
the plugin hook delivers nothing there. Either way the file is yours to edit
afterwards — a row `<slug> = { active = false }` under [rules] switches that rule
off; see the "next steps" this script prints when it finishes.

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
012c730d32c212ab8edd14c8218febb8e757946f1f2b5dd8bbc0d88050638fd3  .claude-plugin/plugin.json
20c24c02b78bf4d86b70f7a6353814252daeef0f19d1488c067bfec2d856637a  .codex-plugin/plugin.json
310cdc794d7d325f428b84da5f4c045bfbd1c2d65ab751cf1ebcdcca13fdbb11  README.md
e14c67c5f4dd36234ef4bc5ba5ef48d5d79e27bad4038c9bc8a64714566aa06a  VERSION
715c06ba485c2428d160f38bcd717bd1bca5b537a0296be417343e74771308d7  hooks/codex-context.mjs
33bd291e8cab52f2b6f3d08eff19ca8e685c5357266f1960c31543076612f986  hooks/codex-hooks.json
a741930673c1fb723ae6cce9421d49579c7eb5bc2dd0385a2b53be8b1d1b27f5  hooks/hooks.json
e904b2c88815ee4924156216c1cde6a181d3264609dcb2bf3b574bbdfa18b2da  hooks/staleness.sh
a224cdcb7a0e2cb1b47c267a3d662d49f840aa49bc9390e21a5f04d451a6cd5c  reference/block-claude.md
b1bb861740d59d7ce08023b889c38b246d70e7ea1f11ff398c0783fb826517a4  reference/block-codex.md
32d15b7d14c252c97a08e1a900e01ebef31a954738fb5f888e8b47f9512bcaa6  reference/block-inline-head.md
10892805ec9c8297e2385bf0c6a552ee64eca7491ca89862ee3941fe60833e32  reference/block-inline-tail.md
81842872e90aa26baa4614c74adb9d0b2520ce3094e1a78c6924ca7fad6898b1  reference/block-inline.md
c65d71a12daf8a3ce8b1c63ef5d32b7ad286ebd9bdf11e5c303b7b4811cd0813  reference/checksums
5c068cf2e5592dab37c0820430166bc101e1725689a1e189abce821bed8b2f6e  reference/invariants.md
eb75b9cf399792e8bab56c42635c7e39fb1555daf11866b7debca4af1ca37733  reference/rules.md
8787be1ab60f4718b6cc0e9ad34293f9bcd21e264aa9ad38a857c48217cbdd3d  reference/trellis.md
095c345b81dfd0dabdbce54c09785091f3ffba0fcd891b8f92950e16644c5d41  reference/version
247037659b91a1c4f0285c1e6b1c42637b18954f4e0b18a79647ef79d5f5d899  skills/remove/SKILL.md
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
# survives every future upgrade: when decision-0072 retired the setup skill, an
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
# Still zero decision logic in the sense AC2's heading means — no marker
# patched, nothing branched on an instructions file. AC2's "never reads" clause
# WAS amended for this branch (see the spec's frontmatter and AC2d): six reads
# over five paths, of which ONE read a file's contents — the managed-block
# opening marker. The enumeration in the static-delivery refusal branch below
# counts them now, and cli/install_script_test.go pins that count. TRL-38 added
# the top-level `governed` key of .trellis/rules.toml, read FIRST and with the
# hook's own matcher, which refuses the render when it says false. The hook
# reads that key before every delivery path and injects nothing on it
# (decision-0070 D5); an installer that rendered a governing file over it put
# the full rules body into always-loaded context for a project that had
# declined, and the hook could then only tell each session to disregard what
# the host had already read. TRL-37 had added a `strictness` read that selected
# one of two posture headers with the hook's own parser; TRL-97 retired it with
# the second header. One header ships, and every project receives its By
# default sentence, which is what the hook delivers too.
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
# Whether THIS HOST reads the shape $static_conflict names. Every shape it can
# take is loaded by Claude Code except one — a managed block in AGENTS.md that
# CLAUDE.md does not import — so `yes` is the default and the gate below is the
# only thing that clears it (TRL-44).
static_host_reads=yes
opted_out=no
bom="$(printf '\357\273\277')"
# The seed's comment line (TRL-97). The seed below writes it above an empty
# [rules] table, and the next steps show the same two lines wherever they tell
# a project to write the file. The hook's accept instruction quotes the same
# bytes; TestInstallScriptSeedMatchesTheHooksAcceptInstruction pins the pair.
trellis_seed_comment='# Every Trellis rule applies. To switch one off, add a row: <slug> = { active = false }'
say_seed() {
  say "  $trellis_seed_comment"
  say "  [rules]"
}
if [ "$scope" = "project" ]; then
  # decision-0070 D5, read BEFORE every other branch, exactly where the hook
  # reads it: an explicit refusal outranks every default and every other
  # branch. The three lines below are the hook's own matcher, copied because
  # the hook lives inside the bundle this script vendors and cannot be shared;
  # a test pins the two copies to each other (decision-0028: a guard per pair).
  # What they establish: a file whose HEAD (the lines before any [table])
  # carries exactly ONE top-level `governed` assignment, and it says false. A
  # `governed = false` under [rules] is not a top-level key and does not opt
  # out — the hook governs such a project normally, and so this script renders
  # for it; the same narrowed divergence the hook records against Codex.
  # An unreadable file leaves $governed_head empty, so it is never an opt-out
  # here — nor in the hook, whose sed fails the same way; that case is handled
  # in the render branch below. Regular-and-readable BEFORE the open: a FIFO at
  # that path (review found it)
  # would block the sed forever waiting for a writer, ahead of the non-regular
  # handling the seed step already has. The hook takes the same guard on its
  # own copy of this read (TRL-43, this change); the parity test pins the
  # guard lines too, so the two cannot drift apart again.
  governed_head=""
  if [ -f "$git_root/.trellis/rules.toml" ] && [ -r "$git_root/.trellis/rules.toml" ]; then
    governed_head="$(sed "1s/^$bom//" "$git_root/.trellis/rules.toml" 2>/dev/null | sed -n '/^[[:space:]]*\[/q;p')"
  fi
  governed_n="$(printf '%s\n' "$governed_head" | LC_ALL=C grep -cE '^[[:space:]]*governed[[:space:]]*=' 2>/dev/null || true)"
  if [ -f "$git_root/.trellis/rules.toml" ] && [ "${governed_n:-0}" -eq 1 ] &&
     printf '%s\n' "$governed_head" | LC_ALL=C grep -qE '^[[:space:]]*governed[[:space:]]*=[[:space:]]*false[[:space:]]*(#.*)?$' 2>/dev/null; then
    opted_out=yes
  fi

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
  # inline block got `inv-directional-flow` in CLAUDE.md AND in the rendered
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
  # BOM. The BOM matters for the same reason it does in staleness.sh: a past setup wrote
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
  # alongside .claude/rules/. Stated because D7 asks for stated, not implied —
  # and because decision-0073 D1's per-component relevance clause requires a
  # deliberate subset of the delivery states said where it is done, with the
  # pointer: decision-0073 (the closed set's normative home; this narrowing is
  # its S4 row scoped to the files this host loads).
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
  #
  # TRL-44. A block in AGENTS.md is a block THIS HOST MAY NEVER READ. Claude Code
  # reaches AGENTS.md only through a standalone `@AGENTS.md` import line in
  # CLAUDE.md (decision-0057: "Claude Code can import AGENTS.md from CLAUDE.md;
  # Codex discovers AGENTS.md directly"), and the hook gates its own AGENTS.md
  # probe on exactly that line — `claude_imports_agents` in staleness.sh — for
  # decision-0073 D2's reason: never be wrong about the reader's state. The
  # fourth content read below closes the gap on the side that was wrong.
  #
  # TRL-48, decision-0091: the gate now reaches the RENDER REFUSAL too, which is
  # the half TRL-44 deliberately left ungated because removing a refusal is
  # decision-0090 D2's fourth clause and had to come to the corpus first. It
  # came, and D1 of 0091 rules that a managed block is a static-delivery conflict
  # for a given delivery only when the host that consumes that delivery loads the
  # block. This script renders .claude/rules/trellis.md, which only Claude Code
  # reads; a block in an AGENTS.md that CLAUDE.md does not import is, so far as
  # either component can see, read by neither — so there is no second delivery
  # for it to be doubled against, and refusing over it said the opposite in
  # terms. "So far as either component can see" is the whole of the claim, and
  # 0091 D1 is careful about it: the ground is the ADAPTER CONTRACT (one
  # standalone @AGENTS.md line, decision-0057 rule 2), which is the line both
  # scripts read, NOT an assertion that the host reaches AGENTS.md only that way.
  # It does not: an @-import chain via a third file is followed by the host and
  # seen by neither probe, and on that repo this gate is wrong. TRL-49 carries
  # the resolver; the NOTE below names the case rather than papering it.
  # Measured on main before the change: a
  # non-opted-out repo with the shipped inline block at AGENTS.md and no
  # import line got "NOT rendering ... this project already delivers the rules
  # statically (managed block in AGENTS.md)" and a remedy telling the author to
  # delete the block — the one thing that DOES deliver rules there, to Codex —
  # while staleness.sh on that same repo injected the full rule set.
  #
  # decision-0073 D1's per-component relevance clause, said where it is done, as
  # that clause requires: the two-file subset above is now a two-file subset with
  # ONE of them conditional, and the condition is the same S4-row scoping — the
  # files THIS HOST loads. The pointer is decision-0073 for the closed set, and
  # decision-0091 for the ruling that the refusal, not only the message, is
  # scoped that way.
  #
  # The asymmetry that kept it ungated, answered rather than dropped, and at the
  # strength the answer actually has: this script WRITES, so a false negative is
  # durable, while the hook only injects and is re-evaluated every session. The
  # durable state has TWO watchers and is GUARANTEED BY NEITHER. Render here, a
  # standalone `@AGENTS.md` line arrives later, and the host loads both — then
  # (a) the hook re-reads that line every session and emits
  # TRELLIS_STATIC_SHAPES_CONFLICT naming AGENTS.md and the remedy, and (b) a
  # re-run of this script takes the gate the other way and reports the double
  # delivery as live. Both verified in sequence on scratch repos. Both are
  # conditional: (a) needs the project-scope plugin to load, which decision-0068
  # records a headless run cannot grant — and headless is the case this install
  # path exists to serve — and (b) needs somebody to re-run. Neither watcher sees
  # the @-import CHAIN at all, which is why that case is a regression here and is
  # stated as one in 0091 rather than filed under pre-existing. What is claimed
  # is only this: the false positive being removed was UNCONDITIONAL, and its
  # remedy told the author to delete the block that governs Codex.
  #
  # ANCHORED, not a substring, and the anchor is the hook's: the contract is one
  # standalone `@AGENTS.md` line, and an unanchored match reads documentation
  # ABOUT the import — prose, or a fenced example — as the import itself.
  claude_imports_agents=no
  grep -qE "^($bom)?[[:space:]]*@AGENTS\.md[[:space:]]*$" "$git_root/CLAUDE.md" 2>/dev/null && claude_imports_agents=yes
  if [ -z "$static_conflict" ]; then
    for f in CLAUDE.md AGENTS.md; do
      grep -q "^\($bom\)\{0,1\}<!-- trellis:begin" "$git_root/$f" 2>/dev/null || continue
      static_conflict="managed block in $f"
      conflict_paths="the managed block in $f, from its trellis:begin marker to its trellis:end marker (there is no overlay directory in this shape)"
      [ "$f" = AGENTS.md ] && [ "$claude_imports_agents" = no ] && static_host_reads=no
      break
    done
  fi
fi
if [ "$scope" = "project" ] && [ "$opted_out" = yes ]; then
  # TRL-38. Before this branch existed the installer rendered a governing rules
  # file over the opt-out, and Claude Code loaded it at launch with no hook
  # involved — the project got the full rules body in always-loaded context
  # every session, while the hook, the component that honours the opt-out,
  # could only inject an override telling the session to disregard what was
  # already read (decision-0070 D5 calls that second-best, and it is). The
  # opt-out was not merely ignored; the thing that respects it was reduced to
  # arguing with the thing that ignored it. Refusing here is what makes the two
  # deliveries agree on such a project: the hook injects nothing, and now
  # nothing is rendered for it to argue with.
  #
  # The bundle IS vendored: the hook that honours the opt-out ships in it, and
  # /trellis:remove with it. The seed is not the path in — this script never
  # overwrites an existing rules.toml — so the opt-out survives the run. Same
  # shape as the static-delivery refusal below: bundle in, rules file out, said
  # out loud, exit 0. The remedy names the opt-out, not a file to delete.
  rendered_note="no rules file — .trellis/rules.toml declares governed = false"
  say "NOT rendering .claude/rules/trellis.md: .trellis/rules.toml declares governed = false,"
  say "so this project has opted out of Trellis (decision-0070 D5). The plugin hook reads"
  say "that key before every delivery path and injects nothing here, the two floor- rules"
  say "included; a rendered rules file would be loaded at launch regardless, governing a"
  say "project that declined. The bundle is vendored (the hook that honours the opt-out"
  say "ships in it); the rules file is not, and .trellis/rules.toml is left exactly as it"
  say "is — this installer never overwrites it."
  # Refusing does not help against what the host ALREADY loads at launch: a
  # rendered file from an earlier run, or a static overlay / managed block
  # (detected above, and the shapes the hook's own opt-out override names —
  # staleness.sh's TRELLIS_NOT_GOVERNING). Each is live right now, and "no
  # rules file" would be false at the moment it is printed. Review found the
  # first version of this branch warning about the rendered file only, so an
  # opted-out legacy project was told it was not governed while its overlay
  # loaded regardless, with no migration remedy. Both are named, with the
  # remedy each needs.
  loaded_anyway=""
  if [ -f "$git_root/.claude/rules/trellis.md" ]; then
    loaded_anyway=".claude/rules/trellis.md"
    say ""
    say "WARNING: .claude/rules/trellis.md ALREADY EXISTS in this project, so Claude Code"
    say "loads those rules at launch over the opt-out. The hook tells each session to"
    say "disregard them, which is second-best — this installer did not create that file"
    say "and has not removed it. Delete it, or run /trellis:remove."
  fi
  if [ -n "$static_conflict" ] && [ "$static_host_reads" = yes ]; then
    loaded_anyway="${loaded_anyway:+$loaded_anyway and }$static_conflict"
    say ""
    say "WARNING: this project also delivers Trellis rules STATICALLY ($static_conflict),"
    say "and Claude Code loads those at launch over the opt-out — the hook can only tell"
    say "each session to disregard them. To stop them loading, delete $conflict_paths,"
    say "or run /trellis:remove to take Trellis out entirely."
  elif [ -n "$static_conflict" ]; then
    # The one shape this host does not read (TRL-44). $loaded_anyway is NOT
    # touched: nothing from this block is loaded over the opt-out, so it must not
    # join the LEFT IN PLACE summary. A NOTE rather than a WARNING, because for a
    # Claude session there is nothing to warn about. Said anyway, because the
    # block is real for the host that does read AGENTS.md, and an opted-out
    # project is entitled to know its opt-out is not portable.
    #
    # EVERY CLAIM IS ABOUT THE BLOCK, not about the project. This branch is
    # independent of the rendered-file warning above and both fire in the same
    # run: opt-out + un-imported AGENTS.md block + a pre-existing
    # .claude/rules/trellis.md prints the WARNING and then this. Review of this
    # PR caught a first version saying "nothing is loaded over the opt-out here"
    # and "the hook stays silent on this project" — false in exactly that run,
    # where the rendered file IS loaded and the hook emits TRELLIS_NOT_GOVERNING.
    # Reproduced before the reword; the fixture below covers it. A message that
    # generalises from its own branch's condition to the whole project is the
    # defect this whole change exists to remove, so it does not get to reappear
    # inside the fix.
    say ""
    say "NOTE: this project carries a Trellis managed block in AGENTS.md, but CLAUDE.md"
    say "does not import it (no standalone @AGENTS.md line), so Claude Code never reads"
    say "that file: nothing from that block is loaded over the opt-out, and the plugin"
    say "hook gates its own AGENTS.md probe on the same line. Other hosts DO read"
    say "AGENTS.md directly (Codex CLI), so the block still governs there: to stop that"
    say "too, delete $conflict_paths,"
    say "or run /trellis:remove to take Trellis out entirely."
  fi
  if [ -n "$loaded_anyway" ]; then
    rendered_note="LEFT IN PLACE — $loaded_anyway present AND .trellis/rules.toml declares governed = false"
  fi
  say "To adopt Trellis here instead, replace .trellis/rules.toml with the two-line file"
  say "shown at the end of this output, followed by any row it sets to active = false,"
  say "and re-run this installer."
elif [ "$scope" = "project" ] && [ -n "$static_conflict" ] && [ "$static_host_reads" = yes ]; then
  # A pre-plugin-delivery consumer whose CLAUDE.md managed block imports
  # @.trellis/internal/trellis.md. Rendering here would put BOTH static chains
  # into context: Claude loads the managed block's imports AND .claude/rules/*.md
  # itself, before any hook runs. The hook's path-A-first ordering suppresses only
  # what the HOOK injects — it cannot un-load a file Claude already read. So this
  # is refused at install time, because there is no runtime fix for it.
  #
  # Every read this script makes of pre-existing project state — FIFTEEN LINES,
  # six paths. An earlier version of this comment claimed "the ONE place this
  # script reads .trellis/", which was false; a later one claimed no contents are
  # read anywhere, which is false while the marker check exists; the two after
  # those were miscounted, once by adjusting the number and once by not touching
  # it. So the unit is now one a reader can re-derive rather than trust: a LINE
  # of this script, outside a comment, that tests or reads one of the six paths
  # below as `"$git_root/…"` or `"$rules_dir/…"`. Writes are not reads and are
  # not counted (the seed's cp, the render's mv, the mkdir). Counting lines, not
  # purposes, is what lets TestInstallScriptReadEnumerationIsCounted fail in the
  # commit that adds the sixteenth — which is the only reason the last three
  # corrections were needed at all (inv-self-improvement).
  #   - .trellis/rules.toml        8  (existence: the opt-out test, the
  #                                    unreadable-file check, the seed's two
  #                                    guards and the legacy no-file guidance;
  #                                    readable: the unreadable-file check;
  #                                    existence-and-readable once, as the guard
  #                                    on the content read; CONTENT x1: the
  #                                    top-level governed key, refusing over the
  #                                    opt-out — TRL-38. The strictness read
  #                                    TRL-37 added retired with the posture
  #                                    headers — TRL-97)
  #   - .claude/rules/trellis.md   3  (existence: the opt-out branch's ALREADY
  #                                    EXISTS warning, the live double-delivery
  #                                    warning below, and the non-regular refusal
  #                                    before the mv. TRL-38 added the first and
  #                                    this line said "x2" for a release after it)
  #   - .trellis/internal/         1  (existence, above)
  #   - .trellis/trellis.md        1  (existence, above)
  #   - CLAUDE.md                  1  (CONTENT: the standalone @AGENTS.md import
  #                                    line, deciding whether THIS HOST reads
  #                                    AGENTS.md at all — TRL-44 for the opt-out
  #                                    branch's warning, TRL-48 for the refusal)
  #   - CLAUDE.md / AGENTS.md      1  (CONTENT: opening marker, one loop)
  # Three are content reads, and decision-0090 splits them by what they DO, not
  # by how many there are. Two only ever REFUSE or print: the marker, one
  # line-anchored string, and the governed key. One REMOVES A REFUSAL: the
  # @AGENTS.md gate, which since TRL-48 decides whether the static-delivery
  # refusal below fires at all. That is D2's fourth clause, so it came to the
  # corpus before it landed and got decision-0091 — TRL-44 shipped this read
  # wired to the MESSAGE only, and said so in terms, for exactly that reason.
  # A fourth, the strictness key, SELECTED which shipped header was rendered and
  # so was owed decision-0088 under D2; it retired with the second header
  # (TRL-97). None patches a marker or writes anything under .trellis/.
  # spec-0005 AC2's second amendment permitted the marker; the rest are argued,
  # and the count is pinned, in cli/install_script_test.go — decision-0090 D1
  # makes that the count's home. The count is unchanged by TRL-48: the gate moved
  # what an EXISTING read decides, and added no read.
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

  # One header ships (TRL-97): every project is rendered from
  # reference/trellis.md and receives its By default sentence, whatever
  # `strictness` its .trellis/rules.toml carries, and nothing here reads that
  # key. The hook delivers the same header, so the two agree without a copied
  # parser. Before TRL-97 that key selected one of two headers here, read with a
  # copy of the hook's parser (TRL-37, decision-0088); the read, its pair guard
  # and the posture note the confirmation used to print retired together.
  #
  # The unreadable-file check stays, and `-r` is what makes it. A regular file
  # this user cannot read (a shared checkout, a stray chmod) passes `-f`. The
  # `governed` read above is guarded regular-and-readable, so such a file yields
  # an empty $governed_head and $opted_out stays `no`: a project whose rules.toml
  # says `governed = false` at mode 000 is rendered over, and governed by the
  # file it declined. decision-0088 D3 takes that deliberately — an install that
  # stops halfway, with a vendored bundle and no rules file, is the worse state,
  # and the install is the one moment someone reads the output. The hook
  # diverges on the same input: it emits TRELLIS_RULES_NOT_LOADED and injects
  # nothing, and once this rendered file exists it stands down to it (path C),
  # so the rendered file is what governs. So the render goes ahead, and the
  # confirmation below says so on a line of its own.
  #
  # What that line names is the OPT-OUT (TRL-44), the bigger thing this input can
  # lose. It used to lead with a posture that could not be honoured; no posture
  # is read now. And it says every rule applies rather than giving a count: the
  # rendered file imports its rows from the file that cannot be read, and a rule
  # with no row applies (TRL-97), so no opt-out takes effect.
  toml_state=absent
  if [ -f "$git_root/.trellis/rules.toml" ]; then
    toml_state=present
    [ -r "$git_root/.trellis/rules.toml" ] || toml_state=unreadable
  fi
  header="reference/trellis.md"

  # The placeholder must exist or the first sed silently emits the whole file
  # minus its last line and the second emits nothing — a truncated render that
  # every other check would pass. Unreachable through a verified fetch; this is
  # defence in depth against a coordinated payload edit.
  grep -q '^@rules\.md[[:space:]]*$' "$stage/bundle/$header" \
    || fail "$header carries no @rules.md placeholder line; refusing to render a truncated rules file"

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
  # Deliberately NOT a prose check. Probing for the By default sentence or the
  # invariants pointer would re-create the exact defect this branch fixed — a
  # legitimate payload reword breaking every install. `-s` asks only "did this
  # step produce anything", which no reword can falsify.
  sed -n '1,/^@rules\.md[[:space:]]*$/p' "$stage/bundle/$header" | sed '$d' > "$stage/render.head"
  [ -s "$stage/render.head" ] || fail "rendering the header prose produced nothing; the bundle's $header is damaged or has no @rules.md placeholder. Nothing was written."
  cat "$stage/bundle/reference/rules.md" > "$stage/render.body"
  [ -s "$stage/render.body" ] || fail "rendering the rules body produced nothing; the bundle's reference/rules.md is missing or empty. Nothing was written."
  sed -n '/^@rules\.md[[:space:]]*$/,$p' "$stage/bundle/$header" | sed '1d' \
    | sed 's|`\.trellis/internal/invariants\.md`|`.claude/skills/trellis/reference/invariants.md`|' > "$stage/render.tail"
  [ -s "$stage/render.tail" ] || fail "rendering the header tail produced nothing; the bundle's $header is damaged. Nothing was written."
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
    # Framing only (TRL-97). The activation rule lives once, in rules.md above:
    # a rule applies unless its row says active = false. The footer adds the
    # heading and the import that brings the rows in, and no sentence of its
    # own; it used to carry one naming `strictness` as authoritative over a
    # posture sentence that no longer varies. The shape is the one the hook's
    # stage machine walks and cli/plugin_hook_test.go's renderedFile builds.
    printf '<!-- trellis:rendered-footer -->\n'
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
  # file could silently lose the activation sentence, which drops every opt-out
  # (every rule applies whatever its row says).
  #
  # So validate the artifact, not the exit status. This is WEAKER than the hook's
  # reader-side check and is not a mirror of it: the probes are matched
  # independently (no ordering), and `grep -c` counts matching LINES where the
  # hook counts DISTINCT ones. Three of the four probes also check strings this
  # script itself printf's, so they cannot catch a payload-derived step failing —
  # that is the -s checks' job, above. What it does guarantee is that a render
  # which would not satisfy the hook never reaches the mv.
  render_defect=""
  for probe in \
    '<!-- trellis:rendered-begin -->' \
    '<!-- trellis:rules-loaded -->' \
    '<!-- trellis:rendered-footer -->' \
    '@../../.trellis/rules.toml'
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

  # decision-0070 D2. Seed the file, so the adoption this run makes is recorded
  # where every other delivery looks for it. The rendered file does not need it
  # — a rule with no row applies — but the plugin on Codex, and a plugin
  # installed outside the repository, adopt a project only by
  # .trellis/rules.toml (decision-0070 D4 and D7, decision-0077); without it a
  # collaborator on either is asked again, or gets nothing.
  #
  # The bytes are $trellis_seed_comment above an empty [rules] table, written
  # literally rather than copied from the bundle (TRL-97): the file the hook's
  # TRELLIS_NOT_YET_GOVERNING accept instruction quotes, pinned to it by
  # TestInstallScriptSeedMatchesTheHooksAcceptInstruction. It holds only what has
  # effect — no strictness, no seeded_from, no row set active = true.
  #
  # Running an installer INSIDE a repository is the adoption act (D1), which is
  # what makes writing here legitimate where the plugin path must ask. Never
  # overwritten: an existing rules.toml is the project's own and outranks a seed.
  #
  # This amends decision-0065:26-29, the plugin/install split ("install.sh is
  # vendoring and never configures ... and continues to never touch .trellis/") —
  # NOT :18-19's "the setup skill writes exactly one file", which bound it and is
  # untouched. That clause was named here, and in three other places, before the
  # boundary was got right. Openly,
  # in decision-0070 D2, not by routing around it. The clause's purpose, that no
  # path silently vendors an overlay, is untouched: one config file, in the repo
  # the user pointed this script at.
  if [ -e "$git_root/.trellis/rules.toml" ] && [ ! -f "$git_root/.trellis/rules.toml" ]; then
    # A non-regular target must be refused BEFORE the write. When the seed was a
    # copy, `cp file dir` copied INTO the directory and returned 0 — measured,
    # that produced a success banner, seeded_rows=yes, a stray copy buried inside
    # .trellis/rules.toml/, and an import resolving to nothing. Exactly the
    # defect already guarded on the rendered file's move. The seed is a redirect
    # now, which fails on a directory but blocks forever on a FIFO, so the guard
    # stays.
    seeded_rows=failed
  elif [ ! -f "$git_root/.trellis/rules.toml" ]; then
    # `2>` before `>`: redirections apply left to right, so the shell's own
    # "cannot create" message is silenced only when stderr moves first.
    if mkdir -p "$git_root/.trellis" 2>/dev/null &&
       printf '%s\n' "$trellis_seed_comment" '[rules]' 2>/dev/null >"$git_root/.trellis/rules.toml"; then
      seeded_rows=yes
    else
      # Not fatal: the rules file is already in place and every rule applies.
      # Say so rather than exit, because the install itself succeeded.
      seeded_rows=failed
    fi
  fi
fi

# --- 4. Confirm — never a git mutation; the commit is yours -----------------------

say "vendored the Trellis plugin ($stamp) to $target"
say "  $nfiles files written; manifest verify OK on every byte before anything was written"
say "  rules: $rendered_note"
if [ "$rendered_note" = ".claude/rules/trellis.md" ] && [ "${toml_state:-}" = unreadable ]; then
  # decision-0088 D3 and TRL-44, on a line of its own; the render branch says why.
  say "  WARNING: .trellis/rules.toml exists but could not be read (permissions?), so no opt-out in it could be honoured and every rule applies. This installer's governed = false check reads that file too and came back empty: if the file declines Trellis, it has just been rendered over, and this project is now governed by the rules file it declined. The plugin hook does NOT do this — on a file it cannot read it emits TRELLIS_RULES_NOT_LOADED and injects nothing, and once this rendered file exists it stands down to it — so this rendered file governs until you fix the permissions and re-run, which will then honour any opt-out, or the governed = false, that the file holds (decision-0088 D3)."
fi
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
  # TRL-48 / decision-0091 D1. The block this run RENDERED OVER, said out loud.
  # The refusal above no longer fires for it, and a removed refusal that says
  # nothing is a silent decision (floor-transparency). Gated on the opt-out too:
  # that branch never renders and prints its own NOTE for the same shape, so
  # without this test an opted-out project would get both.
  #
  # EVERY CLAIM IS ABOUT THE BLOCK AND THIS HOST, the discipline TRL-44's review
  # imposed on the sibling NOTE: what is asserted is that Claude does not read
  # that file, which the line above decides, and NOT that the project is
  # single-delivery — a nested import chain (CLAUDE.md -> foo.md -> @AGENTS.md)
  # is loaded by the host and seen by neither this gate nor the hook's, which is
  # why the last two lines name the condition rather than promise its absence.
  if [ "$opted_out" = no ] && [ -n "$static_conflict" ] && [ "$static_host_reads" = no ]; then
    say "NOTE: this project carries a Trellis managed block in AGENTS.md, and the rules"
    say "file above was rendered anyway. CLAUDE.md carries no standalone @AGENTS.md line,"
    say "which is the only route to that file this installer can see, and the plugin hook"
    say "reads the same one line — so neither expects Claude Code to load the block, and"
    say "the hook stands down to the rendered file. Hosts that DO read AGENTS.md directly"
    say "(Codex CLI) get their rules from the block and nothing from the rendered file —"
    say "deleting the block would ungovern them."
    say "CHECK THIS IF YOUR CLAUDE.md IMPORTS OTHER FILES: an @-import chain that reaches"
    say "AGENTS.md indirectly IS followed by the host and is NOT seen by either component,"
    say "and then the block and the rendered file are the same rules twice with nothing"
    say "reporting it. If that is this project, delete one of the two static shapes."
    say "Adding a standalone @AGENTS.md line to CLAUDE.md later has the same effect, and"
    say "is the case the hook does catch (TRELLIS_STATIC_SHAPES_CONFLICT)."
    say ""
  fi
fi
case "${seeded_rows:-}" in
  yes)
    say "This project is governed now: all sixteen rules are active, followed by default"
    say "with deviations said out loud. .trellis/rules.toml is yours to edit — a row"
    say "<slug> = { active = false } under [rules] switches that rule off."
    ;;
  failed)
    # Every rule applies without the file, since a rule with no row applies; what
    # is missing is the adoption record the seed exists to write (see the seed).
    say "Could not write .trellis/rules.toml (permissions, or something that is not a"
    say "file at that path). The rules file is in place and every rule applies here, but"
    say "the plugin on Codex, and any plugin install outside this repository, adopt a"
    say "project only by that file. Clear the path if it needs it, then write the file"
    say "yourself, containing exactly these two lines:"
    say_seed
    ;;
  *)
    # Nothing was seeded on this run — either the file already existed, or the
    # render was refused (a static-delivery conflict or an opt-out), or this is
    # personal scope. Those are NOT the same state, and the old code only ever
    # printed one line for all of them. `seeded_rows` is unset here, so ask the
    # disk instead.
    if [ "$scope" = "project" ] && [ ! -f "$git_root/.trellis/rules.toml" ]; then
      # Only a refusal over a legacy static shape reaches here with no file: the
      # opt-out needs the file, and the render path seeds it. The floors-only
      # warning was dropped once, and that was a regression: such a repo runs on
      # two rules and was no longer told so. Under that shape's frozen text a
      # rule applies only when its row says active = true, so the file this
      # installer seeds would change nothing there, and telling the user to
      # write it by hand is the wrong advice (TRL-97). The route is the
      # migration the refusal names, then a re-run, which seeds the file, with
      # any opt-out the old shape carried added back under its [rules].
      say "This project has no .trellis/rules.toml, so only floor-transparency and"
      say "floor-intent-gate apply: the legacy shape's frozen rules text applies any other"
      say "rule only when a row says active = true. Do not write that file by hand — the"
      say "file this installer seeds has no such rows, so under that text it would govern"
      say "only the floors too. Migrate as described above, noting any row the old shape"
      say "sets to active = false; then re-run this installer, which seeds the file, and"
      say "add those rows under its [rules]."
    elif [ "$opted_out" = yes ]; then
      say "This project is NOT governed: .trellis/rules.toml declares governed = false, and"
      say "neither this installer nor the plugin hook delivers any rule over that — the"
      say "floors included (decision-0070 D5). To adopt Trellis here, replace that file"
      say "with exactly these two lines, followed by any row it sets to active = false,"
      say "and re-run this installer:"
      say_seed
    elif [ "$scope" = "personal" ]; then
      # A plugin outside the repository governs a project only once that project
      # has the file; the hook announces and asks in one that has none
      # (decision-0070 D4, decision-0077).
      say "A project is governed once it has .trellis/rules.toml: the first session in a"
      say "project without one asks whether to adopt Trellis there and writes the file if"
      say "you accept. In it, a row <slug> = { active = false } under [rules] switches a rule off."
    else
      say "Edit .trellis/rules.toml to switch a rule off: a row <slug> = { active = false }"
      say "under [rules]. That file is yours, and this installer never rewrites it."
    fi
    ;;
esac
