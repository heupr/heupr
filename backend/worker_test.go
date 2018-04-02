package backend

import (
	"reflect"
	"testing"
)

func Test_start(t *testing.T) {
	w := &worker{
		work:    make(chan *work),
		workers: make(chan chan *work),
		repos: &repos{
			internal: make(map[int64]*repo),
		},
		quit: make(chan bool),
	}

	tests := []struct {
		desc string
		wk   *work
		expt *repos
		pass bool
	}{
		{"empty work and worker arguments", new(work), &repos{internal: make(map[int64]*repo)}, false},
	}

	for i := range tests {
		w.start()

		go func() {
			w.work <- tests[i].wk
			w.quit <- true
		}()

		if !reflect.DeepEqual(w.repos, tests[i].expt) {
			t.Error(tests[i].desc)
		}
	}
}
