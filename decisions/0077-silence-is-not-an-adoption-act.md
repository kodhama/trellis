---
id: decision-0077
type: decision
status: approved  # maintainer's intent act 2026-08-28, in session, on this text: "Approved… the damn gating again." An in-PR flip recording that act (decision-0046, decision-0022); author (agent) != approver (maintainer). Two earlier acts the same session settled the forks this record turns on — the direction ("The record — correct 0070 D4", over fixing the hook) and the mechanism ("Superseding-in-part decision", over an appended correction block). The intervening `gated` landing was the agent's over-caution, not the rule: decision-0074 set the precedent for flipping on an in-conversation act with the details accepted at merge. Recorded because the round-trip it cost is friction, and friction here is product research (AGENTS.md)
depends_on: [decision-0070, decision-0073]
informed_by: [decision-0008, decision-0040, decision-0046, spec-0001]
owner: agent
date: 2026-08-28
provenance: filed as TRL-8 against `decision-0073` D3. The finding is older — the global cross-cutting review of run `20260803-093000-delivery-shapes-closed-set` recorded it as L3 ("routes to the MAINTAINER at ship") and it rode the ledger without being closed. That ledger lived in `.grove/`, deleted by `decision-0076` on grove's retirement, so the run citation is historical and no longer resolves in-tree — the verdict text is quoted here rather than pointed at. Both the run record and the issue indicted `decision-0073`; the session that discharged them found the claim originates one record upstream, in `decision-0070` D4, which `0073` restated faithfully
---

# 0077 — silence is not an adoption act

## Context

**`decision-0070` D4 describes a hook that has never existed.** Its second
bullet reads:

> **accept, or no objection** → seed `.trellis/rules.toml` from `rules-b.toml`.
> From the next turn the project is governed at 14/14, posture B.

and its rationale that *"silence never reads as refusal — a user
who ignores it gets what the announcement said would happen"*.

The shipped hook names two actions and neither of them is silence. The
`TRELLIS_NOT_YET_GOVERNING` announcement — path B's user-scope branch, the
`else` arm of the project-scope containment test — instructs the
agent to write `governed = false` if the user declines, to copy the preset
*"If they ACCEPT"*, and then says the opposite of D4 outright:

> without that file this same announcement repeats every session and **the
> project is never governed**

It then `exit 0`s having injected nothing. The message has read that way since
**#218 — the commit that built `decision-0070` itself** — which
`git log -S'this same announcement repeats every session' -- plugins/trellis/hooks/staleness.sh`
returns as the sole commit. (A line-addressed `git log -L` was the original
recipe; it is recorded in this form because line numbers move, and this record's
own change moves that one.) The silence-seeds branch was never implemented; it
was not a regression, and no later record narrowed it (0071–0075 and every spec
checked).

**Measured, not recalled** (`cli/plugin_hook_test.go`,
`TestSilenceNeverAdoptsAfterTheDeclineIsDeleted`): a user-scope plugin, a
project that records `governed = false`, that file then deleted — the
announcement re-arms, and a second session with nothing written announces
*again* and injects zero slugs. Under mutation (the announcement branch falling
through to the preset instead of exiting) the guard goes red with all fifteen.

**The error propagated downstream because it was restated faithfully.**
`decision-0073` D3 says *"one ignored prompt re-governs at 14/14"*, and
`skills/remove/SKILL.md` carries the same sentence to users. Neither is a
drafting slip: both are correct readings of approved upstream. The run record
and TRL-8 both indicted `0073`, and correcting `0073` alone would have put it
out of conformance with a record it names in `depends_on` — the enumeration
defect `0073` exists to end, one layer up.

**Which side to correct was a real fork, and the code won on `decision-0070`'s
own argument.** That record is titled *"adoption is the consent act, not
installation"*, and its D3 makes the vendored bundle the adoption **act**.
D4's second bullet is the one clause in it where adoption happens with nobody
doing anything — and silence is not an act. Correcting the hook instead would
have been a shipped-consent change that newly governs repositories whose users
ignored a prompt; correcting the record changes no behaviour and un-governs
nobody. `decision-0008` and `floor-intent-gate` point the same way.

## Decision

**1. Silence never adopts.** Under a user-scope install in a project with no
`.trellis/rules.toml`, the project is governed only once an explicit act writes
that file. An unanswered announcement leaves the project **ungoverned**, and
recurs next session. `decision-0070` D4's *"accept, or no objection → seed"*
bullet and its *"silence never reads as refusal"* rationale are **superseded**.

