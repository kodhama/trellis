---
id: decision-0091
type: decision
depends_on: [decision-0037, decision-0047, decision-0057, decision-0068, decision-0070, decision-0073, decision-0090]  # coupling, not provenance (decision-0047). decision-0068: D12 is the record this one changes in part -- its read-inventory row at :291 states the refusal outcome for the CONTENT read of CLAUDE.md/AGENTS.md unconditionally, and its ground at :260-261 ("both static chains are loaded by the host before any hook runs") is what D1 below denies for one input. An independent corpus review of this branch found it carried as informed_by, which is the category error decision-0047 names: had D12 said something else there would be nothing here to change. decision-0088 and decision-0090 both carry it in depends_on for the adjacent clause, and this record follows them. decision-0057: the ruling rests on its rule 2 -- the adapter contract is exactly one standalone `@AGENTS.md` line in CLAUDE.md -- which is what both components read and therefore what the gate can know. Rule 2 mandates that line in this repo; it does not rule on how the host resolves AGENTS.md in general, and this record does not read it as doing so (its Open questions carry the counterexample, a nested import chain the contract does not describe and neither component sees). If rule 2 said something else, the gate would be reading the wrong line and D1 would be wrong with it. decision-0073: D1's closed set is what makes "the inline managed block" a nameable shape at all, and D2's test -- refuse only over files this host actually loads -- is the test D1 below applies to the second of the two components. decision-0090: D2's fourth clause, "removes an existing refusal", is why this record exists rather than the change landing alone; TRL-44 shipped the read wired to the message only and said so in terms, deferring exactly this. decision-0070: D5 is the opt-out ruling whose branch this change must not disturb, and the interaction (opted-out project, un-imported block) is one of the two states D2 below rules on. decision-0037 and decision-0047 for the reason decision-0088 and decision-0090 declare them: this record's `owner: agent` and this frontmatter's shape are contingent on their rulings, not informed by them
changes: [decision-0068]  # ONE outcome for ONE input, wherever D12 states it -- scoped by PARAGRAPH, not by sentence, because two successive corpus reviews each found another sentence carrying the same proposition after the previous ones were named. D12's whole opening paragraph :259-262, the :291 inventory row's AGENTS.md half, and :318-319's "is refused" half (its "not repaired" survives). The forward pointer on 0068 states that scope and, for what it does NOT change, says so in those words rather than claiming it stands (decision-0040)
informed_by: [decision-0010, decision-0028, decision-0079, decision-0082, decision-0088]  # decision-0028 is where the copied-matcher guard obligation comes from and is already discharged for this read (TestInstallScriptAgentsImportGateMatchesHook, shipped with TRL-44); it is informed_by rather than depends_on following decision-0090's treatment of the same record. decision-0088 for the budget's history that D4 below leaves untouched
owner: agent
date: 2026-09-04
---

> **Provenance.** Filed as `TRL-48` by the agent session that merged `#273` (`700ca30`), which
> found this defect while fixing its sibling, fixed only the half that needed no record, and said
> so in the code it shipped. Written by a different session, which did not author `#273`.
> The maintainer ruled on 2026-09-04 that it be fixed, record and all; he did not rule on which
> way the substance goes. That is what this record decides.
>
> **The code lands separately**, on `fix/trl-48-agents-md-refusal-gate`. That is not incidental:
> `decision-0090` D2 requires a read that removes a refusal to reach the corpus *before* it lands,
> so on this branch the corpus half is ahead of the behaviour it describes, deliberately. A
> conformance pass over this branch alone will measure `install.sh` still refusing and read D1 as
> false; it is describing the change, not reporting it as shipped.

# 0091 — a static-delivery conflict is a conflict *for a reader*, not for a directory

## Context

Trellis delivers its rules two ways, and both must refuse to deliver where the other already
does. `install.sh` renders `.claude/rules/trellis.md`; the plugin hook (`staleness.sh`) injects at
SessionStart. Each therefore probes the project for a Trellis managed block — the inline shape,
`decision-0073` D1's S4 — and stands down when it finds one.

