# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.2] - 2025-01-18

### Added

- `ralphy init` command to create config files
  - `--local` flag for project config (./ralphy.yaml)
  - Default creates global config (~/.config/ralphy/ralphy.yaml)
  - `--engine` and `--model` flags to set defaults during init

- `ralphy models` command to list and set default models
  - `ralphy models --opencode` - lists models from OpenCode CLI
  - `ralphy models --claude` - shows Claude aliases (opus, sonnet, haiku)
  - `--set-default` for interactive model selection
  - `--model` with `--set-default` for non-interactive setting
  - `--local` to save to project config instead of global

- Per-engine model defaults in config
  ```yaml
  models:
    claude: sonnet
    opencode: anthropic/claude-sonnet-4-5
  ```

- Config precedence: flags > env > local config > global config

## [0.3.1] - 2025-01-17

### Fixed

- Sequential mode now auto-marks YAML/Markdown tasks as complete after each task
- Previously only GitHub issues were auto-marked, causing sequential runs to repeat the same task

## [0.3.0] - 2025-01-17

### Added

- Activity heartbeat indicator in spinner
  - Shows `● active` (green) when engine is outputting
  - Shows `○ Xs ago` when idle, so you know it's still running
- Helps monitor long-running tasks without verbose logging

## [0.2.0] - 2025-01-17

### Added

- `--model` flag to specify which model to use with each AI engine
- `model` configuration option in `ralphy.yaml`
- `RALPHY_MODEL` environment variable support

### Examples

```bash
ralphy --claude --model opus
ralphy --claude --model claude-sonnet-4-5-20250929
ralphy --opencode --model anthropic/claude-sonnet-4-5
ralphy --opencode --model openai/gpt-5.2
```

## [0.1.0] - 2025-01-17

### Added

- Initial release of Ralphy Go CLI
- Support for multiple AI engines: Claude Code, OpenCode, Codex, Cursor
- Task sources: Markdown PRD files, YAML task files, GitHub issues
- Git integration with branch-per-task workflow
- Automatic PR creation with draft PR support
- Parallel task execution with configurable concurrency
- Retry logic with configurable delays and max retries
- Desktop notifications on task completion
- Configuration via CLI flags, environment variables, or config file
- Dry-run mode for previewing actions
- Verbose output mode for debugging

### Documentation

- README with installation and usage instructions
- Example PRD.md and tasks.yaml files
- Configuration file reference

[Unreleased]: https://github.com/ncecere/ralphy/compare/v0.3.2...HEAD
[0.3.2]: https://github.com/ncecere/ralphy/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/ncecere/ralphy/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/ncecere/ralphy/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/ncecere/ralphy/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/ncecere/ralphy/releases/tag/v0.1.0
