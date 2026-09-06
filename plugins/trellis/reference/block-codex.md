<!-- trellis:codex-bootstrap:begin (managed by trellis — edit .trellis/, not this block) -->
# Trellis delivery receipt and fallback

Trellis rules are authoritative only in the installed project files listed below. This bootstrap is a **best-effort** model instruction: it is not proof that a native hook ran, and it never replaces those files.

Before substantive work, assess two independently loaded components:

1. Generated prose is complete only when the exact terminal sentinel `<!-- trellis:rules-loaded -->` is followed, later in the same generated prose, by the exact end marker `<!-- trellis:prose-complete -->` — both written by the generator, in that order, each matched as a whole line and as nothing else. The prose between and around them is free to be reworded and is no part of this test. A sentinel alone, an end marker alone, a diagnostic marker, this bootstrap's mention of either marker, or bare slug-name presence is not completion.
2. Activation TOML is valid when it parses and `strictness`, if present, is exactly `firm` or `adaptive` — absent, read it as `adaptive`. A single top-level `governed = false` is an opt-out, not a row set: nothing is reconciled, no rule applies including the two floor rules, and you say so rather than govern. Otherwise a row set that does not match the canonical list below is **reconciled, never refused** — for this session, in what you load, never by editing the file: a canonical slug carrying no row of its own governs as active, the first occurrence of a repeated slug is kept, and a correctly shaped row naming a slug not in that list, or a later duplicate of one in it, is set aside — commented out with the date and the reason, its value kept verbatim, never deleted. A disabled floor row is understood as overridden-by-floor. Only a genuine syntax fault makes the file invalid: a row not of the form `inv-<name>` or `floor-<name>`, `<name>` lowercase letters and hyphens only, followed by `= { active = <boolean> }`; before the `[rules]` header, a key other than `seeded_from`, `strictness` or `governed`, or one of those repeated; a duplicate or foreign section; a `seeded_from` or `strictness` whose value is not a quoted string, or a `strictness` that is present but is neither value; or a `governed` that is not a boolean.

`inv-directional-flow`, `inv-handover-points`, `inv-intent-locus`, `inv-ratifiable-artifacts`, `inv-graph-maintenance`, `inv-self-improvement`, `inv-deliberate-succession`, `inv-no-orphan-followups`, `inv-gate-at-handover`, `inv-independent-judgment`, `inv-auditable-archive`, `inv-bounded-context`, `inv-minimal-first`, `inv-clarify-before-commit`, `floor-transparency`, `floor-intent-gate`

Use this single-copy fallback table:

- If both the sentinel-plus-end-marker boundary and valid activation TOML are already present from a previously verified generated overlay, use the loaded context and read no Trellis file again.
- If the boundary is present but activation TOML is absent or invalid, read only `.trellis/rules.toml`.
- If valid activation TOML is present but the boundary is absent, read only the three `.trellis/internal/` files.
- If neither component is present, read and validate all four installed inputs.

The four inputs are `.trellis/internal/trellis.md`, `.trellis/internal/rules.md`, `.trellis/internal/version`, and `.trellis/rules.toml`. The generated prose files must be readable and nonempty; trellis.md must contain exactly one exact `@rules.md` expansion point; rules.md must carry the one terminal sentinel; version, after at most one terminal newline is trimmed, must match `^payload@[0-9a-f]{12}$`; and rules.toml must be valid by the test above. The installed files, never plugin-side reference files, are the rule authority.

Missing native-hook delivery is not itself an error: attempt the applicable fallback branch. A reconciled row set is not a failure to load — govern by the reconciled set, and tell the user what you reconciled, row by row, before substantive work. If the required installed components remain absent, unreadable, or invalid, tell the user exactly **“Trellis was not loaded”** and do not claim governed execution.
<!-- trellis:codex-bootstrap:end -->