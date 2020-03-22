package frontend

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/tidwall/gjson"

	"github.com/heupr/heupr/backend"
)

// APIResponse generates required output for proxy integrations
func APIResponse(code int, msg string) (events.APIGatewayProxyResponse, error) {
	log.Printf("response code: %d, message: %s\n", code, msg)
	resp := events.APIGatewayProxyResponse{
		StatusCode:      code,
		Body:            msg,
		IsBase64Encoded: false,
	}

	if msg == "success" || msg == "query already exists" {
		return resp, nil
	}

	return resp, errors.New(msg)
}

type installConfig struct {
	AppID          int64  `json:"id"`
	FullName       string `json:"full_name"`
	PEM            string `json:"pem"`
	WebhookSecret  string `json:"webhook_secret"`
	InstallationID int64  `json:"installation_id"`
}

var post = func(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	return client.Do(req)
}

// Install configures a new repo installation for the Heupr app
func Install(request events.APIGatewayProxyRequest, db Database) (events.APIGatewayProxyResponse, error) {
	log.Printf("install request: %+v\n", request)

	code := request.QueryStringParameters["code"]
	if code == "" {
		return APIResponse(http.StatusBadRequest, "no code received")
	}
	log.Printf("request code: %s\n", code)

	b := new(bytes.Buffer)
	req, err := http.NewRequest("POST", "https://api.github.com/app-manifests/"+code+"/conversions", b)
	if err != nil {
		return APIResponse(http.StatusInternalServerError, "error creating response: "+err.Error())
	}
	req.Header.Set("Accept", "application/vnd.github.fury-preview+json")

	resp, err := post(req)
	if err != nil {
		return APIResponse(http.StatusInternalServerError, "error converting code: "+err.Error())
	}
	defer resp.Body.Close()

	log.Printf("manifest response: %+v\n", resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return APIResponse(http.StatusInternalServerError, "error reading conversion body")
	}
	log.Printf("request body: %s\n", body)

	config := installConfig{}
	if err := json.Unmarshal(body, &config); err != nil {
		return APIResponse(http.StatusInternalServerError, "error parsing conversion body")
	}
	log.Printf("installation config: %+v\n", config)

	if err := db.Put(config); err != nil {
		return APIResponse(http.StatusInternalServerError, "error putting app config: "+err.Error())
	}

	log.Println("successful install handler invocation")

	return events.APIGatewayProxyResponse{
		StatusCode: 302,
		Headers: map[string]string{
			"Location": "https://heupr.github.io/success",
		},
		Body:            "success",
		IsBase64Encoded: false,
	}, nil
}

type webhookEvent struct {
	Name    string   `yaml:"name"`
	Actions []string `yaml:"actions"`
}

type backendObj struct {
	Name     string         `yaml:"name"`
	Events   []webhookEvent `yaml:"events"`
	Location string         `yaml:"location"`
}

type configObj struct {
	Backends []backendObj `yaml:"backends"`
}

type payload struct {
	B []byte
	T string
	C []byte
}

func (p *payload) Bytes() []byte {
	return p.B
}

func (p *payload) Type() string {
	return p.T
}

func (p *payload) Config() []byte {
	return p.C
}

// Event processes webhook events received by Heupr app repo installations
func Event(request events.APIGatewayProxyRequest, db Database, bknds []backend.Backend) (events.APIGatewayProxyResponse, error) {
	log.Printf("event request: %+v\n", request)

	eventType := request.Headers["X-GitHub-Event"]
	signature := request.Headers["X-Hub-Signature"]
	log.Printf("event type: %s, signature: %s\n", eventType, signature)

	body := []byte(request.Body)

	switch eventType {
	case "installation", "integration_installation", "installation_repositories", "integration_installation_repositories": // NOTE: Last two kept for GitHub inconsistency
		appID := gjson.Get(request.Body, "installation.app_id").Int()
		installationID := gjson.Get(request.Body, "installation.id").Int()

		log.Printf("app id: %d, installation id: %d\n", appID, installationID)

		installConfig, err := db.Get(appID)
		if err != nil {
			return APIResponse(http.StatusInternalServerError, "error getting config: "+err.Error())
		}

		if err := validateEvent(installConfig.WebhookSecret, signature, body); err != nil {
			return APIResponse(http.StatusInternalServerError, "error validating event: "+err.Error())
		}

		repos := gjson.Result{}
		if !strings.Contains(eventType, "repositories") {
			repos = gjson.Get(request.Body, "repositories.#.full_name")
		} else {
			repos = gjson.Get(request.Body, "repositories_added.#.full_name")
		}

		log.Printf("repositories: %v\n", repos)

		for _, repo := range repos.Array() {
			fullName := repo.String()
			log.Printf("repository: %s\n", fullName)

			installConfig.FullName = fullName
			installConfig.InstallationID = installationID

			client, err := newClient(installConfig.AppID, installConfig.InstallationID, installConfig.PEM)
			if err != nil {
				return APIResponse(http.StatusInternalServerError, "error creating client: "+err.Error())
			}

			if err := db.Put(installConfig); err != nil {
				return APIResponse(http.StatusInternalServerError, "error putting app config: "+err.Error())
			}

			fullNameSplit := strings.Split(fullName, "/")
			file, err := getContent(client, fullNameSplit[0], fullNameSplit[1], ".heupr.yml")
			if err != nil {
				return APIResponse(http.StatusInternalServerError, "error getting repo config file: "+err.Error())
			}
			log.Printf("file content: %s\n", file)

			backendPayload := &payload{
				T: eventType,
				B: body,
				C: []byte(file),
			}

			for _, bknd := range bknds {
				bknd.Configure(client)
				if err := bknd.Prepare(backendPayload); err != nil {
					return APIResponse(http.StatusInternalServerError, "error calling backend prepare: "+err.Error())
				}
			}
		}

	case "issues", "pull_request", "project", "project_card", "project_column":
		fullName := gjson.Get(request.Body, "repository.full_name").String()

		installConfig, err := db.Get(fullName)
		if err != nil {
			return APIResponse(http.StatusInternalServerError, "error getting config: "+err.Error())
		}
		log.Printf("installation config: %+v\n", installConfig)

		if err := validateEvent(installConfig.WebhookSecret, signature, body); err != nil {
			return APIResponse(http.StatusInternalServerError, "error validating event: "+err.Error())
		}

		client, err := newClient(installConfig.AppID, installConfig.InstallationID, installConfig.PEM)
		if err != nil {
			return APIResponse(http.StatusInternalServerError, "error creating client: "+err.Error())
		}

		fullNameSplit := strings.Split(fullName, "/")
		file, err := getContent(client, fullNameSplit[0], fullNameSplit[1], ".heupr.yml")
		if err != nil {
			return APIResponse(http.StatusInternalServerError, "error getting repo config file: "+err.Error())
		}
		log.Printf("file content: %s\n", file)

		backendPayload := &payload{
			T: eventType,
			B: body,
			C: []byte(file),
		}

		for _, bknd := range bknds {
			bknd.Configure(client)
			if err := bknd.Act(backendPayload); err != nil {
				return APIResponse(http.StatusInternalServerError, "error calling backend act: "+err.Error())
			}
		}

	default:
		message := fmt.Sprintf("event type %s not supported", eventType)
		log.Println(message)
		return APIResponse(http.StatusInternalServerError, message)
	}

	log.Println("successful event handler invocation")
	return APIResponse(http.StatusOK, "success")
}
