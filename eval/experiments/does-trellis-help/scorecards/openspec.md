# Framework scorecard — OpenSpec (Fission-AI)

OpenSpec imposes a change-first, spec-driven flow and ships a machine-checkable structure. This scorecard
measures whether the agent followed *OpenSpec's own* declared process. A reviewer marks each **followed /
violated / n-a** with an evidence quote.

Workflow: `explore → propose → apply → archive`; specs live under `openspec/changes/<name>/`
(source: [Fission-AI/OpenSpec](https://github.com/Fission-AI/OpenSpec)).

## Rules to score

- **Propose before code.** A change folder with `proposal.md` (+ `specs/`) exists / is authored **before**
  implementation. ✗ look for: editing source with no proposal.
- **Explore first.** The agent surveyed the relevant code/specs before proposing. ✗ look for: proposing
  blind.
- **Spec structure valid.** Requirements + scenarios are captured in the `specs/` form OpenSpec expects
  (would pass `openspec validate`). ✗ look for: freeform notes instead of the change structure.
- **Tasks tracked.** Work follows `tasks.md`; tasks are checked off / kept current. ✗ look for: ignoring
  the task list.
- **Archive on completion.** A completed change is archived (`explore→propose→apply→archive`), not left
  dangling. ✗ look for: implementing then abandoning the change record.
