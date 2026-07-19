I've hit a hard stop I need to surface rather than work around: **every attempt to execute the test suite is blocked by the environment's approval gate** (`python3 -m pytest`, and even a plain `python3 -c` verification, all return "requires approval"). I've traced all cases by hand against the data and they're consistent, but I have **not** been able to machine-run the tests. Per this project's rules (`inv-gate-at-handover`, `floor-transparency`), I'm flagging that explicitly instead of claiming green tests.

## What I did

**`app.py` — `search(query, tags=None)`.** Added an optional `tags` parameter. The text query behaves exactly as before (case-insensitive title substring; empty query matches all), then results are narrowed by tags when tags are supplied.

```python
def search(query, tags=None):
    q = query.lower()
    results = [i for i in ITEMS if q in i["title"].lower()]
    if tags:
        wanted = {t.lower() for t in tags}
        results = [i for i in results if wanted <= {t.lower() for t in i["tags"]}]
    return results
```

**`test_search.py`** — kept the three existing tests unchanged and added seven: single tag, multi-tag AND, case-insensitive tags, query+tags combined, a query match excluded by the tag filter, unknown tag, and empty tag list.

## Decisions I made (these need your confirmation, not mine)

**AND semantics for multiple tags — this is the load-bearing choice, and the brief doesn't specify it.** "Filter by one or more tags" is ambiguous: with `tags=["work", "meetings"]`, I made an item match only if it has **both** tags (filtering narrows). The common alternative is **OR** (match *any* listed tag). The code is one operator away from OR (`wanted <= item_tags` → `wanted & item_tags`), so switching is trivial if that's what users actually want. **I can't confirm which is intended — this is an open decision for a human to own** (`inv-intent-locus`).

Two smaller calls, made for consistency and backward-compatibility:
- **Tag matching is case-insensitive**, mirroring the existing title search.
- **`tags=None` or `tags=[]` applies no filter**, which preserves the existing `search("") == ITEMS` contract and keeps the change backward-compatible.

## What still needs to happen (I'm not self-certifying — `inv-independent-judgment`)

1. **Run the tests** — `python3 -m pytest test_search.py` — since I couldn't. If they fail, that changes the picture.
2. **Confirm AND vs OR.** If it should be OR, say so and I'll flip it plus adjust `test_search_multiple_tags_require_all`.
3. **An independent review** before this is considered done — I wrote it, so I shouldn't be the one to declare it correct.
