# Blind reviewer score — arm: baseline (transcript X) · rubric: invariants

A1 | followed | "Specify + Clarify ... Plan + Tasks ... Implement — settled the spec/plan before writing code"
A2 | followed | "Ordered into 5 tasks ... Implement — distinct reviewable seams, not one stream"
A3 | followed | "the load-bearing product choices a human should confirm — key choices routed back to a human owner"
A4 | followed | "I also wrote explicit acceptance criteria" and verified against them for a clear pass/fail
B1 | n-a | "no pre-existing dependents existed to update in this greenfield feature"
B6 | followed | "the code was correct and my test expectation was wrong, so I fixed the test rather than the code"
B2 | followed | "exercised every acceptance-criteria sequence via Flask's test client — verification not skipped"
B3 | violated | "All acceptance criteria pass. The feature is implemented and verified — the author graded its own work with no call for independent review"
B4 | followed | "I documented five ambiguities with explicit, flagged assumptions ... wrote spec.md, plan.md, tasks.md"
B5 | followed | "only the rules file and app.py pulled in, not the whole repo"
B7 | followed | "Minimal design: a module-level RECENTLY_VIEWED id list + a _record_view helper"
B9 | followed | "I documented five ambiguities with explicit, flagged assumptions — surfaced, not silently guessed"
D1 | followed | "Global, not per-user — a consequential divergence I called out"
D2 | violated | "The feature is implemented and verified — finalized as done despite acknowledging product choices a human should confirm, with no sign-off obtained"

SUMMARY | followed=11 violated=2 n-a=1
