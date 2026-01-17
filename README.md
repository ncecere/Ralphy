# Ralphy

Autonomous AI coding loop that runs AI assistants to work through tasks until a PRD is complete.

## Features

- **Multiple AI Engines**: Support for Claude Code, OpenCode, Codex, and Cursor
- **Flexible Task Sources**: Read tasks from Markdown PRDs, YAML files, or GitHub issues
- **Git Integration**: Automatic branch creation and PR workflows
- **Parallel Execution**: Run independent tasks concurrently
- **Retry Logic**: Automatic retries with configurable delays
- **Desktop Notifications**: Get notified when tasks complete

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

1. Create a `PRD.md` file with tasks:

```markdown
# My Project PRD

## Tasks

- [ ] Implement user authentication
- [ ] Add database migrations
- [ ] Create API endpoints
- [x] Set up project structure (completed)
```

2. Run ralphy:

```bash
ralphy --claude
```

## Usage

```bash
ralphy [flags]
```

### AI Engine Flags

| Flag | Description |
|------|-------------|
| `--claude` | Use Claude Code |
| `--opencode` | Use OpenCode |
| `--codex` | Use Codex CLI |
| `--cursor` | Use Cursor agent |
| `--agent` | Alias for `--cursor` |

### Task Source Flags

| Flag | Description |
|------|-------------|
| `--prd <file>` | PRD markdown file (default: `PRD.md`) |
| `--yaml <file>` | Use YAML task file instead of markdown |
| `--github <owner/repo>` | Fetch tasks from GitHub issues |
| `--github-label <label>` | Filter GitHub issues by label |

### Workflow Flags

| Flag | Description |
|------|-------------|
| `--dry-run` | Show what would be done without executing |
| `--fast` | Skip tests and linting |
| `--no-tests`, `--skip-tests` | Skip writing and running tests |
| `--no-lint`, `--skip-lint` | Skip linting |
| `--max-iterations <n>` | Stop after N iterations (0 = unlimited) |
| `--max-retries <n>` | Max retries per task on failure (default: 3) |
| `--retry-delay <n>` | Seconds between retries (default: 5) |
| `-v`, `--verbose` | Show debug output |

### Parallel Execution Flags

| Flag | Description |
|------|-------------|
| `--parallel` | Run independent tasks in parallel |
| `--max-parallel <n>` | Max concurrent tasks (default: 3) |

### Git Workflow Flags

| Flag | Description |
|------|-------------|
| `--branch-per-task` | Create a new git branch for each task |
| `--base-branch <branch>` | Base branch to create task branches from |
| `--create-pr` | Create a pull request after each task |
| `--draft-pr` | Create PRs as drafts |

## Configuration

Ralphy supports configuration via:
1. Command-line flags (highest priority)
2. Environment variables
3. Configuration file (lowest priority)

### Configuration File

Create a `ralphy.yaml` in your project root:

```yaml
# AI engine: claude, opencode, cursor, codex
ai_engine: claude

# Task source
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
base_branch: main
create_pr: false
pr_draft: false

# GitHub integration
github_repo: ""
github_label: ""
github_token: ""  # or use GITHUB_TOKEN/GH_TOKEN env var
```

### Environment Variables

All configuration options can be set via environment variables with the `RALPHY_` prefix:

```bash
export RALPHY_AI_ENGINE=claude
export RALPHY_PRD_FILE=tasks.md
export RALPHY_MAX_RETRIES=5
export GITHUB_TOKEN=ghp_xxxx  # or GH_TOKEN
```

## Task Formats

See the [examples/](examples/) directory for complete example files.

### Markdown (PRD.md)

```markdown
# Project PRD

## Tasks

- [ ] First task to complete
- [ ] Second task to complete
- [x] Already completed task
```

See [examples/PRD.md](examples/PRD.md) for a full example.

### YAML

```yaml
tasks:
  - title: First task
    status: pending
  - title: Second task
    status: pending
  - title: Completed task
    status: done
```

The YAML format supports additional metadata like `priority`, `tags`, `notes`, and `depends_on`.
See [examples/tasks.yaml](examples/tasks.yaml) for a full example.

### GitHub Issues

Tasks are fetched from open issues, ordered oldest-first:

```bash
ralphy --github owner/repo --github-label enhancement --claude
```

## Examples

### Basic Usage

```bash
# Run with Claude Code on PRD.md
ralphy --claude

# Run with OpenCode on a custom file
ralphy --opencode --prd tasks.md

# Dry run to see what would happen
ralphy --claude --dry-run
```

### Git Workflow

```bash
# Create a branch for each task and open PRs
ralphy --claude --branch-per-task --create-pr --base-branch main

# Create draft PRs
ralphy --claude --branch-per-task --create-pr --draft-pr
```

### Parallel Execution

```bash
# Run up to 3 tasks in parallel
ralphy --claude --parallel --max-parallel 3
```

### GitHub Issues

```bash
# Work through GitHub issues labeled "ai-task"
ralphy --claude --github myorg/myrepo --github-label ai-task --create-pr
```

## Requirements

- Go 1.21+ (for building)
- One of the supported AI CLI tools installed:
  - [Claude Code](https://claude.ai/code)
  - [OpenCode](https://github.com/opencode-ai/opencode)
  - [Codex](https://github.com/openai/codex)
  - [Cursor](https://cursor.sh/)
- Git (for branch/PR workflows)
- GitHub CLI (`gh`) for PR creation

## License

MIT

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
