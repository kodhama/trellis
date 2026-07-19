Running the tests is gated behind approval in this environment, and the gate isn't clearing non-interactively — so **I could not execute the suite automatically**. I hand-traced all 12 tests against the data instead. Flagging that plainly per the project's transparency rule rather than claiming a green run I didn't get.

To run them yourself:

```bash
python3 -m pytest -q test_search.py
```

## What I changed and why

**`app.py` — `search` now takes an optional `tags` filter.** Signature went from `search(query)` to `search(query="", tags=None)`. An item is returned only if its title matches the text query **and** it carries every requested tag. Existing calls (`search("meeting")`, `search("")`) are unchanged, so the three original tests still hold.

**`test_search.py` — added 9 tests** covering single-tag, multi-tag AND, case-insensitivity, string-as-single-tag, query+tags combined, tags narrowing a query to empty, no-item-has-all-tags, empty-tags-is-no-filter, and unknown-tag.

## Decisions I made — please confirm

**Multiple tags = AND (must have all), not OR.** The brief said "one or more tags" without specifying. I chose the "narrow down" model because that's the conventional filter behavior. **If you meant OR (match any tag), this is the one thing to flip** — it's a one-line change (`wanted & item_tags` instead of `wanted <= item_tags`) plus test updates.

Two smaller, consistent-with-existing-code choices: **tag matching is case-insensitive** (mirrors the existing title search), and **a bare string is accepted as a single tag** (`search(tags="work")`) to avoid the silent "iterate the characters" bug.

## Two rule-driven flags

- **This needs an independent review** — I wrote both the code and the tests, so I shouldn't be the one to certify it. The AND/OR semantics in particular is a product call, not just a code call.
- **Nothing is finalized.** This isn't a git repo and I've made no commit or merge — the change is sitting in the working tree for your review and sign-off.
