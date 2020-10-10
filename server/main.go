package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	chiadapter "github.com/awslabs/aws-lambda-go-api-proxy/chi"
)

var chiLambda *chiadapter.ChiLambda
var awsSession = session.Must(session.NewSession(&aws.Config{}))

func init() {
	r := newRouter()
	chiLambda = chiadapter.New(r)
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return chiLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(handler)
}
