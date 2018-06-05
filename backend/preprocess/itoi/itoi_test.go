package itoi

import (
	"encoding/json"
	"reflect"
	"testing"

	"heupr/backend/process/preprocess"
)

func Test_findIssueNumbers(t *testing.T) {
	tests := []struct {
		desc string
		text string
		expt []int
	}{
		{"no issue references", "test text body", []int{}},
		{"single issue reference", "Closes #1", []int{1}},
		{"multiple issue references", "Closes #1 and also Fixes #2", []int{1, 2}},
	}

	// The production slice keywords is used in this unit test.
	for i := range tests {
		output := findIssueNumbers(tests[i].text, keywords)
		if !reflect.DeepEqual(tests[i].expt, output) {
			t.Errorf("test %v desc: %v, expected %v, received %v", i+1, tests[i].desc, tests[i].expt, output)
		}
	}
}

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
					Payload: json.RawMessage(`{"action":"opened","issue":{"number":1}}`),
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
					Payload: json.RawMessage(`{"action":"opened","issue":{"number":2}}`),
				},
				&preprocess.Container{
					Event:   "pull_request",
					Payload: json.RawMessage(`{"action":"opened","pull_request":{"number":3,"title":"test title","body":"test body"}}`),
				},
			},
			[]int{0, 0},
			true,
		},
		{
			"one linked pull request",
			[]*preprocess.Container{
				&preprocess.Container{
					Event:   "issues",
					Payload: json.RawMessage(`{"action":"opened","issue":{"number":4}}`),
				},
				&preprocess.Container{
					Event:   "pull_request",
					Payload: json.RawMessage(`{"action":"opened","pull_request":{"number":5,"title":"closes issue","body":"Closes #4"}}`),
				},
			},
			[]int{1, 0},
			true,
		},
		{
			"one pull request referencing two issues",
			[]*preprocess.Container{
				&preprocess.Container{
					Event:   "issues",
					Payload: json.RawMessage(`{"action":"opened","issue":{"number":6}}`),
				},
				&preprocess.Container{
					Event:   "issues",
					Payload: json.RawMessage(`{"action":"opened","issue":{"number":7}}`),
				},
				&preprocess.Container{
					Event:   "pull_request",
					Payload: json.RawMessage(`{"action":"opened","pull_request":{"number":8,"title":"closes two issues","body":"This Fixes #6 and Fixes #7"}}`),
				},
			},
			[]int{1, 1, 0},
			true,
		},
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
