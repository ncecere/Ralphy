package tasks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/goccy/go-yaml/token"
)

type YAMLSource struct {
	Path string
}

func NewYAMLSource(path string) *YAMLSource {
	return &YAMLSource{Path: path}
}

func (y *YAMLSource) Name() string {
	return "yaml"
}

func (y *YAMLSource) Next(ctx context.Context) (*Task, error) {
	_ = ctx
	items, err := y.readTasks()
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if !item.Completed {
			return &Task{ID: item.Title, Title: item.Title, ParallelGroup: item.ParallelGroup}, nil
		}
	}
	return nil, nil
}

func (y *YAMLSource) All(ctx context.Context) ([]Task, error) {
	_ = ctx
	items, err := y.readTasks()
	if err != nil {
		return nil, err
	}
	var tasks []Task
	for _, item := range items {
		if item.Completed {
			continue
		}
		tasks = append(tasks, Task{ID: item.Title, Title: item.Title, ParallelGroup: item.ParallelGroup})
	}
	return tasks, nil
}

func (y *YAMLSource) Summary(ctx context.Context) (Summary, error) {
	_ = ctx
	items, err := y.readTasks()
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

func (y *YAMLSource) MarkComplete(ctx context.Context, task Task) error {
	_ = ctx
	if task.Title == "" {
		return errors.New("task title required")
	}

	file, err := parser.ParseFile(y.Path, parser.ParseComments)
	if err != nil {
		return err
	}

	taskMapping, err := findTaskMapping(file, task.Title)
	if err != nil {
		return err
	}
	if taskMapping == nil {
		return nil
	}

	updated := setMappingBool(taskMapping, "completed", true)
	if !updated {
		return nil
	}

	content := file.String()
	return os.WriteFile(y.Path, []byte(content), 0o644)
}

type yamlItem struct {
	Title         string
	Completed     bool
	ParallelGroup int
}

func (y *YAMLSource) readTasks() ([]yamlItem, error) {
	file, err := parser.ParseFile(y.Path, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	seq, err := findTasksSequence(file)
	if err != nil {
		return nil, err
	}
	if seq == nil {
		return nil, nil
	}

	items := make([]yamlItem, 0, len(seq.Values))
	for _, entry := range seq.Values {
		mapping, ok := entry.(*ast.MappingNode)
		if !ok {
			continue
		}
		item := yamlItem{ParallelGroup: 0}
		for _, value := range mapping.Values {
			key, ok := scalarString(value.Key)
			if !ok {
				continue
			}
			switch key {
			case "title":
				if title, ok := scalarString(value.Value); ok {
					item.Title = title
				}
			case "completed":
				if completed, ok := scalarBool(value.Value); ok {
					item.Completed = completed
				}
			case "parallel_group":
				if group, ok := scalarInt(value.Value); ok {
					item.ParallelGroup = group
				}
			}
		}
		if item.Title != "" {
			items = append(items, item)
		}
	}

	return items, nil
}

func findTasksSequence(file *ast.File) (*ast.SequenceNode, error) {
	path, err := yaml.PathString("$.tasks")
	if err != nil {
		return nil, err
	}

	node, err := path.FilterFile(file)
	if err != nil {
		if errors.Is(err, yaml.ErrNotFoundNode) {
			return nil, nil
		}
		return nil, err
	}

	seq, ok := node.(*ast.SequenceNode)
	if !ok {
		return nil, fmt.Errorf("tasks is not a sequence")
	}
	return seq, nil
}

func findTaskMapping(file *ast.File, title string) (*ast.MappingNode, error) {
	seq, err := findTasksSequence(file)
	if err != nil || seq == nil {
		return nil, err
	}
	for _, entry := range seq.Values {
		mapping, ok := entry.(*ast.MappingNode)
		if !ok {
			continue
		}
		for _, value := range mapping.Values {
			key, ok := scalarString(value.Key)
			if !ok || key != "title" {
				continue
			}
			if titleValue, ok := scalarString(value.Value); ok && titleValue == title {
				return mapping, nil
			}
		}
	}
	return nil, nil
}

func setMappingBool(mapping *ast.MappingNode, key string, value bool) bool {
	for _, entry := range mapping.Values {
		entryKey, ok := scalarString(entry.Key)
		if !ok || entryKey != key {
			continue
		}

		tokenValue := "false"
		if value {
			tokenValue = "true"
		}

		scalar, ok := entry.Value.(ast.ScalarNode)
		if !ok {
			return false
		}

		tk := scalar.GetToken().Clone()
		tk.Type = token.BoolType
		tk.Value = tokenValue
		tk.Origin = tokenValue

		boolNode := ast.Bool(tk)
		_ = entry.Replace(boolNode)
		return true
	}

	return false
}

func scalarString(node ast.Node) (string, bool) {
	scalar, ok := node.(ast.ScalarNode)
	if !ok {
		return "", false
	}
	value := scalar.GetValue()
	switch typed := value.(type) {
	case string:
		return typed, true
	case fmt.Stringer:
		return typed.String(), true
	case int:
		return strconv.Itoa(typed), true
	case int64:
		return strconv.FormatInt(typed, 10), true
	case uint64:
		return strconv.FormatUint(typed, 10), true
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64), true
	case bool:
		if typed {
			return "true", true
		}
		return "false", true
	default:
		text := strings.TrimSpace(node.String())
		if text == "" {
			return "", false
		}
		return text, true
	}
}

func scalarBool(node ast.Node) (bool, bool) {
	scalar, ok := node.(ast.ScalarNode)
	if !ok {
		return false, false
	}
	value := scalar.GetValue()
	switch typed := value.(type) {
	case bool:
		return typed, true
	case string:
		parsed, err := strconv.ParseBool(typed)
		if err != nil {
			return false, false
		}
		return parsed, true
	default:
		text := strings.TrimSpace(node.String())
		parsed, err := strconv.ParseBool(text)
		if err != nil {
			return false, false
		}
		return parsed, true
	}
}

func scalarInt(node ast.Node) (int, bool) {
	scalar, ok := node.(ast.ScalarNode)
	if !ok {
		return 0, false
	}
	value := scalar.GetValue()
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case uint64:
		return int(typed), true
	case float64:
		return int(typed), true
	case string:
		parsed, err := strconv.Atoi(typed)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		text := strings.TrimSpace(node.String())
		parsed, err := strconv.Atoi(text)
		if err != nil {
			return 0, false
		}
		return parsed, true
	}
}
