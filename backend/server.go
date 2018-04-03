package backend

import (
	"database/sql"
	"sync"
)

var (
	integrationQuery = ``
	settingsQuery    = ``
	eventsQuery      = ``
)

type repos struct {
	sync.RWMutex
	internal map[int64]*repo
}

// Server hosts backend in-memory active repos and access to the database.
type Server struct {
	database dataAccess
	repos    *repos
}

// openDatabase is designed to be overridden in unit testing.
var openDatabase = func() (dataAccess, error) {
	db, err := sql.Open("mysql", "root@/heupr?interpolateParams=true&parseTime=true")
	if err != nil {
		return nil, err
	}
	return &database{sqlDB: db}, nil
}

// Start is exported so that cmd/ has access to launch the backend.
func (s *Server) Start() {
	db, _ := openDatabase()
	s.database = db

	s.repos = new(repos)

	intg, _ := s.database.readIntegrations(integrationQuery)
	sets, _ := s.database.readSettings(settingsQuery)
	evts, _ := s.database.readEvents(eventsQuery)

	var wg sync.WaitGroup
	wg.Add(len(intg))

	for i := range intg {
		go func() {
			repo, _ := newRepo(sets[i], intg[i])
			s.repos.internal[i] = repo
			_ = evts
			// TODO: Call necessary learn methods w/ evts[i] argument.
			wg.Done()
		}()
	}

	wiggin := make(chan bool)
	s.timer(wiggin)
}

func (s *Server) timer(ender chan bool) {
	// start ticker + dispatcher
	// begin perpetual goroutine
	// if ticker.C
	// read new integrations, settings, and events into memory
	// - place into work struct containing all three
	// - loop over resulting []*work
	// - - pass each *work into worklaod chan *work
	// if ender
	// - stop ticker, close ender, and return
}
