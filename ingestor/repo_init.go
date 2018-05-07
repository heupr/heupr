package ingestor

import (
	"database/sql"

	"github.com/google/go-github/github"
)

type repoInitService interface {
	addRepo(repo *github.Repository, c githubService)
	repoIntegrationExists(repoID int64) bool
}

type repoInit struct {
	database dataAccess
}

func (r *repoInit) addRepo(repo *github.Repository, c githubService) {
	owner, name := *repo.Owner.Login, *repo.Name
	issues, err := c.getIssues(owner, name, "closed")
	if err != nil {
		_ = err // TODO: Log error correctly.
	}
	r.database.InsertBulkIssues(issues)

	pulls, err := c.getPulls(owner, name, "closed")
	if err != nil {
		_ = err // TODO: Log error correctly.
	}
	r.database.InsertBulkPullRequests(pulls)
}

func (r *repoInit) repoIntegrationExists(repoID int64) bool {
	_, err := r.database.ReadIntegrationByRepoID(repoID)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		_ = err
		return false
	default:
		return true
	}
}
