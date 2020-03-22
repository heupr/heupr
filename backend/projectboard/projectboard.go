package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/google/go-github/v28/github"
	"github.com/tidwall/gjson"

	"github.com/heupr/heupr/backend"
)

/*
Description:
`projectboard` plugin provides update messages for Project Board events.

Setup:
In the `.heupr.yml` file, include a backend option:

```
backends:
- name: projectboard
  settings:
    urls:
      - https://example-target-url.com/
```

- The `urls` array should be URLs that you wish to send updates to.

Notes:
DO NOT include sensative URLs in this config (e.g. Slack URLs).
*/

type helper interface {
	postHTTP(url string, body io.Reader) error
	parseMessage(eventType, payloadString string) string
}

type help struct {
	client *github.Client
}

func (h *help) postHTTP(url string, body io.Reader) error {
	resp, err := http.Post(url, "application/json", body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("non-200 status code: %d", resp.StatusCode)
	}

	return nil
}

func (h *help) parseMessage(eventType, payloadString string) string {
	message := ""

	user := gjson.Get(payloadString, "sender.login").String()
	action := gjson.Get(payloadString, "action").String()

	switch eventType {
	case "project":
		projectName := gjson.Get(payloadString, "project.name").String()
		message = fmt.Sprintf("user %s %s project %s", user, action, projectName)
	case "project_column":
		projectColumnName := gjson.Get(payloadString, "project_column.name")
		message = fmt.Sprintf("user %s %s project column %s", user, action, projectColumnName)
	case "project_card":
		projectCardID := gjson.Get(payloadString, "project_card.id").Int()
		log.Printf("project card ID: %d", projectCardID)

		card, _, err := h.client.Projects.GetProjectCard(context.Background(), projectCardID)
		if err != nil {
			return message
		}
		log.Printf("card: %+v\n", card)

		if action == "moved" && (card.PreviousColumnName == nil || card.ColumnName == nil) {
			message = fmt.Sprintf("user %s %s card", user, action)
		} else if action == "moved" && (card.PreviousColumnName == nil || card.ColumnName == nil) {
			message = fmt.Sprintf("user %s %s %d from %s to %s", user, action, projectCardID, *card.PreviousColumnName, *card.ColumnName)
		} else {
			message = fmt.Sprintf("user %s %s project card %d", user, action, projectCardID)
		}
	default:
		log.Printf("type %s not supported for project board\n", eventType)
		return message
	}

	log.Printf("message: %s\n", message)

	return message
}

// Backend implements the backend package interface
var Backend bnkd

type bnkd struct {
	help helper
}

// Configure configures the backend with a client
func (b *bnkd) Configure(c *github.Client) {
	log.Println("configure project board backend")
	b.help = &help{
		client: c,
	}
}

// Prepare processes performs no action but implements the Backend interface
func (b *bnkd) Prepare(p backend.Payload) error {
	log.Printf("prepare payload bytes: %s\n", string(p.Bytes()))

	return nil
}

type settings struct {
	URLs []string `yaml:"urls"`
}

type backendObj struct {
	Name     string   `yaml:"name"`
	Settings settings `yaml:"settings"`
}

type configObj struct {
	Backends []backendObj `yaml:"backends"`
}

// Act processes Project Board actions and posts messages to the configured URL
func (b *bnkd) Act(p backend.Payload) error {
	log.Printf("act payload bytes: %s\n", string(p.Bytes()))

	payloadString := string(p.Bytes())
	message := b.help.parseMessage(p.Type(), payloadString)
	if message == "" {
		return errors.New("no output message")
	}

	output := fmt.Sprintf(`{"message": %s}`, message)

	config := configObj{}
	if err := yaml.Unmarshal(p.Config(), &config); err != nil {
		return fmt.Errorf("error parsing heupr config: %s", err.Error())
	}

	urls := []string{}
	for _, bknd := range config.Backends {
		if bknd.Name == "projectboard" {
			urls = bknd.Settings.URLs
		}
	}

	for _, url := range urls {
		if err := b.help.postHTTP(url, strings.NewReader(output)); err != nil {
			return fmt.Errorf("error posting target url: %s", err.Error())
		}
	}

	return nil
}

func main() {}
