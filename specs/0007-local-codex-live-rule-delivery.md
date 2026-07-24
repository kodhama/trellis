---
id: spec-0007
type: spec
status: gated
depends_on: [decision-0058, spec-0004, spec-0006]
implements: decision-0058
owner: agent
rubric: rubric-artifact-contract
version: 1
date: 2026-07-24
---

# Spec 0007 — Phase 1 local Codex live-rule delivery

## Purpose

Deliver the installed Trellis overlay to a fresh, trusted local Codex session
through a host-native startup hook, while preserving Claude's existing import
transport and one effective rule copy per host.

## Scope and support boundary

Phase 1 supports:

- setup, ordinary refresh, and product-wide remove through the Trellis plugin;
- Claude's existing `CLAUDE.md` import transport;
- Codex setup and refresh through a separately marked receipt/fallback in
  `AGENTS.md`; and
- a fresh local Codex startup in a trusted repository through a locally
  installed Trellis plugin.

The support claim requires the production hook-contract tests in this spec and
the already recorded live local-startup positive control in `decision-0058`.
Fixture tests for another event or surface do not expand the claim.

The following are explicitly excluded from Phase 1:

- Codex resume, clear, compact, and subagent boundaries;
- Codex desktop, IDE, headless/automation, and cloud surfaces;
- a Claude hook replacement;
- any other host-native transport;
- an explicit preset-reset operation;
- revival of the parked `seed` or `custom` presets;
- mutation of a model context already in flight; and
- a per-host disable or remove operation.

Dormant shared machinery for an excluded boundary may exist only when it is not
registered by default and is not described as supported.

## Runtime assumptions

| Assumption | Contract consequence |
|---|---|
| The Codex plugin is installed locally and discoverable by Codex. | This spec creates no new remote installation or release channel. |
| The repository has passed Codex's local trust boundary. | Trellis does not bypass trust or claim hook execution in an untrusted repository. |
| The hook receives JSON with `hook_event_name`, `source`, and `cwd`, plus an absolute `PLUGIN_ROOT`. | The hook validates those inputs and resolves the project as specified below; it never assumes the process working directory is the project. |
| The host invokes a fresh-start event distinguishable from other lifecycle events. | Only `SessionStart(startup)` is registered and supported in Phase 1. |
| The local plugin host provides Node.js 20 or newer. | Node.js `>=20` is the Phase 1 Codex hook runtime prerequisite. Setup preflights it. If unavailable, setup reports native-hook degradation and installs the bootstrap-only path; it does not claim native delivery. |
| `AGENTS.md` is loaded by Codex and imported by Claude through the repository's existing adapter. | The Codex bootstrap must be harmless in Claude when Claude's Trellis import has already loaded the sentinel and activation rows. |

## Authoritative inputs and validity

The installed overlay remains the sole rule authority. The Codex transport
reads exactly these project files:

| Input | Role |
|---|---|
| `.trellis/internal/trellis.md` | Generated Trellis header and its exact `@rules.md` expansion point |
| `.trellis/internal/rules.md` | Generated complete rule readout whose exact terminal line is the loaded-context sentinel |
| `.trellis/internal/version` | Installed diagnostic stamp |
| `.trellis/rules.toml` | Current consumer-owned activation rows and strictness |

An overlay is valid for Codex delivery only when all four files are readable,
the two generated prose files are nonempty, the version is a nonempty
`payload@<hash>` stamp, and `rules.toml` passes the same validator as setup:

- `strictness` is exactly `"firm"` or `"adaptive"`;
- every shipped rule slug has exactly one row;
- no unknown slug is present; and
- a false floor row is surfaced as overridden-by-floor while the floor remains
  applicable.

The stamp is diagnostic provenance, not rule authority. A valid but older
installed payload may produce the existing staleness warning; staleness alone
does not authorize the hook to substitute plugin-side rule files for the
installed project files.

The version file is valid only after trimming at most one terminal newline and
matching `^payload@[0-9a-f]{12}$`. The hook does not rehash content; this stamp
is diagnostic only.

