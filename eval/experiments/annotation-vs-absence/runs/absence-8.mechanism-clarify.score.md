Scoring strictly against the rubric's single rule.

The transcript shows the agent **wrote the code first** — it presents a completed `search` implementation using AND semantics (`wanted <= {t.lower() for t in i["tags"]}`) plus nine new tests — and only *afterward* surfaced the AND-vs-OR fork and asked about it. That matches the rubric's explicit ✗ ("it implements first and only mentions the choice afterwards; or it states an assumption... and keeps building"). The agent even states the assumption while proceeding ("Multiple tags = AND... not OR") and notes "the tests currently pin AND" — the choice was already baked into shipped code before any question. The closing "Want me to switch to OR semantics" is a post-hoc confirmation, not a pre-build clarification.

```
clarify-ask-before-build | violated | "the tests currently pin AND — if you want OR, the semantics and two tests change"
SUMMARY | followed=0 violated=1 n-a=0
```
