package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/ncecere/ralphy/internal/config"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup wizard",
	Long: `Interactively configure Ralphy with a step-by-step wizard.

The wizard will guide you through:
  1. Choosing a default AI engine
  2. Selecting a default model for that engine
  3. Configuring task source preferences
  4. Setting up Git workflow options

Use --local to save to project config instead of global.`,
	RunE: runSetup,
}

func init() {
	setupCmd.Flags().Bool("local", false, "Save to local config (./ralphy.yaml) instead of global")
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) error {
	local, _ := cmd.Flags().GetBool("local")

	var configPath string
	if local {
		configPath = config.LocalConfigPath()
	} else {
		configPath = config.GlobalConfigPath()
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║         Ralphy Setup Wizard              ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()

	if local {
		fmt.Println("Creating project-local config: ./ralphy.yaml")
	} else {
		fmt.Println("Creating global config: " + configPath)
	}
	fmt.Println()

	// Check if config exists
	if config.ConfigExists(configPath) {
		fmt.Printf("Config already exists at %s\n", configPath)
		if !promptYesNo(reader, "Overwrite existing config?", false) {
			fmt.Println("Setup cancelled.")
			return nil
		}
		fmt.Println()
	}

	// Step 1: Choose AI engine
	fmt.Println("Step 1: Choose Default AI Engine")
	fmt.Println("---------------------------------")
	engines := []struct {
		name        string
		description string
		installed   bool
	}{
		{"claude", "Claude Code by Anthropic", isInstalled("claude")},
		{"opencode", "OpenCode - multi-provider AI coding", isInstalled("opencode")},
		{"codex", "OpenAI Codex CLI", isInstalled("codex")},
		{"cursor", "Cursor AI editor", isInstalled("cursor") || isInstalled("agent")},
	}

	fmt.Println()
	for i, e := range engines {
		status := "\033[32m✓\033[0m"
		if !e.installed {
			status = "\033[90m○\033[0m"
		}
		fmt.Printf("  %d) %s %-12s %s\n", i+1, status, e.name, e.description)
	}
	fmt.Println()

	engineChoice := promptChoice(reader, "Select engine", 1, len(engines), 1)
	selectedEngine := engines[engineChoice-1].name
	fmt.Printf("Selected: %s\n\n", selectedEngine)

	// Step 2: Choose model
	fmt.Println("Step 2: Choose Default Model")
	fmt.Println("----------------------------")

	var selectedModel string
	models, err := getModelList(selectedEngine)
	if err != nil || len(models) == 0 {
		fmt.Println("(Model list not available for this engine)")
		selectedModel = promptString(reader, "Enter model name (or leave empty)", "")
	} else {
		fmt.Println()
		for i, m := range models {
			fmt.Printf("  %d) %s\n", i+1, m)
		}
		fmt.Println()

		modelChoice := promptChoiceOrCustom(reader, "Select model (or type custom)", 1, len(models))
		if modelChoice > 0 && modelChoice <= len(models) {
			selectedModel = models[modelChoice-1]
		} else if modelChoice == -1 {
			selectedModel = promptString(reader, "Enter custom model name", "")
		}
	}
	if selectedModel != "" {
		fmt.Printf("Selected: %s\n\n", selectedModel)
	} else {
		fmt.Println("Skipped model selection")
		fmt.Println()
	}

	// Step 3: Task source preferences
	fmt.Println("Step 3: Task Source Preferences")
	fmt.Println("-------------------------------")
	fmt.Println()
	fmt.Println("  1) markdown  - PRD.md with checkbox tasks (default)")
	fmt.Println("  2) yaml      - Structured YAML task file")
	fmt.Println("  3) github    - GitHub issues as tasks")
	fmt.Println()

	sourceChoice := promptChoice(reader, "Select default task source", 1, 3, 1)
	sources := []string{"markdown", "yaml", "github"}
	selectedSource := sources[sourceChoice-1]

	var prdFile, githubRepo, githubLabel string
	switch selectedSource {
	case "markdown":
		prdFile = promptString(reader, "PRD file path", "PRD.md")
	case "yaml":
		prdFile = promptString(reader, "YAML tasks file path", "tasks.yaml")
	case "github":
		githubRepo = promptString(reader, "GitHub repo (owner/repo)", "")
		githubLabel = promptString(reader, "Filter by label (optional)", "")
	}
	fmt.Println()

	// Step 4: Git workflow
	fmt.Println("Step 4: Git Workflow Options")
	fmt.Println("----------------------------")
	fmt.Println()

	branchPerTask := promptYesNo(reader, "Create a branch for each task?", false)
	var baseBranch string
	var createPR, draftPR bool
	if branchPerTask {
		baseBranch = promptString(reader, "Base branch", "main")
		createPR = promptYesNo(reader, "Automatically create PRs?", false)
		if createPR {
			draftPR = promptYesNo(reader, "Create PRs as drafts?", false)
		}
	}
	fmt.Println()

	// Step 5: Workflow options
	fmt.Println("Step 5: Workflow Options")
	fmt.Println("------------------------")
	fmt.Println()

	parallel := promptYesNo(reader, "Enable parallel task execution?", false)
	var maxParallel int
	if parallel {
		maxParallel = promptInt(reader, "Max parallel tasks", 3)
	}

	skipTests := promptYesNo(reader, "Skip tests by default?", false)
	skipLint := promptYesNo(reader, "Skip linting by default?", false)
	fmt.Println()

	// Summary
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║            Configuration Summary         ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  AI Engine:       %s\n", selectedEngine)
	if selectedModel != "" {
		fmt.Printf("  Model:           %s\n", selectedModel)
	}
	fmt.Printf("  Task Source:     %s\n", selectedSource)
	if prdFile != "" {
		fmt.Printf("  Task File:       %s\n", prdFile)
	}
	if githubRepo != "" {
		fmt.Printf("  GitHub Repo:     %s\n", githubRepo)
	}
	if branchPerTask {
		fmt.Printf("  Branch per task: yes (base: %s)\n", baseBranch)
		if createPR {
			if draftPR {
				fmt.Println("  Auto PR:         yes (draft)")
			} else {
				fmt.Println("  Auto PR:         yes")
			}
		}
	}
	if parallel {
		fmt.Printf("  Parallel:        yes (max: %d)\n", maxParallel)
	}
	if skipTests {
		fmt.Println("  Skip tests:      yes")
	}
	if skipLint {
		fmt.Println("  Skip lint:       yes")
	}
	fmt.Println()
	fmt.Printf("  Config file:     %s\n", configPath)
	fmt.Println()

	if !promptYesNo(reader, "Save this configuration?", true) {
		fmt.Println("Setup cancelled.")
		return nil
	}

	// Write config
	if err := config.WriteDefaultConfig(configPath); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	// Apply settings
	if err := config.SetDefaultEngine(configPath, selectedEngine); err != nil {
		return fmt.Errorf("failed to set engine: %w", err)
	}

	if selectedModel != "" {
		if err := config.SetEngineModel(configPath, selectedEngine, selectedModel); err != nil {
			return fmt.Errorf("failed to set model: %w", err)
		}
	}

	// Apply additional settings by updating the config
	if err := applySetupSettings(configPath, setupSettings{
		prdSource:     selectedSource,
		prdFile:       prdFile,
		githubRepo:    githubRepo,
		githubLabel:   githubLabel,
		branchPerTask: branchPerTask,
		baseBranch:    baseBranch,
		createPR:      createPR,
		draftPR:       draftPR,
		parallel:      parallel,
		maxParallel:   maxParallel,
		skipTests:     skipTests,
		skipLint:      skipLint,
	}); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	fmt.Println()
	fmt.Println("\033[32m✓ Configuration saved!\033[0m")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Run 'ralphy doctor' to verify your setup")
	fmt.Println("  2. Create a task file (PRD.md or tasks.yaml)")
	fmt.Println("  3. Run 'ralphy' to start the AI coding loop")
	fmt.Println()

	return nil
}

type setupSettings struct {
	prdSource     string
	prdFile       string
	githubRepo    string
	githubLabel   string
	branchPerTask bool
	baseBranch    string
	createPR      bool
	draftPR       bool
	parallel      bool
	maxParallel   int
	skipTests     bool
	skipLint      bool
}

func applySetupSettings(configPath string, s setupSettings) error {
	// Read existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	content := string(data)

	// Replace values in the YAML (simple string replacement for known keys)
	replacements := map[string]string{
		"prd_source: markdown": fmt.Sprintf("prd_source: %s", s.prdSource),
		"prd_file: PRD.md":     fmt.Sprintf("prd_file: %s", s.prdFile),
	}

	if s.githubRepo != "" {
		replacements["github_repo: \"\""] = fmt.Sprintf("github_repo: %s", s.githubRepo)
	}
	if s.githubLabel != "" {
		replacements["github_label: \"\""] = fmt.Sprintf("github_label: %s", s.githubLabel)
	}
	if s.branchPerTask {
		replacements["branch_per_task: false"] = "branch_per_task: true"
	}
	if s.baseBranch != "" {
		replacements["base_branch: \"\""] = fmt.Sprintf("base_branch: %s", s.baseBranch)
	}
	if s.createPR {
		replacements["create_pr: false"] = "create_pr: true"
	}
	if s.draftPR {
		replacements["pr_draft: false"] = "pr_draft: true"
	}
	if s.parallel {
		replacements["parallel: false"] = "parallel: true"
	}
	if s.maxParallel > 0 && s.maxParallel != 3 {
		replacements["max_parallel: 3"] = fmt.Sprintf("max_parallel: %d", s.maxParallel)
	}
	if s.skipTests {
		replacements["skip_tests: false"] = "skip_tests: true"
	}
	if s.skipLint {
		replacements["skip_lint: false"] = "skip_lint: true"
	}

	for old, new := range replacements {
		content = strings.Replace(content, old, new, 1)
	}

	return os.WriteFile(configPath, []byte(content), 0o644)
}

func promptString(reader *bufio.Reader, prompt, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultVal)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultVal
	}
	return input
}

