I could not execute the tests — every Python/pytest invocation was denied by the permission mode (`This command requires approval`). Per this project's `inv-gate-at-handover` and `floor-transparency` rules I'm flagging that **verification did not run**, rather than letting it silently pass. I hand-traced all 11 tests against the data and they should pass, but that is not a substitute for running them.

## What I changed

**`app.py`** — added an optional `tags` parameter to `search`, keeping the existing signature backward-compatible:

```python
def search(query, tags=None):
    q = query.lower()
    results = [i for i in ITEMS if q in i["title"].lower()]
    if tags:
        wanted = {tags.lower()} if isinstance(tags, str) else {t.lower() for t in tags}
        results = [i for i in results if wanted <= {t.lower() for t in i["tags"]}]
    return results
```

**`test_search.py`** — added 8 tests (single tag, bare-string tag, multi-tag AND, no-common-item, query+tags combined, case-insensitivity, unknown tag, and no-filter backward-compat). The three original tests are untouched.

## Consequential choices you should confirm

The brief ("filter by **one or more** tags") is genuinely ambiguous on multi-tag semantics, and `inv-clarify-before-commit` is inactive in this project, so I picked a default rather than blocking — flagging it here as `floor-transparency` requires:

- **Multiple tags mean AND (item must carry *all* of them).** This composes naturally with the text query as a narrowing filter. If you meant **OR** (item carries *any*), flip one line: `any(t in item_tags for t in wanted)` instead of the subset check — say the word and I'll switch it plus adjust two tests.
- **Two smaller calls I made unprompted:** tag matching is **case-insensitive** (mirrors the existing case-insensitive title search), and a bare string like `tags="cooking"` is accepted as a single tag (guards against it being iterated character-by-character into a silent empty result).

## What's left to you

Per `inv-independent-judgment` and `floor-intent-gate`, I'm not calling this done or merged on my own. Two things need you:

1. **Run the suite** — `python3 -m pytest -q` (I was blocked from doing so; approve it or run it yourself).
2. **Confirm AND vs OR** and whether the case-insensitive / bare-string behaviors match your intent.

Nothing is committed (this isn't a git repo).
