package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ncecere/ralphy/internal/config"
	"github.com/ncecere/ralphy/internal/runlog"
	"github.com/spf13/cobra"
)

var runsCmd = &cobra.Command{
	Use:   "runs",
	Short: "View run history",
	Long: `View and analyze run history from log files.

Subcommands:
  list    List recent runs
  show    Show details of a specific run
  summary Show aggregate statistics`,
}

var runsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent runs",
	Long: `List recent runs from the log file.

Shows timestamp, engine, tasks completed, and cost for each run.`,
	RunE: runRunsList,
}

var runsShowCmd = &cobra.Command{
	Use:   "show [run-number]",
	Short: "Show run details",
	Long: `Show detailed information about a specific run.

Run numbers are shown in 'ralphy runs list' output (1 = most recent).`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRunsShow,
}

var runsSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show aggregate statistics",
	Long:  `Show aggregate statistics across all runs in the log file.`,
	RunE:  runRunsSummary,
}

func init() {
	runsCmd.PersistentFlags().String("file", "", "Log file path (default: from config or ralphy-runs.log)")
	runsListCmd.Flags().Int("limit", 10, "Maximum runs to show")
	runsShowCmd.Flags().Bool("json", false, "Output as JSON")

	runsCmd.AddCommand(runsListCmd)
	runsCmd.AddCommand(runsShowCmd)
	runsCmd.AddCommand(runsSummaryCmd)
	rootCmd.AddCommand(runsCmd)
}

func getLogFile(cmd *cobra.Command) string {
	// Check flag first
	if file, _ := cmd.Flags().GetString("file"); file != "" {
		return file
	}

	// Check config
	cfg, err := config.LoadWithFile(configFile)
	if err == nil && cfg.LogFile != "" {
		return cfg.LogFile
	}

	// Default
	return "ralphy-runs.log"
}

func runRunsList(cmd *cobra.Command, args []string) error {
	logFile := getLogFile(cmd)
	limit, _ := cmd.Flags().GetInt("limit")

	entries, err := runlog.ReadLog(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("No run history found at %s\n", logFile)
			fmt.Println("Use --log-file flag when running ralphy to create a log.")
			return nil
		}
		return fmt.Errorf("reading log file: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No runs recorded yet.")
		return nil
	}

	fmt.Printf("Run History (%s)\n", logFile)
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	// Show most recent first
	start := 0
	if len(entries) > limit {
		start = len(entries) - limit
	}

	fmt.Printf("%-4s %-20s %-10s %-8s %-10s %s\n", "#", "Timestamp", "Engine", "Tasks", "Cost", "Duration")
	fmt.Println(strings.Repeat("-", 70))

	for i := len(entries) - 1; i >= start; i-- {
		e := entries[i]
		num := len(entries) - i
		ts := e.Timestamp.Local().Format("2006-01-02 15:04")
		tasks := fmt.Sprintf("%d/%d", e.Summary.Succeeded, e.Summary.TotalTasks)
		cost := fmt.Sprintf("$%.4f", e.Summary.EstimatedCost)
		duration := formatDuration(e.DurationSecs)

		fmt.Printf("%-4d %-20s %-10s %-8s %-10s %s\n", num, ts, e.Engine, tasks, cost, duration)
	}

	fmt.Println()
	fmt.Printf("Showing %d of %d runs. Use --limit to see more.\n", min(limit, len(entries)), len(entries))

	return nil
}

