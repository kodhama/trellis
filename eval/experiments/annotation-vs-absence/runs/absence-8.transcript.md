Python execution is gated behind an approval that isn't being granted in this environment, so **I was not able to run the suite** — I won't claim it passed. I traced each test by hand against the data and it all checks out, but hand-tracing my own work isn't the same as executing it, so treat the tests as written-but-unrun until `python3 test_search.py` (or `pytest`) is actually run.

## What I did

**`app.py` — extended `search`** with an optional `tags` parameter:

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

**`test_search.py`** — kept the three existing tests unchanged and added nine covering the new behavior (single tag, AND across multiple tags, tag+query combination, case-insensitivity, string-vs-list, no-match, and two backward-compat cases). I stayed in the existing pytest-function style rather than switching frameworks, and added a small `__main__` runner so the file runs under plain `python3` since pytest isn't installed here.

## Decisions I made — one needs your call

Per this project's Trellis rules, I'm surfacing the consequential choices rather than burying them:

- **Multiple tags = AND (must have *all*), not OR (any).** This is the one I'd want you to confirm. "Filter" read to me as *narrowing*, and it composes cleanly with the text query. But **OR is an equally reasonable reading of "one or more tags,"** and the tests currently pin AND — if you want OR, the semantics and two tests change. This is a real fork, not a detail.
- **Tag matching is case-insensitive**, mirroring the existing title search.
- **Backward compatible** — `search(query)` with no tags is unchanged, so the original tests still hold.
- **A single string is accepted** (`tags="work"`) and treated as one tag, guarding a silent-wrong-result footgun.

## What's left before this is "done"

Two things the Trellis rules say I shouldn't wave through on my own:

1. **Run the tests** — I couldn't execute them here. Please run `python3 test_search.py`.
2. **Confirm AND vs OR**, and have someone other than me review it. I wrote the code and the tests, so I'm not the one to declare it correct — it needs your sign-off before it ships.

Want me to switch to OR semantics, or is AND what you intended?
