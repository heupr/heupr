package ingestor

import (
	"errors"
	"testing"

	"github.com/google/go-github/github"
)

type repoInitClient struct {
	issues    []*github.Issue
	issuesErr error
	pulls     []*github.PullRequest
	pullsErr  error
	toml      string
	tomlErr   error
}

func (r *repoInitClient) getRepoByID(id int64) (*github.Repository, error) {
	return nil, nil
}

func (r *repoInitClient) getIssues(owner, repo, state string) ([]*github.Issue, error) {
	return r.issues, r.issuesErr
}

func (r *repoInitClient) getPulls(owner, repo, state string) ([]*github.PullRequest, error) {
	return r.pulls, r.pullsErr
}

func (r *repoInitClient) getTOML(owner, repo string) (string, error) {
	return r.toml, r.tomlErr
}

type repoInitDB struct {
	intg      *integration
	issues    []*github.Issue
	issuesErr error
	pulls     []*github.PullRequest
	pullsErr  error
	toml      string
	tomlErr   error
}

func (r *repoInitDB) InsertRepositoryIntegration(appID, repoID, installationID int64) {}

func (r *repoInitDB) DeleteRepositoryIntegration(appID, repoID, installationID int64) {}

func (r *repoInitDB) ObliterateIntegration(appID, installationID int64) {}

func (r *repoInitDB) ReadIntegrationByRepoID(repoID int64) (*integration, error) {
	return nil, nil
}

func (r *repoInitDB) InsertBulkIssues(issues []*github.Issue) {
	r.issues = issues
}

func (r *repoInitDB) InsertBulkPullRequests(pulls []*github.PullRequest) {
	r.pulls = pulls
}

func (r *repoInitDB) InsertTOML(content string) {
	r.toml = content
}

func Test_addRepo(t *testing.T) {
	err := errors.New("test error value")
	tests := []struct {
		desc      string
		owner     string
		repo      string
		issues    []*github.Issue
		issuesErr error
		pulls     []*github.PullRequest
		pullsErr  error
		toml      string
		tomlErr   error
	}{
		{
			desc:  "repo with no issues/pulls",
			owner: "grand-moff-tarink",
			repo:  "death-star",
		},
		{
			desc:      "all getters returning errors",
			owner:     "uncle-owen-and-aunt-beru",
			repo:      "lars-moisture-farm",
			issuesErr: err,
			pullsErr:  err,
			tomlErr:   err,
		},
		{
			desc:  "repo with issues/toml and no pull requests",
			owner: "chalmun",
			repo:  "chalmuns-cantina",
			issues: []*github.Issue{
				&github.Issue{
					Title: stringPtr("No droids!"),
				},
			},
			issuesErr: nil,
			toml:      "example toml content",
			tomlErr:   nil,
		},
	}

	for i, tc := range tests {
		r := &repoInit{
			database: &repoInitDB{},
		}
		c := &repoInitClient{
			issues:    tc.issues,
			issuesErr: tc.issuesErr,
			pulls:     tc.pulls,
			pullsErr:  tc.pullsErr,
			toml:      tc.toml,
			tomlErr:   tc.tomlErr,
		}

		r.addRepo(tc.owner, tc.repo, c)
		rec, exp := len(r.database.(*repoInitDB).issues), len(tc.issues)
		if rec != exp {
			t.Errorf("test #%v desc: %v, database issues length %v, expected %v", i+1, tc.desc, rec, exp)
		}
		rec, exp = len(r.database.(*repoInitDB).pulls), len(tc.pulls)
		if rec != exp {
			t.Errorf("test #%v desc: %v, database pulls length %v, expected %v", i+1, tc.desc, rec, exp)
		}
		rec, exp = len(r.database.(*repoInitDB).toml), len(tc.toml)
		if rec != exp {
			t.Errorf("test #%v desc: %v, database toml length %v, expected %v", i+1, tc.desc, rec, exp)
		}
	}
}
