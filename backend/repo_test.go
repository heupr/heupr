package backend

import (
	"testing"
)

func Test_newRepo(t *testing.T) {}

func Test_parseSettings(t *testing.T) {
	tests := []struct {
		desc string
		repo *repo
		sets *settings
		pass bool
	}{
		{"empty repo method pointer", &repo{}, &settings{}, true},
	}

	for i := range tests {
		if err := tests[i].repo.parseSettings(tests[i].sets); (err == nil) != tests[i].pass {
			t.Errorf("test #%v desc: %v, error: %v", i+1, tests[i].desc, err)
		}
		// NOTE: Evaluating repo field population can go here.
	}
}
