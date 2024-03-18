package app

import (
	"github.com/dennishilgert/cloud-computing-2/cmd/config"
	"github.com/dennishilgert/cloud-computing-2/internal/app"
	"github.com/dennishilgert/cloud-computing-2/pkg/concurrency/runner"
	"github.com/dennishilgert/cloud-computing-2/pkg/logger"
	"github.com/dennishilgert/cloud-computing-2/pkg/signals"
	"github.com/joho/godotenv"
)

var log = logger.NewLogger("app")

func Run() {
	// load environment variables on local installation
	godotenv.Load()

	// load configuration from environment variables
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	err = logger.ApplyOptionsToLoggers(&cfg.Logger)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("starting translator -- version %s", "1.0.0")
	log.Infof("log level set to: %s", cfg.Logger.OutputLevel)

	ctx := signals.Context()
	app, err := app.NewApp(ctx, app.Options{
		AppPort:      cfg.AppPort,
		GpcProjectId: cfg.GpcProjectId,
		RedisHost:    cfg.RedisHost,
		RedisPort:    cfg.RedisPort,
	})
	if err != nil {
		log.Fatalf("error while creating translator: %v", err)
	}

	err = runner.NewRunnerManager(
		app.Run,
	).Run(ctx)
	if err != nil {
		log.Fatalf("error while running translator: %v", err)
	}

	log.Info("translator shut down gracefully")
}