They probed differently. The hook checks `AGENTS.md` only behind a standalone `@AGENTS.md` import
line in `CLAUDE.md`, because Claude Code reaches that file no other way (`decision-0057` rule 2).
`install.sh` checked `CLAUDE.md` **and** `AGENTS.md`, ungated. `#273` copied the hook's gate into
the installer and wired it to the **message** only, leaving the **render refusal** on the ungated
probe, and recorded the reason in the script: gating the refusal *removes* a refusal, and
`decision-0090` D2's fourth clause sends that to this corpus before it lands.

**Measured on `700ca30`, one scratch repo, both deliveries.** A project that has not opted out,
carrying the shipped `block-inline-b.md` at `AGENTS.md`, with no `@AGENTS.md` line anywhere:

| component | what it did |
|---|---|
| `install.sh` | `NOT rendering .claude/rules/trellis.md: this project already delivers the rules statically (managed block in AGENTS.md)`, then `Migrate first: delete the managed block in AGENTS.md` |
| `staleness.sh`, same repo | injected the posture header and the full rule set |

The installer declined to install where the plugin governs, and the remedy it offered was to
delete the one thing that *does* deliver rules on that repo — to Codex, which reads `AGENTS.md`
directly. Adding the `@AGENTS.md` line makes the two agree again: both refuse.

### The two readings

- **The hook's.** A block is a conflict only if *this host* will load it. Makes the two
  components agree; removes a refusal.
- **The installer's.** Any managed block is a conflict, because the installer **writes** where the
  hook only **injects**: a false negative in the writer leaves a durable artifact and means rules
  delivered twice, while a false negative in the injector costs one session that is re-decided at
  the next.

### The asymmetry is real, and it is not answered by denying it

The second reading's argument is sound as far as it goes. The installer's mistake persists on
disk with no scheduled re-check; the hook's does not survive the session. If the reader-state that
justified the render later changes — someone adds a standalone `@AGENTS.md` line to `CLAUDE.md` —
the rendered file and the block are both loaded, and nothing about the original install run is
re-run to notice.

