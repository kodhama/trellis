---
id: decision-0079
type: decision
status: gated  # authored + self-checked by agent; the approved flip is the maintainer's intent act (decision-0046)
depends_on: [decision-0011]
changes: [decision-0011]
informed_by: [decision-0075, decision-0076, decision-0078]
owner: agent
date: 2026-08-29
---

# 0079 — Retire the spec stage; planning moves to superpowers

## Context

`decision-0011` added a **spec (behavioral-contract) stage between decisions and build**, and
eight specs were written under it (`spec-0001`–`spec-0008`). Two things have changed the premise:

- **The stage stopped paying for itself.** Of the eight, one is fully superseded (`spec-0008`)
  and four carry `superseded_in_part_by` amendments longer than the clauses they amend —
  `spec-0005`'s single frontmatter comment runs to ~1,400 words. The specs had become a second
  place to maintain what `decisions/` already records, which is the duplication
  `decision-0028` exists to prevent.
- **A planning layer now arrives from outside.** The maintainer has adopted the **superpowers**
  plugin, whose `brainstorming` / `writing-plans` / `executing-plans` skills occupy the slot the
  spec stage held: turn ratified intent into a checkable plan before code. Keeping both means two
  contracts for one job.

This follows `decision-0076` (grove retired) in the same direction: the repo is shedding
process layers it no longer runs.

## Decision

**1. The spec stage retires. `specs/` is deleted.** `decision-0011` is superseded by this
record. Planning between a decision and the build is done with the superpowers skills, which
produce a plan for the work at hand rather than a ratified artifact to maintain afterwards.

**2. `decisions/` and `research/` are unaffected.** They remain append-only and keep their
`spec-*` citations, which are historical (`inv-auditable-archive`).

**3. The artifact contract survives in the rubric, not in a spec.**
`core/rubrics/artifact-contract.md` derived its checks from `spec-0001` §3 and `spec-0002` §4.
Each check was already stated there in full — the spec citations were provenance, not content —
so the rubric becomes the self-standing definition and its `depends_on` re-points to the
decisions that ratified the lifecycle it enforces.

**4. `spec` stays a recognized `type` in the contract.** What retires is the *mandatory stage*,
not the artifact type; a spec written later, here or in a consumer methodology, still has a
contract to meet.

**5. `grove/adr-*` citations stay untouched.** `decision-0076` warned that grove-the-repo
citations are load-bearing and named `spec-0001`'s dependency on `grove/adr-0010` as one of two
anchors. Deleting `specs/` removes that anchor; the other — `decision-0045` superseded in part
by `grove/adr-0010` — is a `decisions/` record and survives, so the rule and its example still
stand.

## Consequences

- **`specs/` no longer exists.** Every current-truth surface citing a spec has been re-pointed
  at a decision or had a decorative citation dropped; `ratify-guard` no longer globs `specs/*.md`.
- **Trellis-core loses no shipped rule.** Catalog, invariant set and rubric keep their content;
  only provenance pointers moved.
- **Cost accepted:** `spec-0005` and `spec-0007` were the written contracts for the curl-install
  and Codex live-rule paths. Their requirement ids (`AC#`, `R##`, `S##`) survive as markers in
  `cli/install_script_test.go`, `cli/codex_hook_test.go` and `install.sh`, whose tests are now
  the executable statement of those contracts. The retired text remains in git history.
- **No orphan follow-ups (`decision-0078`).** This record parks nothing. The one question it
  raised — whether superpowers plans should be retained in-repo — is **dropped**, not deferred:
  plans are session scaffolding, and what survives a change is the decision and the diff.
