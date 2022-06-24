package storage

import (
	"context"

	"github.com/vleukhin/gophermart/internal/types"
)

type Storage interface {
	Ping(ctx context.Context) error
	ShutDown(ctx context.Context) error

	CreateUser(ctx context.Context, name string, password string) (created bool, err error)
	GetUser(ctx context.Context, name string) (*types.User, error)

	Migrate(ctx context.Context) error
}
