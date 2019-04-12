package api

import (
	"errors"
	"log"
	"strings"
	"testing"
)

func Test_newProcessor(t *testing.T) {
	tests := []struct {
		desc string
		name string
		err  bool
	}{
		{
			"no process name provided",
			"",
			false,
		},
		{
			"process.test name provided",
			"process.test",
			true,
		},
	}

	for _, test := range tests {
		p, err := newProcessor(test.name)

		if (err != nil) == test.err {
			t.Errorf("description: %s, error returned creating process: %s", test.desc, err.Error())
		}

		if (p == nil) == test.err {
			t.Errorf("description: %s, process created incorrectly: %+v", test.desc, p)
		}
	}
}

func Test_insertMultiple(t *testing.T) {
	tests := []struct {
		desc  string
		input []byte
		err   error
		ftrs  []*feature
	}{
		{
			"non-object byte array",
			[]byte("a"),
			errors.New("error unmarshalling input: invalid character 'a' looking for beginning of value"),
			nil,
		},
		{
			"incorrect input object type",
			[]byte("{}"),
			errors.New("error unmarshalling input: json: cannot unmarshal object into Go value of type []*api.object"),
			nil,
		},
		{
			"object missing number",
			[]byte("[{}]"),
			errors.New("no input object number"),
			nil,
		},
		{
			"passing object array input",
			[]byte(`[{"object_number": 123, "object_body": "Here is a body with a #321 reference"}]`),
			nil,
			[]*feature{
				&feature{
					Number:     intPtr(123),
					References: []*int{intPtr(321)},
				},
			},
		},
	}

	p, err := newProcessor("process.insertmultiple.test")
	if err != nil {
		log.Fatalf("error creating processor: %s", err.Error())
	}

	for _, test := range tests {
		ftrs, err := p.insertMultiple(test.input)

		if test.err != nil || err != nil {
			if !strings.Contains(err.Error(), test.err.Error()) {
				t.Errorf("description: %s, incorrect error message, received: %s, expected: %s", test.desc, err.Error(), test.err)
			}
		}

		if ftrs != nil {
			for i, ftr := range ftrs {
				if !(*ftr.Number == *test.ftrs[i].Number) {
					t.Errorf("description: %s, object numbers do not match, received: %d, expected: %d", test.desc, *ftr.Number, *test.ftrs[i].Number)
				}

				if !(*ftr.References[0] == *test.ftrs[i].References[0]) {
					t.Errorf("description: %s, reference numbers do not match, received: %v, expected: %v", test.desc, ftr.References, test.ftrs[i].References)
				}
			}
		}
	}
}

func Test_insertSingle(t *testing.T) {
	tests := []struct {
		desc  string
		input []byte
		err   error
		ftr   *feature
	}{
		{
			"non-object byte array",
			[]byte("a"),
			errors.New("error unmarshalling input: invalid character 'a' looking for beginning of value"),
			nil,
		},
		{
			"object missing number",
			[]byte("{}"),
			errors.New("no input object number"),
			nil,
		},
		{
			"passing object input",
			[]byte(`{"object_number": 123, "object_body": "Here is a body with a #321 reference"}`),
			nil,
			&feature{
				Number: intPtr(123),
				References: []*int{
					intPtr(321),
				},
			},
		},
	}

	p, err := newProcessor("process.insertsingle.test")
	if err != nil {
		log.Fatalf("error creating processor: %s", err.Error())
	}

	for _, test := range tests {
		ftr, err := p.insertSingle(test.input)

		if test.err != nil || err != nil {
			if !strings.Contains(err.Error(), test.err.Error()) {
				t.Errorf("description: %s, incorrect error message, received: %s, expected: %s", test.desc, err.Error(), test.err)
			}
		}

		if ftr != nil {
			if !(*ftr.Number == *test.ftr.Number) {
				t.Errorf("description: %s, object numbers do not match, received: %d, expected: %d", test.desc, *ftr.Number, *test.ftr.Number)
			}

			if !(*ftr.References[0] == *test.ftr.References[0]) {
				t.Errorf("description: %s, reference numbers do not match, received: %v, expected: %v", test.desc, ftr.References, test.ftr.References)
			}
		}
	}
}
