Execution of anything that imports `app` is gated in this environment (`python3 --version` runs; `python3 -m pytest` and `python3 -c "import app…"` return "requires approval"), so **I was not able to run the test suite here** — I'm flagging that rather than letting it silently not happen. I hand-traced every assertion against the data in `app.py`, which is what caught the wrong expectation above.

## What I did

**Added tag filtering to `search` in `app.py`** as an additive, backward-compatible change — new optional `tags` parameter, existing `search(query)` calls behave exactly as before.

```python
def search(query, tags=None):
    q = query.lower()
    wanted = {t.lower() for t in (tags or [])}
    results = []
    for i in ITEMS:
        if q not in i["title"].lower():
            continue
        if wanted and not wanted <= {t.lower() for t in i["tags"]}:
            continue
        results.append(i)
    return results
```

**Added six tests** in `test_search.py` covering single-tag filtering, multi-tag AND, query+tag combination, case-insensitivity, empty-tag-list = no filter, and unknown-tag = empty. The three original tests are untouched and still hold.

## The one decision you own

**When multiple tags are given, an item must carry *all* of them (match-all / AND), and tags combine with the title query also by AND.** The brief ("filter by one or more tags") doesn't pin down all-vs-any, so I chose the narrowing reading to match how `search` already narrows by title — but **this is the load-bearing judgment call and it's yours to confirm.** If you want match-*any* (OR) instead, it's a one-line change: `if wanted and not (wanted & {t.lower() for t in i["tags"]})`. I also made tag matching case-insensitive to mirror the existing title search; say if you'd rather it be exact.

## Not done, on purpose

- **The test run is unverified by a runner** — please run `python3 -m pytest test_search.py`. I traced the logic by hand but did not execute it here.
- **Per the project's rules, I'm not treating this as final.** I wrote both the code and its tests, so it needs an independent review, and the AND/OR semantics needs your sign-off before this is merged — I haven't committed or shipped anything.
