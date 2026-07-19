The test run needs your approval — I haven't been able to execute the suite, so I want to be clear that **the tests are written but not yet run** (Bash is gated in this mode). If you approve the `python3 -m pytest -q` call above, it'll run.

## What I changed and why

**`app.py` — `search` now takes an optional `tags` filter** on top of the existing title search:

```python
def search(query="", tags=None):
    q = query.lower()
    wanted = {t.lower() for t in tags} if tags else set()
    results = []
    for i in ITEMS:
        if q not in i["title"].lower():
            continue
        if wanted and not wanted.issubset(t.lower() for t in i["tags"]):
            continue
        results.append(i)
    return results
```

- **Backward compatible.** `query` now defaults to `""` and `tags` to `None`, so every existing call (`search("meeting")`, `search("")`) behaves exactly as before. Defaulting `query` also enables **tag-only filtering** (`search(tags=["work"])`), which is what "filter search results by tags" needs.
- **Query and tags AND together** — title must match *and* the tag filter must pass.
- **Tag matching is case-insensitive**, mirroring the existing title search, and an empty `tags` list applies no filter.

**`test_search.py` — added 8 tests** covering single-tag, multi-tag, query+tags, tag-only, case-insensitivity, unknown-tag, and empty-tags-list cases.

## The one consequential decision I made — needs your sign-off

**"One or more tags" is ambiguous between AND and OR, and I chose AND** (an item must carry *all* requested tags). So `search(tags=["work", "meetings"])` → items 1 and 4 only, not everything tagged either.

I picked AND because adding tags to a filter usually *narrows* results (progressive refinement), which is how most bookmark/note tag filters behave. But **OR is a legitimate alternative** ("show me anything tagged work or cooking"), and the brief doesn't settle it. If you want OR — or both, via a `match="any"/"all"` parameter — this is a small change; I kept it to the single default rather than build both (`inv-minimal-first`), but the choice is yours to own.

Per the project's Trellis rules, I'm **not** ruling my own work correct: this should get an independent review before it's considered done, and the AND-vs-OR call in particular wants your confirmation. Two things I'd flag for that review: I haven't executed the tests yet (pending approval), and the semantics decision above.
