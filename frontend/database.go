package frontend

import (
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type dynamoDBClient interface {
	Query(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error)
	UpdateItem(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error)
}

// Database provides an interface to DynamoDB data storage
type Database interface {
	Put(input installConfig) error
	Get(key interface{}) (installConfig, error)
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
	log.Printf("put input: %+v\n", input)
	updateInput := dynamodb.UpdateItemInput{
		TableName: aws.String("heupr"),
		Key: map[string]*dynamodb.AttributeValue{
			"app_id": {
				N: aws.String(strconv.FormatInt(input.AppID, 10)),
			},
		},
		ReturnValues: aws.String("ALL_NEW"),
	}

	if input.FullName == "" {
		log.Println("installation")
		updateInput.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":webhook_secret": {
				S: aws.String(input.WebhookSecret),
			},
			":pem": {
				S: aws.String(input.PEM),
			},
		}
		updateInput.UpdateExpression = aws.String("set webhook_secret = :webhook_secret, pem = :pem")
	} else {
		log.Println("event")
		updateInput.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":full_name": {
				S: aws.String(input.FullName),
			},
			":installation_id": {
				N: aws.String(strconv.FormatInt(input.InstallationID, 10)),
			},
		}
		updateInput.UpdateExpression = aws.String("set full_name = :full_name, installation_id = :installation_id")
	}

	log.Printf("update input: %s\n", updateInput)

	_, err := d.dynamodb.UpdateItem(&updateInput)
	if err != nil {
		return fmt.Errorf("put item error: %s", err.Error())
	}

	log.Println("successful put method invocation")
	return nil
}

func (d *db) Get(key interface{}) (installConfig, error) {
	keyType := fmt.Sprintf("%T", key) // NOTE: Better/more performent alternative should be implemented
	log.Printf("get input: %v, type: %s\n", key, keyType)

	queryInput := &dynamodb.QueryInput{
		TableName: aws.String("heupr"),
	}

	if keyType == "int64" {
		log.Println("apps")
		queryInput.KeyConditionExpression = aws.String("app_id = :app_id")
		queryInput.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":app_id": {
				N: aws.String(strconv.FormatInt(key.(int64), 10)),
			},
		}
	} else if keyType == "string" {
		log.Println("repos")
		queryInput.KeyConditionExpression = aws.String("full_name = :full_name")
		queryInput.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
			":full_name": {
				S: aws.String(key.(string)),
			},
		}
		queryInput.IndexName = aws.String("repos")
	}

	log.Printf("query input: %+v\n", queryInput)

	output := installConfig{}
	result, err := d.dynamodb.Query(queryInput)
	if err != nil {
		return output, fmt.Errorf("get item error: %s", err.Error())
	}

	for key, item := range result.Items[0] {
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
		case "full_name":
			output.FullName = *item.S
		default:
			return output, fmt.Errorf("key not provided: %s", key)
		}
	}

	log.Println("successful get method invocation")
	return output, nil
}
