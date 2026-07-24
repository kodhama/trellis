---
id: decision-0058
type: decision
status: approved  # maintainer intent act 2026-07-24: document the rollout phases and start implementing them; independent SOUND pass preceded this flip
depends_on: [decision-0039, decision-0043, decision-0051, decision-0053, decision-0057]
owner: agent
date: 2026-07-24
---

# 0058 — phase host-native live-rule delivery; keep the installed overlay as the common truth

> **Human direction (2026-07-24).** The maintainer rejected refresh-time rule
> activation and asked for a progressive rollout: keep Claude's working import
> delivery, begin with the locally proven Codex hook path, validate other Codex
> lifecycle events and surfaces before claiming them, and retain a best-effort
> file-reading fallback rather than stopping merely because a hook marker is
> absent. They also distinguished live row edits from an explicit preset reset.
> This record turns that direction into independently gated phases.

## Context

Trellis already has one sound delivery path. Claude's managed block imports the
installed generated payload and `.trellis/rules.toml`; each new session therefore
sees the current consumer-owned rows without a Trellis refresh
(`decision-0053`). `decision-0057` deliberately left that block in `CLAUDE.md`
while making `AGENTS.md` the shared project-instruction authority.

Codex reads `AGENTS.md`, but does not follow Claude's `@` imports. On
2026-07-23, after installing the prototype as a local Codex plugin and starting
it in a trusted repository, the maintainer reported the model-visible result:

```text
HOOK=TRELLIS_HOOK_CONTEXT payload@0760a802ccd1
RULE=inv-handover-points
```

That is one live positive control for local Codex startup: the hook-carried
marker and one payload rule reached the model. It is not evidence for resume,
clear, compact, subagents, desktop, IDE, automation, or cloud. Phase 1's
production contract tests must independently establish the event schema and
failure behavior used by the build; broader live probes remain promotion gates.

Two tempting shortcuts both fail the product requirement:

- copying the complete rules into `AGENTS.md` duplicates generated policy and
  makes row changes wait for refresh; and
- requiring a hook proof and stopping when it is absent makes an optional
  transport failure fatal even when the installed overlay is readable.

The installed `.trellis/` overlay is already the inspectable common artifact.
The host integration should transport that artifact, not create a second rules
authority.

## Decision

**1. The stable contract is payload plus live rows, not a particular
transport.** Every supported host receives the installed generated rule
payload from `.trellis/internal/` and the current consumer-owned
`.trellis/rules.toml`. A manual row edit takes effect at that host's next
documented context-loading boundary, without `/trellis:setup`, payload refresh,
or regeneration. “Immediate” means that next boundary; Trellis does not claim
to mutate a model context already in flight.

The generated payload carries a small, stable loaded-context sentinel. Native
transports may add the installed `payload@…` stamp for diagnostics, but the
stamp is not a second authority.

**2. Claude's existing import transport stands.** Setup and refresh retain the
managed import block in `CLAUDE.md`. It imports
`.trellis/internal/trellis.md` and `.trellis/rules.toml` exactly as today.
Claude therefore receives one copy of the rule payload: the shared bootstrap
described below recognizes the already loaded sentinel and does not read it
again.

**3. Phase 1 supports local Codex setup and startup through the plugin.** The
Trellis plugin gains a Codex manifest, exposes setup instructions that branch
on the current host without weakening Claude's existing path, preserves the
product-wide remove contract, and adds a host-gated
`SessionStart(startup)` hook.
The hook reads the installed project's `.trellis/internal/trellis.md`,
`.trellis/internal/rules.md`, `.trellis/internal/version`, and current
`.trellis/rules.toml`; it emits one Codex `hookSpecificOutput.additionalContext`
block containing their effective rule content and a diagnostic sentinel. It
does not inject Trellis into Claude, and Claude's staleness hook does not emit
its Claude-shaped output into Codex.

Phase 1's supported claim is deliberately narrow: setup/refresh/remove and a
fresh startup through a trusted local Codex plugin. Codex setup seeds or
refreshes the same `.trellis/` overlay and manages only the Codex bootstrap in
`AGENTS.md`; Claude setup manages only its import block in `CLAUDE.md`. Running
both is safe and produces one transport per host. Remove remains the existing
product-wide clean exit (`spec-0004`): it strips every Trellis-managed host
block byte-safely and then deletes the shared overlay. Phase 1 adds no
per-host disable operation. Direct contract tests cover the event schema, host
isolation, current-row reads, missing-file behavior, setup/remove
reversibility, and single-copy assembly. They do not expand the
supported-surface claim.

**4. `AGENTS.md` carries a receipt and fallback, never the rules.** Setup adds a
separately marked Trellis-managed Codex bootstrap to `AGENTS.md`. Its behavior
is:

1. if the loaded context already contains the Trellis sentinel and activation
   rows, use that context and do not load another copy;
2. otherwise, read the installed generated payload files and current
   `.trellis/rules.toml` before substantive work; and
3. fail visibly only when those required files are absent, unreadable, or
   invalid — not merely because the hook did not run.

