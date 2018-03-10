package preprocess

// Container is the generalied internal object for processing.
type Container struct {
	event   string
	action  string
	payload string
	linked  map[string][]*Container
}

// Executor creates a uniform preprocessing step before training.
type Executor interface {
	Execute(input []*Container) ([]*Container, error)
}
