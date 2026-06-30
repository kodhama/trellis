# `core/fixtures/` — positive controls

Deliberately **broken** artifacts. Their job is to be **rejected**: the conformance check is
trusted only after it flags every violation here (`spec-0001` AC2 — *the verifier must be
demonstrably able to fail*, the B3 positive-control lesson logged from our CI episode).

**Excluded from normal corpus runs** — they are test data, not real artifacts.

## Answer key — `known-bad.md`

The check, run against `known-bad.md`, **must report all four:**
1. **Check 1** — missing required field `owner`.
2. **Check 2** — invalid `status: gated` (not in `{draft, ratified, superseded}`).
3. **Check 4** — dangling `depends_on: [decision-9999]` (no such artifact).
4. **Check 6** — `type: spec` but missing `## Acceptance criteria` / `## Open questions`.

A run that *passes* `known-bad.md`, or reports vague/￼incomplete findings, **fails the check
itself.** (Checks 3/5/7 — uniqueness, directional flow, supersede integrity — are exercised
against the live corpus, where such edges can actually occur.)
