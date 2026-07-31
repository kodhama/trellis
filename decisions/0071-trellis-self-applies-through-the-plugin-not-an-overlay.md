---
id: decision-0071
type: decision
status: gated  # drafted by agent; awaiting the maintainer's intent act. Requested in session 2026-07-31 ("yes, retire it") after tracing #219's step 1 and finding it retires a ratified property rather than performing a cleanup.
depends_on: [decision-0035, decision-0043, decision-0051, decision-0065, decision-0070, kodhama/kodhama-0007-one-render-many-copiers]
owner: agent
date: 2026-07-31
---

# 0071 — Trellis self-applies through the plugin, like every other consumer

## Context

This repository governs itself with a **vendored overlay** — `.trellis/internal/`
plus a managed block in `CLAUDE.md` — and that is the shape `decision-0065`
retired for consumers, `decision-0068` refuses to render over, and
`decision-0070` treats as a legacy migration case. Trellis is the last project
running the delivery mode Trellis tells everyone else to leave.

**It is not an oversight; it is pinned.** `cli/selfapply_test.go` enforces it as a
ratified property:

> *"the self-application sync-guard (`decision-0035`) … the two guards together
> give generator == vendored payload == repo overlay: drift stays impossible, and
> the repo's own overlay is exactly a mechanical copy of the shipped artifact
> (`kodhama-0007` rule 2)."*

`TestSharedProjectInstructionEntrypoints` additionally requires `CLAUDE.md` to be
*exactly* `@AGENTS.md`, a blank line, and the byte-identical `block-claude.md`
payload — `spec-0006`, still `gated`.

**Why this surfaced now.** `#219` proposes retiring `/trellis:setup`, whose only
non-trivial remaining job is the overlay→plugin migration (§3). That migration
exists for exactly one project: this one. So the skill cannot go while the
overlay stays, and the overlay cannot go while `decision-0035`'s sync-guard binds
it. `#219` recorded step 1 as "migrate this repo's own overlay by hand" — a
sentence that, traced, retires a ratified guarantee.

## Decision

**1. This repository stops self-applying through a vendored overlay.**
`.trellis/internal/` and the `CLAUDE.md` managed block are removed. `.trellis/rules.toml`
stays: it is the opt-in signal (`decision-0065`) and the rows (`decision-0070`).

**2. Self-application continues — through the plugin path, which is what
consumers get.** The claim `decision-0035` protects is that Trellis is subject to
its own rules and cannot drift from its shipped payload. That claim survives, and
gets *stronger*: the repo now consumes the same `SessionStart` delivery every
consumer does, rather than a copy that had to be proved identical by test.

**3. What the sync-guard was protecting is retained by other means, and this is
the part that must not be lost in the deletion.** The overlay guard gave three
properties. Two are preserved by guards that already exist:

| property | was pinned by | now pinned by |
|---|---|---|
| vendored payload == generator render | `TestVendoredPayloadIsCurrent` | unchanged — still enforced |
| the shipped bundle is byte-exact in a consumer | `install.sh`'s manifest + `assertBundleVendored` | unchanged |
| repo overlay == shipped payload | `TestRepoOverlayIsCurrent` | **retired with the overlay** |

The third has nothing left to compare, because the artifact it compared is gone.
`decision-0035`'s sync-guard is therefore **superseded in part**: the drift it
prevented is now structurally impossible rather than tested, since the repo reads
the plugin's payload directly instead of holding a copy.

**4. `spec-0006`'s `CLAUDE.md` shape is superseded for this repo.** `CLAUDE.md`
becomes `@AGENTS.md` alone. The spec is `gated` and its entrypoint contract was
written for the managed-block era.

## Consequences

- `decision-0035` gains a `superseded_in_part_by` pointer scoped to the
  repo-overlay sync-guard; its self-application principle stands.
- `TestRepoOverlayIsCurrent` is deleted rather than adjusted — there is no
  overlay to compare. `TestSharedProjectInstructionEntrypoints` is narrowed to
  what still holds.
- **This unblocks `#219`.** With no overlay left in the family, `/trellis:setup`
  §3's migration has no population, and the skill's remaining job is a one-line
  posture change.
- This repo becomes a live test of `decision-0070` D3/D4: it now has
  `.trellis/rules.toml` and no overlay, which is the ordinary config-only shape.

## Open questions

1. **Does anything still prove this repo runs its own rules?** After this, the
   answer is "the same thing that proves it for a consumer" — the hook, on every
   session. That is weaker as a *test* and stronger as a *fact*. Whether a CI
   assertion should replace `TestRepoOverlayIsCurrent` is not resolved here;
   naming it so the deletion is not mistaken for coverage that was never needed.
