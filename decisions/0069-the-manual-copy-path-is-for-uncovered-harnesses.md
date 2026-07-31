---
id: decision-0069
type: decision
status: gated  # drafted by agent; awaiting the maintainer's intent act. Scope ruling requested in session 2026-07-31 after the LP audit found the README instructing users into a state the product nudges them out of.
depends_on: [decision-0043, decision-0053, decision-0065, decision-0068]
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
2. **The hook tells you to leave it.** `hooks/staleness.sh:141` emits: *"This
   project still carries a vendored overlay, **which the plugin no longer writes
   or refreshes**. To move it onto plugin-delivered rules, run `/trellis:setup`
   and accept the migration — it removes `.trellis/internal/` and the managed
   block."*
3. **`install.sh` refuses to render over it.** `decision-0068` D12: a
   `.trellis/internal/` tree or a managed block suppresses the rendered rules
   file, because both static chains load before any hook runs.

A consumer who follows the README on Claude Code is therefore told, at their next
session, to undo what the README just had them do — and has silently opted out of
the delivery path `decision-0068` built.

**What is NOT wrong here, and needs saying because the obvious reading is that
the manual path is stale debt.** It is not. `kodhama-0007`'s "one render, many
copiers" makes a hand-copied payload a first-class consumer, and for a harness
with neither plugin support nor `.claude/rules/` the manual path is the **only**
delivery mechanism that exists. Deleting it would remove the product's
harness-agnostic story, not tidy it.

Nor does path A deliver anything. `staleness.sh:123-145` **checks** the overlay
and exits without injecting; on the manual path the rules reach the session
through the managed block's own imports. So the overlay shape is load-bearing
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

**4. The hook's migration nudge is left exactly as it is.** It is correct for the
population that sees it: the hook runs only on Claude and Codex, which are
precisely the hosts where D2 says the manual path is superseded. A consumer on an
uncovered harness never runs it.

## Consequences

- `README.md`'s manual-copy section gains a scope sentence and an explicit
  exclusivity note. The recipe itself is unchanged — it still works.
- No code changes. This record exists so the next cleanup does not read
  `.trellis/internal/` support as dead weight and delete it; three mechanisms
  currently imply it is dead and none of them says it is retained.

## Open questions

1. **Does an uncovered harness get any staleness surface at all?** `decision-0035`
   makes staleness a floor, and a hand-copied overlay on a harness with no hook
   has none — the consumer has no way to learn their payload is behind. Not
   resolved here; naming it so it is not mistaken for covered.
