# Cursor Plan Usage Design

## Summary

Add Cursor subscription quota reporting to the provider-neutral plan-usage system introduced by PR #4218. Cursor does not currently publish a supported personal-plan usage API or JSON CLI command, so AO will isolate the integration behind a version-gated private-protocol reader. The reader will consume the same structured provider data used by Cursor CLI's `/usage` screen; it will not scrape terminal text or create a visible AO session.

This change is delivered in a Cursor-only PR based on `FEAT-SUBSCRIPTION-USAGE`. It does not include Kimi support.

## Goals

- Show the Cursor plan name, billing reset, Included, Auto, API, and On-Demand usage when Cursor reports them.
- Preserve provider-reported monetary usage, limit, remaining value, and unlimited/disabled/unavailable states.
- Refresh without an AO session or visible Cursor conversation.
- Contain the undocumented integration so Cursor changes fail closed without corrupting stored quota.
- Reuse PR #4218's account-level refresh, persistence, history, API, and Plan Usage surfaces.

## Non-goals

- Do not send `/usage` to a TUI, capture a PTY, remove ANSI sequences, or parse rendered terminal columns.
- Do not expose Cursor credentials or private protocol payloads through AO's API or logs.
- Do not promise compatibility with unverified Cursor CLI versions.
- Do not implement Cursor enterprise analytics or organization-wide reporting.
- Do not combine this work with Kimi support.

## Compatibility strategy

Cursor usage is experimental because the provider protocol is undocumented. AO will support an explicit compatibility table keyed by the exact Cursor CLI build identifier. The initial verified build is `2026.08.11-e8db854`; additional builds require fixtures and a compatibility-table update.

The reader resolves the installed `cursor-agent`, obtains its exact version, and selects a matching private protocol implementation. A version that is absent from the table is treated as unsupported. AO preserves the last successful snapshot and records a clear sanitized refresh error. It never guesses field layouts or silently converts missing fields to zero.

The protocol implementation is isolated behind an internal interface:

```go
type UsageClient interface {
	ReadUsage(ctx context.Context) (RawUsage, error)
}
```

The production implementation invokes the authenticated Cursor dashboard transport corresponding to these structured operations:

- `getPlanInfo`
- `getCurrentPeriodUsage`
- `getHardLimit`

The helper reuses Cursor's local authenticated profile and never returns access or refresh tokens to the quota normalizer. Provider protocol code, version matching, and raw DTOs remain inside the Cursor adapter package. Tests substitute a fake `UsageClient`; no test calls Cursor or the network.

Before shipping, the implementation must pass a live compatibility probe against the verified Cursor build. If an authenticated structured call cannot be made without copying credentials into command arguments, environment variables, logs, or AO storage, implementation stops and the Cursor PR is not opened. Terminal scraping is not an approved fallback.

## Provider-neutral model extension

Cursor's On-Demand display cannot be represented faithfully by only `remaining_value` and `total_value`: the provider can report current spend above a configured hard limit, and it distinguishes active, unlimited, disabled, and unavailable states. Add these optional fields to `domain.QuotaLimit`:

```go
UsedValue *float64
State     QuotaLimitState
```

`QuotaLimitState` accepts `active`, `unlimited`, `disabled`, and `unavailable`. An empty state remains valid for older/provider-neutral snapshots and is normalized to `active` only when numeric quota data makes that state unambiguous.

Add migration `0109_extend_provider_quota_limits.sql` with nullable `used_value` and non-null `limit_state` defaulting to an empty string. Do not modify migration `0108_provider_quota.sql`. Update quota queries, store mapping, sqlc artifacts, API DTOs, OpenAPI, and generated frontend types. History and alert thresholds continue to use percentages and reached transitions; the new monetary fields do not create new alert semantics.

## Response normalization

The Cursor adapter emits provider `cursor`, account `default`, completeness `complete`, and capabilities for read, history, and spend limits.

Plan metadata maps to `PlanType`, a non-secret account label when available, and the billing-cycle reset timestamp shared by the included buckets.

Structured usage maps as follows:

- `included`: `usage_credits`, account scope, provider total percentage.
- `auto`: `usage_credits`, account scope, provider Auto percentage.
- `api`: `usage_credits`, account scope, provider API percentage.
- `on_demand`: `spend_limit`, account scope, provider current spend as `UsedValue`, hard limit as `TotalValue`, derived or provider-reported remainder as `RemainingValue`, provider currency as `Unit`, and the exact state.

