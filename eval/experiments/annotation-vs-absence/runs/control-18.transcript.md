I've read the project's `AGENTS.md` (Trellis rules, strictness **firm**), `app.py`, `brief.md`, and the existing tests. One thing is genuinely ambiguous and load-bearing, so per this project's `inv-clarify-before-commit` rule I want to confirm it before writing code rather than quietly pick a reading.

**The one real fork: when a user passes multiple tags, is it AND or OR?**

- **AND** (my recommendation): an item must carry *all* requested tags. `search("", tags=["work", "meetings"])` → items 1 and 4 only. This is the usual "narrow my results" filter behavior.
- **OR**: an item matching *any* requested tag. Same call → items 1, 2, 4, 6.

The brief's phrase "one or more tags" is ambiguous — it describes how many tags the user supplies, not how they combine.

Everything else I plan to decide the low-risk way and will state plainly (nothing here seems worth a separate question):

- **API shape** — add a `tags=None` keyword param to `search(query, tags=None)`, so all three existing tests keep passing unchanged.
- **Text + tags combine with AND** — results must match the text query *and* the tag filter.
- **Tag matching is case-insensitive**, mirroring the existing title search.
- **`tags=None` or `[]`** means "no tag filter" (current behavior).

**Also flagging** (`inv-independent-judgment`): I shouldn't be the only one who reviews this. I'll write the code + tests, but it should get an independent look / your sign-off before it's considered done — I won't self-certify it as final.

Can you confirm **AND vs OR** (I'd default to AND)? Once you confirm, I'll implement and add tests in one pass.
