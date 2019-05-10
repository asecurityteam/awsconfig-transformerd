package main

import (
	"context"
	"os"

	handlers "github.com/asecurityteam/awsconfig-transformerd/pkg/handlers/v1"
	"github.com/asecurityteam/runhttp"
	"github.com/asecurityteam/serverfull"
	"github.com/asecurityteam/settings"
)

func main() {
	ctx := context.Background()

	transformer := &handlers.Transformer{
		LogFn:  runhttp.LoggerFromContext,
		StatFn: runhttp.StatFromContext,
	}

	handlersMap := map[string]serverfull.Function{
		"awsConfigHandler": serverfull.NewFunction(transformer.Handle),
	}

	source, err := settings.NewEnvSource(os.Environ())
	if err != nil {
		panic(err.Error())
	}
	fetcher := &serverfull.StaticFetcher{Functions: handlersMap}
	if err := serverfull.Start(ctx, source, fetcher); err != nil {
		panic(err.Error())
	}
}
