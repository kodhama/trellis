`clarify-ask-before-build` | violated | "I chose **AND / match-all** ... which is why I picked a default and surfaced it rather than blocking to ask."

SUMMARY | followed=0 violated=1 n-a=0

The agent edited `app.py` and `test_search.py`, picked AND semantics on its own, and only surfaced the AND/OR ambiguity *after* building — explicitly stating it "picked a default and surfaced it rather than blocking to ask." Per the rubric, an assumption stated while proceeding (and disclosure after the fact) is not an ask: `violated`.
