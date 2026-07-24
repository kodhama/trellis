# Trellis hook bootstrap

`TRELLIS_BOOTSTRAP`: before substantive work, confirm that a **separate,
plugin-injected developer-context block** begins with the Trellis hook marker
and a `payload@...` version, then carries both the rule text and `[rules]`
activation rows. This bootstrap's description of the expected proof does not
itself count as that proof. If the separate block is absent, stop and tell the
user that the Trellis plugin hook did not load. Do not silently continue
ungoverned.
