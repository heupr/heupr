package frontend

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/go-github/v28/github"

	"github.com/heupr/heupr/backend"
)

type databaseMock struct {
	putErr  error
	getResp installConfig
	getErr  error
}

func (mock *databaseMock) Put(input installConfig) error {
	return mock.putErr
}

func (mock *databaseMock) Get(key interface{}) (installConfig, error) {
	return mock.getResp, mock.getErr
}

func TestInstall(t *testing.T) {
	tests := []struct {
		desc     string
		code     string
		postResp *http.Response
		postErr  error
		putErr   error
		err      string
		status   int
		respBody string
	}{
		{
			desc:     "no temporary manifest code provided",
			code:     "",
			postResp: nil,
			postErr:  nil,
			putErr:   nil,
			err:      "no code received",
			status:   400,
			respBody: "no code received",
		},
		{
			desc:     "error requesting temporary manifest code",
			code:     "test-code",
			postResp: nil,
			postErr:  errors.New("mock post error"),
			putErr:   nil,
			err:      "error converting code: mock post error",
			status:   500,
			respBody: "error converting code: mock post error",
		},
		{
			desc: "error saving config values",
			code: "test-code",
			postResp: &http.Response{
				Body: ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
			},
			postErr:  nil,
			putErr:   errors.New("mock put error"),
			err:      "error putting app config: mock put error",
			status:   500,
			respBody: "error putting app config: mock put error",
		},
		{
			desc: "successful invocation",
			code: "test-code",
			postResp: &http.Response{
				Body: ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
			},
			postErr:  nil,
			putErr:   nil,
			err:      "",
			status:   302,
			respBody: "success",
		},
	}

	for _, test := range tests {
		post = func(req *http.Request) (*http.Response, error) {
			return test.postResp, test.postErr
		}

		req := events.APIGatewayProxyRequest{
			QueryStringParameters: map[string]string{
				"code": test.code,
			},
		}

		db := &databaseMock{
			putErr: test.putErr,
		}

		resp, err := Install(req, db)
		if err != nil && err.Error() != test.err {
			t.Errorf("description: %s, incorrect error message, received: %s, expected: %s", test.desc, err.Error(), test.err)
		}

		if resp.StatusCode != test.status {
			t.Errorf("description: %s, incorrect status code, received: %d, expected: %d", test.desc, resp.StatusCode, test.status)
		}

		if resp.Body != test.respBody {
			t.Errorf("description: %s, incorrect body, received: %s, expected: %s", test.desc, resp.Body, test.respBody)
		}
	}
}

func Test_payloadMethods(t *testing.T) {
	p := &payload{
		B: []byte("content"),
		T: "type",
	}

	bytesOut := p.Bytes()
	if bytesOut == nil {
		t.Error("description: no byte array returned")
	}

	typeOutput := p.Type()
	if typeOutput == "" {
		t.Error("description: no type string returned")
	}
}

func int64Ptr(input int64) *int64 {
	return &input
}

func stringPtr(input string) *string {
	return &input
}

// type testPayload struct{}
//
// func (tp testPayload) Type() string {
// 	return ""
// }
//
// func (tp testPayload) Bytes() []byte {
// 	return []byte{}
// }

type testBackend struct {
	prepareErr error
	actErr     error
}

func (tb *testBackend) Configure(*github.Client) {}

func (tb *testBackend) Prepare(backend.Payload) error {
	return tb.prepareErr
}

func (tb *testBackend) Act(backend.Payload) error {
	return tb.actErr
}

