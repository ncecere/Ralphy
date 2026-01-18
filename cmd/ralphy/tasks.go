package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ncecere/ralphy/internal/config"
	"github.com/ncecere/ralphy/internal/tasks"
	"github.com/spf13/cobra"
)

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "List and preview tasks",
	Long: `List and preview tasks from the configured task source.

Subcommands:
  list    List all tasks with their status
  next    Show the next task that would be executed`,
}

var tasksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Long: `List all tasks from the configured task source.

Shows task status (pending/completed), title, and parallel group (if applicable).`,
	RunE: runTasksList,
}

var tasksNextCmd = &cobra.Command{
	Use:   "next",
	Short: "Show the next task",
	Long:  `Show the next task that would be executed without actually running it.`,
	RunE:  runTasksNext,
}

func init() {
	// Add source flags to tasks command (inherited by subcommands)
	tasksCmd.PersistentFlags().String("prd", "", "PRD file path")
	tasksCmd.PersistentFlags().String("yaml", "", "Use YAML task file")
	tasksCmd.PersistentFlags().String("github", "", "Fetch tasks from GitHub issues (owner/repo)")
	tasksCmd.PersistentFlags().String("github-label", "", "Filter GitHub issues by label")

	// List-specific flags
	tasksListCmd.Flags().Bool("pending", false, "Show only pending tasks")
	tasksListCmd.Flags().Bool("completed", false, "Show only completed tasks")
	tasksListCmd.Flags().Bool("count", false, "Show only task counts")

	tasksCmd.AddCommand(tasksListCmd)
	tasksCmd.AddCommand(tasksNextCmd)
	rootCmd.AddCommand(tasksCmd)
}

func runTasksList(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadWithFile(configFile)
	if err != nil {
		return err
	}

	// Apply task source flags
	applyTaskSourceFlags(cmd, &cfg)

	source, err := loadTaskSource(cfg)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Get flags
	pendingOnly, _ := cmd.Flags().GetBool("pending")
	completedOnly, _ := cmd.Flags().GetBool("completed")
	countOnly, _ := cmd.Flags().GetBool("count")

	// Get summary
	summary, err := source.Summary(ctx)
	if err != nil {
		return fmt.Errorf("getting task summary: %w", err)
	}

	if countOnly {
		fmt.Printf("Total: %d\n", summary.Completed+summary.Remaining)
		fmt.Printf("Completed: %d\n", summary.Completed)
		fmt.Printf("Pending: %d\n", summary.Remaining)
		return nil
	}

	// Print header
	fmt.Printf("Tasks from %s\n", sourceLabel(cfg))
	fmt.Printf("============%s\n", strings.Repeat("=", len(sourceLabel(cfg))))
	fmt.Println()

	// Get all tasks (we need to read the raw source to show both pending and completed)
	allTasks, completedTasks, err := getAllTasksWithStatus(ctx, cfg, source)
	if err != nil {
		return err
	}

	// Print tasks
	if !completedOnly {
		if len(allTasks) > 0 {
			fmt.Println("Pending:")
			for i, t := range allTasks {
				groupInfo := ""
				if t.ParallelGroup > 0 {
					groupInfo = fmt.Sprintf(" [group %d]", t.ParallelGroup)
				}
				fmt.Printf("  %d. [ ] %s%s\n", i+1, t.Title, groupInfo)
			}
			fmt.Println()
		}
	}

	if !pendingOnly {
		if len(completedTasks) > 0 {
			fmt.Println("Completed:")
			for i, t := range completedTasks {
				fmt.Printf("  %d. [x] %s\n", i+1, t.Title)
			}
			fmt.Println()
		}
	}

	// Print summary
	fmt.Printf("Summary: %d pending, %d completed\n", summary.Remaining, summary.Completed)

	return nil
}

func runTasksNext(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadWithFile(configFile)
	if err != nil {
		return err
	}

	// Apply task source flags
	applyTaskSourceFlags(cmd, &cfg)

	source, err := loadTaskSource(cfg)
	if err != nil {
		return err
	}

	ctx := context.Background()

	task, err := source.Next(ctx)
	if err != nil {
		return fmt.Errorf("getting next task: %w", err)
	}

	if task == nil {
		fmt.Println("No pending tasks.")
		return nil
	}

	fmt.Println("Next Task")
	fmt.Println("=========")
	fmt.Println()
	fmt.Printf("Title: %s\n", task.Title)
	if task.ParallelGroup > 0 {
		fmt.Printf("Parallel Group: %d\n", task.ParallelGroup)
	}
	if task.ID != "" && task.ID != task.Title {
		fmt.Printf("ID: %s\n", task.ID)
	}

	// Show summary
	summary, err := source.Summary(ctx)
	if err == nil {
		fmt.Println()
		fmt.Printf("Progress: %d/%d completed\n", summary.Completed, summary.Completed+summary.Remaining)
	}

	return nil
}

