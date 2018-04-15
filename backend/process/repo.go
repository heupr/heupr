package process

import (
	"sync"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"

	"heupr/backend/response"
)

type Repo struct {
	sync.Mutex
	setting   *Setting
	client    *github.Client
	responses map[string]map[string]*response.Action
}

type Repos struct {
	sync.RWMutex
	Internal map[int64]*Repo
}

func (r *Repo) parseSettings(s *Setting, id int64) error {
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

	r.setting = s
	r.responses = responses

	return nil
}

var NewRepo = func(set *Setting, i *Integration) (*Repo, error) {
	r := new(Repo)

	if err := r.parseSettings(set, i.RepoID); err != nil {
		return nil, errors.Errorf("parse settings error: %v", err)
	}

	// TODO:
	// [ ] create GitHub client and place into repo field
	// [ ] call training methods for necessary responses
	// [ ] call through spun up goroutines use sync.WaitGroup to coordinate

	return r, nil
}
