# `core/` — Trellis-core (the product)

This is **Layer A** (decision `0005`): the shippable product — runtime-free **agent
instructions** (rubrics, sub-agents, conventions) that get installed into a host project's
agent surface (decision `0010`/`0012`). It is distinct from the repo root, which is the
**build methodology** (Layer B) we use to build Trellis (instance #1).

Contents grow with the spine (`spec-0001`):
- `invariants/` — the ratified invariant set (`trellis-product` scope; the load-bearing core).
- `rubrics/` — checkable quality gates (e.g. the artifact contract).
- `fixtures/` — known-bad inputs that the checks must reject (positive controls).

The conformance **sub-agent** that applies these currently runs as the
plugin-carried `grove:conformance-reviewer` in *this* repo (dogfood; the vendored
`.claude/agents/` copy retired — `grove/adr-0026` D1); its product
home is `core/agents/`, which the delivery slice (`0012`) will package and install.
