# AO harness follow-up: real AO screenshots and fresh restore proof

This follow-up supplements the [historical session report](ao-agent-harness-session-2026-10-10.md). All builds, provider calls, fixtures, UI execution and capture ran on the authorized Ubuntu VPS; the laptop was used for SSH transport only. Each integration was tested at its own revision. These are not combined-branch or main results.

The [contributor skill](../../.agents/skills/ao-agent-harness/SKILL.md) provides integration intake and an API-first native/AO lifecycle audit. It introduces a new standing-instruction token only after confirmed kill, then verifies consumption after exact native-session restore. Real AO Electron capture and authorized PR publication are host workflow steps: the runner does not automatically launch the UI or publish PRs.

## Fresh results

All three fresh runs proved initial and second mutations, same-native-ID/workspace restore, initialized empty restored composer, post-restore mutation, history continuity and newly injected post-kill instructions. None received a strict PASS. Credential presence remains configured rather than verified; native model listing/comparison remains NOT_RUN. Open Interpreter also retains an empty AO model catalog FAIL.

| Harness / attempt | PASS | FAIL | BLOCKED | NOT_RUN | Strict verdict | Published report |
| --- | ---: | ---: | ---: | ---: | --- | --- |
| Tau / tau-002 | 25 | 0 | 1 | 2 | BLOCKED (diagnostic) | [#6493](https://github.com/OrchestratorInc/agent-orchestrator/pull/6493#issuecomment-6100859313) |
| Neovate Code / neovate-001 | 25 | 0 | 1 | 2 | BLOCKED (diagnostic) | [#6494](https://github.com/OrchestratorInc/agent-orchestrator/pull/6494#issuecomment-6100907238) |
| Open Interpreter Rust / open-interpreter-002 | 24 | 1 | 1 | 2 | FAIL (diagnostic) | [#6496](https://github.com/OrchestratorInc/agent-orchestrator/pull/6496#issuecomment-6101253645) |

Cancellation evidence covers configured cancel input during an observed active turn and the same terminal settling. It does not establish termination of an already-running shell child. The Tau transcript shows the requested wait completing after restore.

## Preserved failures and the Open Interpreter fix

Tau attempt 001 retained 21 PASS, 4 FAIL, 1 BLOCKED and 2 NOT_RUN. The provider added an extra closing brace to the second JSON mutation. No manual file repair was made; attempt 002 used a fresh workspace with the same functional source, native pin, provider profile and contract.

Open Interpreter attempt 001 retained 19 PASS, 2 FAIL, 2 BLOCKED and 5 NOT_RUN. Its initial standing-instruction token was missing from native developer context; restored terminal readiness also failed, so no continuation was sent. The corrected readiness contract still requires an initialized empty composer with the expected model/workspace.

Native Rust 0.0.56 previews SessionStart context above its default 2,500-token limit, preserving the ends and spilling the middle to a file. AO emitted full instructions, but model-facing context omitted the token in the middle. [Fix 330590df92575a958dc7a507d16654079b17b4fa](https://github.com/OrchestratorInc/agent-orchestrator/commit/330590df92575a958dc7a507d16654079b17b4fa) sets additionalContextLimit=0 only on AO-owned SessionStart (full delivery in the pinned native version) and updates the exact trust hash. Other native/user hooks are unchanged. The existing callback timeout is a time bound, not a byte/token cap.

The released-native regression uses a middle canary in approximately 19 KiB of instructions: it failed before the fix and passed after it. A tail-only canary was preserved as an inadequate control because previews keep the ends. Corrected live native history independently contains both the original token and the fresh post-kill token without a spill footer.

## Actual AO screenshots

These are real scrot captures from owned Xvfb displays running isolated AO Electron. The native preload bridge and exact session were confirmed. After each finalized runner stopped, the supervisor reopened the same isolated data with the identical verified binary. Capture-time native-identity and restored-generation rechecks are explicitly recorded for Tau-002, Neovate-001 and Open Interpreter-001/002. Tau-001 records retained-session recovery and its audit identity, but no explicit capture-time digest/generation recheck; it supplies failure-state UI evidence. These are later retained-session captures, not original gate-time screenshots. Original gate-time images were unavailable because the API-first audit preceded the GUI.

Images were visually inspected for relevance and credentials, and published bytes checked against recorded SHA-256 hashes. Screenshots show visible state; functional assertions establish lifecycle and instruction continuity. Open Interpreter retains upstream “Ask Codex” wording; pinned native binary and session provenance identify Rust Open Interpreter.

### Tau

Native 0.4.7, TUI, Z.ai glm-5.3-flash. Run finished 2026-10-10T18:30:55.127Z. Functional source 189ca905dee3264ba5c682590ae04bcd8da88adf; binary SHA-256 4c82435bec3bac6f08b1073d416c683b4035e20721b3d1b6d345a2a09216d4ef; evidence head b98a1bfb4ebad89a0e6501de204623ab69bb4bb2.

[Detailed gates, failures, provenance and cleanup](https://github.com/nikhilachale/agent-orchestrator/blob/b98a1bfb4ebad89a0e6501de204623ab69bb4bb2/docs/research/tau-20261010-followup.md).

![Tau: Retained original failed audit](https://raw.githubusercontent.com/nikhilachale/agent-orchestrator/b98a1bfb4ebad89a0e6501de204623ab69bb4bb2/docs/research/assets/tau-20261010-followup/tau-001-ao-restored-failure.png)

Capture 2026-10-10T18:26:18.392600Z; AO session audit-tau-4f1b6208-1; native-ID SHA-256 aac1e75b36b34a0f4b0fa47fd9aee3cba5357353628d2d57f9755b6f969037d9; AO source 189ca905dee3264ba5c682590ae04bcd8da88adf; UI source 189ca905dee3264ba5c682590ae04bcd8da88adf; binary SHA-256 4c82435bec3bac6f08b1073d416c683b4035e20721b3d1b6d345a2a09216d4ef; PNG SHA-256 6ca864dc3d6fadb1e321f37245bc9d774f103c791826a4b1e8a4b5d4b05bfa56.

![Tau: Functional diagnostic success](https://raw.githubusercontent.com/nikhilachale/agent-orchestrator/b98a1bfb4ebad89a0e6501de204623ab69bb4bb2/docs/research/assets/tau-20261010-followup/tau-002-ao-restored-success.png)

Capture 2026-10-10T18:32:59.570527Z; AO session audit-tau-0fb37380-1; native-ID SHA-256 a3f87013b77caef5c8860009aea4a55d567173d68e86562b9bed16afd4d8fdad; AO source 189ca905dee3264ba5c682590ae04bcd8da88adf; UI source 189ca905dee3264ba5c682590ae04bcd8da88adf; binary SHA-256 4c82435bec3bac6f08b1073d416c683b4035e20721b3d1b6d345a2a09216d4ef; PNG SHA-256 1902fbcc696799f3b6cf36d80041e9b7248d35acb93d17cfc988833d7c841f69.

### Neovate Code

Native 0.28.5, TUI, Z.ai glm-5.3-flash. Run finished 2026-10-10T18:37:20.725Z. Functional source e82a48fa09c07e7d70cdd19a26aea0ae1bed7e1f; binary SHA-256 7c0340a91489e8dc9e2e306a821eee79e4aad313d343b85201470f53792fb31f; evidence head 49e850d3f4a676e14046755487c751d99bc94ad7.

[Detailed gates, failures, provenance and cleanup](https://github.com/nikhilachale/agent-orchestrator/blob/49e850d3f4a676e14046755487c751d99bc94ad7/docs/research/neovate-20261010-followup.md).

![Neovate Code: Functional diagnostic success](https://raw.githubusercontent.com/nikhilachale/agent-orchestrator/49e850d3f4a676e14046755487c751d99bc94ad7/docs/research/assets/neovate-20261010-followup/neovate-001-ao-restored-success.png)

Capture 2026-10-10T18:39:19.976133Z; AO session audit-neovate-c3090324-1; native-ID SHA-256 b9fb3eadd73e423e648ee1959b17b836afcae16183298fd7119ec281e8316a35; AO source e82a48fa09c07e7d70cdd19a26aea0ae1bed7e1f; UI source e82a48fa09c07e7d70cdd19a26aea0ae1bed7e1f; binary SHA-256 7c0340a91489e8dc9e2e306a821eee79e4aad313d343b85201470f53792fb31f; PNG SHA-256 b082322cfc46f9faace73554c01a54c5ef817a69ee4e41e972d8ef754ee68cfc.

### Open Interpreter Rust

Native Rust 0.0.56, TUI, Z.ai glm-5.3-flash. Run finished 2026-10-10T19:02:25.394Z. Functional source 330590df92575a958dc7a507d16654079b17b4fa; binary SHA-256 8d84f311dac085d2b1d9df1c07019b54e6f5d21816b19e4467d3a2ebc4617c42; evidence head 74d89ab52a951f69538742fd34bf8c947df0f841.

[Detailed gates, failures, provenance and cleanup](https://github.com/OrchestratorInc/agent-orchestrator/blob/74d89ab52a951f69538742fd34bf8c947df0f841/docs/research/open-interpreter-20261010-followup.md).

![Open Interpreter Rust: Retained original FAIL/BLOCKED attempt](https://raw.githubusercontent.com/OrchestratorInc/agent-orchestrator/74d89ab52a951f69538742fd34bf8c947df0f841/docs/research/assets/open-interpreter-20261010-followup/open-interpreter-001-ao-restored-blocked.png)

Capture 2026-10-10T18:48:33.530050Z; AO session audit-open-interpreter-a9b07959-1; native-ID SHA-256 22f95998e786c4f086203476940a56ea5b6036dcf179e43b11c145a60e0291f4; AO source 88c7431c427dda8c08802e9612042545a4febaf3; UI source 88c7431c427dda8c08802e9612042545a4febaf3; binary SHA-256 be272b67a4fecd59c60016957ec3376123dd9899897cda268011dac7cab2c980; PNG SHA-256 0f720af3fb8835e7abdd48c2a35ae215b9a500686a6c329959a94ae3d22cb669.

![Open Interpreter Rust: Corrected functional diagnostic success](https://raw.githubusercontent.com/OrchestratorInc/agent-orchestrator/74d89ab52a951f69538742fd34bf8c947df0f841/docs/research/assets/open-interpreter-20261010-followup/open-interpreter-002-ao-restored-success.png)

Capture 2026-10-10T19:04:31.649461Z; AO session audit-open-interpreter-1b4e895b-1; native-ID SHA-256 0bd058dae3bfcc86e66555bd2aeb6f10594e019e682d7b343680344da5696756; AO source 330590df92575a958dc7a507d16654079b17b4fa; UI source 330590df92575a958dc7a507d16654079b17b4fa; binary SHA-256 8d84f311dac085d2b1d9df1c07019b54e6f5d21816b19e4467d3a2ebc4617c42; PNG SHA-256 1df692b9cfdd088808dac4d7ddc5ecd021830ee48e5206085176f83ef9a50d61.

## Other session candidates

| Candidate | Recorded outcome |
| --- | --- |
| [MiniMax Code #6520](https://github.com/OrchestratorInc/agent-orchestrator/pull/6520#issuecomment-6098803293) | 28 PASS, 1 authentication BLOCKED, no FAIL/NOT_RUN. Real AO success plus three preserved failure captures. Evidence head e83c478c962d3a0b43e617061e3232af078e668d: 23/23 CI passed. |
| [memcode #6515](https://github.com/OrchestratorInc/agent-orchestrator/pull/6515#issuecomment-6096737821) | Unregistered: native v0.38.1 exact resume panics. Native-only failure evidence, no execution-inside-AO claim. |
| [Strands #6517](https://github.com/OrchestratorInc/agent-orchestrator/pull/6517#issuecomment-6097575371) | Unregistered: native cancellation returns while the shell child continues and writes a delayed artifact. Native-only failure captures; AO lifecycle/UI BLOCKED / NOT_RUN. |

MiniMax actual AO success capture; the linked report includes all failed attempts and complete provenance:

![MiniMax: restored session inside actual AO](https://raw.githubusercontent.com/nikhilachale/agent-orchestrator/e83c478c962d3a0b43e617061e3232af078e668d/docs/research/assets/minimax/minimax-ao-004-restored-success-dark.png)

Capture 2026-10-10T14:53:58.011881Z, actual AO Electron/scrot/Xvfb; AO/UI source 17a1467a404234f34d8bc8824d293993e1703564; binary SHA-256 4b47bd780af069862dffd9ace01bb768fca6e5ae9959956da15fe7e75faabf7e; AO session audit-minimax-code-acbea22e-1; native-ID SHA-256 fd7a69d95770906f65b69ddf93dcc012a3457326327446cb044a2cb5706525a4; PNG SHA-256 30419d725e8f6ed3d39547e4d28041d8f9befbfedd2d2afa7e1f67747ef14da0.

## Validation and limitations

Runner source 1c7b3b45b78751bab64d644b066853d410d55351; audit_ao.mjs SHA-256 8887aa2e7101b9e6d3950bfca08ffddf2e1e6c1c3f1d030c8dec2676443aecea. That runner passed 145/145 fixtures and 5/5 GitHub checks. Fixtures establish runner behavior rather than live-provider conformance. This follow-up adds documentation only; its final-head CI is reported separately on the skill PR.

Tau and Neovate each passed 24/24 GitHub checks on their evidence heads. The linked Open Interpreter PR report records final-head CI and exact broader VPS test outcomes, including failures/timeouts. Focused Open Interpreter adapter/CLI race tests passed on corrected source. No complete-suite pass is inferred from focused or provider passes.

The corrected Open Interpreter full VPS test suite was deliberately interrupted after approximately eight minutes when its exact functional-source CI finished 24/24 successful. It remains INCOMPLETE with observed failures in five packages: fake-agent adapter, OpenCode V2 adapter, HTTP controllers, integration and agent service. Local full vet, lint and race were NOT_RUN in this follow-up; their verification is CI-only. These failures were not proven to be baseline failures.

MiniMax’s earlier full VPS race attempt was incomplete at 30 minutes with failures in five packages; its complete 23/23 remote CI results are separate. An earlier VPS frontend typecheck exited 134 and was not rerun after memory pressure. Native OS, frontend and container results are taken from exact-head CI, not labeled Linux VPS passes.

The historical Open Interpreter HTTP-200-during-initialization case could leave an unsubmitted draft. This fresh run used strong native composer readiness; that early-send behavior is not declared fixed.

These results apply to the recorded native versions, source revisions, model/profile and TUI path. They do not qualify Chat, other models/releases, user-draft cleanup or shell-subprocess reaping. Historical 17/17 labels and terminal mirrors remain unchanged with their caveats; these fresh runs supersede only the missing refreshed-instruction and actual-AO-screen proof.

## Cleanup

Each audited session was API-terminated and confirmed. Owned Electron/daemon executables and isolated run-file provenance were checked before stopping. Neovate’s surviving owned scratch auth PTY host was also stopped; incomplete cleanup observations remain recorded. Sources, profiles, original reports and provenance binaries remain on the VPS. Credentials, profile contents, daemon databases and temporary worktrees are not committed. No merge or release was performed.
