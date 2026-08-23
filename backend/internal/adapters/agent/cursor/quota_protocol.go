package cursor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	aoprocess "github.com/aoagents/agent-orchestrator/backend/internal/process"
)

const (
	supportedCursorUsageBuild = "2026.08.11-e8db854"
	maxCursorUsageOutput      = 64 << 10
)

var errCursorUsageBuildUnsupported = errors.New("Cursor build does not expose a verified usage protocol")

// UsageClient reads Cursor's sanitized account usage model.
type UsageClient interface {
	ReadUsage(context.Context) (RawUsage, error)
}

// RawUsage is the credential-free usage model returned by Cursor's dashboard client.
type RawUsage struct {
	PlanName   string
	ResetLabel string
	Included   RawIncludedUsage
	OnDemand   RawOnDemandUsage
}

// RawIncludedUsage contains Cursor's provider-reported included percentages.
type RawIncludedUsage struct {
	TotalPercentUsed float64 `json:"totalPercentUsed"`
	AutoPercentUsed  float64 `json:"autoPercentUsed"`
	APIPercentUsed   float64 `json:"apiPercentUsed"`
}

// RawOnDemandUsage contains Cursor's provider-reported spend state.
type RawOnDemandUsage struct {
	Kind         string  `json:"kind"`
	UsedDollars  float64 `json:"usedDollars"`
	LimitDollars float64 `json:"limitDollars"`
	Scope        string  `json:"scope"`
}

type usageCommandRunner func(context.Context, string, string) ([]byte, error)

type privateUsageClient struct {
	binaryPath string
	version    string
	runtimeDir string
	run        usageCommandRunner
}

func newUsageClient(binaryPath, version, runtimeDir string, runner usageCommandRunner) (UsageClient, error) {
	version = strings.TrimSpace(version)
	if version != supportedCursorUsageBuild {
		return nil, fmt.Errorf("%w: %s", errCursorUsageBuildUnsupported, version)
	}
	if runner == nil {
		runner = runCursorUsageHelper
	}
	return &privateUsageClient{binaryPath: binaryPath, version: version, runtimeDir: runtimeDir, run: runner}, nil
}

func (c *privateUsageClient) ReadUsage(ctx context.Context) (RawUsage, error) {
	out, err := c.run(ctx, c.binaryPath, c.runtimeDir)
	if err != nil {
		return RawUsage{}, fmt.Errorf("Cursor usage helper failed: %w", err)
	}
	if len(out) > maxCursorUsageOutput {
		return RawUsage{}, errors.New("Cursor usage helper returned oversized output")
	}
	var envelope struct {
		CLIVersion string `json:"cliVersion"`
		Usage      struct {
			Kind    string `json:"kind"`
			Message string `json:"message"`
			Model   struct {
				Kind       string           `json:"kind"`
				PlanName   string           `json:"planName"`
				ResetLabel string           `json:"resetLabel"`
				Included   RawIncludedUsage `json:"included"`
				OnDemand   RawOnDemandUsage `json:"onDemand"`
			} `json:"model"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(out, &envelope); err != nil {
		return RawUsage{}, fmt.Errorf("decode Cursor usage helper output: %w", err)
	}
	if envelope.CLIVersion != c.version {
		return RawUsage{}, errors.New("Cursor usage helper build identity changed")
	}
	if envelope.Usage.Kind != "available" || envelope.Usage.Model.Kind != "standard" {
		return RawUsage{}, errors.New("Cursor usage is unavailable for this account")
	}
	return RawUsage{
		PlanName:   envelope.Usage.Model.PlanName,
		ResetLabel: envelope.Usage.Model.ResetLabel,
		Included:   envelope.Usage.Model.Included,
		OnDemand:   envelope.Usage.Model.OnDemand,
	}, nil
}

func runCursorUsageHelper(ctx context.Context, binaryPath, runtimeDir string) ([]byte, error) {
	resolved, err := filepath.EvalSymlinks(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("resolve Cursor executable: %w", err)
	}
	cursorDir := filepath.Dir(resolved)
	node := filepath.Join(cursorDir, "node")
	if runtime.GOOS == "windows" {
		node = filepath.Join(cursorDir, "node.exe")
	}
	helper := filepath.Join(runtimeDir, "ao-cursor-plan-usage.cjs")
	for path, label := range map[string]string{node: "Cursor Node runtime", helper: "AO Cursor usage helper"} {
		info, statErr := os.Stat(path)
		if statErr != nil || info.IsDir() {
			return nil, fmt.Errorf("%s is unavailable", label)
		}
	}
	cmd := aoprocess.CommandContext(ctx, node, helper, cursorDir) //nolint:gosec // fixed AO helper and resolved Cursor runtime.
	out, err := cmd.Output()
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, ctx.Err()
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("helper exited with status %d", exitErr.ExitCode())
		}
		return nil, errors.New("could not execute helper")
	}
	return out, nil
}
