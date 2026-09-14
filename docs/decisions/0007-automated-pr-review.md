---
id: decision-0007
type: decision
status: ratified
ratified: 2026-06-29
depends_on: [invariants-v0, decision-0005]
owner: gundi
date: 2026-06-29
---

# 0007 — Automated PR review (instantiate invariant 5 on our own repo)

**Raised by:** the maintainer — add Claude PR-review machinery by default, ahead of need.

## Context

Our operating method puts independent verification at the execution layer ("the builder
does not grade itself" — invariant 5). So far that gate has been manual (the maintainer
reviewing PRs). Adding an automated reviewer makes invariant 5 concrete in *our own* build
methodology (Layer B, decision `0005`) — dogfooding before we ship the machinery.

## Decision

Add `.github/workflows/claude-code-review.yml`, running the official `code-review` plugin
(the `/code-review` skill) on every PR (`opened`, `synchronize`). Authenticate via
**`CLAUDE_CODE_OAUTH_TOKEN`** (Claude subscription, no separate API bill); an
`ANTHROPIC_API_KEY` swap is documented in the workflow for direct-API billing.

It is an **assistive** reviewer, not a merge gate — the human still owns the intent gate
(invariant 4). It does not block merges; it surfaces findings.

## Prerequisites (repo admin, one-time — *not* doable from the agent)

1. Install the Claude GitHub App: https://github.com/apps/claude
2. Add the auth secret (`CLAUDE_CODE_OAUTH_TOKEN`). Easiest: run `/install-github-app` in
   Claude Code and choose "Skip for now" at the workflow step (this repo already ships the
   workflow file).

## Consequences

- Every future PR gets an automated review pass with no extra ceremony.
- Until the App + secret are configured, the workflow is present but inert (it will error
  on missing secret — visible, not silent: invariant 8).
- Minimal-first (invariant 9): only the review workflow is added, **not** the `@claude`
  mention workflow — add that later if friction warrants.

## Supersedes / superseded by

— (none)
