package tasks

import (
	"os"
	"strings"
)

type lineEnding string

const (
	lineEndingLF   lineEnding = "\n"
	lineEndingCRLF lineEnding = "\r\n"
)

func splitLines(content string) ([]string, lineEnding, bool) {
	ending := lineEndingLF
	if strings.Contains(content, "\r\n") {
		ending = lineEndingCRLF
	}
	trimmed := strings.TrimSuffix(content, string(ending))
	if ending == lineEndingCRLF {
		trimmed = strings.TrimSuffix(trimmed, "\n")
	}
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines, ending, content != trimmed
}

func writeLines(path string, lines []string, ending lineEnding, trailing bool) error {
	content := strings.Join(lines, string(ending))
	if trailing {
		content += string(ending)
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
