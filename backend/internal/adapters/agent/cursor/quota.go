package cursor

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
)

const cursorUsageTimeout = 20 * time.Second

// QuotaPlugin is the installed Cursor capability required by the quota reader.
type QuotaPlugin interface {
	ResolveBinary(context.Context) (string, error)
	AuthStatus(context.Context) (ports.AgentAuthStatus, error)
}

type cursorVersionReader func(context.Context, string) (string, error)
type cursorUsageClientFactory func(string, string, string, usageCommandRunner) (UsageClient, error)

// QuotaRefresher reads Cursor's local authenticated dashboard usage model.
type QuotaRefresher struct {
	plugin        QuotaPlugin
	readVersion   cursorVersionReader
	newClient     cursorUsageClientFactory
	runtimeDir    string
	now           func() time.Time
	commandRunner usageCommandRunner
}

// NewQuotaRefresher creates a daemon-owned Cursor plan-usage reader.
func NewQuotaRefresher(plugin QuotaPlugin) *QuotaRefresher {
	return &QuotaRefresher{
		plugin:      plugin,
		readVersion: readCursorVersion,
		newClient:   newUsageClient,
		runtimeDir:  cursorUsageRuntimeDirectory(),
		now:         func() time.Time { return time.Now().UTC() },
	}
}

// QuotaAccountPresent reports whether the installed, authenticated Cursor
// build has the exact usage protocol AO has verified.
func (r *QuotaRefresher) QuotaAccountPresent(ctx context.Context, provider domain.QuotaProviderID, accountID domain.QuotaAccountID) (bool, error) {
	if r == nil || r.plugin == nil || provider != "cursor" || accountID != "default" {
		return false, nil
	}
	binary, err := r.plugin.ResolveBinary(ctx)
	if err != nil {
		if errors.Is(err, ports.ErrAgentBinaryNotFound) {
			return false, nil
		}
		return false, err
	}
	auth, err := r.plugin.AuthStatus(ctx)
	if err != nil {
		return false, err
	}
	if auth != ports.AgentAuthStatusAuthorized {
		return false, nil
	}
	version, err := r.readVersion(ctx, binary)
	if err != nil {
		return false, err
	}
	return normalizeCursorBuild(version) == supportedCursorUsageBuild, nil
}

// RefreshQuota reads and normalizes Cursor plan usage without opening a session.
func (r *QuotaRefresher) RefreshQuota(ctx context.Context, provider domain.QuotaProviderID, accountID domain.QuotaAccountID) (domain.QuotaSnapshot, error) {
	present, err := r.QuotaAccountPresent(ctx, provider, accountID)
	if err != nil {
		return domain.QuotaSnapshot{}, err
	}
	if !present {
		return domain.QuotaSnapshot{}, ports.ErrQuotaRefreshUnsupported
	}
	binary, err := r.plugin.ResolveBinary(ctx)
	if err != nil {
		return domain.QuotaSnapshot{}, err
	}
	version, err := r.readVersion(ctx, binary)
	if err != nil {
		return domain.QuotaSnapshot{}, err
	}
	client, err := r.newClient(binary, normalizeCursorBuild(version), r.runtimeDir, r.commandRunner)
	if err != nil {
		return domain.QuotaSnapshot{}, err
	}
	readCtx, cancel := context.WithTimeout(ctx, cursorUsageTimeout)
	defer cancel()
	raw, err := client.ReadUsage(readCtx)
	if err != nil {
		return domain.QuotaSnapshot{}, err
	}
	return normalizeCursorUsage(raw, r.now())
}

