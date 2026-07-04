---
id: decision-0023
type: decision
status: draft
depends_on: [decision-0010, decision-0012, decision-0019]
owner: gundi
date: 2026-07-05
---

# 0023 — Trellis's first code: Go, single binary, no package manager; the dev cycle

**Raised by:** the maintainer — the setup CLI (`spec-0003`) is Trellis's **first code**. So far the
repo has been instructions only, with no code dev cycle; now one is needed, and the distribution must
not depend on a package manager.

## Context

Constraints, load-bearing:

- **No package-manager dependency.** The maintainer's enterprise **cannot reach the npm registry**, and
  the internal JFrog mirror isn't configured (keys + access — real friction). For a *locked-down
  enterprise* — Trellis's own target buyer — an npm/npx install path isn't friction, it's a **wall**.
- **curl-installable, cross-platform**, run once for setup.
- **Not a runtime** (`decision-0010`): the CLI is **setup tooling** — you run it to onboard/compose;
  the host's agents then follow instructions. It is *not* in the agent critical path, so shipping a
  binary does not make Trellis a runtime dependency.

## Decision

1. **Language: Go.** A single static cross-platform binary; **no package manager for the user** (which
   sidesteps the npm/JFrog wall entirely); good interactive-prompt libraries; shells out to `git`
   cleanly; simple `go build` / `go test`.
2. **Distribution: GitHub Releases + a `curl … | sh` installer** that detects OS/arch and downloads the
   release binary — **no registry (npm or JFrog)**. Repo **Releases** get activated when the first
   binary ships; a stable install URL (GH raw / a redirect) is the curl endpoint.
3. **Dev cycle: `go test` + GitHub Actions CI** (build · test · vet/lint) required on PRs; the code
   lives in its own module (e.g. `cli/`). **Test-first for non-trivial logic**, riding the same PR
   ritual as the docs (`decision-0007` review + `decision-0022` merge=ratify).
4. **Scope of "no runtime" preserved.** Core (invariants, gates) stays instructions; the CLI is
   support tooling only. No product feature requires the binary at agent-time.

## Consequences

- **Releases + a curl endpoint** on the repo (activated with the first binary).
- **A `cli/` Go module + CI** (build/test/lint); `CLAUDE.md` gains a short "code dev cycle" note
  (test-first, CI-gated, merge=ratify applies to code too).
- **`spec-0003` assumes this stack** — the setup CLI is Go, curl-installed.

## Open questions

- **Curl endpoint host** — GH raw / releases now; a custom domain later (deferred; open per maintainer).
- **Interactive-prompt library** (e.g. `bubbletea` / `survey`) — a build-time detail, not decided here.
- **Module layout** (`cli/` vs a top-level `main`) — settle in the scaffolding PR.

## Supersedes / superseded by

— (none)
