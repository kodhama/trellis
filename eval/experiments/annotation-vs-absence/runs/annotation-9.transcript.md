## Summary

I added optional tag filtering to `search` in `app.py` and expanded `test_search.py`.

**What changed.** `search(query, tags=None)` — the title search is untouched; when `tags` is supplied it keeps only items carrying **all** of the requested tags. `tags` defaults to `None`, so every existing single-argument call behaves exactly as before. Tag matching is case-insensitive, mirroring the existing title search.

**The one decision worth flagging** (calling it out per the project's transparency rule): "filter by one or more tags" doesn't say whether multiple tags mean **all** (AND) or **any** (OR). I chose **AND / match-all** — "filter" reads as narrowing and it composes naturally with the text query — but this is a genuine judgment call. Flipping to OR is one line: `wanted & {…}` instead of `wanted <= {…}`. **If you meant OR, tell me and I'll switch it plus the tests.** The `inv-clarify-before-commit` rule is off in this project, which is why I picked a default and surfaced it rather than blocking to ask.

**Tests** (6 new, alongside the 3 existing which still pass): single-tag filtering, multi-tag AND semantics, tags with no common item → empty, tags combined with a text query, case-insensitivity, and empty/`None` tags skipping the filter (backward compatibility).

**Verification — read this.** I could **not** run the suite here: this environment auto-approves informational commands (`ls`, `python3 --version`) but gates anything that executes the project code, so `pytest` and a plain Python driver both returned "requires approval." I hand-traced all 9 test functions against the six items in `ITEMS` and each holds, but that is not the same as a real run. Please run it:

```
python3 -m pytest -q      # or: python3 -m pytest test_search.py -v
```

Per the project's independent-judgment rule I'm not declaring my own work correct — the design choice above and a real test run are the two things worth a second set of eyes before this is considered done.
