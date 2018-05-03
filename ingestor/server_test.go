package ingestor

import "github.com/google/go-github/github"

type testClient struct {
	repo *github.Repository
	err  error
}

func (tc *testClient) getRepoByID(id int64) (*github.Repository, error) {
	return tc.repo, tc.err
}

func init() {
	newClient = func(appID, installationID int64) githubService {
		return &testClient{}
	}
}
