package internal

import (
	"context"
	"net/http"
	"time"

	"github.com/caarlos0/env"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"

	"github.com/vleukhin/gophermart/internal/services/accrual"
	"github.com/vleukhin/gophermart/internal/services/balance"
	"github.com/vleukhin/gophermart/internal/services/orders"
	"github.com/vleukhin/gophermart/internal/services/users"
	"github.com/vleukhin/gophermart/internal/storage"
)

type AppConfig struct {
	Addr        string `env:"RUN_ADDRESS" envDefault:"localhost:8080"`
	DatabaseURI string `env:"DATABASE_URI"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"debug"`
	JwtKey      string `env:"JWT_KEY"`
	AccrualAddr string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://localhost:8888"`
}

type Application struct {
	Cfg            *AppConfig
	Db             storage.Storage
	UsersService   users.Service
	OrdersService  orders.Service
	BalanceService balance.Service
	AccrualService accrual.Service
}

func (cfg *AppConfig) Parse() error {
	err := env.Parse(cfg)
	if err != nil {
		return err
	}

	addr := pflag.StringP("addr", "a", cfg.Addr, "Server address")
	dsn := pflag.StringP("dsn", "d", cfg.DatabaseURI, "Database connection URI")
	logLevel := pflag.StringP("log-level", "l", cfg.LogLevel, "Application log level")
	jwtKey := pflag.StringP("jwt-key", "j", cfg.JwtKey, "JWT key for authentication")
	accrualAddr := pflag.StringP("acc-addr", "r", cfg.AccrualAddr, "Accrual system address")

	pflag.Parse()

	cfg.Addr = *addr
	cfg.DatabaseURI = *dsn
	cfg.LogLevel = *logLevel
	cfg.JwtKey = *jwtKey
	cfg.AccrualAddr = *accrualAddr

	return nil
}

func NewApplication(ctx context.Context, cfg *AppConfig) (*Application, error) {
	db, err := storage.NewPostgresStorage(cfg.DatabaseURI, time.Second*2)
	if err != nil {
		return nil, err
	}

	accrualService := accrual.NewDefaultAccrualService(cfg.AccrualAddr)
	userService := users.NewService(db, cfg.JwtKey)
	ordersService := orders.NewService(ctx, db, accrualService)
	balanceService := balance.NewService(db)

	app := Application{
		Cfg:            cfg,
		Db:             db,
		UsersService:   userService,
		OrdersService:  ordersService,
		AccrualService: accrualService,
		BalanceService: balanceService,
	}

	err = app.migrate(ctx)
	if err != nil {
		return nil, err
	}

	return &app, nil
}

func (app *Application) Run(err chan<- error) {
	log.Info().Msg("Gopher-mart API listen at: " + app.Cfg.Addr)
	err <- http.ListenAndServe(app.Cfg.Addr, NewRouter(app))
}

func (app *Application) ShutDown() error {
	app.Db.ShutDown()
	app.OrdersService.ShutDown()
	return nil
}

func (app *Application) migrate(ctx context.Context) error {
	return app.Db.Migrate(ctx)
}
