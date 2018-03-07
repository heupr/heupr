package backend

import (
	"sync"

	"github.com/google/go-github/github"

	"heupr/logic"
)

type repo struct {
	sync.Mutex
	settings     *settings
	client       *github.Client
	models       map[string][]*logic.Model
	conditionals map[string][]*logic.Conditional
}

func (s *Server) newRepo() {
	// create sync.WaitGroup w/ count of needed goroutines (possibly)
	// read settings from database (and additional integration/installation info)
	// instantiate conditionals/train models w/ settings (separate methods)
}
