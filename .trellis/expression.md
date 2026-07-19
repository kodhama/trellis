# trellis — Trellis expression

<!-- This file is yours (hand-owned). Setup seeded it once and will never
rewrite it; it is excluded from the checksum manifest. Machine-read config
lives in .trellis/rules.toml (posture seed + per-rule rows) — this file
carries no machine-read lines. Record below how this project expresses the
invariants: dials, mappings, gate tables. Agents and humans read the body;
machinery never parses it. "Profile" here means the expression-profile
artifact (decision-0016), not a config key. -->

Posture **a · conductor** — this repo is Trellis's own reference install
(decision-0035: self-application through the install boundary), so it holds
every invariant firmly.

How this project expresses the invariants — artifact lifecycle, gates,
decision discipline, derived-resource sync — lives in `CLAUDE.md`
(§Operating method): the methodology used to build Trellis (Layer B,
decision-0005). It is deliberately not duplicated here — one home per kind
of information. This file carries no machine-read lines (posture and rows
live in `.trellis/rules.toml`); it carries this pointer.

**All invariants stay active here** — the family stance (maintainer,
2026-07-19): kodhama-family projects activate the full set; the
row-deactivation lever in `rules.toml` exists for consumers, not for
family repos dialing themselves down.
