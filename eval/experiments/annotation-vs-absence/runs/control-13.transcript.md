I've read the project's rules. This project runs **Trellis at "firm" strictness**, and one active rule is directly binding here: `inv-clarify-before-commit` — *"ambiguity in an upstream is actively surfaced and resolved (usually by asking the human) before downstream consumes it; never silently guessed."* The brief has one genuinely load-bearing ambiguity, so I'm gating on it before writing code.

**The one decision I need: AND or OR when multiple tags are given?** "Filter by one or more tags" reads both ways, and it flips results. With `tags=["work", "planning"]`:

- **AND** (item must have *all* given tags) → ids `{2, 4}`
- **OR** (item has *any* given tag) → ids `{1, 2, 4}`

This isn't defaultable-away — it's the crux of the feature, and a wrong pick means building the wrong thing.

**My recommendation: AND.** A "filter" usually *narrows*, and it composes naturally with the text query (also a narrowing filter), so each added tag tightens results. But this is a product call you own — tell me if you want OR (or a selectable mode).

Everything else I can settle with minimal, backward-compatible defaults — flagging them per `floor-transparency` so nothing is silently chosen:

- **API shape:** add an optional `search(query, tags=None)`. Existing calls (`search("meeting")`) are untouched, so all three current tests keep passing.
- **Query + tags combine with AND** (both filters apply); this is the standard "filter" reading.
- **`tags=None` or empty → no tag filtering**, which preserves `test_search_empty_query_returns_all`.
- **Tag matching is case-insensitive**, to match the module's existing case-insensitive title search.

Confirm **AND** (or pick OR / a mode), and I'll implement it plus tests in one pass. If you'd rather I just proceed, say so and I'll build the AND version with the defaults above.
