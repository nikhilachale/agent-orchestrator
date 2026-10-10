# AO harness recovery: Strands and memcode

This recovery completes the actual **TUI-only** AO integrations for PR6517 and PR6515. Both run the official native implementation with supported extension paths, consume private standing instructions, restore exact native history and accept fresh post-Kill instructions. Neither supports Chat, handoff or switching. This supplements, rather than erases, the earlier six-image follow-up and its original blockers.

| Harness | Functional audit | Remaining strict gaps | Published evidence |
| --- | --- | --- | --- |
| Strands | 24 PASS | 1 auth BLOCKED, 1 catalog FAIL, 2 catalog NOT_RUN | [Pinned report](https://github.com/nikhilachale/agent-orchestrator/blob/022084e99b25e5aff7e31c46ebc1b6e869bf320f/docs/research/strands-tui-recovery-2026-10-11.md) · [PR comment](https://github.com/OrchestratorInc/agent-orchestrator/pull/6517#issuecomment-6102363961) |
| memcode | 24 PASS | 1 auth BLOCKED, 1 catalog FAIL, 2 catalog NOT_RUN | [Pinned report](https://github.com/nikhilachale/agent-orchestrator/blob/ef48482c61044cfd006d5745969e319a6c0cb088/docs/research/memcode-tui-recovery-2026-10-11.md) · [PR comment](https://github.com/OrchestratorInc/agent-orchestrator/pull/6515#issuecomment-6102843130) |

Both overall audit verdicts remain **diagnostic FAIL**. Actual authorized Z.ai glm-5.3-flash calls succeeded; configured credentials do not make AO's auth observation verified. The unchanged runner SHA-256 is `8887aa2e7101b9e6d3950bfca08ffddf2e1e6c1c3f1d030c8dec2676443aecea`; its existing145/145 fixtures are separate runner evidence. No runner change, synthetic screenshot or relabeled original failure.

## Supported production paths

Strands CLI0.2.0/SDK1.20.0 uses public Plugin.initAgent after selected tools exist. It wraps only the existing official shell via public addOrReplace, preserving schema/errors/Cedar permission semantics and narrowed children without shell. Each invocation delegates cancellation into official WorkspaceSandbox.execute. Unsupported custom shells/sandboxes/root configuration/custom subagent modules reject; native foreground configuration and public model-driven reconfiguration guards keep the qualified foreground route. Native Escape/Ctrl+C controls prove child-authored PID/start identity gone and no delayed artifacts. Independent actual AO active-shell Kill reaped its exact owned group in389ms with32-second no-late-write and unrelated-control checks. Native Cancelled labels alone previously left a shell running28seconds; that failure remains in the report. Manual profile/plugin takeover is unqualified.

memcode official0.38.1 direct --resume still panics during native widget mount. Supported initialized /resume FULL_ID avoids the defect. The optional manager-owned initialization holds exclusive input, validates confined bounded nonempty exact-ID history, requires fresh current-generation local witness/empty composer, sends the command once, verifies later exact-target witness/loaded messages/empty composer, then publishes and reads back native identity. Startup adoption and every direct resume caller fail closed; no transient fresh ID overwrites the target, saved task replay, timeout fallback or synchronous parked-hook deadlock. Current rendered viewport, rather than stale/incomplete raw ring tail, supplies terminal evidence. Private native instruction chunks must reproduce actual native trim/join exactly, with bounded size/count and explicit rejection. All AO model overrides reject; actual native configured model is demonstrated.

Ask-mode actual memcode control remained Blocked after selection changed to option2, ordinary Send returned409 SESSION_AWAITING_DECISION, and Up+Enter then approved option1 once. Protection is qualified after AO observes/publishes the card; the existing30-second observer ticker remains, with read/commit delays possible. Executing-shell Escape002 reaped exact owned identities593ms, native empty composer921ms, AO idle5081ms; independent AO Kill89ms. Both observed32seconds without the28-second forbidden artifact and kept an unrelated control unchanged immediately/later. Earlier15-second settle criterion FAILED despite253ms reap; it remains unchanged, separately explained by normal polling. Existing45-second audit confirmation is a test bound, not guaranteed latency. Generic turn cancellation does not establish active-shell reaping or erase historical intent.

## Real Electron captures

<!-- publication-images -->
![Real AO Strands restored conversation](https://raw.githubusercontent.com/nikhilachale/agent-orchestrator/022084e99b25e5aff7e31c46ebc1b6e869bf320f/docs/research/assets/strands-recovery-20261011/ao-003-electron-restored.png)

![Real AO memcode restored conversation](https://raw.githubusercontent.com/nikhilachale/agent-orchestrator/ef48482c61044cfd006d5745969e319a6c0cb088/docs/research/assets/memcode-recovery-20261011/ao-004-electron-restored.png)

![Real AO memcode native Ask decision](https://raw.githubusercontent.com/nikhilachale/agent-orchestrator/ef48482c61044cfd006d5745969e319a6c0cb088/docs/research/assets/memcode-recovery-20261011/ao-permission-002-pending-approval.png)
<!-- /publication-images -->

Strands capture2026-10-10T21:08:21.227532UTC: daemon source98365, UI878957, native CLI0.2.0/SDK1.20.0, glm-5.3-flash, native digest750f2289056d635ce4330cd6fcbf0002731214db6e9dbfcb0c8c7d4e86863ee8, generation1dcba694-ef1b-4c71-83c2-513d5c691724; PNG `ffae74a4e45becc91b11c019c8ece1ee128a2aa7c8651ab992c8b84e1b442303`. Later guard/ledger-only source fixes have focused/CI evidence, not extrapolated UI captures. See full report for binary SHA/session/profile boundaries.

memcode restored capture2026-10-10T22:04:34.550047UTC: daemon35a8a3d5debc3e02ae10c59d8dd6775b62b1a8cb, UI3793bde903b948788e6ebfdbca54941cdaa86e53, native0.38.1/glm-5.3-flash/allow-all, binaryb07471106c3d27a9043ea9297e9e5384a10482035b6ce40aaa5a96017a572b9f, sessionaudit-memcode-21bee11c-1, native digest592d5f9268d6a16d0ae0cc6c39190324b0914ce73c9e6495fb4b93ddb1b3453f, generation445cabf9-4e81-45a7-a816-0e2335f1bd35; PNG `cf0aba8d67e040fab4d0bedd805515acd693a9f6eaf3d86890514fb6dad44888`. Electron started a new same-binary isolated supervisor daemon after audit; live identity/generation were verified before/after capture, not inferred from the old audit PID.

memcode decision capture2026-10-10T22:05:37.710377UTC shares daemon/UI/binary/native model, uses Ask/sessionaudit-memcode-21bee11c-2, digest02152774702bdb79bcc52f8e13a96dc46c6b9dc07d76dcb9d6ad75775dfb387d, generation486bbc2b-262a-4d30-aeab-1a32b5929a13; PNG `3972c184113b8c21f520e2a4f7dee923c9326c04d6ee1d96b295b82cde8bf266`. Before/after identity/generation/Blocked matched. Visible option2 was not approved; earlier transient gateway timeout/retry remains uncropped.

## Preserved failures and scope

![Preserved real AO memcode decision observation failure](https://raw.githubusercontent.com/nikhilachale/agent-orchestrator/ef48482c61044cfd006d5745969e319a6c0cb088/docs/research/assets/memcode-recovery-20261011/ao-permission-001-pending-failure.png)

Historical Ask failure PNG `3203b6042d9c54c6b5830767567c3985b295b6ee4bb59335ddde2fb7c87d414d` at21:42:12UTC/source77c36 records native card while AO remained active. No ordinary Send attempted in that unsafe state; owned session killed without approval. Its21:43:03 result includes identity/generation, but explicit capture-time before/after identity checks were not recorded. Do not borrow stronger successful-capture provenance.

![Preserved real Electron thread-limit blank window](https://raw.githubusercontent.com/nikhilachale/agent-orchestrator/022084e99b25e5aff7e31c46ebc1b6e869bf320f/docs/research/assets/strands-recovery-20261011/ao-003-electron-thread-limit.png)

Actual Strands Electron thread-limit failure PNG `c76ba77e564904b5811a937819415c8f2adeab1045f8b51b08cf6422ca6777fd` is preserved; owned worker task capacity was raised for successful capture. Both individual reports retain all original native/AO/provider/CI failures, including headless failures whose gate-time screenshot is unavailable with an explicit reason. Original Tau001 also lacks explicit capture-time native digest/generation recheck: no blanket all-images-reconfirmed claim.

MiniMax6520 (source17a146/evidencee83c,28PASS1authBLOCKED23CI), Tau6493 (b98a1bfb,25PASS1BLOCKED2NOT_RUN24CI), Neovate6494 (49e850d3,samecounts24CI), and OpenInterpreter6496 (74d89ab/source330590,24PASS1catalogFAIL1authBLOCKED2NOT_RUN24CI) are preserved and were not rerun here. OpenInterpreter's early-send HTTP200/unsubmitted-draft behavior remains unfixed; only its readiness-aware route was proven. The earlier broad VPS suite remains incomplete with5failed packages/94test names; no baseline or full-local PASS claim.

## Validation and cleanup

Focused VPS production-boundary regressions and pinned lint passed; generated API/migration ledgers verified. Optional local race compilation aborted before tests for disk pressure, NOT_RUN. No duplicate full VPS suites/frontendOOM rerun or full-local pass. Complete final current evidence-head GitHub CI is verified separately in latest PR comments: Strands `022084e99b25e5aff7e31c46ebc1b6e869bf320f`23/23 success; memcode `ef48482c61044cfd006d5745969e319a6c0cb088` and this consolidated report head require their own completed checks, rather than earlier source-head checks. Counts in PR headers derive current GitHub changed files. Uploaded PNG bytes above were verified against local originals at these pinned URLs.

Individual harness PRs are checked against main; candidate branches have expected shared registry/enum/API/install/ledger textual collisions. Distinct0200/0201/0202 migrations avoid duplicate versions. No sibling import/combined PR/merge. All tests ran on VPS or complete remote CI; no Mac compute/connection/release.

Exact owned audit sessions were terminated, isolated loopback daemons stopped and actual Electron instances stopped. memcode final HTTP calls returned without error, but numeric statuses were not persisted after a post-shutdown ownership assertion; later durable termination/daemon absence/runfile removal were independently verified and the gap remains in its receipt. Private profiles/evidence/proof binaries retained; no credentials/configs/databases/raw unrestricted logs/run state published. Skill instruction update records only the reusable gates demonstrated here.
