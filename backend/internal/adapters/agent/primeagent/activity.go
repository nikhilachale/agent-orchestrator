package primeagent

import "github.com/aoagents/agent-orchestrator/backend/internal/domain"

// DeriveActivityState maps the normalized events emitted by the AO-managed
// Prime Agent extension to durable AO activity state.
func DeriveActivityState(event string, _ []byte) (domain.ActivityState, bool) {
	switch event {
	case "session-start", "user-prompt-submit":
		return domain.ActivityActive, true
	case "stop":
		return domain.ActivityIdle, true
	case "session-end":
		return domain.ActivityExited, true
	default:
		return "", false
	}
}
