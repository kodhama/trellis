---
id: decision-0068
type: decision
status: gated  # drafted by agent; awaiting the maintainer's intent act. Three rulings are recorded, in order. (1) 2026-07-30, direction: "let's focus on claude only for the install path, potentially with the symlink alternative if you think it has merit". (2) 2026-07-30, ROUTING — the agent recommended `/trellis:setup` write the rendered file and said it would not proceed until ruled; the maintainer chose `install.sh`, which is D1, and this frontmatter records the ruling that reversed the agent's own recommendation. (3) 2026-07-30, wording — the maintainer accepted new install-path wording as a `decision-0053` exception; §Wording below records that the exception proved UNNECESSARY on measurement, so the ruling is preserved but not exercised. Three questions remain open for ratification (§Open questions 1-3). The measurements are the agent's; the ratification is not.
depends_on: [decision-0043, decision-0053, decision-0058, decision-0065, decision-0066]
owner: agent
date: 2026-07-30
---

# 0068 — the install path delivers rules through `.claude/rules/`, Claude only

## Context

**The install path delivers no rules at all. This is measured, not inferred.**

`decision-0065` moved rule delivery to the plugin's SessionStart hooks. The
install path was left vendoring the whole bundle into `.claude/skills/trellis/`
and registering nothing — which issue #201 reported and which this record
confirms by experiment.

### The measurement, including the reading that was wrong

Four headless runs in a scratch git repo, asking whether Trellis rules were in
the always-loaded context:

| # | setup | user scope | result |
|---|---|---|---|
| 1 | bundle copied to `.claude/skills/trellis/` — exactly what `install.sh` produces | **included** | TRELLIS=YES |
| 2 | same | **excluded** | **TRELLIS=NO** |
| 3 | nothing installed — baseline | **excluded** | TRELLIS=NO |
| 4 | `.claude/rules/trellis.md` | **excluded** | **TRELLIS=YES** |
| **5** | **the SHIPPED shape: a real file at `.claude/rules/trellis.md` carrying `@../../.trellis/rules.toml`** | **excluded** | **sentinel loaded** |
| **6** | same, with `.trellis/rules.toml` **deleted** — the bare-install state D7 describes | **excluded** | rules body loads; the import contributes nothing, silently, no error |

**Run 1 was read as "the install path works" and that was wrong.** The
user-scope Trellis plugin, installed on the test machine earlier the same day,
governs any repo carrying a `.trellis/rules.toml` — including the scratch one.
**The single-variable control is run 2**, which holds the bundle constant and
removes user scope; run 3 differs from run 1 in two variables at once and an
earlier draft of this record credited it wrongly. Run 4 supplies the positive
control a null result needs. The honest result is **run 2 against run 4**: the
vendored bundle delivers nothing while the rules file delivers, at equal scope.

**Runs 5 and 6 were added after independent review.** An adversary reading this
record observed that every tabulated import measurement was taken in the
*symlink* configuration, which D2 withdraws — so the form this record actually
ships had no recorded evidence. The measurement existed but had never been
written down, which for a reader is the same as not existing. Run 5 records it.
Run 6 answers a gap the same review named: nobody had measured what an import at
a **nonexistent** path does. It degrades silently rather than erroring — which
makes D7's floors-only state real, but reached by silent no-op rather than by
design, and that is worth knowing rather than assuming.

The documented skills-directory-plugin path requires accepting the workspace
trust dialog, which a headless run cannot grant — and headless CI is exactly the
case the install path exists to serve.

### What `.claude/rules/` gives that a hook does not

- Loaded at launch "with the same priority as `.claude/CLAUDE.md`" when the rule
  carries no `paths` frontmatter. No hook, no `settings.json` write, no trust
  dialog, no `.claude-plugin/` manifest.
- **Discovery is symlink-aware, and the link resolves first**: an `@` import
  inside a symlinked rules file resolves relative to the **real file's**
  directory, not the link's. Measured both ways:

  | file at `.claude/rules/trellis.md` | import written | result |
  |---|---|---|
  | symlink → `.trellis/trellis.md` | `@rules.toml` | **loaded** |
  | same symlink | `@../../.trellis/rules.toml` | **NONE** |

- **A `.toml` placed directly in `.claude/rules/` is never loaded.** Only `.md`
  is discovered. Measured: the sentinel was absent across both runs above.

The wrong import form produces **no error and no content**. That is the third
silent-no-op this format has produced during this work and it must be pinned by
a test rather than by care.

