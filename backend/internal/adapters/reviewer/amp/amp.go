// Package amp adapts Amp execute mode for AO code-review sessions.
package amp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	workeramp "github.com/aoagents/agent-orchestrator/backend/internal/adapters/agent/amp"
	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
)

// Reviewer is the Amp code-review adapter.
type Reviewer struct{}

func New() *Reviewer { return &Reviewer{} }

func (*Reviewer) Harness() domain.ReviewerHarness { return domain.ReviewerAmp }

var _ ports.Reviewer = (*Reviewer)(nil)
var _ ports.ReviewerCanceller = (*Reviewer)(nil)
var _ ports.ReviewerReusePolicy = (*Reviewer)(nil)

// ReviewCommand launches one private execute-mode Amp turn. A reviewer-only
// settings file keeps the checkout read-only while allowing the narrow git,
// GitHub, and AO bookkeeping commands required by the central review prompt.
func (*Reviewer) ReviewCommand(ctx context.Context, inv ports.ReviewInvocation) (ports.ReviewCommandSpec, error) {
	binary, err := workeramp.ResolveAmpBinary(ctx)
	if err != nil {
		return ports.ReviewCommandSpec{}, err
	}

	prompt := inv.Prompt
	if inv.SystemPromptFile != "" {
		prompt = fmt.Sprintf("First read and follow the reviewer role in `%s`. Then %s", filepath.ToSlash(inv.SystemPromptFile), inv.Prompt)
	} else if strings.TrimSpace(inv.SystemPrompt) != "" {
		prompt = strings.TrimSpace(inv.SystemPrompt + "\n\n" + inv.Prompt)
	}

	argv := []string{
		binary,
		"--execute", prompt,
		"--visibility", "private",
		"--no-ide",
		"--no-notifications",
		"--plugin-ready-timeout", "30",
	}
	if inv.TaskPromptRoot != "" {
		settingsPath, err := writeReviewerSettings(inv.TaskPromptRoot)
		if err != nil {
			return ports.ReviewCommandSpec{}, err
		}
		argv = append(argv, "--settings-file", settingsPath)
	}
	return ports.ReviewCommandSpec{Argv: argv}, nil
}

func writeReviewerSettings(root string) (string, error) {
	if err := os.MkdirAll(root, 0o700); err != nil {
		return "", fmt.Errorf("create amp reviewer settings directory: %w", err)
	}
	settings, err := loadUserSettings()
	if err != nil {
		return "", err
	}
	settings["amp.dangerouslyAllowAll"] = false
	settings["amp.remoteThreadCreation.enabled"] = false
	settings["amp.updates.mode"] = "disabled"
	settings["amp.mcpServers"] = map[string]any{}
	settings["amp.permissions"] = []map[string]any{
		{"tool": "Read", "action": "allow"},
		{"tool": "Grep", "action": "allow"},
		{"tool": "Glob", "action": "allow"},
		{
			"tool": "Bash",
			"matches": map[string]any{"cmd": []string{
				"git diff*",
				"git log*",
				"git show*",
				"git status*",
				"gh api*",
				"ao review submit*",
				"printf * | gh api*",
				"printf * | ao review submit*",
			}},
			"action": "allow",
		},
		{"tool": "Bash", "action": "reject"},
		{"tool": "Edit", "action": "reject"},
		{"tool": "Write", "action": "reject"},
		{"tool": "*", "action": "reject"},
	}
	data, err := json.Marshal(settings)
	if err != nil {
		return "", fmt.Errorf("encode amp reviewer settings: %w", err)
	}
	path := filepath.Join(root, "amp-settings.json")
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return "", fmt.Errorf("write amp reviewer settings: %w", err)
	}
	return path, nil
}

func loadUserSettings() (map[string]any, error) {
	path := strings.TrimSpace(os.Getenv("AMP_SETTINGS_FILE"))
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return map[string]any{}, nil
		}
		path = filepath.Join(home, ".config", "amp", "settings.json")
	}
	data, err := os.ReadFile(path) //nolint:gosec // user-selected Amp settings path
	if os.IsNotExist(err) {
		return map[string]any{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read amp settings: %w", err)
	}
	settings := make(map[string]any)
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("decode amp settings: %w", err)
	}
	return settings, nil
}

func (*Reviewer) ReviewMessage(_ context.Context, inv ports.ReviewInvocation) (string, error) {
	return inv.Prompt, nil
}

func (*Reviewer) ReviewCancel(context.Context) (ports.ReviewCancelSpec, error) {
	return ports.ReviewCancelSpec{Mode: ports.ReviewCancelInterrupt, Interrupts: 2}, nil
}

func (*Reviewer) ReviewProcessReusable() bool { return false }
