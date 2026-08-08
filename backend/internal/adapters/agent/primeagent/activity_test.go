package primeagent

import (
	"testing"

	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
)

func TestDeriveActivityState(t *testing.T) {
	tests := []struct {
		event string
		state domain.ActivityState
		ok    bool
	}{
		{"session-start", domain.ActivityActive, true},
		{"user-prompt-submit", domain.ActivityActive, true},
		{"stop", domain.ActivityIdle, true},
		{"session-end", domain.ActivityExited, true},
		{"unknown", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.event, func(t *testing.T) {
			state, ok := DeriveActivityState(tt.event, nil)
			if state != tt.state || ok != tt.ok {
				t.Fatalf("DeriveActivityState(%q) = (%q, %v), want (%q, %v)", tt.event, state, ok, tt.state, tt.ok)
			}
		})
	}
}
