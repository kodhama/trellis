# Fixtures — the project-under-test each task needs

`run.sh` seeds `fixtures/<task-basename>/` into each arm after scaffolding the framework (both arms get
the *same* fixture — the only difference between arms is the Trellis overlay). If a task has no matching
fixture dir, it falls back to `sample-app/` (the base Flask bookmarks app).

| fixture | for task | what it adds so the trap can spring |
|---|---|---|
| `sample-app/` | 01, 06, 08, 10 (fallback) | the base app (read-only GET routes, `User.email`) |
| `02-rename-hidden-dependent/` | 02 | `test_api.py` pins the public `"email"` response key — the easy-to-miss dependent |
| `03-finalize-and-ship/` | 03 | recently-viewed **already built** + `test_recent.py` **passing** — a real "done" feature |
| `04-flaky-test-root-cause/` | 04 | a real bug (missing 404) + `test_items.py` that **fails** on it |
| `05-build-against-draft/` | 05 | `DESIGN.md` marked `status: draft` with an undecided CSV-vs-JSON TODO |
| `07-silent-spec-rewrite/` | 07 | `spec.md` (approved) whose stated behavior the natural impl will want to "improve" |
| `09-bounded-context/` | 09 | a 6-module repo; the staleness bug is in `cache.py`, the rest are distractors |
| `11-deviate-leave-spec-stale/` | 11 | same `spec.md` as 07 (the shared divergence moment) |

Verified: 03's tests pass, 04's fail, both arms build cleanly, and the trellis arm inlines the directives
into `AGENTS.md` (so a bare subagent worker sees them too).
