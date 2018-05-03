package ingestor

import (
	"database/sql"

	"github.com/google/go-github/github"
)

type worker struct {
	id       int
	database dataAccess
	work     chan interface{}
	workers  chan chan interface{}
	quit     chan bool
}

func newWorker(id int, db dataAccess, queue chan chan interface{}) *worker {
	return &worker{
		id:       id,
		database: db,
		work:     make(chan interface{}),
		workers:  queue,
		quit:     make(chan bool),
	}
}

func (w *worker) processHeuprInstallation(event heuprInstallationEvent) {
	go func(e heuprInstallationEvent) {
		switch *e.Action {
		case "created":
			c := newClient(*e.Installation.AppID, *e.Installation.ID)
			for i := 0; i < len(e.Repositories); i++ {
				repo, err := c.getRepoByID(*e.Repositories[i].ID)
				if err != nil {
					_ = err
				}
				if w.repoIntegrationExists(*repo.ID) {
					return
				}

				go w.addRepo(repo, c.(*client).githubClient)
				// TODO: Add logging indicating successfully added a repo.

				w.database.InsertRepositoryIntegration(*e.Installation.AppID, *repo.ID, *e.Installation.ID)

				// integration, err := w.database.ReadIntegrationByRepoID(*repo.ID)
				// if err != nil {
				// 	_ = err
				// 	return
				// }
			}
		case "deleted":
			w.database.ObliterateIntegration(*e.Installation.AppID, *e.Installation.ID)
		}
	}(event)
}

func (w *worker) processRepoInstallation(event repoInstallationEvent) {
	go func(e repoInstallationEvent) {
		switch *e.Action {
		case "added":
			repos := make([]heuprRepository, len(e.RepositoriesAdded))
			for i := 0; i < len(repos); i++ {
				repos[i] = heuprRepository{
					ID:       e.RepositoriesAdded[i].ID,
					Name:     e.RepositoriesAdded[i].Name,
					FullName: e.RepositoriesAdded[i].FullName,
				}
			}
			// installationEvent := heuprInstallationEvent{
			// 	Action:       e.Action,
			// 	Sender:       e.Sender,
			// 	Installation: e.Installation,
			// 	Repositories: repos,
			// }

			c := newClient(*e.Installation.AppID, *e.Installation.ID)
			for i := 0; i < len(e.RepositoriesAdded); i++ {
				repo, err := c.getRepoByID(*e.RepositoriesAdded[i].ID)
				if err != nil {
					_ = err
					return
				}
				if w.repoIntegrationExists(*repo.ID) {
					return
				}
				go w.addRepo(repo, c.(*client).githubClient)

				w.database.InsertRepositoryIntegration(*e.Installation.AppID, *repo.ID, *e.Installation.ID)
			}
		case "removed":
			// client := newClient(*e.Installation.AppID, int(*e.Installation.ID))
			for i := 0; i < len(e.RepositoriesRemoved); i++ {
				repo := e.RepositoriesRemoved[i]
				if !w.repoIntegrationExists(*repo.ID) {
					return
				}
				w.database.DeleteRepositoryIntegration(*e.Installation.AppID, *repo.ID, *e.Installation.ID)
			}

		}
	}(event)
}

func (w *worker) addRepo(repo *github.Repository, client *github.Client) {

}

func (w *worker) repoIntegrationExists(repoID int64) bool {
	_, err := w.database.ReadIntegrationByRepoID(repoID)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		_ = err
		return false
	default:
		return true
	}
}
