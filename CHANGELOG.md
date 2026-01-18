# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.2] - 2025-01-18

### Added

- **PR Templates** - Customizable PR body with template variables
  - Variables: `{{.Task}}`, `{{.Engine}}`, `{{.Model}}`, `{{.InputTokens}}`, `{{.OutputTokens}}`, `{{.Cost}}`, `{{.Duration}}`, `{{.FilesChanged}}`
  - Default template shows task details in a formatted table
  - Custom templates via `pr.body_template` in config

- **PR Metadata** - Configure labels, reviewers, and assignees
  ```yaml
  pr:
    labels: ["ai-generated", "needs-review"]
    reviewers: ["teammate1", "teammate2"]
    assignees: ["myself"]
  ```

- **Run History Logging** - Track run metrics over time
  - `--log-file <path>` flag to write run history (JSON/NDJSON format)
  - Logs task status, tokens, cost, duration, retries, PR URLs
  - Aggregate summary per run

### Changed

- PR creation now includes richer metadata from task execution
- Task execution tracks more detailed metrics for logging

## [0.4.1] - 2025-01-18

### Added

- **`ralphy setup` command** - Interactive setup wizard
  - Step-by-step configuration with prompts
  - Shows which AI engines are installed (✓/○ indicators)
  - Lets you select default engine and model
  - Configure task source (markdown/yaml/github)
  - Set up Git workflow options (branch-per-task, PRs)
  - Configure parallel execution settings
  - `--local` flag to create project config instead of global
  - Summary screen before saving
  - Shows next steps after completion

## [0.4.0] - 2025-01-18

### Added

- **`ralphy doctor` command** - Verify system dependencies and configuration
  - Checks AI engine CLIs (claude, opencode, codex, cursor)
  - Checks Git installation and user configuration
  - Checks GitHub CLI and authentication status
  - Checks GitHub token availability
  - Checks global and local config file presence
  - Color-coded output with pass/warn/error status

- **`ralphy config` command** - Manage and inspect configuration
  - `ralphy config show` - Display effective merged configuration
  - `ralphy config show --sources` - Show source of each value (env/local/global/default)
  - `ralphy config show --yaml` - Output configuration as YAML
  - `ralphy config validate` - Validate configuration files
    - Checks YAML syntax
    - Warns on unknown keys
    - Validates enum values (ai_engine, prd_source)
    - Checks file paths exist

- **`ralphy tasks` command** - Preview tasks without running
  - `ralphy tasks list` - List all tasks with status
  - `ralphy tasks list --pending` - Show only pending tasks
  - `ralphy tasks list --completed` - Show only completed tasks
  - `ralphy tasks list --count` - Show task counts only
  - `ralphy tasks next` - Show the next task that would execute
  - Supports `--prd`, `--yaml`, `--github` flags to specify source

- **`--config` flag** - Specify alternate config file path
  - Works with all commands
  - Overrides default config file locations
  - Example: `ralphy --config /path/to/config.yaml --claude`

## [0.3.2] - 2025-01-18

### Added

- **`ralphy init` command** - Initialize configuration files
  - Creates global config at `~/.config/ralphy/ralphy.yaml` by default
  - `--local` flag creates project config at `./ralphy.yaml`
  - `--engine` and `--model` flags to set defaults during initialization
  - `--force` flag to overwrite existing config files

- **`ralphy models` command** - List and set default models for AI engines
  - `ralphy models --claude` - Shows Claude aliases: opus, sonnet, haiku
  - `ralphy models --opencode` - Lists models live from `opencode models` CLI
  - `ralphy models --codex` / `--cursor` - Notes that model listing is not available
  - `--set-default` flag for interactive model selection with numbered list
  - `--model` with `--set-default` for non-interactive model setting
  - `--local` flag to save to project config instead of global

- **Per-engine model defaults** in configuration
  ```yaml
  models:
    claude: sonnet
    opencode: anthropic/claude-sonnet-4-5
    codex: ""
    cursor: ""
  ```

- **Config precedence system** - Settings are loaded in priority order:
  1. Command-line flags (highest)
  2. Environment variables
  3. Local config (`./ralphy.yaml`)
  4. Global config (`~/.config/ralphy/ralphy.yaml`)

### Changed

- Model resolution now checks per-engine defaults when no `--model` flag is provided
- Documentation significantly expanded with full command reference

## [0.3.1] - 2025-01-17

### Fixed

- Sequential mode now auto-marks YAML and Markdown tasks as complete after each task
- Previously only GitHub issues were being marked complete, causing sequential runs to repeat the same task indefinitely

## [0.3.0] - 2025-01-17

### Added

- **Activity heartbeat indicator** in the task spinner
  - Shows `● active` (green dot) when the AI engine is actively outputting
  - Shows `○ Xs ago` when idle, displaying seconds since last output
  - Helps monitor long-running tasks without enabling verbose logging
  - Useful for detecting if the AI engine has stalled

## [0.2.0] - 2025-01-17

### Added

- **`--model` flag** to specify which model to use with each AI engine
- `model` configuration option in `ralphy.yaml`
- `RALPHY_MODEL` environment variable support

### Examples

```bash
# Claude Code - use alias or full model name
ralphy --claude --model opus
ralphy --claude --model sonnet
ralphy --claude --model claude-sonnet-4-5-20250929

# OpenCode - use provider/model format
ralphy --opencode --model anthropic/claude-sonnet-4-5
ralphy --opencode --model openai/gpt-4o

# Codex
ralphy --codex --model o4-mini
```

## [0.1.0] - 2025-01-17

### Added

- **Initial release** of Ralphy Go CLI

#### AI Engine Support
- Claude Code (`--claude`)
- OpenCode (`--opencode`)
- Codex CLI (`--codex`)
- Cursor (`--cursor`, `--agent`)

#### Task Sources
- Markdown PRD files with checkbox syntax (`- [ ]` / `- [x]`)
- YAML task files with structured metadata and parallel groups
- GitHub issues with optional label filtering

#### Git Integration
- Branch-per-task workflow (`--branch-per-task`)
- Automatic PR creation (`--create-pr`)
- Draft PR support (`--draft-pr`)
- Configurable base branch (`--base-branch`)

#### Parallel Execution
- Run independent tasks concurrently (`--parallel`)
- Configurable concurrency limit (`--max-parallel`)
- Git worktree support for parallel branches
- Automatic branch merging and cleanup

#### Reliability Features
- Automatic retries on failure (`--max-retries`, `--retry-delay`)
- Desktop notifications on task completion
- Dry-run mode for previewing actions (`--dry-run`)
- Verbose output for debugging (`--verbose`)

#### Configuration
- YAML config file support (`ralphy.yaml`)
- Environment variable support (`RALPHY_*`)
- GitHub token support via `GITHUB_TOKEN`, `GH_TOKEN`, or config

### Documentation

- README with installation and usage instructions
- Example PRD.md file
- Example tasks.yaml file with parallel groups
- Configuration file reference

[Unreleased]: https://github.com/ncecere/ralphy/compare/v0.4.2...HEAD
[0.4.2]: https://github.com/ncecere/ralphy/compare/v0.4.1...v0.4.2
[0.4.1]: https://github.com/ncecere/ralphy/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/ncecere/ralphy/compare/v0.3.2...v0.4.0
[0.3.2]: https://github.com/ncecere/ralphy/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/ncecere/ralphy/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/ncecere/ralphy/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/ncecere/ralphy/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/ncecere/ralphy/releases/tag/v0.1.0
