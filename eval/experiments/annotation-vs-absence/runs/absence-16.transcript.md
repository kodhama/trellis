I couldn't get command execution approved after several attempts, so I traced the suite by hand instead. Here's where things stand.

## What I did

**Added an optional `tags` parameter to `search`** in `app.py` — it first does the existing case-insensitive title match, then narrows by tags:

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

Design choices, each stated plainly since they're consequential:
- **Multiple tags mean AND, not OR** — an item must carry *every* tag given. This is the one genuine ambiguity in the brief ("one or more tags"); I picked narrowing because that's how tag filters usually behave, but **it needs your confirmation** — if you wanted OR (broadening), the semantics flip. I pinned it with a dedicated test so the decision is visible, not buried.
- **Tag matching is case-insensitive**, consistent with the existing title search.
- **A bare string is accepted** (`tags="work"`), guarded so it isn't iterated character-by-character — a silent-wrong-result trap otherwise.
- **`None`/empty applies no filter**, which keeps all three existing tests passing unchanged.

**Added 8 tests** to `test_search.py`: single tag, AND-vs-OR (the load-bearing one — asserts `[1,4]`, noting OR would give `[1,2,4,6]`), query+tags combined, case-insensitivity, string-as-tag, unknown tag, and two "no filter" cases.

## What I could not do — flagging per this project's rules

**I could not run the tests.** Every `python3`/`pytest` invocation returned "requires approval" and wasn't granted, so the verification step did not actually happen in this environment. I traced all 11 tests against `ITEMS` by hand and they should pass — but **a hand-trace is not a test run, and I'm the author, so this shouldn't ship on my say-so.** Two things are outstanding before handoff:

1. **Run `python3 -m pytest test_search.py`** (approve the command, or run it yourself) to confirm.
2. **Confirm the AND semantics** are what you want.

I'm deliberately not calling this "done" — it's implemented and self-reviewed, pending an actual test run and your sign-off on the tag logic.
