package ingestor

import (
	"database/sql"

	"github.com/google/go-github/github"
)

type dataAccess interface {
	InsertRepositoryIntegration(appID int, repoID, installationID int64)
	DeleteRepositoryIntegration(appID int, repoID, installationID int64)
	ObliterateIntegration(appID int, installationID int64)
	ReadIntegrationByRepoID(repoID int64) (*integration, error)
}

type sqlDB interface {
	Close() error
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type database struct {
	sqlDB sqlDB
	pool  pool
}

type event struct {
	Type    string            `json:"type"`
	Repo    github.Repository `json:"repo"`
	Action  string            `json:"action"`
	Payload interface{}       `json:"payload"`
}

type integration struct {
	RepoID         int64
	AppID          int
	InstallationID int
}

type eventType int

const (
	pullRequest eventType = iota
	issue
	all
)

type eventQuery struct {
	Type eventType
	Repo string
}

func (d *database) InsertRepositoryIntegration(appID int, repoID, installationID int64) {}

func (d *database) DeleteRepositoryIntegration(appID int, repoID, installationID int64) {}

func (d *database) ObliterateIntegration(appID int, installationID int64) {}

func (d *database) ReadIntegrationByRepoID(repoID int64) (*integration, error) {
	return &integration{}, nil
}
