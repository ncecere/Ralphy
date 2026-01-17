package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ncecere/ralphy/internal/config"
	"github.com/ncecere/ralphy/internal/engine"
	"github.com/ncecere/ralphy/internal/git"
	"github.com/ncecere/ralphy/internal/github"
	"github.com/ncecere/ralphy/internal/notify"
	"github.com/ncecere/ralphy/internal/parallel"
	"github.com/ncecere/ralphy/internal/prompt"
	"github.com/ncecere/ralphy/internal/tasks"
	"github.com/ncecere/ralphy/internal/ui"
	"github.com/ncecere/ralphy/internal/utils"
)

type totals struct {
	iterations   int
	inputTokens  int
	outputTokens int
	cost         float64
	branches     []string
}

func Run(ctx context.Context, cfg config.Config) error {
	if err := preflight(cfg); err != nil {
		return err
	}

	source, err := loadTaskSource(cfg)
	if err != nil {
		return err
	}

	if cfg.DryRun && cfg.MaxIterations == 0 {
		cfg.MaxIterations = 1
	}

	if cfg.BranchPerTask && cfg.BaseBranch == "" {
		cfg.BaseBranch, _ = git.CurrentBranch()
	}

	ui.Banner(cfg.AIEngine, cfg.PRDSource, prdLabel(cfg))

	if cfg.Parallel {
		return runParallel(ctx, cfg, source)
	}

	t := &totals{}

	for {
		t.iterations++
		resultCode, err := runSingleTask(ctx, cfg, source, t)
		if err != nil {
			ui.Error(err.Error())
			return err
		}
		switch resultCode {
		case taskResultAllDone:
			showSummary(t)
			notify.Done("Ralphy has completed all tasks!")
			return nil
		case taskResultFailed:
			ui.Warn("Task failed, continuing...")
		}

		if cfg.MaxIterations > 0 && t.iterations >= cfg.MaxIterations {
			ui.Warn("Reached max iterations")
			showSummary(t)
			notify.Done("Ralphy stopped after max iterations")
			return nil
		}

		utils.Sleep(1 * time.Second)
	}
}

func runParallel(ctx context.Context, cfg config.Config, source tasks.Source) error {
	totals, err := parallel.Run(ctx, cfg, source)
	if err != nil {
		return err
	}
	cost := totals.Cost
	if cost == 0 && totals.InputTokens > 0 {
		cost = float64(totals.InputTokens)*0.000003 + float64(totals.OutputTokens)*0.000015
	}
	ui.Summary(totals.Iterations, totals.InputTokens, totals.OutputTokens, cost)
	return nil
}

const (
	taskResultSuccess = iota
	taskResultFailed
	taskResultAllDone
)

