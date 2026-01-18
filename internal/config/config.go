package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const Version = "0.4.2"

// GlobalConfigDir returns the global config directory path.
func GlobalConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "ralphy")
}

// GlobalConfigPath returns the global config file path.
func GlobalConfigPath() string {
	return filepath.Join(GlobalConfigDir(), "ralphy.yaml")
}

// LocalConfigPath returns the local (project) config file path.
func LocalConfigPath() string {
	return "ralphy.yaml"
}

const (
	AIEngineClaude   = "claude"
	AIEngineOpenCode = "opencode"
	AIEngineCursor   = "cursor"
	AIEngineCodex    = "codex"
)

const (
	PRDSourceMarkdown = "markdown"
	PRDSourceYAML     = "yaml"
	PRDSourceGitHub   = "github"
)

// EngineModels holds per-engine model defaults.
type EngineModels struct {
	Claude   string `mapstructure:"claude"`
	OpenCode string `mapstructure:"opencode"`
	Codex    string `mapstructure:"codex"`
	Cursor   string `mapstructure:"cursor"`
}

// PRConfig holds pull request configuration.
type PRConfig struct {
	// Template for PR body. Supports variables: {{.Task}}, {{.Engine}}, {{.Model}},
	// {{.InputTokens}}, {{.OutputTokens}}, {{.Cost}}, {{.Duration}}, {{.FilesChanged}}
	BodyTemplate string `mapstructure:"body_template"`

	// Labels to add to the PR
	Labels []string `mapstructure:"labels"`

	// Reviewers to request (GitHub usernames)
	Reviewers []string `mapstructure:"reviewers"`

	// Assignees for the PR (GitHub usernames)
	Assignees []string `mapstructure:"assignees"`
}

type Config struct {
	SkipTests         bool   `mapstructure:"skip_tests"`
	SkipLint          bool   `mapstructure:"skip_lint"`
	AIEngine          string `mapstructure:"ai_engine"`
	Model             string `mapstructure:"model"`
	DryRun            bool   `mapstructure:"dry_run"`
	MaxIterations     int    `mapstructure:"max_iterations"`
	MaxRetries        int    `mapstructure:"max_retries"`
	RetryDelaySeconds int    `mapstructure:"retry_delay"`
	Verbose           bool   `mapstructure:"verbose"`

	BranchPerTask bool   `mapstructure:"branch_per_task"`
	CreatePR      bool   `mapstructure:"create_pr"`
	BaseBranch    string `mapstructure:"base_branch"`
	PRDraft       bool   `mapstructure:"pr_draft"`

	// PR configuration
	PR PRConfig `mapstructure:"pr"`

	// Log file for run history (JSON format)
	LogFile string `mapstructure:"log_file"`

	Parallel    bool `mapstructure:"parallel"`
	MaxParallel int  `mapstructure:"max_parallel"`

	PRDSource   string `mapstructure:"prd_source"`
	PRDFile     string `mapstructure:"prd_file"`
	GitHubRepo  string `mapstructure:"github_repo"`
	GitHubLabel string `mapstructure:"github_label"`
	GitHubToken string `mapstructure:"github_token"`

	// Per-engine model defaults
	Models EngineModels `mapstructure:"models"`

	ConfigFile string `mapstructure:"-"`
}

// ResolvedModel returns the model to use for the current engine.
// Priority: Model field (from --model flag) > Models.<engine> > empty
func (c *Config) ResolvedModel() string {
	if c.Model != "" {
		return c.Model
	}
	switch c.AIEngine {
	case AIEngineClaude:
		return c.Models.Claude
	case AIEngineOpenCode:
		return c.Models.OpenCode
	case AIEngineCodex:
		return c.Models.Codex
	case AIEngineCursor:
		return c.Models.Cursor
	}
	return ""
}

