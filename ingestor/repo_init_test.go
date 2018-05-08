package ingestor

import (
	"errors"
	"testing"

	"github.com/google/go-github/github"
)

type repoInitDB struct {
	intg        *integration
	issues      []*github.Issue
	pulls       []*github.PullRequest
	tomlContent string
	err         error
}

func (r *repoInitDB) InsertRepositoryIntegration(appID, repoID, installationID int64) {}

func (r *repoInitDB) DeleteRepositoryIntegration(appID, repoID, installationID int64) {}

func (r *repoInitDB) ObliterateIntegration(appID, installationID int64) {}

func (r *repoInitDB) ReadIntegrationByRepoID(repoID int64) (*integration, error) {
	return r.intg, r.err
}

func (r *repoInitDB) InsertBulkIssues(issues []*github.Issue) {
	r.issues = issues
}

func (r *repoInitDB) InsertBulkPullRequests(pulls []*github.PullRequest) {
	r.pulls = pulls
}

func (r *repoInitDB) InsertTOML(content string) {
	r.tomlContent = content
}

type repoInitClient struct {
	repo        *github.Repository
	issues      []*github.Issue
	pulls       []*github.PullRequest
	tomlContent string
	err         error
}

func (r *repoInitClient) getRepoByID(id int64) (*github.Repository, error) {
	return r.repo, r.err
}

func (r *repoInitClient) getIssues(owner, repo, state string) ([]*github.Issue, error) {
	return r.issues, r.err
}

func (r *repoInitClient) getPulls(owner, repo, state string) ([]*github.PullRequest, error) {
	return r.pulls, r.err
}

func (r *repoInitClient) getTOML(owner, repo string) (string, error) {
	return r.tomlContent, r.err
}

func Test_addRepo(t *testing.T) {
	tests := []struct {
		desc        string
		owner       string
		repo        string
		issues      []*github.Issue
		pulls       []*github.PullRequest
		tomlContent string
		err         error
	}{
		{
			desc:  "repo with no issues/pulls",
			owner: "grand-moff-tarink",
			repo:  "death-star",
		},
		{
			desc:  "repo with issues/pulls getters returning errors",
			owner: "uncle-owen-and-aunt-beru",
			repo:  "lars-moisture-farm",
			err:   errors.New("getter returning error"),
		},
		{
			desc:  "repo with issues and no pull requests",
			owner: "chalmun",
			repo:  "chalmuns-cantina",
			issues: []*github.Issue{
				&github.Issue{},
			},
			err: nil,
		},
		// TODO: Other possible scenarios:
		// [ ] repo with issue pass + pull err
		// [ ] repo with pull pass + issue err
		// [ ] repo with issue/pull pass
	}

	for i, tc := range tests {
		r := &repoInit{
			database: &repoInitDB{},
		}
		c := &repoInitClient{
			issues: tc.issues,
			pulls:  tc.pulls,
			err:    tc.err,
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
	}
}
