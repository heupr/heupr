package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v28/github"
)

func Test_getContent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/mustafar/mining-facility/contents/confederacy-leadership.json", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"content":"deceased"}`)
	})

	server := httptest.NewServer(mux)

	c := github.NewClient(nil)
	url, _ := url.Parse(server.URL + "/")
	c.BaseURL = url
	c.UploadURL = url

	h := help{}

	output, err := h.getContent(c, "mustafar", "mining-facility", "confederacy-leadership.json")
	if output == "" {
		t.Errorf("description: error getting file contents, received: %s", output)
	}

	if err != nil {
		t.Errorf("description: error getting file contents, error: %s", err.Error())
	}
}

func Test_getText(t *testing.T) {
	i := &github.Issue{
		Title: stringPtr("the tragedy of darth plagueis the wise"),
		Body:  stringPtr("not a story the jedi would tell "),
	}

	h := help{}

	output := h.getText(i)
	text := " tragedy darth plagueis wise   story jedi tell "
	if output != text {
		t.Errorf("description: error getting clean text, received: %s, expected: %s", output, text)
	}

}

func Test_listIssues(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/hoth/echo-base/issues", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"title":"1", "state":"closed", "assignee":{"login":"rebel-alliance"}}]`)
	})

	server := httptest.NewServer(mux)

	c := github.NewClient(nil)
	url, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	c.BaseURL = url
	c.UploadURL = url

	h := &help{}
	issues, err := h.listIssues(c, "hoth", "echo-base")
	if err != nil {
		t.Errorf("description: error listing issues, error received: %s", err.Error())
	}

	expected := 1
	if len(issues) != 1 {
		t.Errorf("description: error listing issues, issues received: %d, expected: %d", len(issues), expected)
	}
}

func Test_addAssignees(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/lars-homestead/new-droids/issues/1/assignees", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"number":1,"assignees":[{"login":"luke"}]}`)
	})

	server := httptest.NewServer(mux)

	c := github.NewClient(nil)
	url, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	c.BaseURL = url
	c.UploadURL = url

	h := help{}
	if err := h.addAssignee(c, "lars-homestead", "new-droids", "luke", 1); err != nil {
		t.Errorf("description: error assigning contributor, error received: %s", err.Error())
	}
}
