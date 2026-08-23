package cursor

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCursorUsageProtocolAcceptsOnlyVerifiedBuild(t *testing.T) {
	if _, err := newUsageClient("cursor-agent", "2026.08.11-e8db854", "/runtime", nil); err != nil {
		t.Fatalf("verified build rejected: %v", err)
	}
	if _, err := newUsageClient("cursor-agent", "2026.08.12-unknown", "/runtime", nil); !errors.Is(err, errCursorUsageBuildUnsupported) {
		t.Fatalf("unknown build error = %v", err)
	}
}

func TestCursorUsageRunnerCapsOutputWhileReading(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture uses a POSIX shell")
	}
	dir := t.TempDir()
	for name, fixture := range map[string]struct {
		contents string
		mode     os.FileMode
	}{
		"cursor-agent": {contents: "fixture", mode: 0o600},
		"node":         {contents: "#!/bin/sh\nhead -c 70000 /dev/zero\n", mode: 0o700},
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(fixture.contents), fixture.mode); err != nil {
			t.Fatal(err)
		}
	}
	runtimeDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(runtimeDir, "ao-cursor-plan-usage.cjs"), []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := runCursorUsageHelper(context.Background(), filepath.Join(dir, "cursor-agent"), runtimeDir)
	if err == nil || !strings.Contains(err.Error(), "oversized") {
		t.Fatalf("error = %v, want bounded output failure", err)
	}
}

func TestCursorUsageProtocolDecodesSanitizedUsage(t *testing.T) {
	runner := func(context.Context, string, string) ([]byte, error) {
		return []byte(`{
			"cliVersion":"2026.08.11-e8db854",
			"usage":{"kind":"available","model":{"kind":"standard","planName":"Pro+","resetLabel":"Resets Aug 25","included":{"totalPercentUsed":21,"autoPercentUsed":10,"apiPercentUsed":100},"onDemand":{"kind":"fixed","usedDollars":333.68,"limitDollars":1}}}
		}`), nil
	}
	client, err := newUsageClient("cursor-agent", "2026.08.11-e8db854", "/runtime", runner)
	if err != nil {
		t.Fatal(err)
	}
	usage, err := client.ReadUsage(context.Background())
	if err != nil {
		t.Fatalf("ReadUsage: %v", err)
	}
	if usage.PlanName != "Pro+" || usage.ResetLabel != "Resets Aug 25" {
		t.Fatalf("metadata = %#v", usage)
	}
	if usage.Included == nil || usage.Included.TotalPercentUsed == nil || *usage.Included.TotalPercentUsed != 21 ||
		usage.OnDemand == nil || usage.OnDemand.UsedDollars == nil || *usage.OnDemand.UsedDollars != 333.68 ||
		usage.OnDemand.LimitDollars == nil || *usage.OnDemand.LimitDollars != 1 {
		t.Fatalf("usage = %#v", usage)
	}
}

func TestCursorUsageHelperEnvironmentExcludesCredentials(t *testing.T) {
	got := strings.Join(cursorUsageHelperEnv([]string{
		"HOME=/home/me", "PATH=/bin", "CURSOR_DATA_DIR=/profile", "CURSOR_API_KEY=secret", "ANTHROPIC_AUTH_TOKEN=secret", "OTHER=value",
	}), "\n")
	if !strings.Contains(got, "HOME=/home/me") || !strings.Contains(got, "CURSOR_DATA_DIR=/profile") {
		t.Fatalf("required environment missing: %q", got)
	}
	if strings.Contains(got, "secret") || strings.Contains(got, "OTHER=") {
		t.Fatalf("unsafe environment leaked: %q", got)
	}
}

func TestCursorUsageProtocolRejectsUnavailableAndOversizedOutput(t *testing.T) {
	for _, tc := range []struct {
		name   string
		output string
	}{
		{name: "unavailable", output: `{"cliVersion":"2026.08.11-e8db854","usage":{"kind":"unavailable","message":"not available"}}`},
		{name: "oversized", output: strings.Repeat("x", maxCursorUsageOutput+1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, err := newUsageClient("cursor-agent", "2026.08.11-e8db854", "/runtime", func(context.Context, string, string) ([]byte, error) {
				return []byte(tc.output), nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := client.ReadUsage(context.Background()); err == nil {
				t.Fatal("expected protocol error")
			}
		})
	}
}

func TestCursorUsageProtocolRejectsStructurallyIncompleteOutput(t *testing.T) {
	client, err := newUsageClient("cursor-agent", supportedCursorUsageBuild, "/runtime", func(context.Context, string, string) ([]byte, error) {
		return []byte(`{"cliVersion":"2026.08.11-e8db854","usage":{"kind":"available","model":{"kind":"standard","planName":"Pro+"}}}`), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.ReadUsage(context.Background()); err == nil {
		t.Fatal("expected incomplete usage model to be rejected")
	}
}
