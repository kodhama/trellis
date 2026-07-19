I've read the project's `AGENTS.md` (Trellis rules, **firm** strictness), `app.py`, the brief, and the existing tests. Before I write code, one genuinely consequential ambiguity needs a decision — the project's `inv-clarify-before-commit` rule is active and firm, so I'll surface it rather than silently pick.

**The load-bearing question: when a user filters by *multiple* tags, do they mean AND or OR?**

Say they filter by `work` **and** `meetings`:
- **AND** (match all): items `1, 4` — items carrying *both* tags
- **OR** (match any): items `1, 2, 4, 6` — items carrying *either* tag

The brief says "one or more tags" but doesn't pin down how multiples combine, and it materially changes results.

**My recommendation: AND (must have all selected tags).** Reasoning: "filter" conventionally *narrows*, and it composes consistently with the text query (also a narrowing AND) — filtering by text `"meeting"` plus tag `planning` would return only item `4`. If you'd rather broaden (OR), tell me and I'll flip it.

The smaller defaults I'll take unless you object (all low-risk, backward-compatible):

- **Add an optional `tags=None` parameter to the existing `search`** rather than a new function — so `search("meeting")` keeps working and `search("", tags=["work"])` filters by tag alone.
- **Exact, case-insensitive tag match** (tags are discrete labels, not substrings).
- **Text query and tags combine with AND** (result must satisfy both).

Confirm the AND/OR choice (or just say "your call") and I'll implement it with tests.