## Decision

**1. The install path delivers rules by rendering exactly one file, in PROJECT
SCOPE ONLY.** *(Maintainer, 2026-07-30.)*

```
<repo-root>/.claude/rules/trellis.md   posture prose + rules body + `@../../.trellis/rules.toml`
```

**A `--scope personal` run renders nothing and prints why.** The alternative was
measured unworkable in both readings: `~/.claude/rules/trellis.md` would import
`~/.trellis/rules.toml`, which nothing writes — shipping precisely the
silent-no-op artifact this record exists to prevent — and it would govern **every
repo on the machine**, contradicting `decision-0065:112-114` ("a project that
never adopted Trellis is never governed by surprise"). The git-root reading
contradicts `install.sh:193`, which forbids any git invocation under explicit
personal scope, by design and by its own test.

**The cost is named, not hidden:** personal-scope installs keep delivering no
rules, exactly as today. This record fixes one half of the install path and says
so.

`install.sh` writes that file and nothing else new. It registers no hook, edits
no settings file, creates no symlink, and **never touches `.trellis/`**.

**2. `install.sh` owns `.claude/`; `/trellis:setup` owns `.trellis/`.**

An earlier draft of this record put the real file at `.trellis/trellis.md` and
symlinked it into `.claude/rules/`. **Withdrawn**, for three reasons:

- **Windows.** A symlink narrows the Git Bash slice on top of the POSIX-`sh`
  barrier, and a copy fallback there could not use the sibling import form —
  it would need a different one and would fail silently if got wrong. Without
  the link, #210 is purely about the shell again.
- **Ownership.** `/trellis:setup` writes `.trellis/rules.toml` and nothing else,
  ever. A symlink would place an install-owned object inside setup's directory.
- **Count.** One filesystem object rather than two, for the same delivered bytes.

The cost is a longer import path, and it is cosmetic: **whichever form is
correct, the other is a silent no-op**, so the test in D4 is mandatory either
way. Once it exists, `../../` being uglier costs nothing.

**3. The rows stay live. The rendered file is not a snapshot.**

Because the rows are imported rather than inlined, editing `.trellis/rules.toml`
takes effect in the next session with no re-run of the installer. This matches
the plugin path's behaviour, which re-splices the rows every session.

**4. The import form is load-bearing and is pinned by a test.**

The rendered file carries `@../../.trellis/rules.toml`, correct because the file
is real and lives at `.claude/rules/`. A test asserts that exact emitted form and
asserts the sibling form `@rules.toml` is *not* emitted. Both were measured; each
is correct in one location and silently loads nothing in the other. Without the
test a future reword ships a file that loads nothing and passes every other
check.

**5. Wording: `decision-0053`'s clause does not bind here, and the exception the
maintainer granted is not needed.**

The rendered file cannot carry a posture sentence chosen from `strictness`,
because AC2 forbids reading `.trellis/`. The maintainer accepted new
install-path-only wording as a `decision-0053` exception. **On measurement the
exception is unnecessary**, which is a better outcome and is recorded rather than
quietly taken:

`decision-0053:52-55` protects what the `annotation-vs-absence` experiment
validated — *"a specific authority header, a rows-inlined-below-the-rules layout
… and live-rows seed comments"*. Verified by grep across `eval/`: the authority
header **is** in `annotation-vs-absence/run.sh`; the posture sentence
(*"How strictly … Firmly … By default"*) appears **nowhere in `eval/` at all**.
It entered the payload through `decision-0034`, not through the experiment. So
0053's clause never covered it.

Therefore the rendered file emits **`reference/trellis-b.md`'s posture prose as a
shipped constant, not a choice** — `staleness.sh:141-144` already resolves absent
or unreadable strictness to `b`, so the install path inherits a ratified default
instead of inventing one. The two edits it performs (resolving `@rules.md`, repointing the invariants
path) are both edits `staleness.sh` already performs under `decision-0065`'s "one
edit" allowance.

**One sentence of new prose IS added, deliberately, and it is the only one.**
*(Maintainer, 2026-07-30.)* The frozen posture sentence and the imported
`rules.toml` land in the same always-loaded context and can disagree — a project
whose rows say `strictness = "firm"` would read the adaptive sentence above them.
Rather than resolve that by reading `.trellis/` (which AC2's surviving clause
forbids) the rendered file **states which one is authoritative**: the `strictness`
key in `.trellis/rules.toml`, not the sentence above it. The mismatch is recorded
where a reader will hit it, instead of being left to look like a contradiction.

**6. `decision-0058` phase 4 is satisfied by D8, and this record says so rather
than leaving it inferred.** `decision-0058:123` governs "any Claude hook
replacement" and requires that the old transport be removed or disabled in the
same change "so rules still arrive once"; `:195` sets the boundary — "never while
both paths would inject the full rule payload". D8 is that disablement. An
earlier draft declared `decision-0058` a dependency and never argued it, which is
the defect `decision-0065:163-166` records itself having been caught on.

**7. Claude only, deliberately, and stated rather than implied.**

`.claude/rules/` is a Claude Code mechanism. Codex, OpenCode, Devin and VS Code
have no equivalent that inlines a non-markdown file; the research is recorded in
**#209**. The install path therefore claims **Claude Code only**, and the
existing prose authority sentence at `reference/rules.md:1` remains the fallback
for the rows anywhere that cannot import them.

**8. Platform boundary: macOS, Linux, WSL — and this record adds nothing to it.**

`install.sh` is `#!/bin/sh` with POSIX dependencies, so Windows is already
outside the install path. Because D2 drops the symlink, **this change narrows
nothing further** — #210 stays purely about the shell, and a future PowerShell or
Go installer inherits no symlink problem.

**9. With no `.trellis/rules.toml`, the install is inert except the floors — and
that needs no seeding.**

`install.sh` makes no posture choice and writes no `rules.toml`; that is
`/trellis:setup`'s single file. The rules text already defines this state: the
two `floor-` rows *"apply regardless of their row value"*, and every other rule
applies **only** where its row says `active = true`. So a bare install yields
exactly `floor-transparency` and `floor-intent-gate` — a safe, well-defined
default that requires no seed and no decision from the script. The rendered file
directs the reader to `/trellis:setup` to activate the rest.

This is why `decision-0065`'s clause that setup writes "exactly one file …
and nothing else, ever" **needs no amendment**: this record does not widen it.

**10. The plugin hook stands down when the install artifact is present. This is
required, not optional — the alternative was measured and it double-delivers.**

With the user-scope plugin installed *and* `.claude/rules/trellis.md` present,
the rules arrive **twice**: once in the project-instructions block, once in the
hook's `additionalContext`. Measured, not predicted.

`staleness.sh` already has this shape — it detects `.trellis/internal/` and
declines to inject over a vendored overlay, so that "a project that vendored the
overlay never receives the rules twice". The install path needs the same branch
keyed on its own artifact: **if `.claude/rules/trellis.md` exists, the hook
injects nothing** and says so, exactly as it does for the overlay today.

An earlier draft claimed this came free by reusing the `.trellis/internal/`
path-A discriminator. **That was wrong** — it held only while invariants were
also placed there, which this record no longer does. Path A never fires under
this design, so the branch is new code, roughly three lines, and explicit.

This widens the change beyond the installer, and that is stated rather than
discovered at implementation: **`decision-0065`'s hook gains a third path.**

**11. `/trellis:remove` enumerates the rendered file, in this change.**
*(Maintainer, 2026-07-30.)*

`plugins/trellis/skills/remove/SKILL.md` deletes `.trellis/` and does not know
about `.claude/rules/trellis.md`. Unchanged, it would delete the rows while
leaving the governing file behind — a posture header, a full rules body and a
dangling import, still in always-loaded context, with nothing left to activate
any rule. `spec-0005` §1 names that standard itself: *"a vendored install with no
`/trellis:remove` is a governance tool with no clean exit, which spec-0004
already treats as a trust defect."*

Ordering matters: the rendered file is removed **before** `.trellis/`, so an
interrupted removal never leaves the governing file without its rows.

**12. The installer reads static-delivery state, and that widens AC2 a second
time. Flagged, not assumed.** *(Added 2026-07-30 after review; NOT covered by the
D1 ruling.)*

`install.sh` refuses to render when the project already delivers the rules
statically — `.trellis/internal/`, the legacy flat `.trellis/trellis.md`, or a
`trellis:begin` block in `CLAUDE.md`/`AGENTS.md`. There is no alternative: **both
static chains are loaded by the host before any hook runs**, so D10's stand-down
cannot reach them, and rendering blindly ships guaranteed double delivery.

The read is existence-only, on files this script never writes, and selects
nothing — the only outcome is render or refuse-loudly. No posture, no style, no
marker patching. But it is state-dependent behaviour in a script whose spec says
"zero decision logic", so it is an amendment and is recorded as one.

**13. Out of scope, named rather than silently retained: whether the bundle
vendoring stays.** `install.sh` also vendors the whole `plugins/trellis/` tree so
`/trellis:setup` and `/trellis:remove` have a home (`spec-0005` §1). Whether
those skills actually load from that location is **unmeasured** — this record
measured rule delivery only, and does not touch the bundle.

## Consequences

- **A path that shipped nothing now ships something.** That is the whole value;
  everything else is simplification.
- `install.sh` no longer needs to reason about hooks or host manifests for rule
  delivery. The hook files remain bundle bytes under `spec-0005` §1 until D13 is
  answered — and D10 keeps the plugin's copy of them from firing here.
- **The install path vendors, and that is correct here.** `decision-0065`'s split
  is "plugin configures, install.sh vendors". Staleness is inherent to vendoring
  and is what the version stamp exists for. It was only the *plugin* path where
  writing into a consumer repo was the error.
- **No new artifact class.** An earlier draft introduced a committed symlink; D2
  withdrew it, so the install path writes only ordinary files.
- **Invariants stay a vendored file** at `.claude/skills/trellis/reference/invariants.md`,
  and the rendered pointer names that path. That is already better than today,
  where the shipped prose names `.trellis/internal/invariants.md` and nothing on
  the install path creates it. A later record is expected to replace the pointer
  with a skill on both paths; it is **deliberately not cited here** — it is a
  gated draft on an unmerged branch, and `decision-0065:209-220` sets the
  precedent that building on a draft is what `inv-directional-flow` forbids, so
  the dependency is removed rather than dressed up.
- **Cost:** two delivery mechanisms now exist for Claude — hook injection via the
  plugin, file discovery via the install. D10 keeps them mutually exclusive; without
  it a repo with both receives the rules twice, which is measured, not feared.
  `.trellis/rules.toml` cannot be the discriminator because both paths require it,
  so the install artifact itself is the signal.

## Open questions

**Numbering note:** an earlier draft ran 1, 3 — a question was deleted rather than
struck when it became D10, in a record whose self-check claims withdrawn material
is struck. Renumbered and recorded.

**The three ratification acts are RULED (2026-07-30) and have moved into
Decisions — D1 (project scope only), D5 (record the posture mismatch), D11
(`/trellis:remove`). They are struck here rather than deleted.**

1. ~~**Personal scope.**~~ **RULED: project scope only — D1.** `spec-0005:117` and AC4 make `~/.claude/skills/trellis/` a
   supported target; this record measured **project scope only**. Neither reading
   works: `~/.claude/rules/trellis.md` would govern **every repo on the machine**
   and import `~/.trellis/rules.toml`, which nothing writes — shipping exactly the
   silent-no-op artifact this record exists to prevent, and contradicting
   `decision-0065:112-114` ("a project that never adopted Trellis is never
   governed by surprise"). The git-root reading contradicts `install.sh:193`
   ("explicit personal scope: no git invocation at all, by design") and its own
   test. **Ruling needed:** project-scope-only, or something else.

2. ~~**The posture is a snapshot; the rows are live.**~~ **RULED: ship the adaptive constant and state which source is authoritative — D5.** D5 emits `trellis-b`'s
   adaptive sentence as a constant. A firm project installed by curl would read
   *"**By default** — follow them unless you have a clear, specific reason not
   to"* and, a few lines below in the same always-loaded context, its imported
   `strictness = "firm"`. The three resolutions are (a) read `rules.toml` —
   forbidden by AC2's surviving clause; (b) new read-time posture prose — the
   thing D5 just established is unnecessary; (c) accept the mismatch and record
   it. **All three are ratification acts.**

3. ~~**`/trellis:remove`.**~~ **RULED: fixed in this change — D11.** It deletes `.trellis/` and does not enumerate
   `.claude/rules/trellis.md`, so it would leave a governing file with a dangling
   import in always-loaded context while deleting the rows it imports.
   `spec-0005` §1 names the standard itself: "a vendored install with no
   `/trellis:remove` is a governance tool with no clean exit, which spec-0004
   already treats as a trust defect." **Parking this makes the product worse than
   today**, so it is a ruling rather than a follow-up.

4. ~~**The rendered file has no staleness surface.**~~ **CLOSED IN THIS PR** —
   implemented rather than left open, after an independent reviewer found the
   record proposing a fix it had not taken. `install.sh` now embeds
   `<!-- trellis:rendered-from payload@<stamp> -->` as the rendered file's last
   line, and path C compares it against the installed plugin's
   `reference/version`, emitting a nudge naming **both** stamps on mismatch while
   still injecting nothing. The original text is kept below because the reasoning
   is what produced the fix.

   ~~This record's claim that "the version stamp exists for" it is unsupported.~~ `decision-0043` rule 3 pins the
   compare to `.trellis/version`; D1 forbids writing there, and D10 makes the hook
   stand down rather than compare. So the new vendored artifact has zero drift
   visibility in either configuration — with the plugin (D10 silences it) or
   without (no hook runs). `decision-0035`'s floor, "drift is made visible, not
   silent", has no subject here. A cheap fix exists: make D10's branch **nudge on
   a stamp mismatch** rather than stand down silently.

5. **Do the vendored skills load at all?** D13 defers it; the same trust-dialog
   question that defeated hook registration may defeat them. §4 item 5's whole
   point is "run `/trellis:setup`" — a vendored skill — so if they do not load,
   the rendered file instructs the reader to run something absent and the install
   stays permanently at two floors.

6. **Does `VERSION` bump?** `plugin_package_test.go:220` hard-pins `"0.3.0\n"`.
   This changes plugin behaviour that marketplace consumers only receive by
   re-pulling, and the immediately preceding commit bumped 0.2.0 -> 0.3.0 for
   exactly that reason. Unstated means the executor decides silently.

## Self-check (gate)

Re-derived after four independent reviews (two decision-adversary, one
conformance, one cold implementation plan). **Three rows that an earlier draft
graded PASS were false.** They are corrected here rather than quietly restated.

| # | check | result |
|---|---|---|
| 1 | The central claim is measured, not read from docs | **PASS** — six runs, tabulated, with the confound named |
| 2 | A wrong reading is recorded rather than quietly corrected | **PASS** — run 1 and the control that overturned it, plus the control attribution itself corrected from run 3 to run 2 |
| 3 | The **shipped** import form is measured, not just the withdrawn one | **WAS FALSE, NOW PASS** — an earlier draft tabulated only the symlink configuration while shipping a different one. The measurement existed and had never been written down, which for a reader is the same as not existing. Run 5 |
| 4 | Non-Claude harnesses checked before claiming Claude-only | **PASS** — #209, four harnesses, sources cited there |
| 5 | The platform boundary is stated with its real cause | **PASS** — POSIX `sh`, not the symlink; #210 |
| 6 | Scope creep resisted and named | **PASS** — D13 leaves the bundle untouched and says so |
| 7 | Every declared dependency is argued, not just listed | **WAS FALSE, NOW PASS** — `decision-0053` and `decision-0058` were in `depends_on` and appeared nowhere in the body. Now D5 and D6 |
| 8 | No gated draft is cited as settled ground | **WAS FALSE, NOW PASS** — an earlier draft leaned on an unmerged `decision-0067`; removed per `decision-0065:209-220` |
| 9 | The maintainer's rulings are in the record, including one that reversed the author | **PASS** — frontmatter carries all three, and marks the wording exception as granted-but-unnecessary |
| 10 | Double-delivery risk surfaced | **PASS** — measured rather than deferred, so it became D10 |
| 11 | A withdrawn design is struck, not deleted | **PARTIAL** — D2 and D10 record their withdrawals, but an open question was deleted rather than struck; corrected with a note rather than reconstructed |
| 12 | Acceptance criteria | **DEFERRED to the paired `spec-0005` amendment**, which is on this branch and itself returned FAIL on first review |
| 13 | The three blocking ratification acts are ruled | **PASS** — ruled 2026-07-30 and folded into D1, D5 and D11; struck in §Open questions rather than deleted |
| 14 | The record matches the code shipped beside it | **WAS FALSE, NOW PASS** — Open 4 declared "no staleness surface" and proposed a fix that this same PR then implemented, leaving the record contradicting its own branch. Found by review, not by me. `inv-graph-maintenance` applies to a record and its own implementation, not only to downstream artifacts |
| 15 | `status: gated` earned | **PASS** — self-check run after four independent reviews; three false rows corrected, one partial recorded, the blocking questions ruled. What remains open (drift surface, vendored skills, VERSION) is named and none of it blocks implementation |
