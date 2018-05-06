package ingestor

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/github"
)

// openDatabase is designed to be overridden in unit testing.
var openDatabase = func() (dataAccess, error) {
	db, err := sql.Open("mysql", "root@/heupr?interpolateParams=true&parseTime=true")
	if err != nil {
		return nil, err
	}
	pool := newPool()
	return &database{sqlDB: db, pool: pool}, nil
}

type newClient func(int64, int64) githubService

func newGitHubClient(appID, installationID int64) githubService {
	var key string
	if PROD {
		key = "heupr.2017-10-04.private-key.pem"
	} else {
		key = "mikeheuprtest.2017-11-16.private-key.pem"
	}
	// Wrap the shared transport for use with the Github Installation.
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, int(appID), int(installationID), key)
	if err != nil {
		_ = err
	}
	c := github.NewClient(&http.Client{Transport: itr})
	return &client{githubClient: c}
}

// githubService encapsulates the GitHub client library methods.
type githubService interface {
	getRepoByID(id int64) (*github.Repository, error)
}

type client struct {
	githubClient *github.Client
}

func (c *client) getRepoByID(id int64) (*github.Repository, error) {
	repo, _, err := c.githubClient.Repositories.GetByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (c *client) getIssues(owner, repo, state string) ([]*github.Issue, error) {
	issues := []*github.Issue{}

	opt := &github.IssueListByRepoOptions{
		State: state,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		i, resp, err := c.githubClient.Issues.ListByRepo(context.Background(), owner, repo, opt)
		if err != nil {
			return nil, fmt.Errorf("addRepo get issues error: %v", err)
		}
		issues = append(issues, i...)
		if resp.NextPage == 0 {
			break
		} else {
			opt.ListOptions.Page = resp.NextPage
		}
	}

	return issues, nil
}

func (c *client) getPulls(owner, repo, state string) ([]*github.PullRequest, error) {
	pulls := []*github.PullRequest{}

	opt := &github.PullRequestListOptions{
		State: state,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		p, resp, err := c.githubClient.PullRequests.List(context.Background(), owner, repo, opt)
		if err != nil {
			return nil, fmt.Errorf("addRepo get pulls error %v", err)
		}
		pulls = append(pulls, p...)

		if resp.NextPage == 0 {
			break
		} else {
			opt.ListOptions.Page = resp.NextPage
		}
	}

	return pulls, nil
}

// Server holds assets necessary for listening to and processing GitHub events.
type Server struct {
	server   http.Server
	database dataAccess
}

// Start begins server listening.
func (s *Server) Start() {
	db, err := openDatabase()
	if err != nil {
		// TODO: Log error here.
	}
	s.database = db

}

// Stop provides graceful server shutdown.
func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.server.Shutdown(ctx)
	// NOTE: Include successful log shutdown here
}
