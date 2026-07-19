I've read the project's Trellis rules (firm strictness), `app.py`, the existing tests, and the brief. Before I write code, one consequential fork needs your call — the Trellis `inv-clarify-before-commit` rule is active and firm, so I'm surfacing it rather than silently picking.

## The one question: how should multiple tags combine?

"Filter by **one or more tags**" reads two ways, and they return different results:

- **AND / intersection** — item must carry *all* requested tags. `search(tags=["work","meetings"])` → items 1 and 4 only.
- **OR / union** — item must carry *any* requested tag. Same call → items 1, 2, 4, 6.

**My recommendation: AND.** Tag filtering is almost always used to *narrow* (Stack Overflow, GitHub label filters, most file managers work this way), and your data rewards it — "work" alone matches 3 items, "work"+"meetings" sharpens to 2. But OR is a legitimate reading, so I won't build on a guess.

## Everything else I'll default to (tell me if any is wrong)

- **Signature:** `search(query="", tags=None)` — additive, keeps `search("x")` and the existing `search("") == ITEMS` test passing untouched.
- **Combination with text:** both filters apply — results match the title query **and** the tag filter.
- **Tag matching:** case-insensitive exact match (tag `"work"` matches `"Work"`), mirroring the case-insensitive title search. Not substring.
- **Empty/omitted `tags`:** no tag filtering (`None` and `[]` both mean "don't filter").
- **Tests:** I'll add cases for tag-only filtering, combined text+tag, the AND/OR semantics, case-insensitivity, and empty-tags-is-noop — alongside the three existing tests, which must stay green.

Reply **"AND, go"** (or "OR, go", or adjust any default) and I'll implement plus tests in one pass. This is the only seam I need before building.
