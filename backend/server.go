package backend

import (
	"database/sql"
	"sync"
	"time"
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

// Server hosts backend in-memory active repos and access to the database as
// well as channels for processing incoming work.
type Server struct {
	database dataAccess
	work     chan *work
	workers  chan chan *work
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

// openQueues is another function to override in unit testing.
var openQueues = func() (chan *work, chan chan *work) {
	return make(chan *work), make(chan chan *work)
}

// Start is exported so that cmd/ has access to launch the backend.
func (s *Server) Start() {
	db, _ := openDatabase()
	s.database = db

	workQueue, workerQueue := openQueues()
	s.work = workQueue
	s.workers = workerQueue

	s.repos = &repos{
		internal: make(map[int64]*repo),
	}

	integrations, _ := s.database.readIntegrations(integrationQuery)
	settings, _ := s.database.readSettings(settingsQuery)

	for i := range integrations {
		// TODO: There should likely be a check for settings[i] existence.
		repo, _ := newRepo(settings[i], integrations[i])
		s.repos.Lock()
		s.repos.internal[integrations[i].repoID] = repo
		s.repos.Unlock()
	}

	wiggin := make(chan bool)
	s.tick(wiggin, s.work, s.workers)
}

var (
	newIntegrationsQuery = ``
	newSettingsQuery     = ``
	newEventsQuery       = ``
)

func (s *Server) tick(ender chan bool, workQueue chan *work, workerQueue chan chan *work) {
	ticker := time.NewTicker(time.Second * 5)

	dispatcher(s.repos, workQueue, workerQueue)

	go func() {
		for {
			select {
			case <-ticker.C:
				integrations, _ := s.database.readIntegrations(newIntegrationsQuery)
				settings, _ := s.database.readSettings(newSettingsQuery)
				events, _ := s.database.readEvents(newEventsQuery)

				w := make(map[int64]*work)

				for k, i := range integrations {
					w[k].repoID = k
					w[k].integration = i
				}
				for k, s := range settings {
					w[k].repoID = k
					w[k].setting = s
				}
				for k, e := range events {
					w[k].repoID = k
					w[k].events = e
				}

				collector(w, workQueue)

			case <-ender:
				ticker.Stop()
				close(ender)
				return
			}
		}
	}()
}
