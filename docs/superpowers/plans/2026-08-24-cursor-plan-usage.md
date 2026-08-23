# Cursor Plan Usage Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add version-gated Cursor subscription quota reads and faithfully render percentage and on-demand spend states.

**Architecture:** Cursor-specific protocol code produces a sanitized `RawUsage` model through a `UsageClient`; normalization converts it to the provider-neutral quota domain. The adapter fails closed for unverified builds. The quota domain gains optional used-value and state fields that flow additively through SQLite, API, and the generic Plan Usage UI.

**Tech Stack:** Go, SQLite/sqlc, OpenAPI generation, React/TypeScript, Cursor's authenticated private dashboard transport for verified build `2026.08.11-e8db854`.

**Spec:** `docs/superpowers/specs/2026-08-24-cursor-plan-usage-design.md`

## Global Constraints

- Never scrape Cursor's TUI or send `/usage` into a conversation.
- Support only exact Cursor builds with verified fixtures and protocol compatibility.
- Never expose credentials or raw private payloads in arguments, environment variables, logs, SQLite, or API responses.
- Stop implementation and do not open the Cursor PR if the verified build cannot be read through a credential-safe structured path.
- Do not include Kimi changes.

---

### Task 1: Prove the structured Cursor compatibility path

**Files:**
- Create: `backend/internal/adapters/agent/cursor/quota_protocol.go`
- Create: `backend/internal/adapters/agent/cursor/quota_protocol_test.go`

**Interfaces:**
- Produces: `type UsageClient interface { ReadUsage(context.Context) (RawUsage, error) }`
- Produces: `newUsageClient(binaryPath, version string) (UsageClient, error)`

- [ ] **Step 1: Write failing build-gate and redaction tests**

Assert exact support for `2026.08.11-e8db854`, rejection of every other version, cancellation, capped output, and credential redaction. The fake runner returns sanitized structured JSON matching Cursor's standard usage model.

- [ ] **Step 2: Confirm focused tests fail**

Run: `cd backend && go test ./internal/adapters/agent/cursor -run 'TestCursorUsageProtocol' -count=1`

- [ ] **Step 3: Implement the minimal structured helper boundary**

Use Cursor's own authenticated dashboard client for `getPlanInfo`, `getCurrentPeriodUsage`, and `getHardLimit`. The helper returns only plan name, billing reset, included percentages, on-demand kind, used dollars, limit dollars, scope, and currency. Validate the executable build before invocation and sanitize every error.

- [ ] **Step 4: Run the live compatibility probe**

Run the helper against the installed verified build and compare its sanitized fields with Cursor's visible `/usage` screen. No token or raw request metadata may appear in process inspection, stdout/stderr, or AO files. If this condition fails, stop the Cursor branch as required by the spec.

- [ ] **Step 5: Commit the proven protocol boundary**

```bash
git add backend/internal/adapters/agent/cursor/quota_protocol.go backend/internal/adapters/agent/cursor/quota_protocol_test.go
git commit -m "feat: add Cursor usage protocol client"
```

### Task 2: Extend the quota domain and storage

**Files:**
- Modify: `backend/internal/domain/quota.go`
- Test: `backend/internal/domain/quota_test.go`
- Create: `backend/internal/storage/sqlite/migrations/0109_extend_provider_quota_limits.sql`
- Modify: `backend/internal/storage/sqlite/queries/quota.sql`
- Modify: `backend/internal/storage/sqlite/store/quota_store.go`
- Test: `backend/internal/storage/sqlite/store/quota_store_test.go`
- Regenerate: `backend/internal/storage/sqlite/gen/*`

**Interfaces:**
- Produces: `type QuotaLimitState string` with active, unlimited, disabled, unavailable.
- Produces optional `QuotaLimit.UsedValue` and `QuotaLimit.State`.

- [ ] **Step 1: Write failing normalization and SQLite round-trip tests**

Assert used value/state survive normalization, complete/partial merging, persistence, and reload; assert old rows with empty state remain valid.

- [ ] **Step 2: Confirm tests fail**

Run: `cd backend && go test ./internal/domain ./internal/storage/sqlite/store -run 'Test.*Quota.*(State|UsedValue)' -count=1`

- [ ] **Step 3: Add the domain fields and additive migration/query mappings**

Add nullable `used_value REAL` and `limit_state TEXT NOT NULL DEFAULT ''` in migration 0109. Update upserts/selects and store conversions without modifying migration 0108.

- [ ] **Step 4: Regenerate sqlc and run tests**

```bash
npm run sqlc
cd backend && go test ./internal/domain ./internal/service/quota ./internal/storage/sqlite/store
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/domain backend/internal/storage/sqlite
git commit -m "feat: persist quota values and states"
```

### Task 3: Normalize Cursor usage and register its refresher

