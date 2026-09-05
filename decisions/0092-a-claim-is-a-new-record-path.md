---
id: decision-0092
type: decision
depends_on: [decision-0081, decision-0089]
changes: [decision-0089]
informed_by: [decision-0028, decision-0040, decision-0078, decision-0082, decision-0085]
owner: agent
date: 2026-09-05
---

> **Provenance.** `#272` (`85594bf`) fixed the code, marked the record with dated notes, and
> deliberately did **not** perform the supersession — *"that is the maintainer's act, not the
> author's"* (`decision-0089:107-148`, citing `decision-0081`). **TRL-46** was filed by the
> coordinating session from `#272`'s code review, *"as the consumer `decision-0078` requires"*. The
> maintainer ruled the general shape on 2026-09-05, relayed to this session: *"Factual corrections
> get dated notes. Rule divergences get successors."* The two marked points are rules, so this is the
> successor. Written by a session that authored neither `#266` nor `#272`: the divergence below was
> re-derived from the script and its tests rather than taken from the ticket, and neither existing
> summary of point 4 survives it — the record's own marker overstates the point, the ticket
> understates it.

# 0092 — a claim is a new record path, and a branch can collide with itself

## Context

`decision-0089` (`#266`) stated the decision-id guard's rules. `#272` then fixed two confirmed false
passes in `.github/scripts/decision-id-guard.sh`, and fixing the code made the record wrong: a record
on `main` is current truth (`decision-0082` — merging is the acceptance), so two of its rules now
describe a guard that no longer exists.

**The authority for every statement below is the shipped script**, with `cli/decision_id_guard_test.go`
as the executable second reading — it runs the production script as an external file against fixtures,
so each rule here names the test that pins it.

### Point 1 — the failure set is wrong in two directions, and silent about a third

`decision-0089:55-57` (`:54-56` before this change adds a frontmatter line):

> **1. A CI check on the pull request, `decision-id-guard`.** It fails when a `decisions/NNNN-*.md`
> file **newly added** by the branch carries an id that is already on the base branch, or that is also
> newly added by a **lower-numbered open pull request**.

*The first sentence is not in dispute and is not superseded; everything below concerns the second.*

| what it says | what the guard does | where |
| --- | --- | --- |
| two failure conditions | **three** — the third is one diff claiming an id twice | script `:13-21`, `:267-279`; `TestDecisionIDGuardFailsOnDuplicateIDWithinOnePR` |
| a claim is "newly added" | `added`, `copied`, or a rename whose **destination** id differs from its source | script `:233-243`; `…TreatsRenameDestinationAsAClaim`, `…CopiedRowCarriesAPreviousFilename`, `…IgnoresSlugOnlyRename` |
| — | a whole second outcome class: **exit 2, could not run** | script `:47-52`, `:69-73`; `…RequiresItsPRNumber`, `…FailsClosedOnProcessingError`, `…RefusesAPathWithASpace` |

The broadening applies to the **rival** side of the comparison too, not just to this branch's own
files: one `claims` file is built from every row of every open PR (`:243`) *before* this PR's rows are
separated out (`:246`) and the rivals selected from the remainder (`:303`). So a lower-numbered PR's
rename destination is a rival claim exactly as an added file is. *That is read off the single shared
code path, not off a test — no fixture exercises a rival whose claim is a rename.*

### Point 4 — narrower than its own marker says

`decision-0089:80-83` (`:79-82` before this change), clause by clause:

| clause | verdict | where |
| --- | --- | --- |
| "A claim is an *added* file … **nothing else**" | **false** — `copied` and a rename destination claim too | script `:233-243` |
| "a rename is a file that already exists" | **false** of the destination path, which is what claims | script `:235-240` |
| "Modifying `decision-0087` does not claim `0087`" | **stands** | script `:241`; `…IgnoresNonAddedFiles` |
| "`decisions/README.md` and `decisions/0089.md` are not claims" | **stands** | script `:78-86`; `…IgnoresNonDecisionFilenames` |
| "the status filter lives in the script rather than in the workflow's `gh --jq` expression" | **stands** | the workflow passes env and one `run:` line; the filter is at script `:233-243` |

**Both existing summaries of this point are wrong, in opposite directions.** The marker at
`decision-0089:140-141` — *"**Point 4** is replaced outright"* — **overstates it**: two of its three
examples and its whole rationale survive. TRL-46 **understates it**, rendering the point as *"a rename
is not a claim"*, which is one of the five clauses above; the headline rule is false independently of
renames, because of `copied`. Retiring the point wholesale would take three true statements down with
the two false ones — the disposal `decision-0040` guards against by requiring the scope to be stated
at clause level, which is what the forward pointer on `decision-0089` does.

### Wider than points 1 and 4, in one place