**What answers it is that the durable state has a named consumer that re-presents it**
(`decision-0078`'s test), and that consumer is the component on the other side of this very
divergence. The hook re-reads the same import line every session. Verified on a scratch repo, in
sequence:

| state | hook, next session |
|---|---|
| rendered file + un-imported `AGENTS.md` block | stands down: *"already loaded from `.claude/rules/trellis.md`"* — one delivery, correctly reported |
| the same repo, `@AGENTS.md` line added afterwards | `TRELLIS_STATIC_SHAPES_CONFLICT`, naming `AGENTS.md`, the double delivery, and the remedy |

**That watcher is conditional, and an automated reviewer on the code PR was right to press it.**
`decision-0068` records that the skills-directory plugin path *"requires accepting the workspace trust
dialog, which a headless run cannot grant — and headless CI is exactly the case the install path exists to
serve."* Where the plugin does not load, the hook does not run, and `.claude/rules/` is loaded regardless —
so on the headless curl path the session-start watcher is absent exactly where the harm lands. What remains
there is the **second** watcher: `install.sh` itself. A re-run on that repo takes the gate the other way and
prints `WARNING: .claude/rules/trellis.md ALREADY EXISTS … the double delivery described above is live right
now`, with the remedy — verified in sequence on a scratch repo, and the curl path is re-run to upgrade.

So the claim this record actually makes is the narrow one: the writer's false negative is **watched twice
and guaranteed by neither** — by the hook wherever the plugin loads, and by the installer on any re-run.
The false positive it replaces was neither watched nor conditional: it fired on every run, forever, on a
project that never had a Claude-side conflict at all.

### What the ungated probe actually bought, measured rather than assumed

The one case where the ungated probe is right for a reason the gate cannot see is a **nested
import**: `CLAUDE.md` imports `docs/foo.md`, which imports `@AGENTS.md`. The host resolves that
chain; neither script's anchored one-line grep does. **This record does not close that case, and
for it D1 is a regression rather than a neutral blind spot. That correction is owed to a reviewer
on this PR, and it is the sharpest thing said against this ruling.**

An earlier draft claimed the status quo did not pay this cost either, on a measurement showing
`staleness.sh` injecting the full rule set on that repo shape while the host loads the block
through the chain — double delivery already live, the ungated refusal declining only to add a
*third*. **That measurement assumed the hook runs.** Where the plugin does not load — the headless
curl path, which `decision-0068` names as the case the install path exists to serve — the status
quo yields **one** delivery (the block, via the chain) and D1 yields **two**, with no watcher:
measured on the code branch, the installer **renders on run 1 and renders again on run 2 in
silence**, because the re-run consults the same anchored line. (The run-1 message at the time of
that measurement also asserted the host does not read the file; that sentence is withdrawn and the
shipped NOTE now names only the route the installer can see, plus the chain it cannot. The
delivery count and the absent watcher — the substance here — are unaffected by the wording.)

So the honest accounting is: D1 **trades** an unconditional wrong refusal on every mixed-host repo
for a wrong render on the narrower nested-import shape, unwatched in the headless environment. It
is a trade, not a free correction, and a reader who thinks the narrower harm is the worse one has
a real argument rather than a misreading. What the change does **not** get to say is that nothing
is lost. `TRL-49` carries the resolver that would close it; until then the shipped NOTE names the
chain case in terms and tells the reader to check for it, rather than asserting a fact about the
host that one anchored line cannot establish.

### Why "keep the refusal and fix the wording" was not available

If the installer's reading governed, the fix would be to keep the refusal and correct whatever
misdescribes it. There is nothing to correct: the refusal's text — *"this project already
delivers the rules statically … Adding the rendered file would deliver them twice, and no hook can
undo that"* — is not imprecise about a true state. It is false about a repo where the two files
are read by different hosts, and its remedy is destructive there. A refusal whose only honest
wording is *"this may be a conflict for a host that is not the one you are installing for, so you
may not install"* is not a wording fix; it is the ruling below, stated in the negative.

## Decision

**1. A static-delivery conflict is a conflict *for the reader of the delivery being decided*.**
`install.sh` renders `.claude/rules/trellis.md`, a file only Claude Code loads. A managed block
is a conflict against that render only when Claude Code loads the block. The **render refusal
therefore takes the same gate the message already takes** — the hook's own matcher, already copied
and already pinned (`TestInstallScriptAgentsImportGateMatchesHook`, `decision-0028`). This is
`decision-0073` D2's test — *refuse only over files this host actually loads* — applied to the
second component.

**The ground is the adapter contract, not a claim about how the host resolves imports**, and the
difference is load-bearing rather than pedantic. `decision-0057` rule 2 fixes one standalone
`@AGENTS.md` line as the contract; that line is what both components read, so it is what a gate
can know. It is **not** true that `AGENTS.md` reaches a session only that way — the nested chain
measured above is the counterexample, and an earlier draft of this very clause asserted the
universal and was caught by a reviewer on this PR contradicting the record's own Context,
Consequences and frontmatter. What D1 rules is that the installer must not refuse over a file it
has no reason to believe this host loads; where the contract is silent and a chain exists anyway,
D1 is wrong on that repo and says so rather than being read as licence.

So: the two components **agree on every input** — which is what the divergence cost and what this
ruling buys — and agreement is not correctness, on the one shape where both are wrong together.

**2. Removing a refusal does not license removing the statement.** The render path says what it
rendered over: that the block is there, that this host does not read it, that a host which does
(Codex CLI) gets its rules from the block and nothing from the rendered file — so deleting the
block ungoverns *that* host — and that adding the import line would make both load. Every claim in
that NOTE is about **the block and this host**, never about the project, which is the discipline
`#273`'s review imposed on its sibling NOTE and the same failure class `decision-0073` exists to
end. It is mutually exclusive with the opt-out branch's NOTE for the same shape — but **by an
explicit test, not by position**: the render NOTE is printed from the trailing summary, which every
project-scope run reaches, so it carries its own `opted_out` guard. Saying only *"an opted-out project
renders nothing"* would describe the outcome and misdescribe the mechanism, and a reader implementing from
that sentence would omit the guard and ship both messages on one run. A test pins it.

**3. The gate reaches this refusal and no other.** A block in `CLAUDE.md` is refused exactly as
before, with or without an import line. `decision-0070` D5's opt-out branch is untouched: it
refuses on the opt-out itself, not on `$static_conflict`, and its `#273` message-gating stands.
Nothing here licenses gating a *different* refusal on an inference about the reader.

**4. This discharges `decision-0090` D2 for this read, and does not move the budget.** The change
adds no read; `install.sh`'s count and its per-path enumeration are unchanged. What changes is the
read's **classification**: the `@AGENTS.md` gate now removes a refusal, so the enumeration comment
that called it *"selects no artifact and removes no refusal"* is corrected in the same commit and
cites this record (`decision-0028` — a source and its derivative are fixed together). `decision-0090`
D1 keeps the count in `cli/install_script_test.go`; nothing here touches it.

**5. `decision-0068` D12 is changed in part, and the scope is stated by paragraph, not by sentence.**
D12 asserts the unconditional refusal in three places: its **whole opening paragraph** — topic sentence
(*"`install.sh` refuses to render when the project already delivers the rules statically"*), ground (*"both
static chains are loaded by the host before any hook runs"*) and conclusion (*"rendering blindly ships
guaranteed double delivery"*); its **read-inventory row**, `AGENTS.md` half only; and its closing *what this
does not cover* clause, *"a repo carrying an inline block is refused, not repaired"*, whose *"is refused"* half
is false and whose *"not repaired"* half survives — this script still repairs nothing. Each is false for the one
input D1 rules on and for no other. What this record does **not** change is stated in those words rather than
as *"stands"*: the same row's `CLAUDE.md` half, the read count and every other row (D4 adds no read, and takes
no position on whether D12's counts are still accurate — they were overtaken by TRL-37, TRL-38 and TRL-44, and
this pointer does not adopt them), and D12's argument for why a content read exists there at all, which is
untouched and is why the probe still runs on both files.

**Three independent corpus reviews of this branch produced that clause, and none of the three findings was the
author's.** The first: the draft carried `0068` as `informed_by` and claimed to change nothing, which would have
corrected the derivative comment in `install.sh` while leaving the corpus clause it derives from asserting the
opposite — the `inv-graph-maintenance` defect `decision-0090` names, in mirror image, inside the record
correcting it. The second: the pointer then named two of the three places and affirmatively listed the third as
standing. The third: it named three of four, and its *"stands"* list affirmed a row another pointer on the same
field already records as false. **The pattern, not any one miss, is the finding** — sentence-level scoping on a
record that says the same thing four ways will always be one sentence short, so the scope is now the paragraph
and the not-changed list no longer certifies anything as true. That is why all three are recorded here instead
of folded into a clean-looking draft, for the reason `decision-0090`'s Self-check gives for counting its own.

## Consequences

- **A documented mixed-host layout can install Trellis for Claude again.** A repo whose
  Codex-facing rules live in an `AGENTS.md` block, with Claude served by the plugin, was refused
  the rendered file on every run and told to delete the block. It now renders, and each host has
  exactly one delivery.
- **One shape moves from refuse to render, and it is the only one.** Non-opted-out project, block
  in `AGENTS.md`, no standalone `@AGENTS.md` line in `CLAUDE.md`. Every other input to the refusal
  is unchanged, and `TestVendorAgentsBlockRefusalFollowsTheImportGate` pins both directions on the
  same scratch repo — the un-imported block must render, the imported one must still refuse, and a
  block in `CLAUDE.md` must be untouched by either.
- **The argument this ruling rests on is itself pinned by a test.** The write/inject asymmetry is
  answered by the hook catching the late-arriving import line; the subtest that renders, plants the
  line, and asserts `TRELLIS_STATIC_SHAPES_CONFLICT` is where that stops being true if it ever does.
  A record whose reasoning has no failing artifact behind it is the shape `decision-0090` D1 warns
  about, so this one does not rely on being re-read.
- **What is *not* claimed here.** That the two components are *right* on every project. They now
  **agree** on every project, which is what D1 delivers and what the divergence cost — but agreement is
  not correctness: on a nested import chain they agree with each other and both are wrong about the
  host. The residue is joint error, not divergence, and on that shape D1 is a **regression** against
  the status quo in the headless environment, stated above with its measurement rather than filed
  under "pre-existing". Closing it means teaching both the same resolver (`TRL-49`), not widening
  this gate. Nor is it claimed that the shipped NOTE knows the project is single-delivery: it says
  what the installer can see, and names the chain it cannot. That `install.sh` gained a licence to read
  more project state: it reads none — D1 and D2 branch on values two existing reads already
  produced, the count is unchanged, and `decision-0090` D2 still governs the next read on the same
  terms it governed this one. That any refusal may be gated on
  a *guess* about the reader — this one is gated on the adapter contract's own single line, read
  with the hook's own anchored matcher, which is why the two cannot diverge again by drift.
- **The recurrence this ends** is two components shipping the same question with two answers, each
  correct in its own file. `decision-0028` guards the matcher against drift; nothing guarded the
  *use* of it, and one component used it for a message while the other refused without it.

## Self-check

- **Written by a session that did not author `#273`**, on the issue that session filed against its
  own deliberate omission — the same structural advantage `decision-0090` records for itself, and
  with the same limit: it is not independent review of this ruling, which is still owed
  (`inv-independent-judgment`). `corpus-reviewer` checked this record against the artifact
  contract before it was merged and **returned FAIL three times running**, on the three findings D5
  records. That is a conformance check and not a review of the argument, and it still caught what
  would have made this record's own D4 hypocritical. **The argument itself got adversarial review,
  from automated reviewers on both PRs, and every finding landed.** One found that the
  `@AGENTS.md` probe both components share was anchored with no BOM tolerance, which turned this
  ruling's transient false negative into a silent permanent one for a BOM'd `CLAUDE.md`; it is fixed
  in both copies in the code change, and it is the reason the watcher paragraph above can be stated
  at all. The other pressed the watcher claim against `decision-0068`'s headless finding and was
  right — that is what narrowed it, and a follow-up from the same reviewer showed the
  nested-import case is a regression rather than a shared blind spot, which is the sharpest thing
  said against D1 and is now stated as such. A fourth found that **D1's own stated ground
  contradicted three other parts of this record**: it asserted `AGENTS.md` is loaded *only*
  through a standalone line, while the Context measured a chain that defeats it, the Consequences
  conceded both components are wrong about the host there, and the frontmatter said in terms that
  this record does not read `decision-0057` as ruling on host resolution. D1 now rests on the
  adapter contract — what a gate can know — rather than on a universal that this record itself
  disproves. A fifth and sixth, on the draft that fixed the fourth: a measurement clause still
  described the *withdrawn* NOTE wording as if it were what ships, and — the one that matters —
  this Self-check claimed *"the status quo's harm needs no unlucky sequence while both of D1's
  do"*, which is false of the nested-import harm the same bullet had just called a straight
  regression. That one **overstated the ruling's case in the ruling's own favour**, three sentences
  after conceding the opposite. **None of the six was found by the author**, and the pattern across
  them is one failure repeated: stating a claim one step stronger than the evidence carries. The
  count is kept because the rate is the finding — six in one record, all in the same direction —
  and a reader weighing how much to trust the argument should know it.
- **Both behaviours were reproduced before either was argued**, on scratch repos, running both
  deliveries on the same repo — the divergence, the control with the import line, the post-render
  arrival of the import line, and the nested-import case that is the strongest thing the status quo
  had. **The nested-import case was got wrong twice, in opposite directions, and both corrections
  came from outside.** The first draft treated it as a cost the ruling simply accepts. The second
  measured the hook injecting there and concluded the status quo does not pay it either — an
  over-correction, because that measurement silently assumed the hook runs. A reviewer on this PR
  supplied the environment that breaks it, and the third version is the one above: headless, the
  status quo delivers once and D1 delivers twice, unwatched. Arguing against a weakened form of the
  opposing case is a different failure from ignoring it and is no better, and this record did it
  once before catching it.
- **The strongest argument against this ruling, stated at its real strength rather than at the
  strength that is easy to answer.** The installer's reading is the conservative one, and
  conservatism in a writer is not a bug. This record answers it with two watchers, and **both are
  conditional**: the hook's next-session message needs the project-scope plugin to load, which
  `decision-0068` records a headless run cannot grant — and headless CI is the case the install path
  exists to serve; the installer's own re-run needs somebody to re-run it. So a repo that renders,
  later gains an `@AGENTS.md` line, is only ever used headlessly, and is never re-installed will
  double-deliver with nobody told. That residue is real and it is the price of D1. It is paid
  against a false positive that was **unconditional** — every run, forever, on a project with no
  Claude-side conflict, carrying a remedy that ungoverns the one host actually reading the block. A
  reader who weighs an unwatched-but-rare wrong delivery above a certain-and-permanent wrong refusal
  should reject D1 and take the installer's reading, under which the fix is to keep the refusal,
  delete D1, and record the divergence as deliberate. **The counter-case has a second prong, and it
  is stronger than the first:** on the nested-import shape D1 is a straight regression in the
  headless environment — one delivery becomes two, with no watcher — which is not a residue but a
  loss. **The two harms D1 creates are not alike, and an earlier version of this sentence folded
  them together in the ruling's own favour** — a reviewer caught it. The first (render, an
  `@AGENTS.md` line arrives later, nobody re-installs) is *sequence*-dependent and needs a run of
  bad luck. The second (nested chain, headless) is *state*-dependent: it fires on the very first
  install of such a repo, deterministically, and nothing has to go wrong for it. So the weighing is
  not luck against luck. It is: the status quo is wrong **unconditionally, on every mixed-host
  repo**, with a remedy that ungoverns Codex; D1 is wrong **deterministically but only on repos
  whose import graph leaves the adapter contract** (`decision-0057` rule 2), plus one unlucky
  sequence. I take D1 on scope, not on likelihood — and a reader who holds that a deterministic
  wrong delivery outranks a universal wrong refusal should reject it. That is a weighing, not a
  proof, and it is the maintainer's to overturn. A third option exists and is not taken here: gate the render on CLAUDE.md carrying no
  `@`-imports at all, which is sound and would refuse for most repos, buying soundness by giving
  back most of what D1 delivers. It is named so the choice is three-way rather than two.
- **`decision-0081`'s cost-of-reversal framing applies, at that record's own stated weight
  ("proposal, not a decision").** This is cheap to reverse: one condition on one `elif`, one NOTE,
  one enumeration comment and one test function. It removes a refusal, which is the direction `decision-0068` D12 calls a
  *simplification* and treats as the harder direction to justify — and D12's own history is why
  this record carries a measurement for every claim rather than an argument from symmetry.
- **`TRL-48` framed the outcome as open** — status quo or change, either a legitimate result. The
  change is what is delivered, and the issue's "Done when" #2 is met in the direction it did not
  presume: the two components agree because the installer moved, not because the hook did.

## Open questions

- **The nested-import chain.** Both components read one anchored line and the host resolves a
  graph. Closing it means a shared resolver, in both scripts, with the recursion limit the host
  actually applies — real work, on both sides of a release boundary, and out of scope here. Filed
  as `TRL-49` (`decision-0075`, `decision-0078` — a named consumer that will re-present it, not a
  note in a file nobody reads), carrying the reproduction recipe and the measurement above.
