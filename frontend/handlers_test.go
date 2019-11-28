package frontend

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/go-github/v28/github"
)

type databaseMock struct {
	putErr  error
	getResp installConfig
	getErr  error
}

func (mock *databaseMock) Put(input installConfig) error {
	return mock.putErr
}

func (mock *databaseMock) Get(key string) (installConfig, error) {
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
			status:   200,
			respBody: "success",
		},
	}

	for _, test := range tests {
		post = func(url, contentType string, body io.Reader) (resp *http.Response, err error) {
			return test.postResp, test.postErr
		}

		req := events.APIGatewayProxyRequest{
			Headers: map[string]string{
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

func stringEvent(event interface{}) string {
	output := []byte{}
	err := errors.New("")

	switch event.(type) {
	case *github.InstallationRepositoriesEvent:
		output, err = json.Marshal(event.(*github.InstallationRepositoriesEvent))
	}

	if err != nil {
		log.Fatalf("error creating event string: %s", err.Error())
	}

	return string(output)
}

func encode(secret string, payload []byte) string {
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write(payload)
	sig := "sha1=" + hex.EncodeToString(mac.Sum(nil))
	return sig
}

func TestEvent(t *testing.T) {
	tests := []struct {
		desc           string
		body           string
		headers        map[string]string
		getResp        installConfig
		getErr         error
		putErr         error
		clientErr      error
		getContentResp string
		getContentErr  error
		err            string
		status         int
		respBody       string
	}{
		{
			desc: "error getting config",
			body: "test-body",
			headers: map[string]string{
				"X-Hub-Signature": "test-secret",
			},
			getResp:        installConfig{},
			getErr:         errors.New("mock get error"),
			putErr:         nil,
			clientErr:      nil,
			getContentResp: "",
			getContentErr:  nil,
			err:            "error getting config: mock get error",
			status:         500,
			respBody:       "error getting config: mock get error",
		},
		{
			desc: "error validating payload",
			body: "test-body",
			headers: map[string]string{
				"X-Hub-Signature": "test-secret",
			},
			getResp: installConfig{
				WebhookSecret: "test-wrong-secret",
			},
			getErr:         nil,
			putErr:         nil,
			clientErr:      nil,
			getContentResp: "",
			getContentErr:  nil,
			err:            "error validating payload: error parsing signature \"test-secret\"",
			status:         500,
			respBody:       "error validating payload: error parsing signature \"test-secret\"",
		},
		{
			desc: "error parsing payload",
			body: stringEvent(&github.InstallationRepositoriesEvent{
				Action: stringPtr("removed"),
				Installation: &github.Installation{
					ID:    int64Ptr(5),
					AppID: int64Ptr(1977),
				},
			}),
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Hub-Signature": encode("test-secret", []byte(stringEvent(&github.InstallationRepositoriesEvent{
					Action: stringPtr("removed"),
					Installation: &github.Installation{
						ID:    int64Ptr(5),
						AppID: int64Ptr(1977),
					},
				}))),
				"X-Github-Event": "installation_repositories",
			},
			getResp: installConfig{
				WebhookSecret: "test-secret",
			},
			getErr:         nil,
			putErr:         nil,
			clientErr:      nil,
			getContentResp: "",
			getContentErr:  nil,
			err:            "repo event not \"added\", received event: \"removed\"",
			status:         500,
			respBody:       "repo event not \"added\", received event: \"removed\"",
		},
		{
			desc: "error creating install client",
			body: stringEvent(&github.InstallationRepositoriesEvent{
				Action: stringPtr("added"),
				Installation: &github.Installation{
					ID:    int64Ptr(5),
					AppID: int64Ptr(1980),
				},
			}),
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Hub-Signature": encode("test-secret", []byte(stringEvent(&github.InstallationRepositoriesEvent{
					Action: stringPtr("added"),
					Installation: &github.Installation{
						ID:    int64Ptr(5),
						AppID: int64Ptr(1980),
					},
				}))),
				"X-Github-Event": "installation_repositories",
			},
			getResp: installConfig{
				WebhookSecret: "test-secret",
			},
			getErr:         nil,
			putErr:         nil,
			clientErr:      errors.New("mock new client error"),
			getContentResp: "",
			getContentErr:  nil,
			err:            "error creating client: mock new client error",
			status:         500,
			respBody:       "error creating client: mock new client error",
		},
		{
			desc: "error getting install content",
			body: stringEvent(&github.InstallationRepositoriesEvent{
				Action: stringPtr("added"),
				Installation: &github.Installation{
					ID:    int64Ptr(5),
					AppID: int64Ptr(1983),
				},
				RepositoriesAdded: []*github.Repository{
					&github.Repository{
						Owner: &github.User{
							Login: stringPtr("star-wars"),
						},
						Name: stringPtr("iv"),
					},
				},
			}),
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Hub-Signature": encode("test-secret", []byte(stringEvent(&github.InstallationRepositoriesEvent{
					Action: stringPtr("added"),
					Installation: &github.Installation{
						ID:    int64Ptr(5),
						AppID: int64Ptr(1983),
					},
					RepositoriesAdded: []*github.Repository{
						&github.Repository{
							Owner: &github.User{
								Login: stringPtr("star-wars"),
							},
							Name: stringPtr("iv"),
						},
					},
				}))),
				"X-Github-Event": "installation_repositories",
			},
			getResp: installConfig{
				WebhookSecret: "test-secret",
			},
			getErr:         nil,
			putErr:         nil,
			clientErr:      nil,
			getContentResp: "",
			getContentErr:  errors.New("mock get content error"),
			err:            "error getting repo config file: mock get content error",
			status:         500,
			respBody:       "error getting repo config file: mock get content error",
		},
		{
			desc: "error parsing config install content",
			body: stringEvent(&github.InstallationRepositoriesEvent{
				Action: stringPtr("added"),
				Installation: &github.Installation{
					ID:    int64Ptr(5),
					AppID: int64Ptr(1983),
				},
				RepositoriesAdded: []*github.Repository{
					&github.Repository{
						Owner: &github.User{
							Login: stringPtr("star-wars"),
						},
						Name: stringPtr("iv"),
					},
				},
			}),
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Hub-Signature": encode("test-secret", []byte(stringEvent(&github.InstallationRepositoriesEvent{
					Action: stringPtr("added"),
					Installation: &github.Installation{
						ID:    int64Ptr(5),
						AppID: int64Ptr(1983),
					},
					RepositoriesAdded: []*github.Repository{
						&github.Repository{
							Owner: &github.User{
								Login: stringPtr("star-wars"),
							},
							Name: stringPtr("iv"),
						},
					},
				}))),
				"X-Github-Event": "installation_repositories",
			},
			getResp: installConfig{
				WebhookSecret: "test-secret",
			},
			getErr:         nil,
			putErr:         nil,
			clientErr:      nil,
			getContentResp: "-------",
			getContentErr:  nil,
			err:            "error parsing repo config file: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `-------` into frontend.configObj",
			status:         500,
			respBody:       "error parsing repo config file: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `-------` into frontend.configObj",
		},
		{
			desc: "error creating event client",
			body: stringEvent(&github.InstallationRepositoriesEvent{
				Action: stringPtr("open"),
				Installation: &github.Installation{
					ID:    int64Ptr(5),
					AppID: int64Ptr(1999),
				},
			}),
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Hub-Signature": encode("test-secret", []byte(stringEvent(&github.InstallationRepositoriesEvent{
					Action: stringPtr("open"),
					Installation: &github.Installation{
						ID:    int64Ptr(5),
						AppID: int64Ptr(1999),
					},
				}))),
				"X-Github-Event": "issues",
			},
			getResp: installConfig{
				WebhookSecret: "test-secret",
			},
			getErr:         nil,
			putErr:         nil,
			clientErr:      errors.New("mock new client error"),
			getContentResp: "",
			getContentErr:  nil,
			err:            "error creating client: mock new client error",
			status:         500,
			respBody:       "error creating client: mock new client error",
		},
		{
			desc: "error getting event content",
			body: stringEvent(&github.InstallationRepositoriesEvent{
				Action: stringPtr("open"),
				Installation: &github.Installation{
					ID:    int64Ptr(5),
					AppID: int64Ptr(1999),
				},
			}),
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Hub-Signature": encode("test-secret", []byte(stringEvent(&github.InstallationRepositoriesEvent{
					Action: stringPtr("open"),
					Installation: &github.Installation{
						ID:    int64Ptr(5),
						AppID: int64Ptr(1999),
					},
				}))),
				"X-Github-Event": "issues",
			},
			getResp: installConfig{
				WebhookSecret: "test-secret",
			},
			getErr:         nil,
			putErr:         nil,
			clientErr:      nil,
			getContentResp: "",
			getContentErr:  errors.New("mock get content error"),
			err:            "error getting repo config file: mock get content error",
			status:         500,
			respBody:       "error getting repo config file: mock get content error",
		},
		{
			desc: "error parsing config event content",
			body: stringEvent(&github.InstallationRepositoriesEvent{
				Action: stringPtr("open"),
			}),
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Hub-Signature": encode("test-secret", []byte(stringEvent(&github.InstallationRepositoriesEvent{
					Action: stringPtr("open"),
				}))),
				"X-Github-Event": "issues",
			},
			getResp: installConfig{
				WebhookSecret: "test-secret",
			},
			getErr:         nil,
			putErr:         nil,
			clientErr:      nil,
			getContentResp: "-------",
			getContentErr:  nil,
			err:            "error parsing repo config file: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `-------` into frontend.configObj",
			status:         500,
			respBody:       "error parsing repo config file: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `-------` into frontend.configObj",
		},
	}

	for _, test := range tests {
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

		resp, err := Event(req, db)
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
