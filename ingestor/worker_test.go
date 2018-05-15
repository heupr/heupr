package ingestor

/*
import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/go-github/github"
)

type processClient struct {
	repo *github.Repository
	err  error
}

func (pc *processClient) getRepoByID(id int64) (*github.Repository, error) {
	return pc.repo, pc.err
}

func (pc *processClient) getIssues(owner, repo, state string) ([]*github.Issue, error) {
	return nil, nil
}

func (pc *processClient) getPulls(owner, repo, state string) ([]*github.PullRequest, error) {
	return nil, nil
}

func (pc *processClient) getTOML(owner, repo string) (string, error) {
	return "", nil
}

type processDB struct {
	intg *integration
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
	return nil, nil
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
	err := errors.New("example test error")
	tests := []struct {
		desc string
		evnt heuprInstallationEvent
		repo *github.Repository
		err  error
		expt *integration
	}{
		{
			desc: "incorrect event action specified",
			evnt: heuprInstallationEvent{
				Action: stringPtr("test-action"),
			},
			expt: nil,
		},
		{
			desc: "get repo by id returning error",
			evnt: heuprInstallationEvent{
				Action: stringPtr("created"),
				Installation: &heuprInstallation{
					ID:    int64Ptr(1),
					AppID: int64Ptr(1),
				},
			},
			err: err,
		},
		{
			desc: "installation event with no repositories",
			evnt: heuprInstallationEvent{
				Action: stringPtr("created"),
				Installation: &heuprInstallation{
					ID:    int64Ptr(1),
					AppID: int64Ptr(1),
				},
				Repositories: []heuprRepository{},
			},
			err: nil,
		},

		// {
		// 	desc: "getRepoByID returning error",
		// 	evnt: heuprInstallationEvent{
		// 		Action: stringPtr("added"),
		// 		Installation: &heuprInstallation{
		// 			ID:    int64Ptr(2),
		// 			AppID: int64Ptr(2),
		// 		},
		// 		Repositories: []heuprRepository{
		// 			heuprRepository{
		// 				ID: int64Ptr(3),
		// 			},
		// 		},
		// 	},
		// 	repo: nil,
		// 	err:  errors.New("test getRepoByID error"),
		// 	expt: nil,
		// },
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
*/
