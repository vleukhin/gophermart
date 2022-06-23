package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/vleukhin/gophermart/internal"
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

	app, err := internal.NewApplication(cfg)
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
