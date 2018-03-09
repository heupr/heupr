package logic

import (
	"heupr/backend/logic/normalize"
)

// Model is used for complicated response features.
type Model interface {
	Learn(input []*normalize.Container) error
	Predict(input *normalize.Container) (interface{}, error)
}

// Conditional is used for simple response features.
type Conditional interface {
	React(input *normalize.Container) (interface{}, error)
}
