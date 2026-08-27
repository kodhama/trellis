---
id: decision-0076
type: decision
status: gated  # maintainer's intent act 2026-08-27, in-session ("Deliberately retired") on being shown the plugin is absent — that settles the retirement itself. Author (agent) != approver (maintainer). Held at `gated` rather than flipped to `approved`: the act covered the *fact* of retirement, not this record's consequence set — the `spec-0006` AC1 partial supersession, the `.grove/` deletion, and the parked owed-set below are the maintainer's to accept at merge (`decision-0046`, `decision-0022`)
depends_on: [spec-0001, spec-0006, decision-0044, decision-0057]  # spec-0001 and decision-0044 moved in / up after an independent corpus review: Decision points 2 and 5 rest on spec-0001's registry and its versioning delegation, and point 2's correctness is contingent on decision-0044's qualified <repo>/<id> form standing — that is coupling under decision-0047's test, not provenance
informed_by: [decision-0005, decision-0045, decision-0051, decision-0071, decision-0074]
owner: agent
date: 2026-08-27
---

> **Provenance.** Surfaced while scoping `TRL-22` (adopt the superpowers plugin), which asked how
> three layers of process instruction compose. Checking the premise found only two: the grove plugin
> is not installed. The maintainer confirmed in the same session that the retirement is deliberate.
> This record exists because that retirement had never been written down — and neither had the
> adoption.

# 0076 — Grove is retired; its citations are not

## Context

- **Grove was never adopted by a decision record.** It arrived as `fda44d9` (2026-07-08),
  *"chore: install grove as operating model (eleven roles, corpus-reviewer instance)"*. No decision
  record in `decisions/` adopts, ratifies, or retires it — no decision title names grove at all. An
  operating model for the whole repo entered as a chore commit and left the same way. That is the
  finding, and it is about this repo's discipline, not about grove.

- **The plugin is absent, and the repo still instructs agents to use it.** No `grove:*` subagent type
  loads in a session here, and `~/.claude/plugins/cache/kodhama/` holds only `kodhama` and `trellis`.
  Until this change, the managed block in `AGENTS.md` told every agent that work matching a grove
  workflow *"run[s] as grove runs, sequenced through grove's chartered agent roles, loaded from the
  grove plugin as `grove:<role>` subagents (all thirteen)."* The instruction had been unfollowable
  since the plugin left; the last recorded run is `20260803-093000`.

- **"Grove" names two different things, and only one of them is installable.** The conflation is why
  retiring it looked like a far bigger job than it is:
  - **grove-the-plugin** — `grove:<role>` subagents, `/grove:refresh`, `/grove:setup`, the
    plugin-carried companions, the `grove plugin@0.1.0` stamp. *This is what is gone.*
  - **grove-the-repo** — cited across the corpus as artifacts, under `decision-0044`'s qualified
    `<repo>/<id>` form, and still a member of the recognized-repo registry (`spec-0001:97`).
    *This is untouched, and cannot be touched.*

  *An earlier draft justified the second half on the repo still being live — public, last pushed
  2026-08-05. That is true and it is irrelevant: `spec-0001:100–103` sets resolution depth at
  **"shape + registry-membership only … no fetch-and-confirm-the-referent-actually-exists
  mechanism."** What keeps these references valid is grove's row in the registry, which Decision
  point 2 preserves. Liveness was reassurance, not the reason. Corrected after an independent corpus
  review made the distinction, and left visible because a reader who inherited the author's test
  would apply it to a repo that had since been deleted and reach the wrong answer.*

- **The citation leg is load-bearing, not decorative.** `specs/0001-spine-artifact-contract.md:5`
  carries `depends_on: […, grove/adr-0010-versioning-is-operational]` — an approved spec with a
  frontmatter coupling edge into grove. Sharper still, `decisions/0045-artifact-versioning-kinds.md:7`
  reads `superseded_in_part_by: [grove/adr-0010-versioning-is-operational]`: a trellis decision is
  partly superseded **by** a grove ADR. Severing the citation leg would orphan a supersession pointer
  in an append-only record.

