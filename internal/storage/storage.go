package storage

import (
	"context"

	"github.com/vleukhin/gophermart/internal/types"
)

type Storage interface {
	Ping(ctx context.Context) error
	ShutDown(ctx context.Context) error

	CreateUser(name string, password string) error
	GetUser(name string) (*types.User, error)
}
