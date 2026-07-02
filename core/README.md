# `core/` — Bonsai-core (the product)

This is **Layer A** (decision `0005`): the shippable product — runtime-free **agent
instructions** (rubrics, sub-agents, conventions) that get installed into a host project's
agent surface (decision `0010`/`0012`). It is distinct from the repo root, which is the
**build methodology** (Layer B) we use to build Bonsai (instance #1).

Contents grow with the spine (`spec-0001`):
- `invariants/` — the ratified invariant set (`bonsai-product` scope; the load-bearing core).
- `rubrics/` — checkable quality gates (e.g. the artifact contract).
- `fixtures/` — known-bad inputs that the checks must reject (positive controls).

The conformance **sub-agent** that applies these currently lives at
`.claude/agents/conformance-reviewer.md` so it runs in *this* repo (dogfood); its product
home is `core/agents/`, which the delivery slice (`0012`) will package and install.
