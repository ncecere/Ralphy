package tasks

import "context"

type Task struct {
	ID            string
	Title         string
	ParallelGroup int
}

type Summary struct {
	Remaining int
	Completed int
}

type Source interface {
	Next(ctx context.Context) (*Task, error)
	All(ctx context.Context) ([]Task, error)
	Summary(ctx context.Context) (Summary, error)
	MarkComplete(ctx context.Context, task Task) error
	Name() string
}
