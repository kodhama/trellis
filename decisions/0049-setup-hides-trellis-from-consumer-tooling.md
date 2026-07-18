---
id: decision-0049
type: decision
status: approved  # maintainer's intent act 2026-07-18 ("/remote-control: approved 161") — in-PR flip recording the act (decision-0046); independent conformance check (corpus-reviewer) passed before the gate
depends_on: [invariants-v1, signature-catalog-v1]
informed_by: [spec-0005, decision-0043]
owner: agent
date: 2026-07-18
---

> **Provenance.** Lifted from grove's shaping draft `adr-0014`
> (kodhama/grove#66, 2026-07-17), which asked the parallel `/trellis:setup`
> review to cross-read its shared-pattern findings — this is move 2
> ("tooling-invisible install"). Sibling to `decision-0048` (the setup-git
> hand-back); both narrow setup's footprint on the consumer's environment.
> `adr-0014` and `decision-0048` are both **drafts**, cited here as
> provenance in prose, **not** consumed as frontmatter edges.

# 0049 — setup offers to hide `.trellis/` from the consumer's own tooling

## Context

`/trellis:setup` writes the `.trellis/` overlay into the consumer's repo. That overlay is
**vendored trellis territory, not consumer source** — but the consumer's own linters and formatters
do not know that and will act on it. Two consequences, surfaced by the math-quest pilot (via grove's
parallel install, `adr-0014`):

- **A consumer formatter reformatting `.trellis/*.md` breaks trellis's own verify.** The generated
  overlay files (`invariants.md`, `profile.md`, `trellis.md`, and the `version` stamp) are checked
  **byte-for-byte** against the shipped manifest on every refresh (`SKILL.md` step 6). A Prettier or
  markdownlint pass that rewraps or re-styles them changes those bytes, so the next `/trellis:setup`
  refresh **fails its checksum check**. The consumer's tidy-up silently sabotages the install's
  integrity oracle.
- **You should not lint or format a dependency.** Even without the checksum break, `.trellis/` is a
  vendored namespace; flagging or reformatting it is noise on files the consumer neither wrote nor
  owns.

grove hit the sharper *JS* version (ESLint `no-undef` on `.grove/check/**` runtime). trellis's
overlay is markdown, so that JS failure does not apply to us — but the **markdown-formatter** path
does, and it is worse for us, because it corrupts the verify rather than just adding lint noise.

## Decision

**Setup detects the consumer's linters/formatters and offers — never imposes — to add `.trellis/`
to their ignore, augment-never-clobber.**

- **Detect**, best-effort, by config presence: ESLint (`.eslintrc*`, `eslint.config.*`,
  `eslintConfig` in `package.json`), Prettier (`.prettierrc*`, `.prettierignore`), Biome
  (`biome.json`), markdownlint (`.markdownlint*`, `.markdownlintignore`). An undetected tool is
  reported as "none found," never a false claim.
- **Offer, with consent.** If a tool is found, ask whether to add `.trellis/` to its ignore. Never
  write without a yes (`floor-intent-gate`); if declined, note it and move on.
- **Augment-never-clobber.** Add `.trellis/` to the tool's existing ignore mechanism, without
  duplicating an entry already present and without touching any other line — the same idiom setup
  uses for the managed block. Create an ignore file only if the tool needs one and none exists, and
  say so.
- **Surface exactly what was touched** (`floor-transparency`): name each ignore file changed and the
  line added, in the confirm step.

**Scope: the whole `.trellis/` directory** — matching grove's whole-namespace rule, and simplest.
One honest, trellis-specific caveat: `.trellis/expression.md` is the consumer's **hand-owned** file
(excluded from checksum verify), so unlike the generated files it does **not** need ignoring — a
formatter over it is harmless. Ignoring it anyway (whole-directory) costs nothing and keeps the rule
simple; the load-bearing target is the generated files.

## Consequences

- `/trellis:setup` gains an interactive step (detect → offer → augment; new step 7) that, for the
  first time, may touch a consumer file **outside `.trellis/` and the managed block** — always with
  consent, always augment-never-clobber. The skill's "never touch anything outside the bundle or the
  markers" rule gains this one **offered, consented** exception; recorded here so it is not silent
  scope-creep.
- The confirm step reports any ignore file touched (or that the offer was declined / no tooling
  found).
- The `reference/checksums` overlay manifest is unchanged, but the new step edits `SKILL.md`, which
  **advances `install.sh`'s bundle-wide manifest** (`TRELLIS_BUNDLE_MANIFEST` covers the whole
  `plugins/trellis/` tree, `SKILL.md` included) — regenerated in the same change (`decision-0028`),
  as `TestInstallScriptBundleManifestIsCurrent` guards.
- The consumer's lint-ignore is a **separate concern** from trellis's own verify: ignoring
  `.trellis/` in Prettier does not affect setup's step-6 checksum check (which reads the files
  directly). It only stops the consumer's tooling from *mutating* them.
- **Symmetric removal (follow-up, not done here).** `/trellis:remove` should offer to strip any
  `.trellis/` ignore entry setup added — augment-never-clobber in reverse. A matching `remove` skill
  edit; flagged so the residue is not forgotten.

## Open questions

- **`.trellis/` + `.grove/` consolidation** (maintainer, parked explicitly-later per `adr-0014`): if
  the two vendored namespaces move under one hidden root, the ignore target changes to that root.
  Noted so this decision is revisited then, not left silently stale.
- **Non-JS / unknown tools.** Detection covers the common set; an exotic formatter gets "none found."
  Is best-effort detection + an honest "none found" enough, or should setup also print a generic
  "ignore `.trellis/` in whatever formats your repo" line regardless? Leaning best-effort, to avoid
  injecting noise.
- **expression.md:** ignore whole-directory (simple) vs. only the generated files (precise). Chose
  whole-directory; revisit if a consumer wants their own formatter on their own `expression.md`.

## Self-check (gate)

Implements `floor-intent-gate` (offer with consent, never impose) and `floor-transparency` (name
exactly what was touched), on the augment-never-clobber idiom (`inv-minimal-first`) — all
`signature-catalog-v1`. Provenance recorded honestly: grove's `adr-0014` (a cross-repo **draft**)
and the sibling `decision-0048` (also draft) are cited in prose, **not** as frontmatter edges, so no
draft is consumed as an upstream and no `informed_by → draft` edge is asserted. `depends_on` carries
genuine coupling (the invariants the behaviour implements); `spec-0005` and `decision-0043` are
`informed_by` — the manifest/verify context that makes the checksum-break motivation real, and the
payload host — neither strictly correctness-contingent, since "don't format a dependency" stands
without them (`decision-0047`). Depends only on ratified/gated upstreams; no draft consumed. Left at
`draft`: the author does not grade its own decision — gating and the `approved` intent act are the
maintainer's (`decision-0046`), ideally after an independent pass (`inv-independent-judgment`).
