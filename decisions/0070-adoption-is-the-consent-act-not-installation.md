---
id: decision-0070
type: decision
status: gated  # drafted by agent; awaiting the maintainer's intent act. Design supplied by the maintainer in session 2026-07-31 ("bring the user the choice"); this record works it through and flags the two places it had to change to survive contact with the code.
depends_on: [decision-0008, decision-0053, decision-0065, decision-0068]
owner: agent
date: 2026-07-31
---

# 0070 — adoption is the consent act, and every path has one

## Context

**A fresh install governs almost nothing, and the two paths fail differently.**
Measured on `main`:

| state | rules in context | rules that apply |
|---|---|---|
| curl install, no `rules.toml` | all 14 (~5.4KB) | **2** — the floors |
| plugin install, no `rules.toml` | **0 bytes** | 0 |
| either, after `/trellis:setup` | 14 | 14 |

Both shipped presets are **14/14 `active = true`**, so "everything on" is already
the intended steady state. The gap is only the window between installing and
configuring — but the curl path spends the full context cost inside that window
and delivers two rules for it.

**The obvious fix is ruled out, and by this project's own records.** Making
absence of `rules.toml` mean "all 14 active" is symmetric and needs no writer. It
is also the thing `decision-0065` forbids in terms: *"`.trellis/rules.toml` is
the opt-in signal. The plugin may be installed user-wide; a project that never
adopted Trellis is never governed by surprise."* `decision-0068` D1 ruled the
same hazard one day earlier for `~/.claude/rules/`, and `install.sh:364` records
why: *"would govern every repo on the machine."* `decision-0008` makes it a
floor: *"The non-negotiable is **surfacing**, not enforcing … a conscious,
visible choice, **never silent**."*

