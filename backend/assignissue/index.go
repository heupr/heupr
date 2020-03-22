package main

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/blevesearch/bleve"

	"github.com/aws/aws-sdk-go/service/s3"
)

type s3Client interface {
	PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error)
	GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error)
}

type indexHelper interface {
	putIndex(path string) error
	getIndex(path string) error
}

type indexHelp struct {
	s3 s3Client
}

var bleveFiles = []string{"index_meta.json", "store"}

func (h *indexHelp) putIndex(path string) error {
	dir := strings.Replace(path, "/tmp/", "", -1)

	for _, filename := range bleveFiles {
		content, err := ioutil.ReadFile(path + "/" + filename)
		if err != nil {
			return err
		}

		input := &s3.PutObjectInput{
			Body:   bytes.NewReader(content),
			Bucket: stringPtr("heupr"),
			Key:    stringPtr(dir + "/" + filename),
		}

		_, err = h.s3.PutObject(input)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *indexHelp) getIndex(path string) error {
	key := strings.Replace(path, "/tmp/", "", -1)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0755); err != nil {
			return err
		}
	}

	for _, filename := range bleveFiles {
		input := &s3.GetObjectInput{
			Bucket: stringPtr("heupr"),
			Key:    stringPtr(key + "/" + filename),
		}

		output, err := h.s3.GetObject(input)
		if err != nil {
			return err
		}

		file, err := os.Create(path + "/" + filename)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(file, output.Body)
		if err != nil {
			return err
		}
	}

	return nil
}

type assigner interface {
	index(repo, key, value string) error
	search(repo, blob string) (string, error)
}

var newClient = func() assigner {
	return &client{
		help: &indexHelp{
			s3: s3.New(session.New()),
		},
	}
}

type client struct {
	help indexHelper
}

func (c *client) index(path, key, value string) error {
	log.Printf("path: %s, key: %s, value: %s\n", path, key, value)

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New(path, mapping)
	if err != nil {
		return err
	}

	data := struct {
		Corpus string
	}{
		Corpus: value,
	}

	if err := index.Index(key, data); err != nil {
		return err
	}

	index.Close()

	if err := c.help.putIndex(path); err != nil {
		return err
	}

	log.Println("successful index invocation")
	return nil
}

func (c *client) search(path, blob string) (string, error) {
	log.Printf("path: %s, blob: %s\n", path, blob)

	if err := c.help.getIndex(path); err != nil {
		return "", err
	}

	_, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	index, err := bleve.Open(path)
	if err != nil {
		return "", err
	}

	query := bleve.NewQueryStringQuery(blob)
	search := bleve.NewSearchRequest(query)
	searchResults, err := index.Search(search)
	if err != nil {
		return "", err
	}
	log.Printf("search results: %+v\n", searchResults)

	if searchResults.Total == uint64(0) {
		return "", errors.New("no search results found")
	}

	return searchResults.Hits[0].ID, nil
}