The generator places one stable, posture-independent sentinel as the exact
terminal line of `rules.md`; it is part of the manifest-covered generated
readout. `trellis.md` contains no sentinel. Claude can receive the sentinel
only by successfully following `trellis.md`'s nested `@rules.md` import; the
static Claude block does not carry a second copy. The Codex hook requires the
installed `rules.md` to contain exactly one sentinel at that terminal position
before using the file. Both posture variants of `trellis.md` carry the same
fixed post-import footer: the first nonblank text after `@rules.md` is `---`,
followed by the existing ambiguity/fallback sentence. A complete flattened
payload therefore has one observable cross-file boundary: the sentinel is
immediately followed, ignoring only the generated newline, by that exact
footer. The bootstrap treats only that sentinel-plus-footer adjacency as
completion for a setup-verified generated overlay; a bare sentinel does not
count. This is a best-effort duplicate-avoidance receipt, not an unspoofable
source-boundary proof after Claude has flattened imports. Adversarial
post-setup mutation of generated `.trellis/internal/` bytes is outside that
receipt's guarantee and is caught by the next setup checksum verification.
No bootstrap mention or diagnostic marker counts as the receipt. The hook
additionally carries the installed stamp in its diagnostic line.

## Codex hook contract

### Registration and host isolation

The plugin carries both host manifests without making either host consume the
other host's protocol:

| Host | Registered transport | Prohibited cross-host behavior |
|---|---|---|
| Claude | Existing staleness `SessionStart` hook and `CLAUDE.md` imports | The Codex context hook is not registered or emitted. |
| Codex | One `SessionStart` hook matched only to `startup` | Claude-shaped staleness output is not registered or emitted. |

The Codex manifest points its hook registration to
`./hooks/codex-hooks.json`; that file registers a `SessionStart` matcher of
exactly `startup`. The registration must not include `resume`, `clear`,
`compact`, subagent, or surface-wide matchers.

Host detection uses this precedence:

1. when `PLUGIN_ROOT` is set, it must be an absolute existing directory whose
   `.codex-plugin/plugin.json` parses as JSON with `name` exactly `"trellis"`;
   a valid value selects Codex even
   when `CLAUDE_PLUGIN_ROOT` is also set, and an invalid value stops rather
   than falling through;
2. only when `PLUGIN_ROOT` is absent, a valid absolute `CLAUDE_PLUGIN_ROOT`
   whose `.claude-plugin/plugin.json` parses as JSON with `name` exactly
   `"trellis"` selects Claude; and
3. an invalid set root, mismatched manifest, or conflict after applying that
   precedence stops visibly before any project write or hook emission.

### Input and output

The supported hook input is a JSON object with:

- `hook_event_name` exactly `"SessionStart"`;
- `source` exactly `"startup"`; and
- `cwd` as an absolute path to an existing directory.

Starting at `cwd`, the hook finds the nearest ancestor containing a `.git`
directory or `.git` file; that is the trusted repository boundary. It then
selects the nearest directory from `cwd` through that boundary, inclusive,
that contains `.trellis/rules.toml`. It never searches above the nearest Git
boundary. No Git boundary or no overlay inside it is a visible failure.
`PLUGIN_ROOT` must satisfy the exact manifest identity above. Missing,
malformed, relative, nonexistent, wrong-plugin, or wrong-event inputs fail
visibly. The hook never falls back to process `$PWD`, the plugin checkout, a
parent repository, or another repository.

On a valid startup and valid overlay, stdout is exactly one JSON object and no
prose:

```json
{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"<assembled-context>"}}
```

`<assembled-context>` is assembled deterministically:

1. read `.trellis/internal/trellis.md`;
2. require exactly one exact `@rules.md` placeholder and replace that
   placeholder once with `.trellis/internal/rules.md`;
3. append the complete current `.trellis/rules.toml`;
4. append one diagnostic line carrying the installed version stamp.

Zero or multiple exact `@rules.md` placeholders are invalid. The hook does not
paraphrase, filter, regenerate, or source rule content from the plugin payload.
Each installed source appears once in the result. Validation occurs before
stdout is written, so the rules file's terminal sentinel is emitted only as
part of a complete successful assembly.

A successful `additionalContext` must be at most **8,000 UTF-8 bytes**. The
generator/build guard and hook contract tests fail if the largest supported
assembled fixture exceeds that budget. At runtime, an over-budget assembly is
a visible failure and emits no rule payload.

When validation succeeds but one or both floor rows are `active = false`,
success remains the exact `hookSpecificOutput` object above with one optional
top-level sibling:

```json
{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"<assembled-context>"},"systemMessage":"Trellis warning: floor rows set active = false are overridden-by-floor and remain active: <sorted comma-separated slugs>."}
```

No other success-path `systemMessage` is permitted.

On an invalid input, invalid/missing overlay, invalid placeholder count, or
over-budget assembly, stdout is exactly one failure object with no
`hookSpecificOutput`:

```json
{"systemMessage":"Trellis hook did not load rules: <path-or-input>: <validation-class>. The AGENTS.md bootstrap must attempt the installed overlay."}
```

