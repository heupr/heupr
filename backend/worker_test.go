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
			Internal: make(map[int64]*repo),
		},
		quit: make(chan bool),
	}

	tests := []struct {
		desc string
		wk   *work
		expt *repos
		pass bool
	}{
		{"empty work and worker arguments", &work{}, &repos{Internal: make(map[int64]*repo)}, false},
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
				int64(94): &work{},
			},
			1,
		},
		{
			"two work objects",
			map[int64]*work{
				int64(5555): &work{},
				int64(4907): &work{},
			},
			2,
		},
	}

	for i := range tests {
		queue := make(chan *work, 100)
		collector(tests[i].work, queue)
		exp, rec := tests[i].count, len(queue)
		if exp != rec {
			t.Errorf("test #%v desc: %v, channel length expected %v, received %v", i+1, tests[i].desc, exp, rec)
		}
	}
}
