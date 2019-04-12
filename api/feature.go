package api

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type processor interface {
	insertMultiple(input []byte) ([]*feature, error)
	insertSingle(input []byte) (*feature, error)
}

type process struct {
	index indexer
}

func newProcessor(name string) (processor, error) {
	i, err := newIndex(name)
	if err != nil {
		return nil, fmt.Errorf("error calling new index: %s", err.Error())
	}

	return &process{
		index: i,
	}, nil
}

type feature struct {
	Number     *int
	References []*int
	Referenced []*int // NOTE: This is not currently used
}

type object struct {
	ID     *int64  `json:"object_id,omitempty"`
	Type   *string `json:"object_type,omitempty"`
	Action *string `json:"object_action,omitempty"`
	Number *int    `json:"object_number,omitempty"`
	Title  *string `json:"object_title,omitempty"`
	Body   *string `json:"object_body,omitempty"`
}

func (p *process) insertMultiple(input []byte) ([]*feature, error) {
	objs := []*object{}

	if err := json.Unmarshal(input, &objs); err != nil {
		return nil, fmt.Errorf("error unmarshalling input: %s", err.Error())
	}

	re := regexp.MustCompile("(#[0-9]+)")

	ftrs := []*feature{}

	for _, obj := range objs {
		ftr := feature{}

		if obj.Number == nil {
			return nil, fmt.Errorf("no input object number")
		}

		ftr.Number = obj.Number

		references := re.FindAllString(*obj.Body, -1)

		for _, reference := range references {
			number, err := strconv.Atoi(strings.TrimPrefix(reference, "#"))
			if err != nil {
				return nil, fmt.Errorf("error parsing reference number: %s", err.Error())
			}
			ftr.References = append(ftr.References, &number)
		}

		p.index.create(&ftr)

		ftrs = append(ftrs, &ftr)
	}
	return ftrs, nil
}

func (p *process) insertSingle(input []byte) (*feature, error) {
	obj := object{}

	if err := json.Unmarshal(input, &obj); err != nil {
		return nil, fmt.Errorf("error unmarshalling input: %s", err.Error())
	}

	ftr := feature{}

	if obj.Number == nil {
		return nil, fmt.Errorf("no input object number")
	}

	ftr.Number = obj.Number

	re := regexp.MustCompile("(#[0-9]+)")

	references := re.FindAllString(*obj.Body, -1)

	for _, reference := range references {
		number, err := strconv.Atoi(strings.TrimPrefix(reference, "#"))
		if err != nil {
			return nil, fmt.Errorf("error parsing reference number: %s", err.Error())
		}
		ftr.References = append(ftr.References, &number)
	}

	p.index.create(&ftr)

	return &ftr, nil
}
