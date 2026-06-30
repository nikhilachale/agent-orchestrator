package agy

import (
	"context"

	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
)

var _ ports.AgentAuthChecker = (*Plugin)(nil)

// AuthStatus returns the plugin's local authentication status.
func (p *Plugin) AuthStatus(ctx context.Context) (ports.AgentAuthStatus, error) {
	cmd, err := p.GetLaunchCommand(ctx, ports.LaunchConfig{})
	if err != nil || len(cmd) == 0 {
		return ports.AgentAuthStatusUnknown, err
	}
	return ports.AgentAuthStatusAuthorized, nil
}
