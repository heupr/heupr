package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v28/github"
)

func TestConfigure(t *testing.T) {
	b := &Backend{}
	c := github.NewClient(nil)
	b.Configure(c)

	if b.client == nil {
		t.Error("description: error creating backend")
	}
}

func Test_getContent(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/repos/mustafar/mining-facility/contents/confederacy-leadership.json", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"content":"deceased"}`)
	})

	server := httptest.NewServer(mux)

	c := github.NewClient(nil)
	url, _ := url.Parse(server.URL + "/")
	c.BaseURL = url
	c.UploadURL = url

	h := help{}

	output, err := h.getContent(c, "mustafar", "mining-facility", "confederacy-leadership.json")
	if output == "" {
		t.Errorf("description: error getting file contents, received: %s", output)
	}

	if err != nil {
		t.Errorf("description: error getting file contents, error: %s", err.Error())
	}
}

func stringPtr(input string) *string {
	return &input
}

func Test_getText(t *testing.T) {
	i := &github.Issue{
		Title: stringPtr("the tragedy of darth plagueis the wise"),
		Body:  stringPtr("not a story the jedi would tell "),
	}

	h := help{}

	output := h.getText(i)
	text := " tragedy darth plagueis wise   story jedi tell "
	if output != text {
		t.Errorf("description: error getting clean text, received: %s, expected: %s", output, text)
	}

}

func Test_newClient(t *testing.T) {
	es, err := newClient("https://mos-eisley.gov")
	if err != nil {
		t.Errorf("description: error creating es client, error received: %s", err.Error())
	}

	if es == nil {
		t.Errorf("description: error creating es client")
	}
}

func Test_listIssues(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/repos/hoth/echo-base/issues", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"title":"1","labels":[{"name":"est-1"}]}]`)
	})

	server := httptest.NewServer(mux)

	c := github.NewClient(nil)
	url, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	c.BaseURL = url
	c.UploadURL = url

	h := &help{}
	issues, err := h.listIssues(c, "hoth", "echo-base")
	if err != nil {
		t.Errorf("description: error listing issues, error received: %s", err.Error())
	}

	expected := 1
	if len(issues) != 1 {
		t.Errorf("description: error listing issues, issues received: %d, expected: %d", len(issues), expected)
	}
}

func Test_addAssignees(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/repos/lars-homestead/new-droids/issues/1/assignees", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"number":1,"assignees":[{"login":"luke"}]}`)
	})

	server := httptest.NewServer(mux)

	c := github.NewClient(nil)
	url, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	c.BaseURL = url
	c.UploadURL = url

	h := help{}
	if err := h.addAssignee(c, "lars-homestead", "new-droids", "luke", 1); err != nil {
		t.Errorf("description: error assigning contributor, error received: %s", err.Error())
	}
}

type mockPayload struct {
	payloadBytes string
	payloadType  string
}

func (m *mockPayload) Type() string {
	return m.payloadType
}

func (m *mockPayload) Bytes() []byte {
	return []byte(m.payloadBytes)
}

type mockHelp struct {
	listIssuesOutput []*github.Issue
	listIssuesErr    error
	getContentOutput string
	getContentErr    error
	getTextOutput    string
	addAssigneeErr   error
}

type mockES struct {
	esIndexOutput string
	esIndexErr    error
}

func (m *mockES) index(string, string) error {
	return m.esIndexErr
}

func (m *mockES) search(string) (string, error) {
	return m.esIndexOutput, m.esIndexErr
}

func (m *mockHelp) listIssues(c *github.Client, owner, repo string) ([]*github.Issue, error) {
	return m.listIssuesOutput, m.listIssuesErr
}

func (m *mockHelp) getContent(c *github.Client, owner, repo, path string) (string, error) {
	return m.getContentOutput, m.getContentErr
}

func (m *mockHelp) getText(issue *github.Issue) string {
	return m.getTextOutput
}

