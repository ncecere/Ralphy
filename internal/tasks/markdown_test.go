package tasks

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMarkdownSource_ReadTasks(t *testing.T) {
	content := `# PRD

## Tasks
- [ ] First task
- [x] Completed task
- [ ] Second task
`
	path := writeTempFile(t, "PRD.md", content)
	source := NewMarkdownSource(path)

	tasks, err := source.All(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("expected 2 pending tasks, got %d", len(tasks))
	}

	if tasks[0].Title != "First task" {
		t.Errorf("expected first task title 'First task', got %q", tasks[0].Title)
	}

	if tasks[1].Title != "Second task" {
		t.Errorf("expected second task title 'Second task', got %q", tasks[1].Title)
	}
}

func TestMarkdownSource_Next(t *testing.T) {
	content := `- [x] Done
- [ ] Pending
`
	path := writeTempFile(t, "PRD.md", content)
	source := NewMarkdownSource(path)

	task, err := source.Next(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if task == nil {
		t.Fatal("expected a task, got nil")
	}

	if task.Title != "Pending" {
		t.Errorf("expected task title 'Pending', got %q", task.Title)
	}
}

func TestMarkdownSource_Summary(t *testing.T) {
	content := `- [ ] One
- [x] Two
- [ ] Three
- [x] Four
`
	path := writeTempFile(t, "PRD.md", content)
	source := NewMarkdownSource(path)

	summary, err := source.Summary(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.Remaining != 2 {
		t.Errorf("expected 2 remaining, got %d", summary.Remaining)
	}

	if summary.Completed != 2 {
		t.Errorf("expected 2 completed, got %d", summary.Completed)
	}
}

func TestMarkdownSource_MarkComplete(t *testing.T) {
	content := `- [ ] Task to complete
- [ ] Other task
`
	path := writeTempFile(t, "PRD.md", content)
	source := NewMarkdownSource(path)

	err := source.MarkComplete(context.Background(), Task{Title: "Task to complete"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := os.ReadFile(path)
	expected := `- [x] Task to complete
- [ ] Other task
`
	if string(updated) != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, string(updated))
	}
}

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}
