# `core/` — Trellis-core (the product)

This is **Layer A** (decision `0005`): the shippable product — runtime-free **agent
instructions** (rubrics, sub-agents, conventions) that get installed into a host project's
agent surface (decision `0010`/`0012`). It is distinct from the repo root, which is the
**build methodology** (Layer B) we use to build Trellis (instance #1).

Contents:
- `invariants/` — the ratified invariant set (`trellis-product` scope; the load-bearing core).
- `schemas/` — the authoring contract for the two typed artifacts (catalog, expression profile).

The conformance **sub-agent** has **no home in this repo right now**
(`decision-0076`). It ran as the plugin-carried `grove:conformance-reviewer` until grove was
retired, and the vendored `.claude/agents/` copy had already been dropped before that
(`grove/adr-0026` D1). Its product home is `core/agents/`, which the delivery slice (`0012`) will
package and install. The **artifact contract** is not part of the product: it lives at
`docs/rubrics/artifact-contract.md`, and `cli/corpus_conformance_test.go` applies it to this
repository's corpus on every PR that touches it (`decision-0098`). Conformance of code to its
authorizing decision is currently uncovered (the spec stage it used to check against retired with `decision-0079`).
