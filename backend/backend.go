package backend

import "github.com/google/go-github/v28/github"

// Payload defines the value passed between frontend resources and backend packages
type Payload interface {
	Type() string
	Bytes() []byte
	Config() []byte
}

// Backend defines the contract packages must follow for use with the application
type Backend interface {
	Configure(*github.Client)
	Prepare(Payload) error
	Act(Payload) error
}
