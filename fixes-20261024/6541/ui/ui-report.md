# Mastra Code real Electron UI proof

Login/setup and model UI proof. Does not qualify full harness lifecycle, restore, permission enforcement or authenticated UI execution.

- native-shell: **PASS** — Real Electron renderer with window.ao preload, own checkout and isolated daemon; actual product API metadata captured.
- account-action: **PASS** — Unknown auth retains an available native Login/Setup action. Source: 0772781da61fa446ee91c08d4989eba65a31c08d.
- native-model-guidance: **PASS** — Native freeform/direct model-entry guidance displayed; real catalog has zero listed models and no fabricated entries. Source: 0772781da61fa446ee91c08d4989eba65a31c08d.
- native-login-chooser: **PASS** — Mastra native OAuth and Add an API key chooser opened; no option selected. Source: dcc2ce5a8012c6604e1869a813ef169da305eede.
- direct-model-serialization: **PASS** — Actual outgoing New Task payload contains the selected native model. Request aborted before daemon acceptance; no successful response mocked.
- connected-in-this-ui-profile: **BLOCKED** — This UI proof deliberately used a fresh no-credential profile. No login was completed or key entered. Separate root-owned live provider checks do not change this UI run.
- practical-task-after-default-policy-fix: **PASS** — Ordinary unmodified form omitted approvalMode; actual daemon persisted default, native setup/login gate reached, then owned session killed and app stopped. No provider inference. Source: 0772781da61fa446ee91c08d4989eba65a31c08d.

Final practical source: 0772781da61fa446ee91c08d4989eba65a31c08d. Native version: 0.45.0. Persisted permission: default. Owned session terminated; production identity HTTP200.

Evidence: practical-policy.json, practical-session-seed.json, practical-stop.json. Account/Models current screenshots: frozen-head-settings/.
