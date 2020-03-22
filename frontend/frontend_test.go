package frontend

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
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

func stringEvent(event interface{}) string {
	output, err := json.Marshal(event.(*github.InstallationRepositoriesEvent))
	if err != nil {
		log.Fatalf("error creating event string: %s", err.Error())
	}

	return string(output)
}

func encode(secret string, payload []byte) string {
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write(payload)
	sig := "sha1=" + hex.EncodeToString(mac.Sum(nil))
	return sig
}

func Test_parseEvent(t *testing.T) {
	sec := "test-secret"
	sig := encode(sec, []byte(stringEvent(&github.InstallationRepositoriesEvent{
		Action: stringPtr("removed"),
		Installation: &github.Installation{
			ID: int64Ptr(2224),
		},
		RepositoriesAdded: []*github.Repository{
			&github.Repository{
				Name: stringPtr("cody"),
				Owner: &github.User{
					Login: stringPtr("commander"),
				},
			},
		},
	})))
	body := []byte(stringEvent(&github.InstallationRepositoriesEvent{
		Action: stringPtr("removed"),
		Installation: &github.Installation{
			ID: int64Ptr(2224),
		},
		RepositoriesAdded: []*github.Repository{
			&github.Repository{
				Name: stringPtr("cody"),
				Owner: &github.User{
					Login: stringPtr("commander"),
				},
			},
		},
	}))

	if err := validateEvent(sec, sig, body); err != nil {
		t.Errorf("description: error parsing event, error: %s", err.Error())
	}
}
