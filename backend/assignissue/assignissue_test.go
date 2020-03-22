package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/google/go-github/v28/github"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

type mockPayload struct {
	payloadBytes  string
	payloadType   string
	payloadConfig string
}

func (m *mockPayload) Type() string {
	return m.payloadType
}

func (m *mockPayload) Bytes() []byte {
	return []byte(m.payloadBytes)
}

func (m *mockPayload) Config() []byte {
	return []byte(m.payloadConfig)
}

type mockHelp struct {
	listIssuesOutput []*github.Issue
	listIssuesErr    error
	getContentOutput string
	getContentErr    error
	getTextOutput    string
	addAssigneeErr   error
}

type mockBleve struct {
	indexErr     error
	searchOutput string
	searchErr    error
}

func (m *mockBleve) index(repo, key, value string) error {
	return m.indexErr
}

func (m *mockBleve) search(repo, blob string) (string, error) {
	return m.searchOutput, m.searchErr
}

func (m *mockHelp) putIndex(path string) error {
	return nil
}

func (m *mockHelp) getIndex(path string) error {
	return nil
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
		newClientOutput  assigner
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
			err:              "error unmarshalling installation: json: cannot unmarshal array into Go value of type github.InstallationRepositoriesEvent",
		},
		{
			desc:             "error getting config file content",
			payloadBytes:     `{"repositories_added":[{"full_name":"delta-squad/CC-1038"}]}`,
			getContentOutput: "",
			getContentErr:    errors.New("mock get content error"),
			listIssuesOutput: nil,
			listIssuesErr:    nil,
			newClientOutput:  nil,
			err:              "error getting heupr config: mock get content error",
		},
		{
			desc:             "error unmarshalling config file",
			payloadBytes:     `{"repositories_added":[{"full_name":"delta-squad/CC-1038"}]}`,
			getContentOutput: "-------",
			getContentErr:    nil,
			listIssuesOutput: nil,
			listIssuesErr:    nil,
			newClientOutput:  nil,
			err:              "error parsing heupr config: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `-------` into main.configObj",
		},
		{
			desc:             "error listing issues",
			payloadBytes:     `{"repositories_added":[{"full_name":"delta-squad/CC-1038"}]}`,
			getContentOutput: "",
			getContentErr:    nil,
			listIssuesOutput: nil,
			listIssuesErr:    errors.New("mock list issues error"),
			newClientOutput:  nil,
			err:              "error getting issues: mock list issues error",
		},
		{
			desc:             "error indexing value",
			payloadBytes:     `{"repositories_added":[{"full_name":"delta-squad/CC-1038"}]}`,
			getContentOutput: "",
			getContentErr:    nil,
			listIssuesOutput: []*github.Issue{
				{
					State: stringPtr("closed"),
					Assignee: &github.User{
						Login: stringPtr("CC-01/425"),
					},
					Title: stringPtr("field-advisor"),
					Body:  stringPtr("hologram only"),
				},
			},
			listIssuesErr: nil,
			newClientOutput: &mockBleve{
				indexErr: errors.New("mock index error"),
			},
			err: "error indexing key/value: mock index error",
		},
		{
			desc:             "successful invocation",
			payloadBytes:     `{"repositories_added":[{"full_name":"delta-squad/CC-1038"}]}`,
			getContentOutput: "",
			getContentErr:    nil,
			listIssuesOutput: []*github.Issue{
				{
					State: stringPtr("closed"),
					Assignee: &github.User{
						Login: stringPtr("CC-01/425"),
					},
					Title: stringPtr("field-advisor"),
					Body:  stringPtr("hologram only"),
				},
			},
			listIssuesErr: nil,
			newClientOutput: &mockBleve{
				indexErr: nil,
			},
			err: "",
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

		newClient = func() assigner {
			return test.newClientOutput
		}

		b := Backend

		b.help = h
		b.github = github.NewClient(nil)

		err := b.Prepare(p)
		if err != nil && err.Error() != test.err {
			t.Errorf("description: %s, error received: %s, expected: %s", test.desc, err.Error(), test.err)
		}
	}
}

