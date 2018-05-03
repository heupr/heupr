package ingestor

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/go-github/github"
)

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

func int64Ptr(i int64) *int64 {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

func Test_processHeuprInstallation(t *testing.T) {
	tests := []struct {
		desc       string
		evt        heuprInstallationEvent
		repo       *github.Repository
		getRepoErr error
		expt       *integration
	}{
		{
			"getRepoByID returning error",
			heuprInstallationEvent{
				Action: stringPtr("added"),
				Installation: &heuprInstallation{
					ID:    int64Ptr(1),
					AppID: int64Ptr(2),
				},
			},
			nil,
			errors.New("getRepoByID error"),
			nil,
		},
		// [ ] repoIntegrationExists returns true
		// [ ] successful pass to InsertRepositoryIntegration
		// [ ] successful pass to ObliterateIntegration
	}

	for i, tc := range tests {
		w := &worker{
			database: &processDB{},
		}

		w.processHeuprInstallation(tc.evt)

		exp := tc.expt
		rec := w.database.(*processDB).intg
		if !reflect.DeepEqual(rec, exp) {
			t.Errorf("test #%v desc: %v, worker database expected %+v, has %+v", i+1, tc.desc, exp, rec)
		}

	}
}