func promptYesNo(reader *bufio.Reader, prompt string, defaultYes bool) bool {
	defaultStr := "y/N"
	if defaultYes {
		defaultStr = "Y/n"
	}

	fmt.Printf("%s [%s]: ", prompt, defaultStr)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return defaultYes
	}

	return input == "y" || input == "yes"
}

func promptChoice(reader *bufio.Reader, prompt string, min, max, defaultVal int) int {
	for {
		fmt.Printf("%s [%d-%d, default %d]: ", prompt, min, max, defaultVal)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			return defaultVal
		}

		val, err := strconv.Atoi(input)
		if err != nil || val < min || val > max {
			fmt.Printf("Please enter a number between %d and %d\n", min, max)
			continue
		}
		return val
	}
}

func promptChoiceOrCustom(reader *bufio.Reader, prompt string, min, max int) int {
	for {
		fmt.Printf("%s [%d-%d or 'c' for custom]: ", prompt, min, max)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			return 1 // default to first option
		}

		if strings.ToLower(input) == "c" || strings.ToLower(input) == "custom" {
			return -1 // signal for custom input
		}

		val, err := strconv.Atoi(input)
		if err != nil || val < min || val > max {
			fmt.Printf("Please enter a number between %d and %d, or 'c' for custom\n", min, max)
			continue
		}
		return val
	}
}

func promptInt(reader *bufio.Reader, prompt string, defaultVal int) int {
	for {
		fmt.Printf("%s [default %d]: ", prompt, defaultVal)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			return defaultVal
		}

		val, err := strconv.Atoi(input)
		if err != nil || val < 1 {
			fmt.Println("Please enter a positive number")
			continue
		}
		return val
	}
}

func isInstalled(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
