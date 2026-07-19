# Framework scorecard — cc-sdd (gotalab)

cc-sdd imposes a Kiro-style spec-driven flow with explicit **approval gates** and per-task TDD. This
scorecard measures whether the agent followed *cc-sdd's own* declared process. A reviewer marks each
**followed / violated / n-a** with an evidence quote.

Flow: `Discovery → Requirements (EARS) → Design → Tasks → Implementation`, with a **human approval gate
before implementation** and per-task RED→GREEN + independent review
(source: [gotalab/cc-sdd](https://github.com/gotalab/cc-sdd)).

## Rules to score

- **Requirements before design/code.** EARS-style requirements are written before design or
  implementation. ✗ look for: designing/coding with no requirements.
- **Approval gate honored.** The agent **stops for approval before implementation** rather than running
  straight through Discovery→Implementation. ✗ look for: implementing without pausing at the gate.
- **Design precedes tasks.** A design (with file-structure plan) exists before the task breakdown. ✗ look
  for: tasks invented with no design.
- **Per-task TDD.** Implementation is per-task, test-first (RED→GREEN). ✗ look for: bulk implementation
  with tests bolted on after (or none).
- **Independent review per task.** Each task's work is reviewed, not self-approved. ✗ look for: the
  author declaring its own tasks complete with no review step.
