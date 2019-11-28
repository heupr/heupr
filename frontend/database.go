package frontend

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type dynamoDBClient interface {
	GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
	UpdateItem(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error)
}

// Database provides an interface to DynamoDB data storage
type Database interface {
	Put(input installConfig) error
	Get(key string) (installConfig, error)
}

// NewDatabase creates a new instance of a Database implementation
func NewDatabase() Database {
	return &db{
		dynamodb: dynamodb.New(session.New()),
	}
}

type db struct {
	dynamodb dynamoDBClient
}

func (d *db) Put(input installConfig) error {
	updateInput := dynamodb.UpdateItemInput{
		TableName: aws.String("heupr"),
		Key: map[string]*dynamodb.AttributeValue{
			"webhook_secret": {
				S: aws.String(input.WebhookSecret),
			},
		},
	}

	if input.InstallationID == 0 {
		updateInput.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			"installation_id": {
				N: aws.String(string(input.InstallationID)),
			},
			"repo_owner": {
				S: aws.String(input.RepoOwner),
			},
			"repo_name": {
				S: aws.String(input.RepoName),
			},
		}
	} else {
		updateInput.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			"app_id": {
				N: aws.String(string(input.AppID)),
			},
			"pem": {
				S: aws.String(input.PEM),
			},
		}
	}

	_, err := d.dynamodb.UpdateItem(&updateInput)
	if err != nil {
		return fmt.Errorf("put item error: %s", err.Error())
	}
	return nil
}

func (d *db) Get(key string) (installConfig, error) {
	getInput := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"webhook_secret": {
				S: aws.String(key),
			},
		},
		TableName: aws.String("heupr"),
	}

	output := installConfig{}
	result, err := d.dynamodb.GetItem(getInput)
	if err != nil {
		return output, fmt.Errorf("get item error: %s", err.Error())
	}

	for key, item := range result.Item {
		switch key {
		case "app_id":
			value, err := strconv.ParseInt(*item.N, 10, 64)
			if err != nil {
				return output, fmt.Errorf("convert item int: %s", err.Error())
			}
			output.AppID = value
		case "pem":
			output.PEM = *item.S
		case "webhook_secret":
			output.WebhookSecret = *item.S
		case "installation_id":
			value, err := strconv.ParseInt(*item.N, 10, 64)
			if err != nil {
				return output, fmt.Errorf("convert item int: %s", err.Error())
			}
			output.InstallationID = value
		case "repo_owner":
			output.RepoOwner = *item.S
		case "repo_name":
			output.RepoName = *item.S
		default:
			return output, fmt.Errorf("key not provided: %s", key)
		}

	}

	return output, nil
}
