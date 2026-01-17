package parallel

import (
	"context"
	"os"

	"github.com/ncecere/ralphy/internal/config"
	"github.com/ncecere/ralphy/internal/git"
	"github.com/ncecere/ralphy/internal/notify"
	"github.com/ncecere/ralphy/internal/tasks"
	"github.com/ncecere/ralphy/internal/ui"
)

type Totals struct {
	Iterations   int
	InputTokens  int
	OutputTokens int
	Cost         float64
	Branches     []string
}

func Run(ctx context.Context, cfg config.Config, source tasks.Source) (*Totals, error) {
	allTasks, err := source.All(ctx)
	if err != nil {
		return nil, err
	}

	if len(allTasks) == 0 {
		ui.Info("No tasks to run")
		return &Totals{}, nil
	}

	ui.Info("Found " + itoa(len(allTasks)) + " tasks to process")

	baseBranch := cfg.BaseBranch
	if baseBranch == "" {
		baseBranch, _ = git.CurrentBranch()
	}

	worktreeBase, err := os.MkdirTemp("", "ralphy-worktrees-*")
	if err != nil {
		return nil, err
	}
	defer cleanupWorktrees(worktreeBase)

	ui.Info("Base branch: " + baseBranch)

	results := RunAgents(ctx, cfg, allTasks, worktreeBase, baseBranch)

	totals := &Totals{}
	var completedBranches []string

	for _, r := range results {
		totals.Iterations++
		totals.InputTokens += r.InputTokens
		totals.OutputTokens += r.OutputTokens
		totals.Cost += r.Cost

		if r.Success {
			completedBranches = append(completedBranches, r.Branch)
			totals.Branches = append(totals.Branches, r.Branch)
			_ = source.MarkComplete(ctx, r.Task)
			ui.Success("Agent " + itoa(r.AgentNum) + ": " + r.Task.Title)
		} else {
			errMsg := "unknown error"
			if r.Error != nil {
				errMsg = r.Error.Error()
			}
			ui.Error("Agent " + itoa(r.AgentNum) + ": " + r.Task.Title + " - " + errMsg)
		}
	}

	if !cfg.CreatePR && len(completedBranches) > 0 {
		ui.Info("Merging branches into " + baseBranch + "...")
		MergeBranches(ctx, cfg, completedBranches, baseBranch)
	}

	notify.Done("Ralphy completed parallel execution!")

	return totals, nil
}

func cleanupWorktrees(base string) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			path := base + "/" + e.Name()
			_ = git.WorktreeRemove(path)
		}
	}
	_ = os.RemoveAll(base)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
