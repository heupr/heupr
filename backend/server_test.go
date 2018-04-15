package backend

import (
	"testing"
    "time"

	"heupr/backend/process"
	"heupr/backend/process/preprocess"
)

type startTestDB struct {
	intg map[int64]*process.Integration
	sets map[int64]*process.Setting
	evts map[int64][]*preprocess.Container
}

func (s *startTestDB) readIntegrations(query string) (map[int64]*process.Integration, error) {
	return s.intg, nil
}

func (s *startTestDB) readSettings(query string) (map[int64]*process.Setting, error) {
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
		intg map[int64]*process.Integration
		sets map[int64]*process.Setting
		repo *process.Repo
		err  error
		expt int
	}{
		{"all database tables empty", nil, nil, nil, nil, 0},
		{
			"single repo added",
			map[int64]*process.Integration{
				int64(66): &process.Integration{
					RepoID: int64(66),
				},
			},
			map[int64]*process.Setting{
				int64(66): &process.Setting{},
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
		newRepo = func(set *process.Setting, intg *process.Integration) (*process.Repo, error) {
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
	dispatcher = func(r *process.Repos, workQueue chan *process.Work, workerQueue chan chan *process.Work) {}

	result := make(map[int64]*process.Work)
	collector = func(wk map[int64]*process.Work, workQueue chan *process.Work) {
		result = wk
	}

	tests := []struct {
		desc string
		intg map[int64]*process.Integration
		sets map[int64]*process.Setting
		evts map[int64][]*preprocess.Container
	}{
		{"test1", nil, nil, nil},
		{
			"test2",
			map[int64]*process.Integration{
				int64(50): &process.Integration{
					RepoID: int64(50),
				},
			},
			nil,
			nil,
		},
	}

	workQueue := make(chan *process.Work)
	workerQueue := make(chan chan *process.Work)

	for i := range tests {
		ender := make(chan bool)
		s := &Server{
			database: &startTestDB{
				intg: tests[i].intg,
				sets: tests[i].sets,
				evts: tests[i].evts,
			},
			repos: &process.Repos{},
		}

		s.tick(ender, workQueue, workerQueue)
        time.Sleep(time.Second * 5) // TEMPORARY

		t.Error(result)
	}
}
