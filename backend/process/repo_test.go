package process

import (
	"testing"

	"heupr/backend/response"
)

func ptrInt64(input int64) *int64 {
	return &input
}

func Test_parseResponses(t *testing.T) {
	tests := []struct {
		desc string
		repo *Repo
		sets *Settings
		id   int64
		pass bool
	}{
		{"empty repo method pointer", &Repo{}, &Settings{}, int64(0), false},
		{
			"incorrect response name",
			&Repo{},
			&Settings{
				Issues: map[string]map[string]response.Options{
					"opened": map[string]response.Options{
						"fakename": response.Options{},
					},
				},
			},
			int64(0),
			false,
		},
	}

	for i := range tests {
		if err := tests[i].repo.parseResponses(tests[i].sets, tests[i].id); (err == nil) != tests[i].pass {
			t.Errorf("test #%v desc: %v, error: %v", i+1, tests[i].desc, err)
		}
	}
}

func Test_newRepo(t *testing.T) {
	tests := []struct {
		desc string
		sets *Settings
		intg *Integration
		expt *Repo
		pass bool
	}{
		{"no repo id in settings", &Settings{}, &Integration{}, nil, false},
		// {"no repo found on server", &setting{Integration: integration{repoID: ptrInt64(66)}}, nil, false},
	}

	for i := range tests {
		_, err := NewRepo(tests[i].sets, tests[i].intg)
		if (err == nil) != tests[i].pass {
			t.Errorf("test #%v desc: %v, error: %v", i+1, tests[i].desc, err)
		}
	}
}
