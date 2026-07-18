---
id: decision-0048
type: decision
status: approved  # maintainer's intent act 2026-07-18 ("ratify decision-0048 and merge #159", in-session) — in-PR flip recording the act (decision-0046); independent conformance check (corpus-reviewer) passed before the gate
depends_on: [invariants-v1, signature-catalog-v1]
informed_by: [spec-0005, decision-0043]
owner: agent
date: 2026-07-18
---

# 0048 — setup performs no git and injects no landing workflow into the consumer session

## Context

`/trellis:setup` is an **inline** skill: when a consumer runs `/trellis:setup`, its prose executes
in the consumer's *own* conversation, in the consumer's *own* repo. That inlining has a consequence
that only surfaced once we tried to give setup a landing behaviour: **any git-workflow instruction
in setup's body leaks into the consumer session and biases how that session handles git for its
own, unrelated work.**

The chain that exposed it:

- The original friction (math-quest re-install, 2026-07-17; reported by the maintainer): setup was
  **silent** on landing, so the host agent freelanced its own default — "commit to `main` and
  push."
- The first fix attempt added a landing step that **recommended opening a PR** and printed the git
  commands. But because setup is inline, that recommendation is itself an **injected opinion**: it
  imports trellis's preferred git workflow into a consumer project that has its own conventions.
  This was observed live (2026-07-18) — a session that had just been *writing* that step-8 prose
  then applied the same "present the PR command, you run it" hand-off to an unrelated git operation,
  a default it would not otherwise have offered. The instruction leaked from the skill into the
  session.

So an inline skill faces two failure modes pulling opposite ways: **silence** lets the host
freelance a bad landing; **a workflow recommendation** contaminates the host's git defaults. Neither
is acceptable.

## Decision

**Setup performs no git of its own, and injects no landing workflow into the session it runs in.**

- Setup does **no** `add`/`commit`, no branch, no push, no PR. It writes the overlay, verifies it,
  and reports.
- Setup **recommends no landing approach** — not a PR, not a branch, not committing anywhere. How
  the overlay is landed is **the consumer project's decision, by the consumer project's own
  conventions.**
- To avoid the freelance failure mode without injecting an opinion, setup **surfaces the state and
  defers**: it names that the overlay is written and uncommitted, and lets the user decide how to
  land it, following whatever that repo normally does. It never lands the change unasked, and never
  commits onto the current branch.

The line this draws: a **neutral, procedural hand-back** ("I did no git; here is what is
uncommitted; landing is yours, your way") is not an injected opinion — it points at the consumer's
own defaults. A **workflow recommendation** ("open a PR, here is the command") is the injected
opinion, and is what setup must not carry.

**This is a chosen line about what setup may *say*, not a claim the hand-back is perfectly
neutral.** "Surface and let the user decide" is itself a thin procedural norm — it ensures the human
is asked. It is legitimate because it governs the landing of setup's *own* overlay and defers the
*how* entirely to the consumer, importing no git workflow. What it deliberately does **not** do is
steer the consumer toward PR-over-`main`; that steering would be the very opinion being removed. The
trade is explicit: honour "nothing lands unasked" (`floor-intent-gate`) without imposing a workflow.

## Why not run setup cold instead

Isolating setup in a cold sub-agent (`context: fork`) was considered as a structural way to keep its
prose out of the consumer session. **Rejected for M1.** It conflicts with grove's decided model —
interactive/stateful work stays with the driving session, not a one-shot sub-agent (grove ADR-0030,
`.claude/agents/README.md`) — and with setup's interactive contract (posture, content-guard); a
forked sub-agent cannot even prompt the user. And once the *opinion* is removed there is nothing
left to isolate: `inv-minimal-first` favours removing the bias over building an isolation mechanism
to contain it. (Isolation stays the open question for **M2**, whose contamination is a generative
rewrite, not deletable prose — see below.)

## Consequences

- `plugins/trellis/skills/setup/SKILL.md` step 8 becomes a neutral hand-back: setup performs no git,
  recommends no workflow, surfaces the uncommitted state, and defers landing to the consumer. It
  replaces the earlier "ask how to land, PR recommended" step. The `reference/checksums` **overlay**
  manifest is unchanged (`SKILL.md` is not one of its payload files), but the edit **advances
  `install.sh`'s bundle-wide manifest** (`TRELLIS_BUNDLE_MANIFEST` covers the whole `plugins/trellis/`
  tree, `SKILL.md` included) — regenerated in the same change (`decision-0028`), as
  `TestInstallScriptBundleManifestIsCurrent` guards.
- Setup gains **no** git capability (an earlier draft would have had it perform a local
  branch+commit; that is dropped). Its outward line matches the curl `install.sh` (no git mutation;
  landing is the human's) but for a sharper reason: not just "the human owns the merge," but "setup
  owns no git workflow at all."
- No *prose* downstream described the M1 landing step, so no doc needs re-syncing (`decision-0028`).
  The one derived surface the skill edit does touch is `install.sh`'s bundle manifest (above).

## Open questions

- **M2 isolation (parked, not closed).** The morph — `/trellis:setup`'s model-driven rewrite of the
  consumer's own files — carries contamination this decision does **not** address: its bias is a
  generative step carrying ambient context, not an opinion that can be deleted. Isolating M2
  (cold / `context: fork`) remains a live follow-up; deferred here as "probably not needed for the
  immediate concern" (maintainer, 2026-07-18), recorded so it is not silently dropped.
- **Does the neutral hand-back still nudge too much?** "Surface and ask" imposes a thin
  confirm-before-landing norm on the consumer. Judged acceptable (it governs setup's own artifact
  and defers the *how*), but the stricter reading — hand back *silently* and let the host's own
  norms decide — was not taken, on the grounds that silence is exactly what let the host freelance
  in the first place.

## Self-check (gate)

Grounded in the observed mechanism (an inline skill's prose leaking into the consumer session, seen
live 2026-07-18) and the ratified invariants it rests on: `floor-intent-gate` (nothing lands
unasked) and `inv-minimal-first` (remove the bias rather than build isolation to contain it) — both
`signature-catalog-v1`. Framed honestly as a **chosen line about injected opinion, not a claim of
perfect neutrality**: the residual "surface and ask" norm and the un-taken stricter reading are
recorded as open questions, not suppressed, and the M2 gap is parked visibly rather than closed by
silence. `depends_on` carries only genuine coupling (the invariants the behaviour implements);
`spec-0005` and `decision-0043` are `informed_by` (precedent and host context, not
correctness-contingent) per `decision-0047`. Depends only on ratified artifacts; no draft consumed.
Left at `draft`: the author does not grade its own decision — gating and the `approved` intent act
are the maintainer's (`decision-0046`), recorded by the in-PR flip, ideally after an independent
pass (`inv-independent-judgment`).
