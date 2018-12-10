package backend

import (
	"database/sql"
	"fmt"

	"github.com/BurntSushi/toml"
)

type dataAccess interface {
	readIntegrations(query string) (map[int64]*integration, error)
	readSettings(query string) (map[int64]*settings, error)
	readEvents(query string) (map[int64][]*container, error)
}

type sqlDB interface {
	Close() error
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type database struct {
	sqlDB sqlDB
}

func (d *database) readIntegrations(query string) (map[int64]*integration, error) {
	// Reads in all integrations; called by Start() and potentially by tick().
	// Note that the returned map key will be redundant to integration.repoID.
	return nil, nil
}

func (d *database) readSettings(query string) (map[int64]*settings, error) {
	allSettings := make(map[int64]*settings)

	rows, err := d.sqlDB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("database read settings: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var content string
		if err := rows.Scan(&id, &content); err != nil {
			return nil, fmt.Errorf("database read settings row scan: %v", err)
		}

		set := settings{}
		if _, err := toml.Decode(content, &set); err != nil {
			return nil, fmt.Errorf("database read settings toml decode: %v", err)
		}
		allSettings[id] = &set
	}

	return allSettings, nil
}

func (d *database) readEvents(query string) (map[int64][]*container, error) {
	// Reads in all events; called by Start() and possibly tick().
	return nil, nil
}
