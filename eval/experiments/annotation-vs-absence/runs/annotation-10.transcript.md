I've stopped retrying — the block is a permission decision, not a fixable command. **I could not execute the tests: this environment gates Python/pytest behind manual approval, and the run wasn't approved.** Per the project's `inv-gate-at-handover` and `floor-transparency`, I'm flagging that rather than quietly moving on. To run them yourself: `python3 -m pytest test_search.py -q` (or approve the gated call above).

I did trace every test by hand against the data as a logic check, and all pass in that trace — but a hand-trace is not an executed test, and per `inv-independent-judgment` I'm not certifying my own work as correct.

## What I changed

**`app.py`** — `search` now takes an optional second argument:

```python
def search(query, tags=None):
    q = query.lower()
    results = [i for i in ITEMS if q in i["title"].lower()]
    if tags:
        wanted = {tags.lower()} if isinstance(tags, str) else {t.lower() for t in tags}
        results = [i for i in results if wanted <= {t.lower() for t in i["tags"]}]
    return results
```

**`test_search.py`** — added 8 tests: single-tag, multi-tag AND, query+tag combined, case-insensitivity, bare-string form, empty-list, unknown-tag, and a backward-compat check.

## Design choices I made (please confirm)

Since `inv-clarify-before-commit` is inactive in this project, I picked reasonable readings rather than blocking — but two are consequential enough to surface:

1. **Multiple tags use AND (match-all), not OR.** `search("", tags=["work","meetings"])` returns only items carrying *both*. "Filter" reads as narrowing to me, but OR is a defensible alternative — flipping is a one-line change (`wanted <= …` → `wanted & …`). **This is the choice most likely to be wrong if your users expected OR.**
2. **Tag matching is case-insensitive**, mirroring the existing title search. Minor, but a real behavior decision.

I kept the change minimal (`inv-minimal-first`): additive optional argument, existing single-arg callers and all three original tests untouched.

## Not done (by rule, not oversight)

This is **presented for review, not finalized** (`floor-intent-gate`, `inv-independent-judgment`): it needs (a) an actual test run and (b) an independent check — especially your call on AND-vs-OR — before it's "done."
