package itoi

import (
	"encoding/json"
	"testing"

	"heupr/backend/process/preprocess"
)

func TestPreprocess(t *testing.T) {
	tests := []struct {
		desc string
		cnts []*preprocess.Container
		lgth []int
		pass bool
	}{
		{"empty input container", []*preprocess.Container{}, []int{0}, false},
		{
			"empty json payload",
			[]*preprocess.Container{
				&preprocess.Container{
					Event:   "issues",
					Payload: json.RawMessage(``),
				},
			},
			[]int{0},
			false,
		},
		{
			"passing single issue payload",
			[]*preprocess.Container{
				&preprocess.Container{
					Event:   "issues",
					Payload: json.RawMessage(`{"action": "opened","issue":{"id": 1}}`),
				},
			},
			[]int{0},
			true,
		},
		{
			"passing issue and pull request",
			[]*preprocess.Container{
				&preprocess.Container{
					Event:   "issues",
					Payload: json.RawMessage(`{"action": "opened", "issue":{"id": 2}}`),
				},
				&preprocess.Container{
					Event:   "pull_request",
					Payload: json.RawMessage(`{"action": "opened", "pull_request":{"id":3,"title":"test title", "body":"test body"}}`),
				},
			},
			[]int{0, 0},
			true,
		},
		// scenarios:
		// [X] container with empty values
		// [X] container with empty issue payload
		// [X] container with issue event
		// [ ] container with two objects no reference
		// [ ] container with two issues one reference
		// [ ] container with three issues two references (chain)
	}

	p := &P{}

	for i := range tests {
		output, err := p.Preprocess(tests[i].cnts)

		if tests[i].pass != (err == nil) {
			t.Errorf("test %v desc: %v, error: %v", i+1, tests[i].desc, err)
			for k, v := range output {
				rec := len(v.Linked)
				exp := tests[i].lgth[k]
				if exp != rec {
					t.Errorf("test %v, container %v linked length %v, expected %v", i+1, k+1, rec, exp)
				}
			}
		}
	}
}
