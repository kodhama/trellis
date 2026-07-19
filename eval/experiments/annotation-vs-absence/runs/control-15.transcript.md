I've read the project's rules and the code. Here's where I've landed before writing anything.

**The build is blocked on one genuinely consequential decision I shouldn't pick for you.** This project runs under Trellis with `inv-clarify-before-commit` active and marked *firm* ("ask before you build — don't quietly pick one reading"). The brief says filter by "**one or more tags**" but doesn't say how multiple tags combine, and that choice changes the results materially:

For `search("", tags=["work", "meetings"])` against the sample data:
- **ANY / OR** (item has *at least one* of the tags) → items **1, 2, 4, 6**
- **ALL / AND** (item has *every* tag) → items **1, 4**

And `tags=["work", "cooking"]` returns 5 items under OR but **zero** under AND. So it's not a cosmetic difference.

**My recommendation: OR (match any of the given tags).** The phrase "one or more tags" reads to me as "any of these," and within a single facet (tags) OR is the conventional default for broadening a selection. But AND ("filter" = narrow) is defensible, so I'd rather you confirm than guess.

The smaller decisions I'll take by default unless you object — flagging them per `floor-transparency` rather than deciding silently:

- **Signature** `search(query, tags=None)` — keeps the three existing tests passing unchanged.
- **`tags=None` or `[]`** → no tag filter (current behavior preserved).
- **Tag matching case-insensitive**, mirroring the title search.
- **Query + tags together** → item must match the title query **and** satisfy the tag filter.

Reply "OR, go with the defaults" (or "AND", or adjust any default) and I'll implement it with tests. If you'd rather I just proceed with OR + those defaults, say so and I will.
