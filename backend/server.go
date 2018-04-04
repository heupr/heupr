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

	integrations, _ := s.database.readIntegrations(integrationQuery)
	settings, _ := s.database.readSettings(settingsQuery)

	var wg sync.WaitGroup
	wg.Add(len(integrations))

	for i := range integrations {
		go func() {
			repo, _ := newRepo(settings[i], integrations[i])
			s.repos.internal[i] = repo
			wg.Done()
		}()
	}

	wiggin := make(chan bool)
	s.timer(wiggin)
}

var (
	newIntegrationsQuery = ``
	newSettingsQuery     = ``
	newEventsQuery       = ``
)

func (s *Server) timer(ender chan bool) {
	ticker := time.NewTicker(time.Second * 5)

	if err := dispatcher(s.repos, 10); err != nil {
		_ = err
	}

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

				collector(w)

			case <-ender:
				ticker.Stop()
				close(ender)
				return
			}
		}
	}()
}
