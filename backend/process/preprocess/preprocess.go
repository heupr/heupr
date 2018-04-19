package preprocess

import "encoding/json"

// Options provides user settings to selected Actions
type Options struct {
	// Assignment response options
	Blacklist []string
	AsComment bool
	Count     int
	// Label response options
	Default []string
	Types   []string
}

// Settings represents the parsed .heupr.toml file for user settings.
type Settings struct {
	Title  string
	Issues map[string]map[string]Options
}

// Integration represents a new Heupr GitHub integration.
type Integration struct {
	InstallationID int64
	AppID          int
	RepoID         int64
}

// Container is the generalized internal object for processing.
type Container struct {
	repoID  int64
	event   string
	action  string
	payload json.RawMessage
	linked  map[string][]*Container
}

// Work holds the objects necessary for processing by the responses.
type Work struct {
	RepoID      int64
	Settings    *Settings
	Integration *Integration
	Events      []*Container
}

// Preprocessor completes necessary conflation/normalization/etc actions prior
// to passing Container objects into the Model/Conditional operations.
type Preprocessor interface {
	Preprocess(input []*Container) ([]*Container, error)
}
