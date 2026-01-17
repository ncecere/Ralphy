package prompt

import (
	"fmt"
	"strings"

	"github.com/ncecere/ralphy/internal/config"
	"github.com/ncecere/ralphy/internal/tasks"
)

type Options struct {
	Task      tasks.Task
	IssueBody string
	Config    config.Config
}

func Build(opts Options) string {
	cfg := opts.Config
	prompt := buildContext(opts)

	prompt = strings.TrimSuffix(prompt, "\n")
	prompt += "\n1. Find the highest-priority incomplete task and implement it."

	step := 2
	if !cfg.SkipTests {
		prompt += fmt.Sprintf("\n%d. Write tests for the feature.", step)
		prompt += fmt.Sprintf("\n%d. Run tests and ensure they pass before proceeding.", step+1)
		step += 2
	}

	if !cfg.SkipLint {
		prompt += fmt.Sprintf("\n%d. Run linting and ensure it passes before proceeding.", step)
		step++
	}

	switch cfg.PRDSource {
	case config.PRDSourceMarkdown:
		prompt += fmt.Sprintf("\n%d. Update the PRD to mark the task as complete (change '- [ ]' to '- [x]').", step)
	case config.PRDSourceYAML:
		prompt += fmt.Sprintf("\n%d. Update %s to mark the task as completed (set completed: true).", step, cfg.PRDFile)
	case config.PRDSourceGitHub:
		prompt += fmt.Sprintf("\n%d. The task will be marked complete automatically. Just note the completion in progress.txt.", step)
	}

	step++
	prompt += fmt.Sprintf("\n%d. Append your progress to progress.txt.", step)
	prompt += fmt.Sprintf("\n%d. Commit your changes with a descriptive message.", step+1)
	prompt += "\nONLY WORK ON A SINGLE TASK."

	if !cfg.SkipTests {
		prompt += " Do not proceed if tests fail."
	}
	if !cfg.SkipLint {
		prompt += " Do not proceed if linting fails."
	}

	prompt += "\nIf ALL tasks in the PRD are complete, output <promise>COMPLETE</promise>."

	return prompt
}

func buildContext(opts Options) string {
	cfg := opts.Config
	switch cfg.PRDSource {
	case config.PRDSourceGitHub:
		taskLabel := taskLabel(opts.Task)
		context := fmt.Sprintf("Task from GitHub Issue: %s\n\nIssue Description:\n%s\n\n@progress.txt", taskLabel, strings.TrimSpace(opts.IssueBody))
		return strings.TrimSuffix(context, "\n")
	default:
		return fmt.Sprintf("@%s @progress.txt", cfg.PRDFile)
	}
}

func taskLabel(task tasks.Task) string {
	if task.ID == "" {
		return task.Title
	}
	return fmt.Sprintf("%s:%s", task.ID, task.Title)
}
