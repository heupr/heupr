package api

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func Test_newIndex(t *testing.T) {
	tests := []struct {
		desc string
		name string
		err  bool
	}{
		{
			"no index name provided",
			"",
			false,
		},
		{
			"index.test name provided",
			"index.test",
			true,
		},
	}

	for _, test := range tests {
		i, err := newIndex(test.name)

		if (err != nil) == test.err {
			t.Errorf("description: %s, error returned creating index: %s", test.desc, err.Error())
		}

		if (i == nil) == test.err {
			t.Errorf("description: %s, index created incorrectly: %+v", test.desc, i)
		}
	}
}

func intPtr(input int) *int {
	return &input
}

func Test_create(t *testing.T) {
	indexName := "index.create.test"

	tests := []struct {
		desc string
		ftr  *feature
		id   string
		pass bool
	}{
		{
			"empty struct input",
			&feature{},
			"",
			false,
		},
		{
			"successful create invocation",
			&feature{
				Number: intPtr(66),
			},
			"66",
			true,
		},
	}

	i, err := newIndex(indexName)
	if err != nil {
		t.Fatalf("error creating test index create: %s", err.Error())
	}

	for _, test := range tests {
		if err := i.create(test.ftr); (err != nil) == test.pass {
			t.Errorf("description: %s, index create method error: %s", test.desc, err)
		}

		if test.pass {
			idx := i.(index)
			doc, err := idx.store.Document(test.id)
			if err != nil {
				t.Fatalf("error retrieving test index values: %s", err.Error())
			}

			if !strings.Contains(doc.GoString(), test.id) {
				t.Errorf("description: %s, id value not in indexed document, received: %s, expected: %s", test.desc, doc.GoString(), test.id)
			}
		}
	}
}

func Test_read(t *testing.T) {
	tests := []struct {
		desc string
		id   string
		ftr  *feature
		pass bool
		err  string
	}{
		{
			"feature not in index",
			"66",
			nil,
			false,
			"no feature found in index",
		},
		{
			"successful read invocation",
			"5555",
			&feature{
				Number: intPtr(5555),
			},
			true,
			"asdf",
		},
	}

	i, err := newIndex("index.read.test")
	if err != nil {
		t.Fatalf("error creating test index create: %s", err.Error())
	}

	for _, test := range tests {
		idx := i.(index)

		if test.ftr != nil {
			idx.store.Index(strconv.Itoa(*test.ftr.Number), test.ftr)
		}

		ftr, err := i.read(test.id)
		if err != nil {
			if !strings.Contains(err.Error(), test.err) {
				t.Errorf("description: %s, wrong error message, received: %s, expected: %s", test.desc, err.Error(), test.err)
			}
		}

		_ = ftr
		if test.pass {
			if !reflect.DeepEqual(*ftr, *test.ftr) {
				t.Errorf("description: %s, received incorrect feature, received: %+v, expected: %+v", test.desc, *test.ftr, *ftr)
			}
		}
	}
}

func Test_delete(t *testing.T) {
	tests := []struct {
		desc string
		id   string
		ftr  *feature
		err  string
	}{
		{
			"empty id string argument",
			"",
			nil,
			"id must be non-empty string",
		},
		{
			"successful delete invocation",
			"94",
			&feature{
				Number: intPtr(94),
			},
			"",
		},
	}

	i, err := newIndex("index.delete.test")
	if err != nil {
		t.Fatalf("error creating test index create: %s", err.Error())
	}

	for _, test := range tests {
		idx := i.(index)

		if test.ftr != nil {
			idx.store.Index(strconv.Itoa(*test.ftr.Number), test.ftr)
		}

		err := i.delete(test.id)
		if err != nil {
			if !strings.Contains(err.Error(), test.err) {
				t.Errorf("description: %s, incorrect error message, received: %s, expected: %s", test.desc, err.Error(), test.err)
			}
		}
	}
}
