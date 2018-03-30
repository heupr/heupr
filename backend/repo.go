package backend

import (
	"sync"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"

	"heupr/backend/response"
)

type repo struct {
	sync.Mutex
	settings  *settings
	client    *github.Client
	responses map[string]map[string]*response.Action
}

func (r *repo) parseSettings(s *settings, id int64) error {
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

func newRepo(set *settings, i *integration) (*repo, error) {
	r := new(repo)

	if err := r.parseSettings(set, i.repoID); err != nil {
		return nil, errors.Errorf("parse settings error: %v", err)
	}

	// TODO:
	// [ ] create GitHub client and place into repo field
	// [ ] call training methods for necessary responses
	// [ ] call through spun up goroutines use sync.WaitGroup to coordinate

	return r, nil
}
