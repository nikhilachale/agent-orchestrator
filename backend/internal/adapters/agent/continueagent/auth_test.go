package continueagent

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/aoagents/agent-orchestrator/backend/internal/adapters/agent/authprobe"
	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
)

func TestContinueLocalAuthStatusAuthorizedFromEnv(t *testing.T) {
	t.Setenv("CONTINUE_API_KEY", "continue-key")

	status, ok, err := continueLocalAuthStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !ok || status != ports.AgentAuthStatusAuthorized {
		t.Fatalf("status = (%q, %v), want (%q, true)", status, ok, ports.AgentAuthStatusAuthorized)
	}
}

func TestContinueLocalAuthStatusUnknownWithoutEnv(t *testing.T) {
	t.Setenv("CONTINUE_API_KEY", "")
	t.Setenv("HOME", t.TempDir())

	status, ok, err := continueLocalAuthStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ok || status != ports.AgentAuthStatusUnknown {
		t.Fatalf("status = (%q, %v), want (%q, false)", status, ok, ports.AgentAuthStatusUnknown)
	}
}

func TestContinueConfigAuthStatusAuthorizedWithAPIKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("models:\n  - provider: anthropic\n    apiKey: continue-key\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	status, ok, err := continueConfigAuthStatus(path)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || status != ports.AgentAuthStatusAuthorized {
		t.Fatalf("status = (%q, %v), want (%q, true)", status, ok, ports.AgentAuthStatusAuthorized)
	}
}

func TestContinuePrintAuthStatusAuthorizedFromResponse(t *testing.T) {
	restore := stubContinueAuthCommand(t, []string{"-p", "hi"}, []byte("Hi! How can I help?\n"), nil)
	defer restore()

	status, err := continuePrintAuthStatus(context.Background(), "cn")
	if err != nil {
		t.Fatal(err)
	}
	if status != ports.AgentAuthStatusAuthorized {
		t.Fatalf("status = %q, want %q", status, ports.AgentAuthStatusAuthorized)
	}
}

func TestContinuePrintAuthStatusUnauthorizedFromLoginPrompt(t *testing.T) {
	restore := stubContinueAuthCommand(t, []string{"-p", "hi"}, []byte("Login required. Please run cn login.\n"), assertErr("exit status 1"))
	defer restore()

	status, err := continuePrintAuthStatus(context.Background(), "cn")
	if err != nil {
		t.Fatal(err)
	}
	if status != ports.AgentAuthStatusUnauthorized {
		t.Fatalf("status = %q, want %q", status, ports.AgentAuthStatusUnauthorized)
	}
}

type assertErr string

func (e assertErr) Error() string {
	return string(e)
}

func stubContinueAuthCommand(t *testing.T, wantArgs []string, out []byte, err error) func() {
	t.Helper()
	previous := authprobe.CmdRunner
	authprobe.CmdRunner = func(ctx context.Context, name string, arg ...string) ([]byte, error) {
		if name != "cn" || !reflect.DeepEqual(arg, wantArgs) {
			t.Fatalf("command = %s %#v, want cn %#v", name, arg, wantArgs)
		}
		return out, err
	}
	return func() { authprobe.CmdRunner = previous }
}