**This record does not restore that hazard by the back door.** D4 governs a
user-scope project only after telling the user, in that project, that it is about
to — which is what `decision-0008` asks for (*"a conscious, visible choice, never
silent"*) and what `decision-0065`'s clause, written before any announcement
existed, could not distinguish from silence. **`decision-0065`'s "a project that
never adopted Trellis is never governed by surprise" is superseded in part**: such
a project may now be governed, but never *by surprise* — the announcement is the
difference, and it is load-bearing rather than decorative.

**And the suppression the naive fix would remove is measured, not theoretical.**
`eval/experiments/annotation-vs-absence`, 60 runs at `2ec7da8`: with an active
row the rule fires **19/20 (95%)**; with the rule text present and its row
inactive, **0/20**. Absence-means-on would convert a measured 0% into ~95% in
every repository on the machine — including `floor-intent-gate`, which is a
refusal instruction, not advice.

## Decision

**1. The unit of consent is ADOPTION, not installation — except at user scope,
where installation IS the adoption and the product says so out loud.** Installing
Trellis project-locally says "govern this project". Installing it user-wide says
"govern my work", and D4 makes that explicit per project rather than assumed.
Every delivery path has an adoption act:

| path | the adoption act | already in the repo? |
|---|---|---|
| curl | running `install.sh` **in that repo**, which writes into it | yes |
| plugin, **project** scope | the bundle vendored at `<repo>/.claude/skills/trellis/` | yes |
| plugin, **user** scope | the user-wide install itself, **announced per project** (D4) | the announcement, not a file |

**2. `install.sh` seeds `.trellis/rules.toml` from `reference/rules-b.toml` when
none exists.** Running an installer inside a repository is an unambiguous
adoption act, and the script already writes two other things there. This makes
the curl path deliver 14/14 at posture B immediately, and resolves the
`@../../.trellis/rules.toml` import that `decision-0068` measured as contributing
"nothing, silently, no error".

**This amends `decision-0065`'s "setup writes exactly one file … ever"**, which
`decision-0068` D9 declined to widen. It is amended openly here rather than
routed around. The clause's purpose — that no path silently vendors an overlay —
is untouched: one config file, in the project the user ran the installer in.

**3. A project-scoped plugin install is itself adoption; absent rows then mean
all-14, posture B.** The bundle lives in the repository, so it is visible,
greppable and deletable — the properties `decision-0008`'s surfacing floor asks
for. **The hook writes nothing to achieve this.** It reads the absence and
applies the default.

*(An earlier shape of this decision had the hook WRITE `rules.toml` here. Dropped:
a `SessionStart` hook that mutates the repository before the user has typed
anything is a new and surprising behaviour, and it is unnecessary — the default
can simply be applied.)*

**4. A user-scoped plugin in an unadopted project ANNOUNCES, once, and offers to
stop.** This is an opt-OUT, chosen deliberately over the opt-in an earlier draft
carried. A user-scoped install genuinely *is* a broad choice, and a tool that
pretends otherwise — then asks permission it already assumed — is the less honest
design. The maintainer's framing, 2026-07-31:

> "The plugin is installed user scope so this repo will be governed by trellis.
> Do you want to disable this?"

The hook injects that announcement naming the project, **and injects no rules on
that turn** — "will be", not "is". The agent puts it to the human and writes the
answer:

- **decline** → `.trellis/rules.toml` containing `governed = false` and nothing
  else. The hook honours it: none of the twelve configurable rules apply, and it
  says so once rather than vanishing.
- **accept, or no objection** → seed `.trellis/rules.toml` from `rules-b.toml`.
  From the next turn the project is governed at 14/14, posture B.

The prompt states the consequence before asking, and asks for the **negative**
action explicitly, so silence never reads as refusal — a user who ignores it gets
what the announcement said would happen, which is the only reading under which
"will be governed" is a true statement rather than a threat.

**The hook never writes.** It has one channel, injected context, and keeps it.
Writes stay agent-mediated and human-consented, which is what makes this comply
with `decision-0008` rather than merely resemble compliance.

**5. `governed = false` is a new top-level key, and it stops the twelve — not the
two floors.** All-false rows cannot express a project-level opt-out at all:
`reference/rules.md:1` says *"the two `floor-` rows apply regardless of their row
value"*, so every row set to false is indistinguishable from a project that
simply turned everything off. Hence its own key, checked before any row.

**The floors survive it, deliberately.** They are *"the only settings that never
dial to zero"* (`README.md:160`), and `decision-0008` makes the reason ratified:
*"the non-negotiable is **surfacing**, not enforcing."* A project may decline
Trellis's opinions about how work is done; it does not thereby acquire an agent
that hides consequential choices or ships without approval. A project-level
opt-out is still a row-level mechanism and does not reach below the floor.

*(An earlier draft of this clause had `governed = false` silence everything,
floors included, and argued that was "the whole point". That was wrong on the
product's own terms — it let a config file switch off transparency and the intent
gate. Corrected on the maintainer's 2026-07-31 ruling.)*

**6. Scope is detected by containment, and ambiguity resolves to ASKING.** The
hook tests whether `CLAUDE_PLUGIN_ROOT` resolves inside `CLAUDE_PROJECT_DIR`
(user-scope example, measured: `~/.claude/plugins/cache/kodhama/trellis/0.2.0`;
project-scope: `<repo>/.claude/skills/trellis/`). Symlinks and marketplace caches
make this imperfect. **When it cannot tell, it treats the project as unadopted
and asks** — the failure mode is one extra question, never governing by surprise.

## Consequences

- `decision-0065` gains a `superseded_in_part_by` pointer for **two** clauses:
  "setup writes exactly one file … ever" (D2), and "a project that never adopted
  Trellis is never governed by surprise" (D4), the second narrowed to preserve
  *by surprise* while releasing *never governed*.
- `/trellis:setup` stops being the thing that turns rules **on** and becomes the
  thing that changes **posture** and rows. Its name still fits on the user-scope
  path, where the file genuinely may not exist.
- The landing page's `# then run /trellis:setup` stops being load-bearing for
  the curl path; a first-time-user review found it demoted to a grey comment and
  excluded from the copy button.

## Open questions

1. ~~**Where the answer persists, and the cost of an unexpected commit.**~~
   **RULED 2026-07-31 — repo-local, and the commit is the point.** The maintainer:
   *"the user made the choice when he said okay, install this globally … it does
   imply an additional commit that wasn't expected, but I think that's a low cost
   to pay."*

   The reasoning that settles it, and which a `~/.claude/` declined-list would
   have destroyed: **`governed = false` is a project fact, not a personal one.**
   It lands in the diff, so the people working on that project review it like any
   other change. If they agree the project is ungoverned, it is — by their
   decision, recorded. If they disagree, they change the file and say so
   explicitly. A machine-local list would let one developer's global install
   silently decide governance for a repository their colleagues share, with no
   artifact anyone else can see or contest.

   So the "unexpected commit" is not a cost the design tolerates; it is the
   mechanism by which a personal install stops being a personal decision. What
   the record cannot do is control what a developer installs on their own
   machine — it can only make the consequence visible to the project, which is
   `decision-0008`'s floor applied one layer out.
2. **The eval never tested "no rows at all".** `annotation-vs-absence` covers
   `active = false`, not a missing rows file. D3's default rests on a suppression
   mechanism proven for one state and assumed for its neighbour. An arm should be
   added before D3 is relied on.
