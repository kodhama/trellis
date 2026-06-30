---
id: fixture-known-bad
type: spec
status: gated
depends_on: [decision-9999]
---

# Known-bad fixture (deliberately malformed — positive control)

This artifact is broken on purpose. It **omits `owner`**, uses an **invalid `status`
(`gated`)**, **depends on a nonexistent id** (`decision-9999`), and — being declared a
`spec` — **omits the required `## Acceptance criteria` and `## Open questions` sections**.

The conformance check must reject it and name all four violations. See `README.md`.
