package backend

import (
	"heupr/backend/response/preprocess"
)

type work struct {
	repoID      int64
	integration integration
	setting     settings
	events      []*preprocess.Container
}

type worker struct {
	id    int
	work  chan work
	queue chan chan work
	repos map[int64]*repo
	quit  chan bool
}

func (w *worker) start() {
	go func() {
		for {
			w.queue <- w.work
			select {
			case newWork := <-w.work:
				_ = newWork
				// evaluate incoming work
				// if integration present
				// - initialize new repo w/ client and assets into s.repos
				// if settings present + if repo key exists
				// - set pathways + initialize responses
				// if events present + if repo key exists
				// - check to see if new integration is present in work struct
				// - pass events into training/predicting as needed
			case <-w.quit:
				return
			}
		}
	}()
}
