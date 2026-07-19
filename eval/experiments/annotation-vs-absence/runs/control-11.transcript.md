I've read the project's `AGENTS.md` and its Trellis rules (activation rows in `.trellis/rules.toml`), plus `app.py`, the existing tests, and `brief.md`. I have what I need — I won't pull in more.

Before I write code, one genuinely load-bearing ambiguity needs your call, because the project's rules here are firm about **clarifying before building** (`inv-clarify-before-commit`) and building against an agreed definition of done (`inv-ratifiable-artifacts`). Silently picking would risk building the wrong thing.

**The one real question: when a user passes multiple tags, is it AND or OR?**

- **OR (match any):** `tags=["work","cooking"]` returns items tagged work *or* cooking. This is the usual faceted-search convention for values within a single facet.
- **AND (match all):** same input returns only items tagged *both* work and cooking — a narrower "filter down" reading.

These produce different results and different tests, and "filter by one or more tags" doesn't disambiguate it. **My recommendation is OR (match any)** — it's the more common tag-filter default and reads most naturally as "show me things in these tags." But I want your confirmation rather than guess.

Here's the rest of the design I'd build once you confirm semantics — flagging these as my intended defaults so you can veto any in the same pass:

- **Signature:** `search(query="", tags=None)` — `tags` optional, defaulting to no tag filter. This keeps the three existing tests (`search("")`, `search("meeting")`) passing unchanged.
- **Combination with text:** query and tags are ANDed — an item must match the text query *and* the tag filter. (The AND/OR question above is only about tags *among themselves*.)
- **Tag matching:** case-insensitive, to match how title search already works.
- **Input shape:** `tags` is a list of tag strings (mirrors the `tags` list on items); empty list behaves like `None` (no filter).
- **Tests:** I'll add cases for single tag, multiple tags (encoding whichever semantic you pick), tags combined with a text query, a tag with no matches, and that the default `tags=None` leaves existing behavior intact.

Confirm **OR vs AND** (or just say "go with your recommendation") and I'll implement it and run the tests. I won't treat my own implementation as reviewed — I'll flag it for your check when it's done (`inv-independent-judgment`, `floor-intent-gate`).
