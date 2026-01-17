package tasks

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestYAMLSource_ReadTasks(t *testing.T) {
	content := `tasks:
  - title: First task
    completed: false
  - title: Completed task
    completed: true
  - title: Second task
    completed: false
    parallel_group: 1
`
	path := writeTempYAMLFile(t, "tasks.yaml", content)
	source := NewYAMLSource(path)

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

	if tasks[1].ParallelGroup != 1 {
		t.Errorf("expected parallel group 1, got %d", tasks[1].ParallelGroup)
	}
}

func TestYAMLSource_Next(t *testing.T) {
	content := `tasks:
  - title: Done
    completed: true
  - title: Pending
    completed: false
`
	path := writeTempYAMLFile(t, "tasks.yaml", content)
	source := NewYAMLSource(path)

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

func TestYAMLSource_Summary(t *testing.T) {
	content := `tasks:
  - title: One
    completed: false
  - title: Two
    completed: true
  - title: Three
    completed: false
`
	path := writeTempYAMLFile(t, "tasks.yaml", content)
	source := NewYAMLSource(path)

	summary, err := source.Summary(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.Remaining != 2 {
		t.Errorf("expected 2 remaining, got %d", summary.Remaining)
	}

	if summary.Completed != 1 {
		t.Errorf("expected 1 completed, got %d", summary.Completed)
	}
}

func TestYAMLSource_MarkComplete(t *testing.T) {
	content := `tasks:
  - title: Task to complete
    completed: false
  - title: Other task
    completed: false
`
	path := writeTempYAMLFile(t, "tasks.yaml", content)
	source := NewYAMLSource(path)

	err := source.MarkComplete(context.Background(), Task{Title: "Task to complete"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := os.ReadFile(path)
	if !strings.Contains(string(updated), "completed: true") {
		t.Errorf("expected task to be marked complete, got:\n%s", string(updated))
	}
}

func writeTempYAMLFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}
