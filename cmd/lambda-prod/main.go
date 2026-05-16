package main

import (
	awslambda "github.com/pdkovacs/xcaliapp/internal/awslambda"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(awslambda.HandleRequest)
}
