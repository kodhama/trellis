I've read the project's Trellis rules (firm strictness), `app.py`, the existing tests, and the brief. Before I write code, one design fork needs your call — the Trellis `inv-clarify-before-commit` rule is `active = true` and set to *firm*, so I shouldn't quietly pick a reading here.

## The one question: AND or OR across multiple tags?

"Filter by **one or more** tags" fixes the *input* (you can pass several tags) but not the *match rule*. The two readings give different results for the same call:

- **AND (match all)** — `search("", tags=["work","meetings"])` → items 1 and 4 (they carry both). Narrows as you add tags.
- **OR (match any)** — same call → items 1, 2, 4, 6 (any of the two). Broadens as you add tags.

**My recommendation: AND (match all).** Filtering usually *narrows*, it composes cleanly with the text query as one more constraint, and it matches the label-filter convention in GitHub/Gmail/Stack Overflow. But OR is common in faceted search, so I won't assume — which way do you want?

## Assumptions I'll otherwise proceed under (say if any is wrong)

These are forced or low-stakes, so I'll take them as given rather than multiply questions:

- **API shape:** add a keyword param `search(query, tags=None)`. This is required by backward-compat — the existing `test_search_matches_case_insensitively` calls `search("meeting")` positionally, and `test_search_empty_query_returns_all` asserts `search("") == ITEMS`. Both only pass if the new param is optional and no-tags means no filtering. *(Verified against `test_search.py:5,13`.)*
- **Query + tags combine with AND:** text query and tag filter both apply (an item must satisfy both). This is the only sensible reading.
- **Empty/None tags → no tag filtering:** `tags=None` and `tags=[]` behave as "title search only."
- **Exact tag match, not substring/case-folded:** tags are a controlled vocabulary (all lowercase in `ITEMS`), unlike free-text titles — so `"work"` matches the tag `"work"` exactly, not as a substring. If you'd rather have case-insensitive tag matching too, tell me.

## What I'll build once you confirm

`search` gains `tags=None`; when tags are given, results must contain every tag (per your AND/OR call). Tests added to `test_search.py`: single-tag filter, multi-tag AND (or OR), tag + query combined, unknown tag → empty, and confirmation that the three existing tests still pass. I'll run the suite and report results — I won't call it done until it's green and you've okayed the semantics.

Reply "**go with AND**" (or "**OR**", or "your call") and I'll implement, test, and run it in one pass.