Missing Included/Auto/API categories are omitted rather than reported as zero. Percentages are retained as provider-reported values and clamped only by the progress-bar renderer. Monetary values must use a single provider-reported currency; mixed or missing currencies render as unavailable rather than being converted by AO.

## Daemon integration and refresh lifecycle

Daemon startup obtains the Cursor plugin from the existing agent resolver and registers `cursor/default` only when the binary, authenticated profile, and compatible version are present.

Cursor uses the existing startup, Plan Usage page, manual, five-minute, and idle refresh triggers. The existing quota service singleflight and freshness window coalesce account reads. No session-level cache or per-session private RPC client is added.

## Frontend behavior

The generic Plan Usage card gains provider-neutral rendering for numeric quota values and limit states:

- percentage buckets retain the existing progress bar;
- fixed On-Demand shows used, total, and remaining currency values even when used exceeds total;
- unlimited shows `Unlimited` without a misleading progress bar;
- disabled shows `Disabled`;
- unavailable shows `Unavailable`;
- an over-limit fixed bucket uses exhausted styling based on `Reached`/severity.

Rendering switches on category/state, not provider name. Kimi, Codex, Claude, and future providers can use the same value/state behavior.

## Errors and retained state

- Missing Cursor binary, login, or compatible version: account discovery reports unsupported, unless a prior Cursor snapshot exists; a prior snapshot receives a sanitized refresh error.
- Private protocol/auth failure, timeout, incompatible response, or missing required fields: reject the observation and retain the last successful snapshot.
- Raw private payloads and credentials are never logged.
- Errors identify the failing operation and installed build but exclude request metadata and tokens.
- Cancellation and the adapter timeout terminate all helper work and child processes.

## Files and responsibilities

- `backend/internal/domain/quota.go`: `UsedValue` and `QuotaLimitState`.
- Domain/service tests: normalization, merging, severity, and old snapshots with empty state.
- `backend/internal/storage/sqlite/migrations/0109_extend_provider_quota_limits.sql`: additive quota columns.
- `backend/internal/storage/sqlite/queries/quota.sql`, generated sqlc files, and quota store/tests: persistence and round trips.
- `backend/internal/httpd/controllers/dto.go`, quota mapping/tests, OpenAPI, and frontend schema: wire fields.
- `backend/internal/adapters/agent/cursor/quota.go`: account refresher and normalization.
- `backend/internal/adapters/agent/cursor/quota_protocol.go`: compatibility table and private usage-client implementation.
- Cursor adapter tests/fixtures: supported build, unsupported build, all usage states, malformed payloads, auth failures, timeout, and sanitization.
- `backend/internal/daemon/daemon.go` and wiring tests: register `cursor/default` and include Cursor in idle refreshes.
- `frontend/src/renderer/components/usage/PlanUsagePage.tsx` and tests: provider-neutral numeric/state rendering.
- Locale message catalogs: labels for used/total/remaining, unlimited, disabled, and unavailable states.

## Testing and verification

All automated provider tests use fakes or recorded sanitized JSON fixtures. Required cases include:

- exact supported-build selection and unknown-build rejection;
- Included, Auto, and API percentage mapping;
- fixed On-Demand under, at, and above its hard limit;
- unlimited, disabled, and unavailable On-Demand states;
- absent categories are omitted rather than zeroed;
- incompatible payload, auth error, timeout, and cancellation;
- old SQLite rows round-trip with empty state;
- API and generated schema expose optional numeric/state fields;
- frontend renders each state without provider-specific branching;
- errors contain no fixture credentials or raw authorization metadata.

Verification includes the live compatibility probe for the exact supported Cursor build, narrow Go/frontend tests, `npm run sqlc`, `npm run api`, backend tests, frontend typecheck/build, and repository lint.

## Acceptance criteria

- An authenticated, explicitly supported Cursor CLI installation appears as `cursor/default` without an active session.
- AO displays the same structured Included, Auto, API, On-Demand, plan, and reset information available to Cursor's `/usage` UI.
- Unsupported Cursor builds fail closed and never produce a fabricated zero-usage snapshot.
- Fixed, unlimited, disabled, unavailable, and over-limit spend states remain distinguishable through SQLite, API, and UI.
- No Cursor credential or raw private payload is persisted or logged.
- The PR contains no Kimi implementation.
