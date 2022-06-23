package storage

import "context"

type Storage interface {
	Ping(ctx context.Context) error
	ShutDown(ctx context.Context) error

	CreateUser(name string, password string) error
}
