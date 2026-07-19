---
id: decision-0051
type: decision
status: approved  # maintainer's intent act 2026-07-19, in-conversation ("Approved. Continue to implementation") — this flip records it (decision-0046); corpus-reviewer conformance pass (all checks PASS) ran before the gate; grove adr-0018/0020 approved statuses confirmed against grove origin/main frontmatter before the gate
depends_on: [invariants-v1, signature-catalog-v1, decision-0033]
informed_by: [grove/adr-0018-gate-profile-and-trigger-split, grove/adr-0020-dispatcher-honors-gate-profile, decision-0016, decision-0043, decision-0049]
owner: agent
date: 2026-07-19
---

> **Provenance.** Shaped in-session with the maintainer (2026-07-19), from their ask to
> (1) fix the three-way "profile" name collision, (2) adopt grove's config-vs-internal
> separation, and (3) make the consumer's config actually select which rules each session
> carries. The family pattern adopted here is grove's `adr-0018` (authority split, TOML
> config surface, preset-as-seed) and `adr-0020` (config honored mechanically) — both
> **approved** (2026-07-18/19), so cited as `informed_by` frontmatter edges in the
> qualified `repo/id` form (`decision-0044`), unlike `decision-0049`'s prose-only citation
> of a then-draft.

# 0051 — the overlay splits by authority: consumer-owned `rules.toml`, generated files under `.trellis/internal/`

## Context

Three pressures converged:

- **The word "profile" means three things in this repo.** (1) The `expression-profile`
  artifact — the rich per-instance readout (active × C1 × C2 × delivery; `decision-0016`,
  `spec-0002`). (2) The `profile:` frontmatter key in `.trellis/expression.md` — a
  two-value posture enum (`a`/`b`), the overlay's only machine-read line. (3) The
  generated `.trellis/profile.md` — the active-rules readout, "(Generated from your
  profile …)". Nothing cross-links them; a consumer reading the overlay cannot know the
  word once meant the rich bundle. The maintainer hit this trap directly.
- **The config surface is illegible.** The consumer's entire machine-read configuration is
  three lines of YAML frontmatter buried inside a prose file whose name says "expression,"
  not "config." Which `.trellis/` files a consumer may edit vs. must not is knowable only
  by reading each file's comment.
- **The active-subset lever now has a user.** `decision-0033` parked `seed`/`custom`
  because "the one concrete lever a posture has today is *which invariants are active*,
  and only `seed`/`custom` used it" — parked "until the active-subset lever earns its
  keep." The maintainer now asks for exactly that: config that actually selects which
  rules each session carries.

