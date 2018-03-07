package backend

import (
	"database/sql"
)

type settings struct{}

type integration struct {
	installID int64
	appID     int
	repoID    int64
}

type accessDB interface {
	readSettings() ([]settings, error)
	readIntegrations() ([]integration, error)
	readIntegrationByRepoID(repoID int64) (integration, error)
}

type sqlDB interface {
	Open(driverName, dataSourceName string) (*sql.DB, error)
	Close() error
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type database struct {
	sqlDB sqlDB
}

func (d *database) readSettings() ([]*settings, error) {
	return nil, nil
}

func (d *database) readIntegrations() ([]integration, error) {
	return nil, nil
}

func (d *database) readIntegrationByRepoID(repoID int64) (integration, error) {
	return integration{}, nil
}
