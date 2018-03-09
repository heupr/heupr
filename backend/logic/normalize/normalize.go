package normalize

// Container is the generalied internal object for processing.
type Container struct {
	event   string
	action  string
	payload string
	linked  map[string][]*Container
}

// Normalizer creates a uniform pre-processing step before training.
type Normalizer interface {
	Normalize(input []*Container) ([]*Container, error)
}
