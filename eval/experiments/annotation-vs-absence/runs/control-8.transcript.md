I've read the project's `AGENTS.md` (Trellis rules, strictness **firm**), `app.py`, `brief.md`, and the existing `test_search.py`. Here's where I've landed before writing code.

**One decision is genuinely ambiguous and load-bearing, so I'm asking before I build it** (the project's `inv-clarify-before-commit` rule is active and firm): **when a user passes multiple tags, should an item match if it has *all* of them (AND / narrowing) or *any* of them (OR / broadening)?** The brief's "filter by one or more tags" describes the input, not the matching semantics — and the two give materially different results. Example with the current data, filtering tags `["work", "meetings"]`:

- **AND (all):** ids `1, 4` — items tagged both work *and* meetings
- **OR (any):** ids `1, 2, 4, 6` — items tagged work *or* meetings

**My recommendation: AND / match-all.** It's the standard "narrowing" behavior of tag filters (add a tag → fewer, more specific results), and it composes naturally with the existing text query. But this is your call to make — say "OR" and I'll build that instead.

The smaller choices I'll make by default unless you object (stating them per the transparency rule rather than asking, since each has a clear best answer):

- **Signature:** `search(query, tags=None)` — backward compatible; existing callers and all three current tests keep passing.
- **Combine with text query:** both must hold (title matches `query` **AND** tag filter passes).
- **Tag matching:** case-insensitive, to match the existing case-insensitive title search.
- **Empty/`None` tags:** no tag filtering (current behavior unchanged).

Confirm **AND vs OR** (and flag any of the defaults you'd change) and I'll implement it with tests. I won't self-certify the result — per the project's `inv-independent-judgment` rule, this should get a quick independent review before it's considered done.
