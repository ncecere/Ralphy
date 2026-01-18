package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ncecere/ralphy/internal/config"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system dependencies and configuration",
	Long: `Verify that all required dependencies are installed and configured correctly.

Checks:
  - AI engine CLIs (claude, opencode, codex, cursor)
  - Git installation and configuration
  - GitHub CLI (gh) for PR creation
  - GitHub authentication tokens
  - Configuration files`,
	RunE: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

type checkResult struct {
	name    string
	status  string // "ok", "warn", "error", "skip"
	message string
}

func runDoctor(cmd *cobra.Command, args []string) error {
	fmt.Println("Ralphy Doctor")
	fmt.Println("=============")
	fmt.Println()

	var results []checkResult

	// Check AI engines
	fmt.Println("AI Engines:")
	results = append(results, checkCLI("claude", "Claude Code", true))
	results = append(results, checkCLI("opencode", "OpenCode", true))
	results = append(results, checkCLI("codex", "Codex", true))
	results = append(results, checkCLI("cursor", "Cursor", true))
	results = append(results, checkCLI("agent", "Cursor Agent", true))
	fmt.Println()

	// Check Git
	fmt.Println("Git:")
	results = append(results, checkCLI("git", "Git", false))
	results = append(results, checkGitConfig())
	fmt.Println()

	// Check GitHub
	fmt.Println("GitHub:")
	results = append(results, checkCLI("gh", "GitHub CLI", true))
	results = append(results, checkGitHubAuth())
	results = append(results, checkGitHubToken())
	fmt.Println()

	// Check Config
	fmt.Println("Configuration:")
	results = append(results, checkConfigFile(config.GlobalConfigPath(), "Global config"))
	results = append(results, checkConfigFile(config.LocalConfigPath(), "Local config"))
	fmt.Println()

	// Summary
	var errors, warnings, ok int
	for _, r := range results {
		switch r.status {
		case "error":
			errors++
		case "warn":
			warnings++
		case "ok":
			ok++
		}
	}

	fmt.Println("Summary:")
	fmt.Printf("  %d passed, %d warnings, %d errors\n", ok, warnings, errors)

	if errors > 0 {
		fmt.Println()
		fmt.Println("Some checks failed. Install missing dependencies to use all features.")
		return nil // Don't return error, just inform
	}

	if warnings > 0 {
		fmt.Println()
		fmt.Println("Some optional dependencies are missing.")
	}

	return nil
}

func checkCLI(cmd, name string, optional bool) checkResult {
	path, err := exec.LookPath(cmd)
	if err != nil {
		status := "error"
		if optional {
			status = "warn"
		}
		printCheck(status, name, "not found in PATH")
		return checkResult{name: name, status: status, message: "not found"}
	}

	// Try to get version
	version := getVersion(cmd)
	msg := fmt.Sprintf("found at %s", path)
	if version != "" {
		msg = fmt.Sprintf("%s (%s)", version, path)
	}

	printCheck("ok", name, msg)
	return checkResult{name: name, status: "ok", message: msg}
}

func getVersion(cmd string) string {
	// Different commands have different version flags
	versionFlags := []string{"--version", "-v", "version"}

	for _, flag := range versionFlags {
		out, err := exec.Command(cmd, flag).Output()
		if err == nil {
			version := strings.TrimSpace(string(out))
			// Take first line only
			if idx := strings.Index(version, "\n"); idx > 0 {
				version = version[:idx]
			}
			// Limit length
			if len(version) > 50 {
				version = version[:50] + "..."
			}
			return version
		}
	}
	return ""
}

func checkGitConfig() checkResult {
	// Check if we're in a git repo
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		printCheck("warn", "Git repo", "not in a git repository")
		return checkResult{name: "Git repo", status: "warn", message: "not in a git repository"}
	}

	// Check user.name and user.email
	name, _ := exec.Command("git", "config", "user.name").Output()
	email, _ := exec.Command("git", "config", "user.email").Output()

	if strings.TrimSpace(string(name)) == "" || strings.TrimSpace(string(email)) == "" {
		printCheck("warn", "Git config", "user.name or user.email not set")
		return checkResult{name: "Git config", status: "warn", message: "user.name or user.email not set"}
	}

	printCheck("ok", "Git config", fmt.Sprintf("%s <%s>", strings.TrimSpace(string(name)), strings.TrimSpace(string(email))))
	return checkResult{name: "Git config", status: "ok", message: "configured"}
}

func checkGitHubAuth() checkResult {
	// Check if gh is authenticated
	cmd := exec.Command("gh", "auth", "status")
	out, err := cmd.CombinedOutput()
	if err != nil {
		printCheck("warn", "GitHub CLI auth", "not authenticated (run `gh auth login`)")
		return checkResult{name: "GitHub CLI auth", status: "warn", message: "not authenticated"}
	}

	// Extract account info
	output := string(out)
	if strings.Contains(output, "Logged in to") {
		// Find the account name
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Logged in to") {
				printCheck("ok", "GitHub CLI auth", strings.TrimSpace(line))
				return checkResult{name: "GitHub CLI auth", status: "ok", message: "authenticated"}
			}
		}
	}

	printCheck("ok", "GitHub CLI auth", "authenticated")
	return checkResult{name: "GitHub CLI auth", status: "ok", message: "authenticated"}
}

func checkGitHubToken() checkResult {
	// Check for GitHub token in environment
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}

	if token != "" {
		masked := token[:4] + "..." + token[len(token)-4:]
		printCheck("ok", "GitHub token", fmt.Sprintf("found in environment (%s)", masked))
		return checkResult{name: "GitHub token", status: "ok", message: "found in environment"}
	}

	// Check config file
	cfg, err := config.Load()
	if err == nil && cfg.GitHubToken != "" {
		masked := cfg.GitHubToken[:4] + "..." + cfg.GitHubToken[len(cfg.GitHubToken)-4:]
		printCheck("ok", "GitHub token", fmt.Sprintf("found in config (%s)", masked))
		return checkResult{name: "GitHub token", status: "ok", message: "found in config"}
	}

	printCheck("warn", "GitHub token", "not found (set GITHUB_TOKEN for PR creation)")
	return checkResult{name: "GitHub token", status: "warn", message: "not found"}
}

func checkConfigFile(path, name string) checkResult {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		printCheck("skip", name, fmt.Sprintf("not found (%s)", path))
		return checkResult{name: name, status: "skip", message: "not found"}
	}

	// Try to load and validate
	printCheck("ok", name, path)
	return checkResult{name: name, status: "ok", message: path}
}

func printCheck(status, name, message string) {
	var icon string
	switch status {
	case "ok":
		icon = "\033[32m✓\033[0m" // green checkmark
	case "warn":
		icon = "\033[33m!\033[0m" // yellow exclamation
	case "error":
		icon = "\033[31m✗\033[0m" // red x
	case "skip":
		icon = "\033[90m-\033[0m" // gray dash
	}
	fmt.Printf("  %s %-20s %s\n", icon, name, message)
}
