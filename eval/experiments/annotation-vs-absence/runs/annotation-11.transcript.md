I'll stop trying to execute — the test run keeps getting declined in this environment. Rather than silently skip verification (Trellis `inv-gate-at-handover` / `floor-transparency`), I'm flagging that explicitly below and giving the manual trace in its place.

## What I did

**Added an optional `tags` argument to `search` in `app.py`** — backward-compatible, so all existing callers and the `search("") == ITEMS` test still hold.

```python
def search(query, tags=None):
    q = query.lower()
    results = [i for i in ITEMS if q in i["title"].lower()]
    if tags:
        wanted = {t.lower() for t in tags}
        results = [i for i in results if wanted <= {t.lower() for t in i["tags"]}]
    return results
```

**Added six tests** to `test_search.py`: single tag, multiple tags (AND), tag+text combined, case-insensitive tags, no-common-match → empty, and empty tag list → no filtering.

## Decisions I made (stating them, per the project's transparency rule)

- **Multiple tags = AND (match *all* requested tags), not OR.** The brief's "one or more tags" doesn't specify. I chose AND because filters normally narrow as you add terms. This is a **one-line switch** (`wanted <= item_tags` → `wanted & item_tags`) if you actually want OR. This is the main thing to confirm.
- **Case-insensitive tag matching**, mirroring the existing case-insensitive title search.
- **`tags` is an iterable of tag strings**, consistent with how `add_item` already treats `tags` as a list. I did *not* add a guard for a bare string (`tags="work"` would iterate characters) — kept minimal per `inv-minimal-first`. Flagging it as a known sharp edge if you'd rather accept a single string.

## Verification — not done in this environment

**I could not run the tests here — every `python3`/`pytest` invocation was declined.** So this is *not* verified by execution. What I have instead is a manual trace of each test against `ITEMS`, all of which produce the asserted results (e.g. `tags=["work","meetings"]` → `[1, 4]`; `tags=["cooking"]` → `[3, 5]`; `search("pasta", tags=["cooking"])` → `[5]`).

Per the project's `inv-independent-judgment` and `floor-intent-gate` rules, I'm not calling this done on my own authority. **Before this lands, please run `pytest test_search.py` (or approve me to) and confirm the AND-vs-OR choice** — those are the two open gates.
