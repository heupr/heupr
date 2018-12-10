package backend

import (
	"database/sql"
	"time"
)

var (
	integrationQuery = ``
	settingsQuery    = ``
	eventsQuery      = ``
)

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

// Start is exported so that cmd/ has access to launch the backend.
func (s *Server) Start() {
	db, _ := openDatabase()
	s.database = db

	s.work = make(chan *work)
	s.workers = make(chan chan *work)

	s.repos = &repos{
		Internal: make(map[int64]*repo),
	}

	integrations, _ := s.database.readIntegrations(integrationQuery)
	settings, _ := s.database.readSettings(settingsQuery)

	for i := range integrations {
		// TODO: There should likely be a check for settings[i] existence.
		repo, _ := newRepo(settings[i], integrations[i])
		s.repos.Lock()
		s.repos.Internal[integrations[i].RepoID] = repo
		s.repos.Unlock()
	}

	wiggin := make(chan bool)
	s.tick(wiggin)
}

var (
	newIntegrationsQuery = ``
	newSettingsQuery     = ``
	newEventsQuery       = ``
)

func (s *Server) tick(ender chan bool) {
	ticker := time.NewTicker(time.Second * 5)

	dispatcher(s.repos, s.work, s.workers)

	go func() {
		for {
			integrations, _ := s.database.readIntegrations(newIntegrationsQuery)
			settings, _ := s.database.readSettings(newSettingsQuery)
			events, _ := s.database.readEvents(newEventsQuery)

			w := make(map[int64]*work)

			for k, i := range integrations {
				w[k] = &work{
					RepoID:      k,
					integration: i,
				}
			}
			for k, s := range settings {
				w[k] = &work{
					RepoID:   k,
					settings: s,
				}
			}
			for k, e := range events {
				w[k] = &work{
					RepoID: k,
					Events: e,
				}
			}

			collector(w, s.work)

			select {
			case <-ticker.C:
				continue
			case <-ender:
				ticker.Stop()
				close(ender)
				return
			}
		}
	}()
}
