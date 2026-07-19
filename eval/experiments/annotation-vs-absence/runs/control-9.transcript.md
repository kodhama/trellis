I've got what I need. There's no spec/decisions artifact under `.trellis/` — just the internal rules docs — so the "spec" for this task is `brief.md` plus your prompt. I won't over-read the internals (`inv-bounded-context`).

Here's where I've landed. Most of this task is unambiguous, but **one decision genuinely changes behavior, and per this project's firm `inv-clarify-before-commit` rule I want to confirm it before I build rather than silently pick a reading.**

**The crux: when a user supplies multiple tags, match items that have _all_ of them (AND) or _any_ of them (OR)?**

The brief says "filter by one or more tags" — but "one or more" describes how many tags the user passes, not how to combine them. The two readings give different results. Example, filtering by `["work", "planning"]` against the current data:
- **AND (all):** items 2, 4
- **OR (any):** items 1, 2, 4

My recommendation: **AND / match-all**, because "filter" reads as *narrowing* a result set — each added tag makes the results more specific, which is the usual mental model for a tag filter (GitHub labels, file/photo taggers). OR is the reasonable alternative if you're thinking faceted-search "broaden within a facet." I genuinely see it as a real fork, not a formality.

**Everything else I'll settle without asking** (all backward-compatible and minimal, per `inv-minimal-first`):

- **Signature:** `search(query, tags=None)` — keeps the three existing tests and any current callers working; `tags=None`/empty means no tag filtering.
- **Composition:** tags *narrow* the text-search results (query still applies alongside).
- **Case:** tag matching case-insensitive, mirroring the existing case-insensitive title search (the data is lowercase anyway).

If you reply **AND**, **OR**, or "your call" (I'll take that as AND), I'll implement it and the tests in one pass immediately — the code is a one-line filter either way, and I'll add tests covering: single tag, multiple tags (locking in the AND-vs-OR behavior), tags combined with a text query, no-match, and the `tags=None` backward-compat path.

One process note: I'm the author here, so per `inv-independent-judgment` / `inv-gate-at-handover` this change should get a review by someone other than me before it's considered done — I'll flag that again when I hand it over.