- **Several artifacts point at a *carrier* that is gone — not at content that is gone.**
  `spec-0001:90` states the versioning contract as *"shape only"* and delegates the semantics to
  *"the grove versioning companion, now plugin-carried … their single home, deliberately not
  restated here."* `core/rubrics/artifact-contract.md:91` and `.claude/agents/corpus-reviewer.md`
  delegate the same way. `ac63e7c` (2026-07-22, adr-0026 D7) deleted the vendored copies in favour
  of the plugin — sound at the time — and the plugin's removal left those pointers naming a carrier
  that no longer exists. **The content is fine:** grove's charters are readable in the repo —
  `charters/relations.md` (8838 bytes) and `charters/versioning.md` (9142 bytes), both
  `status: approved`. So what is owed is a **pointer repair, not a rehoming.**

  *An earlier draft of this record called those semantics "unhomed" and "dangling" and asked, in an
  Open Question, where they should now live. That was wrong: an independent review fetched the
  charters and found them present. The error mattered — "unhomed" invites a future session to
  restate the semantics locally, which is precisely the duplication `grove/adr-0010` removed. The
  draft also contradicted itself, since the same change described grove's corpus-reviewer charter as
  "still readable" two lines from calling its sibling unreachable.*

## Decision

**1. Grove-the-plugin is retired as this repo's operating model.** The maintainer's act on
2026-08-27 settles it; this record is the first and last decision record grove has here. Every
instruction that routes work to a `grove:<role>` subagent, names a `/grove:*` command, or reads a
plugin-carried companion is withdrawn, because none of them can be followed.

**The operative line is `.claude/settings.json`, not the prose.** That file carried
`"grove@kodhama": true` — committed by `f9d0347` (2026-07-26) *"so a fresh clone has a fleet"* — and
the kodhama marketplace still publishes grove. Deleting `.grove/` and the `AGENTS.md` block while
leaving that entry would have retired grove everywhere except the one file that actually installs
it: the next fresh clone would have reinstated the whole fleet, silently, against a green suite —
the exact failure the new absence assertions exist to prevent. The entry is removed and a test now
asserts it cannot return. **This was not self-caught; an independent review found it**, and it is
the single most consequential line in the change.

