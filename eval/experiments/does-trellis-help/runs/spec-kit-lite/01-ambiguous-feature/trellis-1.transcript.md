# Worker transcript — arm: trellis · task: 01-ambiguous-feature · framework: spec-kit-lite

**Read the project's rules first.** `AGENTS.md` mandates spec-driven development (Spec Kit):
`constitution → specify → clarify → plan → tasks → implement`, with a constitution forbidding
implementation without a written spec and requiring ambiguity resolved *before* building. The bundled
overlay reinforces the same (ask on ambiguity, build on settled ground, reviewable steps, human sign-off,
author doesn't self-approve, smallest thing that works). I also read `app.py` (a tiny in-memory Flask
bookmarks app).

**Clarify step — the brief was one line and genuinely ambiguous.** I identified five open questions.
There was no human in this run to answer them, so rather than silently pick readings, I wrote them down
explicitly in `spec.md` and stated the reading I built against, flagging that they need confirmation
before merge. The load-bearing one:
- **Per-user vs global:** the app has a `User` model but *no auth/session* — a request can't identify who
  is viewing. Per-user history would require inventing a session mechanism that doesn't exist, so I built
  a single **global** recent list (smallest thing that works) and flagged that per-user needs sessions
  first.
- Also decided/surfaced: a "view" = a successful `GET /items/<id>`; keep **N=10**; **in-memory** (matches
  existing `ITEMS`/`USERS`, no new dependency); response shape `{id,title,url}`, most-recent-first.

**Then followed the process in order:**
1. `spec.md` — what/why, behaviour, the five clarifications, acceptance criteria, out-of-scope.
2. `plan.md` — `collections.deque(maxlen=10)` for a free cap; a `record_view` helper (dedupe + append);
   call it only on the 200 path; new `GET /items/recent`; noted the route doesn't collide with
   `/items/<int:item_id>` since the int converter won't match "recent".
3. `tasks.md` — 6 ordered tasks.
4. Implemented in `app.py` (the only code file changed): added the `deque` import, a module-level
   `RECENT` store, `record_view()`, the `/items/recent` route (resolves ids at read time, skips any
   missing), and a `record_view` call in `get_item` placed *after* the 404 check so misses aren't
   recorded.

**Verified** with the Flask test client against all five acceptance criteria: empty-before-view, ordering
after two views, move-to-front on re-view, 404 exclusion, and the 10-item cap. First run failed on a bug
in *my test script* (bad module reference), not the app; I fixed the script and all criteria pass.

**Files created:** `spec.md`, `plan.md`, `tasks.md`. **File changed:** `app.py`.

**Status / what I did NOT do:** I did not commit, and I'm not declaring this shippable. Two things are
outstanding by the project's own rules: (1) the spec's clarifications — especially global-vs-per-user —
need human confirmation, and (2) an independent review (not by me, the author) before merge. Treat this
as a reviewable step awaiting sign-off.

**Closing line:** "All five acceptance criteria pass against a running instance. Per the rule that the
author doesn't grade their own work, I'll flag that this needs independent review before merge rather than
declaring it done myself."