func TestAct(t *testing.T) {
	tests := []struct {
		desc            string
		payloadBytes    string
		payloadType     string
		getTextOutput   string
		payloadConfig   string
		newClientOutput assigner
		addAssigneeErr  error
		err             string
	}{
		{
			desc:            "incorrect event type",
			payloadBytes:    "",
			payloadType:     "",
			getTextOutput:   "",
			payloadConfig:   "",
			newClientOutput: nil,
			addAssigneeErr:  nil,
			err:             "",
		},
		{
			desc:            "error unmarshalling event object",
			payloadBytes:    "[]",
			payloadType:     "issues",
			getTextOutput:   "",
			payloadConfig:   "",
			newClientOutput: nil,
			addAssigneeErr:  nil,
			err:             "error parsing issue: json: cannot unmarshal array into Go value of type github.IssuesEvent",
		},
		{
			desc:            "action not opened",
			payloadBytes:    `{"action":"closed"}`,
			payloadType:     "issues",
			getTextOutput:   "",
			payloadConfig:   "",
			newClientOutput: nil,
			addAssigneeErr:  nil,
			err:             "",
		},
		{
			desc:            "error unmarshalling event object",
			payloadBytes:    `{"action":"opened","issue":{"title":"battle of geonosis","body":"the beginning of the war"},"repository":{"full_name": "grand-plan/dooku"}}`,
			payloadType:     "issues",
			getTextOutput:   "issue corpus",
			payloadConfig:   "-------",
			newClientOutput: nil,
			addAssigneeErr:  nil,
			err:             "error parsing heupr config: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `-------` into main.configObj",
		},
		{
			desc:          "error searching value",
			payloadBytes:  `{"action":"opened","issue":{"title":"battle of geonosis","body":"the beginning of the war"},"repository":{"full_name": "grand-plan/dooku"}}`,
			payloadType:   "issues",
			getTextOutput: "issue corpus",
			payloadConfig: "",
			newClientOutput: &mockBleve{
				searchOutput: "",
				searchErr:    errors.New("mock search error"),
			},
			addAssigneeErr: nil,
			err:            "error searching index: mock search error",
		},
		{
			desc:          "add assignee error",
			payloadBytes:  `{"action":"opened","issue":{"title":"battle of geonosis","body":"the beginning of the war"},"repository":{"full_name": "grand-plan/dooku"}}`,
			payloadType:   "issues",
			getTextOutput: "issue corpus",
			payloadConfig: `{backends: [{name: assignissue, events: [{name: issues, actions: [opened]}], settings: {contributors: [example_github_username]}}]}`,
			newClientOutput: &mockBleve{
				searchOutput: "grandmaster yoda",
				searchErr:    nil,
			},
			addAssigneeErr: errors.New("mock add assignee error"),
			err:            "error searching index: mock search error",
		},
		{
			desc:          "successful invocation",
			payloadBytes:  `{"action":"opened","issue":{"number": 2,"title":"battle of geonosis","body":"the beginning of the war"},"repository":{"full_name":"grand-plan/dooku"}}`,
			payloadType:   "issues",
			getTextOutput: "issue corpus",
			payloadConfig: `{backends: [{name: assignissue, events: [{name: issues, actions: [opened]}], settings: {contributors: [example_github_username]}}]}`,
			newClientOutput: &mockBleve{
				searchOutput: "yoda",
				searchErr:    nil,
			},
			addAssigneeErr: nil,
			err:            "",
		},
	}

	for _, test := range tests {
		p := &mockPayload{
			payloadBytes:  test.payloadBytes,
			payloadType:   test.payloadType,
			payloadConfig: test.payloadConfig,
		}

		newClient = func() assigner {
			return test.newClientOutput
		}

		h := &mockHelp{
			getTextOutput: test.getTextOutput,
			// getContentOutput: test.getContentOutput,
			// getContentErr:    test.getContentErr,
			addAssigneeErr: test.addAssigneeErr,
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
