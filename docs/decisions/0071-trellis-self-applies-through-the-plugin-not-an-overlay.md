---
id: decision-0071
type: decision
status: approved  # maintainer intent act 2026-08-02, in session: "I would say approve them" — an in-PR flip recording that act, per grove/charters/lifecycle.md:61 ("an in-PR flip recording that act is legitimate"). Author (agent) != approver (maintainer). Content re-read before the flip, not flipped on age. Requested in session 2026-07-31 ("yes, retire it") after tracing #219's step 1 and finding it retires a ratified property rather than performing a cleanup.
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
payload — `spec-0006`, which the same 2026-08-02 act ratifies alongside this record.

**Why this surfaced now.** `#219` proposes retiring `/trellis:setup`, whose only
non-trivial remaining job is the overlay→plugin migration (§3). That migration
exists for exactly one project: this one. So the skill cannot go while the
overlay stays, and the overlay cannot go while `decision-0035`'s sync-guard binds
it. `#219` recorded step 1 as "migrate this repo's own overlay by hand" — a
sentence that, traced, retires a ratified guarantee.

## Decision

**1. This repository stops self-applying through a vendored overlay, on BOTH
hosts.** `.trellis/internal/`, the `CLAUDE.md` managed block, **and the
`AGENTS.md` Codex bootstrap block** are removed. `.trellis/rules.toml` stays: it
is the opt-in signal (`decision-0065`) and the rows (`decision-0070`).

*(An earlier draft removed only the Claude half. The Codex bootstrap is the
AGENTS.md counterpart of the same overlay: `block-codex.md:17` says "if valid
activation TOML is present but the boundary is absent, read only the three
`.trellis/internal/` files" — so deleting those files while leaving the block
would have made every Codex session here report "Trellis was not loaded" instead
of self-applying. Found by the Codex reviewer. Retiring one transport and leaving
its fallback pointed at the deleted inputs is worse than retiring neither.)*

**What that costs, named rather than discovered:** the bootstrap existed as a
belt-and-braces fallback for when the native Codex hook does not fire. This repo
now has none — on Codex it is governed by the hook or not at all. That is the
same exposure every consumer already has (`decision-0065`: the plugin path is
the delivery), and accepting it here is the point of self-applying like a
consumer. It is a real reduction in this repo's own safety net, and `trellis#220`
is where a proper Codex path returns.

**2. Self-application continues — through the plugin path, which is what
consumers get — and this repo now DECLARES that dependency.** `trellis@kodhama`
joins `grove@kodhama` in `.claude/settings.json`, at project scope, for the same
reason grove is there: so every clone, contributor and CI agent gets it rather
than relying on whatever happens to be installed on one machine.

*(A reviewer read the absence of that entry as proof the repo consumed nothing,
and I briefly agreed — wrongly. `trellis@kodhama` is enabled at USER scope in
`~/.claude/settings.json`, so the plugin does reach this repo, and with
`.trellis/rules.toml` present `decision-0070`'s path B delivers with no scope
check at all. The finding was right about the DECLARATION and wrong about the
delivery: undeclared, self-application worked here and would not survive a fresh
clone.)*

**What this does not give: dogfooding HEAD.** A marketplace-installed plugin is
the last published release, not the working tree — so this repo now runs the
shipped Trellis rather than the one in `plugins/trellis/`. The overlay's one real
advantage was that `TestRepoOverlayIsCurrent` forced them to be identical.
Recorded as a known reduction, not a discovery; pointing a local session at the
worktree is a development-time concern, not a property this record should assert.

**The consumer path, restated:** The claim `decision-0035` protects is that Trellis is subject to
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
becomes `@AGENTS.md` alone. The spec's entrypoint contract was written for the
managed-block era; approving it on 2026-08-02 approves that contract, not the
managed-block shape this ruling retires for this repo.

**5. This repository is NOT governed on Codex until `#220`, and that is accepted
here rather than left to be found.** `.claude/settings.json` is Claude-specific,
so declaring the plugin at project scope installs nothing for Codex; `README.md`
says plainly that Codex has no plugin installation channel; and D1 removes the
`AGENTS.md` bootstrap that was its only other route. A fresh clone opened with
Codex therefore has neither the hook nor a readable payload.

That is a real reduction, not a technicality. It is accepted because the
alternative is worse: keeping the bootstrap means keeping `.trellis/internal/`
for it to read, which is the overlay this decision exists to retire, and
retaining a whole delivery mode for one repository's second host inverts the
"self-apply like a consumer" principle the record is built on. A Codex consumer
today has no install path either — that is `#220`, the same gap, not a special
one this repo should paper over for itself.

**This supersedes `decision-0058` point 4 in part** — its requirement that the
Codex receipt/fallback be retained — for this repository only. The phased
host-expansion principle stands; what lapses is the fallback's existence here,
and only while Codex has no channel.

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