`<path-or-input>` is exactly one of `stdin`, `hook_event_name`, `source`,
`cwd`, `project-root`, `PLUGIN_ROOT`, `assembled-context`, or one of the four
repository-relative authoritative paths. `<validation-class>` is exactly one
of `invalid-json`, `wrong-event`,
`invalid-cwd`, `project-root-not-found`, `invalid-plugin-root`,
`missing-file`, `unreadable-file`, `empty-prose`, `invalid-version`,
`invalid-rules`, `invalid-placeholder-count`, or `context-over-budget`.
Tests cover every pair the handler can emit. The diagnostic must not claim
governed execution or turn a missing overlay into plausible rule content.

An excluded but well-formed lifecycle event produces no Trellis rule context.
A malformed event fails visibly without reading another repository.

## Codex receipt and fallback

The generated Codex bootstrap is a small, separately manifest-covered payload
file installed between a unique begin/end marker pair in `AGENTS.md`. It
contains:

- a receipt naming Trellis and the installed-file authority;
- the sentinel-plus-fixed-footer boundary proof and activation-row duplicate
  check;
- the canonical inventory of all 14 assessable slugs, as names only;
- instructions to read missing installed components before substantive work;
- the four paths and validity conditions above; and
- the required failure disclosure when no valid transport succeeds.

It contains no rule prose, activation-row values, generated rule readout,
posture-specific content, or embedded copy of any `.trellis/` input. The
14-slug names are permitted solely so the bootstrap can decide whether the
loaded activation rows are complete.

The canonical inventory carried by the bootstrap is exactly:

`inv-directional-flow`, `inv-handover-points`, `inv-intent-locus`,
`inv-ratifiable-artifacts`, `inv-graph-maintenance`,
`inv-self-improvement`, `inv-gate-at-handover`,
`inv-independent-judgment`, `inv-auditable-archive`,
`inv-bounded-context`, `inv-minimal-first`,
`inv-clarify-before-commit`, `floor-transparency`, and
`floor-intent-gate`.

The fallback is a best-effort model instruction. Tests assert its exact
generated bytes and branching instructions; they do not claim deterministic
execution merely because the text is present. In particular, flattened Claude
context cannot make a textual source boundary unspoofable; the no-reload branch
assumes the generated overlay passed setup verification and has not been
adversarially modified since. The Codex hook does not share that limitation:
it validates source-file terminal position before output.

The bootstrap applies this single-copy table:

| Context state | Fallback action |
|---|---|
| The exact sentinel-plus-fixed-footer adjacency and valid activation TOML are both present in a setup-verified generated overlay | Use the loaded context; read no Trellis file again. Valid activation TOML means parseable TOML, valid strictness, exactly one row for each of the 14 canonical slugs, no unknown/duplicate rows, and the floor override understood. A diagnostic marker, bootstrap mention, sentinel alone, or bare slug-name presence is never sufficient. |
| The valid sentinel-plus-footer boundary is present but activation TOML is absent or invalid | Read and validate only `.trellis/rules.toml`; do not reload generated prose. |
| Valid activation TOML is present but the sentinel-plus-footer boundary is absent | Read and validate the three `.trellis/internal/` files; do not reload the rows. |
| Neither component is present | Read and validate all four installed inputs. |

Presence of the hook diagnostic marker alone never proves that rules loaded.
Missing hook execution is not itself an error. If the native hook did not
deliver valid context, the bootstrap attempts the table above. If the required
installed components remain absent, unreadable, or invalid, the agent must tell
the user that Trellis was not loaded and must not claim governed execution.

## Setup, refresh, and remove

### Host-scoped setup and ordinary refresh

Setup first applies the host-detection precedence above and preflights Node.js
`>=20` for Codex native delivery. Missing Node is reported as a visible
bootstrap-only degradation, not mistaken for native-hook support.

Before any write, setup preflights the complete transaction: host identity,
runtime status, config validity or seed choice, every source payload and
manifest check, target-file existence/content, and every relevant marker pair.
Any invalid source, unresolved choice, or marker ambiguity stops with **no
project file changed**.

Both host branches seed or refresh the same overlay:

- generated files under `.trellis/internal/` are copied and verified from the
  payload;
- a missing `.trellis/rules.toml` is seeded through the existing posture
  choice;
- an existing `.trellis/rules.toml` is validated and never overwritten; and
- setup/refresh is not required for a row edit to take effect at the next
  context-loading boundary.

Instruction-file ownership is host-isolated:

