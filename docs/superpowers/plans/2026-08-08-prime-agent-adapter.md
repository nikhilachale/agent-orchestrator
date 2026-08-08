# Prime Agent Adapter Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the safe V1 `prime-agent` harness using AO's terminal runtime, Prime's client-owned ephemeral mode, and an AO-managed lifecycle extension.

**Architecture:** A new `primeagent` Go adapter owns command construction, binary resolution, extension installation, and activity derivation. Existing registry-driven daemon boundaries provide catalog, launch, and active-steering behavior; append-only storage/API/frontend changes make the harness selectable and durable without adding a Prime RPC transport.

**Tech Stack:** Go 1.x, Cobra, SQLite/Goose, embedded TypeScript, React/TypeScript, OpenAPI generation, npm scripts.

## Global Constraints

- Harness ID is exactly `prime-agent`; display name is exactly `Prime Agent`.
- Launch only with `--no-session`; never emit `--continue`, `--resume`, or persistent-session flags.
- Use the existing terminal runtime and `ports.Agent` boundary; add no RPC or network transport.
- Launch argv order is binary, `--no-session`, `--extension`, optional system prompt, optional model, then `--` plus task.
- Omit model/system-prompt flags for whitespace-only values; fail on unreadable nonblank prompt files.
- Ignore every AO permission mode because Prime Agent has no equivalent safe permission contract.
- Install the extension atomically and idempotently at `<AO_DATA_DIR>/agent-runtime/prime-agent/ao-activity.ts`.
- Hook failures are best effort and must never break Prime Agent.
- Do not edit merged migrations or generated storage code; add migration `0032_allow_prime_agent_harness.sql`.
- Do not add network calls to tests.
- Commit generated OpenAPI YAML and frontend schema together.

---

### Task 1: Prime Agent adapter, binary resolver, and lifecycle extension

**Files:**
- Create: `backend/internal/adapters/agent/primeagent/primeagent.go`
- Create: `backend/internal/adapters/agent/primeagent/hooks.go`
- Create: `backend/internal/adapters/agent/primeagent/activity.go`
- Create: `backend/internal/adapters/agent/primeagent/assets/ao-activity.ts`
- Create: `backend/internal/adapters/agent/primeagent/primeagent_test.go`
- Create: `backend/internal/adapters/agent/primeagent/hooks_test.go`
- Create: `backend/internal/adapters/agent/primeagent/activity_test.go`

**Interfaces:**
- Consumes: `ports.Agent`, `ports.AgentBinaryResolver`, `ports.ActiveTurnSteerer`, `ports.ActivitySignaler`, `agentbase.Base`, `binaryutil.ResolveBinary`, and `hookutil.AtomicWriteFile`.
- Produces: `func New() *Plugin`, `func ResolvePrimeAgentBinary(context.Context) (string, error)`, `func (p *Plugin) SteersActiveTurn() bool`, `func DeriveActivityState(string, []byte) (domain.ActivityState, bool)`, and manifest ID `prime-agent`.

- [ ] **Step 1: Write failing adapter contract tests**

Add table tests that pin this exact representative argv and related edge cases:

```go
want := []string{
    "prime-agent", "--no-session",
    "--extension", filepath.Join(dataDir, "agent-runtime", "prime-agent", "ao-activity.ts"),
    "--append-system-prompt", "follow repo rules",
    "--model", "openai/gpt-5",
    "--", "-fix the check",
}
```

Cover manifest/config spec, prompt delivery, optional flag omission, inline/file prompt precedence, missing prompt file, ignored permission modes, unavailable restore, active steering, submit activity true, blocked activity false, cancellation, PATH resolution, Node-manager fallback, and `ErrAgentBinaryNotFound`.

- [ ] **Step 2: Run the adapter tests and verify the package is missing**

Run: `cd backend && go test ./internal/adapters/agent/primeagent`

Expected: FAIL because the package implementation does not exist or its symbols are undefined.

- [ ] **Step 3: Implement the minimal adapter and resolver**

Implement `Plugin` with cached binary resolution and the exact command rules. The restore method must be:

```go
func (p *Plugin) GetRestoreCommand(ctx context.Context, _ ports.RestoreConfig) ([]string, bool, error) {
    if err := ctx.Err(); err != nil {
        return nil, false, err
    }
    return nil, false, nil
}
```

Use whitespace only to decide whether optional values are blank; preserve nonblank system-prompt contents and task contents. Require `LaunchConfig.DataDir` before deriving the extension path.

