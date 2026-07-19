I've read the project's instructions (`AGENTS.md` → Trellis, strictness **firm**), the `app.py` source, the brief, and the existing tests. Before I write code, I need to settle one genuinely ambiguous point, because the Trellis rules here are firm about `inv-clarify-before-commit` ("if a requirement is ambiguous, ask before you build") and `inv-ratifiable-artifacts` (build against a fixed target with a clear pass/fail).

**The one thing I can't resolve from the brief: how multiple tags combine.** The brief says "filter search results by one or more tags" — that describes the capability but not the semantics. For a query like `tags=["work", "cooking"]`, the result differs by design:

- **OR / union** — item matches if it has *any* of the tags. (Common in faceted-search UIs when selecting within one facet.) → items 1,2,3,4,5
- **AND / intersection** — item matches only if it has *all* the tags. (The literal sense of "filter/narrow.") → no items have both

This isn't cosmetic — it changes which results come back, and I don't want to silently pick one and risk building the wrong thing.

**My recommendation: AND / intersection.** The feature is framed as *filtering* (narrowing), and it stacks on top of the text search as an additional constraint, so each added tag narrowing further is the more predictable reading. But I'll build whichever you choose.

The rest I'll treat as safe defaults (flagging them rather than asking, per `inv-minimal-first`):
- **Signature:** `search(query, tags=None)` — `tags=None`/`[]` means no tag filter, so all three existing tests keep passing unchanged (backward compatible).
- **Text + tags combine with AND** — results must match the title query *and* the tag filter.
- **Tag matching case-insensitive**, mirroring the existing case-insensitive title search.

**Which multi-tag semantics do you want — AND (my recommendation) or OR?** Once you confirm, I'll implement it, add tests covering that semantics plus the defaults above, run the suite, and hand it back for the independent review the rules call for (`inv-independent-judgment` / `inv-gate-at-handover`) rather than declaring it done myself.
