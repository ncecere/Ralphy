# Ralphy

Autonomous AI coding loop that runs AI assistants (Claude Code, OpenCode, Codex, Cursor) to work through tasks until a PRD is complete.

## Features

- **Multiple AI Engines**: Support for Claude Code, OpenCode, Codex, and Cursor
- **Flexible Task Sources**: Read tasks from Markdown PRDs, YAML files, or GitHub issues
- **Git Integration**: Automatic branch creation and PR workflows
- **Parallel Execution**: Run independent tasks concurrently with git worktrees
- **Retry Logic**: Automatic retries with configurable delays
- **Desktop Notifications**: Get notified when tasks complete
- **Activity Monitoring**: Live heartbeat indicator shows engine activity
- **Per-Engine Model Defaults**: Configure default models for each AI engine
- **Developer Tools**: Doctor command, config validation, task preview

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Commands](#commands)
  - [Main Command](#main-command-ralphy)
  - [Initialize Config](#initialize-config-ralphy-init)
  - [List Models](#list-models-ralphy-models)
  - [Doctor](#check-dependencies-ralphy-doctor)
  - [Config Management](#config-management-ralphy-config)
  - [Task Preview](#task-preview-ralphy-tasks)
- [Configuration](#configuration)
  - [Config Precedence](#config-precedence)
  - [Config File Reference](#config-file-reference)
  - [Environment Variables](#environment-variables)
  - [GitHub Authentication](#github-authentication)
- [Task Formats](#task-formats)
  - [Markdown PRD](#markdown-prd)
  - [YAML Tasks](#yaml-tasks)
  - [GitHub Issues](#github-issues)
- [Examples](#examples)
- [Requirements](#requirements)
- [License](#license)

## Installation

### Download Binary (Recommended)

Download the latest release for your platform from [GitHub Releases](https://github.com/ncecere/ralphy/releases/latest).

```bash
# macOS (Apple Silicon)
curl -L https://github.com/ncecere/ralphy/releases/latest/download/ralphy_darwin_arm64.tar.gz | tar xz
sudo mv ralphy /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/ncecere/ralphy/releases/latest/download/ralphy_darwin_amd64.tar.gz | tar xz
sudo mv ralphy /usr/local/bin/

# Linux (amd64)
curl -L https://github.com/ncecere/ralphy/releases/latest/download/ralphy_linux_amd64.tar.gz | tar xz
sudo mv ralphy /usr/local/bin/

# Linux (arm64)
curl -L https://github.com/ncecere/ralphy/releases/latest/download/ralphy_linux_arm64.tar.gz | tar xz
sudo mv ralphy /usr/local/bin/
```

### Go Install

```bash
go install github.com/ncecere/ralphy/cmd/ralphy@latest
```

### Build from Source

```bash
git clone https://github.com/ncecere/ralphy.git
cd ralphy
make build
sudo cp build/ralphy /usr/local/bin/
```

## Quick Start

1. **Create a PRD file** with tasks:

```markdown
# My Project PRD

## Tasks

- [ ] Implement user authentication
- [ ] Add database migrations
- [ ] Create API endpoints
- [x] Set up project structure (completed)
```

2. **Initialize config** (optional but recommended):

```bash
# Create global config (~/.config/ralphy/ralphy.yaml)
ralphy init

# Or create project-local config (./ralphy.yaml)
ralphy init --local

# Set your preferred engine and model
ralphy init --engine claude --model sonnet
```

3. **Run ralphy**:

```bash
# Run with Claude Code
ralphy --claude

# Or if you set a default engine in config
ralphy
```

## Commands

### Main Command: `ralphy`

Run the autonomous AI coding loop.

```bash
ralphy [flags]
```

#### AI Engine Flags

| Flag | Description |
|------|-------------|
| `--claude` | Use Claude Code |
| `--opencode` | Use OpenCode |
| `--codex` | Use Codex CLI |
| `--cursor` | Use Cursor agent |
| `--agent` | Alias for `--cursor` |
| `--model <name>` | Model to use (overrides config default) |
| `--config <file>` | Path to config file (overrides default locations) |

#### Model Examples

```bash
# Claude Code - use alias or full model name
ralphy --claude --model opus
ralphy --claude --model sonnet
ralphy --claude --model haiku
ralphy --claude --model claude-sonnet-4-5-20250929

# OpenCode - use provider/model format
ralphy --opencode --model anthropic/claude-sonnet-4-5
ralphy --opencode --model openai/gpt-4o

# Codex
ralphy --codex --model o4-mini
```

#### Task Source Flags

| Flag | Description |
|------|-------------|
| `--prd <file>` | PRD markdown file (default: `PRD.md`) |
| `--yaml <file>` | Use YAML task file instead of markdown |
| `--github <owner/repo>` | Fetch tasks from GitHub issues |
| `--github-label <label>` | Filter GitHub issues by label |

#### Workflow Flags

| Flag | Description |
|------|-------------|
| `--dry-run` | Show what would be done without executing |
| `--fast` | Skip tests and linting (equivalent to `--no-tests --no-lint`) |
| `--no-tests`, `--skip-tests` | Skip writing and running tests |
| `--no-lint`, `--skip-lint` | Skip linting |
| `--max-iterations <n>` | Stop after N iterations (0 = unlimited) |
| `--max-retries <n>` | Max retries per task on failure (default: 3) |
| `--retry-delay <n>` | Seconds between retries (default: 5) |
| `-v`, `--verbose` | Show debug output |

#### Parallel Execution Flags

| Flag | Description |
|------|-------------|
| `--parallel` | Run independent tasks in parallel using git worktrees |
| `--max-parallel <n>` | Max concurrent tasks (default: 3) |

When running in parallel mode, Ralphy:
1. Groups tasks by `parallel_group` (YAML) or runs all concurrently (Markdown/GitHub)
2. Creates separate git worktrees for each concurrent task
3. Merges completed branches back to the base branch
4. Cleans up worktrees after completion

#### Git Workflow Flags

| Flag | Description |
|------|-------------|
| `--branch-per-task` | Create a new git branch for each task |
| `--base-branch <branch>` | Base branch to create task branches from |
| `--create-pr` | Create a pull request after each task |
| `--draft-pr` | Create PRs as drafts |

---

### Initialize Config: `ralphy init`

Create a Ralphy configuration file.

```bash
ralphy init [flags]
```

#### Flags

| Flag | Description |
|------|-------------|
| `--local` | Create config in current directory (`./ralphy.yaml`) instead of global |
| `--force` | Overwrite existing config file |
| `--engine <name>` | Set default AI engine (`claude`, `opencode`, `codex`, `cursor`) |
| `--model <name>` | Set default model for the engine (requires `--engine`) |

#### Examples

```bash
# Create global config at ~/.config/ralphy/ralphy.yaml
ralphy init

# Create project-local config at ./ralphy.yaml
ralphy init --local

# Create config with defaults already set
ralphy init --engine claude --model sonnet

# Create local config for OpenCode project
ralphy init --local --engine opencode --model anthropic/claude-sonnet-4-5

# Overwrite existing config
ralphy init --force
```

---

### List Models: `ralphy models`

List available models for an AI engine and optionally set a default.

```bash
ralphy models [flags]
```

#### Flags

| Flag | Description |
|------|-------------|
| `--claude` | List Claude Code models |
| `--opencode` | List OpenCode models (fetched live from CLI) |
| `--codex` | List Codex models |
| `--cursor` | List Cursor models |
| `--set-default` | Interactively select and save a default model |
| `--model <name>` | Model to set (use with `--set-default` for non-interactive) |
| `--local` | Save to local config instead of global |

#### Examples

```bash
# List available Claude models (aliases)
ralphy models --claude
# Output:
# opus
# sonnet
# haiku

# List available OpenCode models (live from CLI)
ralphy models --opencode

# Interactively select and save a default model
ralphy models --opencode --set-default
# Shows numbered list, prompts for selection

# Non-interactively set a default model
ralphy models --claude --set-default --model sonnet

# Save to local project config
ralphy models --opencode --set-default --model anthropic/claude-sonnet-4-5 --local
```

#### Model Sources

| Engine | Source |
|--------|--------|
| Claude Code | Built-in aliases: `opus`, `sonnet`, `haiku` |
| OpenCode | Live from `opencode models` CLI command |
| Codex | Not available (use `codex --help`) |
| Cursor | Not available (check Cursor settings) |

---

### Check Dependencies: `ralphy doctor`

Verify that all required dependencies are installed and configured correctly.

```bash
ralphy doctor
```

Checks performed:
- **AI Engines**: Claude Code, OpenCode, Codex, Cursor CLI availability
- **Git**: Installation and user configuration
- **GitHub**: CLI installation, authentication status, token availability
- **Configuration**: Global and local config file presence

#### Example Output

```
Ralphy Doctor
=============

AI Engines:
  ✓ Claude Code          2.1.7 (/opt/homebrew/bin/claude)
  ✓ OpenCode             1.1.25 (/opt/homebrew/bin/opencode)
  ✓ Codex                codex-cli 0.63.0 (/opt/homebrew/bin/codex)
  ! Cursor               not found in PATH

Git:
  ✓ Git                  git version 2.48.1 (/usr/bin/git)
  ✓ Git config           Your Name <you@example.com>

GitHub:
  ✓ GitHub CLI           gh version 2.40.0 (/usr/bin/gh)
  ✓ GitHub CLI auth      Logged in to github.com
  ✓ GitHub token         found in environment (ghp_...xxxx)

Configuration:
  ✓ Global config        /home/user/.config/ralphy/ralphy.yaml
  ✓ Local config         ralphy.yaml

Summary:
  10 passed, 1 warnings, 0 errors
```

---

### Config Management: `ralphy config`

View and validate configuration.

#### Show Configuration

Display the effective configuration after merging all sources:

```bash
# Show merged configuration
ralphy config show

# Show with source annotations (env/local/global/default)
ralphy config show --sources

# Output as YAML
ralphy config show --yaml
```

#### Validate Configuration

Check configuration files for errors:

```bash
ralphy config validate
```

Validates:
- YAML syntax
- Known configuration keys (warns on unknown keys)
- Valid values for enums (`ai_engine`, `prd_source`)
- File paths exist (when applicable)

#### Example: Config Show with Sources

```
Configuration with Sources
==========================

AI Engine:
  ai_engine:           claude               [local]
  models.claude:       sonnet               [local]
  models.opencode:                          [default]

Task Source:
  prd_source:          markdown             [local]
  prd_file:            PRD.md               [local]

Workflow:
  skip_tests:          false                [default]
  max_retries:         3                    [global]
  verbose:             true                 [env]
```

---

### Task Preview: `ralphy tasks`

List and preview tasks without running them.

#### List Tasks

```bash
# List all tasks from default source
ralphy tasks list

# List tasks from a specific file
ralphy tasks list --prd myproject.md
ralphy tasks list --yaml tasks.yaml

# List only pending or completed tasks
ralphy tasks list --pending
ralphy tasks list --completed

# Show only counts
ralphy tasks list --count
```

#### Show Next Task

```bash
# Show what task would run next
ralphy tasks next

# With a specific source
ralphy tasks next --yaml tasks.yaml
```

#### Example Output

```
Tasks from examples/tasks.yaml
===============================

Pending:
  1. [ ] Initialize Go module and project structure [group 1]
  2. [ ] Set up configuration management with Viper [group 1]
  3. [ ] Create database connection pool [group 1]
  4. [ ] Implement user authentication [group 2]

Completed:
  1. [x] Research best practices
  2. [x] Design database schema

Summary: 4 pending, 2 completed
```

---

## Configuration

### Config Precedence

Ralphy loads configuration from multiple sources. Higher priority sources override lower ones:

1. **Command-line flags** (highest priority)
2. **Environment variables** (`RALPHY_*`)
3. **Local config** (`./ralphy.yaml` in current directory)
4. **Global config** (`~/.config/ralphy/ralphy.yaml`)

This allows you to:
- Set global defaults in `~/.config/ralphy/ralphy.yaml`
- Override per-project in `./ralphy.yaml`
- Override for a single run with flags or env vars

### Config File Reference

```yaml
# Ralphy configuration
# Location: ~/.config/ralphy/ralphy.yaml (global)
#       or: ./ralphy.yaml (project-local)

# ==============================================================================
# AI Engine Settings
# ==============================================================================

# Default AI engine: claude, opencode, codex, cursor
ai_engine: claude

# Per-engine model defaults
# These are used when no --model flag is provided
# Use `ralphy models --<engine>` to list available models
# Use `ralphy models --<engine> --set-default` to set interactively
models:
  claude: sonnet           # opus, sonnet, haiku, or full model name
  opencode: ""             # e.g., anthropic/claude-sonnet-4-5, openai/gpt-4o
  codex: ""                # e.g., o4-mini
  cursor: ""

# ==============================================================================
# Task Source Settings
# ==============================================================================

# Task source type: markdown, yaml, github
prd_source: markdown

# Path to task file (for markdown and yaml sources)
prd_file: PRD.md

# GitHub integration (for github source)
github_repo: ""            # e.g., owner/repo
github_label: ""           # Filter issues by label
github_token: ""           # Or use GITHUB_TOKEN/GH_TOKEN env var

# ==============================================================================
# Workflow Settings
# ==============================================================================

# Skip tests and linting
skip_tests: false
skip_lint: false

# Preview mode - show what would be done without executing
dry_run: false

# Stop after N iterations (0 = unlimited)
max_iterations: 0

# Retry settings for failed tasks
max_retries: 3             # Max retries per task
retry_delay: 5             # Seconds between retries

# Debug output
verbose: false

# ==============================================================================
# Parallel Execution
# ==============================================================================

# Enable parallel task execution using git worktrees
parallel: false

# Maximum concurrent tasks
max_parallel: 3

# ==============================================================================
# Git Workflow
# ==============================================================================

# Create a new branch for each task
branch_per_task: false

# Base branch to create task branches from (empty = current branch)
base_branch: ""

# Automatically create pull requests after tasks complete
create_pr: false

# Create PRs as drafts
pr_draft: false
```

### Environment Variables

All configuration options can be set via environment variables with the `RALPHY_` prefix:

| Environment Variable | Config Key | Description |
|---------------------|------------|-------------|
| `RALPHY_AI_ENGINE` | `ai_engine` | Default AI engine |
| `RALPHY_MODEL` | `model` | Model override (applies to any engine) |
| `RALPHY_PRD_SOURCE` | `prd_source` | Task source type |
| `RALPHY_PRD_FILE` | `prd_file` | Path to task file |
| `RALPHY_SKIP_TESTS` | `skip_tests` | Skip tests (`true`/`false`) |
| `RALPHY_SKIP_LINT` | `skip_lint` | Skip linting (`true`/`false`) |
| `RALPHY_DRY_RUN` | `dry_run` | Preview mode (`true`/`false`) |
| `RALPHY_MAX_ITERATIONS` | `max_iterations` | Max iterations (0 = unlimited) |
| `RALPHY_MAX_RETRIES` | `max_retries` | Max retries per task |
| `RALPHY_RETRY_DELAY` | `retry_delay` | Seconds between retries |
| `RALPHY_VERBOSE` | `verbose` | Debug output (`true`/`false`) |
| `RALPHY_PARALLEL` | `parallel` | Parallel execution (`true`/`false`) |
| `RALPHY_MAX_PARALLEL` | `max_parallel` | Max concurrent tasks |
| `RALPHY_BRANCH_PER_TASK` | `branch_per_task` | Branch per task (`true`/`false`) |
| `RALPHY_BASE_BRANCH` | `base_branch` | Base branch for task branches |
| `RALPHY_CREATE_PR` | `create_pr` | Create PRs (`true`/`false`) |
| `RALPHY_PR_DRAFT` | `pr_draft` | Draft PRs (`true`/`false`) |
| `RALPHY_GITHUB_REPO` | `github_repo` | GitHub repo (`owner/repo`) |
| `RALPHY_GITHUB_LABEL` | `github_label` | GitHub issue label filter |
| `GITHUB_TOKEN` | `github_token` | GitHub authentication token |
| `GH_TOKEN` | `github_token` | GitHub CLI token (fallback) |

### GitHub Authentication

For GitHub integration (fetching issues or creating PRs), Ralphy needs a GitHub token. It checks these sources in order:

1. `GITHUB_TOKEN` environment variable (recommended)
2. `GH_TOKEN` environment variable (used by GitHub CLI)
3. `github_token` in config file (not recommended for shared repos)

```bash
# Option 1: Export token (recommended)
export GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxx

# Option 2: Use GitHub CLI's token
# If you've authenticated with `gh auth login`, Ralphy will use GH_TOKEN

# Option 3: Add to config file (not recommended)
# github_token: ghp_xxxxxxxxxxxxxxxxxxxx
```

#### Required Token Permissions

| Feature | Required Scopes |
|---------|-----------------|
| Fetch issues | `repo` (private) or `public_repo` (public) |
| Create PRs | `repo` |
| Close issues | `repo` |

Create a token at [GitHub Settings > Developer settings > Personal access tokens](https://github.com/settings/tokens).

---

## Task Formats

### Markdown PRD

The simplest format. Tasks are markdown checkboxes:

```markdown
# Project PRD

## Overview

Description of the project.

## Tasks

### Phase 1: Setup

- [ ] Initialize project structure
- [ ] Set up configuration management
- [ ] Create database connection

### Phase 2: Features

- [ ] Implement user authentication
- [ ] Create API endpoints
- [ ] Add validation

## Completed

- [x] Research best practices
- [x] Design database schema
```

**Rules:**
- `- [ ]` = pending task (will be worked on)
- `- [x]` = completed task (skipped)
- Tasks are processed in order (top to bottom)
- Ralphy marks tasks complete by changing `[ ]` to `[x]`

See [examples/PRD.md](examples/PRD.md) for a full example.

### YAML Tasks

More structured format with metadata and parallel groups:

```yaml
project: my-api
version: "1.0"

tasks:
  # Group 1: Setup (run in parallel)
  - title: Initialize project structure
    completed: false
    parallel_group: 1
    tags: [setup]

  - title: Set up configuration
    completed: false
    parallel_group: 1
    tags: [setup, config]

  # Group 2: Features (run after group 1)
  - title: Implement authentication
    completed: false
    parallel_group: 2
    tags: [auth, api]
    acceptance_criteria:
      - Validate email format
      - Hash passwords with bcrypt

  - title: Create API endpoints
    completed: false
    parallel_group: 2
    tags: [api]
    subtasks:
      - GET /api/items
      - POST /api/items
      - DELETE /api/items/:id

metadata:
  created_at: "2024-01-15"
  owner: developer@example.com
```

**Supported fields:**
| Field | Required | Description |
|-------|----------|-------------|
| `title` | Yes | Task description (sent to AI) |
| `completed` | No | `true`/`false` - whether task is done |
| `parallel_group` | No | Integer - tasks with same group run concurrently |
| `tags` | No | Array of tags (for documentation) |
| `notes` | No | Additional notes (for documentation) |
| `subtasks` | No | List of subtasks (for documentation) |
| `acceptance_criteria` | No | List of acceptance criteria (for documentation) |
| `depends_on` | No | Dependencies (for documentation) |
| `priority` | No | Priority level (for documentation) |

**Parallel Groups:**
- Tasks with the same `parallel_group` number run concurrently
- Groups are processed in numerical order (1, then 2, then 3...)
- Tasks without a group run sequentially

See [examples/tasks.yaml](examples/tasks.yaml) for a full example.

### GitHub Issues

Fetch tasks from GitHub issues:

```bash
# All open issues, oldest first
ralphy --github owner/repo --claude

# Filter by label
ralphy --github owner/repo --github-label ai-task --claude
```

**Behavior:**
- Open issues are fetched, ordered by creation date (oldest first)
- Issue title becomes the task description
- Issue body is included as context
- When a task completes, the issue is closed automatically
- Use `--github-label` to filter to specific issues

---

## Examples

### Basic Usage

```bash
# Run with Claude Code on default PRD.md
ralphy --claude

# Run with OpenCode on a custom file
ralphy --opencode --prd tasks.md

# Run with a specific model
ralphy --claude --model opus

# Dry run to see what would happen
ralphy --claude --dry-run

# Skip tests and linting for faster iteration
ralphy --claude --fast

# Verbose output for debugging
ralphy --claude -v
```

### Using YAML Tasks

```bash
# Run with YAML task file
ralphy --claude --yaml tasks.yaml

# Parallel execution with YAML groups
ralphy --claude --yaml tasks.yaml --parallel
```

### Git Workflow

```bash
# Create a branch for each task
ralphy --claude --branch-per-task

# Create branches and PRs
ralphy --claude --branch-per-task --create-pr

# Create draft PRs from a specific base branch
ralphy --claude --branch-per-task --create-pr --draft-pr --base-branch develop
```

### Parallel Execution

```bash
# Run up to 3 tasks in parallel
ralphy --claude --parallel

# Run up to 5 tasks in parallel
ralphy --claude --parallel --max-parallel 5

# Parallel with branch-per-task and PRs
ralphy --claude --parallel --branch-per-task --create-pr
```

### GitHub Issues

```bash
# Work through all open issues
ralphy --claude --github myorg/myrepo

# Filter to issues labeled "ai-task"
ralphy --claude --github myorg/myrepo --github-label ai-task

# Create PRs for each issue
ralphy --claude --github myorg/myrepo --github-label ai-task --create-pr
```

### Configuration Workflows

```bash
# Set up global defaults once
ralphy init --engine claude --model sonnet

# Then just run with defaults
ralphy

# Override model for a specific run
ralphy --model opus

# Set up project-specific config
cd my-project
ralphy init --local --engine opencode --model anthropic/claude-sonnet-4-5

# List and set models interactively
ralphy models --opencode --set-default
```

---

## Requirements

- **Go 1.21+** (for building from source)
- **One of the supported AI CLI tools:**
  - [Claude Code](https://claude.ai/code) - `claude` CLI
  - [OpenCode](https://github.com/opencode-ai/opencode) - `opencode` CLI
  - [Codex](https://github.com/openai/codex) - `codex` CLI
  - [Cursor](https://cursor.sh/) - `cursor` CLI
- **Git** (for branch/PR workflows)
- **GitHub CLI** (`gh`) for PR creation

---

## Activity Indicator

While running, Ralphy displays a spinner with an activity heartbeat:

```
⠋ Working on task: Implement authentication  ● active
```

- `● active` (green) - AI engine is actively outputting
- `○ 5s ago` - Time since last output (helps identify stalls)

---

## License

MIT

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