Meanwhile grove (the family's operational layer) landed the pattern this decision adopts.
`adr-0018` D5: placement "**by who is authoritative on a file's contents** — *not*
machinery-vs-config" — a consumer-authoritative `.grove/` root ("setup writes defaults;
grove NEVER clobbers") and a grove-authoritative `.grove/internal/` ("regenerated verbatim
on update; hand-edits overwritten"). D9: "user-facing **config** files are uniform
**TOML** … the companions stay **markdown** — reference prose … a different *kind*." D7:
a preset is "a **seed-time operation**, not a file abstraction" — the explicit per-gate
rows are the source of truth, `seeded_from` is non-authoritative provenance, "the rows win
if they diverge."

One inherited rule sits in tension. `kodhama-0007` rule 4 (now homed in
`kodhama/stewards`, per kodhama-0009's relocation) prescribes "**one** hand-owned
declaration file per bundle" whose "**YAML frontmatter** carries the machine-read config"
and whose body "is never parsed by machinery." grove's ADRs diverge from that letter —
multiple hand-owned whole-file TOML configs — and reconcile it **by silence**: neither ADR
names the rule. Trellis is the reference install; repeating the silence here would be the
"quietly drift from the methodology it claims to follow" failure our own overlay names.

## Decision

**1. The overlay splits by authority (grove `adr-0018` D5, applied to `.trellis/`).**

- `.trellis/` **root — consumer-authoritative** (seeded once; setup never clobbers;
  excluded from manifest verification): `expression.md` (hand-owned **prose**, stays
  markdown per D9's kind rule — it is expression, not data) and **`rules.toml`** (new —
  the machine-read config).
- `.trellis/internal/` — **trellis-authoritative** (copied/regenerated verbatim on
  refresh; hand-edits overwritten; manifest-verified byte-for-byte): `trellis.md`,
  **`rules.md`** (the assembled active-rules readout — today's `profile.md`, renamed),
  `invariants.md`, `version`.

**2. `rules.toml` is posture-as-seed, rows-as-truth (grove D7).** Setup's one interactive
question (posture, unchanged from `decision-0033`: A·conductor / B·author-adapt) becomes a
**seed-time expansion** into explicit rows; after seeding, editing the rows *is* the
configuration act, and a refresh reads rows and asks nothing. Shape (keys normative;
grove's "illustrative names" hedge became load-bearing verbatim, so these are stated as
final unless the executor hits a concrete conflict, which must be surfaced):

```toml
seeded_from = "conductor"   # provenance only — the rows below win if they diverge
strictness  = "firm"        # firm (a·conductor) | adaptive (b·author-adapt)

[rules]                     # one row per assessable catalog slug (signature-catalog-v1)
inv-intent-locus         = { active = true }
inv-bounded-context      = { active = true }
inv-clarify-before-commit = { active = true }
# … remaining operating/structural slugs …
floor-transparency       = { active = true }   # floor-held — see rule 3
floor-intent-gate        = { active = true }   # floor-held — see rule 3
```

**3. Floors are not rows a consumer can turn off.** `floor-transparency` and
`floor-intent-gate` rows exist (the table is honest about the full set) but are
**floor-held**: assembly ignores `active = false` on them, includes them anyway, and says
so loudly in the confirm step — fail-open on the floors, never silent (the analog of
grove's every-read floor validation, `adr-0018` F1).

**4. The readout is assembled from manifest-covered fragments at refresh — the copier
discipline survives at fragment granularity.** The payload ships one pre-rendered fragment
per rule (`reference/rules/<slug>.md`), each covered by `reference/checksums`. Setup
**selects and concatenates** the active fragments in catalog order into
`.trellis/internal/rules.md` — a deterministic mechanical assembly, not composition: no
byte is authored at install time, so `kodhama-0007`'s "no second writer" holds. Verify
extends naturally: each included fragment checks against the manifest, and the assembled
file must equal their ordered concatenation. Delivery is unchanged — the managed block's
`@import` (or the inline block) carries the assembled readout into every session, so an
edited row takes effect at the next refresh, with no per-session reader, no hook, no
runtime machinery (trellis has no dispatcher; grove's run-time honoring has no analog
here to build).

**5. The name "profile" is de-collided.**

- The `profile:` frontmatter key in `expression.md` **retires**. Migration on refresh: no
  `rules.toml` + a legacy `profile:` key → seed `rules.toml` from that posture and offer
  to strip the frontmatter; no `rules.toml` + no key → ask the posture question (as
  today). `expression.md` becomes pure hand-owned prose.
- Generated `profile.md` → `.trellis/internal/rules.md` ("Generated from your
  `rules.toml` …" as its closing line).
- "Profile" is **reserved** for the `expression-profile` artifact (`decision-0016`);
  `expression.md`'s seed comment gains a cross-link saying exactly that.

**6. The divergence from `kodhama-0007` rule 4 is named, not silent.** This decision
keeps rule 4's generalized half — "every file in an installed bundle is 100% generated or
100% hand-owned — never mixed" — and strengthens it to directory granularity. It departs
from the letter of the carrier half: the machine-read config moves from "YAML frontmatter
of the one hand-owned file" to a dedicated hand-owned TOML file, and "one hand-owned
declaration file" becomes "one hand-owned file per config axis" (grove D9's one-home-per-
kind). grove already made this move family-wide by silence; this decision records it and
proposes a **stewards-level amendment** of rule 4 as follow-on work (see Open questions —
the stewards repo is not cloned locally).

**7. What this does *not* revive.** `seed` and `custom` postures stay parked
(`decision-0033` is superseded **only in part**: the active-subset lever returns as
consumer-editable rows; the parked *presets* do not). Per-rule C1/C2 dials stay out of
`rules.toml` — `strictness` is one instance-level key, not a per-row column — until
something enforces them (`decision-0033`'s "descriptive, not enforced" finding still
holds; adding dead columns is `inv-minimal-first` failure).

## Consequences

- **`/trellis:setup` (`SKILL.md`) is rewritten**: posture step seeds `rules.toml` (not
  `expression.md` frontmatter); new assembly step (rule 4 above); migration step for
  existing flat-layout overlays (move generated files under `internal/`, delete old-path
  copies, migrate the legacy `profile:` key); verify gains the fragment/concatenation
  checks. The `#112` guard (step 2) extends to `rules.toml` hand-edits: rows are the
  consumer's — never clobbered, so nothing to rescue; the guard's target stays the
  generated files.
- **Derived resources advance in the same change (`decision-0028`):** the payload
  `reference/checksums` manifest (per-fragment entries; `rules.toml` and `expression.md`
  stay excluded); `block-claude.md`'s import path (`@.trellis/internal/trellis.md`) and
  the `block-inline-<p>` variants; the staleness hook's compared path
  (`.trellis/internal/version`, `decision-0043` mechanics unchanged); the `remove` skill;
  `install.sh`'s bundle-wide manifest (`TestInstallScriptBundleManifestIsCurrent`
  guards); the plugin `README`.
- **Installed overlays across the family** (trellis-self, grove, kodhama, wisp,
  design-system) migrate on their next `/trellis:setup` refresh — the migration step
  makes the refresh the rollout vehicle; no flag-day.
- **`decision-0033` gains an append-only forward pointer** to this record on approval
  (superseded-in-part: the active-subset parking; the preset parking stands).
- **`decision-0049`'s ignore-offer scope is unchanged** (whole `.trellis/`), and its
  caveat extends: `rules.toml` joins `expression.md` as consumer-owned files that need no
  checksum protection — the load-bearing ignore target is now `internal/`.
- **No `set-posture` skill yet** (grove's `set-profile` analog): editing `rules.toml` +
  refresh is the whole flow; a wholesale-switch skill is deferred until asked for
  (`inv-minimal-first`).

## Open questions

- **Which rule would a real consumer deactivate first?** `decision-0033`'s "earn its
  keep" question is answered by the maintainer's ask, but a named concrete instance (the
  consultant-mode work project is the candidate) would make the lever's value checkable,
  not asserted.
- **The stewards amendment.** Rule 4's letter is now de facto superseded family-wide;
  `kodhama/stewards` (its new home) is not cloned locally. Who drafts the amendment,
  and does trellis block anything on it? (Position here: no — grove's approved ADRs are
  the operative family pattern; the amendment is repair, not a gate.)
- **Fragment granularity vs. posture.** Fragments are posture-independent today (the
  readout is byte-identical across a/b; only `trellis.md`'s strictness line differs). If
  a future posture wants per-posture rule wording, fragments double — accept then, not
  now.
- **`internal/` naming.** Chose `internal/` for exact family symmetry with
  `.grove/internal/`. Alternative (`generated/`) is more self-describing but breaks the
  symmetry; flagged in case the maintainer weighs legibility over symmetry.

## Self-check (gate)

Implements the authority half of `kodhama-0007` rule 4 while **explicitly** departing from
its carrier half (rule 6 — `floor-transparency`: the divergence is recorded, not silent).
Floors stay non-configurable (`floor-intent-gate`, rule 3). Machinery added is minimal
(`inv-minimal-first`): one new file, one new directory, zero runtime readers; the
per-session hook variant was considered and rejected (no reader exists; the inline path
could not use it). Cross-repo edges use the qualified `repo/id` form (`decision-0044`) and
cite only **approved** upstreams — `grove/adr-0018`, `grove/adr-0020` are approved
2026-07-18/19; no draft is consumed (`inv-directional-flow`). `depends_on` carries genuine
coupling only (the rule set the rows key off, the catalog the slugs resolve to, and the
decision this supersedes in part); grove's ADRs and the sibling context are `informed_by`
(`decision-0047`). Left at `draft`: the author does not grade its own decision — gating
and the `approved` flip are the maintainer's intent act (`decision-0046`), ideally after
an independent pass (`inv-independent-judgment`).
