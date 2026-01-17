package tasks

import (
	"context"
	"errors"
	"os"
	"strings"
)

const (
	markdownTodoPrefix = "- [ ]"
	markdownDonePrefix = "- [x]"
)

type MarkdownSource struct {
	Path string
}

func NewMarkdownSource(path string) *MarkdownSource {
	return &MarkdownSource{Path: path}
}

func (m *MarkdownSource) Name() string {
	return "markdown"
}

func (m *MarkdownSource) Next(ctx context.Context) (*Task, error) {
	_ = ctx
	items, err := m.readTasks()
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if !item.Completed {
			return &Task{ID: item.Title, Title: item.Title}, nil
		}
	}
	return nil, nil
}

func (m *MarkdownSource) All(ctx context.Context) ([]Task, error) {
	_ = ctx
	items, err := m.readTasks()
	if err != nil {
		return nil, err
	}
	var tasks []Task
	for _, item := range items {
		if item.Completed {
			continue
		}
		tasks = append(tasks, Task{ID: item.Title, Title: item.Title})
	}
	return tasks, nil
}

func (m *MarkdownSource) Summary(ctx context.Context) (Summary, error) {
	_ = ctx
	items, err := m.readTasks()
	if err != nil {
		return Summary{}, err
	}
	var summary Summary
	for _, item := range items {
		if item.Completed {
			summary.Completed++
		} else {
			summary.Remaining++
		}
	}
	return summary, nil
}

func (m *MarkdownSource) MarkComplete(ctx context.Context, task Task) error {
	_ = ctx
	if task.Title == "" {
		return errors.New("task title required")
	}

	content, err := os.ReadFile(m.Path)
	if err != nil {
		return err
	}

	lines, ending, hasTrailing := splitLines(string(content))
	updated := false

	for i, line := range lines {
		if !strings.HasPrefix(line, markdownTodoPrefix) {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(line, markdownTodoPrefix))
		if rest != task.Title {
			continue
		}
		lines[i] = strings.Replace(line, markdownTodoPrefix, markdownDonePrefix, 1)
		updated = true
	}

	if !updated {
		return nil
	}

	return writeLines(m.Path, lines, ending, hasTrailing)
}

type markdownItem struct {
	Title     string
	Completed bool
}

func (m *MarkdownSource) readTasks() ([]markdownItem, error) {
	content, err := os.ReadFile(m.Path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")
	items := make([]markdownItem, 0, len(lines))

	for _, line := range lines {
		if strings.HasPrefix(line, markdownTodoPrefix) {
			items = append(items, markdownItem{
				Title:     strings.TrimSpace(strings.TrimPrefix(line, markdownTodoPrefix)),
				Completed: false,
			})
			continue
		}
		if strings.HasPrefix(line, markdownDonePrefix) {
			items = append(items, markdownItem{
				Title:     strings.TrimSpace(strings.TrimPrefix(line, markdownDonePrefix)),
				Completed: true,
			})
		}
	}

	return items, nil
}
