package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"plugin"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/heupr/heupr/backend"
	"github.com/heupr/heupr/frontend"
)

// HANDLER allows for build-time starter configuration
var HANDLER string

func starter(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	db := frontend.NewDatabase()

	files, err := ioutil.ReadDir("/opt/")
	if err != nil {
		return frontend.APIResponse(http.StatusInternalServerError, "error reading plugin files: "+err.Error())
	}

	bknds := []backend.Backend{} // NOTE: Possibly adjust to be a map[string]backend.Backend for index referencing
	for _, file := range files {
		if !strings.Contains(file.Name(), ".so") {
			continue
		}

		plug, err := plugin.Open("/opt/" + file.Name())
		if err != nil {
			return frontend.APIResponse(http.StatusInternalServerError, "error opening plugin file: "+err.Error())
		}

		symBackend, err := plug.Lookup("Backend")
		if err != nil {
			return frontend.APIResponse(http.StatusInternalServerError, "error looking up backend plugin: "+err.Error())
		}

		var bknd backend.Backend
		bknd, ok := symBackend.(backend.Backend)
		if !ok {
			return frontend.APIResponse(http.StatusInternalServerError, "error asserting backend plugin type: "+err.Error())
		}

		bknds = append(bknds, bknd)
	}

	switch HANDLER {
	case "INSTALL":
		return frontend.Install(request, db)
	case "EVENT":
		return frontend.Event(request, db, bknds)
	}

	return frontend.APIResponse(http.StatusInternalServerError, "requested lambda type not available")
}

func main() {
	lambda.Start(starter)
}
