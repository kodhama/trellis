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
  grove plugin as `grove:<role>` subagents (all thirteen)."* Thirteen was itself drift — the install
  commit says eleven. The instruction had been unfollowable since the plugin left; the last recorded
  run is `20260803-093000`.

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

- **`spec-0001` is incomplete on its own terms, and this record only surfaced it.** Line 90 states
  the versioning contract as *"shape only"* and delegates the semantics to *"the grove versioning
  companion, now plugin-carried … their single home, deliberately not restated here."* Two steps got
  it here: `ac63e7c` (2026-07-22, adr-0026 D7) deleted the vendored copies in favour of the plugin —
  sound at the time, since the plugin was installed — and the plugin's later removal left the single
  home with no carrier. The exact removal date is not recoverable from this repo; the last grove run
  is 2026-08-03 and the last `.grove/` commit 2026-08-04, so the gap has been open at least three
  weeks. `core/rubrics/artifact-contract.md:91` and `.claude/agents/corpus-reviewer.md:36,42`
  delegate the same way. The delegation was sound when written. It is dangling now.

## Decision

**1. Grove-the-plugin is retired as this repo's operating model.** The maintainer's act on
2026-08-27 settles it; this record is the first and last decision record grove has here. Every
instruction that routes work to a `grove:<role>` subagent, names a `/grove:*` command, or reads a
plugin-carried companion is withdrawn, because none of them can be followed.

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
  - `config.toml`'s `TEST_CMD` and `TYPECHECK_CMD` are **real and still needed** — they move into
    `AGENTS.md`, where an agent will actually read them.
  - Four corpus tokens — `CONVENTIONS_PATH`, `ARTIFACT_DIRS`, `ARTIFACT_CONTRACT_PATHS`,
    `REPO_TYPED_CHECKS` — need no home, because the repo-owned `corpus-reviewer` already bakes the
    same values into its own charter (its default corpus is `ARTIFACT_DIRS` verbatim, exclusion and
    all). `config.toml:27–38` said so itself: that role *"is authoritative for corpus-linting here
    and has these values baked in."*
  - The remaining **eight** tokens were already `"none …"` — an honest value with nothing to carry
    forward. Two, four and eight account for all fourteen.
  - `gates.toml`'s `intent = "human"` / `ship = "human"` is **already supplied twice over** by
    `floor-intent-gate` in the live `.trellis/rules.toml` overlay and by the maintainer's merge gate
    (`decision-0022`). Stated, not assumed.
  - `runs/` (224K, 2 runs) survives in git history. It is removed from `HEAD`, not from the archive.

**4. `spec-0006` is superseded in part, not edited.** Its AC1 requires *"The full Layer-B method and
Grove block exist in `AGENTS.md`"*. The Grove half is retired; the Layer-B half and the
entrypoint-adapter principle stand. The forward pointer goes on `spec-0006`; its AC text is left
intact, exactly as `decision-0071` handled the same spec's AC3/AC4.

**5. The unhomed semantics are named and parked, not invented.** `spec-0001:90`,
`core/rubrics/artifact-contract.md:91` and `.claude/agents/corpus-reviewer.md:36,42` point at a
companion that no longer resolves. Their pointers are marked honestly in this change so no reader
believes a live home exists — but **rehoming versioning and relations semantics is a contract
decision, not a cleanup, and this record does not make it.** Naming the exemption and leaving it for
the maintainer is what `inv-deliberate-succession` asks for; quietly writing a replacement home into
a spec under cover of a retirement PR is what it forbids.

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
- **The eleven-to-thirteen drift is left as evidence, not corrected.** The install commit says
  eleven roles, the withdrawn block says thirteen. Both are now historical; there is nothing to
  reconcile and no live claim to fix.
- **An `approved` spec now carries a supersession mark from a `gated` record.** No rule forbids it —
  `superseded_in_part_by` is a forward pointer of the `superseded_by` class, explicitly not walked
  as a flow edge, and resolution requires only that the entry resolve. But if this record never
  reaches `approved`, `spec-0006` carries a live mark whose authority never ratified. Surfaced by
  the corpus review and stated here because it is a conscious call at merge, not an oversight.

## Open questions

- **Where do versioning and relations semantics live now?** `spec-0001:90` delegates to an unreachable
  home. Options: restate them in `spec-0001` (reversing the adr-0010 de-reflection), re-cite the grove
  ADRs directly as repo artifacts, or accept the gap. **Owed, and blocking any corpus-reviewer change
  that depends on edge taxonomy.**
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

Two things this record deliberately does **not** do, declared so their absence is not read as
completeness (`floor-transparency`): it does not rehome the unhomed versioning semantics, and it does
not settle superpowers. Both are named above with what they are waiting on. Held at `gated`, not
flipped to `approved`, because the maintainer's act covered the retirement and not this consequence
set (`decision-0046`).
