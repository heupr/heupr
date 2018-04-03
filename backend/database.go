package backend

import (
	"database/sql"

	"heupr/backend/response"
	"heupr/backend/response/preprocess"
)

type integration struct {
	installationID int64
	appID          int
	repoID         int64
}

type settings struct {
	Title  string
	Issues map[string]map[string]response.Options
}

type dataAccess interface {
	readIntegrations(query string) (map[int64]*integration, error)
	readSettings(query string) (map[int64]*settings, error)
	readEvents(query string) (map[int64][]*preprocess.Container, error)
}

type sqlDB interface {
	Close() error
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type database struct {
	sqlDB sqlDB
}

func (d *database) readIntegrations(query string) (map[int64]*integration, error) {
	// reads in all integrations; called by Start()
	// note that the returned map key will be redundant to integration.repoID
	// note the argument may make this flexible to also be called by timer()
	return nil, nil
}

func (d *database) readSettings(query string) (map[int64]*settings, error) {
	// reads in all settings; called by Start
	return nil, nil
}

func (d *database) readEvents(query string) (map[int64][]*preprocess.Container, error) {
	// reads in all events; called by Start
	return nil, nil
}
