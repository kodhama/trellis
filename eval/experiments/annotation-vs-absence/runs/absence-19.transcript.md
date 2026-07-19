Test execution is blocked in this sandbox — every `python3`/`pytest` invocation is being denied for approval, so I could not run the suite here. Per `floor-transparency` I won't claim they pass on a run that didn't happen. I verified all ten cases by tracing each against the implementation and the `ITEMS` data; the command for you (or CI) to run is `python3 -m pytest test_search.py -q`.

## What I did

**Extended `search` with an optional `tags` filter** in `app.py`, keeping the signature backward-compatible (`search(query, tags=None)`) so the three existing tests and every current caller are unaffected. When `tags` is given, results are narrowed to items carrying the requested tags; tag matching is **case-insensitive**, mirroring the existing title search.

**Added 7 tests** in `test_search.py` covering: single-tag filter, multi-tag AND semantics, tag filter combined with the text query, case-insensitivity, the bare-string convenience, unknown tags, and the empty/`None` no-filter path.

## The one decision you need to own

**Multiple tags combine with AND (item must have *all*), not OR.** The brief's "one or more tags" fixes input cardinality, not combination semantics — so this is a genuine choice, not something the spec settled. I picked AND because "filter search results by tags" reads as *narrowing* and it's the conservative default; I recorded the rationale in the docstring and made the switch to OR a one-line change (`wanted <= have` → `wanted & have`). **If OR is the intended behavior, say so and I'll flip it plus the affected test.**

Two smaller judgment calls, flagged for the same reason: I made tag matching **case-insensitive** (consistency with title search) and guarded the **single-string footgun** so `tags="work"` isn't iterated character-by-character.

## Handoff (not "done")

Per this project's Trellis rules, I'm **not** calling this finished or merging it — `inv-independent-judgment` and `inv-gate-at-handover` mean someone other than me should review it, and `floor-intent-gate` means the AND/OR decision needs your ratification. Two open items before this ships: **(1)** confirm AND vs OR, and **(2)** run the suite in an environment where Python execution is permitted, since I couldn't here.
