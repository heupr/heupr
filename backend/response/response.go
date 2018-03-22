package response

import (
	"heupr/backend/response/preprocess"
)

// Model is used for complicated response features.
type Model interface {
	Learn(input []*preprocess.Container) error
	Predict(input *preprocess.Container) (interface{}, error)
}

// Conditional is used for simple response features.
type Conditional interface {
	React(input *preprocess.Container) (interface{}, error)
}

// Action houses options for responses and the required normalization.
type Action struct {
	Model       Model
	Conditional Conditional
}
