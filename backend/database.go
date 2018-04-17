package backend

import (
	"database/sql"

	"heupr/backend/process"
	"heupr/backend/process/preprocess"
)

type dataAccess interface {
	readIntegrations(query string) (map[int64]*process.Integration, error)
	readSettings(query string) (map[int64]*process.Settings, error)
	readEvents(query string) (map[int64][]*preprocess.Container, error)
}

type sqlDB interface {
	Close() error
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type database struct {
	sqlDB sqlDB
}

func (d *database) readIntegrations(query string) (map[int64]*process.Integration, error) {
	// reads in all integrations; called by Start()
	// note that the returned map key will be redundant to integration.repoID
	// note the argument may make this flexible to also be called by tick()
	return nil, nil
}

func (d *database) readSettings(query string) (map[int64]*process.Settings, error) {
	// reads in all settings; called by Start
	return nil, nil
}

func (d *database) readEvents(query string) (map[int64][]*preprocess.Container, error) {
	// reads in all events; called by Start
	return nil, nil
}
