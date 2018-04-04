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

func Test_collector(t *testing.T) {
	tests := []struct {
		desc  string
		work  map[int64]*work
		count int
	}{
		{"empty work map", nil, 0},
		{
			"one work object",
			map[int64]*work{
				int64(94): &work{
					repoID: int64(94),
				},
			},
			1,
		},
	}

	for i := range tests {
		collector(tests[i].work)
		exp, rec := tests[i].count, len(workQueue)
		if exp != rec {
			t.Errorf("test #%v desc: %v, expected %v, received %v", i+1, tests[i].desc, exp, rec)
		}
	}
}
