# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2025-01-17

### Added

- `--model` flag to specify which model to use with each AI engine
- `model` configuration option in `ralphy.yaml`
- `RALPHY_MODEL` environment variable support

### Examples

```bash
ralphy --claude --model opus
ralphy --opencode --model gpt-4o
ralphy --codex --model o3
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

[Unreleased]: https://github.com/ncecere/ralphy/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/ncecere/ralphy/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/ncecere/ralphy/releases/tag/v0.1.0
