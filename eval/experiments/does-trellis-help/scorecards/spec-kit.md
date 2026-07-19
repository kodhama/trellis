# Framework scorecard — GitHub Spec Kit

Spec Kit imposes a **spec-driven** process and lands checkable artifacts. This scorecard measures whether
the agent followed *Spec Kit's own* declared process — the "does Trellis help you follow **your**
methodology?" dimension. A reviewer marks each **followed / violated / n-a** with an evidence quote from
the transcript.

Process: `constitution → specify → (clarify) → plan → tasks → (analyze) → implement`
(source: [github/spec-kit](https://github.com/github/spec-kit)).

## Rules to score

- **Spec before code.** A `spec.md` (feature spec) exists / is consulted **before** implementation code
  is written. ✗ look for: jumping straight to editing source with no spec.
- **Clarify before planning.** Ambiguities in the spec are resolved (the `/clarify` step) before a plan
  is drawn. ✗ look for: planning/implementing an ambiguous spec without clarification.
- **Honor the constitution.** The project's `constitution.md` rules are respected. ✗ look for: violating
  a stated constitution rule.
- **Plan → tasks → implement, in order.** Work follows the ordered `tasks.md`; phases aren't skipped or
  reordered. ✗ look for: implementing before a plan/tasks exist, or ignoring the task breakdown.
- **Artifacts kept current.** The spec/plan/tasks are updated when the work diverges from them. ✗ look
  for: code that no longer matches the spec, with the spec left stale.

> Note: Spec Kit's non-TTY init silently defaults to Copilot; the harness always passes
> `--integration claude` for reproducibility against the same agent.
