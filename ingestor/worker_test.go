package ingestor

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/go-github/github"
)

type processClient struct {
	repo        *github.Repository
	issues      []*github.Issue
	pulls       []*github.PullRequest
	tomlContent string
	err         error
}

func (pc *processClient) getRepoByID(id int64) (*github.Repository, error) {
	return pc.repo, pc.err
}

func (pc *processClient) getIssues(owner, repo, state string) ([]*github.Issue, error) {
	return pc.issues, pc.err
}

func (pc *processClient) getPulls(owner, repo, state string) ([]*github.PullRequest, error) {
	return pc.pulls, pc.err
}

func (pc *processClient) getTOML(owner, repo string) (string, error) {
	return pc.tomlContent, pc.err
}

type processDB struct {
	intg *integration
	err  error
}

func (p *processDB) InsertRepositoryIntegration(appID, repoID, installationID int64) {
	p.intg = &integration{
		AppID:          appID,
		RepoID:         repoID,
		InstallationID: installationID,
	}
}

func (p *processDB) DeleteRepositoryIntegration(appID, repoID, installationID int64) {}

func (p *processDB) ObliterateIntegration(appID, installationID int64) {}

func (p *processDB) ReadIntegrationByRepoID(repoID int64) (*integration, error) {
	return p.intg, p.err
}

func (p *processDB) InsertBulkIssues(issues []*github.Issue) {}

func (p *processDB) InsertBulkPullRequests(pulls []*github.PullRequest) {}

func (p *processDB) InsertTOML(content string) {}

type processRepoInit struct {
	resp bool
}

func (p *processRepoInit) addRepo(repo *github.Repository, c githubService) {}

func (p *processRepoInit) repoIntegrationExists(repoID int64) bool {
	return p.resp
}

func int64Ptr(i int64) *int64 {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

func Test_processHeuprInstallation(t *testing.T) {
	tests := []struct {
		desc string
		evnt heuprInstallationEvent
		repo *github.Repository
		err  error
		expt *integration
	}{
		{
			desc: "installation event with no repositories",
			evnt: heuprInstallationEvent{
				Action: stringPtr("added"),
				Installation: &heuprInstallation{
					ID:    int64Ptr(1),
					AppID: int64Ptr(1),
				},
				Repositories: []heuprRepository{},
			},
			repo: nil,
			err:  nil,
			expt: nil,
		},
		{
			desc: "getRepoByID returning error",
			evnt: heuprInstallationEvent{
				Action: stringPtr("added"),
				Installation: &heuprInstallation{
					ID:    int64Ptr(2),
					AppID: int64Ptr(2),
				},
				Repositories: []heuprRepository{
					heuprRepository{
						ID: int64Ptr(3),
					},
				},
			},
			repo: nil,
			err:  errors.New("test getRepoByID error"),
			expt: nil,
		},
		// [X] heuprInstallationEvent w/ no repositories
		// [ ] getRepoByID returning error
		// [ ] repoIntegrationExists returns true
		// [ ] successful pass to InsertRepositoryIntegration
		// [ ] successful pass to ObliterateIntegration
	}

	for i, tc := range tests {
		w := &worker{
			database: &processDB{},
			repoInit: &processRepoInit{},
		}

		f := func(appID, installationID int64) githubService {
			return &processClient{
				repo: tc.repo,
				err:  tc.err,
			}
		}

		w.processHeuprInstallation(tc.evnt, f)

		exp := tc.expt
		rec := w.database.(*processDB).intg
		if !reflect.DeepEqual(rec, exp) {
			t.Errorf("test #%v desc: %v, worker database expected %+v, has %+v", i+1, tc.desc, exp, rec)
		}

	}
}
