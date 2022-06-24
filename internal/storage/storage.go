package storage

import (
	"context"
	"errors"

	"github.com/vleukhin/gophermart/internal/types"
)

var ErrDuplicateUserName = errors.New("this username is already taken")

type Storage interface {
	Ping(ctx context.Context) error
	ShutDown(ctx context.Context) error

	CreateUser(ctx context.Context, name string, password string) (created bool, err error)
	GetUser(ctx context.Context, name string) (*types.User, error)

	Migrate(ctx context.Context) error
}
