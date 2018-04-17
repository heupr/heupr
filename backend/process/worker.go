package process

import (
	"heupr/backend/process/preprocess"
	"heupr/backend/response"
)

// Integration represents a new Heupr GitHub integration.
type Integration struct {
	InstallationID int64
	AppID          int
	RepoID         int64
}

// Settings represents the parsed .heupr.toml file for user settings.
type Settings struct {
	Title  string
	Issues map[string]map[string]response.Options
}

// Work holds the objects necessary for processing by the responses.
type Work struct {
	RepoID      int64
	Settings    *Settings
	Integration *Integration
	Events      []*preprocess.Container
}

type worker struct {
	id      int
	work    chan *Work
	workers chan chan *Work
	repos   *Repos
	quit    chan bool
}

func (w *worker) start() {
	go func() {
		for {
			w.workers <- w.work
			select {
			case wk := <-w.work:
				if wk.Integration != nil && wk.Settings != nil {
					r, err := NewRepo(wk.Settings, wk.Integration)
					_ = err // TODO: Log this result.
					w.repos.Lock()
					w.repos.Internal[wk.RepoID] = r
					w.repos.Unlock()
				} else if wk.Integration == nil && wk.Settings != nil {
					w.repos.RLock()
					r, ok := w.repos.Internal[wk.RepoID]
					w.repos.RUnlock()
					_ = ok // TODO: Log this result.
					// TODO: This might need a lock as well?
					err := r.parseResponses(wk.Settings, wk.RepoID)
					_ = err // TODO: Log this result.
				}

				if len(wk.Events) != 0 {
					// call predict on events
				}
			case <-w.quit:
				return
			}
		}
	}()
}

// Dispatcher initializes and starts workers to receive incoming work.
func Dispatcher(r *Repos, workQueue chan *Work, workerQueue chan chan *Work) {
	for i := 0; i < cap(workerQueue); i++ {
		w := &worker{
			id:      i + 1,
			work:    make(chan *Work),
			workers: workerQueue,
			repos:   r,
			quit:    make(chan bool),
		}
		w.start()
	}

	go func() {
		for {
			wk := <-workQueue
			go func() {
				wkrs := <-workerQueue
				wkrs <- wk
			}()
		}
	}()
}

// Collector distributes new work objects to active workers.
func Collector(wk map[int64]*Work, workQueue chan *Work) {
	if len(wk) != 0 {
		for _, w := range wk {
			workQueue <- w
		}
	}
}
