---
id: decision-0093
type: decision
depends_on: [decision-0065, decision-0051]
informed_by: [decision-0028, decision-0053, decision-0071, decision-0072, decision-0082]
superseded_in_part_by: [decision-0094]  # 2026-09-06 decision-0094 — TWO clauses, both in Consequences, and no part of the ruling. (1) The closing sentence of the bullet on existingFile: "No caller here can act on the difference between missing and unreadable." Overtaken by #294 in the same week it was written, which added the first caller that DOES act on it — it reports the result to a person, and which fault it is decides which remedy that person needs. The rest of that bullet stands as a description of existingFile, which keeps the flattened contract for its one remaining caller (the overlay-presence half of D2's condition). (2) The first Consequences bullet, "A Codex session on an incomplete overlay is handed a pointer that resolves, instead of one that does not" — true when written, and now false in exactly four cells: an incomplete overlay beside a plugin copy that is zero-byte, newline-only, NUL-filled or mode-0000 no longer takes the fallback, so it keeps its own dead token. That is 0094's D4 and is disclosed there; the clause is marked here so a reader of 0093 alone does not meet it as current truth. Its second sentence stands — a complete overlay is still unaffected. THE RULING IS UNTOUCHED: D1's consulted-vs-delivered distinction is what 0094 obeys, and D2's fallback condition is applied rather than amended — 0094 reads "the plugin's copy exists" as the usability test D2 itself says it is ("stops the fallback replacing one dead pointer with another"), which a bare stat does not deliver on an unusable copy.
owner: agent
date: 2026-09-06
---

> **Provenance.** TRL-58, filed from a cold review of `#283` (TRL-52). That PR closed the
> invariants pointer on the **plugin-native** Codex branch and deliberately left the vendored
> branch alone, arguing the overlay carries its own authoritative `invariants.md`. The review
> reproduced the shape where it does not. This record rules on that shape and on the README
> sentence the same review found false.

# A consulted pointer may fall back; a delivered payload may not

## Context

`codex-context.mjs` chooses its payload source from the **presence of the
`.trellis/internal/` directory**, and states its own doctrine at `:907-912`:

> Present -> vendored: every file within is required, and a missing one is a broken overlay that
> must fail loudly rather than silently falling through to the plugin's payload. Absent ->
> plugin-native. Using a file as the discriminator would turn a half-deleted overlay into a
> silent mode switch.

`VENDORED_PAYLOAD` names three required files — `trellis.md`, `rules.md`, `version`. It does not
name `invariants.md`, and nothing else on that path checks for it. So an overlay carrying only
those three injects a context whose last line tells the model to read
`` `.trellis/internal/invariants.md` `` — **TRL-52's exact defect, surviving on the arm `#283` did
not change.**

That shape is not hypothetical. The overlay's writer was the retired setup skill, and a
**skill is model-executed instructions rather than a script** (`2f6a21a:plugins/trellis/skills/setup/SKILL.md:121`
carried the `cp`). A session that skipped the copy leaves exactly this tree, and `decision-0071`
retired the overlay without ever auditing the ones already on disk. Nothing creates an overlay
today: `install.sh:529` only *detects* `.trellis/internal/` and refuses over it as a legacy
static-delivery conflict. **The population is historical and unaudited — which is precisely why it
cannot be assumed complete.**

Two constraints appear to forbid a fix, and the ruling below is that neither reaches this case.

## Decision

**A file that is *delivered* and a file that is *consulted* are not governed by the same rule.**

1. **The `:907-912` doctrine governs delivered payload.** Its concern is a *mode switch* — an
   overlay silently falling through to the plugin's copy for the injected chain, so the model is
   governed by rules the project did not vendor. That is why the three files are required and why
   the *directory*, not a file, is the discriminator. `invariants.md` is **consulted**: read on
   demand, when a rule seems ambiguous, and never injected. A session that meets no ambiguous rule
   never opens it. Failing such a session closed would trade a dead pointer for **no governance at
   all**, which is strictly worse on the axis the doctrine exists to protect.

