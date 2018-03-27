package backend

import (
	"testing"

	"heupr/backend/response"
)

func ptrInt64(input int64) *int64 {
	return &input
}

func Test_parseSettings(t *testing.T) {
	tests := []struct {
		desc string
		repo *repo
		sets *settings
		pass bool
	}{
		{"empty repo method pointer", &repo{}, &settings{}, false},
		{
			"incorrect response name",
			&repo{},
			&settings{
				Issues: map[string]map[string]response.Options{
					"opened": map[string]response.Options{
						"fakename": response.Options{},
					},
				},
				Integration: integration{repoID: ptrInt64(65)},
			},
			false,
		},
	}

	for i := range tests {
		if err := tests[i].repo.parseSettings(tests[i].sets); (err == nil) != tests[i].pass {
			t.Errorf("test #%v desc: %v, error: %v", i+1, tests[i].desc, err)
		}
	}
}

func Test_newRepo(t *testing.T) {
	s := &Server{
		repos: map[int64]*repo{
			int64(66): &repo{},
		},
	}

	tests := []struct {
		desc string
		sets *settings
		expt *repo
		pass bool
	}{
		{"no repo id in settings", &settings{}, nil, false},
		// {"no repo found on server", &settings{Integration: integration{repoID: ptrInt64(66)}}, nil, false},
	}

	for i := range tests {
		_, err := s.newRepo(tests[i].sets)
		if (err == nil) != tests[i].pass {
			t.Errorf("test #%v desc: %v, error: %v", i+1, tests[i].desc, err)
		}
	}
}
