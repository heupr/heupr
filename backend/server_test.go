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
		evts map[int64][]*preprocess.Container
		expt int
	}{
		{"all database tables empty", nil, nil, nil, 0},
	}

	for i := range tests {
		openDatabase = func() (dataAccess, error) {
			db := &startTestDB{
				intg: tests[i].intg,
				sets: tests[i].sets,
				evts: tests[i].evts,
			}
			return db, nil
		}

		s.Start()

		if len(s.repos.internal) != tests[i].expt {
			t.Errorf("test #%v desc: %v", i+1, tests[i].desc)
		}
	}
}

func Test_timer(t *testing.T) {}
