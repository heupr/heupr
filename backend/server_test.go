package backend

import (
	"testing"

	"heupr/backend/response/preprocess"
)

type startTestDB struct {
	intg map[int64]*integration
	sets map[int64]*setting
	evts map[int64][]*preprocess.Container
}

func (s *startTestDB) readIntegrations(query string) (map[int64]*integration, error) {
	return s.intg, nil
}

func (s *startTestDB) readSettings(query string) (map[int64]*setting, error) {
	return s.sets, nil
}

func (s *startTestDB) readEvents(query string) (map[int64][]*preprocess.Container, error) {
	return s.evts, nil
}

func TestStart(t *testing.T) {
	s := &Server{}

	tests := []struct {
		desc string
		intg map[int64]*integration
		sets map[int64]*setting
		repo *repo
		err  error
		expt int
	}{
		{"all database tables empty", nil, nil, nil, nil, 0},
		{
			"single repo added",
			map[int64]*integration{
				int64(66): &integration{
					repoID: int64(66),
				},
			},
			map[int64]*setting{
				int64(66): &setting{},
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
		newRepo = func(set *setting, intg *integration) (*repo, error) {
			return tests[i].repo, tests[i].err
		}

		s.Start()

		exp, rec := tests[i].expt, len(s.repos.internal)
		if exp != rec {
			t.Errorf("test #%v desc: %v, internal map expected length %v, received %v", i+1, tests[i].desc, exp, rec)
		}
	}
}