func (m *mockHelp) addAssignee(c *github.Client, owner, repo, actor string, number int) error {
	return m.addAssigneeErr
}

func TestPrepare(t *testing.T) {
	tests := []struct {
		desc             string
		payloadBytes     string
		getContentOutput string
		getContentErr    error
		listIssuesOutput []*github.Issue
		listIssuesErr    error
		newClientOutput  es
		newClientErr     error
		err              string
	}{
		{
			desc:             "incorrect event received",
			payloadBytes:     "[]",
			getContentOutput: "",
			getContentErr:    nil,
			listIssuesOutput: nil,
			listIssuesErr:    nil,
			newClientOutput:  nil,
			newClientErr:     nil,
			err:              "error unmarshalling installation: json: cannot unmarshal array into Go value of type github.InstallationRepositoriesEvent",
		},
		{
			desc:             "error getting config file content",
			payloadBytes:     `{"repositories_added":[{"owner":{"login":"delta-squad"},"name":"CC-1038"}]}`,
			getContentOutput: "",
			getContentErr:    errors.New("mock get content error"),
			listIssuesOutput: nil,
			listIssuesErr:    nil,
			newClientOutput:  nil,
			newClientErr:     nil,
			err:              "error getting heupr config: mock get content error",
		},
		{
			desc:             "error unmarshalling config file",
			payloadBytes:     `{"repositories_added":[{"owner":{"login":"delta-squad"},"name":"CC-1038"}]}`,
			getContentOutput: "-------",
			getContentErr:    nil,
			listIssuesOutput: nil,
			listIssuesErr:    nil,
			newClientOutput:  nil,
			newClientErr:     nil,
			err:              "error parsing heupr config: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `-------` into main.configObj",
		},
		{
			desc:             "error listing issues",
			payloadBytes:     `{"repositories_added":[{"owner":{"login":"delta-squad"},"name":"CC-1038"}]}`,
			getContentOutput: "",
			getContentErr:    nil,
			listIssuesOutput: nil,
			listIssuesErr:    errors.New("mock list issues error"),
			newClientOutput:  nil,
			newClientErr:     nil,
			err:              "error getting issues: mock list issues error",
		},
		{
			desc:             "error creating es client",
			payloadBytes:     `{"repositories_added":[{"owner":{"login":"delta-squad"},"name":"CC-1038"}]}`,
			getContentOutput: "",
			getContentErr:    nil,
			listIssuesOutput: []*github.Issue{
				&github.Issue{
					State: stringPtr("closed"),
					Assignee: &github.User{
						Login: stringPtr("CC-01/425"),
					},
					Title: stringPtr("field-advisor"),
					Body:  stringPtr("hologram only"),
				},
			},
			listIssuesErr:   nil,
			newClientOutput: nil,
			newClientErr:    errors.New("mock new client error"),
			err:             "error creating es client: mock new client error",
		},
		{
			desc:             "error indexing value",
			payloadBytes:     `{"repositories_added":[{"owner":{"login":"delta-squad"},"name":"CC-1038"}]}`,
			getContentOutput: "",
			getContentErr:    nil,
			listIssuesOutput: []*github.Issue{
				&github.Issue{
					State: stringPtr("closed"),
					Assignee: &github.User{
						Login: stringPtr("CC-01/425"),
					},
					Title: stringPtr("field-advisor"),
					Body:  stringPtr("hologram only"),
				},
			},
			listIssuesErr: nil,
			newClientOutput: &mockES{
				esIndexErr: errors.New("mock index error"),
			},
			newClientErr: nil,
			err:          "error indexing key/value: mock index error",
		},
		{
			desc:             "successful invocation",
			payloadBytes:     `{"repositories_added":[{"owner":{"login":"delta-squad"},"name":"CC-1038"}]}`,
			getContentOutput: "",
			getContentErr:    nil,
			listIssuesOutput: []*github.Issue{
				&github.Issue{
					State: stringPtr("closed"),
					Assignee: &github.User{
						Login: stringPtr("CC-01/425"),
					},
					Title: stringPtr("field-advisor"),
					Body:  stringPtr("hologram only"),
				},
			},
			listIssuesErr: nil,
			newClientOutput: &mockES{
				esIndexErr: nil,
			},
			newClientErr: nil,
			err:          "",
		},
	}

	for _, test := range tests {
		p := &mockPayload{
			payloadBytes: test.payloadBytes,
		}

		h := &mockHelp{
			listIssuesOutput: test.listIssuesOutput,
			listIssuesErr:    test.listIssuesErr,
			getContentOutput: test.getContentOutput,
			getContentErr:    test.getContentErr,
		}

		newClient = func(url string) (es, error) {
			return test.newClientOutput, test.newClientErr
		}

		b := Backend{
			help:   h,
			client: github.NewClient(nil),
		}

		err := b.Prepare(p)
		if err != nil && err.Error() != test.err {
			t.Errorf("description: %s, error received: %s, expected: %s", test.desc, err.Error(), test.err)
		}
	}
}

