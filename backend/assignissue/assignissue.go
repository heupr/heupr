package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/bbalet/stopwords"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/go-yaml/yaml"
	"github.com/google/go-github/v28/github"

	"github.com/heupr/heupr/backend"
)

/*
Description:
`assignissue` plugin provides automated GitHub issue assignment.

Setup:
In the `.heupr.yml` file, include a backend option:

```
backends:
- name: assignissue
  events:
  - name: issues
    actions:
    - closed
  settings:
    es_endpoint: 'https://example-es-endpoint.com'
    contributors:
      - example_github_username
  location: http://s3-aws-region.amazonaws.com/heupr/assignissue.so
```

- The `es_endpoint` should be a valid URL at which an ElasticSearch
  index is available.
- The `contributors` array should be GitHub usernames for possible
  individuals to assign issues to.

Notes:
Currently, the logic is relatively simple for identifying which user
is assigned a text corpus it and only looks for "assignee" fields on
previously closed issues for training. Additional logic could be
added in the future to include things like commit messages that close
issues or the text bodies for associated pull requests.
*/

type webhookEvent struct {
	Name    string   `yaml:"name"`
	Actions []string `yaml:"actions"`
}

type settings struct {
	ESEndpoint   string   `yaml:"es_endpoint"`
	Contributors []string `yaml:"contributors"`
}

type backendObj struct {
	Name     string         `yaml:"name"`
	Events   []webhookEvent `yaml:"events"`
	Settings settings       `yaml:"settings"`
	Location string         `yaml:"location"`
}

type configObj struct {
	Backends []backendObj `yaml:"backends"`
}

type helper interface {
	getContent(c *github.Client, owner, repo, path string) (string, error)
	getText(issue *github.Issue) string
	listIssues(c *github.Client, owner, repo string) ([]*github.Issue, error)
	addAssignee(c *github.Client, owner, repo, actor string, number int) error
}

type help struct{}

func (h *help) getContent(c *github.Client, owner, repo, path string) (string, error) {
	opts := &github.RepositoryContentGetOptions{}
	file, _, _, err := c.Repositories.GetContents(context.Background(), owner, repo, path, opts)
	if err != nil {
		return "", errors.New("error getting content: " + err.Error())
	}

	return *file.Content, nil
}

func (h *help) getText(issue *github.Issue) string {
	// NOTE: Add checks for empty title/body fields.
	title := stopwords.CleanString(strings.ToLower(*issue.Title), "en", false)
	body := stopwords.CleanString(strings.ToLower(*issue.Body), "en", false)

	return title + " " + body
}

func (h *help) addAssignee(c *github.Client, owner, repo, actor string, number int) error {
	_, _, err := c.Issues.AddAssignees(context.Background(), owner, repo, number, []string{actor})
	return err
}

func (h *help) listIssues(c *github.Client, owner, repo string) ([]*github.Issue, error) {
	output := []*github.Issue{}

	opts := &github.IssueListByRepoOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		issues, resp, err := c.Issues.ListByRepo(context.Background(), owner, repo, opts)
		if err != nil {
			return nil, err
		}
		output = append(output, issues...)

		if resp.NextPage == 0 {
			break
		} else {
			opts.ListOptions.Page = resp.NextPage
		}
	}

	return output, nil
}

// Backend implements the backend package interface
type Backend struct {
	client *github.Client
	help   helper
}

// Configure configures the backend with a client and helper struct
func (b *Backend) Configure(c *github.Client) {
	b.client = c
	b.help = &help{}
}

type es interface {
	index(string, string) error
	search(string) (string, error)
}

type client struct {
	esClient *elasticsearch.Client
}

var newClient = func(url string) (es, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{url},
	}
	c, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating client: %s", err.Error())
	}

	return &client{
		esClient: c,
	}, nil
}

func (c *client) index(key, value string) error {
	request := esapi.IndexRequest{
		Index:   "assignment",
		Body:    strings.NewReader(`{"actor" : "` + key + `", "blob" : "` + value + `"}`),
		Refresh: "true",
	}
	response, err := request.Do(context.Background(), c.esClient)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return nil
}

