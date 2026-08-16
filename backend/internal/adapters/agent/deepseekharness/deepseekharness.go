// Package deepseekharness implements the DeepSeek Harness (dsh) agent
// adapter. It is an experimental, developer-preview integration: the upstream
// `dsh@0.1.0-rc.6` launcher ships no stable interactive TUI session CLI, so AO
// drives the one-shot `dsh --profile headless` task runner and relies on its
// own process supervisor to detect exit (see ExitDetectionMode).
//
// The adapter is deliberately minimal:
//   - launches one task and exits, so AO does not need an interactive terminal;
//   - does not attempt native resume (`--resume` lives on the unshipped
//     `@deepseek-ai/dsh-terminal` profile), so GetRestoreCommand is false;
//   - does not consult or install hooks (`dsh` exposes no plugin event surface
//     at the launcher level), so the inherited no-op GetAgentHooks is correct.
package deepseekharness

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/aoagents/agent-orchestrator/backend/internal/adapters"
	"github.com/aoagents/agent-orchestrator/backend/internal/adapters/agent/agentbase"
	"github.com/aoagents/agent-orchestrator/backend/internal/adapters/agent/binaryutil"
	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
)

// adapterID is the registry id and the value users pass to
// `ao spawn --agent`. It matches domain.HarnessDeepSeek.
const adapterID = "deepseek-harness"

// dshBinarySpec locates the DeepSeek Harness CLI: PATH first, then the
// install-script and Homebrew locations, the standard Node-managed fallback
// paths, and the npm shims under %APPDATA% on Windows.
var dshBinarySpec = binaryutil.BinarySpec{
	Label:         "deepseek-harness",
	Names:         []string{"dsh"},
	WinNames:      []string{"dsh.cmd", "dsh.exe", "dsh"},
	UnixPaths:     []string{"/usr/local/bin/dsh", "/opt/homebrew/bin/dsh"},
	UnixHomePaths: binaryutil.NodeManagedUnixHomePaths("dsh"),
	NodeManaged:   true,
	WinPaths: []binaryutil.WinPath{
		{Base: binaryutil.WinAppData, Parts: []string{"npm", "dsh.cmd"}},
		{Base: binaryutil.WinAppData, Parts: []string{"npm", "dsh.exe"}},
	},
}

// Plugin is the DeepSeek Harness agent adapter. It is safe for concurrent use;
// the binary path is resolved once and cached under binaryMu.
type Plugin struct {
	agentbase.Base
	binaryMu       sync.Mutex
	resolvedBinary string
}

// New returns a ready-to-register DeepSeek Harness adapter.
func New() *Plugin {
	return &Plugin{}
}

var _ adapters.Adapter = (*Plugin)(nil)
var _ ports.Agent = (*Plugin)(nil)
var _ ports.AgentAuthChecker = (*Plugin)(nil)
var _ ports.AgentBinaryResolver = (*Plugin)(nil)
var _ ports.AgentExitDetector = (*Plugin)(nil)

// Manifest returns the adapter's static self-description. The description
// calls out the developer-preview status so the dashboard reflects the
// upstream warning that the harness is not GA.
func (p *Plugin) Manifest() adapters.Manifest {
	return adapters.Manifest{
		ID:          adapterID,
		Name:        "DeepSeek Harness",
		Description: "Run DeepSeek Harness one-shot tasks (developer preview).",
		Version:     "0.0.1",
		Capabilities: []adapters.Capability{
			adapters.CapabilityAgent,
		},
	}
}

// GetConfigSpec reports no agent-specific config keys. Model selection is
// profile-owned inside `dsh`; no CLI flag is documented, so there is nothing
// for AO to expose.
func (p *Plugin) GetConfigSpec(ctx context.Context) (ports.ConfigSpec, error) {
	if err := ctx.Err(); err != nil {
		return ports.ConfigSpec{}, err
	}
	return ports.ConfigSpec{}, nil
}

// GetLaunchCommand builds the argv to run a single DeepSeek Harness task:
//
//	dsh --profile headless <prompt>
//
// The prompt is the positional argument (no documented `--` separator exists).
// Prompts beginning with "-" may be misinterpreted as flags in the developer
// preview; callers are expected to validate that. The runtime sets cwd to
// cfg.WorkspacePath, matching every other shipped adapter.
func (p *Plugin) GetLaunchCommand(ctx context.Context, cfg ports.LaunchConfig) (cmd []string, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.Prompt) == "" {
		return nil, fmt.Errorf("deepseek-harness: initial prompt is required for headless profile")
	}

	binary, err := p.dshBinary(ctx)
	if err != nil {
		return nil, err
	}

	cmd = []string{binary, "--profile", "headless", cfg.Prompt}
	return cmd, nil
}

// SessionInfo surfaces DeepSeek Harness metadata persisted by AO under the
// shared ports.MetadataKey* keys. The headless profile does not write any of
// its own, so ok is false until something populates metadata.
func (p *Plugin) SessionInfo(ctx context.Context, session ports.SessionRef) (ports.SessionInfo, bool, error) {
	if err := ctx.Err(); err != nil {
		return ports.SessionInfo{}, false, err
	}
	info, ok := agentbase.StandardSessionInfo(session)
	return info, ok, nil
}

// ExitDetectionMode opts DeepSeek Harness into AO's process supervisor. The
// headless profile runs one task and exits, so AO must detect the process exit
// rather than waiting on TUI activity hooks.
func (p *Plugin) ExitDetectionMode() ports.AgentExitDetectionMode {
	return ports.AgentExitDetectionSupervisor
}

// ResolveDSHBinary returns the path to the `dsh` binary on this machine,
// searching PATH then a handful of well-known install locations. It returns a
// wrapped ports.ErrAgentBinaryNotFound when `dsh` is absent. The runtime does
// not auto-fall-back to `npx @deepseek-ai/dsh` — the spike showed that times
// out on first install because of transitive deps — so install guidance lives
// in SuggestedInstallCommand instead.
func ResolveDSHBinary(ctx context.Context) (string, error) {
	return binaryutil.ResolveBinary(ctx, dshBinarySpec)
}

// dshBinary returns the cached resolved path, populating it on first use.
func (p *Plugin) dshBinary(ctx context.Context) (string, error) {
	p.binaryMu.Lock()
	defer p.binaryMu.Unlock()

	if p.resolvedBinary != "" {
		return p.resolvedBinary, nil
	}

	binary, err := ResolveDSHBinary(ctx)
	if err != nil {
		return "", err
	}
	p.resolvedBinary = binary
	return binary, nil
}
