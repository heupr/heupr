package response

import "heupr/backend/process/preprocess"

// Model is used for complicated response features.
type Model interface {
	Learn(input []*preprocess.Container) error
	Predict(input *preprocess.Container) (interface{}, error)
}

// Conditional is used for simple response features.
type Conditional interface {
	React(input *preprocess.Container) (interface{}, error)
}

// Action houses Options for responses and the required normalization.
type Action struct {
	Options     preprocess.Options
	Model       Model
	Conditional Conditional
}
