package engine

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func Run(ctx context.Context, engine string, prompt string, opts RunOptions) (RunResult, error) {
	if opts.OutputFile == "" {
		return RunResult{}, fmt.Errorf("output file required")
	}
	if opts.WorkDir == "" {
		opts.WorkDir = "."
	}

	switch engine {
	case "opencode":
		return runOpenCode(ctx, prompt, opts)
	case "cursor":
		return runCursor(ctx, prompt, opts)
	case "codex":
		return runCodex(ctx, prompt, opts)
	default:
		return runClaude(ctx, prompt, opts)
	}
}

func runClaude(ctx context.Context, prompt string, opts RunOptions) (RunResult, error) {
	args := []string{"--dangerously-skip-permissions", "--verbose", "--output-format", "stream-json"}
	if opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}
	args = append(args, "-p", prompt)
	cmd := exec.CommandContext(ctx, "claude", args...)
	return execute(ctx, cmd, "claude", opts)
}

func runOpenCode(ctx context.Context, prompt string, opts RunOptions) (RunResult, error) {
	args := []string{"run", "--format", "json"}
	if opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}
	args = append(args, prompt)
	cmd := exec.CommandContext(ctx, "opencode", args...)
	cmd.Env = append(os.Environ(), `OPENCODE_PERMISSION={"*":"allow"}`)
	return execute(ctx, cmd, "opencode", opts)
}

func runCursor(ctx context.Context, prompt string, opts RunOptions) (RunResult, error) {
	args := []string{"--print", "--force", "--output-format", "stream-json"}
	if opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}
	args = append(args, prompt)
	cmd := exec.CommandContext(ctx, "agent", args...)
	return execute(ctx, cmd, "cursor", opts)
}

func runCodex(ctx context.Context, prompt string, opts RunOptions) (RunResult, error) {
	args := []string{"exec", "--full-auto", "--json", "--output-last-message", opts.OutputFile + ".last"}
	if opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}
	args = append(args, prompt)
	cmd := exec.CommandContext(ctx, "codex", args...)
	return execute(ctx, cmd, "codex", opts)
}

func execute(ctx context.Context, cmd *exec.Cmd, engine string, opts RunOptions) (RunResult, error) {
	cmd.Dir = opts.WorkDir

	outFile, err := os.Create(opts.OutputFile)
	if err != nil {
		return RunResult{}, err
	}
	defer outFile.Close()

	// If activity tracker is provided, wrap the writer to track output
	if opts.Activity != nil {
		aw := &activityWriter{w: outFile, tracker: opts.Activity}
		cmd.Stdout = aw
		cmd.Stderr = aw
	} else {
		cmd.Stdout = outFile
		cmd.Stderr = outFile
	}

	start := time.Now()
	if err := cmd.Run(); err != nil {
		return RunResult{}, err
	}

	result, err := parseOutput(opts.OutputFile, engine)
	if err != nil {
		return RunResult{}, err
	}
	result.Engine = engine
	result.OutputFile = opts.OutputFile
	result.Duration = time.Since(start)

	return result, nil
}
