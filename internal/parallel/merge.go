package parallel

import (
	"context"
	"os"

	"github.com/ncecere/ralphy/internal/config"
	"github.com/ncecere/ralphy/internal/engine"
	"github.com/ncecere/ralphy/internal/git"
	"github.com/ncecere/ralphy/internal/ui"
)

func MergeBranches(ctx context.Context, cfg config.Config, branches []string, baseBranch string) {
	if len(branches) == 0 {
		return
	}

	if err := git.Checkout(baseBranch); err != nil {
		ui.Warn("Could not checkout " + baseBranch)
		ui.Info("Branches created:")
		for _, b := range branches {
			ui.Info("  " + b)
		}
		return
	}

	var failed []string

	for _, branch := range branches {
		ui.Info("Merging " + branch + "...")
		if err := git.Merge(branch); err != nil {
			ui.Warn("Conflict in " + branch)
			failed = append(failed, branch)
		} else {
			_ = git.DeleteBranch(branch)
		}
	}

	if len(failed) == 0 {
		ui.Success("All branches merged successfully!")
		return
	}

	ui.Info("Using AI to resolve merge conflicts...")

	var stillFailed []string
	for _, branch := range failed {
		conflicted, _ := git.ConflictedFiles()
		if len(conflicted) == 0 {
			_ = git.MergeAbort()
			if err := git.Merge(branch); err == nil {
				_ = git.DeleteBranch(branch)
				continue
			}
			stillFailed = append(stillFailed, branch)
			_ = git.MergeAbort()
			continue
		}

		resolved := resolveConflicts(ctx, cfg, conflicted)
		if resolved {
			ui.Success("AI resolved " + branch)
			_ = git.DeleteBranch(branch)
		} else {
			ui.Error("Could not resolve " + branch)
			stillFailed = append(stillFailed, branch)
			_ = git.MergeAbort()
		}
	}

	if len(stillFailed) > 0 {
		ui.Warn("Some conflicts need manual resolution:")
		for _, b := range stillFailed {
			ui.Info("  " + b)
		}
	}
}

func resolveConflicts(ctx context.Context, cfg config.Config, files []string) bool {
	prompt := buildConflictPrompt(files)

	outputFile, err := tempFile()
	if err != nil {
		return false
	}

	_, err = engine.Run(ctx, cfg.AIEngine, prompt, engine.RunOptions{
		WorkDir:    ".",
		OutputFile: outputFile,
		Model:      cfg.Model,
	})
	if err != nil {
		return false
	}

	remaining, _ := git.ConflictedFiles()
	return len(remaining) == 0
}

func buildConflictPrompt(files []string) string {
	prompt := `You are resolving a git merge conflict. The following files have conflicts:

`
	for _, f := range files {
		prompt += f + "\n"
	}

	prompt += `
For each conflicted file:
1. Read the file to see the conflict markers (<<<<<<< HEAD, =======, >>>>>>> branch)
2. Understand what both versions are trying to do
3. Edit the file to resolve the conflict by combining both changes intelligently
4. Remove all conflict markers
5. Make sure the resulting code is valid and compiles

After resolving all conflicts:
1. Run 'git add' on each resolved file
2. Run 'git commit --no-edit' to complete the merge

Be careful to preserve functionality from BOTH branches. The goal is to integrate all features.`

	return prompt
}

func tempFile() (string, error) {
	f, err := os.CreateTemp("", "ralphy-merge-*.log")
	if err != nil {
		return "", err
	}
	name := f.Name()
	f.Close()
	return name, nil
}
