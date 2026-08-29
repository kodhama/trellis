---
id: decision-0079
type: decision
status: approved  # maintainer's intent act 2026-08-29, in conversation ("Approve approve approve") — this flip RECORDS that act, it does not perform it (decision-0046). Author (agent) != approver (maintainer)
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

**3. The artifact contract survives in the rubric; the typed-artifact schema moves to `core/`.**
`core/rubrics/artifact-contract.md` derived its *checks* from `spec-0001` §3 and `spec-0002` §4,
and every check was already stated there in full — those citations were provenance, not content —
so the rubric becomes the self-standing definition of the contract.

`spec-0002` §1–§3 is different: it is the **authoring schema** for the two typed artifacts (catalog
field list; profile delivery axes and per-invariant fields; the D2 lifecycle gate), and
`decision-0016` fixed only their *existence and scope*, explicitly deferring schema and lifecycle
to the spec. Deleting it would have left `core/catalog/` and `profiles/` with no contract to be
authored against. Those sections are migrated **unchanged in substance** to
**`core/schemas/typed-artifacts.md`** (`id: schema-typed-artifacts`) — product definition moving
into `core/`, where the product lives. Only cross-references were retargeted.

**3a. Retired-artifacts registry.** Deleting `specs/` would otherwise dangle every retained
`depends_on: [… spec-000N …]` — 17 artifacts carry one, including `decision-0076`. Rubric check 4
and the `corpus-reviewer` checklist gain a clause resolving these ids against the registry below,
the same historical-reference exemption the invariant-set's Identifiers registry already grants
(`decision-0013`): *a retirement does not reach back and edit the append-only records that cite
it.* No citing artifact is edited.

| Retired id | Was | Where its content lives now |
|---|---|---|
| `spec-0001` | spine artifact contract | `core/rubrics/artifact-contract.md` (checks 1–7) |
| `spec-0002` | catalog + profile schema and lifecycle | `core/schemas/typed-artifacts.md`; checks in the rubric (8–11) |
| `spec-0003` | advisor delivery machinery | superseded in part by `decision-0043`/`decision-0072`; the surviving path is `install.sh` + the plugin |
| `spec-0004` | clean exits, uninstall and remove | `plugins/trellis/skills/remove/SKILL.md` + `cli/remove_skill_test.go` |
| `spec-0005` | curl-install mechanical vendoring | `install.sh` + `cli/install_script_test.go` (its `AC#` markers) |
| `spec-0006` | shared project-instruction entrypoints | `decision-0057` + `cli/selfapply_test.go` |
| `spec-0007` | local Codex live-rule delivery | `decision-0058` + `cli/codex_hook_test.go` (its `R##`/`S##` markers) |
| `spec-0008` | plugin release and surface contract | already `superseded_by: [decision-0060]` before this record |

Their text remains in git history; the registry is what makes the references resolve.

**4. `decision-0011` carries the forward pointer.** Its status becomes `superseded` with
`superseded_by: [decision-0079]` and a pointer note appended under its existing
*Supersedes / superseded by* heading. Nothing else in it is edited — without this, a reader
entering at `0011` still sees a mandatory spec stage as live.

**5. `spec` stays a recognized `type` in the contract.** What retires is the *mandatory stage*,
not the artifact type; a spec written later, here or in a consumer methodology, still has a
contract to meet.

**6. `grove/adr-*` citations stay untouched.** `decision-0076` warned that grove-the-repo
citations are load-bearing and named `spec-0001`'s dependency on `grove/adr-0010` as one of two
anchors. Deleting `specs/` removes that anchor; the other — `decision-0045` superseded in part
by `grove/adr-0010` — is a `decisions/` record and survives, so the rule and its example still
stand.

## Consequences

- **`specs/` no longer exists.** Every current-truth surface citing a spec has been re-pointed
  at a decision or had a decorative citation dropped; `ratify-guard` no longer globs `specs/*.md`.
- **Trellis-core loses no shipped rule.** Catalog, invariant set and rubric keep their content;
  the typed-artifact schema moved into `core/schemas/` rather than being dropped, and the rest
  is provenance pointers.
- **Three findings from the automated review on PR #255 are folded in** — the dangling
  `depends_on` set (registry, D3a), the typed-artifact schema hole (migration, D3), and the
  missing `decision-0011` forward pointer (D4). Each was verified against the source before
  being accepted.
- **Cost accepted:** `spec-0005` and `spec-0007` were the written contracts for the curl-install
  and Codex live-rule paths. Their requirement ids (`AC#`, `R##`, `S##`) survive as markers in
  `cli/install_script_test.go`, `cli/codex_hook_test.go` and `install.sh`, whose tests are now
  the executable statement of those contracts. The retired text remains in git history.
- **No orphan follow-ups (`decision-0078`).** This record parks nothing. The one question it
  raised — whether superpowers plans should be retained in-repo — is **dropped**, not deferred:
  plans are session scaffolding, and what survives a change is the decision and the diff.
