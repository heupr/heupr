package frontend

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/google/go-github/v28/github"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func Test_newClient(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/repos/temple/archives/contents/kamino.txt", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"content":"should be here"}`)
	})

	server := httptest.NewServer(mux)

	c := github.NewClient(nil)
	url, _ := url.Parse(server.URL + "/")
	c.BaseURL = url
	c.UploadURL = url

	output, err := getContent(c, "temple", "archives", "kamino.txt")
	if output == "" {
		t.Errorf("description: error getting file contents, received: %s", output)
	}

	if err != nil {
		t.Errorf("description: error getting file contents, error: %s", err.Error())
	}
}
