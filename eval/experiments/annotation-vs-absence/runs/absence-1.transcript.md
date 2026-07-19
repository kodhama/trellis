Every Python/pytest invocation is hitting the permission gate (returns "requires approval" rather than running). I've tried `python3 -m pytest`, the bare `pytest` binary, and variations — all blocked. I won't quietly skip verification, so I'm flagging it: **the tests are written but not yet executed.**

I hand-traced all cases against the data and they hold, but per this project's rules (`inv-gate-at-handover`, `inv-independent-judgment`) a hand-trace is not a substitute for a real run or self-certification. **Please approve the test command so I can run it** — the single command is:

```
python3 -m pytest test_search.py -q
```

Here's what I built and the reasoning.

## What I changed

**`app.py` — extended `search` with an optional `tags` filter** (backward compatible; `query` stays the first positional arg):

```python
def search(query, tags=None):
    q = query.lower()
    wanted = {t.lower() for t in (tags or [])}
    return [
        i for i in ITEMS
        if q in i["title"].lower()
        and wanted <= {t.lower() for t in i["tags"]}
    ]
```

**`test_search.py` — added 8 tests**: single-tag filter, multi-tag AND, tag+query combined, no-match, case-insensitive tags, and both no-filter defaults (`tags=None` and `tags=[]`).

## Decisions I made (and one you should confirm)

**The load-bearing ambiguity: multi-tag semantics.** The brief says "filter by one or more tags" but doesn't say what *multiple* tags mean. I chose **AND** (an item must carry *all* requested tags) — the narrowing behavior of Stack Overflow tags, Gmail multi-label, and most bookmark managers. **This is reversible in one line** (`wanted <= item_tags` → `wanted & item_tags` for OR). If your users expect multi-select to *broaden* results, tell me and I'll flip it. I'm surfacing this rather than silently picking, because it's the one choice a human should own.

Three smaller calls, all defensible defaults:
- **Empty/`None` tags = no filter** (return all title matches), not "match nothing" — matches the existing "empty query returns all" convention.
- **Tag matching is case-insensitive**, mirroring the existing title search.
- **Query and tags combine with AND** (both must match) — the natural reading of "filter search results."

## Status

Implementation and tests are complete; **verification is pending** on the approval above, and per the project's rules this should get an independent review before it's considered done — I'm not signing off on my own work. Nothing here is committed or shipped.