**2. What stands in D4, untouched.** The announcement itself; that it names the
project; that it **injects no rules on that turn** (*"will be"*, not *"is"*);
the decline bullet writing `governed = false`; and **"the hook never writes"** —
writes stay agent-mediated and human-consented. Only the disposition of
*silence* changes.

**3. The mechanism's honest name is announce-then-accept.** D4 called itself
*"an opt-OUT, chosen deliberately over the opt-in an earlier draft carried"*
(D4's opening). That characterisation does not survive point 1: a project that
never answers is never governed, which is opt-in per project. What remains true
is the reasoning underneath it — a user-scope install *is* a broad choice
already made, so the honest surface announces it rather than asking permission
it already assumed. The announcement is the disclosure; the file is the consent.

**4. `decision-0073` D3's second bullet inherits this correction.** Its
surviving half is *"deleting the recorded decline re-arms the adoption
announcement"* — verified true, because the branch is a bare file-existence test
with no persisted already-announced state. Its *"one ignored prompt re-governs"*
half goes with D4. The consequence the remove skill must state is that deleting
a decline returns the project to the **unadopted** state: the announcement
returns, and the project is governed only if someone then accepts.

**5. The claim is pinned by behaviour, not prose.** A `cli/` fixture runs the
hook twice over an unanswered announcement and asserts zero slugs injected, the
announcement repeating, and that the hook wrote no file — the assertion class
`decision-0073` D4 reserves for shell components. Prose surfaces that restate
the consequence to a user are textual pins, as before.

## Consequences

- **`decision-0070` gains `superseded_in_part_by: [decision-0077]`** and an
  append-only forward pointer. `spec-0001` §2 and `decision-0040` point 5 class
  that marking as *"a marking, not an edit-in-substance"*; D1–D3 and D5–D7 are
  untouched, and D4 keeps everything point 2 above names.
- **`decision-0073` gains the same marking**, scoped to D3's second bullet. The
  rest of that record — the closed set, the transaction ordering, the morph
  preflight, the predicate discipline — is unaffected.
- **`skills/remove/SKILL.md` states the corrected consequence**, and
  `cli/remove_skill_test.go`'s needles move with it. The skill is a
  revise-in-place product surface, so it carries the current truth rather than a
  pointer — the same split `decision-0074` already established there when the
  rule count moved to fifteen while `0073` stayed frozen at fourteen.
- **`staleness.sh`'s D4 rationale comment is corrected.** It currently sits
  directly above the announcement saying *"silence cannot read as refusal"* —
  describing D4 rather than the code beneath it. Its bytes change, so
  `install.sh`'s bundle manifest hash for `hooks/staleness.sh` is recomputed in
  the same change (`decision-0028`).
- **No product behaviour changes, and no repository becomes governed or
  ungoverned by this record.** That is the point of the direction chosen.
- **Codex is unaffected, and the gap narrows.** `decision-0070` D7 already
  bounds D3/D4 to the Claude path; `codex-context.mjs` has no announcement
  branch and has always required an explicit `.trellis/rules.toml`. Correcting
  the record toward explicit-act adoption moves the two hosts closer together,
  not further apart.
- **`decision-0015`'s open question is deliberately not settled here.** Whether
  an append-only block that leaves the original text untouched is an "edit" in
  the sense `AGENTS.md` forbids remains open. This record takes the
  supersede-in-part route — the mechanism ratified in writing — precisely so
  that it does not settle that question by accident, and adds no third contested
  instance to it.

## Self-check

Sections present and per-type correct (`Context` / `Decision` / `Consequences`).
Every load-bearing code claim is either quoted from source with a line cite or
measured this session by a test that goes red under mutation.

**No claim in this record cites a line number in a file this change edits.**
That is a repair, not a style preference: three separate review rounds — a
corpus review and two automated passes — each found stale line citations here,
because every edit this change made shifted the lines its own prose pointed at.
The citations now anchor on content that moves with the text it names (a quoted
phrase, a decision point, a named branch, `git log -S` over a distinctive
string) rather than on a coordinate the next edit invalidates. Line numbers into
*unedited* files remain fine; the defect is self-reference across an edit.

`depends_on` names
both records this supersedes in part, and both receive their forward marking in
this change, so no reader lands on an outgrown half without a link. The
direction, the mechanism and the approval were all the maintainer's intent acts,
recorded in the frontmatter with the wording they used; the drafting is the
agent's and is not claimed as theirs. This record landed `gated` first and was
flipped on the maintainer's act minutes later — an avoidable round-trip, since
`decision-0074` had already established that an in-conversation act carries a
flip with the details accepted at merge. Noted here rather than smoothed over,
because a gate that fires when no one asked for it is a cost the methodology
pays without buying anything.
