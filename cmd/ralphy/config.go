package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/ncecere/ralphy/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage ralphy configuration",
	Long: `View and validate ralphy configuration.

Subcommands:
  show      Display merged configuration with sources
  validate  Validate configuration files`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display merged configuration",
	Long: `Display the effective configuration after merging all sources.

Shows values from:
  1. Global config (~/.config/ralphy/ralphy.yaml)
  2. Local config (./ralphy.yaml)
  3. Environment variables (RALPHY_*)

Use --sources to see where each value comes from.`,
	RunE: runConfigShow,
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration files",
	Long: `Check configuration files for errors and issues.

Validates:
  - YAML syntax
  - Known configuration keys
  - Valid values for enums (ai_engine, prd_source)
  - File paths exist (when applicable)`,
	RunE: runConfigValidate,
}

func init() {
	configShowCmd.Flags().Bool("sources", false, "Show source of each configuration value")
	configShowCmd.Flags().Bool("yaml", false, "Output as YAML")

	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configValidateCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	showSources, _ := cmd.Flags().GetBool("sources")
	asYAML, _ := cmd.Flags().GetBool("yaml")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if asYAML {
		return outputConfigYAML(cfg)
	}

	if showSources {
		return outputConfigWithSources(cfg)
	}

	return outputConfigTable(cfg)
}

func outputConfigYAML(cfg config.Config) error {
	// Create a clean struct for output (exclude internal fields)
	out := map[string]interface{}{
		"ai_engine":       cfg.AIEngine,
		"models":          cfg.Models,
		"prd_source":      cfg.PRDSource,
		"prd_file":        cfg.PRDFile,
		"skip_tests":      cfg.SkipTests,
		"skip_lint":       cfg.SkipLint,
		"dry_run":         cfg.DryRun,
		"max_iterations":  cfg.MaxIterations,
		"max_retries":     cfg.MaxRetries,
		"retry_delay":     cfg.RetryDelaySeconds,
		"verbose":         cfg.Verbose,
		"parallel":        cfg.Parallel,
		"max_parallel":    cfg.MaxParallel,
		"branch_per_task": cfg.BranchPerTask,
		"base_branch":     cfg.BaseBranch,
		"create_pr":       cfg.CreatePR,
		"pr_draft":        cfg.PRDraft,
		"github_repo":     cfg.GitHubRepo,
		"github_label":    cfg.GitHubLabel,
	}

	enc := yaml.NewEncoder(os.Stdout)
	enc.SetIndent(2)
	return enc.Encode(out)
}

func outputConfigTable(cfg config.Config) error {
	fmt.Println("Effective Configuration")
	fmt.Println("=======================")
	fmt.Println()

	// Show config file sources
	globalPath := config.GlobalConfigPath()
	localPath := config.LocalConfigPath()

	fmt.Println("Sources:")
	if _, err := os.Stat(globalPath); err == nil {
		fmt.Printf("  Global: %s\n", globalPath)
	} else {
		fmt.Printf("  Global: (not found)\n")
	}
	if _, err := os.Stat(localPath); err == nil {
		fmt.Printf("  Local:  %s\n", localPath)
	} else {
		fmt.Printf("  Local:  (not found)\n")
	}
	fmt.Println()

	fmt.Println("AI Engine:")
	fmt.Printf("  %-20s %s\n", "engine:", cfg.AIEngine)
	fmt.Printf("  %-20s %s\n", "model (resolved):", cfg.ResolvedModel())
	if cfg.Models.Claude != "" {
		fmt.Printf("  %-20s %s\n", "models.claude:", cfg.Models.Claude)
	}
	if cfg.Models.OpenCode != "" {
		fmt.Printf("  %-20s %s\n", "models.opencode:", cfg.Models.OpenCode)
	}
	if cfg.Models.Codex != "" {
		fmt.Printf("  %-20s %s\n", "models.codex:", cfg.Models.Codex)
	}
	if cfg.Models.Cursor != "" {
		fmt.Printf("  %-20s %s\n", "models.cursor:", cfg.Models.Cursor)
	}
	fmt.Println()

	fmt.Println("Task Source:")
	fmt.Printf("  %-20s %s\n", "source:", cfg.PRDSource)
	fmt.Printf("  %-20s %s\n", "file:", cfg.PRDFile)
	if cfg.GitHubRepo != "" {
		fmt.Printf("  %-20s %s\n", "github_repo:", cfg.GitHubRepo)
	}
	if cfg.GitHubLabel != "" {
		fmt.Printf("  %-20s %s\n", "github_label:", cfg.GitHubLabel)
	}
	fmt.Println()

	fmt.Println("Workflow:")
	fmt.Printf("  %-20s %v\n", "skip_tests:", cfg.SkipTests)
	fmt.Printf("  %-20s %v\n", "skip_lint:", cfg.SkipLint)
	fmt.Printf("  %-20s %v\n", "dry_run:", cfg.DryRun)
	fmt.Printf("  %-20s %d\n", "max_iterations:", cfg.MaxIterations)
	fmt.Printf("  %-20s %d\n", "max_retries:", cfg.MaxRetries)
	fmt.Printf("  %-20s %d\n", "retry_delay:", cfg.RetryDelaySeconds)
	fmt.Printf("  %-20s %v\n", "verbose:", cfg.Verbose)
	fmt.Println()

	fmt.Println("Parallel Execution:")
	fmt.Printf("  %-20s %v\n", "parallel:", cfg.Parallel)
	fmt.Printf("  %-20s %d\n", "max_parallel:", cfg.MaxParallel)
	fmt.Println()

	fmt.Println("Git Workflow:")
	fmt.Printf("  %-20s %v\n", "branch_per_task:", cfg.BranchPerTask)
	fmt.Printf("  %-20s %s\n", "base_branch:", valueOrDefault(cfg.BaseBranch, "(current)"))
	fmt.Printf("  %-20s %v\n", "create_pr:", cfg.CreatePR)
	fmt.Printf("  %-20s %v\n", "pr_draft:", cfg.PRDraft)

	return nil
}