- [ ] **Step 4: Write failing extension installer and activity tests**

Assert two installs produce the same managed file, stale content is replaced, mode is `0600`, cancelled contexts fail, and a blank data dir fails. Assert the embedded source subscribes to `session_start`, `before_agent_start`, `agent_start`, `agent_end`, and `session_shutdown`; invokes all four normalized commands; includes a timeout; invokes `ao` without prompt interpolation; and catches callback/spawn failures.

Pin activity mapping:

```go
tests := map[string]domain.ActivityState{
    "session-start":      domain.ActivityActive,
    "user-prompt-submit": domain.ActivityActive,
    "stop":               domain.ActivityIdle,
    "session-end":        domain.ActivityExited,
}
```

- [ ] **Step 5: Implement extension installation, source, and derivation**

Embed `assets/ao-activity.ts`. Store the prompt from `before_agent_start`, emit it at `agent_start`, and use guarded synchronous child-process execution with JSON stdin, `ctx.cwd`, ignored stdout, captured diagnostics, and a bounded timeout. Return from every caught failure without throwing.

- [ ] **Step 6: Run focused adapter tests**

Run: `cd backend && go test ./internal/adapters/agent/primeagent`

Expected: PASS.

- [ ] **Step 7: Commit the adapter unit**

```bash
git add backend/internal/adapters/agent/primeagent
git commit -m "feat: add Prime Agent adapter"
```

### Task 2: Domain, registry, activity dispatch, daemon, and catalog wiring

**Files:**
- Modify: `backend/internal/domain/harness.go`
- Modify: `backend/internal/domain/harness_test.go` if present; otherwise add coverage in the nearest domain harness validation test
- Modify: `backend/internal/adapters/agent/registry/registry.go`
- Modify: `backend/internal/adapters/agent/registry/registry_test.go`
- Modify: `backend/internal/adapters/agent/activitydispatch/dispatch.go`
- Modify: `backend/internal/adapters/agent/activitydispatch/dispatch_test.go`
- Modify: `backend/internal/daemon/wiring_test.go`
- Modify: `backend/internal/service/agent/catalog_test.go`

**Interfaces:**
- Consumes: `primeagent.New`, `primeagent.DeriveActivityState`, and adapter capability interfaces from Task 1.
- Produces: `domain.HarnessPrimeAgent`, domain validation, registry resolution, activity support, daemon active steering, and catalog label `Prime Agent`.

- [ ] **Step 1: Add failing domain and registry tests**

Require `domain.AgentHarness("prime-agent").IsKnown()` to be true, require registry construction to include exactly one `prime-agent` adapter, and require its manifest name to equal `Prime Agent`.

- [ ] **Step 2: Run the focused domain/registry tests**

Run: `cd backend && go test ./internal/domain ./internal/adapters/agent/registry`

Expected: FAIL because `HarnessPrimeAgent` and its constructor are absent.

- [ ] **Step 3: Add the domain constant and registry constructor**

Add:

```go
HarnessPrimeAgent AgentHarness = "prime-agent"
```

Append it to `AllHarnesses`, import `primeagent`, and add `primeagent.New()` next to `pi.New()` in the stable constructor list.

- [ ] **Step 4: Add failing activity, daemon, and catalog wiring tests**

Require `activitydispatch.SupportsHarness(domain.HarnessPrimeAgent)`, verify `session-end` derives exited through the dispatcher, require `buildAgentResolver` to return the Prime adapter, require `activeTurnSteering` to return true for Prime, and require the supported catalog entry `{ID: "prime-agent", Label: "Prime Agent"}`.

- [ ] **Step 5: Wire activity dispatch and satisfy registry-driven paths**

Import `primeagent` in `activitydispatch`, register:

```go
"prime-agent": primeagent.DeriveActivityState,
```

Update explicit expected harness lists in daemon/catalog tests without adding Prime-specific daemon production branches.

- [ ] **Step 6: Run focused wiring tests**

Run: `cd backend && go test ./internal/domain ./internal/adapters/agent/registry ./internal/adapters/agent/activitydispatch ./internal/daemon ./internal/service/agent`

Expected: PASS.

- [ ] **Step 7: Commit the integration wiring**

```bash
git add backend/internal/domain backend/internal/adapters/agent/registry backend/internal/adapters/agent/activitydispatch backend/internal/daemon/wiring_test.go backend/internal/service/agent/catalog_test.go
git commit -m "feat: wire Prime Agent harness"
```