func TestEvent(t *testing.T) {
	tests := []struct {
		desc           string
		body           string
		headers        map[string]string
		bknds          []backend.Backend
		getResp        installConfig
		getErr         error
		putErr         error
		validateErr    error
		clientErr      error
		getContentResp string
		getContentErr  error
		err            string
		status         int
		respBody       string
	}{
		{
			desc: "incorrect event type",
			body: "",
			headers: map[string]string{
				"X-GitHub-Event":  "test-event",
				"X-Hub-Signature": "test-signature",
			},
			bknds:          []backend.Backend{},
			getResp:        installConfig{},
			getErr:         nil,
			putErr:         nil,
			validateErr:    nil,
			clientErr:      nil,
			getContentResp: "",
			getContentErr:  nil,
			err:            "event type test-event not supported",
			status:         500,
			respBody:       "event type test-event not supported",
		},
		{
			desc: "error getting config",
			body: `{"installation": {"app_id": 1038}}`,
			headers: map[string]string{
				"X-GitHub-Event":  "installation_repositories",
				"X-Hub-Signature": "test-signature",
			},
			bknds:          []backend.Backend{},
			getResp:        installConfig{},
			getErr:         errors.New("mock get error"),
			putErr:         nil,
			validateErr:    nil,
			clientErr:      nil,
			getContentResp: "",
			getContentErr:  nil,
			err:            "error getting config: mock get error",
			status:         500,
			respBody:       "error getting config: mock get error",
		},
		{
			desc: "error validating received event",
			body: `{"installation": {"app_id": 1038}}`,
			headers: map[string]string{
				"X-GitHub-Event":  "installation_repositories",
				"X-Hub-Signature": "test-signature",
			},
			bknds:          []backend.Backend{},
			getResp:        installConfig{},
			getErr:         nil,
			putErr:         nil,
			validateErr:    errors.New("mock validate error"),
			clientErr:      nil,
			getContentResp: "",
			getContentErr:  nil,
			err:            "error validating event: mock validate error",
			status:         500,
			respBody:       "error validating event: mock validate error",
		},
		{
			desc: "error creating client",
			body: `{"installation": {"app_id": 1, "id": 2}, "repositories_added": [{"owner": {"login": "test-login"}, "name": "test-name", "full_name": "test-fullname"}]}`,
			headers: map[string]string{
				"X-GitHub-Event":  "installation_repositories",
				"X-Hub-Signature": "test-signature",
			},
			bknds:          []backend.Backend{},
			getResp:        installConfig{},
			getErr:         nil,
			putErr:         nil,
			validateErr:    nil,
			clientErr:      errors.New("mock client error"),
			getContentResp: "",
			getContentErr:  nil,
			err:            "error creating client: mock client error",
			status:         500,
			respBody:       "error creating client: mock client error",
		},
		{
			desc: "error putting config data",
			body: `{"installation": {"app_id": 1, "id": 2}, "repositories_added": [{"owner": {"login": "test-login"}, "name": "test-name", "full_name": "test-fullname"}]}`,
			headers: map[string]string{
				"X-GitHub-Event":  "installation_repositories",
				"X-Hub-Signature": "test-signature",
			},
			bknds:          []backend.Backend{},
			getResp:        installConfig{},
			getErr:         nil,
			putErr:         errors.New("mock put error"),
			validateErr:    nil,
			clientErr:      nil,
			getContentResp: "",
			getContentErr:  nil,
			err:            "error putting app config: mock put error",
			status:         500,
			respBody:       "error putting app config: mock put error",
		},
		{
			desc: "error getting repo config content",
			body: `{"installation": {"app_id": 1, "id": 2}, "repositories_added": [{"full_name": "test-owner/test-name"}]}`,
			headers: map[string]string{
				"X-GitHub-Event":  "installation_repositories",
				"X-Hub-Signature": "test-signature",
			},
			bknds:          []backend.Backend{},
			getResp:        installConfig{},
			getErr:         nil,
			putErr:         nil,
			validateErr:    nil,
			clientErr:      nil,
			getContentResp: "",
			getContentErr:  errors.New("mock content error"),
			err:            "error getting repo config file: mock content error",
			status:         500,
			respBody:       "error getting repo config file: mock content error",
		},
		{
			desc: "error calling install event backend prepare",
			body: `{"installation": {"app_id": 1, "id": 2}, "repositories_added": [{"full_name": "test-owner/test-name"}]}`,
			headers: map[string]string{
				"X-GitHub-Event":  "installation_repositories",
				"X-Hub-Signature": "test-signature",
			},
			bknds: []backend.Backend{
				&testBackend{
					prepareErr: errors.New("mock prepare error"),
					actErr:     nil,
				},
			},
			getResp:        installConfig{},
			getErr:         nil,
			putErr:         nil,
			validateErr:    nil,
			clientErr:      nil,
			getContentResp: "",
			getContentErr:  nil,
			err:            "error calling backend prepare: mock prepare error",
			status:         500,
			respBody:       "error calling backend prepare: mock prepare error",
		},
		{
			desc: "successful install event invocation",
			body: `{"installation": {"app_id": 1, "id": 2}, "repositories_added": [{"full_name": "test-owner/test-name"}]}`,
			headers: map[string]string{
				"X-GitHub-Event":  "installation_repositories",
				"X-Hub-Signature": "test-signature",
			},
			bknds: []backend.Backend{
				&testBackend{
					prepareErr: nil,
					actErr:     nil,
				},
			},
			getResp:        installConfig{},
			getErr:         nil,
			putErr:         nil,
			validateErr:    nil,
			clientErr:      nil,
			getContentResp: "",
			getContentErr:  nil,
			err:            "",
			status:         200,
			respBody:       "success",
		},
		{
			desc: "error getting config from database",
			body: `{"repository": {"full_name": "test-owner/test-name"}}`,
			headers: map[string]string{
				"X-GitHub-Event":  "issues",
				"X-Hub-Signature": "test-signature",
			},
			bknds:          []backend.Backend{},
			getResp:        installConfig{},
			getErr:         errors.New("mock get error"),
			putErr:         nil,
			validateErr:    nil,
			clientErr:      nil,
			getContentResp: "",
			getContentErr:  nil,
			err:            "error getting config: mock get error",
			status:         500,
			respBody:       "error getting config: mock get error",
		},
		{
			desc: "error validating received event",
			body: `{"repository": {"full_name": "test-owner/test-name"}}`,
			headers: map[string]string{
				"X-GitHub-Event":  "issues",
				"X-Hub-Signature": "test-signature",
			},
			bknds:          []backend.Backend{},
			getResp:        installConfig{},
			getErr:         nil,
			putErr:         nil,
			validateErr:    errors.New("mock validate error"),
			clientErr:      nil,
			getContentResp: "",
			getContentErr:  nil,
			err:            "error validating event: mock validate error",
			status:         500,
			respBody:       "error validating event: mock validate error",
		},
		{
			desc: "error creating client",
			body: `{"repository": {"full_name": "test-owner/test-name"}}`,
			headers: map[string]string{
				"X-GitHub-Event":  "issues",
				"X-Hub-Signature": "test-signature",
			},
			bknds:          []backend.Backend{},
			getResp:        installConfig{},
			getErr:         nil,
			putErr:         nil,
			validateErr:    nil,
			clientErr:      errors.New("mock client error"),
			getContentResp: "",
			getContentErr:  nil,
			err:            "error creating client: mock client error",
			status:         500,
			respBody:       "error creating client: mock client error",
		},
		{
			desc: "error getting repo config content",
			body: `{"repository": {"full_name": "test-owner/test-name"}}`,
			headers: map[string]string{
				"X-GitHub-Event":  "issues",
				"X-Hub-Signature": "test-signature",
			},
			bknds:          []backend.Backend{},
			getResp:        installConfig{},
			getErr:         nil,
			putErr:         nil,
			validateErr:    nil,
			clientErr:      nil,
			getContentResp: "",
			getContentErr:  errors.New("mock content error"),
			err:            "error getting repo config file: mock content error",
			status:         500,
			respBody:       "error getting repo config file: mock content error",
		},
		{
			desc: "error calling issue event backend act",
			body: `{"repository": {"full_name": "test-owner/test-name"}}`,
			headers: map[string]string{
				"X-GitHub-Event":  "issues",
				"X-Hub-Signature": "test-issues",
			},
			bknds: []backend.Backend{
				&testBackend{
					prepareErr: nil,
					actErr:     errors.New("mock act error"),
				},
			},
			getResp:        installConfig{},
			getErr:         nil,
			putErr:         nil,
			validateErr:    nil,
			clientErr:      nil,
			getContentResp: "",
			getContentErr:  nil,
			err:            "error calling backend act: mock act error",
			status:         500,
			respBody:       "error calling backend act: mock act error",
		},
		{
			desc: "successful issue event invocation",
			body: `{"repository": {"full_name": "test-owner/test-name"}}`,
			headers: map[string]string{
				"X-GitHub-Event":  "issues",
				"X-Hub-Signature": "test-signature",
			},
			bknds: []backend.Backend{
				&testBackend{
					prepareErr: nil,
					actErr:     nil,
				},
			},
			getResp:        installConfig{},
			getErr:         nil,
			putErr:         nil,
			validateErr:    nil,
			clientErr:      nil,
			getContentResp: "",
			getContentErr:  nil,
			err:            "",
			status:         200,
			respBody:       "success",
		},
	}

	for _, test := range tests {
		validateEvent = func(secret, signature string, body []byte) error {
			return test.validateErr
		}

		newClient = func(appID, installationID int64, file string) (*github.Client, error) {
			return github.NewClient(nil), test.clientErr
		}

		getContent = func(c *github.Client, owner, repo, path string) (string, error) {
			return test.getContentResp, test.getContentErr
		}

		req := events.APIGatewayProxyRequest{
			Body:    test.body,
			Headers: test.headers,
		}

		db := &databaseMock{
			getResp: test.getResp,
			getErr:  test.getErr,
			putErr:  test.putErr,
		}

		resp, err := Event(req, db, test.bknds)
		if err != nil && err.Error() != test.err {
			t.Errorf("description: %s, incorrect error message, received: %s, expected: %s", test.desc, err.Error(), test.err)
		}

		if resp.StatusCode != test.status {
			t.Errorf("description: %s, incorrect status code, received: %d, expected: %d", test.desc, resp.StatusCode, test.status)
		}

		if resp.Body != test.respBody {
			t.Errorf("description: %s, incorrect body, received: %s, expected: %s", test.desc, resp.Body, test.respBody)
		}
	}
}