func runRunsShow(cmd *cobra.Command, args []string) error {
	logFile := getLogFile(cmd)
	asJSON, _ := cmd.Flags().GetBool("json")

	entries, err := runlog.ReadLog(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no run history found at %s", logFile)
		}
		return fmt.Errorf("reading log file: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no runs recorded")
	}

	// Determine which run to show
	runNum := 1 // Most recent by default
	if len(args) > 0 {
		var err error
		runNum, err = strconv.Atoi(args[0])
		if err != nil || runNum < 1 || runNum > len(entries) {
			return fmt.Errorf("invalid run number: %s (valid range: 1-%d)", args[0], len(entries))
		}
	}

	// Convert run number to index (1 = most recent = last entry)
	idx := len(entries) - runNum
	entry := entries[idx]

	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(entry)
	}

	// Pretty print
	fmt.Printf("Run #%d Details\n", runNum)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	fmt.Println("Overview:")
	fmt.Printf("  Timestamp:  %s\n", entry.Timestamp.Local().Format(time.RFC1123))
	fmt.Printf("  Version:    %s\n", entry.Version)
	fmt.Printf("  Engine:     %s\n", entry.Engine)
	if entry.Model != "" {
		fmt.Printf("  Model:      %s\n", entry.Model)
	}
	fmt.Printf("  Source:     %s\n", entry.TaskSource)
	if entry.TaskFile != "" {
		fmt.Printf("  File:       %s\n", entry.TaskFile)
	}
	fmt.Printf("  Duration:   %s\n", formatDuration(entry.DurationSecs))
	fmt.Println()

	fmt.Println("Tasks:")
	for i, t := range entry.Tasks {
		status := "\033[32m✓\033[0m"
		if t.Status == "failed" {
			status = "\033[31m✗\033[0m"
		} else if t.Status == "skipped" {
			status = "\033[33m○\033[0m"
		}

		fmt.Printf("  %d. %s %s\n", i+1, status, t.Title)
		if t.InputTokens > 0 {
			fmt.Printf("       Tokens: %d in / %d out\n", t.InputTokens, t.OutputTokens)
		}
		if t.Cost > 0 {
			fmt.Printf("       Cost: $%.4f\n", t.Cost)
		}
		if t.DurationSecs > 0 {
			fmt.Printf("       Duration: %s\n", formatDuration(t.DurationSecs))
		}
		if t.Retries > 0 {
			fmt.Printf("       Retries: %d\n", t.Retries)
		}
		if t.PRUrl != "" {
			fmt.Printf("       PR: %s\n", t.PRUrl)
		}
		if t.Error != "" {
			fmt.Printf("       Error: %s\n", t.Error)
		}
	}
	fmt.Println()

	fmt.Println("Summary:")
	fmt.Printf("  Total Tasks:    %d\n", entry.Summary.TotalTasks)
	fmt.Printf("  Succeeded:      %d\n", entry.Summary.Succeeded)
	fmt.Printf("  Failed:         %d\n", entry.Summary.Failed)
	if entry.Summary.Skipped > 0 {
		fmt.Printf("  Skipped:        %d\n", entry.Summary.Skipped)
	}
	fmt.Printf("  Total Tokens:   %d\n", entry.Summary.TotalTokens)
	fmt.Printf("  Estimated Cost: $%.4f\n", entry.Summary.EstimatedCost)

	return nil
}

func runRunsSummary(cmd *cobra.Command, args []string) error {
	logFile := getLogFile(cmd)

	entries, err := runlog.ReadLog(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no run history found at %s", logFile)
		}
		return fmt.Errorf("reading log file: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no runs recorded")
	}

	// Aggregate stats
	var totalRuns, totalTasks, totalSucceeded, totalFailed int
	var totalTokens, totalInputTokens, totalOutputTokens int
	var totalCost, totalDuration float64
	engineCounts := make(map[string]int)

	for _, e := range entries {
		totalRuns++
		totalTasks += e.Summary.TotalTasks
		totalSucceeded += e.Summary.Succeeded
		totalFailed += e.Summary.Failed
		totalTokens += e.Summary.TotalTokens
		totalInputTokens += e.Summary.InputTokens
		totalOutputTokens += e.Summary.OutputTokens
		totalCost += e.Summary.EstimatedCost
		totalDuration += e.DurationSecs
		engineCounts[e.Engine]++
	}

	// Find date range
	oldest := entries[0].Timestamp
	newest := entries[len(entries)-1].Timestamp

	fmt.Printf("Run History Summary (%s)\n", logFile)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	fmt.Println("Overview:")
	fmt.Printf("  Total Runs:     %d\n", totalRuns)
	fmt.Printf("  Date Range:     %s to %s\n",
		oldest.Local().Format("2006-01-02"),
		newest.Local().Format("2006-01-02"))
	fmt.Printf("  Total Duration: %s\n", formatDuration(totalDuration))
	fmt.Println()

	fmt.Println("Tasks:")
	fmt.Printf("  Total:          %d\n", totalTasks)
	fmt.Printf("  Succeeded:      %d (%.1f%%)\n", totalSucceeded, percent(totalSucceeded, totalTasks))
	fmt.Printf("  Failed:         %d (%.1f%%)\n", totalFailed, percent(totalFailed, totalTasks))
	fmt.Println()

	fmt.Println("Tokens:")
	fmt.Printf("  Input:          %d\n", totalInputTokens)
	fmt.Printf("  Output:         %d\n", totalOutputTokens)
	fmt.Printf("  Total:          %d\n", totalTokens)
	fmt.Println()

	fmt.Println("Cost:")
	fmt.Printf("  Total:          $%.4f\n", totalCost)
	fmt.Printf("  Avg per Run:    $%.4f\n", totalCost/float64(totalRuns))
	fmt.Printf("  Avg per Task:   $%.4f\n", totalCost/float64(max(totalTasks, 1)))
	fmt.Println()

	fmt.Println("Engines Used:")
	for engine, count := range engineCounts {
		fmt.Printf("  %-12s %d runs (%.1f%%)\n", engine, count, percent(count, totalRuns))
	}

	return nil
}

func formatDuration(secs float64) string {
	d := time.Duration(secs * float64(time.Second))
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", secs)
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

func percent(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