func normalizeCursorUsage(raw RawUsage, observedAt time.Time) (domain.QuotaSnapshot, error) {
	if observedAt.IsZero() {
		return domain.QuotaSnapshot{}, errors.New("Cursor usage observation time is required")
	}
	reset := parseCursorResetLabel(raw.ResetLabel, observedAt)
	limits := []domain.QuotaLimit{
		cursorPercentLimit("included", "Included", raw.Included.TotalPercentUsed, reset, observedAt),
		cursorPercentLimit("auto", "Auto", raw.Included.AutoPercentUsed, reset, observedAt),
		cursorPercentLimit("api", "API", raw.Included.APIPercentUsed, reset, observedAt),
	}
	spend := domain.QuotaLimit{
		ID: "on_demand", Name: "On-Demand", Category: domain.QuotaSpendLimit,
		Scope: domain.QuotaAccountScope, Unit: "USD", ObservedAt: observedAt,
	}
	switch raw.OnDemand.Kind {
	case "fixed":
		spend.State = domain.QuotaLimitActive
		spend.UsedValue = float64Pointer(raw.OnDemand.UsedDollars)
		spend.TotalValue = float64Pointer(raw.OnDemand.LimitDollars)
		spend.RemainingValue = float64Pointer(math.Max(0, raw.OnDemand.LimitDollars-raw.OnDemand.UsedDollars))
		reached := raw.OnDemand.UsedDollars >= raw.OnDemand.LimitDollars
		spend.Reached = &reached
	case "unlimited":
		spend.State = domain.QuotaLimitUnlimited
	case "disabled":
		spend.State = domain.QuotaLimitDisabled
	case "unavailable", "":
		spend.State = domain.QuotaLimitUnavailable
	default:
		return domain.QuotaSnapshot{}, fmt.Errorf("unsupported Cursor on-demand usage state %q", raw.OnDemand.Kind)
	}
	limits = append(limits, spend)
	return domain.NormalizeQuotaSnapshot(domain.QuotaSnapshot{
		Provider: "cursor", AccountID: "default", AccountLabel: "Cursor", PlanType: raw.PlanName,
		Capabilities: domain.QuotaCapabilities{SupportsRead: true, SupportsHistory: true, SupportsCredits: true, SupportsSpendLimits: true},
		Limits:       limits, ObservedAt: observedAt, Completeness: domain.QuotaComplete,
	}), nil
}

func cursorPercentLimit(id domain.QuotaLimitID, name string, used float64, reset *time.Time, observedAt time.Time) domain.QuotaLimit {
	return domain.QuotaLimit{
		ID: id, Name: name, Category: domain.QuotaUsageCredits, Scope: domain.QuotaAccountScope,
		WindowType: "billing_cycle", UsedPercent: float64Pointer(used), ResetsAt: reset, ObservedAt: observedAt,
	}
}

func parseCursorResetLabel(label string, observedAt time.Time) *time.Time {
	value := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(label), "Resets "))
	parsed, err := time.ParseInLocation("Jan 2 2006", value+" "+fmt.Sprint(observedAt.Year()), time.UTC)
	if err != nil {
		return nil
	}
	day := time.Date(observedAt.Year(), observedAt.Month(), observedAt.Day(), 0, 0, 0, 0, time.UTC)
	if parsed.Before(day) {
		parsed = parsed.AddDate(1, 0, 0)
	}
	return &parsed
}

func float64Pointer(value float64) *float64 { return &value }

func readCursorVersion(ctx context.Context, binary string) (string, error) {
	out, err := exec.CommandContext(ctx, binary, "--version").Output() //nolint:gosec // resolved Cursor adapter binary.
	if err != nil {
		return "", errors.New("could not read Cursor version")
	}
	return strings.TrimSpace(string(out)), nil
}

func normalizeCursorBuild(version string) string {
	for _, field := range strings.Fields(version) {
		if field == supportedCursorUsageBuild {
			return field
		}
	}
	return strings.TrimSpace(version)
}

func cursorUsageRuntimeDirectory() string {
	if configured := strings.TrimSpace(os.Getenv("AO_ACP_RUNTIME_DIR")); configured != "" {
		return configured
	}
	executable, err := os.Executable()
	if err != nil {
		return ""
	}
	resources := filepath.Dir(filepath.Dir(executable))
	for _, candidate := range []string{filepath.Join(resources, "acp-runtime"), filepath.Join(resources, "resources", "acp-runtime")} {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return ""
}
