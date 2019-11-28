#!/bin/bash

go build -buildmode=plugin -o estimatepr.so backend/estimatepr/estimatepr.go
aws s3 mv estimatepr.so s3://heupr/estimatepr.so

go build -buildmode=plugin -o assignissue.so backend/assignissue/assignissue.go
aws s3 mv assignissue.so s3://heupr/assignissue.so

GOARCH=amd64 GOOS=linux go build -ldflags "-X main.HANDLER=INSTALL" -o install
zip heupr-install.zip install
aws lambda update-function-code --function-name heupr-install --zip-file fileb://heupr-install.zip --region us-east-1

GOARCH=amd64 GOOS=linux go build -ldflags "-X main.HANDLER=EVENT" -o event
zip heupr-event.zip event
aws lambda update-function-code --function-name heupr-event --zip-file fileb://heupr-event.zip --region us-east-1
