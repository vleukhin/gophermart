package main

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vleukhin/gophermart/internal"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := &internal.AppConfig{}

	if err := cfg.Parse(); err != nil {
		log.Fatal().Msg(err.Error())
	}

	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal().Msg(err.Error())
		os.Exit(1)
	}

	zerolog.SetGlobalLevel(logLevel)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := internal.NewApplication(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create application")
		os.Exit(1)
	}

	errChan := make(chan error)
	sigChan := make(chan os.Signal, 1)

	go app.Run(errChan)
	defer func(app *internal.Application) {
		err := app.ShutDown()
		if err != nil {
			panic(err)
		}
	}(app)

	signal.Ignore(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-sigChan:
		log.Info().Msg("Terminating...")
		os.Exit(0)
	case err := <-errChan:
		log.Error().Msg("Application error: " + err.Error())
		os.Exit(1)
	}
}
