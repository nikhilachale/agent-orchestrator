// Package aider adapts Aider's one-shot mode for AO code-review sessions.
package aider

import (
	"context"
	"strings"

	workeraider "github.com/aoagents/agent-orchestrator/backend/internal/adapters/agent/aider"
	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
)

// Reviewer is the Aider code-review adapter.
type Reviewer struct{}

func New() *Reviewer { return &Reviewer{} }

func (*Reviewer) Harness() domain.ReviewerHarness { return domain.ReviewerAider }

var _ ports.Reviewer = (*Reviewer)(nil)
var _ ports.ReviewerCanceller = (*Reviewer)(nil)
var _ ports.ReviewerReusePolicy = (*Reviewer)(nil)

// ReviewCommand runs one Aider turn over AO's immutable task file. Dry-run and
// ask mode prevent Aider edits; --yes-always lets its suggested read-only git
// and reporting commands run without an unattended confirmation prompt. Aider
// has no command allowlist, so the central reviewer prompt remains the boundary
// around the shell commands it is asked to execute.
func (*Reviewer) ReviewCommand(ctx context.Context, inv ports.ReviewInvocation) (ports.ReviewCommandSpec, error) {
	binary, err := workeraider.ResolveAiderBinary(ctx)
	if err != nil {
		return ports.ReviewCommandSpec{}, err
	}
	argv := []string{
		binary,
		"--chat-mode", "ask",
		"--dry-run",
		"--yes-always",
		"--no-auto-commits",
		"--no-dirty-commits",
		"--no-gitignore",
		"--no-check-update",
		"--no-stream",
		"--no-pretty",
	}
	if inv.SystemPromptFile != "" {
		argv = append(argv, "--read", inv.SystemPromptFile)
	}
	if inv.TaskPromptFile != "" {
		argv = append(argv, "--message-file", inv.TaskPromptFile)
	} else if strings.TrimSpace(inv.Prompt) != "" {
		argv = append(argv, "--message", strings.TrimSpace(inv.SystemPrompt+"\n\n"+inv.Prompt))
	}
	return ports.ReviewCommandSpec{Argv: argv}, nil
}

func (*Reviewer) ReviewMessage(_ context.Context, inv ports.ReviewInvocation) (string, error) {
	return inv.Prompt, nil
}

func (*Reviewer) ReviewCancel(context.Context) (ports.ReviewCancelSpec, error) {
	return ports.ReviewCancelSpec{Mode: ports.ReviewCancelInterrupt, Interrupts: 2}, nil
}

func (*Reviewer) ReviewProcessReusable() bool { return false }
