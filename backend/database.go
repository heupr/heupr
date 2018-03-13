package backend

import (
	"database/sql"

	"heupr/backend/response/preprocess"
)

type settings struct {
	// .heupr.toml file contents
	// integration information
	// installation information
}

type integration struct {
	installID int64
	appID     int
	repoID    int64
}

type accessDB interface {
	read() (map[int64][]*preprocess.Container, error)
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

func (d *database) read() (map[int64][]*preprocess.Container, error) {
	// used to pull out new objects; called by timer() method
	return nil, nil
}

func (d *database) readSettings() ([]*settings, error) {
	// used on Re/Start to boot up necessary repo settings into memory
	return nil, nil
}

func (d *database) readIntegrations() ([]integration, error) {
	return nil, nil
}

func (d *database) readIntegrationByRepoID(repoID int64) (integration, error) {
	return integration{}, nil
}
