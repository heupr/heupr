package frontend

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type mockDBClient struct {
	getItemOutput    *dynamodb.GetItemOutput
	getItemErr       error
	updateItemOutput *dynamodb.UpdateItemOutput
	updateItemErr    error
}

func (m *mockDBClient) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	return m.getItemOutput, m.getItemErr
}

func (m *mockDBClient) UpdateItem(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	return m.updateItemOutput, m.updateItemErr
}

func TestPut(t *testing.T) {
	tests := []struct {
		desc             string
		config           installConfig
		updateItemOutput *dynamodb.UpdateItemOutput
		updateItemErr    error
		err              string
	}{
		{
			desc: "error updating item",
			config: installConfig{
				WebhookSecret: "secret",
				AppID:         66,
				PEM:           "execute all of the jedi",
			},
			updateItemOutput: nil,
			updateItemErr:    errors.New("mock update error"),
			err:              "put item error: mock update error",
		},
		{
			desc: "successful invocation repo installation info",
			config: installConfig{
				WebhookSecret:  "secret",
				InstallationID: 4,
				RepoOwner:      "gar",
				RepoName:       "orders",
			},
			updateItemOutput: nil,
			updateItemErr:    nil,
			err:              "",
		},
		{
			desc: "successful invocation app installation info",
			config: installConfig{
				WebhookSecret: "secret",
				AppID:         4,
				PEM:           "gar contingency command",
			},
			updateItemOutput: nil,
			updateItemErr:    nil,
			err:              "",
		},
	}

	for _, test := range tests {
		db := db{
			dynamodb: &mockDBClient{
				updateItemOutput: test.updateItemOutput,
				updateItemErr:    test.updateItemErr,
			},
		}

		err := db.Put(test.config)

		if err != nil && err.Error() != test.err {
			t.Errorf("description: %s, error received: %s, expected: %s", test.desc, err.Error(), test.err)
		}
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		desc          string
		key           string
		getItemOutput *dynamodb.GetItemOutput
		getItemErr    error
		output        installConfig
		err           string
	}{
		{
			desc:          "error getting item",
			key:           "key",
			getItemOutput: nil,
			getItemErr:    errors.New("get mock error"),
			output:        installConfig{},
			err:           "get item error: get mock error",
		},
		{
			desc: "successful invocation",
			key:  "test-key",
			getItemOutput: &dynamodb.GetItemOutput{
				Item: map[string]*dynamodb.AttributeValue{
					"app_id": {
						N: aws.String("1"),
					},
					"pem": {
						S: aws.String("tatoo-i-tatoo-ii-ghomrassen-guermessa-chenini"),
					},
					"webhook_secret": {
						S: aws.String("skywalker"),
					},
					"installation_id": {
						N: aws.String("2"),
					},
					"repo_owner": {
						S: aws.String("outer-rim"),
					},
					"repo_name": {
						S: aws.String("tatooine"),
					},
				},
			},
			getItemErr: nil,
			output: installConfig{
				AppID:          1,
				PEM:            "tatoo-i-tatoo-ii-ghomrassen-guermessa-chenini",
				WebhookSecret:  "skywalker",
				InstallationID: 2,
				RepoOwner:      "outer-rim",
				RepoName:       "tatooine",
			},
			err: "",
		},
	}

	for _, test := range tests {
		db := db{
			dynamodb: &mockDBClient{
				getItemOutput: test.getItemOutput,
				getItemErr:    test.getItemErr,
			},
		}

		output, err := db.Get(test.key)

		if err != nil && err.Error() != test.err {
			t.Errorf("description: %s, error received: %s, expected: %s", test.desc, err.Error(), test.err)
		}

		if output.PEM != test.output.PEM {
			t.Errorf("description: %s, output received: %+v, expected: %+v", test.desc, output, test.output)
		}
	}
}
