---
id: decision-0044
type: decision
status: approved  # ratified by PR #133 merge (2026-07-10)
depends_on: [spec-0001, decision-0037, decision-0042, kodhama-0004-uniform-lifecycle]
owner: agent
date: 2026-07-10
---

> **Approved and merged.** Originally authored during a
> session where the maintainer was away and unavailable for the normal interactive shaping
> conversation (`shaper`-style back-and-forth) this family's process calls for before a new
> decision is drafted; every genuinely open design choice was flagged rather than quietly
> resolved. **On 2026-07-10 the maintainer weighed in directly on the four substantive open
> questions** — delimiter, registry membership, retrofit vs. grandfather, and adoption scope —
> now recorded as resolved rather than open in `## Open questions` below. That input is what
> moved this decision from `draft` to `gated`: self-checked and shaped by the maintainer's actual
> call, then approved at the merge gate — the maintainer's own merge of PR #133 (2026-07-10),
> which is itself the ratification act (per this repo's own lifecycle mapping, `decision-0037`
> point 3); PR #136 followed as frontmatter-only bookkeeping recording that bump
> (`gated → approved`). Per that same mapping, already declared in `CLAUDE.md`: `owner: agent`
> here still means *authorship*, not accountability — the accountable human was the maintainer,
> exercised at the merge gate (`decision-0022`), which is exactly what happened via PR #133's
> merge.
>
> **Two items were not specifically weighed and default to the draft's own recommendation** —
> depth of resolution (v0: shape + registry-membership check only, no fetch-and-verify) and
> wisp's non-frontmatter citation channel (left unresolved here, tracked separately as wisp's own
> GH issue #13). Both are marked as defaults, not confirmed maintainer calls, in `## Open
> questions` below — revisable if either assumption turns out wrong.
>
> **Self-illustration, not a defect.** This decision's own `depends_on` cites
> `kodhama-0004-uniform-lifecycle` — a bare, unqualified cross-repo id. This decision is now
> `approved` and merged, and the follow-on contract-author pass it called for has since landed
> too (`spec-0001` §1 amended via PR #137, merged 2026-07-10, to recognize the qualified
> `<repo>/<id>` form). That amendment does not retroactively qualify this citation, though: it is
> still bare and unqualified today, so under the now-ratified contract it remains a dangling
> reference by the letter of rubric check 4. It is left exactly as-is here, on purpose: it was
> the live instance of the gap this decision is about, not something to paper over while shaping
> the fix — retrofitting it (to `kodhama/kodhama-0004-uniform-lifecycle`) is separate follow-on
> work, the same class the Consequences section already scopes out for the other known dangling
> references.

# 0044 — Cross-repo `depends_on` references: a qualified `repo/id` form (proposal)

## Context

A 2026-07-10 family-wide consistency sweep (`corpus-reviewer` + `conformance-reviewer` across
kodhama, trellis, grove, wisp, design-system, math-quest — recorded in kodhama's
`conductor/wave-consistency-sweep.md`) found the single highest-leverage recurring gap: **no
repo declares a convention for `depends_on` references that cross repo boundaries.** Concrete
instances:

1. **kodhama** `decisions/0001-family-delivery.md:5` —
   `depends_on: [adr-0030-espalier, discovery-espalier-runtime-viz]`, both ids living in
   math-quest's own corpus, not kodhama's.
2. **trellis** (this repo) `specs/0005-curl-install-mechanical-vendoring.md:5` — `depends_on`
   includes `kodhama-0007-one-render-many-copiers`, dangling by this repo's own contract; that
   spec's own self-check incorrectly graded the entry PASS by reasoning about the referent's
   real-world status rather than checking it against the declared allowlist (`spec-0001` §1
   names only `brief-§…`).
3. **wisp** `protocol.ts:5` and `dashboard.html:353` cite math-quest's ADR-0030 in code
   comments, with zero footprint anywhere in wisp's own `decisions/`/`specs/` — the same shape
   of gap, but through a different citation channel (a code comment, not YAML frontmatter) —
   see the note under Consequences.
4. **grove** decisions cite `kodhama-0003`/`kodhama-0007`; these happen to resolve today, but no
   declared convention makes that legitimate rather than accidental.

**What `spec-0001` already has, and what it doesn't.** §1 already treats external references as
a first-class concept — *"a `depends_on` entry that is not an artifact `id` must match a
declared external-ref prefix (v0 allowlist: `brief-§…`). Anything else is a dangling reference →
fail"* — and the rubric's check 4 enforces the same allowlist. So the *mechanism* (a declared
allowlist gates what counts as a legitimate non-local reference) already exists; only the
*allowlist's contents* are narrow. `brief-§…` is itself a soft, unverifiable anchor (a section
cite into a planning brief, e.g. `decisions/0001`'s own `depends_on: [brief-§9.1, brief-§7]`) —
generalizing it to cover another repo's *artifact id* is an extension of an existing pattern,
not a new concept.

**A naming-convention asymmetry that bears directly on the mechanism choice.** Checking every
family repo's own `id:` convention (read directly, not assumed):

| Repo | Own id convention | Self-prefixed with repo name? |
|---|---|---|
| kodhama | `kodhama-0007-one-render-many-copiers` | **Yes** |
| trellis | `decision-0037` (no slug appended) | No — but also doesn't collide, see below |
| math-quest | `adr-0030-espalier`, `discovery-espalier-runtime-viz` | **No** |
| grove | `adr-0001-corpus-reviewer-lift` | **No** |
| wisp, design-system | template only so far (`adr-000x-short-slug`) | **No** |

Only kodhama's own ids happen to carry an unambiguous repo-name prefix already. Every other
repo that authors decisions uses a generic `adr-*` (or, for trellis, `decision-*`) numbering
local to itself — the same shape of id could exist in two different repos' corpora with no
textual way to tell them apart. This matters for the choice below.

## Decision

**Adopt a qualified `repo/id` form as the recognized cross-repo external-reference shape**, to
be built into `spec-0001` §1's allowlist by a follow-on contract-author pass once this decision
is approved (not built here — this decision proposes the mechanism only, per this repo's own
stage discipline: contract-author writes specs from an *approved* decision, never a draft).

**Form:** `<repo>/<id>`, where `<id>` is the referenced artifact's own id exactly as declared in
its home corpus (e.g. `math-quest/adr-0030-espalier`, `kodhama/kodhama-0007-one-render-many-copiers`).
`<repo>` must be a member of the recognized registry — resolved by the maintainer 2026-07-10 (see
Open questions): kodhama's own declared family list (kodhama, trellis, grove, wisp,
design-system, homebrew-tap) **plus math-quest**, chosen over the family list alone because
math-quest is the most-cited external corpus in the concrete motivating instances above.

**Why the qualified form over a looser repo-name-prefix allowlist** (the two options this
decision was asked to weigh):

- **Option (a) — qualified `repo/id`, chosen.** Unambiguous by construction: a delimiter
  structurally separates the origin repo from the local id, regardless of whether that repo's
  own convention happens to self-prefix. Checked directly: no existing `id:` value anywhere in
  trellis, kodhama, grove, or math-quest's corpora contains an embedded colon or slash, so
  either delimiter is a safe, non-colliding choice today.
- **Option (b) — a looser allowlist of recognized repo-name *prefixes*.** This is the
  weaker fit for the actual concrete instances above, not just a stylistic difference: it only
  works when a repo's own id convention already self-embeds its name, which — per the table
  above — is true for **kodhama alone**. It resolves instance 4 (grove citing
  `kodhama-0003`/`kodhama-0007`) and would have resolved instance 2 if trellis's own ids
  self-prefixed (they don't, but the *referent* here does — `kodhama-0007-...` — so it happens
  to work by the referent's convention, not trellis's). It does **not** resolve instance 1 —
  kodhama's own reference to math-quest's `adr-0030-espalier` and
  `discovery-espalier-runtime-viz` — because neither carries a `math-quest-` prefix, and
  requiring one would mean editing math-quest's own ratified ids, which no repo has standing to
  do. A fallback reading of (b) — "presume external if the id doesn't match this repo's *own*
  local id pattern" — would resolve instance 1, but at a real cost: it converts a structural
  check into a heuristic one, and a heuristic that treats "doesn't match my local pattern" as
  "therefore legitimately external" is exactly the shape of false-pass risk `spec-0001` AC1
  ("no false pass / no vague fail") and the rubric's honesty clause exist to prevent — a genuine
  typo'd or missing local id would silently read the same as a deliberate cross-repo reference.
- **corpus-reviewer's check 4, concretely:** gains one new branch — an entry matching
  `<registered-repo>/<rest>` is accepted as a declared external reference (recognized, not
  further resolved against this repo's own corpus — the same non-resolution treatment `brief-§…`
  already gets, since this repo generally cannot fetch another repo's live corpus to verify the
  referent actually exists). v0 checks shape and registry membership only, matching the existing
  `brief-§…` precedent's own strictness level; whether a *stronger* fetch-and-verify form of
  resolution is ever worth building stays out of scope by default — the maintainer did not
  specifically weigh this one (see Open questions).

**What this decision does not do:** it does not itself amend `spec-0001`, and does not itself
retrofit any of the four existing instances. Per the maintainer's 2026-07-10 resolution on
adoption scope (see Open questions), it does *not* additionally require a separate per-repo
adoption act in kodhama/grove/wisp/design-system either — the follow-on `spec-0001` amendment
composes onto them automatically. See Consequences.

## Consequences

- **`spec-0001` §1 gains a second recognized external-ref form**, alongside `brief-§…` — built
  by a follow-on contract-author pass once this decision is `approved`, not by this decision.
  `core/rubrics/artifact-contract.md` check 4 (and its "Open questions" note, which already
  named "external-ref mechanism: an allowlist prefix (v0) vs a registry artifact — revisit when
  refs multiply") gets the matching update at the same time.
- **This is trellis's own contract, and per the maintainer's 2026-07-10 resolution it suffices
  family-wide on its own — no separate per-repo adoption step is needed.** `spec-0001` is
  explicitly framed (§2, by analogy to `decision-0037`'s status precedent) as *"what this repo
  runs and what setup composes onto a project that brings no lifecycle of its own."* The
  maintainer weighed this against the two-step pattern `kodhama-0004-uniform-lifecycle` +
  trellis's own `decision-0042` already used for the lifecycle vocabulary, and chose the simpler
  option: amending `spec-0001` here is sufficient, composing onto every sibling repo the same way
  this repo's other defaults already do, with no matching family-wide declaration + per-repo
  adoption decision owed in kodhama, grove, wisp, or design-system for this convention. This
  resolves what the earlier draft of this decision left open.
- **wisp's code-comment citation of ADR-0030 is a different channel, not directly covered — left
  unresolved here, as originally scoped.** This decision's mechanism is a `depends_on`-frontmatter
  form; `protocol.ts`/`dashboard.html` cite math-quest's ADR-0030 in source comments, outside any
  artifact's frontmatter entirely. The maintainer did not specifically weigh in on whether the
  same qualified-id convention should also govern non-frontmatter citations, or whether wisp's
  right fix is instead to file its own local decision formally adopting/citing ADR-0030; this
  decision defaults to leaving it unresolved, per its own original scoping. The question stays
  tracked as wisp's own GH issue #13 (the family sweep's parked item #6) — independent of this
  decision, and revisable if this default turns out wrong.
- **Retrofit (not grandfathering) is the maintainer's chosen direction, resolved 2026-07-10** —
  but the retrofit itself is not executed by this decision, and cannot be until the follow-on
  contract-author pass actually amends `spec-0001` to recognize the qualified form. Of the three
  known existing references this affects, two live outside this repo — kodhama's own
  `decisions/0001` and grove's `kodhama-0003`/`kodhama-0007` citations — and retrofitting those is
  separate follow-on work scoped to kodhama's and grove's own repos respectively, not something
  this trellis-only decision reaches into or executes here. (Re-checked directly: grove's
  citations already resolve to real, existing ids today — `kodhama-0003-family-naming` and
  `kodhama-0007-one-render-many-copiers` both exist verbatim in kodhama's own `decisions/`; the
  gap there is that no declared convention makes citing them bare *legitimate*, not that the
  citation is broken — consistent with the Context section's framing above.) The third instance —
  this repo's own `specs/0005-curl-install-mechanical-vendoring.md` — is trellis's own to
  retrofit, once `spec-0001` itself is amended; not done by this decision either. Note for
  whoever picks up any of these: trellis's own PR #131 already established a working precedent
  that a **frontmatter-only, fact-preserving correction to an already-`approved`/ratified
  artifact** (adding `superseded_by`/`superseded_in_part_by` fields that were owed but missing) is
  treated as legitimate bookkeeping, not a forbidden edit-in-place of ratified content — the same
  class of touch this retrofit is.

## Open questions

Four of the six items below were put to the maintainer directly and are now resolved
(2026-07-10). The remaining two were not specifically weighed and default to this proposal's own
stated recommendation instead — flagged as such, not silently treated as settled.

- **Delimiter — resolved 2026-07-10: `/`.** The maintainer chose this proposal's own
  recommendation (`math-quest/adr-0030-espalier`) over the alternative `:` form illustrated in
  the original dispatching brief (`math-quest:adr-0030-espalier`). Both were mechanically safe —
  no existing id anywhere in the checked corpora contains either character — but `/` was
  preferred: it mirrors the `org/repo`-style qualification the maintainer already reads daily via
  `gh`/GitHub URLs, and reads less like a YAML mapping key at a glance during review.
- **Registry membership — resolved 2026-07-10: kodhama's declared family list PLUS
  math-quest.** The maintainer chose this proposal's own recommendation over the family list
  alone: kodhama, trellis, grove, wisp, design-system, homebrew-tap, **and math-quest** — chosen
  specifically because math-quest is the single most-cited external corpus in the actual
  motivating instances (kodhama's `decisions/0001`, wisp's code comments), and the family list
  alone would have left that motivating instance still unresolvable. Where this registry lives
  (duplicated per repo, or one canonical source every repo's contract points at) remains for the
  follow-on contract-author pass to settle when it amends `spec-0001`.
- **Retrofit vs. grandfather — resolved 2026-07-10: retrofit.** The maintainer chose this
  proposal's own recommendation over grandfathering: a corpus that still carries known-dangling
  references after the fix ships is a worse resting state than one that doesn't, and the PR #131
  precedent (frontmatter-only corrections to already-ratified artifacts are legitimate
  bookkeeping) supports the touch. This reaches into kodhama's and grove's own repos and is
  explicitly scoped as **separate follow-on work in those repos** — not executed inside this
  trellis-only decision or PR. See `## Consequences` for the concrete breakdown of what's owed
  where.
- **Adoption scope — resolved 2026-07-10: amending trellis's `spec-0001` alone suffices
  family-wide.** The maintainer chose the simpler single-step option over the two-step
  `kodhama-0004-uniform-lifecycle` → trellis `decision-0042` pattern used for the lifecycle
  vocabulary: no separate per-repo adoption decision is needed in kodhama, grove, wisp, or
  design-system.
- **Depth of resolution — not specifically weighed by the maintainer; defaults to this
  proposal's own recommendation.** v0 recognizes a `repo/id` entry on shape + registry
  membership alone, matching `brief-§…`'s existing non-verified treatment — no
  fetch-and-confirm-the-referent-exists mechanism for now. This was always this proposal's stated
  default; it stands because the maintainer did not weigh in on it specifically, and remains
  revisable later if that assumption turns out wrong.
- **wisp's non-frontmatter (code-comment) citation channel — not specifically weighed by the
  maintainer; defaults to this proposal's own recommendation.** Left unresolved by this decision,
  as originally scoped — tracked separately as wisp's own GH issue #13 (the family sweep's parked
  item #6), which covers the parked question of whether wisp should file its own local decision
  citing ADR-0030. Revisable later if leaving this unresolved turns out wrong.
