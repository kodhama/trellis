No decision docs or prior specs for this feature — the ground is clean. There's **one genuinely consequential ambiguity I need to settle before building**, per this project's rules (which firmly require asking rather than quietly picking a reading).

**When a caller passes multiple tags, should an item match ALL of them (AND / narrowing) or ANY of them (OR / widening)?** The brief's "one or more tags" says how many tags a caller can supply, not how to combine them — and the two readings give materially different results. Example, filtering to tags `["work", "planning"]`:

- **AND** → items 2, 4 (only those carrying both)
- **OR** → items 1, 2, 4 (those carrying either)

**My recommendation: AND (narrowing).** It's the dominant convention for tag filters (GitHub labels, Gmail, faceted search all narrow with AND), and "add a filter" typically means "show me less, more specific." Reply "AND", "OR", or "your call" and I'll build it.

The smaller decisions I've already resolved (and will note in the code), so you don't have to weigh in unless you disagree:

- **Backward-compatible signature** — add an optional `tags=None` parameter so the three existing tests (`search("meeting")`, etc.) keep passing unchanged. Query and tags combine (AND): text match *and* tag match.
- **Case-insensitive tag matching**, mirroring the existing case-insensitive title search.
- **No tags passed → tags don't filter** (pure text search, today's behavior).

I'll implement, add tests, and run the suite once you confirm the AND/OR call. I won't mark it done myself — the rules here require an independent review before this is considered finished, and I'll flag that at handover.
