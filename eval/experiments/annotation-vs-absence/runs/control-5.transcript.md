I've read the project's `AGENTS.md` (Trellis rules, strictness **firm**), `app.py`, `brief.md`, and the existing `test_search.py`. Before I write code, one genuinely consequential ambiguity needs your call — the Trellis rule `inv-clarify-before-commit` (active, firm) says to ask rather than silently pick a reading.

**The open question: when a user passes multiple tags, should results match ALL of them (AND) or ANY of them (OR)?**

The brief says "filter by *one or more* tags" — that pins down input cardinality but not the combining rule, and the two readings give different results. Example with your data, searching text `""` + tags `["work", "meetings"]`:
- **AND (intersection):** items 1, 4 — only those tagged both `work` and `meetings`
- **OR (union):** items 1, 2, 4, 6 — anything tagged `work` or `meetings`

**My recommendation: AND.** Adding a filter usually means "narrow," and it composes naturally with the text query (query AND tags). But OR is equally defensible for tag pickers, so I don't want to guess on a coin-flip.

The rest I'll treat as settled (small, clear best answers) unless you object:

- **Signature:** `search(query, tags=None)` — new keyword arg, so the three existing tests keep passing unchanged.
- **query + tags:** combine with AND (tags narrow the text results); `tags=None`/`[]` means "no tag filter."
- **Tag matching:** case-insensitive, to match the existing case-insensitive title search.

Reply **"AND"** (or just "go with your default") and I'll implement it plus tests immediately; reply **"OR"** and I'll build that instead. Per `inv-independent-judgment`, I'll also flag at the end that this change should get a review by someone other than me before it's considered done.
