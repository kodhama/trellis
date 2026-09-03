---
id: decision-0088
type: decision
depends_on: [decision-0037, decision-0047, decision-0068, decision-0070, decision-0079]  # decision-0070 and decision-0079 moved up from informed_by after an independent corpus review: both are coupling under decision-0047's test, not provenance. D2's "still seeded rules-b.toml, so seed and header still agree" is FALSE, not merely less evidenced, if decision-0070 D2 stopped seeding or seeded rules-a.toml; and the Context argument that spec-0005's prohibition is now only a test rests on decision-0079 having retired the spec stage. Same correction decision-0076 made for the same reason. decision-0037 and decision-0047 added in review of #265: the Self-check characterises this record's own owner: agent by 0037's ruling, and this frontmatter's shape is decided by 0047's coupling test -- both contingent, not provenance, and both declared as depends_on everywhere else in the corpus
changes: [decision-0068]
informed_by: [decision-0028, decision-0035, decision-0053, decision-0078, decision-0081, decision-0082]
owner: agent
date: 2026-09-03
---

# 0088 — the install path renders the posture the project's own rows ask for

## Context

`decision-0068` D5 ruled that the rendered `.claude/rules/trellis.md` emits
`reference/trellis-b.md`'s adaptive posture prose **as a shipped constant, not a choice**. Its
Open question 2 named three resolutions — read `rules.toml`, write new read-time prose, or accept
the mismatch — and ruled the third, adding one sentence of footer prose saying the rows win over
the frozen sentence.

**The consequence it predicted has now been observed in the field.** `TRL-35`, probing Trellis
delivery in a Claude Code cloud session, hit it on *this repository*: `.trellis/rules.toml` here
says `strictness = "firm"`, and the rendered file opened with *"**By default** — follow them unless
you have a clear, specific reason not to."* The plugin's own `SessionStart` hook serves the **firm**
header to that same project (`plugins/trellis/hooks/staleness.sh`, path B). So the two deliveries
disagreed about how strictly a project's rules are to be followed, and which one a session got
depended only on which delivery path it was on. `TRL-37` filed it as a defect.

### Why D5's stated ground no longer holds

D5 does not argue that a constant is *better*. It argues, twice and explicitly, that the choice is
**unavailable**:

> The rendered file cannot carry a posture sentence chosen from `strictness`, because AC2 forbids
> reading `.trellis/`.

That ground has since been removed from under it, in two steps neither of which was about posture:

1. **`decision-0070` D2 already has `install.sh` write `.trellis/rules.toml`** — seeding the rows
   when none exist. `decision-0068`'s "never touches `.trellis/`" clause is marked
   `superseded_in_part_by: [decision-0070]` in its own frontmatter. A script that may *write* the
   file to make the rules apply is not one that may not *read* one key of it to render them
   correctly.
2. **AC2's "never reads" clause was already amended inside `decision-0068` itself** (D12), which
   permits one content read — the managed-block marker — and records the count as an amendment
   rather than a silent widening. And `decision-0079` retired the spec stage: `spec-0005`'s only
   surviving statement is `cli/install_script_test.go`. The constraint is now a **test**, which is
   changed by argument in the same commit, not a ratified document that outranks the change.

What survives from D5 untouched is its *other* half, and this record keeps it: the rendered sentence
is frozen at install time while the rows are read fresh every session, so the footer must go on
saying the rows are authoritative.

## Decision

1. **When `.trellis/rules.toml` exists and is readable, the render selects its header from that
   file's `strictness`, exactly as `staleness.sh` path B does:** `firm` → `reference/trellis-a.md`,
   every other value → `reference/trellis-b.md`. Head and tail change source; body, footer, import
   line and stamp are unchanged. The head is the shipped payload's own bytes, not new prose —
   `decision-0053`'s rule that the tested wording is the shipped wording holds on this path too.
2. **Every other readable input resolves to `trellis-b.md`** — no file, no key, or an unrecognised
   value — which is the hook's own fall-through, not a new default invented here. A project with no
   `rules.toml` is still seeded `rules-b.toml`, so seed and header still agree. A file that cannot be
   read also resolves to `trellis-b.md`, but that is **not** what the hook does; D3 records the
   divergence rather than claiming parity.
