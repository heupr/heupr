package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/google/go-github/v28/github"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func Test_helpers(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/kamino/tipoca/pulls", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"number":1,"labels":[{"name":"est-1"}]}]`)
	})

	mux.HandleFunc("/repos/kamino/tipoca/pulls/1/commits", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"sha":"1"}]`)
	})

	mux.HandleFunc("/repos/kamino/tipoca/issues/1/comments", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"id":1}`)
	})

	mux.HandleFunc("/repos/tatooine/mos-eisley/pulls/1/commits", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"commit":{"author":{"date":"1977-05-25T20:00:00Z"}}},{"commit":{"author":{"date":"2005-05-19T20:00:00Z"}}}]`)
	})

	mux.HandleFunc("/repos/tatooine/mos-eisley/issues/1/comments", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"id":1}`)
	})

	server := httptest.NewServer(mux)

	c := github.NewClient(nil)
	url, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	c.BaseURL = url
	c.UploadURL = url

	prs := []*github.PullRequest{}
	h := &help{}
	t.Run("test get pull requests", func(t *testing.T) {
		prs, err = h.pullRequests(c, "kamino", "tipoca")
		if err != nil {
			t.Errorf("description: bulk pull request retrieval error, received: %s", err.Error())
		}

		if len(prs) != 1 {
			t.Errorf("description: incorrect pull request slice, received: %v", prs)
		}
	})

	t.Run("test get commit", func(t *testing.T) {
		cmts, err := h.commits(c, "kamino", "tipoca", 1)
		if err != nil {
			t.Errorf("description: bulk commit retrieval error, received: %s", err.Error())
		}

		if len(cmts) != 1 {
			t.Errorf("description: incorrect commit slice, received: %v", cmts)
		}
	})

	t.Run("test apply comment single commit", func(t *testing.T) {
		if err := h.comment(c, "kamino", "tipoca", prs[0]); err != nil {
			t.Errorf("description: pull request comment error, received: %s", err.Error())
		}
	})

	t.Run("test apply comment multiple commits", func(t *testing.T) {
		if err := h.comment(c, "tatooine", "mos-eisley", prs[0]); err != nil {
			t.Errorf("description: pull request comment error, received: %s", err.Error())
		}
	})
}

func TestConfigure(t *testing.T) {
	b := Backend
	c := github.NewClient(nil)
	b.Configure(c)

	if b.client == nil {
		t.Error("description: error creating backend")
	}

	if b.help == nil {
		t.Error("description: error creating backend")
	}
}

type mockHelp struct {
	pullRequestsOutput []*github.PullRequest
	pullRequestsErr    error
	commitsOutput      []*github.RepositoryCommit
	commitsErr         error
	commentErr         error
}

func (mock *mockHelp) pullRequests(c *github.Client, owner, repo string) ([]*github.PullRequest, error) {
	return mock.pullRequestsOutput, mock.pullRequestsErr
}

func (mock *mockHelp) commits(c *github.Client, owner, repo string, number int) ([]*github.RepositoryCommit, error) {
	return mock.commitsOutput, mock.commitsErr
}

func (mock *mockHelp) stringPtr(input string) *string {
	return &input
}

func (mock *mockHelp) comment(c *github.Client, owner, repo string, pr *github.PullRequest) error {
	return mock.commentErr
}

type mockPayload struct {
	payload     string
	payloadType string
}

func (mock *mockPayload) Type() string {
	return mock.payloadType
}

func (mock *mockPayload) Bytes() []byte {
	return []byte(mock.payload)
}

func (mock *mockPayload) Config() []byte {
	return []byte("")
}

func boolPtr(input bool) *bool {
	return &input
}

func timePtr(input time.Time) *time.Time {
	return &input
}

func TestPrepare(t *testing.T) {
	tests := []struct {
		desc               string
		payload            string
		payloadType        string
		pullRequestsOutput []*github.PullRequest
		pullRequestsErr    error
		commentErr         error
		err                string
	}{
		{
			desc:               "error parsing payload bytes",
			payload:            "incorrect-payload",
			payloadType:        "",
			pullRequestsOutput: nil,
			pullRequestsErr:    nil,
			commentErr:         nil,
			err:                "error unmarshalling installation event: invalid character 'i' looking for beginning of value",
		},
		{
			desc:               "error parsing payload bytes",
			payload:            `{"repositories_added":[{"full_name": "test-name/test-login"}]}`,
			payloadType:        "pull_request",
			pullRequestsOutput: nil,
			pullRequestsErr:    errors.New("mock get pull requests error"),
			commentErr:         nil,
			err:                "error getting pull requests: mock get pull requests error",
		},
		{
			desc:        "error commenting on pull request",
			payload:     `{"repositories_added":[{"full_name": "test-name/test-login"}]}`,
			payloadType: "pull_request",

			pullRequestsOutput: []*github.PullRequest{
				{
					ClosedAt: timePtr(time.Now()),
					Merged:   boolPtr(true),
				},
			},
			pullRequestsErr: nil,
			commentErr:      errors.New("mock comment error"),
			err:             "error posting comment: mock comment error",
		},
		{
			desc:        "successful invocation",
			payload:     `{"repositories_added":[{"full_name": "test-name/test-login"}]}`,
			payloadType: "pull_request",

			pullRequestsOutput: []*github.PullRequest{
				{
					ClosedAt: timePtr(time.Now()),
					Merged:   boolPtr(true),
				},
			},
			pullRequestsErr: nil,
			commentErr:      nil,
			err:             "",
		},
	}

	for _, test := range tests {
		p := &mockPayload{
			payload:     test.payload,
			payloadType: test.payloadType,
		}

		h := &mockHelp{
			pullRequestsOutput: test.pullRequestsOutput,
			pullRequestsErr:    test.pullRequestsErr,
			commentErr:         test.commentErr,
		}

		b := Backend
		b.help = h

		err := b.Prepare(p)
		if err != nil && err.Error() != test.err {
			t.Errorf("description: %s, error received: %s", test.desc, err.Error())
		}
	}
}

func TestAct(t *testing.T) {
	tests := []struct {
		desc       string
		payload    string
		commentErr error
		err        string
	}{
		{
			desc:       "error parsing payload bytes",
			payload:    "incorrect-payload",
			commentErr: nil,
			err:        "error unmarshaling event: invalid character 'i' looking for beginning of value",
		},
		{
			desc:       "error commenting on pull request",
			payload:    `{"action":"closed","pull_request":{"merged":true},"repository":{"full_name": "test-owner/test-login"}}`,
			commentErr: errors.New("mock comment error"),
			err:        "error posting comment: mock comment error",
		},
		{
			desc:       "successful invocation",
			payload:    `{"action":"closed","pull_request":{"merged":true},"repository":{"full_name": "test-owner/test-login"}}`,
			commentErr: nil,
			err:        "",
		},
	}

	for _, test := range tests {
		p := &mockPayload{
			payload: test.payload,
		}

		h := &mockHelp{
			commentErr: test.commentErr,
		}

		b := Backend
		b.help = h

		err := b.Act(p)
		if err != nil && err.Error() != test.err {
			t.Errorf("description: %s, error received: %s, expected: %s", test.desc, err.Error(), test.err)
		}
	}
}

func Test_main(t *testing.T) {
	main() // invoking for test coverage
}
