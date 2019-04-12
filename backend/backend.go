package backend

// Backend provides the basis for implementing the Integration interface
type Backend struct {
}

// New generates the default resource for processing events
func New() (*Backend, error) {
	// outline:
	// [ ] check if elastic available
	// - [ ] false: run command line download
	// - [ ] https://www.elastic.co/guide/en/elasticsearch/reference/current/zip-targz.html
	// - [ ] true: continue

	// outline:
	// [ ] create new elasticsearch instance
	// [ ] test instance health
	// [ ] add to struct field
	// [ ] return backend w/ instance running
	// note: http://olivere.github.io/elastic/
	// note: https://github.com/elastic/go-elasticsearch
	// note: https://www.elastic.co/downloads/elasticsearch

	return &Backend{}, nil
}

// Learn implements the Integration interface
func (b *Backend) Learn(input []byte) ([]byte, error) {
	// outline:
	// [ ] parse input to json array
	// [ ] call feature web logic
	// - [ ] pass in json object iteratively
	// - [ ] generate json array output object
	// [ ] call index on elastic search
	// - [ ] insert objects to features index
	// [ ] return json object response byte encoded

	return []byte{}, nil
}

// Predict implements the Integration interface
func (b *Backend) Predict(input []byte) ([]byte, error) {
	// outline:
	// [ ] parse input to json object
	// [ ] call search on elastic search
	// - [ ] retrieve/parse received json object result
	// [ ] call feature web logic
	// - [ ] pass in json object
	// - [ ] generate json array output object
	// [ ] call index on elastic search
	// - [ ] insert objects to features index
	// [ ] return json object response byte encoded
	// - [ ] include output prediction

	return []byte{}, nil
}
