package git

import (
	"regexp"
	"strings"
)

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

func Slugify(text string) string {
	slug := strings.ToLower(text)
	slug = nonAlphaNum.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if len(slug) > 50 {
		slug = slug[:50]
	}
	return slug
}

func TaskBranchName(taskTitle string) string {
	return "ralphy/" + Slugify(taskTitle)
}

func AgentBranchName(agentNum int, taskTitle string) string {
	return "ralphy/agent-" + itoa(agentNum) + "-" + Slugify(taskTitle)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
