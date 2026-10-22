# AO harness audit

Started 2026-10-22T17:43:49.543Z · AO checkout ~/Desktop/agent-orchestrator/.worktrees/feat-mimo-code (beaf8950b5 2026-10-22) · Chat bridge: found

| Agent | Pass | Fail | Blocked | Skip |
|---|---|---|---|---|
| stakpak | 8 | 3 | 0 | 10 |

## stakpak

| Check | Result | Why | Evidence |
|---|---|---|---|
| cli_installed | PASS | found ~/.local/bin/stakpak |  |
| cli_version | PASS | stakpak 0.3.88 | `evidence/stakpak/02-cli_version.json` |
| cli_login | PASS | the CLI reports a logged-in account | `evidence/stakpak/03-cli_login.json` |
| ao_registered | PASS | AO knows this harness | `evidence/stakpak/04-ao_registered.json` |
| ao_installed | PASS | AO found the installed CLI | `evidence/stakpak/05-ao_installed.json` |
| ao_authenticated | PASS | AO verified the login (authorized) | `evidence/stakpak/06-ao_authenticated.json` |
| ao_models | PASS | AO uses free-form model entry (the CLI has no model list) | `evidence/stakpak/07-ao_models.json` |
| tui_spawn | PASS | AO created a TUI session | `evidence/stakpak/08-tui_spawn.json` |
| tui_prompt | FAIL | the agent did not write the proof file before the timeout (prompt not delivered or not submitted); the terminal screen is saved in the evidence file | `evidence/stakpak/09-tui_prompt.json` |
| tui_worktree | SKIP | the first prompt produced no proof file |  |
| tui_system_prompt | SKIP | the first prompt produced no proof file |  |
| tui_agents_md | SKIP | the first prompt produced no proof file |  |
| tui_activity | SKIP | the first prompt produced no proof file |  |
| tui_native_id | SKIP | the first prompt produced no proof file |  |
| tui_follow_up | SKIP | the first prompt produced no proof file |  |
| tui_usage | SKIP | the first prompt produced no proof file |  |
| tui_stop | SKIP | the first prompt produced no proof file |  |
| tui_restore | SKIP | the first prompt produced no proof file |  |
| tui_restore_history | SKIP | the first prompt produced no proof file |  |
| role_worker | FAIL | the worker gave no answer before the timeout |  |
| role_orchestrator | FAIL | the orchestrator gave no answer before the timeout |  |

### Fix (stakpak)

**AO implementation**

- `tui_prompt`: the agent did not write the proof file before the timeout (prompt not delivered or not submitted); the terminal screen is saved in the evidence file. Look in: `backend/internal/adapters/agent/stakpak/ (prompt delivery, ready cue)`
- `role_worker`: the worker gave no answer before the timeout. Look in: `backend/internal/session_manager/prompt.go`, `the harness's system-prompt channel`
- `role_orchestrator`: the orchestrator gave no answer before the timeout. Look in: `backend/internal/session_manager/prompt.go`, `the harness's system-prompt channel`

_Fix in the AO implementation, add a Go test that fails without the fix, then rerun this agent._

