package cursor

import (
	"context"
	"testing"
	"time"

	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
)

type fakeCursorQuotaPlugin struct {
	binary string
	auth   ports.AgentAuthStatus
}

func (f fakeCursorQuotaPlugin) ResolveBinary(context.Context) (string, error) { return f.binary, nil }
func (f fakeCursorQuotaPlugin) AuthStatus(context.Context) (ports.AgentAuthStatus, error) {
	return f.auth, nil
}

func TestNormalizeCursorUsageIncludesCreditsAndOnDemandSpend(t *testing.T) {
	observedAt := time.Date(2026, 8, 24, 4, 1, 34, 0, time.UTC)
	snapshot, err := normalizeCursorUsage(RawUsage{
		PlanName: "Pro+", ResetLabel: "Resets Aug 25",
		Included: RawIncludedUsage{TotalPercentUsed: 21, AutoPercentUsed: 10, APIPercentUsed: 100},
		OnDemand: RawOnDemandUsage{Kind: "fixed", UsedDollars: 333.68, LimitDollars: 1},
	}, observedAt)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Provider != "cursor" || snapshot.AccountID != "default" || snapshot.PlanType != "Pro+" {
		t.Fatalf("identity = %+v", snapshot)
	}
	if !snapshot.Capabilities.SupportsRead || !snapshot.Capabilities.SupportsCredits || !snapshot.Capabilities.SupportsSpendLimits {
		t.Fatalf("capabilities = %+v", snapshot.Capabilities)
	}
	if len(snapshot.Limits) != 4 {
		t.Fatalf("limits = %d, want 4", len(snapshot.Limits))
	}
	assertCursorFloat(t, snapshot.Limits[0].UsedPercent, 21)
	assertCursorFloat(t, snapshot.Limits[1].UsedPercent, 10)
	assertCursorFloat(t, snapshot.Limits[2].UsedPercent, 100)
	spend := snapshot.Limits[3]
	assertCursorFloat(t, spend.UsedValue, 333.68)
	assertCursorFloat(t, spend.TotalValue, 1)
	assertCursorFloat(t, spend.RemainingValue, 0)
	if spend.State != domain.QuotaLimitActive || spend.Unit != "USD" || spend.Reached == nil || !*spend.Reached {
		t.Fatalf("spend = %+v", spend)
	}
	if snapshot.Limits[0].ResetsAt == nil || !snapshot.Limits[0].ResetsAt.Equal(time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("reset = %v", snapshot.Limits[0].ResetsAt)
	}
}

func TestNormalizeCursorUsagePreservesNonNumericOnDemandStates(t *testing.T) {
	for _, kind := range []string{"unlimited", "disabled", "unavailable"} {
		t.Run(kind, func(t *testing.T) {
			snapshot, err := normalizeCursorUsage(RawUsage{OnDemand: RawOnDemandUsage{Kind: kind}}, time.Now().UTC())
			if err != nil {
				t.Fatal(err)
			}
			if got := snapshot.Limits[3].State; string(got) != kind {
				t.Fatalf("state = %q, want %q", got, kind)
			}
			if snapshot.Limits[3].TotalValue != nil {
				t.Fatalf("total = %v, want unknown", snapshot.Limits[3].TotalValue)
			}
		})
	}
}

func TestCursorQuotaAccountRequiresAuthorizedSupportedBuild(t *testing.T) {
	r := NewQuotaRefresher(fakeCursorQuotaPlugin{binary: "/cursor-agent", auth: ports.AgentAuthStatusAuthorized})
	r.readVersion = func(context.Context, string) (string, error) { return supportedCursorUsageBuild, nil }
	present, err := r.QuotaAccountPresent(context.Background(), "cursor", "default")
	if err != nil || !present {
		t.Fatalf("present = %v, err = %v", present, err)
	}
	r.readVersion = func(context.Context, string) (string, error) { return "new-build", nil }
	present, err = r.QuotaAccountPresent(context.Background(), "cursor", "default")
	if err != nil || present {
		t.Fatalf("unsupported present = %v, err = %v", present, err)
	}
}

func assertCursorFloat(t *testing.T, got *float64, want float64) {
	t.Helper()
	if got == nil || *got != want {
		t.Fatalf("value = %v, want %v", got, want)
	}
}
