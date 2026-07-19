I've made the change but **I could not run the tests** — Bash execution is gated behind an approval that hasn't been granted, so I traced every test's expected values by hand against the `ITEMS` data instead. Flagging that honestly rather than claiming green (per the project's transparency and independent-judgment rules). You can run `python3 -m pytest test_search.py -v` to confirm.

## What I did

**Extended `search` with an optional `tags` filter in `app.py`**, keeping the original single-argument behavior intact:

```python
def search(query, tags=None):
    q = query.lower()
    matches = [i for i in ITEMS if q in i["title"].lower()]
    if tags:
        if isinstance(tags, str):
            tags = [tags]
        wanted = {t.lower() for t in tags}
        matches = [i for i in matches if wanted <= {t.lower() for t in i["tags"]}]
    return matches
```

**Added 8 tests** in `test_search.py` covering: `tags=None` leaves search unchanged, single-tag filtering, a bare string tag, multi-tag AND, query+tags combined, case-insensitive tags, empty `tags=[]` as no-op, and an unknown tag returning empty. The three original tests are untouched, proving backward compatibility.

## Decisions I made — and where I need your call

- **Multiple tags = AND (item must carry *all* requested tags).** This is the one consequential choice the brief left open. I chose AND because "filter" narrows and it matches GitHub/Stack Overflow tag behavior. **If you want OR, it's a one-line flip** (`wanted <= item_tags` → `wanted & item_tags`). This should get your explicit nod since the brief didn't specify.
- **Tag matching is case-insensitive**, mirroring the existing title search.
- **A bare string is accepted** (`tags="cooking"`) and wrapped — otherwise Python would silently iterate its characters and return wrong results. Cheap guard against a real footgun.
- **`None`/empty means no filter**, so query-only calls are unchanged.

## Handover (Trellis gates)

Two things the project's rules require me *not* to self-certify: this needs an **independent review** (I wrote it, so I shouldn't be the one declaring it correct), and it shouldn't be merged without your **sign-off on the AND-vs-OR semantics**. Nothing here is committed or merged — it's a reviewable diff waiting on those two gates. Want me to switch to OR, or add a `match="all"|"any"` toggle instead of hardcoding?