| Current host | Managed target | Required behavior | Must remain byte-identical |
|---|---|---|---|
| Claude | Existing Trellis block in `CLAUDE.md` | Install or replace it with the manifest-covered Claude import block. | `AGENTS.md`, including any Codex bootstrap |
| Codex | Codex receipt/fallback in `AGENTS.md` | Install or replace it with the manifest-covered Codex bootstrap. | `CLAUDE.md`, including its Claude adapter and Trellis block |

A fresh target appends one managed block with one separator newline. A refresh
replaces from the first matching begin marker through the paired end marker,
inclusive. Bytes outside the markers are preserved exactly. Duplicate,
unpaired, nested, or ambiguous markers stop visibly before any edit.

Running Claude setup and Codex setup in either order is idempotent and leaves
one transport per host. Setup never places the Claude import block in
`AGENTS.md`, the Codex bootstrap in `CLAUDE.md`, or full rule text in either
bootstrap.

### Product-wide remove

`/trellis:remove` remains the product-wide clean exit from `spec-0004`. Phase 1
adds no per-host disable operation. From either host, remove:

1. preflights every recognized instruction file, marker pair, overlay path,
   consumer-owned-content consent, and cleanup target before any edit;
2. stops with no block or overlay change if any marker or consent is
   ambiguous;
3. strips every detected managed block byte-safely, including its one
   setup-added separator newline;
4. preserves all surrounding user bytes;
5. deletes the shared `.trellis/` overlay only after the managed blocks have
   been handled; and
6. reports every removed, retained, ambiguous, or absent item.

If an instruction file becomes empty only because Trellis created it, remove
may delete it under the existing clean-exit rule. Otherwise it remains.
Unpaired, duplicate, nested, or ambiguous markers stop destructive cleanup and
are reported; remove does not guess. A second remove is a reported no-op.

## Derived resources and documentation

The implementation change updates all of these source/derivative pairs
together:

| Source or contract | Required derivative / guard |
|---|---|
| Terminal sentinel in generated `rules.md` + fixed post-import footer in both `trellis-<p>.md` variants | Payload checksum manifest, hook terminal-position validator, bootstrap boundary-proof wording, equality guard for both footer variants, Claude nested-import positive/negative fixtures, and a test asserting exactly one sentinel-plus-footer boundary in each host's complete context |
| Generated Codex bootstrap | Payload checksum manifest, setup paste oracle, remove marker logic, and byte-preservation tests |
| Codex hook implementation | Codex hook registration/manifest and production input/output fixtures |
| Plugin tree | Bundle-wide manifest used by the vendoring path and its regenerate-and-diff guard |
| Changed reference payload | Payload `version` regenerated from the rendered content and repository self-overlay sync guard updated |
| Setup/remove behavior | Both skills' host branches, reversibility tests, and plugin README |
| Supported-surface boundary | Root and plugin README support tables naming only trusted local Codex fresh startup for Phase 1 |
| Codex hook runtime | Setup preflight, Node.js `>=20` documentation, bootstrap-only degradation message, and runtime matrix tests |

The Claude manifest, Claude import block, and Claude staleness behavior remain
covered by regression tests. Documentation must distinguish:

- installed project files as authority versus plugin files as source;
- row activation at the next host context-loading boundary versus refresh;
- native hook success versus best-effort fallback;
- fresh trusted local Codex startup versus excluded Codex boundaries/surfaces;
- product-wide remove versus the nonexistent per-host disable; and
- ordinary refresh versus the separately deferred, confirmed preset reset;
- Node.js `>=20` as a local native-hook prerequisite versus the
  no-project-runtime bootstrap fallback.

No README, manifest default, help text, or setup confirmation may claim an
excluded Phase 2–4 boundary.

## Acceptance criteria

### EARS requirements

- **R1 — authority.** The Codex transport shall derive rule content only from
  the four installed project inputs listed under `Authoritative inputs and
  validity`.
- **R2 — live rows.** When a consumer edits a valid row, the next supported
  host context-loading boundary shall read the edited `rules.toml` without
  requiring setup, refresh, or regeneration.
- **R3 — in-flight limit.** Trellis shall not claim that a row edit mutates an
  already-running model context.
- **R4 — narrow registration.** The Codex manifest shall register the native
  context hook through `./hooks/codex-hooks.json` only for
  `SessionStart(startup)`.
- **R5 — host isolation.** While running in Codex, the plugin shall not emit
  Claude-shaped hook output; while running in Claude, it shall not emit the
  Codex context response.
- **R6 — project isolation.** When `cwd` is absent, relative, nonexistent, or
  has no nearest Git boundary containing an overlay at or below that boundary,
  the Codex hook shall emit only the exact failure `systemMessage` object and
  shall not guess a project from `$PWD` or search a parent repository.
