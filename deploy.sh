#!/bin/bash

# build backend plugin files
mv backend/estimatepr/estimatepr.go .
go build -buildmode=plugin -o estimatepr.so estimatepr.go
rm estimatepr.go

mv backend/assignissue/assignissue.go backend/assignissue/helper.go backend/assignissue/index.go .
go build -buildmode=plugin -o assignissue.so assignissue.go helper.go index.go
rm assignissue.go helper.go index.go

mv backend/projectboard/projectboard.go .
go build -buildmode=plugin -o projectboard.so projectboard.go
rm projectboard.go

zip heupr-plugins.zip estimatepr.so assignissue.so projectboard.so
aws s3 mv heupr-plugins.zip s3://heupr/

# build app lambda functions
GOARCH=amd64 GOOS=linux go build -ldflags "-X main.HANDLER=INSTALL" -o install
zip heupr-install.zip install

GOARCH=amd64 GOOS=linux go build -ldflags "-X main.HANDLER=EVENT" -o event
zip heupr-event.zip event

aws s3 mv heupr-install.zip s3://heupr/
aws s3 mv heupr-event.zip s3://heupr/

# deploy cloudformation template resources
aws cloudformation deploy --template-file cft.yml --stack-name heupr --parameter-overrides HeuprBucket=heupr --capabilities CAPABILITY_NAMED_IAM  --region us-east-1 --no-fail-on-empty-changeset

# update lambda code
aws lambda update-function-code --function-name heupr-install --s3-bucket heupr --s3-key heupr-install.zip --region us-east-1
aws lambda update-function-code --function-name heupr-event --s3-bucket heupr --s3-key heupr-event.zip --region us-east-1

# publish layer/retrieve arn
ARN=$(aws lambda publish-layer-version --layer-name HeuprEventLayer --content S3Bucket=heupr,S3Key=heupr-plugins.zip --compatible-runtimes go1.x --region us-east-1 | jq -r '.LayerVersionArn')

aws lambda update-function-configuration --function-name heupr-event --layers $ARN --region us-east-1
