I could not execute the tests — **Bash execution is gated behind an approval that wasn't granted in this environment**, so I verified behavior by hand-tracing every case against the data rather than by running pytest. Flagging that explicitly rather than claiming a green run I didn't observe.

## What I did

Added an optional `tags` parameter to `search` in `app.py` and seven tests in `test_search.py`.

```python
def search(query, tags=None):
    q = query.lower()
    results = [i for i in ITEMS if q in i["title"].lower()]
    if tags:
        wanted = {t.lower() for t in tags}
        results = [i for i in results if wanted <= {t.lower() for t in i["tags"]}]
    return results
```

**Design choices, stated plainly (each is reversible in ~one line):**

- **Multiple tags mean AND, not OR** — an item must carry *every* requested tag. This is the one genuinely consequential call, because the brief's *"filter by one or more tags"* doesn't specify. I chose AND because "filter" reads as narrowing and it's consistent with how the query and tag filter combine. **If you meant OR (match any tag), that's the one change to confirm** — swap the subset check `wanted <= item_tags` for an intersection `wanted & item_tags`.
- **Query and tags combine with AND** — tag filtering is applied *on top of* the title-search results, matching the brief's "filter search results by tags."
- **Tag matching is case-insensitive**, mirroring the existing case-insensitive title search.
- **Backward compatible** — `tags=None`/`[]` leaves results unchanged, so the three existing tests and all current positional `search(query)` callers are unaffected.

**One footgun I left in, deliberately:** `tags` must be an iterable of strings; passing a bare string (`tags="work"`) would iterate characters and misbehave silently. I kept it minimal rather than adding string-coercion, but say so in case you'd rather guard it.

## What's left to you

Two Trellis obligations I can't discharge myself: this needs an **independent review** (I wrote it, so I shouldn't be the one to certify it), and the **AND-vs-OR semantics is a decision a human should ratify** before this is considered done. There's no git repo here, so the reviewable artifact is the diff to `app.py` + `test_search.py`. Once you (or a reviewer) can run `pytest`, that's the last verification gate I couldn't close.