func TestAct(t *testing.T) {
	tests := []struct {
		desc             string
		payloadBytes     string
		payloadType      string
		getContentOutput string
		getContentErr    error
		newClientOutput  es
		newClientErr     error
		addAssigneeErr   error
		err              string
	}{
		{
			desc:             "incorrect event type",
			payloadBytes:     "",
			payloadType:      "",
			getContentOutput: "",
			getContentErr:    nil,
			newClientOutput:  nil,
			newClientErr:     nil,
			addAssigneeErr:   nil,
			err:              "",
		},
		{
			desc:             "error unmarshalling event object",
			payloadBytes:     "[]",
			payloadType:      "issues",
			getContentOutput: "",
			getContentErr:    nil,
			newClientOutput:  nil,
			newClientErr:     nil,
			addAssigneeErr:   nil,
			err:              "error parsing issue: json: cannot unmarshal array into Go value of type github.IssuesEvent",
		},
		{
			desc:             "error unmarshalling event object",
			payloadBytes:     "[]",
			payloadType:      "issues",
			getContentOutput: "",
			getContentErr:    nil,
			newClientOutput:  nil,
			newClientErr:     nil,
			addAssigneeErr:   nil,
			err:              "error parsing issue: json: cannot unmarshal array into Go value of type github.IssuesEvent",
		},
		{
			desc:             "action not closed",
			payloadBytes:     `{"action":"open"}`,
			payloadType:      "issues",
			getContentOutput: "",
			getContentErr:    nil,
			newClientOutput:  nil,
			newClientErr:     nil,
			addAssigneeErr:   nil,
			err:              "",
		},
		{
			desc:             "error getting file content",
			payloadBytes:     `{"action":"closed","issue":{"title":"battle of geonosis","body":"the beginning of the war","repository":{"owner":{"login":"grand-plan"},"name":"dooku"}}}`,
			payloadType:      "issues",
			getContentOutput: "",
			getContentErr:    errors.New("mock get content error"),
			newClientOutput:  nil,
			newClientErr:     nil,
			addAssigneeErr:   nil,
			err:              "error getting heupr config: mock get content error",
		},
		{
			desc:             "error unmarshalling event object",
			payloadBytes:     `{"action":"closed","issue":{"title":"battle of geonosis","body":"the beginning of the war","repository":{"owner":{"login":"grand-plan"},"name":"dooku"}}}`,
			payloadType:      "issues",
			getContentOutput: "-------",
			getContentErr:    nil,
			newClientOutput:  nil,
			newClientErr:     nil,
			addAssigneeErr:   nil,
			err:              "error parsing heupr config: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `-------` into main.configObj",
		},
		{
			desc:             "error creating new es client",
			payloadBytes:     `{"action":"closed","issue":{"title":"battle of geonosis","body":"the beginning of the war","repository":{"owner":{"login":"grand-plan"},"name":"dooku"}}}`,
			payloadType:      "issues",
			getContentOutput: "",
			getContentErr:    nil,
			newClientOutput:  nil,
			newClientErr:     errors.New("mock creating es client error"),
			addAssigneeErr:   nil,
			err:              "error creating es client: mock creating es client error",
		},
		{
			desc:             "error creating new es client",
			payloadBytes:     `{"action":"closed","issue":{"title":"battle of geonosis","body":"the beginning of the war","repository":{"owner":{"login":"grand-plan"},"name":"dooku"}}}`,
			payloadType:      "issues",
			getContentOutput: "",
			getContentErr:    nil,
			newClientOutput:  nil,
			newClientErr:     errors.New("mock creating es client error"),
			addAssigneeErr:   nil,
			err:              "error creating es client: mock creating es client error",
		},
		{
			desc:             "error searching value",
			payloadBytes:     `{"action":"closed","issue":{"title":"battle of geonosis","body":"the beginning of the war","repository":{"owner":{"login":"grand-plan"},"name":"dooku"}}}`,
			payloadType:      "issues",
			getContentOutput: "",
			getContentErr:    nil,
			newClientOutput: &mockES{
				esIndexOutput: "",
				esIndexErr:    errors.New("mock search error"),
			},
			newClientErr:   nil,
			addAssigneeErr: nil,
			err:            "error searching index: mock search error",
		},
		{
			desc:             "add assignee error",
			payloadBytes:     `{"action":"closed","issue":{"title":"battle of geonosis","body":"the beginning of the war","repository":{"owner":{"login":"grand-plan"},"name":"dooku"}}}`,
			payloadType:      "issues",
			getContentOutput: `{backends: [{name: assignissue, events: [{name: issues, actions: [closed]}], settings: {es_endpoint: 'https://example-es-endpoint.com', contributors: [example_github_username]}, location: 'https://github.com/heupr/heupr/assignissue.so'}]}`,
			getContentErr:    nil,
			newClientOutput: &mockES{
				esIndexOutput: "grandmaster yoda",
				esIndexErr:    nil,
			},
			newClientErr:   nil,
			addAssigneeErr: errors.New("mock add assignee error"),
			err:            "error searching index: mock search error",
		},
		{
			desc:             "successful invocation",
			payloadBytes:     `{"action":"closed","issue":{"number": 2,"title":"battle of geonosis","body":"the beginning of the war","repository":{"owner":{"login":"grand-plan"},"name":"dooku"}}}`,
			payloadType:      "issues",
			getContentOutput: `{backends: [{name: assignissue, events: [{name: issues, actions: [closed]}], settings: {es_endpoint: 'https://example-es-endpoint.com', contributors: [yoda]}, location: 'https://github.com/heupr/heupr/assignissue.so'}]}`,
			getContentErr:    nil,
			newClientOutput: &mockES{
				esIndexOutput: "yoda",
				esIndexErr:    nil,
			},
			newClientErr:   nil,
			addAssigneeErr: nil,
			err:            "",
		},
	}

	for _, test := range tests {
		p := &mockPayload{
			payloadBytes: test.payloadBytes,
			payloadType:  test.payloadType,
		}

		newClient = func(url string) (es, error) {
			return test.newClientOutput, test.newClientErr
		}

		h := &mockHelp{
			getContentOutput: test.getContentOutput,
			getContentErr:    test.getContentErr,
			addAssigneeErr:   test.addAssigneeErr,
		}

		b := Backend{
			help: h,
		}

		err := b.Act(p)
		if err != nil && err.Error() != test.err {
			t.Errorf("description: %s, error received: %s, expected: %s", test.desc, err.Error(), test.err)
		}
	}
}

func Test_main(t *testing.T) {
	main() // invoking for test coverage
}