- **R7 — success response.** When the startup event and overlay are valid, the
  hook shall emit exactly the specified `hookSpecificOutput` JSON, containing
  each installed source once and no more than 8,000 UTF-8 bytes of
  `additionalContext`.
- **R8 — invalid-overlay response.** When any required input is missing,
  unreadable, empty where prose is required, malformed, or fails setup's rule
  validation, the hook shall emit a visible diagnostic and no rule payload.
- **R9 — diagnostic honesty.** The hook shall not claim that Trellis loaded
  when it emitted no valid rule payload.
- **R10 — stamp role.** The installed `payload@<hash>` shall be diagnostic and
  shall not supersede the installed payload or activation rows as authority.
- **R11 — bootstrap content.** The generated Codex bootstrap shall contain the
  receipt, duplicate check, fallback paths, validity conditions, and failure
  disclosure, and shall contain no rules or row values.
- **R12 — best-effort boundary.** Documentation and tests shall describe the
  bootstrap as best-effort model instruction, not deterministic enforcement.
- **R13 — single-copy behavior.** Only when context contains the exact
  completion-sentinel-plus-fixed-footer adjacency and activation TOML
  satisfying the full validity predicate shall the bootstrap instruct the
  agent not to reload Trellis; marker, bootstrap mention, sentinel alone, or
  bare slug-name presence shall not prove a complete load. This is a
  best-effort receipt for setup-verified generated files, not a claim of
  tamper-proof source-boundary detection in flattened Claude context.
- **R14 — partial context.** When exactly one of a valid sentinel-plus-footer
  boundary or fully valid activation TOML is present, the bootstrap shall
  instruct the agent to load only the missing component.
- **R15 — missing-hook tolerance.** When no native hook marker is present, the
  bootstrap shall attempt installed files rather than treating marker absence
  as failure.
- **R16 — terminal fallback failure.** When neither native transport nor
  fallback yields a valid overlay, the bootstrap shall require a visible
  “Trellis was not loaded” disclosure and prohibit a governed-execution claim.
- **R17 — Claude setup isolation.** While setup/refresh runs in Claude, it
  shall manage only the Claude block and shall preserve `AGENTS.md`
  byte-for-byte.
- **R18 — Codex setup isolation.** While setup/refresh runs in Codex, it shall
  manage only the Codex bootstrap and shall preserve `CLAUDE.md`
  byte-for-byte.
- **R19 — shared overlay.** Setup in either host shall seed or refresh the same
  `.trellis/` layout and validation contract.
- **R20 — consumer authority.** While a valid `rules.toml` exists, ordinary
  setup/refresh shall not overwrite it.
- **R21 — marker safety.** When a managed marker pair is absent, setup shall
  append one block; when exactly one valid pair exists, setup shall replace
  only that pair; when markers are ambiguous, setup shall stop with no project
  file changed.
- **R22 — co-installation.** Running both host setup branches in either order
  shall produce exactly one host-correct transport per host and no full rule
  copy in `AGENTS.md`.
- **R23 — product-wide remove.** When remove runs from either host, it shall
  strip all recognized Trellis-managed host blocks before deleting the shared
  overlay.
- **R24 — remove preservation.** Remove shall preserve bytes outside managed
  markers and shall preflight all recognized markers and consent before any
  edit; any ambiguity shall leave every block and the overlay unchanged.
- **R25 — remove idempotency.** When no Trellis-managed block or overlay
  remains, remove shall make no change and shall report a no-op.
- **R26 — manifest propagation.** When a generated payload, bootstrap,
  manifest, or hook registration changes, its declared derivatives and
  regenerate-and-diff guards shall advance in the same change.
- **R27 — Claude regression.** The Phase 1 change shall preserve Claude's
  import block, live-row semantics, and staleness behavior.
- **R28 — docs boundary.** Phase 1 documentation shall claim only trusted local
  Codex fresh startup and shall list every excluded boundary in this spec.
- **R29 — no preset reset.** Ordinary setup/refresh shall not replace rows,
  strictness, or `seeded_from` as a wholesale configuration reset.
- **R30 — no hidden runtime.** Rule delivery shall remain instructions plus
  host-native integration and shall not require a Trellis daemon, network
  service, or project runtime. Node.js `>=20` is an explicit local Codex-plugin
  hook prerequisite, not a project dependency.
- **R31 — exact input.** The Codex hook shall accept only JSON with
  `hook_event_name = "SessionStart"`, `source = "startup"`, and an absolute
  existing-directory `cwd`; it shall find the nearest `.git` file/directory
  boundary and select an overlay only from `cwd` through that boundary.
