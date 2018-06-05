package process

import (
	"reflect"
	"testing"

	"heupr/backend/preprocess"
)

func Test_start(t *testing.T) {
	w := &worker{
		work:    make(chan *preprocess.Work),
		workers: make(chan chan *preprocess.Work),
		repos: &Repos{
			Internal: make(map[int64]*Repo),
		},
		quit: make(chan bool),
	}

	tests := []struct {
		desc string
		wk   *preprocess.Work
		expt *Repos
		pass bool
	}{
		{"empty work and worker arguments", &preprocess.Work{}, &Repos{Internal: make(map[int64]*Repo)}, false},
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
		work  map[int64]*preprocess.Work
		count int
	}{
		{"empty work map", nil, 0},
		{
			"one work object",
			map[int64]*preprocess.Work{
				int64(94): &preprocess.Work{},
			},
			1,
		},
		{
			"two work objects",
			map[int64]*preprocess.Work{
				int64(5555): &preprocess.Work{},
				int64(4907): &preprocess.Work{},
			},
			2,
		},
	}

	for i := range tests {
		queue := make(chan *preprocess.Work, 100)
		Collector(tests[i].work, queue)
		exp, rec := tests[i].count, len(queue)
		if exp != rec {
			t.Errorf("test #%v desc: %v, channel length expected %v, received %v", i+1, tests[i].desc, exp, rec)
		}
	}
}