2. **So the pointer falls back and the payload does not.** On the vendored branch, when the
   overlay carries no `invariants.md` **and** the plugin's copy exists, the consulted pointer is
   repointed at the plugin's copy. `sources` is untouched: prose, rules and version still come
   from the overlay. **Only the pointer moves.** Both halves of that condition are load-bearing —
   the second stops the fallback replacing one dead pointer with another.

3. **`README.md:27` overstated the symmetry and is corrected.** It said that where an overlay
   exists *"the hooks detect it, read from it, and inject nothing"* — **on both hosts**. Measured:
   `staleness.sh` emits **zero bytes** on a vendored project; `codex-context.mjs` reads the
   overlay and injects what it finds. The clause that follows — *"the plugin's `reference/` files
   stay installation sources rather than runtime substitutes"* — stands, and is not in tension
   with (2): substitution presupposes something to substitute **for**, and this arm runs only when
   there is not.

4. **The same claim in `decision-0065:191-193` is marked, not just the README's copy.** That
   paragraph states *"A project that still has an overlay keeps its import transport and the hooks
   inject nothing"* — false on Codex for the same measured reason, and it is the record this one
   depends on. `decision-0065` now carries `decision-0093` in `superseded_in_part_by` for that
   clause alone. Its **discriminator** half stands (both hooks do discriminate on
   `.trellis/internal/`), as does *"a project without one has no import block and receives the
   injection"*. Worth recording why this was missed: the neighbouring `:194-195` sentence in the
   same paragraph was re-pointed by `decision-0068` while this one was not, so the paragraph
   already looked marked. **The claim was wrong in two places and marked in neither** — correcting
   the README while leaving the record it derives from would have preserved exactly that.

**What is not decided here:** whether the vendored-overlay path should exist at all. That is
`decision-0072`'s open question and TRL-31's, and it is untouched — this record makes the path
honest, not permanent.

## Consequences

- A Codex session on an incomplete overlay is handed a pointer that resolves, instead of one that
  does not. A complete overlay is unaffected: its own `invariants.md` still wins.
- `TestCodexRepointsWhenAVendoredOverlayLacksInvariants` pins the new arm, and asserts the
  delivery source did **not** move — the guard against this becoming the mode switch (1) forbids.
  `TestCodexLeavesAVendoredInvariantsPointerAlone` pins the complete-overlay side; a mutation
  inverting the guard fails **both**, so the condition is pinned in both directions.
- The fixture asserts `writeValidCodexOverlay` writes no `invariants.md`, so this test cannot
  quietly decay into a duplicate of its sibling if that helper changes.
- `existingFile` joins `existingDirectory` with a deliberately identical contract: absolute paths
  only, every stat failure reading as absent. No caller here can act on the difference between
  missing and unreadable.
- A payload change, so a release: `VERSION` 0.14.0 → 0.15.0, both `plugin.json` manifests, and
  `install.sh`'s baked `TRELLIS_BUNDLE_MANIFEST` (`decision-0028`).
- `decision-0053` is not engaged: the shipped *prose* is unchanged. The token still ships as
  written; what changed is which address one channel resolves it to — the same class of edit
  `decision-0065:106-111` already ratified for the plugin-native mode.

## Self-check (gate)

- **Smaller thing that works?** Yes, and two smaller ones were rejected with reasons. *Add
  `invariants.md` to `VENDORED_PAYLOAD`* would fail every incomplete overlay closed — the
  disproportion (1) names. *Leave it* keeps a shipped line telling the model to read a file that
  is not there.
- **Grounded in an artifact?** The fix is code plus two guards, not prose about a rule.
- **Reversible?** One `else if`. Deleting it restores today's behaviour exactly.
- **Intent gate.** Authored by an agent; the merge is the maintainer's act.
