package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/go-yaml/yaml"
	"github.com/google/go-github/v28/github"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func Test_postHTTP(t *testing.T) {
	tests := []struct {
		desc string
		path string
		err  string
	}{
		{
			desc: "incorrect url provided to app",
			path: "not-endpoint",
			err:  "non-200 status code: 404",
		},
		{
			desc: "error from target url",
			path: "failure",
			err:  "non-200 status code: 500",
		},
		{
			desc: "successful invocation",
			path: "success",
			err:  "",
		},
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"message":"success"}`)
	})

	mux.HandleFunc("/failure", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"failure"}`, http.StatusInternalServerError)
	})

	server := httptest.NewServer(mux)

	url := server.URL

	h := &help{}
	for _, test := range tests {
		err := h.postHTTP(url+"/"+test.path, strings.NewReader(`{"message":"message"}`))
		if err != nil && err.Error() != test.err {
			t.Errorf("description: %s, error received: %s, expected: %s", test.desc, err.Error(), test.err)
		}
	}
}

func Test_parseMessage(t *testing.T) {
	tests := []struct {
		desc         string
		eventType    string
		eventPayload string
		msg          string
	}{
		{
			desc:         "create project type received",
			eventType:    "project",
			eventPayload: `{"sender":{"login":"bane"},"action":"created","project":{"name":"grand-plan"}}`,
			msg:          "user bane created project grand-plan",
		},
		{
			desc:         "create project column type received",
			eventType:    "project_column",
			eventPayload: `{"sender":{"login":"sidious"},"action":"created","project_column":{"name":"clone-wars"}}`,
			msg:          "user sidious created project column clone-wars",
		},
		{
			desc:         "create project card type received",
			eventType:    "project_card",
			eventPayload: `{"sender":{"login":"sidious"},"action":"created","project_card":{"id":66}}`,
			msg:          "user sidious created project card 66",
		},
		{
			desc:         "moved project card type received",
			eventType:    "project_card",
			eventPayload: `{"sender":{"login":"sidious"},"action":"moved","project_card":{"id":66}}`,
			msg:          "user sidious moved 66 from pre-war to clone-wars",
		},
		{
			desc:         "non-project event received",
			eventType:    "non-project",
			eventPayload: "",
			msg:          "",
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/projects/columns/cards/66", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"previous_column_name": "pre-war", "column_name": "clone-wars"}`)
	})

	server := httptest.NewServer(mux)

	c := github.NewClient(nil)
	url, _ := url.Parse(server.URL + "/")
	c.BaseURL = url
	c.UploadURL = url

	h := &help{
		client: c,
	}

	for _, test := range tests {
		msg := h.parseMessage(test.eventType, test.eventPayload)
		if msg != test.msg {
			t.Errorf("description: %s, message received: %s, expected: %s", test.desc, msg, test.msg)
		}
	}
}

type mockPayload struct {
	payloadBytes  string
	payloadType   string
	payloadConfig []byte
}

func (mock *mockPayload) Type() string {
	return mock.payloadType
}

func (mock *mockPayload) Bytes() []byte {
	return []byte(mock.payloadBytes)
}

func (mock *mockPayload) Config() []byte {
	return mock.payloadConfig
}

type mockHelp struct {
	postHTTPErr        error
	parseMessageOutput string
}

func (m *mockHelp) postHTTP(url string, body io.Reader) error {
	return m.postHTTPErr
}

func (m *mockHelp) parseMessage(eventType, payloadString string) string {
	return m.parseMessageOutput
}

func TestPrepare(t *testing.T) {
	p := &mockPayload{}
	b := Backend
	if err := b.Prepare(p); err != nil {
		t.Errorf("description: error calling prepare, error: %s", err.Error())
	}
}

func TestAct(t *testing.T) {
	tests := []struct {
		desc               string
		payloadBytes       string
		payloadType        string
		postHTTPErr        error
		parseMessageOutput string
		err                string
	}{
		{
			desc:               "non-project event received",
			payloadBytes:       ``,
			payloadType:        "non-project",
			postHTTPErr:        nil,
			parseMessageOutput: "",
			err:                "no output message",
		},
		{
			desc:               "project event received",
			payloadBytes:       ``,
			payloadType:        "project",
			postHTTPErr:        nil,
			parseMessageOutput: "output",
			err:                "",
		},
		{
			desc:               "project card event received",
			payloadBytes:       ``,
			payloadType:        "project_card",
			postHTTPErr:        nil,
			parseMessageOutput: "output",
			err:                "",
		},
		{
			desc:               "project column event received",
			payloadBytes:       ``,
			payloadType:        "project_column",
			postHTTPErr:        nil,
			parseMessageOutput: "output",
			err:                "",
		},
	}

	for _, test := range tests {
		c := configObj{
			Backends: []backendObj{
				backendObj{
					Name: "projectboard",
					Settings: settings{
						URLs: []string{
							"https://test.com",
						},
					},
				},
			},
		}

		config, err := yaml.Marshal(&c)
		if err != nil {
			t.Fatal(err)
		}

		p := &mockPayload{
			payloadBytes:  test.payloadBytes,
			payloadType:   test.payloadType,
			payloadConfig: config,
		}

		h := &mockHelp{
			postHTTPErr:        test.postHTTPErr,
			parseMessageOutput: test.parseMessageOutput,
		}

		b := Backend
		b.help = h

		err = b.Act(p)
		if err != nil && err.Error() != test.err {
			t.Errorf("description: %s, error received: %s, expected: %s", test.desc, err.Error(), test.err)
		}
	}
}
