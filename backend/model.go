package backend

type model interface {
	learn(input []*container) error
	predict(input *container) error
}
