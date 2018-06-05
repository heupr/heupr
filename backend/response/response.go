package response

import "heupr/backend/preprocess"

// Model is used for processing incoming events within Containers.
type Model interface {
	Learn(input []*preprocess.Container) error
	Act(input *preprocess.Container, opts *preprocess.Options) (interface{}, error)
}

// Action houses Options for responses and the required normalization.
type Action struct {
	Preprocess preprocess.Preprocessor
	Options    preprocess.Options
	Model      Model
}

// Learn is a wrapper to call the given Action's Model Learn method.
func (a *Action) Learn(input []*preprocess.Container) error {
	// call preprocess from field value
	// call learn method
	// return errors
	return nil
}

// Act is a wrapper to call the given Action's Model Act method.
func (a *Action) Act(input *preprocess.Container) (interface{}, error) {
	// call act method
	// pass in input + options field value
	// return value + errors
	results := true // TEMPORARY
	return results, nil
}
