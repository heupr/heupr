package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

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
    - opened
  settings:
    contributors:
      - example_github_username
```

- The `contributors` array should be GitHub usernames for possible
  individuals to assign issues to.

Notes:
Currently, the logic is relatively simple for identifying which user
is assigned a text corpus it and only looks for "assignee" fields on
previously closed issues for training. Additional logic could be
added in the future to include things like commit messages that close
issues or the text bodies for associated pull requests.
*/

func stringPtr(input string) *string {
	return &input
}

type webhookEvent struct {
	Name    string   `yaml:"name"`
	Actions []string `yaml:"actions"`
}

type settings struct {
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

// Backend implements the backend package interface
var Backend bnkd

type bnkd struct {
	github *github.Client
	help   helper
}

// Configure configures the backend with a client and helper struct
func (b *bnkd) Configure(c *github.Client) {
	b.github = c
	b.help = &help{}
}

func parseRepos(event string, payloadBytes []byte) ([]*github.Repository, error) {
	repos := []*github.Repository{}
	if event == "installation" {
		installation := github.InstallationEvent{}
		if err := json.Unmarshal(payloadBytes, &installation); err != nil {
			return nil, err
		}

		repos = installation.Repositories
	} else if event == "installation_repositories" {
		installation := github.InstallationRepositoriesEvent{}
		if err := json.Unmarshal(payloadBytes, &installation); err != nil {
			return nil, err
		}

		repos = installation.RepositoriesAdded
	}

	return repos, nil
}

// Prepare processes existing issues and establishes indexes for target repos
func (b *bnkd) Prepare(p backend.Payload) error {
	log.Printf("prepare payload bytes: %s\n", string(p.Bytes()))

	repos, err := parseRepos(p.Type(), p.Bytes())
	if err != nil {
		return errors.New("error unmarshalling installation event: " + err.Error())
	}

	log.Printf("repositories: %+v\n", repos)
	for _, repo := range repos {
		log.Printf("repository: %s\n", *repo.FullName)

		config := configObj{}
		if err := yaml.Unmarshal([]byte(p.Config()), &config); err != nil {
			return fmt.Errorf("error parsing heupr config: %s", err.Error())
		}

		fullName := strings.Split(*repo.FullName, "/")
		issues, err := b.help.listIssues(b.github, fullName[0], fullName[1])
		if err != nil {
			return fmt.Errorf("error getting issues: %s", err.Error())
		}
		log.Printf("issues: %+v, count: %d\n", issues, len(issues))

		indexContent := make(map[string]string)
		for _, issue := range issues {
			actor := *issue.Assignee.Login
			if _, ok := indexContent[actor]; ok {
				indexContent[actor] += " " + b.help.getText(issue)
			} else {
				indexContent[actor] = b.help.getText(issue)
			}
		}
		log.Printf("index content: %+v", indexContent)

		bleveClient := newClient()
		for actor, corpus := range indexContent {
			log.Printf("actor: %s, corpus: %s\n", actor, corpus)
			path := "/tmp/" + strings.Replace(*repo.FullName, "/", "_", -1) + ".bleve"
			if err := bleveClient.index(path, actor, corpus); err != nil {
				return fmt.Errorf("error indexing key/value: %s", err.Error())
			}
		}
	}

	log.Println("successful prepare invocation")
	return nil
}

// Act processes new issues and assigns available contributors
func (b *bnkd) Act(p backend.Payload) error {
	log.Printf("act payload bytes: %s\n", string(p.Bytes()))
	if p.Type() != "issues" {
		log.Printf("type %s not supported for issue assignment\n", p.Type())
		return nil
	}

	event := github.IssuesEvent{}
	if err := json.Unmarshal(p.Bytes(), &event); err != nil {
		return fmt.Errorf("error parsing issue: %s", err.Error())
	}

	log.Printf("action: %s\n", *event.Action)
	if *event.Action != "opened" {
		return nil // (?)
	}

	corpus := b.help.getText(event.Issue)
	log.Printf("corpus: %s\n", corpus)

	log.Printf("repository: %s\n", *event.Repo.FullName)
	fullName := strings.Split(*event.Repo.FullName, "/")

	config := configObj{}
	if err := yaml.Unmarshal(p.Config(), &config); err != nil {
		return fmt.Errorf("error parsing heupr config: %s", err.Error())
	}

	contributors := []string{}
	for _, bknd := range config.Backends {
		if bknd.Name == "assignissue" {
			contributors = bknd.Settings.Contributors
		}
	}
	log.Printf("contributors: %v\n", contributors)

	bleveClient := newClient()
	path := "/tmp/" + strings.Replace(*event.Repo.FullName, "/", "_", -1) + ".bleve"
	actor, err := bleveClient.search(path, corpus)
	if err != nil {
		return fmt.Errorf("error searching index: %s", err.Error())
	}
	log.Printf("actor: %s\n", actor)

	for _, contributor := range contributors {
		if contributor == actor {
			if err := b.help.addAssignee(b.github, fullName[0], fullName[1], actor, *event.Issue.Number); err != nil {
				return fmt.Errorf("error adding assignee: %s", err.Error())
			}
		}
	}

	log.Println("successful act invocation")
	return nil
}

func main() {}
