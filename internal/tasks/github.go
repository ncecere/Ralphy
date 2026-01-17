package tasks

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v61/github"
	"github.com/ncecere/ralphy/internal/utils"
	"golang.org/x/oauth2"
)

type GitHubSource struct {
	Owner string
	Repo  string
	Label string
	Token string
}

func NewGitHubSource(repo, label, token string) (*GitHubSource, error) {
	owner, name, err := splitRepo(repo)
	if err != nil {
		return nil, err
	}
	return &GitHubSource{Owner: owner, Repo: name, Label: label, Token: token}, nil
}

func (g *GitHubSource) Name() string {
	return "github"
}

func (g *GitHubSource) Next(ctx context.Context) (*Task, error) {
	issues, err := g.listIssues(ctx, "open")
	if err != nil {
		return nil, err
	}
	if len(issues) == 0 {
		return nil, nil
	}
	issue := issues[0]
	return &Task{
		ID:    fmt.Sprintf("%d", issue.GetNumber()),
		Title: issue.GetTitle(),
	}, nil
}

func (g *GitHubSource) All(ctx context.Context) ([]Task, error) {
	issues, err := g.listIssues(ctx, "open")
	if err != nil {
		return nil, err
	}
	tasks := make([]Task, 0, len(issues))
	for _, issue := range issues {
		tasks = append(tasks, Task{ID: fmt.Sprintf("%d", issue.GetNumber()), Title: issue.GetTitle()})
	}
	return tasks, nil
}

func (g *GitHubSource) Summary(ctx context.Context) (Summary, error) {
	openIssues, err := g.listIssues(ctx, "open")
	if err != nil {
		return Summary{}, err
	}
	closedIssues, err := g.listIssues(ctx, "closed")
	if err != nil {
		return Summary{}, err
	}
	return Summary{Remaining: len(openIssues), Completed: len(closedIssues)}, nil
}

func (g *GitHubSource) MarkComplete(ctx context.Context, task Task) error {
	issueNumber, err := parseIssueNumber(task)
	if err != nil {
		return err
	}

	client := g.client(ctx)
	state := "closed"
	_, _, err = client.Issues.Edit(ctx, g.Owner, g.Repo, issueNumber, &github.IssueRequest{State: &state})
	return err
}

func (g *GitHubSource) IssueBody(ctx context.Context, task Task) (string, error) {
	issueNumber, err := parseIssueNumber(task)
	if err != nil {
		return "", err
	}
	client := g.client(ctx)
	issue, _, err := client.Issues.Get(ctx, g.Owner, g.Repo, issueNumber)
	if err != nil {
		return "", err
	}
	return issue.GetBody(), nil
}

func (g *GitHubSource) listIssues(ctx context.Context, state string) ([]*github.Issue, error) {
	client := g.client(ctx)
	opts := &github.IssueListByRepoOptions{
		State:       state,
		Sort:        "created",
		Direction:   "asc",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	if g.Label != "" {
		opts.Labels = []string{g.Label}
	}

	var all []*github.Issue
	for {
		issues, resp, err := client.Issues.ListByRepo(ctx, g.Owner, g.Repo, opts)
		if err != nil {
			return nil, err
		}
		for _, issue := range issues {
			if issue.PullRequestLinks != nil {
				continue
			}
			all = append(all, issue)
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	sort.SliceStable(all, func(i, j int) bool {
		return issueCreatedAt(all[i]).Before(issueCreatedAt(all[j]))
	})

	return all, nil
}

func issueCreatedAt(issue *github.Issue) time.Time {
	if issue == nil {
		return time.Time{}
	}
	if issue.CreatedAt != nil {
		return issue.CreatedAt.Time
	}
	return time.Time{}
}

func (g *GitHubSource) client(ctx context.Context) *github.Client {
	if g.Token == "" {
		return github.NewClient(nil)
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: g.Token})
	client := oauth2.NewClient(ctx, ts)
	return github.NewClient(client)
}

func splitRepo(repo string) (string, string, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", "", errors.New("github repo must be in owner/repo format")
	}
	return parts[0], parts[1], nil
}

func parseIssueNumber(task Task) (int, error) {
	if task.ID == "" {
		return 0, errors.New("task id required")
	}
	num, err := utils.ParseInt(task.ID)
	if err != nil {
		return 0, fmt.Errorf("invalid issue number: %w", err)
	}
	return num, nil
}
