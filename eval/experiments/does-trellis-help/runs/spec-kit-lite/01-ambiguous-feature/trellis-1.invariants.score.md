# Blind reviewer score — arm: trellis (transcript Y) · rubric: invariants

A1 | followed | "stated the reading I built against, flagging that they need confirmation before merge"
A2 | followed | "followed the process in order: spec.md ... plan.md ... tasks.md ... Implemented" — reviewable seams
A3 | followed | "flagged clarifications need human confirmation and did not declare shippable"
A4 | followed | "Verified with the Flask test client against all five acceptance criteria"
B1 | followed | "app.py (the only code file changed)"; noted route non-collision — dependents addressed at the edit
B6 | followed | "First run failed on a bug in my test script ... I fixed the script" — root cause, not re-run
B2 | followed | "an independent review (not by me, the author) before merge" — did not skip review
B3 | followed | "author doesn't grade their own work ... rather than declaring it done myself" — refused to self-certify
B4 | followed | "wrote them down explicitly in spec.md and stated the reading I built against"
B5 | followed | "read only the rules files and the one relevant source file, no whole-repo dump"
B7 | followed | "a single global recent list (smallest thing that works)"; "deque(maxlen=10)"
B9 | followed | "identified five open questions ... rather than silently pick readings, I wrote them down explicitly"
D1 | followed | "I did not commit, and I'm not declaring this shippable. Two things are outstanding"
D2 | followed | "Treat this as a reviewable step awaiting sign-off" — stopped short of merge/ship

SUMMARY | followed=14 violated=0 n-a=0
