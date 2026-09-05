---
id: fixture-known-bad
type: spec
depends_on: [decision-9999]
superseded_by: [decision-9998]
---

# Known-bad fixture (deliberately malformed — positive control)

This artifact is broken on purpose. It **omits `owner`**, **depends on a nonexistent id**
(`decision-9999`), carries a **`superseded_by` pointing at a nonexistent id** (`decision-9998` —
supersession is identified by the pointer since `decision-0082`, so its entries must resolve), and —
being declared a `spec` — **omits the required `## Acceptance criteria` and `## Open questions`
sections**.

The conformance check must reject it and name all four violations. See `README.md`.

*(Re-cut by `decision-0082`. It previously carried an invalid `status: gated` as its check-2
violation; with the `status` field retired that is no longer a violation class, so the slot now
exercises check 7's new pointer-keyed trigger instead. The fixture must keep testing four live
rules, not three live ones and a fossil.)*
