package response

import "heupr/backend/process/preprocess"

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
	Options     Options
	Model       Model
	Conditional Conditional
}
