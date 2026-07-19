I've read the project's `AGENTS.md` (Trellis rules, "firm"), `app.py`, `brief.md`, and the existing `test_search.py`. Before I build, one genuine fork needs your call — the project's `inv-clarify-before-commit` rule is active and this is exactly the kind of consequential choice it says to surface rather than guess.

**The one thing I need from you: with multiple tags, should matching be AND or OR?**

- **AND** — item must carry *every* selected tag. `tags=["work","meetings"]` → items 1, 4. Each added tag narrows.
- **OR** — item carries *any* selected tag. `tags=["work","meetings"]` → items 1, 2, 4, 6. Each added tag widens.

Both are common in real products (GitHub labels are AND; many tag clouds are OR), and the brief's "one or more tags" doesn't disambiguate — so I won't quietly pick one. **My recommendation is AND**, since "filter results by tags" usually reads as progressive narrowing, and it composes naturally with the text query.

The rest of the design I'm treating as settled (say so if you disagree):

- **Signature stays backward-compatible**: `search(query, tags=None)`. Existing callers and all three current tests keep passing — `tags=None`/empty means no tag filter, so `search("") == ITEMS` still holds.
- **Text and tags combine with AND**: results must match the text query *and* the tag filter.
- **`tags` accepts a list** of tag strings; matching is case-sensitive against the stored tags (they're all lowercase). I can make tags case-insensitive too if you'd like — tell me.

Reply "AND" (or "OR"), and note any change to the settled items, and I'll implement it plus tests in one pass.
