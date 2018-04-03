package backend

import (
	"heupr/backend/response/preprocess"
)

var (
	workQueue   = make(chan *work, 100)
	workerQueue chan chan *work
)

type work struct {
	repoID      int64
	settings    *settings
	integration *integration
	events      []*preprocess.Container
}

type worker struct {
	id      int
	work    chan *work
	workers chan chan *work
	repos   *repos
	quit    chan bool
}

func (w *worker) start() {
	go func() {
		for {
			w.workers <- w.work
			select {
			case wk := <-w.work:
				if wk.integration != nil && wk.settings != nil {
					r, err := newRepo(wk.settings, wk.integration)
					_ = err // TODO: Log this result.
					// w.repos.RLock()
					w.repos.Lock()
					w.repos.internal[wk.repoID] = r
					w.repos.Unlock()
				} else if wk.integration == nil && wk.settings != nil {
					w.repos.RLock()
					r, ok := w.repos.internal[wk.repoID]
					w.repos.RUnlock()
					_ = ok // TODO: Log this result.
					// TODO: This might need a lock as well?
					err := r.parseSettings(wk.settings, wk.repoID)
					_ = err // TODO: Log this result.
				}

				if len(wk.events) != 0 {
					// call predict on events
				}
			case <-w.quit:
				return
			}
		}
	}()
}

func dispatcher(r *repos, count int) error {
	workerQueue = make(chan chan *work, count)

	for i := 0; i < count; i++ {
		w := &worker{
			id:      i + 1,
			work:    make(chan *work),
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

	return nil
}

var collector = func(wk map[int64]*work) {
	if len(wk) != 0 {
		for _, w := range wk {
			workQueue <- w
		}
	}
}
