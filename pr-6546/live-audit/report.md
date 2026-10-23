# AO harness audit

Started 2026-10-23T16:06:55.600Z · AO checkout source (c9e14d966 2026-10-19) · Chat bridge: not built

| Agent | Pass | Fail | Blocked | Skip |
|---|---|---|---|---|
| codebuddy | 5 | 1 | 0 | 1 |

## codebuddy

| Check | Result | Why | Evidence |
|---|---|---|---|
| cli_installed | PASS | found native/lib/node_modules/@tencent-ai/codebuddy-code/bin/codebuddy |  |
| cli_version | PASS | 2.164.0 | `evidence/codebuddy/02-cli_version.json` |
| ao_registered | PASS | AO knows this harness | `evidence/codebuddy/03-ao_registered.json` |
| ao_installed | PASS | AO found the installed CLI | `evidence/codebuddy/04-ao_installed.json` |
| ao_authenticated | FAIL | Authentication check was inconclusive. | `evidence/codebuddy/05-ao_authenticated.json` |
| ao_models | PASS | AO uses free-form model entry (the CLI has no model list) | `evidence/codebuddy/06-ao_models.json` |
| live | SKIP | AO cannot launch the harness yet (see above); live checks did not run |  |

### Fix (codebuddy)

**AO implementation**

- `ao_authenticated`: Authentication check was inconclusive.. Look in: `backend/internal/adapters/agent/codebuddy/ (auth probe)`, `backend/internal/service/agentauth/`

_Fix in the AO implementation, add a Go test that fails without the fix, then rerun this agent._

