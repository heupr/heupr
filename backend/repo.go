package backend

import (
	"sync"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"

	"heupr/backend/response"
)

type repo struct {
	sync.Mutex
	id        int64
	settings  *settings
	client    *github.Client
	responses map[string]map[string]*response.Action
}

func (r *repo) parseResponses(settings *settings) error {
	if r.id <= 0 {
		return errors.Errorf("invalid repo id %d", r.id)
	}

	responses := make(map[string]map[string]*response.Action)

	for action, opts := range settings.Issues {
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

	r.settings = settings
	r.responses = responses

	return nil
}

var newRepo = func(settings *settings, i *integration) (*repo, error) {
	repo := &repo{id: i.repoID}

	if err := repo.parseResponses(settings); err != nil {
		return nil, errors.Errorf("parse settings error: %v", err)
	}

	// TODO:
	// [ ] create GitHub client and place into repo field
	// [ ] call training methods for necessary responses
	// [ ] call through spun up goroutines use sync.WaitGroup to coordinate

	return repo, nil
}
