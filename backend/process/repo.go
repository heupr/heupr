package process

import (
	"net/http"
	"sync"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"

	"heupr/backend/process/preprocess"
	"heupr/backend/response"
)

// Repo represents an active GitHub repository.
type Repo struct {
	sync.Mutex
	ID        int64
	settings  *preprocess.Settings
	client    *github.Client
	responses map[string]map[string]*response.Action
}

// Repos provides a concurrency-safety wrap around the Repo map.
type Repos struct {
	sync.RWMutex
	Internal map[int64]*Repo
}

func (r *Repo) parseResponses(s *preprocess.Settings, id int64) error {
	if id == 0 {
		return errors.New("repo id not found")
	}

	responses := make(map[string]map[string]*response.Action)

	for action, opts := range s.Issues {
		for name, opt := range opts {
			if resp, ok := r.responses["issues-"+action][name]; ok {
				resp.Options = opt
				responses["issues-"+action][name] = resp
			} else {
				switch name {
				case "assignment":
					// initialize assignment w/ response.Options struct
					// r.responses["issues-"+action][name] = Action + options
				case "label":
					// initialize label w/ response.Options struct
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

// NewRepo is a helper function to create a new Repo instance.
var NewRepo = func(set *preprocess.Settings, i *preprocess.Integration) (*Repo, error) {
	r := new(Repo)

	if err := r.parseResponses(set, i.RepoID); err != nil {
		return nil, errors.Errorf("parse settings error: %v", err)
	}

	// TODO:
	// [ ] create GitHub client and place into repo field
	// [ ] call training methods for necessary responses
	// [ ] call through spun up goroutines use sync.WaitGroup to coordinate

	return r, nil
}
