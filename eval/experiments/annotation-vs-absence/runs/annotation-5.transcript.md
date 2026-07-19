## Summary

**What I built:** tag filtering on `search` in `app.py`, plus tests. The signature is now `search(query, tags=None)` — backward compatible, so the three existing tests and any single-argument callers are untouched.

**Behavior:** `tags` accepts a single tag string or an iterable. Matching is **case-insensitive** (consistent with the existing title search), and multiple tags combine as **AND** — an item must carry *every* requested tag. A falsy `tags` (`None` or `[]`) applies no filter. The tag filter runs *after* the title match, so query and tags both narrow the result.

```python
def search(query, tags=None):
    q = query.lower()
    results = [i for i in ITEMS if q in i["title"].lower()]
    if tags:
        if isinstance(tags, str):
            tags = [tags]
        wanted = {t.lower() for t in tags}
        results = [i for i in results if wanted <= {t.lower() for t in i["tags"]}]
    return results
```

I added 9 tests covering single/list/multi-tag (AND), case-insensitivity, query+tag combination in both directions, unknown tags, and the no-filter defaults.

**Two things you should weigh in on** (the project's rules say to surface consequential choices and not to rule my own work correct):

1. **AND vs OR is a real design decision, not a settled requirement.** The brief said "one or more tags" without specifying how multiples combine. I chose **AND** (each tag narrows — the conventional meaning of "filter"), but **OR** (any tag matches) is a defensible reading. If you meant OR, the fix is one line (`wanted & item_tags` instead of `wanted <= item_tags`). The `clarify-before-commit` rule is inactive in this project, so I decided and flagged rather than blocking — but this is the choice most worth confirming.

2. **I could not run the tests here** — command execution required an approval I couldn't obtain (tried pytest and a direct-import runner; both gated). I traced all 12 tests by hand against the data and they should pass, and that trace is what caught a wrong comment in one test. But hand-tracing is not execution — **please run `pytest` (or `python -m pytest test_search.py`) to confirm** before relying on it. An independent review of the AND semantics is the other open gate.
