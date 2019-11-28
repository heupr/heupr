package frontend

import (
	"context"
	"errors"
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
		return "", errors.New("error getting content: " + err.Error())
	}

	return *file.Content, nil
}
