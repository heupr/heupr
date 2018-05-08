package ingestor

import "database/sql"

type repoInitService interface {
	addRepo(owner, repo string, c githubService)
	repoIntegrationExists(repoID int64) bool
}

type repoInit struct {
	database dataAccess
}

func (r *repoInit) addRepo(owner, repo string, c githubService) {
	issues, err := c.getIssues(owner, repo, "closed")
	if err != nil {
		_ = err // TODO: Log error correctly.
	}
	r.database.InsertBulkIssues(issues)

	pulls, err := c.getPulls(owner, repo, "closed")
	if err != nil {
		_ = err // TODO: Log error correctly.
	}
	r.database.InsertBulkPullRequests(pulls)

	file, err := c.getTOML(owner, repo)
	if err != nil {
		_ = err // TODO: Log error correctly.
	}
	r.database.InsertTOML(file)
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
