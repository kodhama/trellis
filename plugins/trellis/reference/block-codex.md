<!-- trellis:codex-bootstrap:begin (managed by trellis — edit .trellis/, not this block) -->
# Trellis delivery receipt and fallback

Trellis rules are authoritative only in the installed project files listed below. This bootstrap is a **best-effort** model instruction: it is not proof that a native hook ran, and it never replaces those files.

Before substantive work, assess two independently loaded components:

1. Generated prose is complete only when the exact terminal sentinel `<!-- trellis:rules-loaded -->` is followed, after only its generated newline, by the fixed footer whose first nonblank line is `---` and whose next text is the ambiguity/fallback sentence. A sentinel alone, a diagnostic marker, this bootstrap's mention of the sentinel, or bare slug-name presence is not completion.
2. Activation TOML is complete only when it parses, strictness is exactly `firm` or `adaptive`, every canonical slug below occurs exactly once, no unknown or duplicate slug occurs, and any disabled floor row is understood as overridden-by-floor:

`inv-directional-flow`, `inv-handover-points`, `inv-intent-locus`, `inv-ratifiable-artifacts`, `inv-graph-maintenance`, `inv-self-improvement`, `inv-gate-at-handover`, `inv-independent-judgment`, `inv-auditable-archive`, `inv-bounded-context`, `inv-minimal-first`, `inv-clarify-before-commit`, `floor-transparency`, `floor-intent-gate`

Use this single-copy fallback table:

- If both the sentinel-plus-fixed-footer boundary and valid activation TOML are already present from a setup-verified generated overlay, use the loaded context and read no Trellis file again.
- If the boundary is present but activation TOML is absent or invalid, read only `.trellis/rules.toml`.
- If valid activation TOML is present but the boundary is absent, read only the three `.trellis/internal/` files.
- If neither component is present, read and validate all four installed inputs.

The four inputs are `.trellis/internal/trellis.md`, `.trellis/internal/rules.md`, `.trellis/internal/version`, and `.trellis/rules.toml`. The generated prose files must be readable and nonempty; trellis.md must contain exactly one exact `@rules.md` expansion point; rules.md must carry the one terminal sentinel; version, after at most one terminal newline is trimmed, must match `^payload@[0-9a-f]{12}$`; and rules.toml must satisfy the complete activation predicate above. The installed files, never plugin-side reference files, are the rule authority.

Missing native-hook delivery is not itself an error: attempt the applicable fallback branch. If the required installed components remain absent, unreadable, or invalid, tell the user exactly **“Trellis was not loaded”** and do not claim governed execution.
<!-- trellis:codex-bootstrap:end -->