func outputConfigWithSources(cfg config.Config) error {
	fmt.Println("Configuration with Sources")
	fmt.Println("==========================")
	fmt.Println()

	globalPath := config.GlobalConfigPath()
	localPath := config.LocalConfigPath()

	// Load configs separately to determine sources
	var globalCfg, localCfg map[string]interface{}

	if data, err := os.ReadFile(globalPath); err == nil {
		yaml.Unmarshal(data, &globalCfg)
	}
	if data, err := os.ReadFile(localPath); err == nil {
		yaml.Unmarshal(data, &localCfg)
	}

	// Helper to determine source
	getSource := func(key string, value interface{}) string {
		// Check environment variable first
		envKey := "RALPHY_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		if os.Getenv(envKey) != "" {
			return "env"
		}

		// Check local config
		if localCfg != nil {
			if v, ok := getNestedValue(localCfg, key); ok && !isZero(v) {
				return "local"
			}
		}

		// Check global config
		if globalCfg != nil {
			if v, ok := getNestedValue(globalCfg, key); ok && !isZero(v) {
				return "global"
			}
		}

		return "default"
	}

	printWithSource := func(key string, value interface{}) {
		source := getSource(key, value)
		sourceTag := ""
		switch source {
		case "env":
			sourceTag = "\033[33m[env]\033[0m"
		case "local":
			sourceTag = "\033[32m[local]\033[0m"
		case "global":
			sourceTag = "\033[36m[global]\033[0m"
		default:
			sourceTag = "\033[90m[default]\033[0m"
		}
		fmt.Printf("  %-20s %-20v %s\n", key+":", value, sourceTag)
	}

	fmt.Println("AI Engine:")
	printWithSource("ai_engine", cfg.AIEngine)
	printWithSource("models.claude", cfg.Models.Claude)
	printWithSource("models.opencode", cfg.Models.OpenCode)
	printWithSource("models.codex", cfg.Models.Codex)
	printWithSource("models.cursor", cfg.Models.Cursor)
	fmt.Println()

	fmt.Println("Task Source:")
	printWithSource("prd_source", cfg.PRDSource)
	printWithSource("prd_file", cfg.PRDFile)
	printWithSource("github_repo", cfg.GitHubRepo)
	printWithSource("github_label", cfg.GitHubLabel)
	fmt.Println()

	fmt.Println("Workflow:")
	printWithSource("skip_tests", cfg.SkipTests)
	printWithSource("skip_lint", cfg.SkipLint)
	printWithSource("dry_run", cfg.DryRun)
	printWithSource("max_iterations", cfg.MaxIterations)
	printWithSource("max_retries", cfg.MaxRetries)
	printWithSource("retry_delay", cfg.RetryDelaySeconds)
	printWithSource("verbose", cfg.Verbose)
	fmt.Println()

	fmt.Println("Parallel Execution:")
	printWithSource("parallel", cfg.Parallel)
	printWithSource("max_parallel", cfg.MaxParallel)
	fmt.Println()

	fmt.Println("Git Workflow:")
	printWithSource("branch_per_task", cfg.BranchPerTask)
	printWithSource("base_branch", cfg.BaseBranch)
	printWithSource("create_pr", cfg.CreatePR)
	printWithSource("pr_draft", cfg.PRDraft)

	return nil
}

