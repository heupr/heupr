package process

import (
	"reflect"
	"testing"
)

func Test_start(t *testing.T) {
	w := &worker{
		work:    make(chan *Work),
		workers: make(chan chan *Work),
		repos: &Repos{
			Internal: make(map[int64]*Repo),
		},
		quit: make(chan bool),
	}

	tests := []struct {
		desc string
		wk   *Work
		expt *Repos
		pass bool
	}{
		{"empty work and worker arguments", &Work{}, &Repos{Internal: make(map[int64]*Repo)}, false},
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
		work  map[int64]*Work
		count int
	}{
		{"empty work map", nil, 0},
		{
			"one work object",
			map[int64]*Work{
				int64(94): &Work{},
			},
			1,
		},
		{
			"two work objects",
			map[int64]*Work{
				int64(5555): &Work{},
				int64(4907): &Work{},
			},
			2,
		},
	}

	for i := range tests {
		queue := make(chan *Work, 100)
		Collector(tests[i].work, queue)
		exp, rec := tests[i].count, len(queue)
		if exp != rec {
			t.Errorf("test #%v desc: %v, channel length expected %v, received %v", i+1, tests[i].desc, exp, rec)
		}
	}
}