func (c *client) search(blob string) (string, error) {
	response, err := c.esClient.Search(
		c.esClient.Search.WithContext(context.Background()),
		c.esClient.Search.WithIndex("assignment"),
		c.esClient.Search.WithBody(strings.NewReader(`{"query" : { "match" : { "blob" : "`+blob+`" } }}`)),
	)
	if err != nil {
		return "", fmt.Errorf("error performing search: %s", err.Error())
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)

	return buf.String(), nil
}

// Prepare processes existing pull requests and calculates points estimates versus actual
func (b *Backend) Prepare(p backend.Payload) error {
	installation := github.InstallationRepositoriesEvent{}

	if err := json.Unmarshal(p.Bytes(), &installation); err != nil {
		return fmt.Errorf("error unmarshalling installation: %s", err.Error())
	}

	for _, repo := range installation.RepositoriesAdded {
		owner := *repo.Owner.Login
		repo := *repo.Name

		fileContent, err := b.help.getContent(b.client, owner, repo, ".heupr.yml")
		if err != nil {
			return fmt.Errorf("error getting heupr config: %s", err.Error())
		}

		config := configObj{}
		if err := yaml.Unmarshal([]byte(fileContent), &config); err != nil {
			return fmt.Errorf("error parsing heupr config: %s", err.Error())
		}

		issues, err := b.help.listIssues(b.client, owner, repo)
		if err != nil {
			return fmt.Errorf("error getting issues: %s", err.Error())
		}

		indexContent := make(map[string]string)
		for _, issue := range issues {
			if *issue.State == "closed" && issue.Assignee != nil {
				actor := *issue.Assignee.Login
				if _, ok := indexContent[actor]; ok {
					indexContent[actor] += b.help.getText(issue)
				} else {
					indexContent[actor] = b.help.getText(issue)
				}
			}
		}

		esEndpoint := ""
		for _, bknd := range config.Backends {
			if bknd.Name == "assignissue" {
				esEndpoint = bknd.Settings.ESEndpoint
			}
		}

		esClient, err := newClient(esEndpoint)
		if err != nil {
			return fmt.Errorf("error creating es client: %s", err.Error())
		}
		for actor, corpus := range indexContent {
			if err := esClient.index(actor, corpus); err != nil {
				return fmt.Errorf("error indexing key/value: %s", err.Error())
			}
		}
	}

	return nil
}

// Act processes new pull requests and calculates points estimates versus actual
func (b *Backend) Act(p backend.Payload) error {
	if p.Type() != "issues" {
		return nil
	}

	event := github.IssuesEvent{}
	if err := json.Unmarshal(p.Bytes(), &event); err != nil {
		return fmt.Errorf("error parsing issue: %s", err.Error())
	}

	if *event.Action != "closed" {
		return nil // (?)
	}

	corpus := b.help.getText(event.Issue)

	owner := *event.Issue.Repository.Owner.Login
	repo := *event.Issue.Repository.Name

	fileContent, err := b.help.getContent(b.client, owner, repo, ".heupr.yml")
	if err != nil {
		return fmt.Errorf("error getting heupr config: %s", err.Error())
	}

	config := configObj{}
	if err := yaml.Unmarshal([]byte(fileContent), &config); err != nil {
		return fmt.Errorf("error parsing heupr config: %s", err.Error())
	}

	esEndpoint := ""
	contributors := []string{}
	for _, bknd := range config.Backends {
		if bknd.Name == "assignissue" {
			esEndpoint = bknd.Settings.ESEndpoint
			contributors = bknd.Settings.Contributors
		}
	}

	esClient, err := newClient(esEndpoint)
	if err != nil {
		return fmt.Errorf("error creating es client: %s", err.Error())
	}

	actor, err := esClient.search(corpus)
	if err != nil {
		return fmt.Errorf("error searching index: %s", err.Error())
	}

	for _, contributor := range contributors {
		if contributor == actor {
			if err := b.help.addAssignee(b.client, owner, repo, actor, *event.Issue.Number); err != nil {
				return fmt.Errorf("error adding assignee: %s", err.Error())
			}
		}
	}

	return nil
}

func main() {}
