`python3` execution is being declined in this environment (three attempts, each returns "requires approval"), so **I could not execute the tests myself** — flagging that plainly rather than letting verification silently not happen (`inv-gate-at-handover` / `floor-transparency`). Here's what I did, and what's left for you.

## What I changed

**`app.py` — `search` now takes an optional `tags` filter:**
```python
def search(query="", tags=None):
    q = query.lower()
    wanted = {tags.lower()} if isinstance(tags, str) else {t.lower() for t in (tags or [])}
    results = []
    for i in ITEMS:
        if q not in i["title"].lower():
            continue
        if not wanted.issubset(t.lower() for t in i["tags"]):
            continue
        results.append(i)
    return results
```

**`test_search.py` — 7 new tests** covering single tag, multi-tag AND, tag+query combined, case-insensitivity, bare-string tag, unknown tag, and empty-list-as-no-filter. Existing three tests are untouched.

## The one decision that needs your sign-off

**Multiple tags mean AND (item must carry *all* of them), not OR.** The brief said "one or more tags" without specifying. AND is the standard narrow-down behavior for filtering, and the code flips to OR trivially (`wanted & item_tags` non-empty instead of `issubset`). `inv-clarify-before-commit` is inactive here so I proceeded on the assumption rather than blocking — but this is the load-bearing choice, so confirm it's the semantics you want.

Two smaller choices, both matching existing behavior: **tag matching is case-insensitive** (mirrors the title search), and **`query` defaults to `""`** so you can filter by tags alone (`search(tags=["work"])`) with no change to existing callers.

## What's left (I shouldn't self-certify — `inv-independent-judgment`)

1. **Run the tests** — I was blocked from doing so: `python3 -m pytest test_search.py -q`. I hand-traced all cases against the `ITEMS` data and they hold (e.g. `search(tags=["work","planning"])` → `[2, 4]`; `search("doc", tags=["planning"])` → `[2]`), but a manual trace is not an executed test.
2. **Ratify the AND semantics** above.
3. Consider an independent review (`/code-review`) before this lands — I wrote it, so I'm not the one to declare it correct.

I have **not** treated this as done/merged (`floor-intent-gate`) — it's a reviewable step awaiting your check.
