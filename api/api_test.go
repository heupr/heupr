package api

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var indexNames = []string{
	"index.test",
	"index.create.test",
	"index.read.test",
	"index.delete.test",
	"process.test",
	"process.insertmultiple.test",
	"process.insertsingle.test",
	"heupr.bleve",
}

func TestMain(m *testing.M) {
	code := m.Run()

	for _, name := range indexNames {
		if err := os.RemoveAll(name); err != nil {
			log.Fatalf("error deleting directory: %s", err.Error())
		}
	}

	os.Exit(code)
}

type MockIntegration struct {
	result []byte
	err    error
}

func (i MockIntegration) Learn(input []byte) ([]byte, error) {
	return i.result, i.err
}

func (i MockIntegration) Predict(input []byte) ([]byte, error) {
	return i.result, i.err
}

func Test_learn(t *testing.T) {
	tests := []struct {
		desc   string
		method string
		body   string
		result []byte
		err    error
		status int
		resp   string
	}{
		{
			"learn method error scenario",
			"POST",
			"request-body",
			[]byte(""),
			errors.New("new-error"),
			500,
			"new-error",
		},
		{
			"successful request scenario",
			"POST",
			"request-body",
			[]byte("learn-method-result"),
			nil,
			200,
			"learn-method-result",
		},
	}

	for _, test := range tests {
		req, err := http.NewRequest(test.method, "/learn", bytes.NewBufferString(test.body))
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()

		mock := MockIntegration{
			result: test.result,
			err:    test.err,
		}
		handler := http.HandlerFunc(learn(mock))

		handler.ServeHTTP(rec, req)

		if status := rec.Code; status != test.status {
			t.Errorf("description: %s, handler returned incorrect status code, received: %v, expected: %v", test.desc,
				status, test.status)
		}

		if resp := rec.Body.String(); !strings.Contains(resp, test.resp) {
			t.Errorf("description: %s, handler returned incorrect response, received: %v, expected: %v",
				test.desc, resp, test.resp)
		}
	}
}

func Test_predict(t *testing.T) {
	tests := []struct {
		desc   string
		method string
		body   string
		result []byte
		err    error
		status int
		resp   string
	}{
		{
			"predict method error scenario",
			"POST",
			"request-body",
			[]byte(""),
			errors.New("new-error"),
			500,
			"new-error",
		},
		{
			"successful request scenario",
			"POST",
			"request-body",
			[]byte("predict-method-result"),
			nil,
			200,
			"predict-method-result",
		},
	}

	for _, test := range tests {
		req, err := http.NewRequest(test.method, "/predict", bytes.NewBufferString(test.body))
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()

		mock := MockIntegration{
			result: test.result,
			err:    test.err,
		}
		handler := http.HandlerFunc(predict(mock))

		handler.ServeHTTP(rec, req)

		if status := rec.Code; status != test.status {
			t.Errorf("description: %s, handler returned incorrect status code, received: %d, expected: %d", test.desc,
				status, test.status)
		}

		if resp := rec.Body.String(); !strings.Contains(resp, test.resp) {
			t.Errorf("description: %s, handler returned incorrect response, received: %s, expected: %s",
				test.desc, resp, test.resp)
		}
	}
}

func TestNewServer(t *testing.T) {
	i := MockIntegration{}
	s := New(i, "127.0.0.1")

	if s == nil || s.server.Addr == "" || s.server.Handler == nil {
		t.Errorf("server created incorrectly: %+v", s)
	}
}
