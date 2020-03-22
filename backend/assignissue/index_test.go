package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/blevesearch/bleve"
)

type mockS3Client struct {
	putObjectOutput *s3.PutObjectOutput
	putObjectErr    error
	getObjectOutput *s3.GetObjectOutput
	getObjectErr    error
}

func (m *mockS3Client) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return m.putObjectOutput, m.putObjectErr
}

func (m *mockS3Client) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return m.getObjectOutput, m.getObjectErr
}

func Test_putIndex(t *testing.T) {
	tests := []struct {
		desc      string
		path      string
		putOutput *s3.PutObjectOutput
		putErr    error
		err       string
	}{
		{
			desc:      "no file available",
			path:      "",
			putOutput: nil,
			putErr:    nil,
			err:       "open /index_meta.json: no such file or directory",
		},
		{
			desc:      "error putting file in s3",
			path:      "test.bleve",
			putOutput: nil,
			putErr:    errors.New("mock put error"),
			err:       "mock put error",
		},
		{
			desc:      "successful invocation",
			path:      "put_test.bleve",
			putOutput: nil,
			putErr:    nil,
			err:       "",
		},
	}

	for _, test := range tests {
		h := &indexHelp{
			s3: &mockS3Client{
				putObjectOutput: test.putOutput,
				putObjectErr:    test.putErr,
			},
		}

		if test.path != "" {
			err := os.Mkdir(test.path, 0755)
			if err != nil {
				t.Fatalf("description: %s, error: %s", test.desc, err.Error())
			}

			for _, filename := range bleveFiles {
				if err := ioutil.WriteFile(filepath.Join(test.path, filename), []byte("test"), 0644); err != nil {
					t.Fatalf("description: %s, error: %s", test.desc, err.Error())
				}
			}
		}

		err := h.putIndex(test.path)
		if err != nil && err.Error() != test.err {
			t.Errorf("description: %s, received: %s, expected: %s", test.desc, err.Error(), test.err)
		}

		os.RemoveAll(test.path)
	}
}

func Test_getIndex(t *testing.T) {
	tests := []struct {
		desc      string
		path      string
		getOutput *s3.GetObjectOutput
		getErr    error
		err       string
	}{
		{
			desc:      "no path provided",
			path:      "",
			getOutput: nil,
			getErr:    nil,
			err:       "mkdir : no such file or directory",
		},
		{
			desc:      "error get file from s3",
			path:      "get_test.bleve",
			getOutput: nil,
			getErr:    errors.New("mock get error"),
			err:       "mock get error",
		},
		{
			desc: "successful invocation",
			path: "get_test.bleve",
			getOutput: &s3.GetObjectOutput{
				Body: ioutil.NopCloser(bytes.NewBufferString("test")),
			},
			getErr: nil,
			err:    "",
		},
	}

	for _, test := range tests {
		h := &indexHelp{
			s3: &mockS3Client{
				getObjectOutput: test.getOutput,
				getObjectErr:    test.getErr,
			},
		}

		err := h.getIndex(test.path)
		if err != nil && err.Error() != test.err {
			t.Errorf("description: %s, received: %s, expected: %s", test.desc, err.Error(), test.err)
		}

		os.RemoveAll(test.path)
	}
}

type mockIndexHelper struct {
	putIndexErr error
	getIndexErr error
}

func (m *mockIndexHelper) putIndex(path string) error {
	return m.putIndexErr
}

func (m *mockIndexHelper) getIndex(path string) error {
	return m.getIndexErr
}

func Test_index(t *testing.T) {
	tests := []struct {
		desc        string
		putIndexErr error
		err         string
	}{
		{
			desc:        "error putting into index",
			putIndexErr: errors.New("mock put error"),
			err:         "mock put error",
		},
		{
			desc:        "successful invocation",
			putIndexErr: nil,
			err:         "",
		},
	}

	for _, test := range tests {
		repo, key, value := "test-index.bleve", "test-key", "test-value"

		c := &client{
			help: &mockIndexHelper{
				putIndexErr: test.putIndexErr,
			},
		}

		err := c.index(repo, key, value)
		if err != nil && err.Error() != test.err {
			t.Errorf("description: %s, received: %s, expected: %s", test.desc, err.Error(), test.err)
		}

		os.RemoveAll(repo)
	}
}

type data struct {
	Corpus string
}

func Test_search(t *testing.T) {
	tests := []struct {
		desc        string
		path        string
		data        []data
		getIndexErr error
		id          string
		err         string
	}{
		{
			desc:        "error getting index file",
			path:        "",
			data:        nil,
			getIndexErr: errors.New("mock search error"),
			id:          "",
			err:         "mock search error",
		},
		{
			desc: "successful invocation",
			path: "test-search.bleve",
			data: []data{
				data{
					Corpus: "first-blob",
				},
				data{
					Corpus: "second-content",
				},
				data{
					Corpus: "third-text",
				},
			},
			getIndexErr: nil,
			id:          "test-key-2",
			err:         "",
		},
	}

	for _, test := range tests {
		c := &client{
			help: &mockIndexHelper{
				getIndexErr: test.getIndexErr,
			},
		}

		if test.path != "" {
			mapping := bleve.NewIndexMapping()
			index, err := bleve.New(test.path, mapping)
			if err != nil {
				t.Fatal(err)
			}

			for i, data := range test.data {
				if err := index.Index("test-key-"+strconv.Itoa(i), data); err != nil {
					t.Fatal(err)
				}
			}

			index.Close()
		}

		id, err := c.search(test.path, "text")
		if err != nil && err.Error() != test.err {
			t.Errorf("description: %s, error received: %s, expected: %s", test.desc, err.Error(), test.err)
		}

		if id != test.id {
			t.Errorf("description: %s, id received: %s, expected: %s", test.desc, id, test.id)
		}

		os.RemoveAll(test.path)
	}
}