- **R32 — host precedence.** A set `PLUGIN_ROOT` shall either validate and
  select Codex (even if `CLAUDE_PLUGIN_ROOT` is set) or stop; only an absent
  `PLUGIN_ROOT` permits a valid Claude root and manifest to select Claude.
  Each selected manifest shall parse as JSON with `name = "trellis"`.
  Invalid, wrong-plugin, or conflicting set roots shall stop visibly before
  writes.
- **R33 — stamp shape.** After trimming at most one terminal newline, the
  installed version shall match `^payload@[0-9a-f]{12}$`; the hook shall not
  rehash content.
- **R34 — deterministic expansion.** The hook shall replace exactly one exact
  `@rules.md` placeholder with installed `rules.md` content once; zero or
  multiple placeholders shall fail.
- **R35 — completion sentinel.** Assembly shall expose the completion sentinel
  exactly once as the terminal line of manifest-covered installed `rules.md`.
  `trellis.md` and the Claude block shall contain no sentinel. The Codex hook
  shall validate that terminal position and all inputs before writing stdout.
  Both generated `trellis.md` variants shall place the same exact fixed footer
  immediately after the expanded rules, and the bootstrap shall require that
  observable sentinel-plus-footer adjacency.
- **R36 — budget guard.** Generation/build and contract tests shall fail when
  the largest supported successful assembly exceeds 8,000 UTF-8 bytes; runtime
  overflow shall emit failure `systemMessage` only.
- **R37 — floor warning.** A false floor row shall remain a successful
  additional-context delivery and shall add only the exact optional top-level
  `systemMessage` warning specified by this contract.
- **R38 — setup atomicity.** Setup shall validate config/seed choice, runtime,
  all sources, manifests, targets, and markers before its first write; any
  failed preflight shall leave every project file unchanged.
- **R39 — remove atomicity.** Remove shall validate all recognized targets,
  markers, and required consent before its first edit; any failed preflight
  shall leave every block and the shared overlay unchanged.
- **R40 — runtime degradation.** When Node.js `>=20` is unavailable, Codex
  setup shall report bootstrap-only delivery, install no claim of native hook
  success, and keep the installed-file fallback usable.
- **R41 — hook diagnostic vocabulary.** Every Codex-hook failure shall use the
  one exact JSON template, one enumerated input/path label, and one enumerated
  validation class defined under `Input and output`; an unenumerated hook
  diagnostic is a contract failure. Setup and remove retain their human-facing
  report contracts and do not emit hook-protocol JSON.

### GWT scenarios

**S1 — valid local startup**

- **Given** a trusted local repository, the locally installed Codex plugin, a
  production-shaped startup event, and a valid installed overlay,
- **When** Codex invokes `SessionStart(startup)`,
- **Then** the hook emits one valid additional-context response containing the
  current installed payload, current rows, sentinel, and installed stamp once.

**S2 — row edit becomes live**

- **Given** a prior successful startup and a consumer edit that changes one
  valid rule row,
- **When** a new supported local startup begins without setup or refresh,
- **Then** its additional context contains the edited row and does not contain
  the previous row value.

**S3 — hook did not run**

- **Given** `AGENTS.md` contains the Codex bootstrap, the installed overlay is
  valid, and no hook marker is in context,
- **When** the agent follows the bootstrap before substantive work,
- **Then** it attempts the four installed files and does not fail merely
  because the marker is absent.

**S4 — Claude already loaded Trellis**

- **Given** Claude imported `AGENTS.md`, then loaded the existing Claude
  Trellis imports so that sentinel and complete rows are present,
- **When** it encounters the shared Codex bootstrap,
- **Then** the bootstrap directs it to use the loaded context and read no
  Trellis component again.

**S5 — partial context**

- **Given** exactly one of the loaded-context sentinel or complete activation
  rows is present,
- **When** the bootstrap evaluates context,
- **Then** it directs the agent to read only the missing component and the
  resulting effective context has one payload copy and one row copy.

**S6 — invalid installed overlay**

- **Given** a production-shaped startup event and one missing, unreadable,
  empty, malformed, unknown-row, missing-row, or invalid-stamp input,
- **When** the Codex hook runs,
- **Then** it emits a protocol-valid diagnostic naming the failure, emits no
  rule payload, and leaves the bootstrap able to attempt the installed files.

**S7 — both transports fail**

- **Given** the native hook did not deliver rules and the bootstrap cannot
  validate the installed overlay,
- **When** substantive work would begin,
- **Then** the agent is instructed to report that Trellis was not loaded and
  not claim governed execution.

