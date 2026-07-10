---
id: decision-0043
type: decision
status: gated
depends_on: [decision-0010, decision-0025, decision-0035, decision-0036, decision-0039, decision-0041]
owner: gundi
date: 2026-07-10
---

# 0043 — Generator-only CLI; the overlay stamp is the payload stamp, compared file-to-file

## Context

`kodhama-0007` rule 5 (ratified in `kodhama/kodhama`) retired trellis's end-user binary channel
— tap deprecated, curl/brew gone, M2 hosting moved to the setup skill — and left one question
open ("whether the Go code survives generator-only or is replaced is an implementation
question"). Implementing it in #120 surfaced a second, sharper conflict: the shipped staleness
scheme (install stamp `plugin@<sha>`, `hooks/staleness.sh` matching `plugin@*` against the
plugin checkout's git HEAD — `decision-0039` rule 2) contradicts `kodhama-0007` rule 3's letter
and the file-to-file payload-stamp compare already running in the first migrated consumer
(math-quest). The maintainer ruled on 2026-07-10 (issue #120, addendum 4): payload-stamp
compare wins. This record resolves both, so neither lives only in a PR body.

## Decision

1. **The Go code survives generator-only.** The command surface is `payload` (+ `version`/
   `help`): render the pre-built bundle + checksum manifest into the vendored home,
   `plugins/trellis/reference/`. The interactive `setup` TUI, `status`, `remove`, `uninstall`,
   the harness detection, and the binary's M2 path are deleted, not merely undocumented — their
   live homes are the plugin skills (`/trellis:setup` incl. the M2 morph, `/trellis:remove`),
   the bundled staleness hook, and the README's manual copy path. The package's tests remain the
   CI guards (regenerate-and-diff on the vendored payload, the repo-overlay sync-guard,
   docs-consistency, the hook contract). The CLI is dependency-free again (`x/term` retired with
   the TUI — `decision-0030` mooted). "Replaced by a simpler generator" was rejected: the render
   code already is the generator, shares its parsing with the guards, and rewriting it would
   re-derive rendered content — the drift class `kodhama-0007` exists to kill.

2. **The overlay stamp is the render stamp** (maintainer ruling, #120 addendum 4). Every
   *installing* writer — the setup skill, the manual copy path — copies the payload's `version`
   file verbatim to `.trellis/version` (`payload@<content-hash>`); `plugin@<sha>` stamping goes
   away entirely. One deliberate exception: the repo's own self-hosted overlay carries no stamp
   (`.trellis/version` is gitignored) — its currency is guarded by CI regenerate-and-diff, not a
   stamp, so a per-copy stamp would be noise (rationale in `TestRepoOverlayIsCurrent`'s doc
   comment and the `.gitignore` entry, per `decision-0035`). This supersedes `decision-0039`
   rule 2 (the stamp format — its rule 1, SessionStart + `additionalContext` as the only
   agent-facing surface, stands unchanged) and folds the stamp into the manifest-verified copy
   set.

3. **Staleness is a file-to-file compare; `trellis status` retires.** `hooks/staleness.sh`
   compares `.trellis/version` against the installed plugin's `reference/version`: warn on
   mismatch, no-op when either side is missing or empty. No git, no binary, no network. Legacy
   stamps (`plugin@<sha>`, CLI semver) mismatch by construction and draw a one-time migration
   nudge — deliberate: with the binary gone their old surface (`trellis status`) no longer
   exists, and `decision-0035`'s floor ("drift is made visible, not silent") must survive the
   retirement. This supersedes `decision-0035`'s rule-3 surface (`trellis status`) and
   `decision-0036`'s "CLI path keeps its own drift surface" consequence; the repo-facing half of
   `decision-0035` (CI regenerate-and-diff — drift *impossible*, not merely visible) is
   untouched.

4. **Release machinery retires with the channel.** `auto-release.yml` (including its tap-notify
   dispatch), `release.yml`, and `install.sh` are removed; the M2 CLI e2e workflow goes with the
   deleted M2 code path. The shipped artifact is the vendored payload at HEAD (plugin versions
   are commits, `decision-0036`), so `decision-0025`'s guard 2 (auto-release) has nothing left
   to guard; its guard 1 (docs-consistency) stands. The tap formula is deprecated in
   `kodhama/homebrew-tap` with a pointer to the plugin + manual paths (`decision-0041`
   superseded in part; `v0.2.29` binaries stay downloadable, frozen).

   > **Note (2026-07-10, `kodhama/trellis#124`): a file also named `install.sh` returned at the
   > repo root — a different, much smaller artifact class than the one removed above, not a
   > reversal of this rule.** The retired script (this rule) was a release-era end-user *binary*
   > installer, tied to the tap/release channel that also died with `kodhama-0007` rule 5. The
   > #124 script downloads no binary and makes no product decision: it is a **plugin vendor
   > script** — it fetches the whole `plugins/trellis/` tree, verifies it against a manifest baked
   > into the script, and writes it to disk as a [skills-directory
   > plugin](https://code.claude.com/docs/en/plugins-reference#skills-directory-plugins) (project
   > or personal scope). Every decision the retired script used to make (posture, target file,
   > block style) stays exclusively in `plugins/trellis/skills/setup/SKILL.md`, run unmodified once
   > the plugin is on disk — the #124 script is one layer further out than that skill, vending the
   > *plugin*, never the *overlay*. First attempted as `#128` (reimplemented the setup skill's
   > decision logic a second time in shell — closed without merging, the exact drift class
   > `kodhama-0007` exists to close); `#124`'s corrected version is the one that landed.

## Consequences

- Forward annotations added in the same PR: `decision-0025`, `-0030`, `-0035` (pointer in #120's
  diff), `-0036`, `-0039`, `-0041`, and specs `0003`/`0004` (binary-era flows preserved as
  ratified records).
- Consumers with pre-#120 overlays get one staleness nudge on their next session with the
  updated plugin; a refresh restamps them onto the payload vocabulary. math-quest already
  conforms.
- The repo's own overlay refresh is now the manual copy path applied to ourselves (`cp` from
  `plugins/trellis/reference/` — see `TestRepoOverlayIsCurrent`'s doc comment).
- The eval harness's +trellis arm applies the overlay by mechanical copy instead of shelling out
  to `trellis setup` (`eval/run.sh`).

## Open questions

- The marketplace-side gap is unchanged (`decision-0035`/`-0039` open questions): the hook says
  the overlay is behind the *installed* plugin, not how far behind the *marketplace*; plugin
  auto-update stays off by default for third-party marketplaces.
- Non-Claude harnesses have no SessionStart hook; their staleness check is the same file compare
  run by hand (`cmp .trellis/version <clone>/plugins/trellis/reference/version`), documented on
  the manual path. If a real non-Claude consumer materializes, consider a checkable home for it.

## Supersedes / superseded by

Supersedes in part: `decision-0025` (guard 2), `decision-0035` (rule-3 status surface),
`decision-0036` (CLI drift-surface consequence), `decision-0039` (rule-2 stamp format),
`decision-0041` (trellis's own formula + formula-sync mechanics). Moots `decision-0030` (the
TUI's dependency). Implements `kodhama-0007` rule 5 and its rule 3 as ruled in #120 addendum 4.
Consistent with `decision-0010` (the CLI was always optional support, never a runtime).