**No point of `decision-0089` states what a rename does to the id it leaves behind.** The shipped
guard decides it: the source id stays taken while the PR is open, and the guard says so in the error
rather than leaving the reader with a message that reads as a contradiction (script `:288-299`,
`…ExplainsAnIDThisPRIsVacating`). TRL-46 names the behaviour, in point 4's row; the *record* has no
clause there to supersede, so point 3 below is an addition rather than a correction, and it is stated
here because a reader of `decision-0089` alone would have no way to reach it.

## Decision

**1. A claim is a file the pull request puts at a `decisions/NNNN-*.md` path it does not already
occupy.** `added` and `copied` claim; `renamed` claims its **destination** id when that differs from
its source. `modified`, `removed`, `changed`, anything else, and a slug-only rename claim nothing.
Concretely: `0088-old.md` → `0090-new.md` claims `0090`; `0087-one-gateway.md` →
`0087-a-better-slug.md` claims nothing, or every legitimate retitle would go red against its own
record on the base branch.

**2. Three failure conditions, not two.** An id on the base branch; an id also claimed by a
lower-numbered open pull request; and **two files in one diff claiming one id**. The third takes no
tie-break — both files are on one branch and the author picks which moves:

> `::error::PR #300 adds two files claiming decision-0090: decisions/0090-a.md decisions/0090-b.md — one id, one record. Renumber one of them.`

**3. A rename claims its destination and does not release its source.** The id a record is moving
*off* stays taken until the pull request merges; handing it to a second branch on the strength of an
unmerged rename would cause the collision the guard exists to prevent. Wrong, if ever, in the
direction of one wasted number rather than a duplicate — and the message says so rather than leaving
the reasoning in a source comment: *"a branch cannot free an id for its own use"*.

**4. Not-a-collision is not the same as green.** Exit 0 is clean (a `::notice` is still 0), exit 1 is
a collision, **exit 2 is "could not run"** — a missing `GUARD_PR_NUMBER`, an unfetchable base branch,
an unreadable PR list, any failing processing step, or a `decisions/NNNN…` path containing a space
that whitespace-splitting would otherwise silence into "no claim". This is `sh`, not bash, so
`pipefail` is unavailable and every step's status is checked instead. **An internal failure must
never read as "no id claimed, nothing to check."**

**5. What of `decision-0089` stands, exactly.** Points 2, 3 (its rule), 5 and 6 are untouched — and so
is **point 1's first sentence**, that the check is a CI check on the pull request named
`decision-id-guard`: nothing above restates it, point 5 independently names the same script and
workflow, and it remains true. What point 1 loses is its **second sentence**, the rule. Within
point 4: modifying a record does not claim its id, `decisions/README.md` and `decisions/0089.md` are
not claims, and the status filter belongs in the script because the tests can only reach that half.
The forward pointer added to `decision-0089` carries this scope, and nothing wider.

## Consequences

- **`AGENTS.md`'s one-line derivative is updated in this change** (`decision-0028`: a source and its
  derivative move together). Its allocation row said an id must be free *"on `main` **and** on every
  open PR"* — true, and now incomplete by exactly the failure condition point 2 above adds. It gains
  *"and within your own diff"*. That line is where an agent actually reads the rule; the record is
  where it is argued.

- **The record ↔ script pair has no guard, and this is the second time it has cost something.**
  `decision-0028` asks for a check per source/derivative pair, and the script/workflow pair has one
  (`TestDecisionIDGuardWorkflowRunsTheScript`). The script/record pair has nobody: `#272` changed the
  behaviour and `decision-0089` went stale in silence until a human read both. Not fixed here — a
  test asserting that English prose matches a shell script is either a brittle string match or a
  judgment no test makes. Named so the next divergence is a known cost rather than a surprise.

- **The dated notes `#272` left on `decision-0089` stay exactly as they are** (`inv-auditable-archive`;
  `decision-0085` §5, a note beside the claim rather than an edit that makes the record look prescient).
  They are the record of the interval in which the corpus knowingly disagreed with the code. This
  record closes that interval; it does not tidy it away. The one permitted edit to `decision-0089` is
  the forward pointer, which `decision-0082` makes the whole mark of supersession.

- **`TRL-46` is discharged, which is what `decision-0078`'s address test asked for.** `#272` parked
  the supersession *with* a named consumer; this record is that consumer arriving. The deferral cost
  one day and two dated notes, and it was the right trade: the author of a fix marking his own record
  superseded is the act `decision-0081` reserves for the maintainer, and merging this is that act.

- **Whether a rename should *release* its source id is still open, and deliberately.** The script
  names the question and declines it; so does point 3. If the answer ever changes, it changes here,
  not in a comment.

- **The maintainer's ruling that directed this record is not minted as a corpus rule.** *"Factual
  corrections get dated notes. Rule divergences get successors"* is quoted in the provenance as this
  record's warrant, not stated as general law — that would be a second logical change, and a rule
  about how the corpus is maintained deserves its own argument rather than a ride on this one.

## Self-check (gate)

