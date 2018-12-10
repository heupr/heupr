package backend

import (
	"net/http"
	"sync"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
)

// repo represents an active GitHub repository.
type repo struct {
	sync.Mutex
	ID        int64
	settings  *settings
	client    *github.Client
	responses map[string]map[string]model
}

// repos provides a concurrency-safety wrap around the repo map.
type repos struct {
	sync.RWMutex
	Internal map[int64]*repo
}

func (r *repo) parseResponses(s *settings, id int64) error {
	if id == 0 {
		return errors.New("repo id not found")
	}

	responses := make(map[string]map[string]model)

	for action, opts := range s.Issue {
		for name := range opts {
			if resp, ok := r.responses["issues-"+action][name]; ok {
				responses["issues-"+action][name] = resp
			} else {
				switch name {
				case "assignment":
					// initialize assignment w/ response.issueOptions struct
					// r.responses["issues-"+action][name] = Action + options
				case "label":
					// initialize label w/ response.issueOptions struct
				default:
					return errors.Errorf("response %v not recognized", name)
				}
			}
		}
	}

	r.settings = s
	r.responses = responses

	return nil
}

// NewClient returns a new GitHub client.
var NewClient = func(appID, installationID int) *github.Client {
	key := ""
	if PROD {
		key = "heupr.2017-10-04.private-key.pem"
	} else {
		key = "mikeheuprtest.2017-11-16.private-key.pem"
	}

	tr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, appID, installationID, key)
	if err != nil {
		_ = err
	}

	client := github.NewClient(&http.Client{Transport: tr})
	return client
}

var newRepo = func(set *settings, i *integration) (*repo, error) {
	r := new(repo)

	if err := r.parseResponses(set, i.RepoID); err != nil {
		return nil, errors.Errorf("parse settings error: %v", err)
	}

	// TODO:
	// [ ] create GitHub client and place into repo field
	// [ ] call training methods for necessary responses
	// [ ] call through spun up goroutines use sync.WaitGroup to coordinate

	return r, nil
}
