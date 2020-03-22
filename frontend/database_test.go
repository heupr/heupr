package frontend

import (
	"errors"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func TestNewDatabase(t *testing.T) {
	os.Setenv("AWS_REGION", "us-east-1")

	db := NewDatabase()
	if db == nil {
		t.Errorf("description: create database error, received: %+v", db)
	}

	os.Unsetenv("AWS_REGION")
}

type mockDBClient struct {
	queryItemOutput  *dynamodb.QueryOutput
	queryErr         error
	updateItemOutput *dynamodb.UpdateItemOutput
	updateItemErr    error
}

func (m *mockDBClient) Query(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	return m.queryItemOutput, m.queryErr
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
			desc: "successful invocation app installation info",
			config: installConfig{
				WebhookSecret: "secret",
				AppID:         4,
				PEM:           "gar-contingency-command",
			},
			updateItemOutput: nil,
			updateItemErr:    nil,
			err:              "",
		},
		{
			desc: "successful invocation repo installation info",
			config: installConfig{
				WebhookSecret:  "secret",
				InstallationID: 4,
				FullName:       "Contingency Orders for the Grand Army of the Republic: Order Initiation, Orders 1 Through 150",
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
		desc            string
		key             interface{}
		queryItemOutput *dynamodb.QueryOutput
		queryErr        error
		output          installConfig
		err             string
	}{
		{
			desc:            "error getting item",
			key:             "watto",
			queryItemOutput: nil,
			queryErr:        errors.New("query mock error"),
			output:          installConfig{},
			err:             "get item error: query mock error",
		},
		{
			desc: "invalid key",
			key:  "shmi",
			queryItemOutput: &dynamodb.QueryOutput{
				Items: []map[string]*dynamodb.AttributeValue{
					map[string]*dynamodb.AttributeValue{
						"mother": {
							S: aws.String("skywalker"),
						},
					},
				},
			},
			queryErr: nil,
			output:   installConfig{},
			err:      "key not provided: mother",
		},
		{
			desc: "successful string invocation",
			key:  "anakin",
			queryItemOutput: &dynamodb.QueryOutput{
				Items: []map[string]*dynamodb.AttributeValue{
					map[string]*dynamodb.AttributeValue{
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
						"full_name": {
							S: aws.String("tatooine"),
						},
					},
				},
			},
			queryErr: nil,
			output: installConfig{
				AppID:          1,
				PEM:            "tatoo-i-tatoo-ii-ghomrassen-guermessa-chenini",
				WebhookSecret:  "skywalker",
				InstallationID: 2,
			},
			err: "",
		},
		{
			desc: "successful int64 invocation",
			key:  int64(1038),
			queryItemOutput: &dynamodb.QueryOutput{
				Items: []map[string]*dynamodb.AttributeValue{
					map[string]*dynamodb.AttributeValue{
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
						"full_name": {
							S: aws.String("tatooine"),
						},
					},
				},
			},
			queryErr: nil,
			output: installConfig{
				AppID:          1,
				PEM:            "tatoo-i-tatoo-ii-ghomrassen-guermessa-chenini",
				WebhookSecret:  "skywalker",
				InstallationID: 2,
			},
			err: "",
		},
	}

	for _, test := range tests {
		db := db{
			dynamodb: &mockDBClient{
				queryItemOutput: test.queryItemOutput,
				queryErr:        test.queryErr,
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
