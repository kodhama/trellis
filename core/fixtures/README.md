# `core/fixtures/` — positive controls

Deliberately **broken** artifacts. Their job is to be **rejected**: the conformance check is
trusted only after it flags every violation here (`rubric-artifact-contract` — *the verifier must be
demonstrably able to fail*, the B3 positive-control lesson logged from our CI episode).

**Excluded from normal corpus runs** — they are test data, not real artifacts.

## Answer key — `known-bad.md`

The check, run against `known-bad.md`, **must report all four:**
1. **Check 1** — missing required field `owner`.
2. **Check 4** — dangling `depends_on: [decision-9999]` (no such artifact).
3. **Check 6** — `type: spec` but missing `## Acceptance criteria` / `## Open questions`.
4. **Check 7** — dangling `superseded_by: [decision-9998]` (no such artifact). Since
   `decision-0082`, supersession is identified by the **pointer** rather than a status value, so the
   pointer's entries must resolve — this is the positive control for that trigger.

A run that *passes* `known-bad.md`, or reports vague/￼incomplete findings, **fails the check
itself.** (Check 3 — id uniqueness — is exercised against the live corpus, where collisions can actually
occur. There is no separate check-5 case to exercise: since `decision-0082`, directional flow for
trellis-self *is* the check-4 resolution test, which the dangling `depends_on` above already covers.)

**The fixture carries no `status` field, and a run that asks for one is itself wrong**
(`decision-0082` retired it). The slot that used to hold an invalid-`status` violation now holds the
check-7 case above — the fixture must test four *live* rules, not three live ones and a fossil.
