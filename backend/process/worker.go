package process

import (
	"heupr/backend/response"
	"heupr/backend/process/preprocess"
)

type Integration struct {
	InstallationID int64
	AppID          int
	RepoID         int64
}

type Setting struct {
	Title  string
	Issues map[string]map[string]response.Options
}

type Work struct {
	RepoID      int64
	Setting     *Setting
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
				if wk.Integration != nil && wk.Setting != nil {
					r, err := NewRepo(wk.Setting, wk.Integration)
					_ = err // TODO: Log this result.
					w.repos.Lock()
					w.repos.Internal[wk.RepoID] = r
					w.repos.Unlock()
				} else if wk.Integration == nil && wk.Setting != nil {
					w.repos.RLock()
					r, ok := w.repos.Internal[wk.RepoID]
					w.repos.RUnlock()
					_ = ok // TODO: Log this result.
					// TODO: This might need a lock as well?
					err := r.parseSettings(wk.Setting, wk.RepoID)
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

var Dispatcher = func(r *Repos, workQueue chan *Work, workerQueue chan chan *Work) {
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

var Collector = func(wk map[int64]*Work, workQueue chan *Work) {
	if len(wk) != 0 {
		for _, w := range wk {
			workQueue <- w
		}
	}
}
