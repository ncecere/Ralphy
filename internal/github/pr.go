package github

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"text/template"
	"time"

	"github.com/google/go-github/v61/github"
	gitpkg "github.com/ncecere/ralphy/internal/git"
	"golang.org/x/oauth2"
)

// PROptions contains all options for creating a PR.
type PROptions struct {
	Title string
	Body  string
	Head  string
	Base  string
	Draft bool
	Owner string
	Repo  string
	Token string

	// New fields for enhanced PR creation
	Labels    []string
	Reviewers []string
	Assignees []string

	// Template data for body generation
	TemplateData *PRTemplateData
	BodyTemplate string
}

// PRTemplateData contains variables available in PR body templates.
type PRTemplateData struct {
	Task         string
	Engine       string
	Model        string
	InputTokens  int
	OutputTokens int
	TotalTokens  int
	Cost         float64
	Duration     time.Duration
	FilesChanged int
	Branch       string
	BaseBranch   string
}

// DefaultPRTemplate is used when no custom template is provided.
const DefaultPRTemplate = `## Summary

Automated PR created by Ralphy.

**Task:** {{.Task}}

## Details

| Metric | Value |
|--------|-------|
| Engine | {{.Engine}} |
{{- if .Model}}
| Model | {{.Model}} |
{{- end}}
{{- if .InputTokens}}
| Input Tokens | {{.InputTokens}} |
| Output Tokens | {{.OutputTokens}} |
| Total Tokens | {{.TotalTokens}} |
{{- end}}
{{- if gt .Cost 0.0}}
| Est. Cost | ${{printf "%.4f" .Cost}} |
{{- end}}
{{- if .Duration}}
| Duration | {{.Duration}} |
{{- end}}
{{- if .FilesChanged}}
| Files Changed | {{.FilesChanged}} |
{{- end}}
`

// CreatePR creates a pull request with optional labels, reviewers, and assignees.
func CreatePR(ctx context.Context, opts PROptions) (string, error) {
	if opts.Token == "" {
		return "", errors.New("github token required for PR creation")
	}

	owner, repo := opts.Owner, opts.Repo
	if owner == "" || repo == "" {
		var err error
		owner, repo, err = gitpkg.ParseRepoFromRemote()
		if err != nil {
			return "", err
		}
	}

	// Generate body from template if template data is provided
	body := opts.Body
	if opts.TemplateData != nil {
		tmplStr := opts.BodyTemplate
		if tmplStr == "" {
			tmplStr = DefaultPRTemplate
		}

		var err error
		body, err = renderTemplate(tmplStr, opts.TemplateData)
		if err != nil {
			// Fall back to simple body on template error
			body = fmt.Sprintf("Automated PR created by Ralphy.\n\nTask: %s", opts.TemplateData.Task)
		}
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: opts.Token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	pr := &github.NewPullRequest{
		Title: &opts.Title,
		Body:  &body,
		Head:  &opts.Head,
		Base:  &opts.Base,
		Draft: &opts.Draft,
	}

	created, _, err := client.PullRequests.Create(ctx, owner, repo, pr)
	if err != nil {
		return "", err
	}

	prNumber := created.GetNumber()
	prURL := created.GetHTMLURL()

	// Add labels if specified
	if len(opts.Labels) > 0 {
		_, _, err := client.Issues.AddLabelsToIssue(ctx, owner, repo, prNumber, opts.Labels)
		if err != nil {
			// Log but don't fail - PR was created
			fmt.Printf("Warning: failed to add labels: %v\n", err)
		}
	}

	// Request reviewers if specified
	if len(opts.Reviewers) > 0 {
		reviewers := github.ReviewersRequest{
			Reviewers: opts.Reviewers,
		}
		_, _, err := client.PullRequests.RequestReviewers(ctx, owner, repo, prNumber, reviewers)
		if err != nil {
			fmt.Printf("Warning: failed to request reviewers: %v\n", err)
		}
	}

	// Add assignees if specified
	if len(opts.Assignees) > 0 {
		_, _, err := client.Issues.AddAssignees(ctx, owner, repo, prNumber, opts.Assignees)
		if err != nil {
			fmt.Printf("Warning: failed to add assignees: %v\n", err)
		}
	}

	return prURL, nil
}

// renderTemplate renders a PR body template with the given data.
func renderTemplate(tmplStr string, data *PRTemplateData) (string, error) {
	tmpl, err := template.New("pr").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GetFilesChanged returns the number of files changed between base and head.
func GetFilesChanged(base string) (int, error) {
	out, err := gitpkg.RunInDir(".", "diff", "--name-only", base+"...HEAD")
	if err != nil {
		return 0, err
	}

	count := 0
	for _, line := range splitLines(out) {
		if line != "" {
			count++
		}
	}
	return count, nil
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
