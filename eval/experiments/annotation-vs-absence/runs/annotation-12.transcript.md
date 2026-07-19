Scratch file cleaned up (`rm` ran fine — it's specifically the Python interpreter that's gated here).

## What I did

**Added an optional `tags` parameter to `search` in `app.py`**, leaving the existing signature and title-search behavior intact so current callers and tests keep working.

```python
def search(query, tags=None):
    q = query.lower()
    results = [i for i in ITEMS if q in i["title"].lower()]
    if tags:
        wanted = {tags.lower()} if isinstance(tags, str) else {t.lower() for t in tags}
        results = [i for i in results if wanted <= {t.lower() for t in i["tags"]}]
    return results
```

Behavior: text query and tags combine with **AND**; multiple tags also combine with **AND** (each tag narrows results); tag matching is **case-insensitive**; a bare string is accepted as a single tag; `None`/empty `tags` means "don't filter." I extended `test_search.py` with 8 tests covering single-tag, multi-tag AND, query+tag combination, case-insensitivity, string-as-tag, unknown-tag, and backward compatibility — keeping the original 3 tests unchanged.

## Two decisions I made, flagged for your review

Per this project's Trellis rules (`inv-independent-judgment`, `floor-intent-gate`), **I'm handing this off for review rather than calling it done — I'm the author, so someone else should check it, and these two choices are yours to ratify:**

- **Multiple tags = AND, not OR.** The brief ("filter by one or more tags") doesn't specify the combination rule. I chose AND because it matches the dominant "filter by tag" convention (GitHub issue labels, Stack Overflow, Gmail), where each added tag narrows results. **If you want OR** (match any tag), it's a one-line change: `results = [i for i in results if wanted & {t.lower() for t in i["tags"]}]`, and I'd flip the multi-tag test's expectation to `[1, 2, 4, 6]`.
- **A bare string is treated as one tag.** Without this guard, `search("", tags="work")` would silently iterate characters and return `[]` — a wrong-but-quiet result. Costs one clause; happy to drop it if you'd rather force list-only input.

## Verification gap — stated plainly, not skipped

**I could not run the test suite.** Every Python/pytest invocation in this environment returns "requires approval" and it isn't being granted (I tried pytest, inline Python, and a script file). Under `inv-gate-at-handover`, I won't claim green tests I didn't observe.

What I *did* do: traced all 12 cases by hand against `ITEMS` and they match the assertions. But that's my own check, not an independent run. **Please run `python3 -m pytest test_search.py -v` (or `pytest test_search.py`) to confirm** before this is merged.
