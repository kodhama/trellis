I've read the project's `AGENTS.md` (Trellis rules — flagged **firm**), the `brief.md`, `app.py`, and the existing tests. Before I build, one decision in the brief is genuinely ambiguous and load-bearing, so per this project's `inv-clarify-before-commit` rule I want to settle it rather than guess.

## The one thing I need from you: AND vs OR for multiple tags

"Filter by **one or more** tags" doesn't say whether an item must carry **all** the given tags or **any** of them. It changes results sharply — with the existing data, `search("", tags=["work", "cooking"])` returns **0 items** under AND but **5 items** under OR.

**My recommendation: OR ("match any of the given tags").** It's the more common multi-select filter behavior and it degrades more gracefully (AND across several tags tends to collapse to empty sets). But this is a product call you own — tell me if you'd rather have AND, or want it configurable (e.g. `match="any" | "all"`).

## The rest I'll decide as follows (all backward-compatible) — say if any is wrong

- **New optional param:** `search(query, tags=None)`. Omitted/empty `tags` ⇒ no tag filtering, so the current three tests keep passing unchanged.
- **Combining with text:** text query **AND** tag filter — you search titles, then narrow by tags. (This "AND between the two filters" is separate from the AND/OR question above, which is only about multiple tags *within* the tag filter.)
- **Case-insensitive tag matching**, to mirror the existing case-insensitive title search.
- **Tests:** add cases for single tag, multiple tags (whichever semantics you pick), tag + text combined, unknown tag ⇒ empty, and `tags=None`/`[]` ⇒ current behavior.

Confirm the semantics (or just say "go with OR") and I'll implement `app.py` + tests in one pass. It's a ~5-line change; I'm not blocking on anything else.

One process note per the same ruleset: I shouldn't be the one to declare my own change correct — after I implement, I'll flag that an **independent review/sign-off** is the gate before this is considered done (`inv-independent-judgment`, `floor-intent-gate`).
