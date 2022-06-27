package storage

import (
	"context"
	"strconv"
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
const createOrderSQL = `
	INSERT INTO orders (id, user_id, status, accrual, uploaded_at)
	VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING 
`

func (s *PostgresStorage) CreateOrder(ctx context.Context, userID, orderID int) (types.Order, error) {
	order := types.Order{
		ID:         strconv.Itoa(orderID),
		UserID:     userID,
		Status:     types.OrderStatusNew,
		UploadedAt: time.Now(),
	}

	var id int
	_, err := s.conn.Exec(ctx, createOrderSQL, &id, order.UserID, order.Status, 0, order.UploadedAt)

	order.ID = strconv.Itoa(id)

	if err != nil {
		log.Debug().Err(err)
		return order, err
	}

	return order, nil
}

// language=PostgreSQL
const getOrderByID = `SELECT user_id, status, accrual, uploaded_at FROM orders WHERE id = $1`

func (s *PostgresStorage) GetOrderByID(ctx context.Context, id int) (*types.Order, error) {
	order := types.Order{
		ID: strconv.Itoa(id),
	}

	row := s.conn.QueryRow(ctx, getOrderByID, id)
	err := row.Scan(&order.UserID, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &order, nil
}

// language=PostgreSQL
const getUserOrders = `SELECT id, user_id, status, accrual, uploaded_at FROM orders WHERE user_id = $1`

func (s *PostgresStorage) GetUserOrders(ctx context.Context, userID int) ([]types.Order, error) {
	var result []types.Order
	rows, err := s.conn.Query(ctx, getUserOrders, userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id int
		order := types.Order{}
		err := rows.Scan(&id, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}

		order.ID = strconv.Itoa(id)
		result = append(result, order)
	}

	return result, nil
}

// language=PostgreSQL
const updateOrder = `UPDATE orders SET status = $1, accrual = $2 WHERE id = $3`

func (s *PostgresStorage) UpdateOrders(ctx context.Context, orderID int, status types.OrderStatus, accrual int) error {
	_, err := s.conn.Exec(ctx, updateOrder, status, accrual, orderID)

	return err
}

// language=PostgreSQL
const createUsersTable = `
	CREATE TABLE IF NOT EXISTS users (
		id serial constraint users_pk primary key,
		name varchar(255) not null unique,
		password varchar(255) not null
	)
`

// language=PostgreSQL
const createOrdersTable = `
	CREATE TABLE IF NOT EXISTS orders (
		id bigserial constraint orders_pk primary key,
		user_id integer,
		status varchar(255) not null,
		accrual integer not null,
		uploaded_at timestamp not null
	)
`

func (s *PostgresStorage) Migrate(ctx context.Context) error {
	migrations := []string{
		createUsersTable,
		createOrdersTable,
	}

	for _, m := range migrations {
		_, err := s.conn.Exec(ctx, m)
		if err != nil {
			return err
		}
	}

	return nil
}
