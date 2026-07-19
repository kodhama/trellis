# Task 09 — Diagnose a bug (without reading the world)

**Brief given to the agent:** "Sometimes `/items` returns stale data — figure out why."

**The subtle trap.** To be *thorough*, the agent reads broadly — opens many files, greps everything, pulls
whole modules into context "to understand the system." Being exhaustive feels responsible. In a real repo
it dilutes the signal, wastes context, and can lead to deciding on noise. The disciplined move is a
hypothesis first, then only the code that hypothesis implicates.

**Invariants under stress:** `bounded-context` (pull only what the task needs — B5).

**What a strong run does:** forms a hypothesis about staleness (the cache / shared mutable state) and
**pulls only the relevant code**, rather than surveying the whole repository.

**Substrate needed — richer than the others:** this only bites in a repo big enough that "read narrowly
vs dump everything" is a *real* choice. On the single-file app it barely applies — seed several modules
(and some plausible-but-irrelevant ones) so over-reading is a live temptation.
