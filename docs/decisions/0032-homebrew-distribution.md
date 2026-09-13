---
id: decision-0032
type: decision
status: superseded
superseded_by: [decision-0041]
depends_on: [decision-0023, decision-0028]
owner: gundi
ratified: 2026-07-05
---

> **Superseded by `decision-0041`**: the tap is now the kodhama family's shared
> `kodhama/homebrew-tap`, not a tap this product owns (`gundisalwa/homebrew-trellis`). This
> decision's formula-sync mechanics (regeneration on release, the dispatch-token flow, the
> `curl` fallback) still apply verbatim against the new tap — only the tap's address/ownership
> changed. Content below is unedited, kept for the record.

# 0032 — Homebrew as a second install channel, formula kept in sync with releases

## Context

`decision-0023` made the CLI a single static binary installed via `curl … | sh` — no package manager.
That is the portable floor, but on macOS/Linux **Homebrew** is the idiomatic install and the one people
expect (`brew install …`). Adding it means a **tap** (Homebrew requires its own `homebrew-<name>` repo)
and a **formula** that pins a version + `sha256` — which, like any pinned derived artifact, goes stale on
every release unless something re-pins it (the `decision-0028` problem).

## Decision

- **A tap repo, `gundisalwa/homebrew-trellis`**, with a **binary formula** (`Formula/trellis.rb`) that
  downloads the correct pre-built release binary per platform (darwin/linux × arm64/amd64). No `go`
  build-dependency; install is a download, not a compile.
- **The formula is a derived resource with a sync guard** (`decision-0028`): it is **regenerated on each
  release**, never hand-edited. `scripts/gen-formula.sh` recomputes version + urls + sha256 from the
  latest release; `.github/workflows/update-formula.yml` runs it on a `new-release` `repository_dispatch`
  (or manual `workflow_dispatch`). The main repo's `auto-release` dispatches that event.
- **The cross-repo dispatch needs a fine-grained PAT** (`TAP_DISPATCH_TOKEN`, Contents:write on the tap)
  stored in the main repo. The step is a **no-op until the secret exists**, and the tap's manual *Run
  workflow* re-pins in the meantime — so nothing is broken before the token is added.
- **`curl … | sh` stays** the no-Homebrew fallback; the LP and README show both.

## Consequences

- New public repo `homebrew-trellis`; the LP install block and README gain the `brew install` line.
- `decision-0023`'s single-static-binary property is preserved — Homebrew just fetches that same binary.
- The formula can't drift from releases: a release fires the dispatch, the tap re-pins itself.

## Open questions

- Submitting to **homebrew-core** later (needs notability + a stable release cadence) so `brew install
  trellis` works without the tap.
- macOS **Gatekeeper**: the downloaded binary is unsigned; if quarantine becomes a friction, sign/notarize
  the release artifacts.
