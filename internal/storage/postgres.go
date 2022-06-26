package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog/log"

	"github.com/vleukhin/gophermart/internal/types"
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

// language=PostgreSQL
const createUserSQL = `
	INSERT INTO users (name, password)
	VALUES ($1, $2) ON CONFLICT DO NOTHING 
`

func (s *PostgresStorage) CreateUser(ctx context.Context, name string, password string) (bool, error) {
	t, err := s.conn.Exec(ctx, createUserSQL, name, password)

	if err != nil {
		log.Debug().Err(err)
		return false, err
	}

	if t.RowsAffected() == 0 {
		return false, nil
	}

	return true, err
}

// language=PostgreSQL
const getUserSQL = `SELECT id, name, password FROM users WHERE name = $1`

func (s *PostgresStorage) GetUser(ctx context.Context, name string) (*types.User, error) {
	var user types.User

	row := s.conn.QueryRow(ctx, getUserSQL, name)
	err := row.Scan(&user.ID, &user.Name, &user.Password)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

// language=PostgreSQL
const getUserByIDSQL = `SELECT id, name, password FROM users WHERE id = $1`

func (s *PostgresStorage) GetUserByID(ctx context.Context, id int) (*types.User, error) {
	var user types.User

	row := s.conn.QueryRow(ctx, getUserByIDSQL, id)
	err := row.Scan(&user.ID, &user.Name, &user.Password)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

// language=PostgreSQL
const createUsersTable = `
	CREATE TABLE IF NOT EXISTS users (
		id    serial constraint table_name_pk primary key,
		name  varchar(255) not null unique,
		password  varchar(255) not null
	)
`

// language=PostgreSQL
const createSessionsTable = `
	CREATE TABLE IF NOT EXISTS sessions (
		id            varchar(191) not null constraint sessions_id_unique unique,
		user_id       bigint,
		last_activity integer not null
	)
`

func (s *PostgresStorage) Migrate(ctx context.Context) error {
	migrations := []string{
		createUsersTable,
		createSessionsTable,
	}

	for _, m := range migrations {
		_, err := s.conn.Exec(ctx, m)
		if err != nil {
			return err
		}
	}

	return nil
}
