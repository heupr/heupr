package backend

import (
	"reflect"
	"testing"
)

type startTestDB struct {
	intg map[int64]*integration
	sets map[int64]*settings
	evts map[int64][]*container
}

func (s *startTestDB) readIntegrations(query string) (map[int64]*integration, error) {
	return s.intg, nil
}

func (s *startTestDB) readSettings(query string) (map[int64]*settings, error) {
	return s.sets, nil
}

func (s *startTestDB) readEvents(query string) (map[int64][]*container, error) {
	return s.evts, nil
}

/*
func TestStart(t *testing.T) {
	// TODO: For each test pass in true to the close map right away.
	s := &Server{}

	tests := []struct {
		desc string
		intg map[int64]*integration
		sets map[int64]*settings
		repo *repo
		err  error
		expt int
	}{
		{"all database tables empty", nil, nil, nil, nil, 0},
		{
			"single repo added",
			map[int64]*integration{
				int64(66): &integration{
					RepoID: int64(66),
				},
			},
			map[int64]*settings{
				int64(66): &settings{},
			},
			&repo{},
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
		newRepo = func(set *settings, intg *integration) (*repo, error) {
			return tests[i].repo, tests[i].err
		}

		s.Start()

		exp, rec := tests[i].expt, len(s.repos.Internal)
		if exp != rec {
			t.Errorf("test #%v desc: %v, internal map expected length %v, received %v", i+1, tests[i].desc, exp, rec)
		}
	}
}
*/

func Test_tick(t *testing.T) {
	dispatcher = func(r *repos, workQueue chan *work, workerQueue chan chan *work) {}

	result := make(map[int64]*work)
	holder := collector
	collector = func(wk map[int64]*work, workQueue chan *work) {
		result = wk
	}

	tests := []struct {
		desc string
		intg map[int64]*integration
		sets map[int64]*settings
		evts map[int64][]*container
		expt map[int64]*work
	}{
		{"no values returned from database", nil, nil, nil, make(map[int64]*work)},
		{
			"single integration value in database",
			map[int64]*integration{
				int64(50): &integration{
					RepoID: int64(50),
				},
			},
			nil,
			nil,
			map[int64]*work{
				int64(50): &work{
					RepoID: int64(50),
					integration: &integration{
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
			work:    make(chan *work),
			workers: make(chan chan *work),
			repos:   &repos{},
		}

		ender := make(chan bool)
		s.tick(ender)
		ender <- true

		exp, rec := tests[i].expt, result
		if !reflect.DeepEqual(exp, rec) {
			t.Errorf("test #%v desc: %v, expected map %v, received %v", i+1, tests[i].desc, exp, rec)
		}
	}

	collector = holder
}
