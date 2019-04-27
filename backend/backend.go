package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/bbalet/stopwords"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type feature struct {
	ID         *int64  `json:"feature_id,omitempty"`
	Type       *string `json:"feature_type,omitempty"`
	Action     *string `json:"feature_action,omitempty"`
	Number     *int    `json:"feature_number,omitempty"`
	Title      *string `json:"feature_title,omitempty"`
	Body       *string `json:"feature_body,omitempty"`
	ActorID    *int64  `json:"feature_actor_id,omitempty"`
	ActorName  *string `json:"feature_actor_name,omitempty"`
	References []*int  `json:"feature_references,omitempty"`
	Referenced []*int  `json:"feature_referenced,omitempty"` // NOTE: This is not currently used
}

type elastcisearch interface {
	search(string) (string, error)
	index(string, string) error
}

type client struct {
	client *elasticsearch.Client
}

func (c *client) search(blob string) (string, error) {
	response, err := c.client.Search(
		c.client.Search.WithContext(context.Background()),
		c.client.Search.WithIndex("assignment"),
		c.client.Search.WithBody(strings.NewReader(`{"query" : { "match" : { "blob" : "`+blob+`" } }}`)),
	)
	if err != nil {
		return "", fmt.Errorf("error performing search: %s", err.Error())
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)

	return buf.String(), nil
}

func (c *client) index(key, value string) error {
	request := esapi.IndexRequest{
		Index:   "assignment",
		Body:    strings.NewReader(`{"actor" : "` + key + `", "blob" : "` + value + `"}`),
		Refresh: "true",
	}
	response, err := request.Do(context.Background(), c.client)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return nil
}

// Backend provides the basis for implementing the Integration interface
type Backend struct {
	esURL  string
	client elastcisearch
}

// New generates the default resource for processing events
func New() (*Backend, error) {
	b := &Backend{}

	url := os.Getenv("HEUPR_ES_URL")
	if url == "" {
		return nil, errors.New("no elasticsearch url set")
	}

	b.esURL = url

	cfg := elasticsearch.Config{
		Addresses: []string{
			b.esURL,
		},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating client: %s", err.Error())
	}

	b.client = &client{
		client: es,
	}

	return b, nil
}

func getText(ref int, ftrs []*feature) string {
	for _, ftr := range ftrs {
		if ref == *ftr.Number {
			title := stopwords.CleanString(strings.ToLower(*ftr.Title), "en", false)
			body := stopwords.CleanString(strings.ToLower(*ftr.Body), "en", false)

			return title + " " + body
		}
	}
	return ""
}

// Learn implements the Integration interface
func (b *Backend) Learn(input []byte) ([]byte, error) {
	features := []*feature{}

	if err := json.Unmarshal(input, &features); err != nil {
		return nil, fmt.Errorf("error unmarshaling: %s", err.Error())
	}

	for _, feature := range features {
		actor := feature.ActorName
		title := stopwords.CleanString(strings.ToLower(*feature.Title), "en", false)
		body := stopwords.CleanString(strings.ToLower(*feature.Body), "en", false)

		blob := title + " " + body

		output, err := b.client.search(*actor)
		if err != nil {
			return nil, fmt.Errorf("error searching: %s", err.Error())
		}

		if feature.References != nil || len(feature.References) != 0 {
			for _, reference := range feature.References {
				referenceBlob := getText(*reference, features)
				blob = blob + " " + referenceBlob
			}
		}

		if output != "" {
			blob = blob + " " + output
		}

		if err := b.client.index(*actor, blob); err != nil {
			return nil, fmt.Errorf("error indexing: %s", err.Error())
		}
	}

	return []byte("successful learn"), nil
}

// Predict implements the Integration interface
func (b *Backend) Predict(input []byte) ([]byte, error) {
	feature := &feature{}

	if err := json.Unmarshal(input, feature); err != nil {
		return nil, fmt.Errorf("error unmarshalling: %s", err.Error())
	}

	title := stopwords.CleanString(strings.ToLower(*feature.Title), "en", false)
	body := stopwords.CleanString(strings.ToLower(*feature.Body), "en", false)

	blob := title + " " + body

	assignee, err := b.client.search(blob)
	if err != nil {
		return nil, fmt.Errorf("error searching: %s", err.Error())
	}

	return []byte(assignee), nil
}
