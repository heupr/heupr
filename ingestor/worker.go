package ingestor

import (
	"context"

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
			client := newClient(*e.Installation.AppID, int(*e.Installation.ID))
			for i := 0; i < len(e.Repositories); i++ {
				repo, _, err := client.Repositories.GetByID(context.Background(), *e.Repositories[i].ID)
				if err != nil {
					_ = err
				}
				if w.repoIntegrationExists(*repo.ID) {
					return
				}

				go w.addRepo(repo, client)
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

}

func (w *worker) addRepo(repo *github.Repository, client *github.Client) {

}

func (w *worker) repoIntegrationExists(repoID int64) bool {
	// TODO: More shit goes here; copy-paste from legacy code.
	return true
}
