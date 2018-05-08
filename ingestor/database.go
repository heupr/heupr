package ingestor

import (
	"bytes"
	"database/sql"

	"github.com/google/go-github/github"
)

type dataAccess interface {
	InsertRepositoryIntegration(appID, repoID, installationID int64)
	DeleteRepositoryIntegration(appID, repoID, installationID int64)
	ObliterateIntegration(appID, installationID int64)
	ReadIntegrationByRepoID(repoID int64) (*integration, error)
	InsertBulkIssues(issues []*github.Issue)
	InsertBulkPullRequests(pulls []*github.PullRequest)
	InsertTOML(content string)
}

type sqlDB interface {
	Close() error
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
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
	AppID          int64
	InstallationID int64
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

func (d *database) InsertRepositoryIntegration(appID, repoID, installationID int64) {
	var buffer bytes.Buffer
	integrationsInsert := "INSERT INTO integrations(repo_id, app_id, installation_id) VALUES"
	valuesFmt := "(?,?,?)"

	buffer.WriteString(integrationsInsert)
	buffer.WriteString(valuesFmt)
	result, err := d.sqlDB.Exec(buffer.String(), repoID, appID, installationID)
	if err != nil {
		_ = err
	}
	rows, _ := result.RowsAffected()
	_ = rows
}

func (d *database) DeleteRepositoryIntegration(appID, repoID, installationID int64) {
	result, err := d.sqlDB.Exec("DELETE FROM integrations where repo_id = ? and app_id = ? and installation_id = ?", repoID, appID, installationID)
	if err != nil {
		_ = err
	}
	rows, _ := result.RowsAffected()
	_ = rows
}

func (d *database) ObliterateIntegration(appID, installationID int64) {
	result, err := d.sqlDB.Exec("DELETE FROM integrations where app_id = ? and installation_id = ?", appID, installationID)
	if err != nil {
		_ = err
	}
	rows, _ := result.RowsAffected()
	_ = rows
}

func (d *database) ReadIntegrationByRepoID(repoID int64) (*integration, error) {
	intg := new(integration)
	err := d.sqlDB.QueryRow("SELECT repo_id, app_id, installation_id FROM integrations WHERE repo_id = ?", repoID).Scan(&intg.RepoID, &intg.AppID, &intg.InstallationID)
	if err != nil {
		if err != sql.ErrNoRows {
			_ = err
		}
		return nil, err
	}
	return intg, nil
}

func (d *database) InsertBulkIssues(issues []*github.Issue) {
	// TODO: Implement this method.
}

func (d *database) InsertBulkPullRequests(pulls []*github.PullRequest) {
	// TODO: Implement this method.
}

func (d *database) InsertTOML(content string) {
	// TODO: Implement this method
}
