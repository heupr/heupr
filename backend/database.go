package backend

import (
	"database/sql"

	"heupr/backend/response/preprocess"
)

type issues struct {
	// Assignment response options
	Blacklist []string
	AsComment bool
	Count     int
	// Label response options
	Default []string
	Types   []string
}

type settings struct {
	Title  string
	Issues map[string]map[string]issues
	// integration information
	// installation information
}

type integration struct {
	installationID int64
	appID          int
	repoID         int64
}

type dataAccess interface {
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