### Task 3: Append-only SQLite harness migration

**Files:**
- Create: `backend/internal/storage/sqlite/migrations/0032_allow_prime_agent_harness.sql`
- Modify: `backend/internal/storage/sqlite/migrate_test.go`

**Interfaces:**
- Consumes: the exact current `sessions.harness` CHECK including `fake` from migration 0026.
- Produces: a migrated schema accepting `prime-agent` and a down schema identical to the pre-0032 constraint.

- [ ] **Step 1: Add the failing migration expectation and insert test**

Add `domain.HarnessPrimeAgent` to `TestMigrateAllowsEveryShippedHarness`. Add a focused test that runs migrations, creates a project, and inserts a session row whose harness is `prime-agent` without a CHECK violation.

- [ ] **Step 2: Run the migration test and verify failure**

Run: `cd backend && go test ./internal/storage/sqlite -run 'TestMigrateAllowsEveryShippedHarness|PrimeAgent'`

Expected: FAIL because the live CHECK omits `prime-agent`.

- [ ] **Step 3: Add migration 0032**

Copy the guarded `PRAGMA writable_schema` structure from migration 0026. Replace the exact current CHECK text with the same text plus `'prime-agent'` in Up, and reverse only that addition in Down. Keep `-- +goose NO TRANSACTION` and `PRAGMA writable_schema = RESET`.

- [ ] **Step 4: Run migration tests**

Run: `cd backend && go test ./internal/storage/sqlite`

Expected: PASS.

- [ ] **Step 5: Commit the migration**

```bash
git add backend/internal/storage/sqlite/migrations/0032_allow_prime_agent_harness.sql backend/internal/storage/sqlite/migrate_test.go
git commit -m "feat: allow Prime Agent sessions in SQLite"
```

### Task 4: CLI and generated HTTP API contract

**Files:**
- Modify: `backend/internal/cli/spawn.go`
- Modify: `backend/internal/cli/spawn_test.go`
- Modify: `backend/internal/cli/agent_test.go`
- Modify: `backend/internal/cli/dto_drift_e2e_test.go` if its explicit enum fixture requires the new harness
- Modify: `backend/internal/httpd/controllers/dto.go`
- Modify: `backend/internal/httpd/controllers/dto_test.go`
- Modify: `backend/internal/httpd/apispec/openapi.yaml` (generated)
- Modify: `frontend/src/api/schema.ts` (generated)

**Interfaces:**
- Consumes: `domain.HarnessPrimeAgent` and manifest catalog wiring from Task 2.
- Produces: CLI-visible harness help and the `prime-agent` OpenAPI enum consumed by the frontend.

- [ ] **Step 1: Add failing CLI and DTO enum tests**

Assert `ao spawn --help` includes `prime-agent`, an explicit Prime spawn request is serialized unchanged, and reflected `SpawnSessionRequest.harness` enum includes `prime-agent` exactly once.

- [ ] **Step 2: Run focused contract tests**

Run: `cd backend && go test ./internal/cli ./internal/httpd/controllers ./internal/httpd/apispec/...`

Expected: FAIL because CLI help and DTO enum omit Prime.

- [ ] **Step 3: Update source CLI help and controller enum**

Add `prime-agent` after `pi` in the visible CLI list and in the controller's harness enum tag. Do not hand-edit generated files yet.

- [ ] **Step 4: Regenerate the API artifacts**

Run: `npm run api`

Expected: `openapi.yaml` and `frontend/src/api/schema.ts` both gain `prime-agent` in the spawn harness enum.

- [ ] **Step 5: Run API parity and CLI tests**

Run: `cd backend && go test ./internal/cli ./internal/httpd/...`

Expected: PASS.

- [ ] **Step 6: Commit source and generated contract together**

```bash
git add backend/internal/cli backend/internal/httpd/controllers/dto.go backend/internal/httpd/controllers/dto_test.go backend/internal/httpd/apispec/openapi.yaml frontend/src/api/schema.ts
git commit -m "feat: expose Prime Agent in CLI and API"
```

### Task 5: Frontend fallback catalog, public branding, and user documentation

