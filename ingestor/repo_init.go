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
	if issues, err := c.getIssues(owner, name, "closed"); err != nil {
		_ = err // TODO: Log error correctly.
	} else {
		for i := range issues {
			issues[i].Repository = repo
		}
		r.database.InsertBulkIssues(issues)
	}

	if pulls, err := c.getPulls(owner, name, "closed"); err != nil {
		_ = err // TODO: Log error correctly.
	} else {
		r.database.InsertBulkPullRequests(pulls)
	}

	if file, err := c.getTOML(owner, name); err != nil {
		_ = err // TODO: Log error correctly.
	} else {
		r.database.InsertTOML(file)
	}
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
