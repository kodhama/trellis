---
id: decision-0060
type: decision
status: gated  # maintainer's explicit reset intent on 2026-07-24; records intent for independent review, not agent self-approval
depends_on: [decision-0043, decision-0058, decision-0059, spec-0007@v1, spec-0008@v2]
owner: agent
date: 2026-07-24
---

# 0060 — retire family release certification; keep Trellis's independent local live-rule delivery

> **Human direction (2026-07-24).** The maintainer reset Trellis's scope:
> preserve the independently decided and implemented local hook/live-rule
> work, retire the family release-certification contract and all machinery it
> would require, and make any future headless test depend only on a narrow
> Stewards CI marketplace-provisioning skill. This gated record captures that
> intent for independent review; it does not record agent approval.

## Context

`decision-0058` and `spec-0007@v1` independently define Trellis's current
host-native work: Claude keeps its import path, and trusted local Codex fresh
startups may receive the installed overlay and current live rows through the
Trellis hook, with a small file-reading fallback. That work landed in PR #181
and deliberately does not claim resume, clear, compact, subagent, desktop,
IDE, headless, automation, or cloud support.

`decision-0059` later adopted a family plugin-release contract. Its deriving
`spec-0008` specified product-owned SemVer, immutable tags, release metadata
and inventory, surface certification, support derivatives, release history,
human release approvals, release workflows, and a shared validator-runtime
protocol. None of that machinery is required to preserve the local delivery
contract above.

The family adoption also narrowed `decision-0043`'s standing result that the
old release channel and its machinery were retired. Keeping that narrowing
would make Trellis build a release-certification system solely to participate
in the family contract, contrary to the maintainer's reset.

Future headless testing has one smaller acquisition prerequisite: CI must be
able to make the Trellis plugin available from its marketplace before the
test begins. Stewards can own that mechanical provisioning step without
owning or certifying Trellis behavior.

## Decision

### 1. The independent live-rule contract remains current truth

`decision-0058` and `spec-0007@v1` remain unchanged and authoritative for
Trellis's local hook, installed-overlay, current-row, fallback, host-isolation,
and supported-boundary behavior. This decision neither promotes a candidate
surface nor changes setup, refresh, remove, payload generation, payload
stamping, or host manifest behavior.

The semantics of PR #181 remain exactly its bounded Phase 1 claim. PR #180's
shared `AGENTS.md` authority, PR #184's lifecycle bookkeeping for
`decision-0043`, and PR #187's GitHub Actions runtime upgrade also remain
unchanged.

### 2. The family release-certification adoption retires in full

This decision supersedes `decision-0059`. `spec-0008`, including both its v1
and v2 recorded forms, is retired with no successor implementation spec.
Trellis will not implement that decision/spec pair's:

- family release certification or product-extension validation;
- canonical shared SemVer authority or cross-host version-equality contract;
- immutable release-tag contract;
- release metadata, inventory, surface contract, or generated support
  derivatives;
- release history or release-approval records;
- pre-tag, tag, adoption, or release workflow; or
- shared validator runtime, runtime-store, digest, request/result, sandbox,
  or audit protocol.

No partial requirement from `decision-0059` or `spec-0008` remains current
merely because it was described there. A future Trellis package-version or
release choice, if one becomes necessary, requires a new project-local
decision on its own evidence.

### 3. The earlier Trellis boundaries are not rewritten

`decision-0059`'s narrowing of `decision-0043` point 4 is withdrawn.
`decision-0043` stands as recorded, including its generator-only CLI,
content-derived payload stamp, file-to-file staleness comparison,
plugin-vendor-script distinction, and retired release-channel machinery.

This reset does not revive `decision-0036` as a new dual-host package contract.
Its historical supersession remains auditable; current host behavior stays
where `decision-0058` and `spec-0007@v1` put it. The new decision changes
neither those artifacts nor the code delivered from them.

### 4. Future headless testing has one narrow Stewards prerequisite

If Trellis later tests a headless Codex surface, its only family dependency is
a narrow Stewards-owned CI skill that provisions the Trellis plugin from the
marketplace before agent startup. That skill may establish acquisition and
pre-test availability only.

It must not certify a Trellis release, define Trellis package versioning,
validate Trellis product behavior, own Trellis setup, generate support state,
or promote headless support. Trellis still owns the test and must independently
prove the exact headless context-loading, live-row, failure, and
duplicate-context behavior before changing the support boundary.

### 5. Supersession pointers are the complete artifact propagation

`decision-0059` receives a full `superseded_by: [decision-0060]` pointer and an
append-only supersession record. `spec-0008` receives the matching
`superseded` lifecycle state and forward pointer. There are no release
metadata, support tables, histories, approvals, validators, workflows, or
other product derivatives to update because the contract that would authorize
their creation is retired before implementation.

## Consequences

- Trellis keeps its independently justified local Claude/Codex live-rule
  behavior without taking on family release certification.
- The repository gains no SemVer, tagging, history, approval, surface,
  runtime-store, or validator-runtime machinery from the retired contract.
- `decision-0043`, `decision-0058`, `spec-0007@v1`, and the behavior delivered
  by PRs #180, #181, #184, and #187 remain intact.
- A future headless experiment can acquire the marketplace plugin through one
  bounded Stewards CI skill, while all behavioral claims remain Trellis-owned
  and evidence-gated.
- The reason for adopting and then retiring the family contract stays
  auditable through the append-only decision chain.

## Open questions

- What exact marketplace and CI environment will the future Stewards
  provisioning skill target? Resolve that in the skill's own narrow contract
  when a headless experiment is scheduled, not in Trellis release machinery.

## Self-check

This record consumes only settled upstreams, identifies the exact contract it
retires, and leaves the independent local delivery artifacts unchanged. It
does not infer headless support from marketplace acquisition. Its
`status: gated` records the maintainer's reset intent while leaving approval
to the configured independent/human gate.
