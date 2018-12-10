package backend

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
					w.repos.Lock()
					w.repos.Internal[wk.RepoID] = r
					w.repos.Unlock()
				} else if wk.integration == nil && wk.settings != nil {
					w.repos.RLock()
					r, ok := w.repos.Internal[wk.RepoID]
					w.repos.RUnlock()
					_ = ok // TODO: Log this result.
					// TODO: This might need a lock as well?
					err := r.parseResponses(wk.settings, wk.RepoID)
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

// dispatcher initializes and starts workers to receive incoming work.
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

// collector distributes new work objects to active workers.
var collector = func(wk map[int64]*work, workQueue chan *work) {
	if len(wk) != 0 {
		for _, w := range wk {
			workQueue <- w
		}
	}
}
