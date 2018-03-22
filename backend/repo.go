package backend

import (
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

func (s *Server) newRepo(settings settings) {}

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

	r.settings = s
	return nil
}
