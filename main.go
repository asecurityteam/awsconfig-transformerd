package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

	handlers "bitbucket.org/asecurityteam/awsconfig-transformerd/pkg/handlers/v1"
	serverfull "github.com/asecurityteam/serverfull/pkg"
	serverfulldomain "github.com/asecurityteam/serverfull/pkg/domain"
	"github.com/asecurityteam/settings"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {

	streamApplianceEndpoint := mustEnv("STREAM_APPLIANCE_ENDPOINT")
	streamApplianceURL, err := url.Parse(streamApplianceEndpoint)
	if err != nil {
		panic(err.Error())
	}

	ctx := context.Background()

	handler := handlers.AWSConfigChangeEventHandler{
		Queuer: handlers.NewEventQueuer(streamApplianceURL)}

	handlersMap := map[string]serverfulldomain.Handler{
		"awsConfigHandler": lambda.NewHandler(handler.Handle(streamApplianceURL))}

	source, err := settings.NewEnvSource(os.Environ())
	if err != nil {
		panic(err.Error())
	}
	rt, err := serverfull.NewStatic(ctx, source, handlersMap)
	if err != nil {
		panic(err.Error())
	}
	if err := rt.Run(); err != nil {
		panic(err.Error())
	}
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("%s is required", key))
	}
	return val
}
