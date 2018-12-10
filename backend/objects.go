package backend

import (
	"encoding/json"

	"github.com/google/go-github/github"
)

type issueOptions struct {
	// Assignment response options
	Blacklist   []string `toml:"blacklist"`
	AsComment   bool     `toml:"as_comment"`
	MaxAssigned int      `toml:"max_assigned"`

	// Label response options
	DefaultLabels []string `toml:"default_labels"`
	LabelTypes    []string `toml:"label_types"`
}

type settings struct {
	Title string                             `toml:"title"`
	Issue map[string]map[string]issueOptions `toml:"issue"`
}

type integration struct {
	InstallationID int64
	AppID          int
	RepoID         int64
}

type container struct {
	RepoID      int64
	Event       string
	Action      string
	Payload     json.RawMessage
	Issue       *github.Issue
	PullRequest *github.PullRequest
	Linked      map[string][]*container
}

type work struct {
	RepoID      int64
	settings    *settings
	integration *integration
	Events      []*container
}