func runConfigValidate(cmd *cobra.Command, args []string) error {
	fmt.Println("Validating Configuration")
	fmt.Println("========================")
	fmt.Println()

	var errors []string
	var warnings []string

	globalPath := config.GlobalConfigPath()
	localPath := config.LocalConfigPath()

	// Validate global config
	if _, err := os.Stat(globalPath); err == nil {
		fmt.Printf("Checking %s...\n", globalPath)
		errs, warns := validateConfigFile(globalPath)
		errors = append(errors, errs...)
		warnings = append(warnings, warns...)
	} else {
		fmt.Printf("Global config: not found (optional)\n")
	}

	// Validate local config
	if _, err := os.Stat(localPath); err == nil {
		fmt.Printf("Checking %s...\n", localPath)
		errs, warns := validateConfigFile(localPath)
		errors = append(errors, errs...)
		warnings = append(warnings, warns...)
	} else {
		fmt.Printf("Local config: not found (optional)\n")
	}

	// Try loading merged config
	fmt.Println()
	fmt.Println("Checking merged configuration...")
	cfg, err := config.Load()
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to load config: %v", err))
	} else {
		// Validate merged config values
		errs, warns := validateConfigValues(cfg)
		errors = append(errors, errs...)
		warnings = append(warnings, warns...)
	}

	fmt.Println()

	if len(errors) > 0 {
		fmt.Println("\033[31mErrors:\033[0m")
		for _, e := range errors {
			fmt.Printf("  ✗ %s\n", e)
		}
		fmt.Println()
	}

	if len(warnings) > 0 {
		fmt.Println("\033[33mWarnings:\033[0m")
		for _, w := range warnings {
			fmt.Printf("  ! %s\n", w)
		}
		fmt.Println()
	}

	if len(errors) == 0 && len(warnings) == 0 {
		fmt.Println("\033[32m✓ Configuration is valid\033[0m")
	} else if len(errors) == 0 {
		fmt.Printf("\033[33m✓ Configuration is valid with %d warning(s)\033[0m\n", len(warnings))
	} else {
		fmt.Printf("\033[31m✗ Configuration has %d error(s)\033[0m\n", len(errors))
		return fmt.Errorf("validation failed")
	}

	return nil
}

func validateConfigFile(path string) (errors []string, warnings []string) {
	data, err := os.ReadFile(path)
	if err != nil {
		errors = append(errors, fmt.Sprintf("%s: cannot read file: %v", path, err))
		return
	}

	// Check YAML syntax
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		errors = append(errors, fmt.Sprintf("%s: invalid YAML: %v", path, err))
		return
	}

	// Check for unknown keys
	knownKeys := map[string]bool{
		"ai_engine": true, "model": true, "models": true,
		"prd_source": true, "prd_file": true,
		"skip_tests": true, "skip_lint": true, "dry_run": true,
		"max_iterations": true, "max_retries": true, "retry_delay": true,
		"verbose": true, "parallel": true, "max_parallel": true,
		"branch_per_task": true, "base_branch": true, "create_pr": true, "pr_draft": true,
		"github_repo": true, "github_label": true, "github_token": true,
	}

	for key := range raw {
		if !knownKeys[key] {
			warnings = append(warnings, fmt.Sprintf("%s: unknown key '%s'", path, key))
		}
	}

	return
}

func validateConfigValues(cfg config.Config) (errors []string, warnings []string) {
	// Validate ai_engine
	validEngines := map[string]bool{
		"claude": true, "opencode": true, "codex": true, "cursor": true,
	}
	if !validEngines[cfg.AIEngine] {
		errors = append(errors, fmt.Sprintf("invalid ai_engine '%s' (must be claude, opencode, codex, or cursor)", cfg.AIEngine))
	}

	// Validate prd_source
	validSources := map[string]bool{
		"markdown": true, "yaml": true, "github": true,
	}
	if !validSources[cfg.PRDSource] {
		errors = append(errors, fmt.Sprintf("invalid prd_source '%s' (must be markdown, yaml, or github)", cfg.PRDSource))
	}

	// Check PRD file exists for file-based sources
	if cfg.PRDSource == "markdown" || cfg.PRDSource == "yaml" {
		if _, err := os.Stat(cfg.PRDFile); os.IsNotExist(err) {
			warnings = append(warnings, fmt.Sprintf("prd_file '%s' does not exist", cfg.PRDFile))
		}
	}

	// Check GitHub source requirements
	if cfg.PRDSource == "github" && cfg.GitHubRepo == "" {
		errors = append(errors, "github_repo is required when prd_source is 'github'")
	}

	// Validate numeric ranges
	if cfg.MaxRetries < 0 {
		errors = append(errors, "max_retries cannot be negative")
	}
	if cfg.RetryDelaySeconds < 0 {
		errors = append(errors, "retry_delay cannot be negative")
	}
	if cfg.MaxParallel < 1 {
		errors = append(errors, "max_parallel must be at least 1")
	}

	return
}

func valueOrDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func getNestedValue(m map[string]interface{}, key string) (interface{}, bool) {
	parts := strings.Split(key, ".")
	current := m

	for i, part := range parts {
		if i == len(parts)-1 {
			v, ok := current[part]
			return v, ok
		}
		if nested, ok := current[part].(map[string]interface{}); ok {
			current = nested
		} else {
			return nil, false
		}
	}
	return nil, false
}

func isZero(v interface{}) bool {
	if v == nil {
		return true
	}
	return reflect.ValueOf(v).IsZero()
}
