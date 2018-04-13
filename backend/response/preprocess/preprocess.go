package preprocess

import (
	"encoding/json"
)

// Container is the generalized internal object for processing.
type Container struct {
	repoID  int64
	event   string
	action  string
	payload json.RawMessage
	linked  map[string][]*Container
}
