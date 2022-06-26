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

type SessionStorage interface {
	// Open opens new session and returns session ID
	Open(ctx context.Context) string
	// Close closes existing session
	Close(ctx context.Context, ID string) error
	// GC Cleanups old sessions
	GC(ctx context.Context)
}
