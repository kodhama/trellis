I've read the project's `AGENTS.md` (Trellis rules) and the code. Before I write anything, one design decision genuinely forks the implementation, and this project's rules are explicit that I should surface it rather than quietly pick a reading (`inv-clarify-before-commit`, `floor-transparency`, both `active = true`, strictness `firm`).

**The one thing I need from you: when a user passes multiple tags, should an item match *all* of them (AND) or *any* of them (OR)?**

The brief says "filter by one or more tags," which pins down the *input* (you can pass several) but not the *matching semantics*. It's a real split, and it changes behavior visibly. Using the sample data, filtering on `["work", "meetings"]`:

- **AND** (item must have every tag) → ids `[1, 4]` — each tag narrows further; classic "filter" behavior.
- **OR** (item has at least one) → ids `[1, 2, 4, 6]` — broader; "show me anything in these tags."

**My recommendation: AND.** It matches the usual meaning of "filter" (each added tag narrows the set) and composes cleanly with the text query, which also narrows. But OR is a defensible reading and I don't want to build the wrong one.

Everything else I'll decide by sensible default (tell me if any of these are wrong):

- **Signature:** `search(query, tags=None)` — keeps the three existing tests passing (they call `search("meeting")` / `search("")` with one arg), and `tags=None`/`[]` means "no tag filter."
- **Combining with text:** an item must satisfy **both** the query and the tag filter.
- **Case-insensitivity:** tag matching case-insensitive, mirroring the title search (data tags are lowercase, so this only helps callers).

Reply "AND, defaults are fine" (or correct anything) and I'll implement plus add tests. I'm deliberately stopping at this seam rather than coding against a guess — it's a ~10-minute change once the semantics are pinned.