**2. Grove-the-repo remains a cited source, and nothing citing it changes.** `grove/adr-00NN`
references, `spec-0001:5`'s `depends_on`, `decision-0045:7`'s `superseded_in_part_by`, the amendment
banners across `research/` and `specs/`, and grove's entry in the recognized-repo registry
(`spec-0001:97`, `core/rubrics/artifact-contract.md:39`) all stand. **Uninstalling a plugin is not a
statement about a repository.** Conflating the two would have rewritten append-only history to
record a tooling change, which `decision-0042` forbids ("no relabel — that would forge history for
zero information gain").

**3. `.grove/` is deleted, after harvest.** It has **zero runtime readers** — nothing in `install.sh`,
`cli/`, `plugins/`, or `.github/workflows/` reads it; only test assertions and a skip-list entry
name it. What it supplied is accounted for rather than assumed away (`inv-deliberate-succession`,
backward direction):
  - `config.toml`'s `TEST_CMD`, `TYPECHECK_CMD` and `LINT_CMD` are **real and still needed** — all
    three move into `AGENTS.md`, where an agent will actually read them. `LINT_CMD`'s value is the
    honest *"none dedicated; `gofmt -l cli/` is available locally but not CI-enforced"*, worth
    carrying precisely because it stops a future session inventing a lint gate that was never there.
  - **`agents/run-resumer.md` carried one repo fact, and it is harvested too:** trellis branches are
    `<category>/<slug>` and deliberately do **not** encode an issue number, so a branch is not
    findable by searching for one. That is stated nowhere else in the tree, and it goes to
    `AGENTS.md` rather than out with the directory.
  - Four corpus tokens — `CONVENTIONS_PATH`, `ARTIFACT_DIRS`, `ARTIFACT_CONTRACT_PATHS`,
    `REPO_TYPED_CHECKS` — need no home, because the repo-owned `corpus-reviewer` already bakes the
    same values into its own charter (its default corpus is `ARTIFACT_DIRS` verbatim, exclusion and
    all). `config.toml:27–38` said so itself: that role *"is authoritative for corpus-linting here
    and has these values baked in."*
  - The remaining **seven** tokens were already `"none …"` — an honest value with nothing to carry
    forward. Three, four and seven account for all fourteen. *(An earlier draft said two and eight.
    It had missed that the `AGENTS.md` bullet it wrote already carried `LINT_CMD`'s content —
    harvested in fact while recorded as discarded.)*
  - `gates.toml`'s `intent = "human"` / `ship = "human"` is **already supplied twice over** by
    `floor-intent-gate` in the live `.trellis/rules.toml` overlay and by the maintainer's merge gate
    (`decision-0022`). Stated, not assumed.
  - `runs/` (224K, 2 runs) survives in git history. It is removed from `HEAD`, not from the archive.

**4. `spec-0006` is superseded in part, not edited.** Its AC1 requires *"The full Layer-B method and
Grove block exist in `AGENTS.md`"*. The Grove half is retired; the Layer-B half and the
entrypoint-adapter principle stand. The forward pointer goes on `spec-0006`; its AC text is left
intact, exactly as `decision-0071` handled the same spec's AC3/AC4.

**5. The stale carrier-pointers in ratified artifacts are named and parked, not rewritten here.**
`spec-0001:90` and `core/rubrics/artifact-contract.md:91` describe the companion as
*plugin-carried*, which is no longer true. Both are **ratified** (the rubric is `owner: gundi`), and
this record deliberately does not edit them: repairing an approved spec and a ratified product
rubric under cover of a retirement PR is the move `inv-deliberate-succession` names as resolving a
boundary in prose nobody approved. They are owed, and now known to be small — a re-citation, since
the content is readable at `kodhama/grove/charters/`.

`.claude/agents/corpus-reviewer.md` **is** repaired here, and the asymmetry is deliberate: it is a
repo-owned agent file rather than a ratified artifact, and it is a **live input to a running agent**,
so leaving it naming a dead carrier would degrade every corpus check taken with it.

**6. One new standing rule enters with this change, declared rather than smuggled.** `AGENTS.md`
gains: *invoke the repo-owned `corpus-reviewer` before merging a change to `decisions/`, `specs/`,
`research/` or `core/`.* No such obligation was written down anywhere before. It belongs with this
retirement because grove's departure removes the last trace of a chartered review fleet, and
`decision-0010` puts artifact conformance in an agent rather than CI — so with grove gone and no CI
check covering the contract, an unstated convention was the only thing between the corpus and no
independent review at all (`inv-gate-at-handover`). It is called out as its own decision point
because an independent review caught it arriving **undeclared**, which is the exact shape point 5
forbids. Declaring it is the fix; the rule itself stands.

## Consequences

- **`AGENTS.md` loses the managed block (`:144–159`) and gains the harvested commands.** No
  `/grove:refresh` exists to regenerate that block, so it could never have been maintained in place —
  removal was the only sanctioned exit available.
- **Four assertions in `cli/selfapply_test.go` invert** (`:102`, `:110`, `:113`, `:164–165`): the
  grove markers and `.grove/` files go from *required-present* to *required-absent*. Asserting the
  absence matters as much as the presence did — a stray branch merge restoring `.grove/` would
  otherwise reintroduce an operating model nobody chose, silently and greenly.
- **The repo's own corpus-reviewer is the one working agent grove's absence degrades.** Its checks 4
  and 5 cite the unreachable relations companion. It still runs; its edge-taxonomy reasoning now
  rests on `decision-0047` alone, which is where the trellis-side rule actually lives.
- **`core/README.md:13–16` was claiming a conformance mechanism that does not exist** — that
  conformance *"currently runs as the plugin-carried `grove:conformance-reviewer` in this repo."* It
  has been false since the plugin left. Corrected here, which means `core/` has no independent
  conformance sub-agent in this repo until `0012` packages one.
- **This does not resolve `TRL-22`.** It clears the ground the superpowers question was standing on.
  The layering record is separate and comes next, onto an `AGENTS.md` that tells the truth.
- **The eleven-to-thirteen difference is grove growing, not trellis drifting.** An earlier draft of
  this record called it drift and filed it as a finding about this repo's discipline. Checked
  against the upstream: grove publishes a *"thirteen-role operating model"* today and carries
  fourteen role charters, against eleven at install. The block's count tracked its upstream
  correctly. Retracted rather than deleted — a false disciplinary finding about oneself is worth
  showing the correction for.
- **An `approved` spec now carries a supersession mark from a `gated` record.** No rule forbids it —
  `superseded_in_part_by` is a forward pointer of the `superseded_by` class, explicitly not walked
  as a flow edge, and resolution requires only that the entry resolve. But if this record never
  reaches `approved`, `spec-0006` carries a live mark whose authority never ratified. Surfaced by
  the corpus review and stated here because it is a conscious call at merge, not an oversight.

## Open questions

- **Who repairs the ratified carrier-pointers, and when?** `spec-0001:90` and
  `core/rubrics/artifact-contract.md:91` still describe the companion as *plugin-carried*. The
  content is readable at `kodhama/grove/charters/{versioning,relations}.md`, so the repair is a
  re-citation rather than a decision about where semantics belong — but it touches an approved spec
  and a ratified rubric, so it is the maintainer's call, not an agent's. **Take all four, not just
  `:90`:** `spec-0001` also names deleted `.grove/internal/` paths at `:8`, `:25` and `:54`. An
  earlier draft of this record enumerated only `:90`, so whoever resolved it would have left three
  behind.
- **Does the repo want a chartered-role operating model at all, or was grove's real contribution the
  four-gate vocabulary?** The gates outlived the roles here. Worth answering before adopting another
  role-shaped methodology.
- **What stops the next operating model from arriving as a chore commit?** The gap this record fills
  was open for seven weeks and nothing detected it. That is a self-improvement trigger
  (`inv-self-improvement`), not just a one-off correction.

## Self-check (gate)

The maintainer's prompt — *"I think grove has been uninstalled, probably it's just your agent.md not
up to date"* — was verified before it was acted on, not agreed with: two independent checks (no
`grove:*` agent types in session; no grove directory in the plugin cache) confirmed it, and the
git log dated it. That mattered, because the *reason* offered turned out to be the smaller half —
`AGENTS.md` is stale, but so is an approved spec, a live agent's inputs, and `core/README.md`.

