---
id: decision-0038
type: decision
status: ratified
depends_on: [invariants-v1, decision-0013, decision-0034]
owner: gundi
ratified: 2026-07-07
---

# 0038 — Retire the display codes; slugs are the only name, for reference *and* display

**Raised by:** the maintainer (2026-07-07, from the instance-#1 conductor session): *"I want to
retire the B3, C1… ids and just use the invariant slugs everywhere because I have no idea what
you are talking about when you use them."*

## Context

`decision-0013` made slugs the canonical *reference* form but kept the `A/B/C/D`+number
ordinals as **frozen display labels** — "convenient to say." The convenience turned out to be
one-directional: convenient for the file, opaque to the human. A display label whose only
audience cannot decode it has failed its one job. This is the same disease `decision-0034`
diagnosed on the always-loaded block — a cold-read agent scored internal codes as
*"invit[ing] paralysis or performative compliance"* — now recognized at the human-facing
layer: the codes cost the *maintainer* the conversation, at the intent gate of all places.
Slugs carry their meaning in their name (`inv-independent-judgment` needs no registry lookup;
`B3` always does).

## Decision

1. **Slugs are the ONLY name** — for reference (`decision-0013`, unchanged) **and for
   display**. The letter-number codes are retired everywhere, including headings, badges, and
   prose asides. Group letters (`A/B/C/D` as section labels) go with them; the groups are
   named by what they are: *structural (admission gate) · operating · dials · floors*.
2. **A legacy map stays, in one place** — the `invariants-v1` Identifiers registry keeps the
   code column, retitled *"Legacy code (retired)"* — so readers of old decisions can resolve
   `B3` → `inv-independent-judgment`. That registry remains the resolution mechanism for
   every historical reference; nothing else may use the codes.
3. **Append-only artifacts are not rewritten.** Decisions and past research keep their
   original code-bearing text (`decision-0013`'s historical-reference exemption, intact);
   the legacy map is how those reads resolve.

## Consequences

- **Swept in this change (slug-first):** `core/invariants/trellis-invariants-v1.md` (headings,
  prose, registry retitled), `core/catalog/signature-catalog-v1.md` + its three verbatim
  copies (`cli/assets/`, `plugins/trellis/reference/`, `.trellis/invariants.md`),
  `docs/invariants.html` (card badges now show slugs; group letters dropped),
  `core/lexicon.md`, `profiles/trellis-self.md`, `README.md`, `CLAUDE.md`, the plugin setup
  skill, `core/rubrics/artifact-contract.md`, the `conformance-reviewer` agent, and the
  **eval invariant scorecard** — `eval/scorecards/invariants.md` is a *derived resource of the
  catalog* (`decision-0028`), so it migrates like the other derivatives: its generator
  (`eval/gen-invariant-scorecard.py`) parsed entries *by the retired codes* and now parses
  slug-only headings, emitting `## <slug>` sections (and failing loudly on a zero-entry parse
  instead of writing an empty scorecard). *Honest note: the first push of this change
  misclassified the scorecard as exempt; the `decision-0028` sync guard (`eval-scorecard-sync`)
  went red and caught it — the derived-resource check doing exactly its job.*
- **Named residuals (deliberate, not oversights):**
  - **`spec-0001`–`spec-0004` and the research notes** still carry codes in ratified prose —
    grandfathered by `decision-0013`, resolvable via the legacy map; migrate **opportunistically
    as each file is next edited** (`inv-ride-existing-rituals` — no sweep ceremony over
    revise-in-place files nothing else touches).
  - **Schema field names** `C1`/`C2`/`default_C1`/`default_C2` (`spec-0002` §1–2, the catalog
    entries, profile columns) are **identifiers, not display** — renaming them is a spec-0002
    schema amendment with a conformance-check cascade; see Open questions.
  - **Go-internal names/comments** (`C1Lean` etc.) are developer-internal; exempt. Past
    **eval run outputs** (`eval/runs/`) are historical records of already-scored runs; exempt
    like decisions. *(The live eval scorecard is NOT exempt — it is a derived resource and was
    swept; see above.)*
- The retired codes stay **frozen** — never reassigned to future invariants, so the legacy
  map can never become ambiguous.

## Open questions

- **Rename the schema fields?** `default_C1` → `default_strength`, `default_C2` →
  `default_gatekeeper` (and profile columns to match) would finish the job at the schema
  layer — a `spec-0002` amendment + rubric/agent/catalog/profile cascade. Decide when
  `spec-0002` is next opened; the dial *slugs* (`dial-enforcement-strength`,
  `dial-gatekeeper`) already name the axes.

## Supersedes / superseded by

**Supersedes `decision-0013` in part** — the "frozen display label" role of the ordinals is
retired. The rest of `decision-0013` stands and is strengthened: stable slugs, references use
slugs, registry-resolution for merges/retirements, no citing artifact edited to chase a
rename.
