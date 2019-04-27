package backend

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		desc string
		err  string
		url  string
	}{
		{
			"no url environment variable",
			"no elasticsearch url set",
			"",
		},
		{
			"passing environment variable scenario",
			"asdf",
			"non-url",
		},
	}

	for _, test := range tests {
		if test.url != "" {
			os.Setenv("HEUPR_ES_URL", test.url)
		}

		b, err := New()
		if err != nil {
			if !strings.Contains(err.Error(), test.err) {
				t.Errorf("description: %s, incorrect error message, received: %s, expected: %s", test.desc, err.Error(), test.err)
			}
		}

		if b != nil {
			if b.esURL != test.url {
				t.Errorf("description: %s, url field incorrectly set, received: %s, expected: %s", test.desc, b.esURL, test.url)
			}
		}

		os.Unsetenv("HEUPR_ES_URL")
	}
}

func stringPtr(input string) *string {
	return &input
}

func intPtr(input int) *int {
	return &input
}

func Test_getText(t *testing.T) {
	tests := []struct {
		desc   string
		ref    int
		ftrs   []*feature
		output string
	}{
		{
			"no match in feature",
			66,
			[]*feature{
				&feature{
					Number: intPtr(65),
					Title:  stringPtr("example feature title"),
					Body:   stringPtr("this text contains #64 as well as #67"),
				},
			},
			"",
		},
		{
			"two matches in features",
			94,
			[]*feature{
				&feature{
					Number: intPtr(94),
					Title:  stringPtr("another feature title"),
					Body:   stringPtr("another reference is #94"),
				},
			},
			" feature title   reference ",
		},
	}

	for _, test := range tests {
		output := getText(test.ref, test.ftrs)
		if output != test.output {
			t.Errorf("description: %s, output does not match, received: %s, expected: %s", test.desc, output, test.output)
		}
	}
}

type testLearnClient struct {
	resp string
	sErr error
	iErr error
}

func (c *testLearnClient) search(actor string) (string, error) {
	return c.resp, c.sErr
}

func (c *testLearnClient) index(key, value string) error {
	return c.iErr
}

func TestLearn(t *testing.T) {
	tests := []struct {
		desc  string
		input []byte
		resp  string
		err   error
		sErr  error
		iErr  error
	}{
		{
			"incorrect input byte object",
			[]byte("{}"),
			"",
			errors.New("error unmarshaling: json: cannot unmarshal object into Go value of type []*backend.feature"),
			nil,
			nil,
		},
		{
			"error response from search",
			[]byte(`[{"feature_actor_name": "test-name", "feature_title": "test title", "feature_body": "test body"}]`),
			"",
			errors.New("error: mock search"),
			errors.New("error: mock search"),
			nil,
		},
		{
			"error response from index",
			[]byte(`[{"feature_actor_name": "test-name", "feature_title": "test title", "feature_body": "test body"}]`),
			"",
			errors.New("error: mock index"),
			nil,
			errors.New("error: mock index"),
		},
		{
			"passing scenario",
			[]byte(`[{"feature_actor_name": "test-name", "feature_title": "test title", "feature_body": "test body"}]`),
			"successful learn",
			nil,
			nil,
			nil,
		},
	}

	for _, test := range tests {
		client := &testLearnClient{
			resp: test.resp,
			sErr: test.sErr,
			iErr: test.iErr,
		}

		b := &Backend{
			esURL:  "test-es-url",
			client: client,
		}

		output, err := b.Learn(test.input)

		if test.err != nil || err != nil {
			if !strings.Contains(err.Error(), test.err.Error()) {
				t.Errorf("description: %s, incorrect error message, received: %s, expected: %s", test.desc, err.Error(), test.err)
			}
		}

		if output != nil && string(output) != test.resp {
			t.Errorf("description: %s, incorrect response, received: %s, expected: %s", test.desc, string(output), test.resp)
		}
	}
}

func TestPredict(t *testing.T) {
	tests := []struct {
		desc   string
		input  []byte
		output []byte
		err    error
		sResp  string
		sErr   error
	}{
		{
			"incorrect input type",
			[]byte("[]"),
			nil,
			errors.New("error unmarshalling: json: cannot unmarshal array into Go value of type backend.feature"),
			"",
			nil,
		},
		{
			"error response from search",
			[]byte(`{"feature_actor_name": "test-name", "feature_title": "test title", "feature_body": "test body"}`),
			nil,
			errors.New("error searching: error: mock search"),
			"",
			errors.New("error: mock search"),
		},
		{
			"passing scenario",
			[]byte(`{"feature_actor_name": "test-name", "feature_title": "test title", "feature_body": "test body"}`),
			[]byte("test-name"),
			nil,
			"test-name",
			nil,
		},
	}

	for _, test := range tests {
		client := &testLearnClient{
			resp: test.sResp,
			sErr: test.sErr,
		}

		b := &Backend{
			esURL:  "test-es-url",
			client: client,
		}

		output, err := b.Predict(test.input)

		if test.err != nil || err != nil {
			if !strings.Contains(err.Error(), test.err.Error()) {
				t.Errorf("description: %s, incorrect error message, received: %s, expected: %s", test.desc, err.Error(), test.err)
			}
		}

		if output != nil && string(output) != string(test.output) {
			t.Errorf("description: %s, incorrect response, received: %s, expected: %s", test.desc, string(output), test.output)
		}
	}
}