**S8 — malformed or unbounded event root**

- **Given** a startup event whose `cwd` is missing/malformed, has no Git
  boundary, or belongs to a nested repository whose own boundary has no
  overlay while a parent repository does,
- **When** the hook parses the event,
- **Then** it emits the exact enumerated failure and reads neither process
  working directory, plugin checkout, nor parent repository as the project.

**S9 — excluded lifecycle event**

- **Given** a valid event fixture for resume, clear, compact, or subagent start,
- **When** the Phase 1 registration and handler are evaluated,
- **Then** no Trellis rule context is emitted and documentation does not mark
  the event supported.

**S10 — Codex setup beside Claude**

- **Given** a repository whose `CLAUDE.md` already carries the valid Claude
  adapter and Trellis block,
- **When** Codex setup runs,
- **Then** it refreshes the shared overlay, installs one Codex bootstrap in
  `AGENTS.md`, and leaves `CLAUDE.md` byte-identical.

**S11 — Claude setup beside Codex**

- **Given** a repository whose `AGENTS.md` already carries a valid Codex
  bootstrap,
- **When** Claude setup runs,
- **Then** it refreshes the shared overlay, installs one Claude Trellis block,
  and leaves `AGENTS.md` byte-identical.

**S12 — ordinary refresh preserves rows**

- **Given** a valid, consumer-edited `rules.toml`,
- **When** setup refresh runs in either host,
- **Then** generated files and the current host block advance while
  `rules.toml` remains byte-identical.

**S13 — marker ambiguity**

- **Given** duplicate, unpaired, nested, or otherwise ambiguous Trellis
  markers in the current host's instruction file,
- **When** setup would patch the file,
- **Then** it reports the ambiguity and changes no project file, including the
  overlay and both instruction files.

**S14 — product-wide remove from either host**

- **Given** valid Claude and Codex managed blocks, surrounding user prose, and
  a shared overlay,
- **When** remove runs from Claude or Codex,
- **Then** both managed blocks and the overlay are removed, all surrounding
  prose is byte-preserved, and the removal report names each action.

**S15 — remove ambiguity**

- **Given** an ambiguous managed marker set,
- **When** remove runs,
- **Then** it reports the ambiguity, does not guess at marker ownership, and
  changes neither managed block nor the shared overlay.

**S16 — remove rerun**

- **Given** a successful product-wide remove,
- **When** remove runs again,
- **Then** it changes no file and reports that Trellis is already absent.

**S17 — derivative drift**

- **Given** a changed sentinel, Codex bootstrap, hook registration, or plugin
  payload,
- **When** regenerate-and-diff and manifest tests run without the required
  derivatives updated,
- **Then** at least one named guard fails with the stale source/derivative pair.

**S18 — support-claim audit**

- **Given** the Phase 1 README, manifests, default hook registration, and test
  inventory,
- **When** they are audited together,
- **Then** only trusted local Codex fresh startup is claimed and every excluded
  lifecycle event and surface remains unregistered or explicitly unsupported.

**S19 — exact expansion and budget**

- **Given** valid input and installed files with exactly one `@rules.md`
  placeholder,
- **When** the hook assembles context,
- **Then** it replaces that placeholder with installed `rules.md` once,
  validates its one terminal sentinel, produces at most 8,000 UTF-8 bytes, and
  exposes the sentinel only in the complete result; a missing/nonterminal/
  duplicate sentinel, zero/multiple placeholders, or overflow emits the exact
  failure `systemMessage` object only.

**S20 — host detection with overlapping environment**

- **Given** valid Codex and Claude plugin-root variables are both present,
- **When** setup detects its host,
- **Then** the `PLUGIN_ROOT/.codex-plugin/plugin.json` that parses with
  `name = "trellis"` selects Codex; an invalid or wrong-plugin Codex root does
  not silently fall through when the roots conflict.

**S21 — missing Node runtime**

- **Given** Codex setup in a trusted repository without Node.js `>=20`,
- **When** setup preflights native delivery,
- **Then** it reports bootstrap-only degradation, installs the valid overlay
  and Codex fallback, and does not claim native hook success.

**S22 — false floor row**

- **Given** an otherwise valid overlay with one or both floor rows false,
- **When** the startup hook succeeds,
- **Then** stdout contains the exact success `hookSpecificOutput` plus the
  exact optional top-level floor-warning `systemMessage`, and the floor rules
  remain in `additionalContext`.

**S23 — Claude rules import fails**

- **Given** the valid Claude block and `trellis.md`, but a missing, unreadable,
  or sentinel-absent installed `rules.md`,
