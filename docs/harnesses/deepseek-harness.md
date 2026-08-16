# DeepSeek Harness Adapter

> **Status:** Developer preview. The upstream `dsh@0.1.0-rc.6` launcher is
> unstable and may break between releases. Expect the adapter to be revised
> as DeepSeek ships a stable CLI surface.

DeepSeek Harness (`dsh`) is integrated into Agent Orchestrator (AO) as an
experimental agent harness. AO launches `dsh --profile headless` as a
one-shot task runner inside the project's worktree. The adapter is registered
under the `deepseek-harness` harness ID in the daemon's agent catalog.

## Install

Install the DeepSeek Harness CLI globally with npm. Node 20 or newer is
required.

```bash
npm install -g @deepseek-ai/dsh
```

AO does not auto-install `dsh` and will not fall back to `npx
@deepseek-ai/dsh` at runtime. `npx` may also time out on first invocation
because of the package's transitive dependencies, so install the launcher
explicitly before starting a session.

Verify the install:

```bash
dsh --version
```

If AO reports the binary as missing, the launcher's resolved path is not on
`PATH` and not under a well-known location. The adapter looks in:

- `PATH`
- `/usr/local/bin/dsh`, `/opt/homebrew/bin/dsh`
- The npm-managed global bin directory (`NPM_CONFIG_PREFIX`,
  `~/.npm-global/bin/dsh`, `~/Library/Pnpm/...`, or similar)
- `%AppData%\npm\dsh.cmd` and `%AppData%\npm\dsh.exe` on Windows

## Auth

DeepSeek Harness resolves credentials from the active profile. AO probes two
sources:

1. The `DEEPSEEK_API_KEY` environment variable. If set, AO treats the
   adapter as authorized and will spawn `dsh` without an interactive prompt.
2. A configured `dsh` profile (for example, a logged-in session created by
   `dsh` itself). If no environment variable is present, AO reports the
   adapter as auth-unknown and defers to whatever the profile supplies.

Set the API key before starting the daemon:

```bash
export DEEPSEEK_API_KEY=sk-...
```

Permission scoping is controlled by `DSH_PERMISSION_MODE` and is delegated to
the `dsh` profile. AO does not surface permission overrides in its config
spec; the adapter's config surface is intentionally empty.

## Supported AO modes

DeepSeek Harness is exposed only through AO's **Terminal UI** mode. The
adapter launches the headless one-shot task runner:

```bash
dsh --profile headless "<prompt>"
```

AO sets the working directory to the session's worktree and passes the
initial prompt as a single positional argument. The process prints the final
assistant response and exits; AO's process supervisor detects the exit
through the `AgentExitDetectionSupervisor` mode rather than waiting on TUI
activity hooks.

The full `dsh web` profile serves a Web UI and is not a usable AO agent
runtime: it exposes no prompt, resume, or workspace CLI flags. AO will not
launch it as a fake session inside tmux or a terminal multiplexer.

## Known limitations

- **No Chat handoff.** AO cannot hand a session off to a DeepSeek Harness
  chat surface today. The shipped launcher has no stable structured
  protocol or SDK for chat-style interactions. Until one exists, only the
  one-shot headless runner is supported.
- **No native session resume.** `--resume <id>` is documented only for the
  separately-installable `@deepseek-ai/dsh-terminal` TUI profile, which is
  not shipped with `@deepseek-ai/dsh` at the `0.1.0-rc.6` line. AO's
  `restore` flow therefore starts a fresh headless run and reports
  `ok=false` when no session metadata is available.
- **No model flag.** `dsh` profiles own model selection; no CLI flag is
  documented. AO does not append a `--model` argument.
- **Prompt-as-flag risk.** `dsh --profile headless` does not document a
  `--` separator. Prompts beginning with `-` may be misinterpreted as
  flags in the developer preview; pass them verbatim and avoid leading
  dashes if possible.

## Developer-preview compatibility warning

DeepSeek Harness is in active development upstream. The `@deepseek-ai/dsh`
package is at the `0.1.0-rc.6` release line, and DeepSeek's own
documentation warns that breaking changes are expected before a stable
release. AO mirrors that status:

- The adapter's manifest description labels the harness as a developer
  preview.
- The CLI surface, flag names, profile names, and auth shape can change
  without notice.
- Session restore, chat handoff, and any other currently-unsupported
  capabilities depend on upstream shipping a stable protocol or SDK and may
  arrive alongside breaking changes to the headless runner.

Treat DeepSeek Harness as best-effort for now, and revisit the adapter when
DeepSeek announces a stable release.
