package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseClaude(t *testing.T) {
	content := `{"type":"assistant","message":"thinking..."}
{"type":"result","result":"Task completed successfully","usage":{"input_tokens":100,"output_tokens":50}}`

	path := writeTempFile(t, "output.log", content)
	result, err := parseOutput(path, "claude")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Response != "Task completed successfully" {
		t.Errorf("expected response 'Task completed successfully', got %q", result.Response)
	}

	if result.InputTokens != 100 {
		t.Errorf("expected 100 input tokens, got %d", result.InputTokens)
	}

	if result.OutputTokens != 50 {
		t.Errorf("expected 50 output tokens, got %d", result.OutputTokens)
	}
}

func TestParseClaude_Error(t *testing.T) {
	content := `{"type":"error","error":{"message":"API error"}}`

	path := writeTempFile(t, "output.log", content)
	_, err := parseOutput(path, "claude")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "API error" {
		t.Errorf("expected error message 'API error', got %q", err.Error())
	}
}

func TestParseCursor(t *testing.T) {
	content := `{"type":"assistant","message":{"content":"working..."}}
{"type":"result","result":"Done","duration_ms":5000}`

	path := writeTempFile(t, "output.log", content)
	result, err := parseOutput(path, "cursor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Response != "Done" {
		t.Errorf("expected response 'Done', got %q", result.Response)
	}

	if result.Duration.Milliseconds() != 5000 {
		t.Errorf("expected 5000ms duration, got %d", result.Duration.Milliseconds())
	}
}

func TestParseCodex(t *testing.T) {
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "output.log")
	lastPath := mainPath + ".last"

	_ = os.WriteFile(mainPath, []byte("{}"), 0o644)
	_ = os.WriteFile(lastPath, []byte("Task completed successfully."), 0o644)

	result, err := parseOutput(mainPath, "codex")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Response != "" {
		t.Errorf("expected empty response after trimming prefix, got %q", result.Response)
	}
}

func TestParseEmpty(t *testing.T) {
	path := writeTempFile(t, "output.log", "")
	_, err := parseOutput(path, "claude")
	if err != ErrEmptyOutput {
		t.Errorf("expected ErrEmptyOutput, got %v", err)
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
