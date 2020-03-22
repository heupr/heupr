package frontend

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v28/github"
)

var newClient = func(appID, installationID int64, file string) (*github.Client, error) {
	tr, err := ghinstallation.New(http.DefaultTransport, appID, installationID, []byte(file))
	if err != nil {
		return nil, err
	}

	return github.NewClient(&http.Client{
		Transport: tr,
	}), nil
}

var getContent = func(c *github.Client, owner, repo, path string) (string, error) {
	opts := &github.RepositoryContentGetOptions{}
	file, _, _, err := c.Repositories.GetContents(context.Background(), owner, repo, path, opts)
	if err != nil {
		return "", errors.New("error getting content object: " + err.Error())
	}

	content, err := file.GetContent()
	if err != nil {
		return "", errors.New("error getting content string: " + err.Error())
	}

	return content, nil
}

var validateEvent = func(secret, signature string, body []byte) error {
	log.Printf("validate event body: %s\n", body)
	if err := github.ValidateSignature(signature, body, []byte(secret)); err != nil {
		return err
	}

	return nil
}
