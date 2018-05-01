package ingestor

import "github.com/google/go-github/github"

// heuprInstallation represents an installation of the Heupr app.
type heuprInstallation struct {
	ID              *int64       `json:"id,omitempty"`
	Account         *github.User `json:"account,omitempty"`
	AppID           *int         `json:"app_id,omitempty"`
	AccessTokensURL *string      `json:"access_tokens_url,omitempty"`
	RepositoriesURL *string      `json:"repositories_url,omitempty"`
	HTMLURL         *string      `json:"html_url,omitempty"`
}

// heuprRepository represents a repo with the Heupr app installed on it.
type heuprRepository struct {
	ID       *int64  `json:"id,omitempty"`
	Name     *string `json:"name,omitempty"`
	FullName *string `json:"full_name,omitempty"`
}

// heuprInstallationEvent is the action that was performed. Can be either
// "created" or "deleted" and is a workaround to an API library limitation.
type heuprInstallationEvent struct {
	Action       *string            `json:"action,omitempty"`
	Sender       *github.User       `json:"sender,omitempty"`
	Installation *heuprInstallation `json:"installation,omitempty"`
	Repositories []heuprRepository  `json:"repositories,omitempty"`
}

// repoInstallationEvent is the repository action that was performed as it
// relates to the Heupr app and can be either "added" or "removed".
type repoInstallationEvent struct {
	Action              *string              `json:"action,omitempty"`
	RepositoriesAdded   []*github.Repository `json:"repositories_added,omitempty"`
	RepositoriesRemoved []*github.Repository `json:"repositories_removed,omitempty"`
	Sender              *github.User         `json:"sender,omitempty"`
	Installation        *heuprInstallation   `json:"installation,omitempty"`
}
