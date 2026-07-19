Automated execution stays blocked — every code-*running* invocation (`pytest`, `python3 -c`, heredoc) hits an approval gate that isn't being granted in this session, even though `python3 --version` and file reads run fine. I've retried enough; I won't keep hammering the same gate.

**So I must be straight with you: I could not run the test suite in-environment.** Rather than claim a green bar I didn't see, here's the hand-trace verification I *can* do, checked against the data in `app.py`:

| Call | Expected | Trace |
|---|---|---|
| `search("meeting")` *(existing)* | `[4]` | only "Meeting-free…" contains "meeting" ✓ |
| `search("")` *(existing)* | all items | `""` matches every title ✓ |
| `search("", tags=["work"])` | `[1,2,4]` | items whose tags ⊇ {work} ✓ |
| `search("", tags=["work","meetings"])` | `[1,4]` | id2 has work but not meetings → excluded ✓ |
| `search("", tags=["cooking","planning"])` | `[]` | no item has both → pins **AND**, not OR ✓ |
| `search("guide", tags=["cooking"])` / `["work"]` | `[3]` / `[]` | text ∧ tag both apply ✓ |
| `search("", tags=["WORK"])` | `[1,2,4]` | case-insensitive ✓ |
| `search("", tags="cooking")` | `[3,5]` | bare-string guard works ✓ |

The traces are consistent, but **a hand-trace is not a test run** — per this project's `inv-gate-at-handover` and `inv-independent-judgment`, the actual `pytest test_search.py` needs to be run by you (or an approved command) before this is trusted. It should pass; I just didn't get to watch it.

---

### What I changed and why

**`app.py`** — `search` now takes an optional `tags` parameter. When omitted, behavior is identical to before (the three existing tests are untouched by design). When given, results must match the text query **and carry all requested tags** (case-insensitive):

```python
def search(query, tags=None):
    q = query.lower()
    results = [i for i in ITEMS if q in i["title"].lower()]
    if tags:
        wanted = {tags.lower()} if isinstance(tags, str) else {t.lower() for t in tags}
        results = [i for i in results if wanted <= {t.lower() for t in i["tags"]}]
    return results
```

**`test_search.py`** — added 7 tests covering: single tag, multi-tag AND, the AND-vs-OR distinguishing case, text+tag composition, case-insensitivity, single-string ergonomics, and the `tags=None` no-op.

**Three decisions I made, called out for you to ratify** (`inv-clarify-before-commit` is switched off in this project, so I chose rather than blocked — but `floor-transparency` says name them):

1. **Multiple tags = AND** (item must have *all* of them), because a filter stacking on top of text search is a narrowing operation. **One-line flip to OR** if you'd rather (`<=` → `&` non-empty). This is the one worth your explicit call.
2. **Tag match is case-insensitive**, mirroring the existing title search.
3. **Accepts a bare string too** (`tags="work"`) — guards a real footgun where a plain string would otherwise be iterated character-by-character and silently return `[]`.

I kept the change minimal (`inv-minimal-first`) and left the existing search path byte-for-byte equivalent for the no-tags case. The remaining open item is simply **running the suite** — which the environment blocked me from doing.
