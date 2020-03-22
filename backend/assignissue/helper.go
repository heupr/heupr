package main

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/bbalet/stopwords"

	"github.com/google/go-github/v28/github"
)

type helper interface {
	getContent(c *github.Client, owner, repo, path string) (string, error)
	getText(issue *github.Issue) string
	listIssues(c *github.Client, owner, repo string) ([]*github.Issue, error)
	addAssignee(c *github.Client, owner, repo, actor string, number int) error
}

type help struct{}

func (h *help) getContent(c *github.Client, owner, repo, path string) (string, error) {
	opts := &github.RepositoryContentGetOptions{}
	file, _, _, err := c.Repositories.GetContents(context.Background(), owner, repo, path, opts)
	if err != nil {
		return "", errors.New("error getting content: " + err.Error())
	}

	content, err := file.GetContent()
	if err != nil {
		return "", errors.New("error parsing content: " + err.Error())
	}

	return content, nil
}

func (h *help) getText(issue *github.Issue) string {
	// NOTE: Add checks for empty title/body fields.
	title := stopwords.CleanString(strings.ToLower(*issue.Title), "en", false)
	body := stopwords.CleanString(strings.ToLower(*issue.Body), "en", false)

	return title + " " + body
}

func (h *help) addAssignee(c *github.Client, owner, repo, actor string, number int) error {
	_, _, err := c.Issues.AddAssignees(context.Background(), owner, repo, number, []string{actor})
	return err
}

func (h *help) listIssues(c *github.Client, owner, repo string) ([]*github.Issue, error) {
	output := []*github.Issue{}

	opts := &github.IssueListByRepoOptions{
		State:    "closed",
		Assignee: "*",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		issues, resp, err := c.Issues.ListByRepo(context.Background(), owner, repo, opts)
		if err != nil {
			return nil, err
		}
		output = append(output, issues...)
		log.Printf("issues: %+v\n", issues)

		if resp.NextPage == 0 {
			break
		} else {
			opts.ListOptions.Page = resp.NextPage
		}
	}

	return output, nil
}
