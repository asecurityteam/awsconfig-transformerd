package main

import (
	"context"
	"os"

	handlers "bitbucket.org/asecurityteam/awsconfig-transformerd/pkg/handlers/v1"
	serverfull "github.com/asecurityteam/serverfull/pkg"
	serverfulldomain "github.com/asecurityteam/serverfull/pkg/domain"
	"github.com/asecurityteam/settings"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	ctx := context.Background()

	handlersMap := map[string]serverfulldomain.Handler{
		"awsConfigHandler": lambda.NewHandler(handlers.Handle)}

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