func DefaultConfig() Config {
	return Config{
		SkipTests:         false,
		SkipLint:          false,
		AIEngine:          AIEngineClaude,
		Model:             "",
		DryRun:            false,
		MaxIterations:     0,
		MaxRetries:        3,
		RetryDelaySeconds: 5,
		Verbose:           false,
		BranchPerTask:     false,
		CreatePR:          false,
		BaseBranch:        "",
		PRDraft:           false,
		Parallel:          false,
		MaxParallel:       3,
		PRDSource:         PRDSourceMarkdown,
		PRDFile:           "PRD.md",
		GitHubRepo:        "",
		GitHubLabel:       "",
		GitHubToken:       "",
	}
}

// Load loads configuration from default locations.
func Load() (Config, error) {
	return LoadWithFile("")
}

// LoadWithFile loads configuration, optionally using a specific config file.
// If configFile is provided, it takes highest priority after environment variables.
func LoadWithFile(configFile string) (Config, error) {
	cfg := DefaultConfig()

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvPrefix("RALPHY")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	v.AutomaticEnv()
	_ = v.BindEnv("github_token", "GITHUB_TOKEN", "GH_TOKEN")

	setDefaults(v, cfg)

	// If explicit config file provided, use only that
	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return cfg, fmt.Errorf("reading config file %s: %w", configFile, err)
		}
	} else {
		// Load global config first (lowest priority)
		globalPath := GlobalConfigPath()
		if _, err := os.Stat(globalPath); err == nil {
			v.SetConfigFile(globalPath)
			if err := v.ReadInConfig(); err != nil {
				return cfg, fmt.Errorf("reading global config: %w", err)
			}
		}

		// Load local config (higher priority, merges with global)
		localPath := LocalConfigPath()
		if _, err := os.Stat(localPath); err == nil {
			v.SetConfigFile(localPath)
			if err := v.MergeInConfig(); err != nil {
				return cfg, fmt.Errorf("reading local config: %w", err)
			}
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	cfg.ConfigFile = v.ConfigFileUsed()
	return cfg, nil
}

func BindFlags(cmd *cobra.Command) {
	defaults := DefaultConfig()
	flags := cmd.Flags()

	flags.Bool("no-tests", defaults.SkipTests, "Skip writing and running tests")
	flags.Bool("skip-tests", defaults.SkipTests, "Skip writing and running tests")
	flags.Bool("no-lint", defaults.SkipLint, "Skip linting")
	flags.Bool("skip-lint", defaults.SkipLint, "Skip linting")
	flags.Bool("fast", false, "Skip tests and linting")

	flags.Bool("claude", false, "Use Claude Code")
	flags.Bool("opencode", false, "Use OpenCode")
	flags.Bool("cursor", false, "Use Cursor agent")
	flags.Bool("agent", false, "Use Cursor agent")
	flags.Bool("codex", false, "Use Codex CLI")
	flags.String("model", "", "Model to use (e.g., opus, sonnet, gpt-4o)")

	flags.Bool("dry-run", defaults.DryRun, "Show what would be done without executing")
	flags.Int("max-iterations", defaults.MaxIterations, "Stop after N iterations (0 = unlimited)")
	flags.Int("max-retries", defaults.MaxRetries, "Max retries per task on failure")
	flags.Int("retry-delay", defaults.RetryDelaySeconds, "Seconds between retries")

	flags.Bool("parallel", defaults.Parallel, "Run independent tasks in parallel")
	flags.Int("max-parallel", defaults.MaxParallel, "Max concurrent tasks")

	flags.Bool("branch-per-task", defaults.BranchPerTask, "Create a new git branch for each task")
	flags.String("base-branch", defaults.BaseBranch, "Base branch to create task branches from")
	flags.Bool("create-pr", defaults.CreatePR, "Create a pull request after each task")
	flags.Bool("draft-pr", defaults.PRDraft, "Create PRs as drafts")

	flags.String("prd", defaults.PRDFile, "PRD file path")
	flags.String("yaml", "", "Use YAML task file instead of markdown")
	flags.String("github", defaults.GitHubRepo, "Fetch tasks from GitHub issues (owner/repo)")
	flags.String("github-label", defaults.GitHubLabel, "Filter GitHub issues by label")

	flags.BoolP("verbose", "v", defaults.Verbose, "Show debug output")

	flags.String("log-file", defaults.LogFile, "Write run history to file (JSON format)")
}

func ApplyFlagOverrides(cmd *cobra.Command, cfg *Config) error {
	flags := cmd.Flags()

	if flags.Changed("fast") {
		cfg.SkipTests = true
		cfg.SkipLint = true
	}

	if flags.Changed("no-tests") {
		value, _ := flags.GetBool("no-tests")
		cfg.SkipTests = value
	}

	if flags.Changed("skip-tests") {
		value, _ := flags.GetBool("skip-tests")
		cfg.SkipTests = value
	}

	if flags.Changed("no-lint") {
		value, _ := flags.GetBool("no-lint")
		cfg.SkipLint = value
	}

	if flags.Changed("skip-lint") {
		value, _ := flags.GetBool("skip-lint")
		cfg.SkipLint = value
	}

	engine, err := resolveEngineFlag(flags)
	if err != nil {
		return err
	}
	if engine != "" {
		cfg.AIEngine = engine
	}

	if flags.Changed("model") {
		value, _ := flags.GetString("model")
		cfg.Model = value
	}

	if flags.Changed("dry-run") {
		value, _ := flags.GetBool("dry-run")
		cfg.DryRun = value
	}

	if flags.Changed("max-iterations") {
		value, _ := flags.GetInt("max-iterations")
		cfg.MaxIterations = value
	}

	if flags.Changed("max-retries") {
		value, _ := flags.GetInt("max-retries")
		cfg.MaxRetries = value
	}

	if flags.Changed("retry-delay") {
		value, _ := flags.GetInt("retry-delay")
		cfg.RetryDelaySeconds = value
	}

	if flags.Changed("parallel") {
		value, _ := flags.GetBool("parallel")
		cfg.Parallel = value
	}

	if flags.Changed("max-parallel") {
		value, _ := flags.GetInt("max-parallel")
		cfg.MaxParallel = value
	}

	if flags.Changed("branch-per-task") {
		value, _ := flags.GetBool("branch-per-task")
		cfg.BranchPerTask = value
	}

	if flags.Changed("base-branch") {
		value, _ := flags.GetString("base-branch")
		cfg.BaseBranch = value
	}

	if flags.Changed("create-pr") {
		value, _ := flags.GetBool("create-pr")
		cfg.CreatePR = value
	}

	if flags.Changed("draft-pr") {
		value, _ := flags.GetBool("draft-pr")
		cfg.PRDraft = value
	}

	if flags.Changed("prd") {
		value, _ := flags.GetString("prd")
		cfg.PRDFile = value
		cfg.PRDSource = PRDSourceMarkdown
	}

	if flags.Changed("yaml") {
		value, _ := flags.GetString("yaml")
		if value == "" {
			value = "tasks.yaml"
		}
		cfg.PRDFile = value
		cfg.PRDSource = PRDSourceYAML
	}

	if flags.Changed("github") {
		value, _ := flags.GetString("github")
		cfg.GitHubRepo = value
		cfg.PRDSource = PRDSourceGitHub
	}

	if flags.Changed("github-label") {
		value, _ := flags.GetString("github-label")
		cfg.GitHubLabel = value
	}

	if flags.Changed("verbose") {
		value, _ := flags.GetBool("verbose")
		cfg.Verbose = value
	}

	if flags.Changed("log-file") {
		value, _ := flags.GetString("log-file")
		cfg.LogFile = value
	}

	return nil
}

func resolveEngineFlag(flags *pflag.FlagSet) (string, error) {
	engineFlags := []struct {
		name  string
		value string
	}{
		{name: "claude", value: AIEngineClaude},
		{name: "opencode", value: AIEngineOpenCode},
		{name: "cursor", value: AIEngineCursor},
		{name: "agent", value: AIEngineCursor},
		{name: "codex", value: AIEngineCodex},
	}

	selected := ""
	for _, flag := range engineFlags {
		if !flags.Changed(flag.name) {
			continue
		}
		value, _ := flags.GetBool(flag.name)
		if !value {
			continue
		}
		if selected != "" && selected != flag.value {
			return "", fmt.Errorf("multiple AI engine flags provided")
		}
		selected = flag.value
	}

	return selected, nil
}

func setDefaults(v *viper.Viper, defaults Config) {
	v.SetDefault("skip_tests", defaults.SkipTests)
	v.SetDefault("skip_lint", defaults.SkipLint)
	v.SetDefault("ai_engine", defaults.AIEngine)
	v.SetDefault("model", defaults.Model)
	v.SetDefault("dry_run", defaults.DryRun)
	v.SetDefault("max_iterations", defaults.MaxIterations)
	v.SetDefault("max_retries", defaults.MaxRetries)
	v.SetDefault("retry_delay", defaults.RetryDelaySeconds)
	v.SetDefault("verbose", defaults.Verbose)
	v.SetDefault("branch_per_task", defaults.BranchPerTask)
	v.SetDefault("create_pr", defaults.CreatePR)
	v.SetDefault("base_branch", defaults.BaseBranch)
	v.SetDefault("pr_draft", defaults.PRDraft)
	v.SetDefault("parallel", defaults.Parallel)
	v.SetDefault("max_parallel", defaults.MaxParallel)
	v.SetDefault("prd_source", defaults.PRDSource)
	v.SetDefault("prd_file", defaults.PRDFile)
	v.SetDefault("github_repo", defaults.GitHubRepo)
	v.SetDefault("github_label", defaults.GitHubLabel)
	v.SetDefault("github_token", defaults.GitHubToken)
	v.SetDefault("models.claude", defaults.Models.Claude)
	v.SetDefault("models.opencode", defaults.Models.OpenCode)
	v.SetDefault("models.codex", defaults.Models.Codex)
	v.SetDefault("models.cursor", defaults.Models.Cursor)
}

// SetEngineModel sets the default model for an engine in the specified config file.
func SetEngineModel(configPath, engine, model string) error {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(configPath)

	// Read existing config if it exists
	if _, err := os.Stat(configPath); err == nil {
		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("reading config: %w", err)
		}
	}

	// Set the model for the engine
	v.Set("models."+engine, model)

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Write config
	if err := v.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

// SetDefaultEngine sets the default AI engine in the specified config file.
func SetDefaultEngine(configPath, engine string) error {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(configPath)

	// Read existing config if it exists
	if _, err := os.Stat(configPath); err == nil {
		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("reading config: %w", err)
		}
	}

	v.Set("ai_engine", engine)

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	if err := v.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

// WriteDefaultConfig writes a default config file to the specified path.
func WriteDefaultConfig(configPath string) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	content := `# Ralphy configuration
# This file can be placed at:
#   - ~/.config/ralphy/ralphy.yaml (global)
#   - ./ralphy.yaml (project-local, overrides global)

# Default AI engine: claude, opencode, codex, cursor
ai_engine: claude

# Per-engine model defaults
# These are used when no --model flag is provided
models:
  claude: ""
  opencode: ""
  codex: ""
  cursor: ""

# Task source
prd_source: markdown
prd_file: PRD.md

# Workflow
skip_tests: false
skip_lint: false
dry_run: false
max_iterations: 0
max_retries: 3
retry_delay: 5
verbose: false

# Parallel execution
parallel: false
max_parallel: 3

# Git workflow
branch_per_task: false
base_branch: ""
create_pr: false
pr_draft: false

# GitHub integration
github_repo: ""
github_label: ""
# Supports GITHUB_TOKEN, GH_TOKEN env vars, or explicit value here
github_token: ""
`
	return os.WriteFile(configPath, []byte(content), 0o644)
}

// ConfigExists checks if a config file exists at the given path.
func ConfigExists(configPath string) bool {
	_, err := os.Stat(configPath)
	return err == nil
}
