package cursor

import (
	"context"
	"errors"
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
	if usage.Included.TotalPercentUsed != 21 || usage.OnDemand.UsedDollars != 333.68 || usage.OnDemand.LimitDollars != 1 {
		t.Fatalf("usage = %#v", usage)
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
