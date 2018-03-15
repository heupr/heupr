package backend

import (
	// "errors"
	"sync"

	"github.com/google/go-github/github"

	"heupr/backend/response"
)

type repo struct {
	sync.Mutex
	settings  *settings
	client    *github.Client
	responses map[string][]*response.Action
}

func (s *Server) newRepo(settings settings) {
	// called by:
	// - a loop over a map[int64]settings on server Start
	// - in timer when a new repo is installed and added to the ingestor database
	// accept settings as function argument
	// place settings into respective repo field (locking map?)
	// call parseSettings
	// < more logic here? >
	// places new repo struct into the repos field on the server struct
}

func (r *repo) parseSettings(s *settings) error {
	old := new(settings)
	if r.settings != nil {
		old = r.settings
	}

	if s.Issues != nil {
		for action, responses := range s.Issues {
			if oldResponses, ok := old.Issues[action]; ok {
				for name, options := range responses {
					if _, ok := oldResponses[name]; !ok {
						_ = options
						// boot up new responses w/ options
					}
				}
			} else {
				for name, options := range responses {
					_ = name
					_ = options
					// boot up new responses w/ options
				}
			}
		}
	}

	// create sync.WaitGroup w/ count of needed goroutines (possibly)
	// provides an "interpretation" of the user requirements
	// this method is responsible for:
	// 1) identifying models/conditionals to instantiate/train
	// 2) instantiate necessary conflation/normalization logic based on logic requirements
	// called by:
	// - newRepo (when a new repo is being installed)
	// - settings updates (when an event from the database includes changes to the .heupr.toml file)
	r.settings = s
	return nil
}
