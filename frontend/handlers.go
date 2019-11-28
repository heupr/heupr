package frontend

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"plugin"

	"github.com/aws/aws-lambda-go/events"
	"github.com/go-yaml/yaml"
	"github.com/google/go-github/v28/github"

	"github.com/heupr/heupr/backend"
)

func createResponse(code int, msg string) (events.APIGatewayProxyResponse, error) {
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
	PEM            string `json:"pem"`
	WebhookSecret  string `json:"webhook_secret"`
	InstallationID int64  `json:"installation_id"`
	RepoOwner      string `json:"repo_owner"`
	RepoName       string `json:"repo_name"`
}

var post = http.Post // for test mocks

// Install configures a new repo installation for the Heupr app
func Install(request events.APIGatewayProxyRequest, db Database) (events.APIGatewayProxyResponse, error) {
	log.Printf("install request: %+v", request)

	code := request.Headers["code"]
	if code == "" {
		return createResponse(http.StatusBadRequest, "no code received")
	}

	b := new(bytes.Buffer)
	resp, err := post("https://github.com/app-manifests/"+code+"/conversions", "application/json", b)
	if err != nil {
		return createResponse(http.StatusInternalServerError, "error converting code: "+err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return createResponse(http.StatusInternalServerError, "error reading conversion body")
	}

	installConfig := installConfig{}
	if err := json.Unmarshal(body, &installConfig); err != nil {
		return createResponse(http.StatusInternalServerError, "error parsing conversion body")
	}

	if err := db.Put(installConfig); err != nil {
		return createResponse(http.StatusInternalServerError, "error putting app config: "+err.Error())
	}

	return createResponse(http.StatusOK, "success")
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
}

func (p *payload) Bytes() []byte {
	return p.B
}

func (p *payload) Type() string {
	return p.T
}

func processBackends(client *github.Client, payloadInput backend.Payload, backends []backendObj, method string) error {
	for _, b := range backends {
		soFile := b.Name + ".so"

		if _, err := os.Stat(soFile); err != nil {
			resp, err := http.Get(b.Location)
			if err != nil {
				return fmt.Errorf("error fetching external %s file: %s", soFile, err.Error())
			}
			defer resp.Body.Close()

			out, err := os.Create(soFile)
			if err != nil {
				return fmt.Errorf("error creating external %s file: %s", soFile, err.Error())
			}
			defer out.Close()

			_, err = io.Copy(out, resp.Body)
			if err != nil {
				return fmt.Errorf("error copying %s file: %s", soFile, err.Error())
			}
			defer out.Close()
		}

		plug, err := plugin.Open(soFile)
		if err != nil {
			return fmt.Errorf("error opening %s file: %s", soFile, err.Error())
		}

		symBackend, err := plug.Lookup("Backend")
		if err != nil {
			return fmt.Errorf("error looking up backend plugin: %s", err.Error())
		}

		var bknd backend.Backend
		bknd, ok := symBackend.(backend.Backend)
		if !ok {
			return errors.New("error type asserting backend plugin")
		}

		bknd.Configure(client)

		if method == "prepare" {
			if err := bknd.Prepare(payloadInput); err != nil {
				return fmt.Errorf("error calling backend prepare method: %s", err.Error())
			}
		} else if method == "act" {
			if err := bknd.Act(payloadInput); err != nil {
				return fmt.Errorf("error calling backend prepare method: %s", err.Error())
			}
		}
	}

	return nil
}

// Event processes webhook events received by Heupr app repo installations
func Event(request events.APIGatewayProxyRequest, db Database) (events.APIGatewayProxyResponse, error) {
	signature := request.Headers["X-Hub-Signature"]
	installConfig, err := db.Get(signature)
	if err != nil {
		return createResponse(http.StatusInternalServerError, "error getting config: "+err.Error())
	}

	githubPayload := []byte(request.Body)
	if err := github.ValidateSignature(signature, githubPayload, []byte(installConfig.WebhookSecret)); err != nil {
		return createResponse(http.StatusInternalServerError, "error validating payload: "+err.Error())
	}

	eventType := request.Headers["X-Github-Event"]
	event, err := github.ParseWebHook(eventType, githubPayload)
	if err != nil {
		return createResponse(http.StatusInternalServerError, "error parsing webhook event: "+err.Error())
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return createResponse(http.StatusInternalServerError, "error marshalling event: "+err.Error())
	}

	payloadInput := &payload{
		B: eventBytes,
		T: eventType,
	}

	switch v := event.(type) {
	case *github.InstallationRepositoriesEvent:
		if *v.Action != "added" {
			return createResponse(http.StatusInternalServerError, fmt.Sprintf("repo event not \"added\", received event: \"%s\"", *v.Action))
		}

		client, err := newClient(*v.Installation.ID, *v.Installation.AppID, installConfig.PEM)
		if err != nil {
			return createResponse(http.StatusInternalServerError, "error creating client: "+err.Error())
		}

		for _, repo := range v.RepositoriesAdded { // NOTE: Refactor into helper function
			owner := *repo.Owner.Login
			repo := *repo.Name

			installConfig.InstallationID = *v.Installation.ID
			installConfig.RepoOwner = owner
			installConfig.RepoName = repo

			if err := db.Put(installConfig); err != nil {
				return createResponse(http.StatusInternalServerError, "error putting app config: "+err.Error())
			}

			file, err := getContent(client, owner, repo, ".heupr.yml")
			if err != nil {
				return createResponse(http.StatusInternalServerError, "error getting repo config file: "+err.Error())
			}

			repoConfig := configObj{}
			if err := yaml.Unmarshal([]byte(file), &repoConfig); err != nil {
				return createResponse(http.StatusInternalServerError, "error parsing repo config file: "+err.Error())
			}

			if err := processBackends(client, payloadInput, repoConfig.Backends, "prepare"); err != nil {
				return createResponse(http.StatusInternalServerError, "error processing event: "+err.Error())
			}
		}

	case *github.IssuesEvent, *github.PullRequestEvent:
		client, err := newClient(installConfig.InstallationID, installConfig.AppID, installConfig.PEM)
		if err != nil {
			return createResponse(http.StatusInternalServerError, "error creating client: "+err.Error())
		}

		file, err := getContent(client, installConfig.RepoOwner, installConfig.RepoName, ".heupr.yml")
		if err != nil {
			return createResponse(http.StatusInternalServerError, "error getting repo config file: "+err.Error())
		}

		repoConfig := configObj{}
		if err := yaml.Unmarshal([]byte(file), &repoConfig); err != nil {
			return createResponse(http.StatusInternalServerError, "error parsing repo config file: "+err.Error())
		}

		if err := processBackends(client, payloadInput, repoConfig.Backends, "act"); err != nil {
			return createResponse(http.StatusInternalServerError, "error processing event: "+err.Error())
		}

	default:
		return createResponse(http.StatusInternalServerError, "event type "+eventType+" not supported")
	}

	return createResponse(http.StatusOK, "success")
}
