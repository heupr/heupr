package normalizer

// Container is the generalied internal object for processing.
type Container struct {
	event   string
	action  string
	payload string
	linked  map[string][]*Container
}
