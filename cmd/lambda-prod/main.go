package main

import (
	awslambda "github.com/pdkovacs/xcaliapp/internal/awslambda"
	"github.com/pdkovacs/xcaliapp/internal/logging"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	logging.Init()
	lambda.Start(awslambda.HandleRequest)
}
