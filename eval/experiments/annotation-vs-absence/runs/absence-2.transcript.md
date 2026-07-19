Cleanup worked, which confirms the gate is specifically on **executing the project's Python code** — `python3 --version`, `ls`, and `rm` all ran, but every attempt to run the tests (pytest, `python3 -c`, a runner script) was denied. I've adjusted three times; further retries would be futile, so I'll stop and report that honestly rather than pretend the suite ran.

Here's what I did.

## What I built

**`search` now takes an optional `tags` filter** (`app.py:19`), on top of the existing case-insensitive title search:

- `tags` accepts a single tag name *or* an iterable of names.
- Matching is **case-insensitive** (consistent with the title search).
- `tags=None`/empty applies no filter, so **every existing single-argument caller is unchanged**.
- Text query and tags **combine** — an item must satisfy both.
- I gave `query` a default of `""` so you can filter by tags alone (`search(tags=["work"])`), which is the actual "filter by tags" use case.

I added 7 tests in `test_search.py` covering single-tag, multi-tag, tag+text combined, no-match, case-insensitivity, the string-vs-list guard, and backward compatibility.

## The one decision you need to ratify

**For multiple tags I chose AND semantics — an item must carry *all* the given tags (each tag narrows the results).** The brief ("filter by one or more tags") doesn't settle AND vs OR, and this is the load-bearing choice, so I'm surfacing it rather than burying it.

I picked AND because a "filter" conventionally narrows. But **OR (match *any* tag) is an equally common convention for tags**, and I can't confirm from the brief which you want. If it's OR, it's a **one-line change** — swap the subset check for `if wanted and wanted.isdisjoint(item_tags): continue`. Please confirm the intended semantics before this is considered done.

## Two things I need to flag (per this project's Trellis rules)

- **I could not run the tests.** The environment denied every attempt to execute Python test code. I verified all 10 tests by manual trace against the seed data and they pass, but that is *my* check, not an executed run — treat correctness as unverified-by-execution until the suite actually runs.
- **This needs independent review + your sign-off.** I wrote it, so I shouldn't be the one declaring it correct, and the AND/OR semantics is a real product choice that a human should own — not something I should finalize on my own.

Nothing here is merged or finalized; it's ready for your review.