**Files:**
- Create: `frontend/src/renderer/lib/agent-options.test.ts`
- Modify: `frontend/src/renderer/lib/agent-options.ts`
- Modify: `frontend/src/landing/components/LandingAgentsBar.tsx`
- Modify: `frontend/src/landing/components/LandingHero.tsx`
- Modify: `frontend/src/landing/app/design-partners/page.tsx`
- Modify: `frontend/src/landing/app/landing/layout.tsx`
- Modify: `frontend/src/landing/app/layout.tsx`
- Modify: `frontend/src/landing/content/docs/plugins/agents/index.mdx`
- Modify: `frontend/src/landing/content/docs/plugins/agents/meta.json`
- Create: `frontend/src/landing/content/docs/plugins/agents/prime-agent.mdx`
- Modify: `README.md`

**Interfaces:**
- Consumes: generated frontend harness type and catalog identity `Prime Agent`.
- Produces: renderer fallback option, public Prime Agent listing/favicon, count 24, and documentation of safe V1 omissions.

- [ ] **Step 1: Add a failing fallback catalog test**

Assert `AGENT_OPTIONS` contains `prime-agent` exactly once and remains free of duplicates.

- [ ] **Step 2: Run the focused frontend test**

Run: `cd frontend && npm test -- --run src/renderer/lib/agent-options.test.ts`

Expected: FAIL because `prime-agent` is absent.

- [ ] **Step 3: Add Prime to frontend/public catalog surfaces**

Add `prime-agent` next to `pi` in `AGENT_OPTIONS`. Add the landing chip:

```ts
{ name: "Prime Agent", id: "prime-agent", src: "https://www.google.com/s2/favicons?domain=primeintellect.ai&sz=64" }
```

Change current public counts from 23 to 24 and “20 more” metadata copy to “21 more”. Add Prime to the README supported harness list.

- [ ] **Step 4: Write Prime Agent documentation**

Create the agent page with binary `prime-agent`, the exact `--no-session` launch shape, model configuration, extension path, active steering, fresh-only restore semantics, and a warning that AO permission modes do not sandbox Prime's IPython/host commands. Add it to the docs navigation and overview table/grid.

- [ ] **Step 5: Run focused frontend tests and typecheck**

Run: `cd frontend && npm test -- --run src/renderer/lib/agent-options.test.ts`

Run: `npm run frontend:typecheck`

Expected: PASS for both.

- [ ] **Step 6: Commit frontend and documentation changes**

```bash
git add frontend/src/renderer/lib/agent-options.ts frontend/src/renderer/lib/agent-options.test.ts frontend/src/landing README.md
git commit -m "docs: document Prime Agent support"
```

### Task 6: Full verification, review, push, and PR

**Files:**
- Modify only files required to fix verified failures or review findings.

**Interfaces:**
- Consumes: all completed tasks.
- Produces: verified branch and a PR against `main` that reports intentional V1 omissions.

- [ ] **Step 1: Run focused Go verification**

Run: `cd backend && go test ./internal/adapters/agent/primeagent ./internal/adapters/agent/registry ./internal/adapters/agent/activitydispatch ./internal/domain ./internal/daemon ./internal/service/agent ./internal/storage/sqlite ./internal/cli ./internal/httpd/...`

Expected: PASS.

- [ ] **Step 2: Regenerate contracts and verify no drift**

Run: `npm run api`

Run: `git status --short`

Expected: generation succeeds and produces no uncommitted drift beyond intentional fixes.

- [ ] **Step 3: Run repository-required checks**

Run: `npm run lint`

Run: `npm run frontend:typecheck`

Run: `cd backend && go test ./...`

Run: `cd frontend/src/landing && npm run build`

Expected: PASS for every command, including the Electron renderer typecheck and
the separately configured Next.js landing/docs build.

- [ ] **Step 4: Review requirements and code quality**

Compare the final diff to `docs/superpowers/specs/2026-08-08-prime-agent-adapter-design.md`. Review command ordering, lifecycle safety, extension failure containment, cancellation, migration exactness, API generation, and public permission/restore warnings. Fix each confirmed finding test-first and commit with a focused conventional message.

- [ ] **Step 5: Push the feature branch**

Run: `git push -u origin ao/agent-orchestrator-167/root`

Expected: branch push succeeds without force.

- [ ] **Step 6: Open the PR against main**

Use a conventional PR title such as `feat: add Prime Agent adapter`. The body must summarize terminal/ephemeral architecture, extension activity mapping, model/catalog/API/storage/docs work, list exact test commands, and explicitly state that V1 omits native restore and permission-mode enforcement.
