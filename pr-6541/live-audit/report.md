# AO harness audit

Started 2026-10-23T16:23:37.416Z · AO checkout repo (adb222647 2026-10-19) · Chat bridge: not built

| Agent | Pass | Fail | Blocked | Skip |
|---|---|---|---|---|
| mastra-code | 4 | 2 | 0 | 1 |

## mastra-code

| Check | Result | Why | Evidence |
|---|---|---|---|
| cli_installed | PASS | found native/node_modules/mastracode/dist/cli.js |  |
| cli_version | FAIL | `mastracode --version` exited 1 | `evidence/mastra-code/02-cli_version.json` |
| ao_registered | PASS | AO knows this harness | `evidence/mastra-code/03-ao_registered.json` |
| ao_installed | PASS | AO found the installed CLI | `evidence/mastra-code/04-ao_installed.json` |
| ao_authenticated | FAIL | Authentication check was inconclusive. | `evidence/mastra-code/05-ao_authenticated.json` |
| ao_models | PASS | AO uses free-form model entry (the CLI has no model list) | `evidence/mastra-code/06-ao_models.json` |
| live | SKIP | AO cannot launch the harness yet (see above); live checks did not run |  |

### Fix (mastra-code)

**AO implementation**

- `ao_authenticated`: Authentication check was inconclusive.. Look in: `backend/internal/adapters/agent/mastracode/ (auth probe)`, `backend/internal/service/agentauth/`

_Fix in the AO implementation, add a Go test that fails without the fix, then rerun this agent._

**Harness contract or upstream CLI**

- `cli_version`: `mastracode --version` exited 1

_The CLI itself fails outside AO: fix the contract's command, or this is an upstream CLI problem._

