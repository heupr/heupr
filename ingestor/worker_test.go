package ingestor

import "testing"

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

func (p *processDB) DeleteRepositoryIntegration(appID int, repoID, installationID int64) {}

func (p *processDB) ObliterateIntegration(appID int, installationID int64) {}

func (p *processDB) ReadIntegrationByRepoID(repoID int64) (*integration, error) {
	return p.intg, p.err
}

func int64Ptr(i int64) *int64 {
	return &i
}

func Test_processHeuprInstallation(t *testing.T) {
	w := &worker{}
	added := "added"
	// removed := "removed"

	tests := []struct {
		evt heuprInstallationEvent
	}{
		{
			evt: heuprInstallationEvent{
				Action: &added,
				Installation: &heuprInstallation{
					ID:    int64Ptr(1),
					AppID: int64Ptr(2),
				},
			},
		},
	}

	for i := range tests {
		w.processHeuprInstallation(tests[i].evt)
	}

}