- **Every rule above was read off the script and matched to a test**, named individually in the two
  tables. Where no test exists — the rival-side broadening — the record says so rather than implying
  coverage.

- **The counter-check ran, and it moved the answer.** `TRL-46` and `decision-0089`'s own marker say
  *points 1 and 4*. Point 4 turns out to be **narrower** than "replaced outright" (three of its five
  clauses stand) and the divergence is **wider** in one place the marker does not reach (a rename's
  source id). Both are recorded above against the specific lines that make them true.

- **Point 5 was checked as a candidate divergence and is not one.** Its parenthetical names three
  injectable inputs where the script has five, but `GUARD_BASE_REF` and `GUARD_REPO` were already
  there at `7c323e3` — the commit that merged `#266`. `#272` did not touch that, so widening the
  supersession to point 5 would be wrong.

- **Points 2 and 3's rule were verified against the emitted strings, not assumed.** The tie-break is
  `[ "$rival" -lt "$pr" ]` (script `:311`) and the two messages match `decision-0089`'s quoted text
  word for word (`:312`, `:315`). Point 3's known error — its second example naming `#299` where the
  script interpolates the current PR — is already flagged in that record and is an example, not a
  rule; it stays where it is.

- **The line cites of `decision-0089` are given for the file as this change leaves it**, with the
  pre-change numbers in parentheses. Adding `superseded_in_part_by` to its frontmatter displaces its
  body by one line — the identical off-by-one `decision-0085:16-17` disclosed and `decision-0089`'s
  own gate caught in four places. Citing only the `main` numbers would repeat the defect that record
  paid to fix.

- **The independent pass ran, and its one FAIL is not this branch's.** `corpus-reviewer` returned
  PASS on every check that reaches this change — frontmatter with no `status`, `decision-0092`'s id
  unique, all seven of its frontmatter references resolving, the required sections, and the
  clause-level scope of the pointer on `decision-0089` — and it independently confirmed the
  post-change line cites at `:55-57` and `:80-83`. Its one FAIL is
  `decisions/0044-cross-repo-depends-on-convention.md:5`'s `kodhama-0004-uniform-lifecycle`, the
  dangling entry that record discloses about itself and that `decision-0076`, `decision-0088` and
  `decision-0090` each recorded before this one; it predates this branch by seven weeks and is
  untouched by it. Repeated here so a clean change is not read as a clean corpus.

- **The review's own tooling was degraded, and that is disclosed rather than papered over.** Its
  `Grep`/`Glob` failed on every call (`ENOENT … posix_spawn 'rg'`, the known sandbox defect), which
  silently skips the corpus-wide checks. It enumerated the corpus from the repository index and read
  all 106 artifacts individually instead, so the sweep was completed by another route rather than
  truncated.

- **The author's own sweep is repository-wide, and its first version was not.** `0089` occurs on
  **25 lines across 9 files** outside the two records, and they fall into four groups that add to 25:

  | group | count | sites |
  | --- | --- | --- |
  | the number used as a fixture path, an example, or an id someone might allocate | 15 | 12 assertion/fixture strings in `cli/decision_id_guard_test.go`, `decision-id-guard.sh:75`, and 2 in `docs/superpowers/` |
  | provenance headers naming TRL-40 / this record | 3 | `decision-id-guard.sh:4`, `decision-id-guard.yml:4`, `decision_id_guard_test.go:5` |
  | citations of a point that **stands** | 6 | `cli-ci.yml:17` → p5 · `decision_id_guard_test.go:120` → p3 · `decision-0087:28` → p2 · `decision-0078:7` and `:160` → p6 · `decision-id-guard.sh:77` → the clause of p4 that survives |
  | the derivative this change updates | 1 | `AGENTS.md:51` |

  **Nothing outside `AGENTS.md:51` restates point 1's rule or a superseded clause of point 4.** *An
  earlier draft of this bullet said "four sites", from a sweep over `decisions/`, `research/`,
  `core/`, `profiles/`, `docs/` and the two root instruction files — scoped to the corpus while the
  sentence claimed the repository, so `.github/` and `cli/` went unswept. Its replacement then
  double-counted two test lines as fixtures. Both corrected against a re-run before merge; the
  conclusion survived each pass and only the count moved. Recorded rather than quietly fixed: a
  record about a claim left standing after its subject moved is the wrong place to leave one.*

- **The unverified half is unchanged and still unverified.** These tests never reach `gh` or the
  network, so the live fetch path — pagination over every open PR, `previous_filename` arriving as the
  fourth field — is exercised only by real runs. Point 1's third row and point 3 both depend on it.
  This record's own pull request is one such run: `#280` fetched `main`, walked the open list and
  printed `- decision-0092 (decisions/0092-a-claim-is-a-new-record-path.md) — free.` The id was
  also checked offline first, by running the same script against `origin/main`'s `decisions/` and
  the real file rows of the one open PR — a check, not the eyeballing this guard replaced.
