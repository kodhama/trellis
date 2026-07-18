---
id: decision-0050
type: decision
status: draft
depends_on: [invariants-v1, signature-catalog-v1]
informed_by: [decision-0048, decision-0043]
owner: agent
date: 2026-07-18
---

> **Provenance.** Closes the M2 open question parked in `decision-0048`
> (2026-07-18). The mechanism findings it rests on were gathered earlier in
> the same session (Claude Code's `context: fork`; grove ADR-0030's
> cold/interactive split) — cited in prose, not as frontmatter edges.

# 0050 — the M2 morph's rewrite runs in a cold, isolated sub-agent

## Context

`decision-0048` fixed M1's contamination by **removing** the injected opinion, and explicitly parked
M2: the **morph** — `/trellis:setup`'s one **model-driven** step, which rewrites the consumer's *own*
instruction files "in the project's own voice" (`SKILL.md` M2 §4, "Perform the rewrite yourself") —
carries a contamination that deletion cannot fix. Its bias is a **generative act carrying ambient
context**: run inside a warm session (e.g. this trellis-context one), the rewrite can bleed
trellis-isms into the consumer's files. You can strip an opinion from prose; you cannot strip the
ambient context a generative step reads. Only **isolation** contains it.

The earlier investigation (this session) established the shape of the fix:

- Claude Code supports cold execution (`context: fork`); a dispatched sub-agent reads only its
  declared inputs, never the invoking conversation (grove's cold model — `executor`: *"reads only
  the artifact… never conversation history"*).
- A cold sub-agent **cannot run interactive prompts** — which is exactly why M1 stayed warm
  (`decision-0048`). But M2's interactive/stateful parts and its one generative step are
  **separable**: refuse-without-git, posture, and branch/rollback are interactive; handing a diff to
  the human is the gate; only the **rewrite itself** is generative, and it is non-interactive.

## Decision

**The M2 rewrite runs in a cold, isolated sub-agent that reads only the consumer's own instruction
files + the declared posture + the invariants to bake in — never the invoking session's
conversation.** The warm driving session keeps the interactive/stateful bookends:

- **Warm (driving session):** refuse without git; determine posture (read `expression.md`, or ask);
  record the rollback point and create the `trellis/morph` branch + tag; **dispatch** the cold
  rewrite; then hand the resulting diff to the human.
- **Cold (dispatched sub-agent):** given the posture, the target instruction files, and the
  invariants to encode, **rewrite those files in the project's own voice** — reading only those
  inputs, not the ambient conversation. It writes the rewritten files on the branch and returns a
  summary; it makes no git decision and asks the human nothing.

This contains the generative contamination **structurally** (`inv-bounded-context`: each operation
reads only its declared inputs): the ambient session cannot reach the rewrite, so no ambient bias
can bleed in. It also fits grove's model — the interactive bookends stay with the driving session
(grove ADR-0030), and only the non-interactive generative core is dispatched. The human diff-review
(`SKILL.md` M2 §5) remains the gate, unchanged.

**Mechanism (chosen by the maintainer, 2026-07-18):** the warm M2 flow **dispatches** the rewrite as
a sub-agent (the Task/Agent tool) with a precise, self-contained prompt. Dispatch keeps the
interactive bookends in one flow and hands the cold agent only its declared inputs — **no new
shipped skill or agent**. The alternatives (a `context: fork` sub-skill, or a shipped
`.claude/agents/` morph agent) were weighed and set aside as heavier for equivalent isolation; a
later promotion to a first-class agent stays available if M2 grows its own test surface.

## Consequences

- `plugins/trellis/skills/setup/SKILL.md`'s **M2 §4** changes from "perform the rewrite yourself" to
  "dispatch a cold sub-agent to perform the rewrite, reading only the project's files + posture +
  invariants; the warm session never does the rewrite in-context." §§1–3 and §5 stay warm. **This
  edit advances `install.sh`'s bundle manifest** (`SKILL.md` is manifest-covered — `decision-0028`).
- Closes `decision-0048`'s parked M2 open question.
- **Not implemented in this decision** — the `SKILL.md` change is downstream of ratification (build
  only on settled ground); this decision is the settled ground it will build on.

## Open questions

- **Mechanism — resolved** (maintainer, 2026-07-18): **dispatch a sub-agent**, keeping M2 in one
  skill; the `context: fork` sub-skill and shipped-agent alternatives were set aside as heavier for
  equivalent isolation. A later promotion to a first-class `.claude/agents/` morph agent (testable)
  stays available if M2 grows its own test surface.
- **What the cold agent may read.** To write "in the project's voice" it must read the project's
  existing instruction files — that is its input, not contamination. Bound it to the files being
  rewritten (+ their immediate siblings), not the whole repo (`inv-bounded-context`); the exact
  bound is worth pinning before implementation.
- **Verifiability.** M2 already gates on the human diff-review. Does cold execution want an
  additional check that the rewrite imported no foreign content, or is the human review the
  sufficient gate? Leaning sufficient — worth confirming.

## Self-check (gate)

Grounded in a quoted, ratified invariant (`inv-bounded-context`, "each operation reads only its
declared inputs" — `signature-catalog-v1`) and the session's mechanism findings (`context: fork`;
grove ADR-0030), cited in prose. Framed honestly as **an architecture with a recommended mechanism
and real open forks**, not a settled implementation — the mechanism choice and the read-bound are
surfaced as open questions, not buried. `depends_on` carries the invariant the behaviour implements;
`decision-0048` (parked gap) and `decision-0043` (M2 hosting) are `informed_by`, not
correctness-contingent, per `decision-0047`. Depends only on ratified upstreams; no draft consumed.
Left at `draft`: the author does not grade its own decision — gating and the `approved` intent act
are the maintainer's (`decision-0046`), ideally after a shaping pass on the mechanism fork.
