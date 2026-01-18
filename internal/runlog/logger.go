// Package runlog provides run history logging functionality.
package runlog

import (
	"encoding/json"
	"os"
	"time"
)

// RunEntry represents a single run's log entry.
type RunEntry struct {
	Timestamp    time.Time    `json:"timestamp"`
	Version      string       `json:"version"`
	Engine       string       `json:"engine"`
	Model        string       `json:"model,omitempty"`
	TaskSource   string       `json:"task_source"`
	TaskFile     string       `json:"task_file,omitempty"`
	Tasks        []TaskEntry  `json:"tasks"`
	Summary      SummaryEntry `json:"summary"`
	DurationSecs float64      `json:"duration_secs"`
}

// TaskEntry represents a single task within a run.
type TaskEntry struct {
	Title        string  `json:"title"`
	Status       string  `json:"status"` // "success", "failed", "skipped"
	Branch       string  `json:"branch,omitempty"`
	PRUrl        string  `json:"pr_url,omitempty"`
	InputTokens  int     `json:"input_tokens,omitempty"`
	OutputTokens int     `json:"output_tokens,omitempty"`
	Cost         float64 `json:"cost,omitempty"`
	DurationSecs float64 `json:"duration_secs,omitempty"`
	Error        string  `json:"error,omitempty"`
	Retries      int     `json:"retries,omitempty"`
}

// SummaryEntry contains aggregate statistics for the run.
type SummaryEntry struct {
	TotalTasks    int     `json:"total_tasks"`
	Succeeded     int     `json:"succeeded"`
	Failed        int     `json:"failed"`
	Skipped       int     `json:"skipped"`
	TotalTokens   int     `json:"total_tokens"`
	InputTokens   int     `json:"input_tokens"`
	OutputTokens  int     `json:"output_tokens"`
	EstimatedCost float64 `json:"estimated_cost"`
}

// Logger handles writing run history to a log file.
type Logger struct {
	path    string
	entry   *RunEntry
	started time.Time
}

// New creates a new run logger.
func New(path string) *Logger {
	if path == "" {
		return nil
	}
	return &Logger{
		path:    path,
		started: time.Now(),
		entry: &RunEntry{
			Timestamp: time.Now(),
			Tasks:     make([]TaskEntry, 0),
		},
	}
}

// SetRunInfo sets the run-level metadata.
func (l *Logger) SetRunInfo(version, engine, model, taskSource, taskFile string) {
	if l == nil {
		return
	}
	l.entry.Version = version
	l.entry.Engine = engine
	l.entry.Model = model
	l.entry.TaskSource = taskSource
	l.entry.TaskFile = taskFile
}

// AddTask adds a task entry to the log.
func (l *Logger) AddTask(task TaskEntry) {
	if l == nil {
		return
	}
	l.entry.Tasks = append(l.entry.Tasks, task)

	// Update summary
	l.entry.Summary.TotalTasks++
	switch task.Status {
	case "success":
		l.entry.Summary.Succeeded++
	case "failed":
		l.entry.Summary.Failed++
	case "skipped":
		l.entry.Summary.Skipped++
	}
	l.entry.Summary.InputTokens += task.InputTokens
	l.entry.Summary.OutputTokens += task.OutputTokens
	l.entry.Summary.TotalTokens += task.InputTokens + task.OutputTokens
	l.entry.Summary.EstimatedCost += task.Cost
}

// Write writes the log entry to the file.
// Appends to existing file as newline-delimited JSON (NDJSON).
func (l *Logger) Write() error {
	if l == nil {
		return nil
	}

	l.entry.DurationSecs = time.Since(l.started).Seconds()

	data, err := json.Marshal(l.entry)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(append(data, '\n'))
	return err
}

// ReadLog reads all entries from a log file.
func ReadLog(path string) ([]RunEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entries []RunEntry
	decoder := json.NewDecoder(&bytesReader{data: data})

	for decoder.More() {
		var entry RunEntry
		if err := decoder.Decode(&entry); err != nil {
			continue // Skip malformed entries
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// bytesReader wraps a byte slice for json.Decoder.
type bytesReader struct {
	data []byte
	pos  int
}

func (r *bytesReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, os.ErrClosed
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
