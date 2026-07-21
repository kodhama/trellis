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
# principle as the rest of this family). It NEVER touches a project's .trellis/ —
# that is /trellis:setup's job entirely, not this script's — and it NEVER runs a git
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
# instead: plugins/trellis/reference/checksums covers only the 11 rendered M1 payload
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

This is the ONLY decision this script makes. Everything else — posture, which
instructions file to patch, and so on — is asked by /trellis:setup once the plugin
is on disk; see the "next steps" this script prints when it finishes.

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
NONINTERACTIVE=0
while [ $# -gt 0 ]; do
  case "$1" in
    --scope)     [ $# -ge 2 ] || fail "--scope needs a value (personal or project)"; SCOPE_FLAG="$2"; shift ;;
    --scope=*)   SCOPE_FLAG="${1#--scope=}" ;;
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
if [ -n "$SCOPE_FLAG" ]; then
  requested="$SCOPE_FLAG"; requested_origin="--scope"
elif [ -n "${TRELLIS_SKILLS_SCOPE:-}" ]; then
  requested="$TRELLIS_SKILLS_SCOPE"; requested_origin="\$TRELLIS_SKILLS_SCOPE"
fi
if [ -n "$requested" ]; then
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
trap 'rm -rf "$stage"' EXIT

# The bundle manifest — baked in, covers the whole plugins/trellis/ tree. Advance-
# guarded by cli/install_script_test.go:TestInstallScriptBundleManifestIsCurrent.
bundle_manifest() {
  cat <<'TRELLIS_BUNDLE_MANIFEST'
9953fcbc0a2a8de509c2bdc585b72a67e9cf1091d05e71ef09a5e6ab50c1c3aa  .claude-plugin/plugin.json
93f58d26e41505d253ac75fae6cf5f7792db02674da92737cc37ae36534c6729  README.md
a289f0cd911c4392a89f3339d03feead7a2735dacfb893ff886ccb625bd2c809  hooks/hooks.json
3becb23c17b78140a666dcccbaae14657cb5180320b887874e81ce5f5b63fd75  hooks/staleness.sh
a224cdcb7a0e2cb1b47c267a3d662d49f840aa49bc9390e21a5f04d451a6cd5c  reference/block-claude.md
c277d931c9f8512e948b8d79e50d7c60859b1f875f4f5e682ba07a228890a0a7  reference/block-inline-a-head.md
2b91b456eedd0815d6d5b685e7bb3534bddbb3c63e02669755a461fcff1e431b  reference/block-inline-a.md
32d15b7d14c252c97a08e1a900e01ebef31a954738fb5f888e8b47f9512bcaa6  reference/block-inline-b-head.md
c81b5b9b99032b6689bcc00f8ab5c886c18c2e1487a0fb4252d39ef1c22c909f  reference/block-inline-b.md
a33f9904a063986caab0ccc156a229b959c232829c1f1b90c252377d3c028795  reference/block-inline-tail.md
d44d9ffe5764261a79a105873917762fcc50ed943a6b97efe1027f20213f803d  reference/checksums
10ac03e9e3e519abf012f82b0833bda5562417b9c1cfb5d6395481418a6edf97  reference/invariants.md
a675233ee08c0c41b5c0490a163f4d6ff4e95c6bbf9964eac59e4772f6597454  reference/rules-a.toml
534c9178b4c5173f6dd51f382a48b970cc263a83d410c0e9fff6c41a7c937386  reference/rules-b.toml
fee59a9e8da798ade502cbff6284a4e543eafbf72d3178d110f8b482d8e67a0b  reference/rules.md
e61d7cdd4419141e94d5a9ce86a804a5cdba05cd0f1e89744cd526dc034bb625  reference/trellis-a.md
7d479f89409416a0fffe147080de576976a289bf81394c0ca5a874a3950520ee  reference/trellis-b.md
63949812931e0b0768b0fef5209341b83b0bfeee73c49c406cf6474556ec8c39  reference/version
ba28856e52f5fa5ada5bd7e978c57f8f33a28743ca8652352819abcf1948aa0b  skills/remove/SKILL.md
d8c50e60f9746c7fbee9d6c6000b1d003b10ff7a0df10d7eb4956e20a78b3dc4  skills/setup/SKILL.md
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

# --- 3. Write — overwrite the plugin's own files; .trellis/ is untouched, always -
#         (the setup skill owns .trellis/ entirely; this script never looks at it,
#         and this script never runs a git command that mutates anything)

mkdir -p "$target"
for f in $bundle_files; do
  mkdir -p "$target/$(dirname "$f")"
  cp "$stage/bundle/$f" "$target/$f"
done
chmod +x "$target/hooks/staleness.sh"

stamp="$(head -n1 "$stage/bundle/reference/version" 2>/dev/null | tr -d '[:space:]')"
nfiles="$(printf '%s\n' "$bundle_files" | wc -l | tr -d ' ')"

# --- 4. Confirm — never a git mutation; the commit is yours -----------------------

say "vendored the Trellis plugin ($stamp) to $target"
say "  $nfiles files written; manifest verify OK on every byte before anything was written"
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
  say "  git -C \"$git_root\" add .claude/skills/trellis && git -C \"$git_root\" commit -m 'chore: vendor the Trellis plugin'"
fi
say ""
say "Then run /trellis:setup in the project you want to govern. That skill (the real"
say "interactive writer — LLM-driven, no decision logic in this script) reads your"
say "posture, writes .trellis/, patches your instructions file, and verifies itself."