func applyTaskSourceFlags(cmd *cobra.Command, cfg *config.Config) {
	if prd, _ := cmd.Flags().GetString("prd"); prd != "" {
		cfg.PRDFile = prd
		cfg.PRDSource = config.PRDSourceMarkdown
	}

	if yamlFile, _ := cmd.Flags().GetString("yaml"); yamlFile != "" {
		cfg.PRDFile = yamlFile
		cfg.PRDSource = config.PRDSourceYAML
	}

	if github, _ := cmd.Flags().GetString("github"); github != "" {
		cfg.GitHubRepo = github
		cfg.PRDSource = config.PRDSourceGitHub
	}

	if label, _ := cmd.Flags().GetString("github-label"); label != "" {
		cfg.GitHubLabel = label
	}
}

func loadTaskSource(cfg config.Config) (tasks.Source, error) {
	switch cfg.PRDSource {
	case config.PRDSourceYAML:
		if _, err := os.Stat(cfg.PRDFile); os.IsNotExist(err) {
			return nil, fmt.Errorf("YAML file not found: %s", cfg.PRDFile)
		}
		return tasks.NewYAMLSource(cfg.PRDFile), nil
	case config.PRDSourceGitHub:
		if cfg.GitHubRepo == "" {
			return nil, fmt.Errorf("github repo required (use --github owner/repo)")
		}
		return tasks.NewGitHubSource(cfg.GitHubRepo, cfg.GitHubLabel, cfg.GitHubToken)
	default:
		if _, err := os.Stat(cfg.PRDFile); os.IsNotExist(err) {
			return nil, fmt.Errorf("PRD file not found: %s", cfg.PRDFile)
		}
		return tasks.NewMarkdownSource(cfg.PRDFile), nil
	}
}

func sourceLabel(cfg config.Config) string {
	switch cfg.PRDSource {
	case config.PRDSourceGitHub:
		label := cfg.GitHubRepo
		if cfg.GitHubLabel != "" {
			label += " (label: " + cfg.GitHubLabel + ")"
		}
		return label
	default:
		return cfg.PRDFile
	}
}

func getAllTasksWithStatus(ctx context.Context, cfg config.Config, source tasks.Source) (pending, completed []tasks.Task, err error) {
	// Get pending tasks through the normal interface
	pending, err = source.All(ctx)
	if err != nil {
		return nil, nil, err
	}

	// For completed tasks, we need to read the source directly
	switch cfg.PRDSource {
	case config.PRDSourceMarkdown:
		completed, err = getCompletedMarkdownTasks(cfg.PRDFile)
	case config.PRDSourceYAML:
		completed, err = getCompletedYAMLTasks(cfg.PRDFile)
	case config.PRDSourceGitHub:
		// GitHub source doesn't track completed in the same way
		// We'd need to query closed issues, which is expensive
		completed = nil
	}

	return pending, completed, err
}

func getCompletedMarkdownTasks(path string) ([]tasks.Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var completed []tasks.Task
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Match completed tasks: - [x] or - [X]
		if strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "- [X]") {
			title := strings.TrimSpace(trimmed[5:])
			if title != "" {
				completed = append(completed, tasks.Task{ID: title, Title: title})
			}
		}
	}

	return completed, nil
}

func getCompletedYAMLTasks(path string) ([]tasks.Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Simple parsing for completed tasks
	var completed []tasks.Task
	lines := strings.Split(string(data), "\n")

	var currentTitle string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "title:") {
			currentTitle = strings.TrimSpace(strings.TrimPrefix(trimmed, "title:"))
			// Remove quotes if present
			currentTitle = strings.Trim(currentTitle, "\"'")
		}

		if strings.HasPrefix(trimmed, "completed:") && currentTitle != "" {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "completed:"))
			if value == "true" {
				completed = append(completed, tasks.Task{ID: currentTitle, Title: currentTitle})
			}
			currentTitle = ""
		}
	}

	return completed, nil
}
