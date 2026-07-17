---
id: decision-0048
type: decision
status: draft
depends_on: [invariants-v1, signature-catalog-v1]
informed_by: [spec-0005, decision-0043]
owner: agent
date: 2026-07-17
---

# 0048 — /trellis:setup asks how to land, and never mutates git remotely

## Context

M1 setup (`/trellis:setup`, steps 1–7) writes the overlay into the working tree and stops —
it says **nothing** about how the change reaches version control. That silence is the defect:
with no landing guidance, the running model fills the gap with its own default, and on a real
install (math-quest, 2026-07-17; reported by the maintainer) that default was **"commit to
`main` and push"** — a silent, unreviewed publish to a shared branch.

Both of setup's neighbouring surfaces already hold the opposite line — the tool never lands the
change; the human does:

- the curl `install.sh` path *"never runs `git add`/`git commit`; it prints the command and
  leaves the commit to you"* — committing is *"the human's decision to make"* (`spec-0005` §4,
  AC9/AC10);
- the **M2 morph** (same skill, step 5) hands the diff to the human: *"the merge is theirs,
  never yours."*

M1 was the one setup surface with no landing rule, so it was free to do the exact thing the pack
exists to prevent. The silent-push default also violates Trellis's own invariants:
`inv-handover-points` (*"Work in reviewable steps with clear stopping points — a plan, a spec, a
PR — not one unbroken stream"*) and `floor-intent-gate` (never finalize/ship/merge at a human's
intent locus without the human).

## Decision

**M1 setup stops at a reviewable seam and asks the human how to land the overlay, every run. It
never commits to the default branch, never pushes, and never opens a PR itself — no remote
mutation ever happens inside the skill.** The choices, PR recommended:

1. **Open a PR (recommended)** — the skill does the **local** half only: branch off the current
   ref and commit **only the files it wrote** (`.trellis/` + the patched instructions file, never
   `git add -A`), then **prints** the `git push … && gh pr create …` for the human to run. The
   remote half is theirs.
2. **Commit to a new branch** — the same local, scoped commit; no push.
3. **Leave it in the working tree** — no git action.

Nothing to land (idempotent refresh, or no git repo) → nothing to do. No human to answer (an
autonomous run) → leave it in the working tree and report; never mutate git unasked.

**Why hand off the remote step rather than perform it.** The push and the PR-create are the only
*remote* mutations in play, and the incident that motivated this was exactly a surprise remote
push. Keeping the skill's own git actions **local-only** (branch + scoped commit, at most) makes
that surprise structurally impossible and mirrors `install.sh`'s "print the command, leave it to
you" precedent. This is a chosen **for-now** boundary, not a logical necessity: the maintainer
noted a push to a **freshly created branch, with the choice asked every run**, would be acceptable
too (2026-07-17). That looser option is recorded below rather than baked in.

## Consequences

- `plugins/trellis/skills/setup/SKILL.md` gains a landing step (new step 8) encoding the
  three-way ask and the local-only rule; step 7 stays "confirm what was written." **No
  payload/checksum change** — `SKILL.md` is skill logic, not manifest-covered bundle content.
- The skill now performs *local* git (branch + scoped commit) on an explicit choice — a capability
  it did not have. This does **not** contradict `spec-0005` AC9 ("no git mutation, ever"): that AC
  binds the `install.sh` **script**, a different tool. The script and the skill keep the same
  *outward* line — no remote mutation, the human owns the merge — by different means.
- Nothing downstream described the M1 landing step (it did not exist), so there is no derived
  surface to re-sync (`decision-0028`) beyond the skill itself.

## Open questions

- **May the skill perform the remote push/PR** on the explicit "open a PR" choice — freshly
  created branch, asked every run — which the maintainer signalled is acceptable, rather than hand
  off the printed command? Kept as hand-off for now; revisit if printing the command proves to be
  friction.
- **Non-Claude harnesses** reach setup via the manual copy path, not this skill, so they get no
  landing prompt. Is a parallel note needed there, or is "you copied the files, now commit them"
  self-evident on that path?

## Self-check (gate)

Grounded in quoted, ratified sources: two invariants (`inv-handover-points`, `floor-intent-gate`
— `signature-catalog-v1`) and `spec-0005`'s enforced no-git-mutation precedent. Framed honestly
as a **chosen for-now boundary, not a deduction** — the maintainer's "a fresh-branch push asked
every run is fine too" is recorded as an open question, not suppressed. `depends_on` carries only
genuine coupling (the invariants the behaviour implements); `spec-0005` and `decision-0043` are
`informed_by` (precedent and host context, not correctness-contingent) per `decision-0047`.
Depends only on ratified artifacts; no draft consumed. Left at `draft`: the author does not grade
its own decision — gating and the `approved` intent act are the maintainer's (`decision-0046`),
recorded by the in-PR flip, ideally after an independent pass (`inv-independent-judgment`).
