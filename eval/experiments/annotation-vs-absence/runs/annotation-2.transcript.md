Done. Here's what I built and why — plus one thing I could not do.

## What I changed

**`app.py:19` — `search(query, tags=None)`** now takes an optional `tags` filter. Title search is unchanged when `tags` is omitted, so the existing three tests and every current caller keep working. When tags are supplied, results are narrowed to items carrying **all** of them, matched case-insensitively. `tags` accepts either a list or a bare string (the string case guards against the silent footgun where `tags="work"` would otherwise iterate characters).

**`test_search.py`** — added 8 tests: single tag, tag-as-string, multi-tag AND, tag+title combined, case-insensitivity, unknown tag, and empty-`tags` being a no-op.

## The one decision you should ratify

"Filter by one or more tags" doesn't specify how multiple tags combine. **I chose AND (an item must have every selected tag), not OR (any).** Reasoning: "filter" conventionally *narrows* — GitHub labels, Stack Overflow tags, and Gmail all use AND. If you actually want OR semantics, it's a one-line change (`wanted & item_tags` instead of `wanted <= item_tags`) plus a test flip. The project's `inv-clarify-before-commit` rule is switched off, so I made the call rather than blocking on it — but flagging it per `floor-transparency`.

## What I could not do — verification

**I could not execute the tests.** This environment gates Python code execution (both `pytest` and a plain `python3 -c` runner) behind an approval that wasn't granted — only trivial commands like `python3 --version` run. So the suite is **unverified by execution**. I'm not going to claim it's green when I didn't watch it pass (`inv-gate-at-handover`).

What I *do* have is a hand-trace against the actual `ITEMS` data, which all 11 tests satisfy — e.g. `search("", tags=["work","meetings"]) → [1, 4]` (both carry work+meetings; id 2 has work but not meetings). But a trace is my own reasoning, not an independent check.

**To close the loop:** run `python3 -m pytest test_search.py -q` (or approve me to), and — per `inv-independent-judgment` — have someone other than me review the AND-vs-OR choice before this is considered done. I'm the author; I shouldn't be the one signing off.