3. **The unreadable case is said out loud, and it is a recorded divergence from the hook**
   (`decision-0035`, `floor-transparency`). `[ -f ]` is true for a file the invoking user cannot
   read. The installer renders the adaptive header, reports that it could not honour the file's
   posture, and does not abort — the first version of this change did abort there, under `set -eu`
   inside a command substitution, leaving a vendored bundle with no rules file, no seed and only the
   shell's `cannot open` as diagnosis. The hook does **not** serve adaptive on that input: its
   strictness `awk` comes back empty, but its row validator then produces no usable report and
   `staleness.sh` emits `TRELLIS_RULES_NOT_LOADED` and injects nothing — its fail-closed branch
   above the reconciler. So on an unreadable file the two deliveries disagree: installer adaptive
   and announced, hook nothing and announced. And once the rendered file exists the hook stands down
   to it (path C), so the rendered adaptive file is what governs. The fail-closed alternative — the
   installer refusing to render over a file it cannot read — was available and is not taken here: an
   install that stops halfway is the worse state, and the install is the one moment someone reads
   the output. It is the same shape as the `governed = false` residual under Consequences, and it
   is noted on `TRL-38` beside it.
4. **The parser is the hook's, copied byte for byte, with a guard pinning the pair**
   (`decision-0028`). The hook ships *inside* the bundle `install.sh` vendors, so the two cannot
   share a file; a copy drifts unless something fails when it does.
5. **The content-read budget goes from one to two, and the second is named** rather than absorbed.
   The test that pins the count said a second read "has to come here and argue itself"; this record
   and that test are the argument.

## Consequences

- **`decision-0068` is changed in part, not superseded.** What is corrected: D5's constant-header
  ruling, the read inventory it states, and D9's clause that *"`install.sh` makes no posture
  choice"* — its other clause, *"writes no `rules.toml`"*, was already changed by `decision-0070`
  D2. Its delivery mechanism, project-scope-only ruling (D1), the two payload edits, the stamp, D11
  and the footer sentence all stand. The forward pointer on `0068` records the scope.
- **The two deliveries now agree on the posture header of any *governed* project whose `rules.toml`
  is readable**, and a test proves it by running both against the same repository rather than
  asserting each side against a fixture. The qualifiers are load-bearing, not hedging: on a
  `governed = false` project they still disagree, for the separate reason in the last bullet; on an
  unreadable file they disagree as D3 records; and the test exercises governed, readable fixtures
  only.
- **What is *not* claimed here:** that `install.sh` may read `.trellis/` generally, or branch on any
  other project state. One key, for one purpose, counted and guarded. A third read still has to come
  and argue itself.
- **Not fixed, and tracked rather than parked** (`decision-0078`, `inv-no-orphan-followups`): a
  `rules.toml` holding only the `governed = false` opt-out still gets a governing rules file
  rendered into it, after which the hook stands down and the rendered file governs a project that
  switched Trellis off. That is pre-existing behaviour, older than this change and outside `TRL-37`;
  it is filed as `TRL-38`, which is the consumer that will re-present it, and the unreadable-file
  divergence in D3 is noted there beside it as the same shape. The installer's own output no longer
  claims hook parity on either input.

## Self-check

- **The maintainer accepted this record on 2026-09-03, on `TRL-37`; his merge is the act**
  (`decision-0082`). It is authored by the agent that wrote the code change (`decision-0037` —
  authorship, not accountability), and it reverses a ruling he gave on 2026-07-30. It was written to
  be rejectable — reverting is one `case` statement and the record retires — and the D3 wording was
  corrected in review before merge, after an automated reviewer showed the hook-parity claim for
  unreadable files was false.
- **`decision-0081`'s cost-of-reversal framing applies and is cited at its own stated weight** — that
  record calls itself *"(proposal, not a decision)"*. Under it this is cheap to reverse, which is the
  argument for the agent proceeding rather than blocking; it is not authority to have decided.
- **The defect this corrects was measured, not reasoned** — on this repository, by a session that was
  probing something else. The prediction was in `0068`'s own Open question 2 the whole time; what was
  missing was that nobody had read the rendered file next to the rows it imports.
- **`corpus-reviewer` was run on this record and found one defect in it**, now fixed: `decision-0070`
  sat in `informed_by` when D2's seed-and-header-agree claim is contingent on `0070` D2's seeding
  clause — coupling under `decision-0047`, not provenance. `decision-0079` moved up with it on the
  same test. A second pass, in review of #265, found four more edge defects under the same test —
  `0037` and `0047` undeclared though this record's own frontmatter and `owner` rest on them, `0078`
  undeclared though the TRL-38 filing rests on its address test, and `0053` declared but unargued —
  all fixed in the same review round. The one violation the review reports elsewhere in the corpus
  (`decision-0044`'s bare `kodhama-0004-uniform-lifecycle`) predates this change and is left alone:
  that record preserves it deliberately as the live instance of the gap it fixes.
- **Written because a review bot insisted on it.** The code change was opened with this recorded only
  as a note offering the record as a possible follow-up. An automated reviewer on
  `kodhama/trellis#261` called that what it is — executable behaviour contradicting an accepted
  decision on `main` — and `inv-graph-maintenance` says fix the decision, not patch around it. The
  agent had already identified the tension and chose to defer it; being right about the tension and
  wrong about deferring it is the part worth recording.
