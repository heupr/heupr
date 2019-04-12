package api

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/blevesearch/bleve"
)

type indexer interface {
	create(input *feature) error
	read(id string) (*feature, error)
	delete(id string) error
}

type index struct {
	store bleve.Index
}

func (i index) create(input *feature) error {
	if input.Number == nil {
		return errors.New("feature id cannot be nil")
	}

	return i.store.Index(strconv.Itoa(*input.Number), *input)
}

func (i index) read(id string) (*feature, error) {
	doc, err := i.store.Document(id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving feature: %s", err.Error())
	}

	if doc == nil {
		return nil, fmt.Errorf("no feature found in index")
	}

	featureID, err := strconv.Atoi(doc.ID)
	if err != nil {
		return nil, fmt.Errorf("error parsing id string: %s", err.Error())
	}

	return &feature{
		Number: &featureID,
	}, nil
}

func (i index) delete(id string) error {
	if id == "" {
		return fmt.Errorf("id must be non-empty string")
	}

	return i.store.Delete(id)
}

func newIndex(name string) (indexer, error) {
	mapping := bleve.NewIndexMapping()
	i, err := bleve.New(name, mapping)
	if err != nil {
		return nil, fmt.Errorf("error creating feature index: %s", err.Error())
	}

	return index{
		store: i,
	}, nil
}
