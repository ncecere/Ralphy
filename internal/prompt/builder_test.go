package prompt

import (
	"strings"
	"testing"

	"github.com/ncecere/ralphy/internal/config"
	"github.com/ncecere/ralphy/internal/tasks"
)

func TestBuild_Markdown(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.PRDSource = config.PRDSourceMarkdown
	cfg.PRDFile = "PRD.md"

	prompt := Build(Options{
		Task:   tasks.Task{Title: "Test task"},
		Config: cfg,
	})

	if !strings.Contains(prompt, "@PRD.md") {
		t.Error("expected prompt to contain @PRD.md")
	}

	if !strings.Contains(prompt, "@progress.txt") {
		t.Error("expected prompt to contain @progress.txt")
	}

	if !strings.Contains(prompt, "Find the highest-priority incomplete task") {
		t.Error("expected prompt to contain task instruction")
	}

	if !strings.Contains(prompt, "Write tests") {
		t.Error("expected prompt to contain test instruction when skip_tests is false")
	}

	if !strings.Contains(prompt, "Run linting") {
		t.Error("expected prompt to contain lint instruction when skip_lint is false")
	}

	if !strings.Contains(prompt, "- [ ]") {
		t.Error("expected prompt to contain markdown completion instruction")
	}
}

func TestBuild_SkipTests(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.SkipTests = true

	prompt := Build(Options{
		Task:   tasks.Task{Title: "Test task"},
		Config: cfg,
	})

	if strings.Contains(prompt, "Write tests") {
		t.Error("expected prompt to NOT contain test instruction when skip_tests is true")
	}
}

func TestBuild_SkipLint(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.SkipLint = true

	prompt := Build(Options{
		Task:   tasks.Task{Title: "Test task"},
		Config: cfg,
	})

	if strings.Contains(prompt, "Run linting") {
		t.Error("expected prompt to NOT contain lint instruction when skip_lint is true")
	}
}

func TestBuild_YAML(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.PRDSource = config.PRDSourceYAML
	cfg.PRDFile = "tasks.yaml"

	prompt := Build(Options{
		Task:   tasks.Task{Title: "Test task"},
		Config: cfg,
	})

	if !strings.Contains(prompt, "@tasks.yaml") {
		t.Error("expected prompt to contain @tasks.yaml")
	}

	if !strings.Contains(prompt, "completed: true") {
		t.Error("expected prompt to contain YAML completion instruction")
	}
}

func TestBuild_GitHub(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.PRDSource = config.PRDSourceGitHub

	prompt := Build(Options{
		Task:      tasks.Task{ID: "123", Title: "Fix bug"},
		IssueBody: "This is the issue body",
		Config:    cfg,
	})

	if !strings.Contains(prompt, "GitHub Issue") {
		t.Error("expected prompt to contain GitHub Issue reference")
	}

	if !strings.Contains(prompt, "This is the issue body") {
		t.Error("expected prompt to contain issue body")
	}
}

func TestBuildParallel(t *testing.T) {
	prompt := BuildParallel(tasks.Task{Title: "Implement feature"})

	if !strings.Contains(prompt, "TASK: Implement feature") {
		t.Error("expected parallel prompt to contain task title")
	}

	if !strings.Contains(prompt, "Do NOT modify PRD") {
		t.Error("expected parallel prompt to contain PRD instruction")
	}
}
