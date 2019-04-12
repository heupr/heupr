package frontend

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-github/github"
)

func TestNewServer(t *testing.T) {
	s := New("test-secret", "127.0.0.1", "test-target")

	if s == nil || s.secret == nil || s.server.Addr == "" || s.server.Handler == nil {
		t.Errorf("frontend server created incorrectly: %+v", s)
	}
}

func stringEvent(event interface{}) string {
	output := []byte{}
	err := errors.New("")

	switch event.(type) {
	case *github.IssuesEvent:
		output, err = json.Marshal(event.(*github.IssuesEvent))
	case *github.IssueCommentEvent:
		output, err = json.Marshal(event.(*github.IssueCommentEvent))
	case *github.PullRequestEvent:
		output, err = json.Marshal(event.(*github.PullRequestEvent))
	case *github.PullRequestReviewEvent:
		output, err = json.Marshal(event.(*github.PullRequestReviewEvent))
	case *github.PullRequestReviewCommentEvent:
		output, err = json.Marshal(event.(*github.PullRequestReviewCommentEvent))
	case *github.CheckRunEvent:
		output, err = json.Marshal(event.(*github.CheckRunEvent))
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

func Test_events(t *testing.T) {
	tests := []struct {
		desc    string
		body    string
		secret  string
		headers map[string]string
		status  int
		resp    string
	}{
		{
			"no application/json header type",
			"test-body",
			"test-secret",
			map[string]string{},
			500,
			"error validating github payload: Webhook request has unsupported Content-Type \"\"",
		},
		{
			"incorrect secret key provided",
			"test-body",
			"test-secret",
			map[string]string{
				"Content-Type": "application/json",
			},
			500,
			"error validating github payload: missing signature",
		},
		{
			"incorrect payload value",
			"test-body",
			"",
			map[string]string{
				"Content-Type": "application/json",
			},
			500,
			"error parsing github webhook: unknown X-Github-Event in message",
		},
		{
			"unsupported event type",
			stringEvent(&github.CheckRunEvent{}),
			"test-secret",
			map[string]string{
				"Content-Type":    "application/json",
				"X-Hub-Signature": encode("test-secret", []byte(stringEvent(&github.CheckRunEvent{}))),
				"X-GitHub-Event":  "check_run",
			},
			500,
			"event type not supported",
		},
		{
			"passing payload",
			stringEvent(&github.IssuesEvent{
				Issue: nil,
			}),
			"test-secret",
			map[string]string{
				"Content-Type":    "application/json",
				"X-Hub-Signature": encode("test-secret", []byte(stringEvent(&github.IssuesEvent{}))),
				"X-GitHub-Event":  "issues",
			},
			200,
			"",
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `output`)
	})
	server := httptest.NewServer(mux)
	target, err := url.Parse(server.URL)
	if err != nil {
		log.Fatalf("error parsing test url: %s", err.Error())
	}

	for _, test := range tests {
		req, err := http.NewRequest("POST", "/events", bytes.NewBufferString(test.body))
		if err != nil {
			t.Fatal(err)
		}

		if len(test.headers) > 0 {
			for header, value := range test.headers {
				req.Header.Set(header, value)
			}
		}

		rec := httptest.NewRecorder()

		handler := http.HandlerFunc(events(test.secret, target.String()))

		handler.ServeHTTP(rec, req)

		if status := rec.Code; status != test.status {
			t.Errorf("description: %s, handler returned incorrect status code, received: %v, expected: %v", test.desc,
				status, test.status)
		}

		if resp := rec.Body.String(); !strings.Contains(resp, test.resp) {
			t.Errorf("description: %s, handler returned incorrect response, received: %v, expected: %v",
				test.desc, resp, test.resp)
		}
	}
}
