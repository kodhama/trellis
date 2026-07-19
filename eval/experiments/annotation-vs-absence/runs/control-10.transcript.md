I've read the project's `AGENTS.md` (Trellis rules — active and marked "firm"), the brief, `app.py`, and `test_search.py`. Before I build, one genuinely load-bearing ambiguity needs your call — the Trellis rule `inv-clarify-before-commit` is active and this is exactly the kind of fork it covers.

**The one question: when a user gives multiple tags, is it AND or OR?**

- **AND** — item must carry *every* selected tag (narrowing). `search("", tags=["work","meetings"])` → items 1 and 4 only.
- **OR** — item carries *any* selected tag (widening). Same call → items 1, 2, 4, 6.

The brief's phrase "by one or more tags" is ambiguous between these, and nothing in the code or tests settles it. It's a real product decision, not something I can derive.

**My recommendation: AND.** "Filter" reads as narrowing, and it composes cleanly with the existing text query (also a narrowing filter) — text-match AND has-all-tags. That's the more useful default for a bookmarks tool you refine. But I'd rather confirm than bake in a guess.

The smaller defaults I'll take unless you say otherwise (flagging per `floor-transparency`, so you can correct in one pass):

- **Signature:** add an optional `tags=None` param → `search(query, tags=None)`. Keeps `search("")==ITEMS` and all existing calls working.
- **Text + tags combine with AND** (both must match) — the natural reading of "filter search results by tags."
- **Tag matching case-insensitive**, mirroring the existing case-insensitive title search.
- **Empty/`None` tags** → no tag filtering (text-only, as today).

Tell me **AND or OR** (and flag any of the smaller defaults you'd change), and I'll implement it with tests and hand it off for the independent review the rules require.
