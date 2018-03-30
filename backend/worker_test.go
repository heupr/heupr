package backend

import (
	"reflect"
	"testing"
)

func Test_start(t *testing.T) {
	w := &worker{
		work:    make(chan *work),
		workers: make(chan chan *work),
		repos:   make(map[int64]*repo),
		quit:    make(chan bool),
	}

	defer close(w.work)
	defer close(w.workers)
	defer close(w.quit)

	tests := []struct {
		desc string
		wk   *work
		expt map[int64]*repo
		pass bool
	}{
		{"blank work and worker arguments", new(work), make(map[int64]*repo), false},
	}

	for i := range tests {
		w.start()

		go func() {
			w.work <- tests[i].wk
			w.quit <- true
		}()

		if !reflect.DeepEqual(w.repos, tests[i].expt) {
			t.Error("no match")
		}
	}
}
