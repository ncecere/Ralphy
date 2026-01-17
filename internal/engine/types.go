package engine

import "time"

type RunOptions struct {
	WorkDir    string
	OutputFile string
	Model      string
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
