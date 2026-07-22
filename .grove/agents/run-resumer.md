# run-resumer addendum — trellis

*Optional per-role addendum (grove adr-0026 D3): local rules / worked examples
the generic `grove:run-resumer` reads when present. Consumer-owned.*

**Finding the branch of a run to resume.** trellis's branch-naming convention is
`<category>/<slug>` — e.g. `chore/family-tap`, `lane/t3-tokenize-lp`,
`decision/0042-family-lifecycle`. trellis does **not** uniformly encode an issue
number into the branch name, so do not search for the branch by issue number;
search by the task's own slug instead: `git branch -r | grep <slug>`.
