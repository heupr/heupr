package backend

import (
	"reflect"
	"testing"
	// "time"
	// "sync"

	"heupr/backend/process"
	"heupr/backend/process/preprocess"
)

type startTestDB struct {
	intg map[int64]*preprocess.Integration
	sets map[int64]*preprocess.Settings
	evts map[int64][]*preprocess.Container
}

func (s *startTestDB) readIntegrations(query string) (map[int64]*preprocess.Integration, error) {
	return s.intg, nil
}

func (s *startTestDB) readSettings(query string) (map[int64]*preprocess.Settings, error) {
	return s.sets, nil
}

func (s *startTestDB) readEvents(query string) (map[int64][]*preprocess.Container, error) {
	return s.evts, nil
}

func TestStart(t *testing.T) {
	// TODO: For each test pass in true to the close map right away.
	s := &Server{}

	tests := []struct {
		desc string
		intg map[int64]*preprocess.Integration
		sets map[int64]*preprocess.Settings
		repo *process.Repo
		err  error
		expt int
	}{
		{"all database tables empty", nil, nil, nil, nil, 0},
		{
			"single repo added",
			map[int64]*preprocess.Integration{
				int64(66): &preprocess.Integration{
					RepoID: int64(66),
				},
			},
			map[int64]*preprocess.Settings{
				int64(66): &preprocess.Settings{},
			},
			&process.Repo{},
			nil,
			1,
		},
	}

	for i := range tests {
		openDatabase = func() (dataAccess, error) {
			db := &startTestDB{
				intg: tests[i].intg,
				sets: tests[i].sets,
			}
			return db, nil
		}
		newRepo = func(set *preprocess.Settings, intg *preprocess.Integration) (*process.Repo, error) {
			return tests[i].repo, tests[i].err
		}

		s.Start()

		exp, rec := tests[i].expt, len(s.repos.Internal)
		if exp != rec {
			t.Errorf("test #%v desc: %v, internal map expected length %v, received %v", i+1, tests[i].desc, exp, rec)
		}
	}
}

func Test_tick(t *testing.T) {
	dispatcher = func(r *process.Repos, workQueue chan *preprocess.Work, workerQueue chan chan *preprocess.Work) {}

	result := make(map[int64]*preprocess.Work)
	collector = func(wk map[int64]*preprocess.Work, workQueue chan *preprocess.Work) {
		result = wk
	}

	tests := []struct {
		desc string
		intg map[int64]*preprocess.Integration
		sets map[int64]*preprocess.Settings
		evts map[int64][]*preprocess.Container
		expt map[int64]*preprocess.Work
	}{
		{"no values returned from database", nil, nil, nil, make(map[int64]*preprocess.Work)},
		{
			"single integration value in database",
			map[int64]*preprocess.Integration{
				int64(50): &preprocess.Integration{
					RepoID: int64(50),
				},
			},
			nil,
			nil,
			map[int64]*preprocess.Work{
				int64(50): &preprocess.Work{
					RepoID: int64(50),
					Integration: &preprocess.Integration{
						RepoID: int64(50),
					},
				},
			},
		},
	}

	for i := range tests {
		s := &Server{
			database: &startTestDB{
				intg: tests[i].intg,
				sets: tests[i].sets,
				evts: tests[i].evts,
			},
			work:    make(chan *preprocess.Work),
			workers: make(chan chan *preprocess.Work),
			repos:   &process.Repos{},
		}

		ender := make(chan bool)
		s.tick(ender)
		ender <- true

		exp, rec := tests[i].expt, result
		if !reflect.DeepEqual(tests[i].expt, result) {
			t.Errorf("test #%v desc: %v, expected map %v, received %v", i+1, tests[i].desc, exp, rec)
		}
	}
}