The blast radius was **narrowed by evidence, not by preference**: an initial reading treated every
`grove` string as breakage, which would have rewritten append-only history. Splitting the references
in two cut roughly two thirds of the proposed diff. `decision-0045:7`'s `superseded_in_part_by` edge
into grove was found *after* that split and confirms it; had the first reading been executed, that
supersession pointer would now be orphaned.

**The split was right for the wrong reason, and the independent review caught it.** The author's
test was *"does the repo still exist?"*; the contract's test is registry membership
(`spec-0001:100–103`). Conclusion unchanged, justification replaced — recorded rather than quietly
swapped, because the wrong test yields the wrong answer the day a cited repo is deleted. The same
review moved `spec-0001` into `depends_on` and `decision-0044` from `informed_by` to `depends_on`:
the two edges the record originally declared were not the ones its Decision section leaned hardest
on (`decision-0047`'s coupling test). Neither was self-caught.

The review returned one **FAIL** against the corpus, and it is not this change's:
`decisions/0044-cross-repo-depends-on-convention.md:5` carries the bare unqualified
`kodhama-0004-uniform-lifecycle`, dangling under rubric check 4. That artifact declares the gap
itself and it predates this record by seven weeks. Named here so a green-looking review is not
mistaken for a clean corpus.

The `.grove/` deletion was put to the maintainer rather than taken, because it discards 224K of
review records from `HEAD` — the standing evidence that past gates ran. Approved on the reasoning
that git history is the archive `inv-auditable-archive` asks for.

**A second independent review — fresh context, whole diff — returned nine more, one of them
critical.** `.claude/settings.json` still enabled `grove@kodhama`, so the next fresh clone would
have reinstalled the fleet and undone the retirement against a green suite. The "unhomed semantics"
framing was simply false — the charters are readable, and the same change said so two lines from
calling them unreachable. The eleven-to-thirteen "drift" was grove growing, not this repo slipping,
so the record had filed a false disciplinary finding against itself. `LINT_CMD` was harvested in
fact while recorded as discarded. A new merge-gate rule had entered `AGENTS.md` undeclared — the
precise shape Decision point 5 forbids, committed by the record that forbids it. And the parked
owed-set named one stale pointer where there are four.

**None of the twelve was self-caught.** Three came from the corpus review, nine from the diff
review; the author caught two arithmetic errors and no substantive one. A record about a repo that
let an operating model arrive and depart without a decision needed two rounds of review to stop it
departing wrongly. The count stays in rather than being smoothed away, and it is the strongest
available argument for Decision point 6.

A third leg failed and is named rather than passed over (`floor-transparency`): a cross-family
(Codex) review was attempted and died on a usage limit, returning no findings.

Two things this record deliberately does **not** do, declared so their absence is not read as
completeness: it does not repair the two ratified carrier-pointers, and it does not settle
superpowers. Both are named above with what they are waiting on. Held at `gated`, not flipped to
`approved`, because the maintainer's act covered the retirement and not this consequence set
(`decision-0046`).
