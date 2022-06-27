package internal

import (
	"context"
	"github.com/caarlos0/env"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/vleukhin/gophermart/internal/services"
	"github.com/vleukhin/gophermart/internal/storage"
	"github.com/vleukhin/gophermart/internal/types"
	"net/http"
	"time"
)

type AppConfig struct {
	Addr        string `env:"RUN_ADDRESS" envDefault:"localhost:8080"`
	DatabaseURI string `env:"DATABASE_URI"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"debug"`
	JwtKey      string `env:"JWT_KEY"`
}

type Application struct {
	cfg            *AppConfig
	db             storage.Storage
	UsersService   *services.UsersService
	OrdersService  *services.OrdersService
	AccrualService *services.AccrualService
}

func (cfg *AppConfig) Parse() error {
	err := env.Parse(cfg)
	if err != nil {
		return err
	}

	addr := pflag.StringP("addr", "a", cfg.Addr, "Server address")
	logLevel := pflag.StringP("log-level", "l", cfg.LogLevel, "Application log level")
	jwtKey := pflag.StringP("jwt-key", "j", cfg.JwtKey, "JWT key for authentication")

	pflag.Parse()

	cfg.Addr = *addr
	cfg.LogLevel = *logLevel
	cfg.JwtKey = *jwtKey

	return nil
}

func NewApplication(cfg *AppConfig) (*Application, error) {
	db, err := storage.NewPostgresStorage(cfg.DatabaseURI, time.Second*2)
	if err != nil {
		return nil, err
	}

	ordersCh := make(chan types.Order)

	userService := services.NewUserService(db, cfg.JwtKey)
	ordersService := services.NewOrdersService(db, ordersCh)
	accrualService := services.NewAccrualService(ordersService, ordersCh)

	go accrualService.Run(context.Background())

	app := Application{
		cfg:            cfg,
		db:             db,
		UsersService:   userService,
		OrdersService:  ordersService,
		AccrualService: accrualService,
	}

	err = app.migrate()
	if err != nil {
		return nil, err
	}

	return &app, nil
}

func (app *Application) Run(err chan<- error) {
	log.Info().Msg("Gopher-mart API listen at: " + app.cfg.Addr)
	err <- http.ListenAndServe(app.cfg.Addr, NewRouter(app))
}

func (app *Application) ShutDown() error {
	app.db.ShutDown()
	return nil
}

func (app *Application) migrate() error {
	return app.db.Migrate(context.TODO())
}
