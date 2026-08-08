# Prime Agent Adapter Design

## Status

Approved V1 design, recorded before implementation. Implementation planning and
code changes remain gated on human review of this document.

## Goal

Add Prime Agent as a first-class Agent Orchestrator worker harness with harness
ID `prime-agent` and display name `Prime Agent`. The adapter will run Prime Agent
inside AO's existing terminal runtime, report lifecycle activity through an
AO-managed Prime Agent extension, expose model selection, and integrate with the
same daemon, API, CLI, frontend, storage, documentation, and branding surfaces
as the existing agent adapters.

The safety goal is as important as basic launch support: terminating an AO
session must terminate the Prime Agent work that AO launched. V1 therefore uses
Prime Agent's client-owned ephemeral mode and does not create or attach to
persistent resident Prime sessions.

## Source Material

The design is based on Prime Agent's official repository and documentation:

- [Prime Agent repository](https://github.com/PrimeIntellect-ai/prime-agent)
- [Usage and CLI reference](https://github.com/PrimeIntellect-ai/prime-agent/blob/main/packages/coding-agent/docs/usage.md)
- [Daemon architecture](https://github.com/PrimeIntellect-ai/prime-agent/blob/main/packages/coding-agent/docs/daemon.md)
- [Extension API and lifecycle events](https://github.com/PrimeIntellect-ai/prime-agent/blob/main/packages/coding-agent/docs/extensions.md)

The Prime documentation distinguishes resident workers from client-owned
workers. Normal interactive sessions use detached resident workers that survive
TUI disconnection. Interactive `--no-session` uses an in-memory session with a
client-owned worker whose lifecycle follows the client. Prime also documents
`--extension`, `--append-system-prompt`, `--model`, and `--` as the relevant CLI
contracts, and documents mid-turn Enter submissions as queued steering rather
than follow-up messages.

## Selected Architecture

V1 will add a conventional daemon-side `ports.Agent` adapter, following the
nearby `pi` adapter for launch/config/binary behavior and the OpenCode adapter
for an embedded TypeScript activity extension. It will continue to use AO's
existing terminal adapter and runtime boundary. No Prime RPC, JSON, or daemon
transport will be introduced.

Two alternatives are intentionally rejected:

1. A Prime RPC or daemon integration would create a second runtime/control path
   beside AO's terminal runtime, widening lifecycle, cancellation, and recovery
   behavior without being needed for V1.
2. Persistent Prime sessions launched normally or with `--continue` would use
   resident workers. Closing the terminal client can leave those workers and
   descendants alive, so AO runtime destruction would no longer be sufficient
   to implement AO kill semantics.

The selected command shape is:

```text
prime-agent --no-session --extension <AO-managed-extension> [--append-system-prompt <instructions>] [--model <model>] -- <task>
```

`--no-session` and the explicit extension are mandatory. The system prompt and
model flags are optional. The `--` separator is mandatory whenever a task is
present so a task beginning with `-` cannot be parsed as a Prime option. When
the task is empty, both the separator and positional task are omitted and Prime
opens an ephemeral interactive session.

## Adapter Contract

### Identity and capabilities

The adapter manifest will use:

- ID: `prime-agent`
- Name: `Prime Agent`
- Description consistent with nearby worker adapters
- Agent capability only

It will implement the standard `ports.Agent` surface, the optional
`ports.AgentBinaryResolver` surface, and `ports.ActiveTurnSteerer`.
`SteersActiveTurn` will return `true`: Prime's queue distinguishes steering
submitted during the current run from a follow-up delivered after all work.
That documented steering queue is the behavior AO wants for active-session
coordination.

The adapter will declare submit activity but not blocked activity. Its managed
extension reports prompt submission, but V1 has no safe permission-prompt and
post-tool correlation contract. AO must therefore not use its Enter-based
blocked-state confirmation loop for Prime Agent.

### Configuration

`GetConfigSpec` will expose the existing typed `model` field as a string. A
nonblank value becomes `--model <trimmed-model>`. A missing or whitespace-only
value emits no model flag. No new project config shape is required.

Permission-mode mapping is intentionally unsupported. Every AO permission mode
will produce identical Prime argv with no permission flag. Prime's built-in
IPython environment can execute arbitrary host Python and commands, and Prime
does not expose a CLI permission contract equivalent to AO's modes. Pretending
that an AO permission setting constrains Prime would be unsafe; the public docs
must state that Prime Agent sessions run with the user's host permissions.

### Launch argv

`GetLaunchCommand` will resolve the executable first and then build argv in this
exact order:

1. resolved `prime-agent` executable;
2. `--no-session`;
3. `--extension` and the stable AO-managed extension path;
4. nonblank `--append-system-prompt` and its text;
5. nonblank `--model` and its trimmed value;
6. if a task exists, `--` and the task as one positional argument.

The initial prompt uses the in-command delivery strategy. AO will not trim or
rewrite task text. Tests will pin both ordinary tasks and tasks beginning with
hyphens.

For standing instructions, nonblank inline `SystemPrompt` wins. If it is blank
and `SystemPromptFile` is nonblank, the adapter reads the file and passes its
contents inline because Prime's append flag accepts text. An unreadable file is
a launch error. Whitespace-only inline text or file contents emit no system
prompt flag. This makes the blank-flag rule explicit without changing the
contents of a nonblank prompt.

### Binary resolution and cancellation

The resolver will use `binaryutil.BinarySpec` with executable name
`prime-agent`, PATH-first lookup, common `/usr/local/bin` and Homebrew locations,
and the existing Node/npm global and version-manager fallbacks used for
Node-distributed agents. Windows names and npm shim locations will follow the
same resolver conventions even though Prime's current public installer targets
macOS and Linux.

Resolution will cache the successful path under the adapter's mutex, as nearby
adapters do. All public adapter methods will check or propagate context
cancellation. Filesystem work for extension installation and prompt-file reads
will not continue after an already-cancelled context.

### Restore behavior

`GetRestoreCommand` will always return unavailable (`cmd=nil`, `ok=false`) after
honoring context cancellation. It will not inspect native Prime metadata and
will never emit `--continue`, `--resume`, or a persistent session flag.

This means V1 has no native transcript-preserving Prime restore. AO's existing
generic restore fallback may start a new ephemeral Prime process and replay the
saved task; when that happens it is a fresh/saved-prompt relaunch, not a resumed
Prime session. The relaunched command still includes `--no-session` and the
managed extension, preserving kill semantics.

`SessionInfo` remains the standard no-metadata behavior. The activity extension
does not claim a native Prime session ID because V1 deliberately cannot resume
one.

## AO-Managed Prime Extension

### Installation and ownership

The adapter will embed a real TypeScript source file under its Go package and
materialize it at this stable path:

```text
<AO_DATA_DIR>/agent-runtime/prime-agent/ao-activity.ts
```

`GetAgentHooks` will perform this installation before launch using the existing
session-manager preparation boundary. A blank data directory is an error. The
installer will create only this AO-owned directory, write the source with
owner-only file permissions using the shared atomic-write helper, and be
idempotent: an identical installed source is already satisfied, while an older
or damaged AO-owned copy is replaced atomically. Multiple session preparations
may converge safely on the same immutable source. Nothing is written into the
user repository or Prime's global config, and no workspace cleanup hook is
needed for this data-dir asset.

`GetLaunchCommand` derives the same path from `LaunchConfig.DataDir` and passes
it through Prime's explicit `--extension` flag. The extension is therefore
loaded only for the AO-launched process; it does not affect the user's unrelated
Prime Agent sessions.

### Lifecycle mapping

The extension will export the documented Prime `ExtensionAPI` factory and map
Prime lifecycle callbacks to AO's hidden hook CLI as follows:

| Prime lifecycle event | Extension behavior                                                                        | AO hook event        | AO activity |
| --------------------- | ----------------------------------------------------------------------------------------- | -------------------- | ----------- |
| `session_start`       | report ephemeral client startup                                                           | `session-start`      | active      |
| `before_agent_start`  | capture the submitted prompt for the immediately following start                          | none directly        | none        |
| `agent_start`         | report one prompt submission using the captured prompt, or an empty prompt if unavailable | `user-prompt-submit` | active      |
| `agent_end`           | report completion of that prompt's agent loop                                             | `stop`               | idle        |
| `session_shutdown`    | report extension/session teardown                                                         | `session-end`        | exited      |

The adapter will register a Prime activity deriver under token `prime-agent`.
It will map `session-start` and `user-prompt-submit` to active, `stop` to idle,
and `session-end` to exited. `before_agent_start` is used only to retain the
prompt until `agent_start`, so there is one AO submit event per Prime prompt and
the payload can still carry the prompt text.

Each hook invocation will execute `ao hooks prime-agent <event>` with a small
JSON payload on stdin and the session worktree as cwd. The existing AO runtime
environment supplies `AO_SESSION_ID`, `AO_DATA_DIR`, and the AO binary on PATH,
so no Prime session identifier is needed. Payloads contain only lifecycle facts
such as prompt or reason; the extension does not invent resumability metadata.

### Best-effort failure behavior

Hook reporting must never break Prime Agent. The extension will invoke `ao`
without interpolating prompt text into a shell command, cap each invocation,
and preserve lifecycle order. Missing binaries, spawn failures, timeouts,
non-zero exits, malformed event data, and logging failures are caught and never
re-thrown into Prime. Diagnostics may be written to stderr, while the existing
`ao hooks` command continues to record daemon delivery failures in
`AO_DATA_DIR/hooks.log` and exit successfully.

Prime's documented extension behavior also logs extension errors and continues,
but the AO extension will not rely on that final safety net: every callback has
its own best-effort guard.

## Repository Integration

### Domain, registry, daemon, and activity

`domain.HarnessPrimeAgent` will join `AllHarnesses`, making project role
validation and spawn validation accept `prime-agent`. The agent registry will
construct the adapter in the existing stable list. Because daemon resolution,
catalog construction, and lifecycle steering are registry-driven, adding the
adapter and its declared interfaces makes those paths available without a new
daemon transport or service.

Focused tests will still pin the resolver, catalog label, activity dispatcher,
activity-support policy, and adapter-backed active steering so future registry
drift is visible.

### SQLite migration

Add the new append-only migration
`0032_allow_prime_agent_harness.sql`. It will widen the `sessions.harness` CHECK
from the current list (including `fake`) to also include `prime-agent`. No merged
migration will be edited. The down migration removes only `prime-agent` and
restores the prior exact CHECK.

Migration coverage will verify the live post-migration schema contains Prime
Agent and that inserting a `prime-agent` session succeeds. The existing
all-shipped-harness migration test will include the new domain constant so an
exact-text replacement that silently no-ops is caught.

### CLI and HTTP contract

The `ao spawn --agent/--harness` help catalog will include `prime-agent` and its
tests will pin the visible help. `ao agent ls` will receive `Prime Agent` from
the daemon catalog rather than duplicating a CLI-only label.

The `SpawnSessionRequest.Harness` enum in the controller DTO will include
`prime-agent`. The code-first API will then be regenerated with `npm run api`,
updating both `backend/internal/httpd/apispec/openapi.yaml` and
`frontend/src/api/schema.ts`. DTO/spec parity and CLI wire-drift tests will cover
the updated contract.

### Frontend catalog and branding

The generated frontend schema will accept `prime-agent`, and the renderer's
fallback agent option list will include it so the project configuration UI also
works before or without a successful catalog refresh. Catalog-backed surfaces
will display the manifest label `Prime Agent`.

Public supported-agent surfaces will add Prime Agent and update the shipped
harness count from 23 to 24. The landing agent marquee will follow its existing
newer-agent convention and use the Prime Intellect site favicon for the Prime
Agent chip; README/docs lists may use the existing text-only fallback where a
dedicated local icon is not already required. No unrelated frontend behavior or
daemon logic will move into Electron.

### Documentation

The supported-agent documentation will describe:

- harness ID and executable name;
- ephemeral `--no-session` launch and AO kill semantics;
- model configuration;
- active-turn steering support;
- activity extension installation under `AO_DATA_DIR`;
- lack of native Prime restore in V1;
- explicit lack of AO permission-mode enforcement and Prime's arbitrary host
  Python/command access.

Stale supported-agent counts and lists touched by the addition will be updated,
but unrelated legacy architecture prose will not be broadly rewritten.

## Error Handling

The following are spawn-blocking errors because a safe launch cannot be built:

- cancelled context;
- unresolved `prime-agent` executable;
- missing/blank `AO_DATA_DIR` when deriving or installing the extension;
- extension directory or atomic-write failures;
- an unreadable nonblank system prompt file.

Whitespace-only model or system-prompt values are not errors; their flags are
omitted. Unsupported permission modes are normalized and ignored rather than
rejected. Hook delivery errors after Prime starts are best effort and never
become agent failures.

## Test Strategy

Implementation will be test-driven, with failing focused tests added before
each behavior. Coverage will include:

- manifest ID/name and model config spec;
- exact argv ordering, mandatory `--no-session`, extension path, and task
  separator;
- a task beginning with hyphens;
- blank/whitespace model and system prompt handling;
- inline system prompt precedence, prompt-file contents, and file read errors;
- all AO permission modes producing no Prime permission flags;
- active-turn steering and submit/blocked activity declarations;
- unavailable native restore and context cancellation;
- PATH/common-location/npm resolver behavior and not-found errors through the
  shared resolver patterns;
- extension installation location, content, permissions, replacement,
  idempotence, cancellation, and concurrent-safe convergence;
- embedded extension registration for all five Prime lifecycle events, exact
  `ao hooks prime-agent` commands, event/payload mapping, timeout, and swallowed
  failures without executing network calls;
- Prime activity derivation and dispatcher support;
- domain validation, registry construction, daemon resolver/catalog/steering
  wiring;
- the new migration and insertion of a Prime session;
- CLI help/catalog, HTTP enum/spec drift, generated frontend schema, fallback
  frontend catalog, public count, and Prime branding entry.

Extension tests will inspect or run the committed local asset with local fakes;
they will not contact Prime, npm, GitHub, model providers, or any other network
service.

## Verification and Review

After implementation, verification will run in this order:

1. focused Go tests for the Prime adapter, activity dispatcher, registry,
   daemon wiring, migration, controller, and CLI surfaces;
2. focused frontend tests for the catalog/branding surfaces;
3. `npm run api`, followed by the HTTP spec drift/parity tests;
4. `npm run lint`;
5. `npm run frontend:typecheck`;
6. broader backend tests and frontend build/checks as warranted by the files
   touched.

Before opening the PR, the final diff will receive a requirements review and a
code-quality review. The PR will target `main`, link no invented issue, use a
conventional title, list the exact verification results, and call out the two
intentional V1 omissions: native Prime session restore and permission-mode
mapping.

## V1 Non-Goals

V1 does not add:

- Prime RPC, JSON, or custom daemon transport;
- persistent Prime sessions, `--continue`, `--resume`, attach, or native
  transcript restoration;
- AO permission-mode emulation or claims of sandboxing;
- provider authentication probing beyond the existing catalog's unknown state;
- Prime-specific review-agent support;
- autonomous mode, goals, schedules, heartbeats, peer routing, or retained
  subagent management;
- changes to AO's primary loopback listener or any new network-facing bind.