This fallback is intentionally best-effort model instruction, not claimed
deterministic enforcement. It gives Codex a useful degraded path and lets
Claude coexist with the shared `AGENTS.md` without double-loading its existing
imports. Refresh updates the small bootstrap in place; remove strips it
byte-safely.

**5. Expansion is evidence-gated by lifecycle and surface.**

| Phase | Candidate delivery | Claim required before promotion |
|---|---|---|
| 1 | Local Codex setup/refresh/remove and fresh startup | Existing live startup positive control plus production contract and reversibility tests |
| 2 | Codex resume, clear, compact, and subagents | One live end-to-end positive and negative control for each boundary; prove no duplicate context |
| 3 | Codex desktop, IDE, headless/automation, and cloud surfaces | Verify hook availability, trust/install behavior, context shape, and live-row reload separately on each surface |
| 4 | Other hosts, including any Claude hook replacement | A host-native proof at least as strong as Phase 1; remove or disable the old transport in the same change so rules still arrive once |

An implementation may carry dormant/shared hook machinery for a later phase,
but README support claims and default hook registration advance only with the
named evidence. Unsupported surfaces retain the installed-file bootstrap where
their instruction system can read it; Trellis does not call that equivalent to
a verified native transport.

**6. Refresh and preset application are different acts.** Ordinary setup
refresh remains a payload update and validator. It never overwrites an existing
consumer-owned `.trellis/rules.toml`. A new explicit “apply preset” operation
may, after showing the proposed replacement and receiving confirmation, replace
the rows, strictness, and `seeded_from` provenance with the conductor or
author-adapt seed. Manual edits after that operation remain authoritative and
live at the next host context boundary. Parked `seed` and `custom` presets do
not return.

Preset application is a configuration reset, not a prerequisite for Phase 1
transport. It may ship as the next bounded slice rather than enlarging the
first hook PR.

**7. Failure is visible without being needlessly fatal.** A Codex hook that
runs but cannot assemble a valid installed overlay emits a visible diagnostic
and no rules payload; the bootstrap then attempts the same installed files.
All four hook inputs are required. The generated prose files must be nonempty,
the version must have the installed `payload@…` shape, and `rules.toml` must
pass setup's existing strictness, complete-known-row, and floor validation.
Missing hook execution is not itself a failure. If neither transport nor
fallback can deliver valid rules, the agent must tell the user that Trellis
was not loaded and must not claim governed execution.

**8. Earlier contracts narrow at exact points.**

- `decision-0051`'s seed-once/consumer-authority rule stands for setup and
  refresh. Its “no set-posture skill yet” deferral is superseded only enough to
  permit the explicit, confirmed preset-reset operation in point 6.
- `decision-0053`'s live-row result stands. Its “immediately” wording is
  clarified to mean the next host context-loading boundary, and its two listed
  transports gain the Codex hook plus bootstrap transport defined here.
- `decision-0057` and `spec-0006` stand for shared prose and Claude's retained
  import block. Their prohibition on any Trellis-managed `AGENTS.md` block is
  superseded only for the small Codex receipt/fallback in point 4; full Trellis
  rule text still never lives there.
- `decision-0039`'s Claude staleness surface and `decision-0043`'s vendored
  overlay/stamp mechanics stand.

With this approval, decisions 0051, 0053, and 0057 gain
`superseded_in_part_by: [decision-0058]` plus append-only forward-pointer notes
stating these exact boundaries. `spec-0006` gains the matching current-truth
pointer; its already satisfied historical acceptance result is not rewritten.

## Consequences

- A consumer's Claude and local Codex sessions receive the same installed
  payload and current rows through different native transports, with a
  single-copy invariant.
- Row edits stay cheap: edit `rules.toml`, then cross the host's context-loading
  boundary. Refresh is for generated payload updates; explicit preset
  application is for intentional wholesale row replacement.
- `.trellis/internal/` remains useful rather than vestigial: it is the shared
  inspectable payload, fallback source, staleness target, and uninstall unit.
- The plugin package gains Codex metadata, hook assembly, tests, setup/remove
  propagation, README support boundaries, and checksum/version regeneration.
- Phase 1 does not claim resume, compact, subagent, IDE, desktop, automation, or
  cloud correctness. Those are named next experiments, not fine print.
- The Claude import may eventually be replaced by a verified Claude hook, but
  never while both paths would inject the full rule payload.

## Open questions

- Does each Codex surface expose the same plugin root, project root, hook
  schema, trust boundary, and context-size behavior as the local CLI?
- On subagent start, does Codex inherit parent developer context, receive the
  subagent hook context, or both? Phase 2 must measure this before enabling the
  hook by default.
- Should the explicit preset operation be a dedicated `set-preset` skill or an
  explicit mode of setup? Its consent and overwrite semantics are fixed here;
  its command surface belongs to that slice's spec.

## Self-check

This record preserves the proven Claude path and bounds the Codex claim to the
one maintainer-reported live positive control. The fallback degrades transport
strength honestly rather than presenting
best-effort instruction as deterministic enforcement. One installed payload
remains authoritative, one host receives one effective copy, and row edits do
not acquire a refresh dependency. The broad rollout is split into checkable
promotion gates, while Phase 1 remains small enough for an independent
implementation and review.
