package backend

import (
	"database/sql"
	"fmt"

	"github.com/BurntSushi/toml"

	"heupr/backend/process/preprocess"
)

type dataAccess interface {
	readIntegrations(query string) (map[int64]*preprocess.Integration, error)
	readSettings(query string) (map[int64]*preprocess.Settings, error)
	readEvents(query string) (map[int64][]*preprocess.Container, error)
}

type sqlDB interface {
	Close() error
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type database struct {
	sqlDB sqlDB
}

func (d *database) readIntegrations(query string) (map[int64]*preprocess.Integration, error) {
	// reads in all integrations; called by Start()
	// note that the returned map key will be redundant to integration.repoID
	// note the argument may make this flexible to also be called by tick()
	return nil, nil
}

func (d *database) readSettings(query string) (map[int64]*preprocess.Settings, error) {
	settings := make(map[int64]*preprocess.Settings)

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

		set := preprocess.Settings{}
		if _, err := toml.Decode(content, &set); err != nil {
			return nil, fmt.Errorf("database read settings toml decode: %v", err)
		}
		settings[id] = &set
	}

	return settings, nil
}

func (d *database) readEvents(query string) (map[int64][]*preprocess.Container, error) {
	// reads in all events; called by Start
	return nil, nil
}
