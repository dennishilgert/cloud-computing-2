package app

import (
	"context"
	"fmt"

	"github.com/dennishilgert/cloud-computing-2/internal/app/cache"
	"github.com/dennishilgert/cloud-computing-2/internal/app/http"
	"github.com/dennishilgert/cloud-computing-2/internal/app/translate"
	"github.com/dennishilgert/cloud-computing-2/pkg/concurrency/runner"
	"github.com/dennishilgert/cloud-computing-2/pkg/logger"
)

var log = logger.NewLogger("app")

// App is the main application
type App interface {
	Run(ctx context.Context) error
}

// Options contains the options for `NewApp`.
type Options struct {
	AppPort      int
	GpcProjectId string
	RedisHost    string
	RedisPort    int
}

type app struct {
	httpServer http.Server
}

func NewApp(ctx context.Context, opts Options) (App, error) {
	translator := translate.NewTranslator(ctx, translate.Options{
		ProjectId: opts.GpcProjectId,
	})

	cache := cache.NewCache(cache.Options{
		Host: opts.RedisHost,
		Port: opts.RedisPort,
	})

	return &app{
		httpServer: http.NewHttpServer(translator, cache, http.Options{
			Port: opts.AppPort,
		}),
	}, nil
}

func (a *app) Run(ctx context.Context) error {
	log.Info("app is starting")

	runner := runner.NewRunnerManager(
		func(ctx context.Context) error {
			if err := a.httpServer.Run(ctx); err != nil {
				return fmt.Errorf("failed to run http server: %v", err)
			}
			return nil
		},
		func(ctx context.Context) error {
			if err := a.httpServer.Ready(ctx); err != nil {
				return fmt.Errorf("http server did not become ready in time: %v", err)
			}
			log.Info("http server started")
			<-ctx.Done()
			return nil
		},
	)

	return runner.Run(ctx)
}
