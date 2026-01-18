package parallel

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ncecere/ralphy/internal/config"
	"github.com/ncecere/ralphy/internal/engine"
	"github.com/ncecere/ralphy/internal/git"
	"github.com/ncecere/ralphy/internal/github"
	"github.com/ncecere/ralphy/internal/prompt"
	"github.com/ncecere/ralphy/internal/tasks"
)

type AgentResult struct {
	Task         tasks.Task
	AgentNum     int
	Success      bool
	Branch       string
	InputTokens  int
	OutputTokens int
	Cost         float64
	Error        error
}

type agentWork struct {
	task     tasks.Task
	agentNum int
}

func RunAgents(ctx context.Context, cfg config.Config, allTasks []tasks.Task, worktreeBase, baseBranch string) []AgentResult {
	results := make([]AgentResult, 0, len(allTasks))

	for batchStart := 0; batchStart < len(allTasks); batchStart += cfg.MaxParallel {
		batchEnd := batchStart + cfg.MaxParallel
		if batchEnd > len(allTasks) {
			batchEnd = len(allTasks)
		}
		batch := allTasks[batchStart:batchEnd]

		batchResults := runBatch(ctx, cfg, batch, batchStart, worktreeBase, baseBranch)
		results = append(results, batchResults...)

		if cfg.MaxIterations > 0 && len(results) >= cfg.MaxIterations {
			break
		}
	}

	return results
}

func runBatch(ctx context.Context, cfg config.Config, batch []tasks.Task, startIdx int, worktreeBase, baseBranch string) []AgentResult {
	var wg sync.WaitGroup
	resultsCh := make(chan AgentResult, len(batch))

	for i, task := range batch {
		agentNum := startIdx + i + 1
		wg.Add(1)
		go func(t tasks.Task, num int) {
			defer wg.Done()
			result := runAgent(ctx, cfg, t, num, worktreeBase, baseBranch)
			resultsCh <- result
		}(task, agentNum)
	}

	wg.Wait()
	close(resultsCh)

	results := make([]AgentResult, 0, len(batch))
	for r := range resultsCh {
		results = append(results, r)
	}
	return results
}

func runAgent(ctx context.Context, cfg config.Config, task tasks.Task, agentNum int, worktreeBase, baseBranch string) AgentResult {
	result := AgentResult{Task: task, AgentNum: agentNum}

	branchName := git.AgentBranchName(agentNum, task.Title)
	worktreeDir := filepath.Join(worktreeBase, fmt.Sprintf("agent-%d", agentNum))

	if err := git.WorktreePrune(); err != nil {
		result.Error = err
		return result
	}

	_ = git.DeleteBranch(branchName)

	if _, err := git.RunInDir(".", "branch", branchName, baseBranch); err != nil {
		result.Error = fmt.Errorf("create branch: %w", err)
		return result
	}

	_ = os.RemoveAll(worktreeDir)
	if err := git.WorktreeAdd(worktreeDir, branchName); err != nil {
		result.Error = fmt.Errorf("create worktree: %w", err)
		return result
	}

	defer func() {
		if !git.HasUncommittedChanges() {
			_ = git.WorktreeRemove(worktreeDir)
		}
	}()

	if cfg.PRDSource == config.PRDSourceMarkdown || cfg.PRDSource == config.PRDSourceYAML {
		src := cfg.PRDFile
		dst := filepath.Join(worktreeDir, filepath.Base(cfg.PRDFile))
		data, err := os.ReadFile(src)
		if err == nil {
			_ = os.WriteFile(dst, data, 0o644)
		}
	}

	progressPath := filepath.Join(worktreeDir, "progress.txt")
	if _, err := os.Stat(progressPath); os.IsNotExist(err) {
		_ = os.WriteFile(progressPath, []byte(""), 0o644)
	}

	promptText := prompt.BuildParallel(task)

	outputFile, err := os.CreateTemp("", "ralphy-agent-*.log")
	if err != nil {
		result.Error = err
		return result
	}
	outputPath := outputFile.Name()
	outputFile.Close()
	defer os.Remove(outputPath)

	var lastErr error
	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		engineResult, err := engine.Run(ctx, cfg.AIEngine, promptText, engine.RunOptions{
			WorkDir:    worktreeDir,
			OutputFile: outputPath,
			Model:      cfg.ResolvedModel(),
		})
		if err != nil {
			lastErr = err
			continue
		}
		result.InputTokens = engineResult.InputTokens
		result.OutputTokens = engineResult.OutputTokens
		result.Cost = engineResult.ActualCost
		lastErr = nil
		break
	}

	if lastErr != nil {
		result.Error = lastErr
		return result
	}

	commitCount, _ := git.CommitCount(baseBranch)
	if commitCount == 0 {
		result.Error = fmt.Errorf("no commits created")
		return result
	}

	result.Branch = branchName
	result.Success = true

	if cfg.CreatePR {
		if err := git.Push(branchName); err == nil {
			_, _ = github.CreatePR(ctx, github.PROptions{
				Title: task.Title,
				Body:  fmt.Sprintf("Automated implementation by Ralphy (Agent %d)", agentNum),
				Head:  branchName,
				Base:  baseBranch,
				Draft: cfg.PRDraft,
				Token: cfg.GitHubToken,
			})
		}
	}

	return result
}
