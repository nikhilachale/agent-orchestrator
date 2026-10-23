# CodeBuddy Code real Electron UI proof

Login/setup and model UI proof. Does not qualify full harness lifecycle, restore, permission enforcement or authenticated UI execution.

- native-shell: **PASS** — Real Electron renderer with window.ao preload, own checkout and isolated daemon; actual product API metadata captured.
- account-action: **PASS** — Unknown auth retains an available native Login/Setup action. Source: 8a7252bfbb1eff54cda3e035305348efe0730eb3.
- native-model-guidance: **PASS** — Native freeform/direct model-entry guidance displayed; real catalog has zero listed models and no fabricated entries. Source: 8a7252bfbb1eff54cda3e035305348efe0730eb3.
- native-login-chooser: **PASS** — Native first-run CodeBuddy login-method chooser opened after trusting only its owned empty auth workspace; no site/account selected. Source: 9a0060f5e684c5d605ea32813a6792382afd5a1e.
- direct-model-serialization: **PASS** — Actual outgoing New Task payload contains the selected native model. Request aborted before daemon acceptance; no successful response mocked.
- connected-in-this-ui-profile: **BLOCKED** — This UI proof deliberately used a fresh no-credential profile. No login was completed or key entered. Separate root-owned live provider checks do not change this UI run.
- practical-task-after-default-policy-fix: **PASS** — Ordinary unmodified form omitted approvalMode; actual daemon persisted default, native setup/login gate reached, then owned session killed and app stopped. No provider inference. Source: 8a7252bfbb1eff54cda3e035305348efe0730eb3.
- manual-login-reopen: **NOT_RUN** — First-run chooser was already open; injecting /login plus Enter could select the default site.

Final practical source: 8a7252bfbb1eff54cda3e035305348efe0730eb3. Native version: 2.173.0. Persisted permission: default. Owned session terminated; production identity HTTP200.

Evidence: practical-policy.json, practical-session-seed.json, practical-stop.json. Account/Models current screenshots: frozen-head-settings/.

Version provenance caveat: the final practical native executable was2.173.0 from the private Node24 toolchain prefix, shadowing pinned2.164.0. See practical-version-provenance.json. Earlier2.164.0 chooser evidence is retained.

Pinned native API retry: **PASS** on native2.164.0 at final source8a7252bf. Ordinary noapprovalMode API request persisted default, actual pinned process path was verified, real mux viewport reached native login-method chooser, and own session/daemon were stopped. This is API/native evidence, not an additional UI interaction. See exact-native-api-retry/. Original native2.173 UI evidence is also retained under native-2.173-path-shadow/.