func runSingleTask(ctx context.Context, cfg config.Config, source tasks.Source, t *totals) (int, error) {
	summary, err := source.Summary(ctx)
	if err != nil {
		return taskResultFailed, err
	}

	currentTask, err := source.Next(ctx)
	if err != nil {
		return taskResultFailed, err
	}
	if currentTask == nil {
		return taskResultAllDone, nil
	}

	issueBody := ""
	if cfg.PRDSource == config.PRDSourceGitHub {
		if ghSource, ok := source.(*tasks.GitHubSource); ok {
			issueBody, _ = ghSource.IssueBody(ctx, *currentTask)
		}
	}

	ui.TaskHeader(t.iterations, summary.Completed, summary.Remaining)

	var branchName string
	if cfg.BranchPerTask {
		branchName = git.TaskBranchName(currentTask.Title)
		ui.Info("Working on branch: " + branchName)
		stashed, _ := git.StashPush("ralphy-autostash")
		if err := git.CreateBranch(branchName, cfg.BaseBranch); err != nil {
			ui.Warn("Failed to create branch: " + err.Error())
		}
		if stashed {
			_ = git.StashPop()
		}
		t.branches = append(t.branches, branchName)
	}

	promptText := prompt.Build(prompt.Options{Task: *currentTask, IssueBody: issueBody, Config: cfg})

	if cfg.DryRun {
		ui.Info("DRY RUN - Would execute:")
		println(promptText)
		returnToBase(cfg)
		return taskResultSuccess, nil
	}

	outputFile, err := tempOutputFile()
	if err != nil {
		returnToBase(cfg)
		return taskResultFailed, err
	}
	defer os.Remove(outputFile)

	activity := engine.NewActivityTracker()
	spinner := ui.NewSpinner(truncate(currentTask.Title, 40))
	spinner.SetActivity(activity)
	spinner.Start()

	var result engine.RunResult
	var lastErr error
	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		result, err = engine.Run(ctx, cfg.AIEngine, promptText, engine.RunOptions{WorkDir: ".", OutputFile: outputFile, Model: cfg.Model, Activity: activity})
		if err != nil {
			lastErr = err
			spinner.SetStep("Retrying")
			if attempt < cfg.MaxRetries {
				utils.Sleep(time.Duration(cfg.RetryDelaySeconds) * time.Second)
				continue
			}
			spinner.Stop()
			ui.TaskFailed(currentTask.Title)
			returnToBase(cfg)
			return taskResultFailed, err
		}
		lastErr = nil
		break
	}

	spinner.Stop()

	if lastErr != nil {
		ui.TaskFailed(currentTask.Title)
		returnToBase(cfg)
		return taskResultFailed, lastErr
	}

	ui.TaskDone(currentTask.Title)

	if result.Response != "" {
		println()
		println(result.Response)
	}

	t.inputTokens += result.InputTokens
	t.outputTokens += result.OutputTokens
	if result.ActualCost > 0 {
		t.cost += result.ActualCost
	}

	if cfg.PRDSource == config.PRDSourceGitHub {
		_ = source.MarkComplete(ctx, *currentTask)
	}

	if cfg.CreatePR && branchName != "" {
		if err := git.Push(branchName); err == nil {
			prURL, err := github.CreatePR(ctx, github.PROptions{
				Title: currentTask.Title,
				Body:  "Automated PR created by Ralphy",
				Head:  branchName,
				Base:  cfg.BaseBranch,
				Draft: cfg.PRDraft,
				Token: cfg.GitHubToken,
			})
			if err == nil {
				ui.Success("PR created: " + prURL)
			} else {
				ui.Warn("Failed to create PR: " + err.Error())
			}
		}
	}

	returnToBase(cfg)

	remaining, err := source.Summary(ctx)
	if err != nil {
		return taskResultFailed, err
	}

	if remaining.Remaining == 0 {
		return taskResultAllDone, nil
	}

	rawOutput := result.RawOutput
	if strings.Contains(rawOutput, "<promise>COMPLETE</promise>") && remaining.Remaining > 0 {
		ui.Debug(cfg.Verbose, "AI claimed completion but tasks remain, continuing...")
	}

	return taskResultSuccess, nil
}

func returnToBase(cfg config.Config) {
	if cfg.BranchPerTask {
		_ = git.Checkout(cfg.BaseBranch)
	}
}

func loadTaskSource(cfg config.Config) (tasks.Source, error) {
	switch cfg.PRDSource {
	case config.PRDSourceYAML:
		return tasks.NewYAMLSource(cfg.PRDFile), nil
	case config.PRDSourceGitHub:
		return tasks.NewGitHubSource(cfg.GitHubRepo, cfg.GitHubLabel, cfg.GitHubToken)
	default:
		return tasks.NewMarkdownSource(cfg.PRDFile), nil
	}
}

func preflight(cfg config.Config) error {
	switch cfg.PRDSource {
	case config.PRDSourceMarkdown, config.PRDSourceYAML:
		if _, err := os.Stat(cfg.PRDFile); errors.Is(err, os.ErrNotExist) {
			return errors.New(cfg.PRDFile + " not found")
		}
	case config.PRDSourceGitHub:
		if cfg.GitHubRepo == "" {
			return errors.New("github repo required")
		}
	}
	return ensureProgressFile()
}

func ensureProgressFile() error {
	if _, err := os.Stat("progress.txt"); errors.Is(err, os.ErrNotExist) {
		return os.WriteFile("progress.txt", []byte(""), 0o644)
	}
	return nil
}

func tempOutputFile() (string, error) {
	file, err := os.CreateTemp("", "ralphy-output-*.log")
	if err != nil {
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	return filepath.Abs(file.Name())
}

func prdLabel(cfg config.Config) string {
	if cfg.PRDSource == config.PRDSourceGitHub {
		return cfg.GitHubRepo
	}
	return cfg.PRDFile
}

func showSummary(t *totals) {
	cost := t.cost
	if cost == 0 && t.inputTokens > 0 {
		cost = float64(t.inputTokens)*0.000003 + float64(t.outputTokens)*0.000015
	}
	ui.Summary(t.iterations, t.inputTokens, t.outputTokens, cost)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
