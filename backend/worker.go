package backend

import (
	"heupr/backend/response/preprocess"
)

type work struct {
	repoID      int64
	setting     *setting
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
				if wk.integration != nil && wk.setting != nil {
					r, err := newRepo(wk.setting, wk.integration)
					_ = err // TODO: Log this result.
					w.repos.Lock()
					w.repos.internal[wk.repoID] = r
					w.repos.Unlock()
				} else if wk.integration == nil && wk.setting != nil {
					w.repos.RLock()
					r, ok := w.repos.internal[wk.repoID]
					w.repos.RUnlock()
					_ = ok // TODO: Log this result.
					// TODO: This might need a lock as well?
					err := r.parseSettings(wk.setting, wk.repoID)
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

var dispatcher = func(r *repos, workQueue chan *work, workerQueue chan chan *work) {
	for i := 0; i < cap(workerQueue); i++ {
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
}

var collector = func(wk map[int64]*work, workQueue chan *work) {
	if len(wk) != 0 {
		for _, w := range wk {
			workQueue <- w
		}
	}
}
