package storage

import (
	"context"
	"github.com/jackc/pgx/v4"
	"time"
)

type PostgresStorage struct {
	conn *pgx.Conn
}

func NewPostgresStorage(dsn string, connTimeout time.Duration) (*PostgresStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connTimeout)
	defer cancel()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{
		conn: conn,
	}, nil
}

func (s *PostgresStorage) ShutDown(ctx context.Context) error {
	return s.conn.Close(ctx)
}

func (s *PostgresStorage) Ping(ctx context.Context) error {
	return s.conn.Ping(ctx)
}

func (s *PostgresStorage) CreateUser(name string, password string) error {
	return nil
}