- **When** Claude loads the import path and then evaluates the shared
  bootstrap,
- **Then** no valid sentinel-plus-fixed-footer boundary proves completion, the
  bootstrap does not take the no-reload branch, and failure to recover the
  installed rules is disclosed rather than called governed execution.

Adversarial post-setup edits that copy the public sentinel/footer bytes into a
different generated-file position are not part of S23: no text-only bootstrap
can distinguish that after Claude flattens imports. The next setup checksum
verification must reject that generated-file drift.

## Required test coverage

| Test class | Required observations | Scenarios |
|---|---|---|
| Production Codex hook fixtures | Exact normalized input; nearest-ancestor resolution; exact success/failure JSON; malformed root; excluded events | S1, S6, S8, S9 |
| Current-row contract | Two startups around a row-only edit with no setup/refresh between them | S2 |
| Assembly and dedup | Exact single-placeholder expansion, source ordering/count, 8,000-byte guard, completion-sentinel placement, 14-slug completeness, full/partial/no-context branches | S1, S3–S5, S19 |
| Invalid-overlay matrix | Each required path missing/unreadable, empty prose, exact stamp-regex failures, invalid strictness, unknown/missing row, zero/multiple placeholders, and over-budget assembly | S6, S7, S19 |
| Floor-override validation | False floor rows retain successful exact JSON and add only the exact optional top-level warning | S22 |
| Host-isolation setup matrix | Fresh and refresh runs for Claude-only, Codex-only, and co-installed repositories; before/after byte snapshots | S10–S13 |
| Host/runtime preflight | Root precedence with both variables, invalid/conflicting manifests, Node `>=20`, and bootstrap-only degradation | S20, S21 |
| Setup transaction atomicity | Every config/source/manifest/marker preflight failure produces a whole-project before/after identity | S13 |
| Product-wide remove matrix | Remove from each host, co-installed cleanup, all-target preflight, consent/marker ambiguity, whole-project preservation, rerun | S14–S16 |
| Manifest and sync guards | Known-stale fixtures for each source/derivative pair; payload version regeneration | S17 |
| Documentation/registration audit | README claim allowlist and default matcher allowlist | S9, S18 |
| Existing regression suite | Claude nested-import positive/negative sentinel fixtures, imports/staleness, setup validation, self-overlay sync, vendoring manifest, build/vet/format/diff gates | S4, S10–S17, S23 |

A mocked JSON object invented solely for the unit test is not sufficient
evidence for the production event schema. The supported startup fixture must
come from the production Codex contract or a captured local positive control,
with volatile values normalized and provenance stated beside the fixture.

## Open questions

- Which exact Codex local plugin-root and project-root fields are stable across
  versions? Phase 1 fixes behavior through the production fixture; a schema
  change is a visible contract-test failure, not permission to guess.
- Do excluded Codex surfaces expose the same trust, root, and hook schema?
  Answer separately in Phase 3 before expanding support.
- On subagent start, does context arrive by inheritance, hook execution, or
  both? Answer with positive and negative controls in Phase 2 before
  registration.

## Rubric check

Self-checked against `core/rubrics/artifact-contract.md`:

| Check | Result | Evidence |
|---|---|---|
| 1. Required frontmatter present and typed | PASS | `id`, `type`, `status`, `depends_on`, and `owner` are present; lists and strings are well-typed. |
| 2. Type and lifecycle | PASS | `type: spec`; `status: gated` is in this repo's declared lifecycle. |
| 3. Unique id | PASS | Repository scan finds no other `spec-0007`. |
| 4. Dependencies resolve | PASS | `decision-0058`, `spec-0004`, and `spec-0006` resolve locally. The specs predate version materialization and carry no version to pin; the append-only decision must not be version-pinned. |
| 5. Directional flow | PASS | Dependencies are approved/ratified/gated, never draft. |
| 6. Required spec sections | PASS | `## Acceptance criteria` and `## Open questions` are present. Acceptance criteria include both EARS requirements and GWT scenarios. |
| 7. Supersede integrity | N/A | This new spec supersedes no artifact. |
| 8–11. Typed catalog/profile checks | N/A | This artifact is neither a signature catalog nor an expression profile. |
| 12. Version cross-check | N/A | `version: 1` initializes this behavioral spec; `changes:` applies to significant-change decisions, not this spec. |
| Honesty clause | PASS | Best-effort fallback and unsupported surfaces are stated without stronger enforcement or support claims. |

The check passes, so the spec is promoted `draft → gated`. `approved` remains
a human intent act. Independent spec-adversary review is still required.
