Both pytest and inline-python runs are blocked pending approval in this environment, so I could not execute the suite. Per the project's transparency rule I'm flagging that rather than implying it ran. I hand-traced all 11 tests against the data and code below; to actually run them: `python3 -m pytest test_search.py -q` (needs your approval, or approve the run I attempted).

## What I did

**Extended `search` to filter by tags while keeping every existing call working.** New signature: `search(query="", tags=None)`.

```python
def search(query="", tags=None):
    q = query.lower()
    wanted = {t.lower() for t in (tags or [])}
    results = []
    for i in ITEMS:
        if q not in i["title"].lower():
            continue
        if wanted and not wanted.issubset(t.lower() for t in i["tags"]):
            continue
        results.append(i)
    return results
```

- **Text and tags stack as narrowing filters.** An item must match the title query *and* the tag filter. `query` now defaults to `""`, so `search(tags=["work"])` filters by tag alone.
- **Tag matching is case-insensitive**, mirroring the existing title behavior (`search(tags=["WORK"])` works).
- **Backward compatible.** The three original tests are unaffected — `tags=None` applies no filter, empty query still matches all.

## The one decision you should ratify

**Multiple tags use AND semantics — an item must carry *all* requested tags.** The brief says "one or more tags" without specifying combination, so this is a genuine choice, not something the spec settled. I picked AND because stacking tags to *narrow* results is how faceted tag filters usually behave, and it composes consistently with the text search (also narrowing).

**If you want OR instead** (match *any* tag — broadening), it's a one-line flip: change `wanted.issubset(...)` to `wanted & {t.lower() for t in i["tags"]}` (non-empty intersection). Tell me and I'll switch it plus the affected tests.

## Tests added (8 new, in `test_search.py`)

Tag-only filtering, multi-tag AND (`["work","planning"]` → `[2,4]`), no-common-item empty, text+tag combined and text+tag exclusion, case-insensitive tags, and empty-tag-list-is-no-filter. All hand-trace to passing against `ITEMS`.

**Two things per the project's gate rules:** these tests are the author's own — they should get an independent review and an actual test run before this is merged, and I haven't done either (nor is this a git repo to commit into). And the AND-vs-OR choice above is yours to confirm — I don't consider "done" ratified until you do.
