package main

import (
	"errors"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/heupr/heupr/frontend"
)

// HANDLER allows for build-time starter configuration
var HANDLER string

func starter(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	db := frontend.NewDatabase()

	switch HANDLER {
	case "INSTALL":
		return frontend.Install(request, db)
	case "EVENT":
		return frontend.Event(request, db)
	}

	return events.APIGatewayProxyResponse{
		StatusCode:      500,
		Body:            "requested lambda type not available",
		IsBase64Encoded: false,
	}, errors.New("requested lambda type not available")
}

func main() {
	lambda.Start(starter)
}
