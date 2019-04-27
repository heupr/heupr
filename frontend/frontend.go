package frontend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
	"github.com/tidwall/gjson"
)

func stringPtr(input string) *string {
	return &input
}

type object struct {
	ID        *int64  `json:"feature_id,id,omitempty"`
	Type      *string `json:"feature_type,omitempty"`
	Action    *string `json:"feature_action,omitempty"`
	Number    *int    `json:"feature_number,number,omitempty"`
	Title     *string `json:"feature_title,title,omitempty"`
	Body      *string `json:"feature_body,body,omitempty"`
	ActorID   *int64  `json:"feature_actor_name,omitempty"`
	ActorName *string `json:"feature_actor_id,omitempty"`
}

func events(secret, target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := github.ValidatePayload(r, []byte(secret))
		if err != nil {
			http.Error(w, fmt.Sprintf("error validating github payload: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		event, err := github.ParseWebHook(github.WebHookType(r), payload)
		if err != nil {
			http.Error(w, fmt.Sprintf("error parsing github webhook: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		output := object{}
		issueStr := ""

		switch evt := event.(type) {
		case *github.IssuesEvent:
			output.Action = evt.Action

			issue := evt.Issue
			buf := new(bytes.Buffer)

			json.NewEncoder(buf).Encode(issue)
			issueStr = string(buf.Bytes())
			if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
				http.Error(w, fmt.Sprintf("error unmarshalling event object: %s", err.Error()), http.StatusInternalServerError)
				return
			}

		case *github.PullRequestEvent:
		default:
			http.Error(w, "event type not supported", http.StatusInternalServerError)
			return
		}

		actorID := gjson.Get(issueStr, "assignee.id")
		actorName := gjson.Get(issueStr, "assignee.login")

		if !actorID.Exists() || !actorName.Exists() {
			http.Error(w, "assignee id/name not found", http.StatusInternalServerError)
			return
		}

		outputBytes, err := json.Marshal(output)
		if err != nil {
			http.Error(w, fmt.Sprintf("error marshaling github output object: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		req, err := http.NewRequest("POST", target, bytes.NewBuffer(outputBytes))
		if err != nil {
			http.Error(w, fmt.Sprintf("error creating post request: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, fmt.Sprintf("error sending post request %s", err.Error()), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
	}
}

// Server hosts the GitHub webhook listener
type Server struct {
	secret *string
	server http.Server
	target *string
}

// New instantiates a frontend listening server struct without starting it
func New(secret, addr, target string) *Server {
	r := mux.NewRouter()
	r.HandleFunc("/events", events(secret, target)).Methods("POST").Schemes("https")

	s := &Server{
		secret: &secret,
		server: http.Server{
			Addr:         addr,
			Handler:      r,
			WriteTimeout: 10 * time.Second,
			ReadTimeout:  10 * time.Second,
		},
		target: &target,
	}

	return s
}
