package github

import (
	"context"
	"errors"

	"github.com/google/go-github/v61/github"
	gitpkg "github.com/ncecere/ralphy/internal/git"
	"golang.org/x/oauth2"
)

type PROptions struct {
	Title string
	Body  string
	Head  string
	Base  string
	Draft bool
	Owner string
	Repo  string
	Token string
}

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

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: opts.Token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	pr := &github.NewPullRequest{
		Title: &opts.Title,
		Body:  &opts.Body,
		Head:  &opts.Head,
		Base:  &opts.Base,
		Draft: &opts.Draft,
	}

	created, _, err := client.PullRequests.Create(ctx, owner, repo, pr)
	if err != nil {
		return "", err
	}

	return created.GetHTMLURL(), nil
}
