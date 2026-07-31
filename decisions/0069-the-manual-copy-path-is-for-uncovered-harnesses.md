---
id: decision-0069
type: decision
status: gated  # drafted by agent; awaiting the maintainer's intent act. Scope ruling requested in session 2026-07-31 after the LP audit found the README instructing users into a state the product nudges them out of.
depends_on: [decision-0035, decision-0043, decision-0065, decision-0068, kodhama/kodhama-0007-one-render-many-copiers]  # 0053 was declared in an earlier draft and never argued in the body — the defect decision-0068's self-check row 7 records being caught on — so it is dropped rather than left decorative. decision-0035, decision-0043 (Open question 1) and kodhama-0007 ARE argued below; the first and last were missing.
owner: agent
date: 2026-07-31
---

# 0069 — the manual copy path is for harnesses the plugin does not cover, and says so

## Context

`README.md:107-140` tells a human to build a vendored overlay by hand:
`.trellis/internal/{trellis,rules,invariants,version}.md` plus
`cat "$ref"/block-claude.md >> CLAUDE.md`.

**Three mechanisms now disagree about that state, and a consumer meets all
three.** Measured on `main` at `6a920cb`, not inferred:

1. **The README creates it.** The copy recipe above is the documented path for
   "any other harness".
2. **The hook tells you to leave it** — but only once the payload has moved on.
   `hooks/staleness.sh:142` emits: *"This project still carries a vendored
   overlay, **which the plugin no longer writes or refreshes**. To move it onto
   plugin-delivered rules, run `/trellis:setup` and accept the migration — it
   removes `.trellis/internal/` and the managed block **and keeps your
   `.trellis/rules.toml` rows**."* The final clause matters and an earlier draft
   truncated it away: the migration is not destructive of the consumer's rows.
3. **`install.sh` refuses to render over it.** `decision-0068` D12: a
   `.trellis/internal/` tree or a managed block suppresses the rendered rules
   file, because both static chains load before any hook runs.

A consumer who follows the README on Claude Code has silently opted out of the
delivery path `decision-0068` built, and will eventually be told to undo what the
README just had them do.

**"Eventually", not "at their next session" — an earlier draft of this record said
the latter and it is measurably false.** A fresh manual copy takes
`reference/version` from the same clone, so `overlay == current` and path A emits
**nothing**; verified by running the hook against a freshly built overlay and
getting empty output. The nudge appears only after the installed payload moves
ahead of the copy. The conflict is real and deferred, not immediate, and the
record should not borrow urgency it does not have.

**What is NOT wrong here, and needs saying because the obvious reading is that
the manual path is stale debt.** It is not. `kodhama-0007`'s "one render, many
copiers" makes a hand-copied payload a first-class consumer, and for a harness
with neither plugin support nor `.claude/rules/` the manual path is the **only**
delivery mechanism that exists. Deleting it would remove the product's
harness-agnostic story, not tidy it.

Nor does path A deliver the rules. `staleness.sh:123-145` **checks** the overlay
and injects **no rule payload** — though it is not silent: three `emit` calls on
that path report a missing stamp, an incomplete overlay, or a stale one. (An
earlier draft said it "exits without injecting", which reads as emitting nothing
at all and is wrong.) On the manual path the rules reach the session through the
managed block's own imports, not through the hook. So the overlay shape is load-bearing
for exactly the consumers the plugin cannot reach, and inert for everyone else.

## Decision

**1. The manual copy path is retained, and is scoped to harnesses the plugin does
not cover.** It is not legacy, not deprecated, and `staleness.sh`'s path A is not
cleanup debt — a future prune that deletes either must supersede this record.

**2. On Claude Code and Codex CLI the manual path is superseded, and the README
must say so at the point of instruction.** Those two hosts are covered by the
plugin (`decision-0065`), and Claude additionally by the rendered rules file
(`decision-0068`). A reader on either host who follows the copy recipe lands in
the state mechanisms 2 and 3 above are built to discourage.

**3. The curl path and the manual path are mutually exclusive, stated rather than
discovered.** Running `install.sh` in a repo that already carries a hand-built
overlay refuses to render and says why; that is correct behaviour and is now
documented as a property of the pair rather than as a surprise.

**4. The hook's migration nudge is left exactly as it is**, with one case named
rather than papered over. It is correct for the population that mostly sees it:
the hook runs only on Claude and Codex, which are precisely the hosts where D2
says the manual path is superseded.

**The exception, which "a consumer on an uncovered harness never runs it" would
have hidden:** a project may run Claude Code *and* an uncovered harness off one
hand-built overlay — that multi-copier case is exactly what `kodhama-0007`
contemplates, so it cannot be waved away by the same record that leans on it.
There, Claude sees the nudge, and accepting the migration deletes the overlay,
leaving the uncovered harness with **no delivery at all**. This is inference from
the code, not an observed report. It is not resolved here; it is Open question 2
so that the next change to the migration has to see it.

## Consequences

- `README.md`'s manual-copy section gains a scope sentence and an explicit
  exclusivity note. The recipe itself is unchanged — it still works.
- No code changes. This record exists so the next cleanup does not read
  `.trellis/internal/` support as dead weight and delete it; three mechanisms
  currently imply it is dead and none of them says it is retained.

## Open questions

1. **Does an uncovered harness get any staleness surface at all?** `decision-0035`
   makes staleness a floor, and a hand-copied overlay on a harness with no hook
   has none — the consumer has no way to learn their payload is behind.

   **`decision-0043:106-108` asserts the answer already exists**: the manual
   path's staleness check is "the same file compare run by hand … documented on
   the manual path". Verified against this branch: **it is not documented** —
   `README.md` contains no `cmp` and no compare procedure of any kind, and 0043
   names `.trellis/version`, a path superseded by `decision-0051`. So the floor is
   claimed as met by a mechanism that was never written down. Not resolved here.

2. **What happens to a project running one hand-built overlay across a covered
   and an uncovered harness?** See D4. Accepting the migration on the covered host
   removes the uncovered host's only delivery. Unmeasured, and no consumer report
   exists; recorded so it is not discovered by a consumer first.
