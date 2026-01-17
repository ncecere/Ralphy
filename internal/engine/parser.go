package engine

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type streamEvent struct {
	Type       string          `json:"type"`
	Result     string          `json:"result"`
	Usage      *usageInfo      `json:"usage"`
	DurationMS int64           `json:"duration_ms"`
	Part       json.RawMessage `json:"part"`
	Error      *streamError    `json:"error"`
	Message    json.RawMessage `json:"message"`
}

type assistantMessage struct {
	Content any `json:"content"`
}

type streamError struct {
	Message string `json:"message"`
}

type usageInfo struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type openCodeTokens struct {
	Tokens struct {
		Input  int `json:"input"`
		Output int `json:"output"`
	} `json:"tokens"`
	Cost float64 `json:"cost"`
	Text string  `json:"text"`
}

type openCodePart struct {
	Text   string         `json:"text"`
	Cost   float64        `json:"cost"`
	Part   openCodeTokens `json:"part"`
	Tokens openCodeTokens `json:"tokens"`
}

func parseOutput(path string, engine string) (RunResult, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return RunResult{}, err
	}

	trimmed := strings.TrimSpace(string(content))
	if trimmed == "" {
		return RunResult{}, ErrEmptyOutput
	}

	result := RunResult{RawOutput: string(content)}

	switch engine {
	case "opencode":
		return parseOpenCode(trimmed, result)
	case "cursor":
		return parseCursor(trimmed, result)
	case "codex":
		return parseCodex(path, result)
	default:
		return parseClaude(trimmed, result)
	}
}

func parseClaude(content string, result RunResult) (RunResult, error) {
	lines := scanJSONLines(content)
	var lastResult *streamEvent
	for _, line := range lines {
		evt, ok := decodeEvent(line)
		if !ok {
			continue
		}
		if evt.Type == "error" {
			return result, parseError(evt)
		}
		if evt.Type == "result" {
			lastResult = evt
		}
	}

	if lastResult == nil {
		return result, nil
	}

	result.Response = lastResult.Result
	if lastResult.Usage != nil {
		result.InputTokens = lastResult.Usage.InputTokens
		result.OutputTokens = lastResult.Usage.OutputTokens
	}
	return result, nil
}

func parseOpenCode(content string, result RunResult) (RunResult, error) {
	lines := scanJSONLines(content)
	var textBuilder strings.Builder
	var inputTokens int
	var outputTokens int
	var cost float64
	for _, line := range lines {
		evt, ok := decodeEvent(line)
		if !ok {
			continue
		}
		if evt.Type == "error" {
			return result, parseError(evt)
		}
		switch evt.Type {
		case "text":
			var part openCodePart
			if err := json.Unmarshal(evt.Part, &part); err == nil {
				if part.Text != "" {
					textBuilder.WriteString(part.Text)
				} else if part.Part.Text != "" {
					textBuilder.WriteString(part.Part.Text)
				}
			}
		case "step_finish":
			var part openCodePart
			if err := json.Unmarshal(evt.Part, &part); err == nil {
				if part.Tokens.Tokens.Input > 0 {
					inputTokens = part.Tokens.Tokens.Input
					outputTokens = part.Tokens.Tokens.Output
				}
				if part.Cost > 0 {
					cost = part.Cost
				} else if part.Part.Cost > 0 {
					cost = part.Part.Cost
				}
				if part.Part.Tokens.Input > 0 {
					inputTokens = part.Part.Tokens.Input
					outputTokens = part.Part.Tokens.Output
				}
			}
		}
	}

	result.Response = strings.TrimSpace(textBuilder.String())
	if inputTokens > 0 {
		result.InputTokens = inputTokens
		result.OutputTokens = outputTokens
	}
	if cost > 0 {
		result.ActualCost = cost
	}
	if result.Response == "" {
		result.Response = "Task completed"
	}
	return result, nil
}

func parseCursor(content string, result RunResult) (RunResult, error) {
	lines := scanJSONLines(content)
	var lastResult *streamEvent
	var lastAssistant *streamEvent
	for _, line := range lines {
		evt, ok := decodeEvent(line)
		if !ok {
			continue
		}
		if evt.Type == "error" {
			return result, parseError(evt)
		}
		switch evt.Type {
		case "result":
			lastResult = evt
		case "assistant":
			lastAssistant = evt
		}
	}

	if lastResult != nil {
		result.Response = strings.TrimSpace(lastResult.Result)
		if lastResult.DurationMS > 0 {
			result.Duration = time.Duration(lastResult.DurationMS) * time.Millisecond
		}
	}

	if result.Response == "" && lastAssistant != nil {
		result.Response = extractAssistantText(lastAssistant)
	}

	if result.Response == "" {
		result.Response = "Task completed"
	}
	return result, nil
}

func parseCodex(path string, result RunResult) (RunResult, error) {
	lastMessagePath := path + ".last"
	content, err := os.ReadFile(lastMessagePath)
	if err == nil {
		result.Response = strings.TrimSpace(string(content))
		result.Response = strings.TrimPrefix(result.Response, "Task completed successfully.")
		result.Response = strings.TrimSpace(result.Response)
		return result, nil
	}

	return result, nil
}

func scanJSONLines(content string) []string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	lines := make([]string, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func decodeEvent(line string) (*streamEvent, bool) {
	var evt streamEvent
	if err := json.Unmarshal([]byte(line), &evt); err != nil {
		return nil, false
	}
	return &evt, true
}

func parseError(evt *streamEvent) error {
	if evt.Error != nil && evt.Error.Message != "" {
		return errors.New(evt.Error.Message)
	}
	if len(evt.Message) > 0 {
		var message any
		if err := json.Unmarshal(evt.Message, &message); err == nil {
			switch msg := message.(type) {
			case string:
				if msg != "" {
					return errors.New(msg)
				}
			}
		}
	}
	return errors.New("unknown engine error")
}

func extractAssistantText(evt *streamEvent) string {
	if len(evt.Message) == 0 {
		return ""
	}
	var msg assistantMessage
	if err := json.Unmarshal(evt.Message, &msg); err != nil {
		return ""
	}
	if msg.Content == nil {
		return ""
	}

	switch content := msg.Content.(type) {
	case string:
		return strings.TrimSpace(content)
	case []any:
		var builder strings.Builder
		for _, item := range content {
			text, ok := extractContentText(item)
			if !ok {
				continue
			}
			builder.WriteString(text)
		}
		return strings.TrimSpace(builder.String())
	default:
		return ""
	}
}

func extractContentText(item any) (string, bool) {
	mapValue, ok := item.(map[string]any)
	if !ok {
		return "", false
	}
	textValue, ok := mapValue["text"]
	if !ok {
		return "", false
	}
	switch text := textValue.(type) {
	case string:
		return text, true
	default:
		return fmt.Sprintf("%v", textValue), true
	}
}
