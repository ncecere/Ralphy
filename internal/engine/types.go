package engine

import (
	"io"
	"sync"
	"time"
)

type RunOptions struct {
	WorkDir    string
	OutputFile string
	Model      string
	Activity   *ActivityTracker
}

// ActivityTracker tracks when the last output was received from the engine.
// It can be used by the UI to show that the engine is still active.
type ActivityTracker struct {
	mu         sync.RWMutex
	lastOutput time.Time
	started    time.Time
}

// NewActivityTracker creates a new activity tracker.
func NewActivityTracker() *ActivityTracker {
	now := time.Now()
	return &ActivityTracker{
		lastOutput: now,
		started:    now,
	}
}

// Touch updates the last output time to now.
func (a *ActivityTracker) Touch() {
	a.mu.Lock()
	a.lastOutput = time.Now()
	a.mu.Unlock()
}

// LastOutputAgo returns how long ago the last output was received.
func (a *ActivityTracker) LastOutputAgo() time.Duration {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return time.Since(a.lastOutput)
}

// HasOutput returns true if any output has been received since start.
func (a *ActivityTracker) HasOutput() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.lastOutput.After(a.started)
}

// activityWriter wraps an io.Writer and updates the tracker on each write.
type activityWriter struct {
	w       io.Writer
	tracker *ActivityTracker
}

func (aw *activityWriter) Write(p []byte) (n int, err error) {
	n, err = aw.w.Write(p)
	if n > 0 {
		aw.tracker.Touch()
	}
	return
}

type RunResult struct {
	Engine       string
	Response     string
	InputTokens  int
	OutputTokens int
	ActualCost   float64
	Duration     time.Duration
	OutputFile   string
	RawOutput    string
}