**Files:**
- Create: `backend/internal/adapters/agent/cursor/quota.go`
- Create: `backend/internal/adapters/agent/cursor/quota_test.go`
- Modify: `backend/internal/daemon/daemon.go`
- Test: relevant daemon tests.

**Interfaces:**
- Consumes: `UsageClient.ReadUsage` from Task 1.
- Produces: `NewQuotaRefresher(plugin cursorPlugin, options ...QuotaOption) *QuotaRefresher` implementing `quota.AccountRefresher`.

- [ ] **Step 1: Write failing normalization/refresher tests**

Cover Included/Auto/API omission and mapping, fixed on-demand below/above limit, unlimited, disabled, unavailable, billing reset, unsupported build, auth error, timeout, and sanitized errors.

- [ ] **Step 2: Confirm tests fail**

Run: `cd backend && go test ./internal/adapters/agent/cursor -run 'TestCursorQuota' -count=1`

- [ ] **Step 3: Implement snapshot normalization and daemon registration**

Emit `cursor/default`, completeness `complete`, usage-credit percentage limits, and an on-demand spend limit with used/total/remaining/state. Register via the existing agent resolver and add Cursor to the idle-refresh harness predicate.

- [ ] **Step 4: Run adapter, daemon, and quota tests**

Run: `cd backend && go test ./internal/adapters/agent/cursor ./internal/daemon ./internal/service/quota -count=1`

- [ ] **Step 5: Commit**

```bash
git add backend/internal/adapters/agent/cursor backend/internal/daemon
git commit -m "feat: refresh Cursor plan usage"
```

### Task 4: Expose optional value/state fields through the API

**Files:**
- Modify: `backend/internal/httpd/controllers/dto.go`
- Modify: `backend/internal/httpd/controllers/quota.go`
- Modify: `backend/internal/httpd/controllers/quota_test.go`
- Regenerate: `backend/internal/httpd/apispec/openapi.yaml`
- Regenerate: `frontend/src/api/schema.ts`

**Interfaces:**
- Produces optional `usedValue` and `state` in `QuotaLimitResponse`.

- [ ] **Step 1: Write a failing HTTP mapping test**

Assert a Cursor on-demand limit returns used, total, remaining, and `unlimited|disabled|unavailable|active` state without exposing raw provider data.

- [ ] **Step 2: Confirm failure**

Run: `cd backend && go test ./internal/httpd/controllers -run TestQuotaAPIListsProviderNeutralSnapshots -count=1`

- [ ] **Step 3: Add DTO/mapping fields and regenerate contracts**

```bash
npm run api
cd backend && go test ./internal/httpd/...
```

- [ ] **Step 4: Commit generated artifacts together**

```bash
git add backend/internal/httpd frontend/src/api/schema.ts
git commit -m "feat: expose quota values and states"
```

### Task 5: Render Cursor spend states generically

**Files:**
- Modify: `frontend/src/renderer/components/usage/PlanUsagePage.tsx`
- Modify: `frontend/src/renderer/components/usage/PlanUsagePage.test.tsx`
- Modify: locale catalogs under `frontend/src/renderer/i18n/`.

**Interfaces:**
- Consumes optional `usedValue`, `totalValue`, `remainingValue`, `unit`, and `state`.

- [ ] **Step 1: Write failing UI tests for every state**

Assert fixed over-limit currency values, unlimited, disabled, unavailable, and existing percentage bars. Assert production rendering switches on category/state rather than provider name.

- [ ] **Step 2: Confirm failure**

Run: `cd frontend && npm test -- PlanUsagePage.test.tsx`

- [ ] **Step 3: Implement provider-neutral numeric/state cards and translations**

Keep percentage bars unchanged. Add a numeric spend presentation with `Intl.NumberFormat` when currency is a valid ISO code and a unit-preserving fallback otherwise.

- [ ] **Step 4: Run frontend tests and typecheck**

```bash
cd frontend && npm test -- PlanUsagePage.test.tsx
npm run frontend:typecheck
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/renderer/components/usage frontend/src/renderer/i18n
git commit -m "feat: render quota spend states"
```

### Task 6: Full verification

**Files:** No new production files.

- [ ] **Step 1: Run generation drift checks and focused suites**

```bash
npm run sqlc
npm run api
cd backend && go test ./internal/adapters/agent/cursor ./internal/domain ./internal/service/quota ./internal/storage/sqlite/store ./internal/httpd/... ./internal/daemon
cd frontend && npm test -- PlanUsagePage.test.tsx
```

- [ ] **Step 2: Run repository checks**

```bash
npm run frontend:typecheck
npm run lint
git diff --check
git status --short
```

- [ ] **Step 3: Repeat the credential-safe live compatibility probe**

Confirm the installed supported Cursor build returns the expected sanitized model and produces no secret-bearing files or output.
