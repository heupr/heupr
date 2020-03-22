package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/google/go-github/v28/github"

	"github.com/heupr/heupr/backend"
)

/*
Description:
`estimatepr` plugin provides expected vs actual pull request
timeline projections.

Setup:
In the `.heupr.yml` file, include a backend option:

```
backends:
- name: estimatepr
  events:
  - name: pull_request
    actions:
    - merged
  settings: {}
```

Notes:
This plugin will search for pull requests bearing a label prefixed
with "est-" and a number (e.g. 1, 3, 10), which it will take as the
estimated time in days ("est-1" = 1 day) and place a comment on the
pull request when it is merged with the actual number of days it took
to complete based on commit activity.
*/

type helper interface {
	pullRequests(c *github.Client, owner, repo string) ([]*github.PullRequest, error)
	commits(c *github.Client, owner, repo string, number int) ([]*github.RepositoryCommit, error)
	stringPtr(input string) *string
	comment(c *github.Client, owner, repo string, pr *github.PullRequest) error
}

type help struct{}

func (h *help) pullRequests(c *github.Client, owner, repo string) ([]*github.PullRequest, error) {
	output := []*github.PullRequest{}

	opts := &github.PullRequestListOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		pullRequests, resp, err := c.PullRequests.List(context.Background(), owner, repo, opts)
		if err != nil {
			return nil, err
		}
		log.Printf("pull requests: %+v\n", pullRequests)
		output = append(output, pullRequests...)

		if resp.NextPage == 0 {
			break
		} else {
			opts.ListOptions.Page = resp.NextPage
		}
	}

	return output, nil
}

func (h *help) commits(c *github.Client, owner, repo string, number int) ([]*github.RepositoryCommit, error) {
	commits, _, err := c.PullRequests.ListCommits(context.Background(), owner, repo, number, &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	return commits, nil
}

func (h *help) stringPtr(input string) *string {
	return &input
}

func (h *help) comment(c *github.Client, owner, repo string, pr *github.PullRequest) error {
	commits, err := h.commits(c, owner, repo, *pr.Number)
	if err != nil {
		return errors.New("error getting commits: " + err.Error())
	}
	log.Printf("commits: %+v\n", commits)

	actual := 1
	estimated := ""

	if len(commits) > 1 {
		start := *commits[0].Commit.Author.Date
		stop := *commits[len(commits)-1].Commit.Author.Date

		actual += int(stop.Sub(start).Seconds() / 86400)
	}

	re := regexp.MustCompile("[0-9]+")
	for _, label := range pr.Labels {
		if strings.Contains(*label.Name, "est-") {
			estimated = re.FindAllString(*label.Name, -1)[0]
		}
	}
	log.Printf("estimated: %s, actual: %d\n", estimated, actual)

	cmt := &github.IssueComment{
		Body: h.stringPtr(fmt.Sprintf("### Completion results\n- Estimated day(s): **%s**\n- Actual day(s): **%d**\n", estimated, actual)),
	}

	_, _, err = c.Issues.CreateComment(context.Background(), owner, repo, *pr.Number, cmt)
	if err != nil {
		return fmt.Errorf("error posting pull request comment: %s", err.Error())
	}

	return nil
}

// Backend implements the backend package interface
var Backend bnkd

type bnkd struct {
	client *github.Client
	help   helper
}

// Configure configures the backend with a client and helper struct
func (b *bnkd) Configure(c *github.Client) {
	log.Println("configure estimate pull request backend")
	b.client = c
	b.help = &help{}
}

func parseRepos(event string, payloadBytes []byte) ([]*github.Repository, error) {
	repos := []*github.Repository{}
	if event == "installation" {
		installation := github.InstallationEvent{}
		if err := json.Unmarshal(payloadBytes, &installation); err != nil {
			return nil, err
		}

		repos = installation.Repositories
	} else if event == "installation_repositories" {
		installation := github.InstallationRepositoriesEvent{}
		if err := json.Unmarshal(payloadBytes, &installation); err != nil {
			return nil, err
		}

		repos = installation.RepositoriesAdded
	}

	return repos, nil
}

// Prepare processes existing pull requests and calculates points estimates versus actual
func (b *bnkd) Prepare(p backend.Payload) error {
	log.Printf("prepare payload bytes: %s\n", string(p.Bytes()))

	repos, err := parseRepos(p.Type(), p.Bytes())
	if err != nil {
		return errors.New("error unmarshalling installation event: " + err.Error())
	}

	log.Printf("repositories: %+v\n", repos)
	for _, repo := range repos {
		log.Printf("repository: %s\n", *repo.FullName)
		fullName := strings.Split(*repo.FullName, "/")
		pullRequests, err := b.help.pullRequests(b.client, fullName[0], fullName[1])
		if err != nil {
			return errors.New("error getting pull requests: " + err.Error())
		}

		log.Printf("pull requests: %+v\n", pullRequests)
		for _, pr := range pullRequests {
			log.Printf("pull request: %+v\n", pr)
			closed := pr.ClosedAt
			merged := *pr.Merged
			log.Printf("closed: %s, merged: %t\n", closed, merged)
			if closed != nil && merged {
				if err := b.help.comment(b.client, fullName[0], fullName[1], pr); err != nil {
					return errors.New("error posting comment: " + err.Error())
				}
				log.Println("posting pull request comment")
			}
		}
	}

	log.Println("successful prepare invocation")
	return nil
}

// Act processes new pull requests and calculates points estimates versus actual
func (b *bnkd) Act(p backend.Payload) error {
	log.Printf("act payload bytes: %s\n", string(p.Bytes()))
	if p.Type() != "pull_request" {
		log.Printf("type %s not supported for pr estimation\n", p.Type())
		return nil
	}

	event := github.PullRequestEvent{}
	if err := json.Unmarshal(p.Bytes(), &event); err != nil {
		return fmt.Errorf("error unmarshaling event: %s", err.Error())
	}

	action := *event.Action
	merged := *event.PullRequest.Merged
	fullName := strings.Split(*event.Repo.FullName, "/")
	log.Printf("action: %s, merged: %t, repository: %s\n", action, merged, *event.Repo.FullName)
	if action == "closed" && merged {
		if err := b.help.comment(b.client, fullName[0], fullName[1], event.PullRequest); err != nil {
			return errors.New("error posting comment: " + err.Error())
		}
	}

	log.Println("successful act invocation")
	return nil
}

func main() {}
