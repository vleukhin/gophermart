package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/vleukhin/gophermart/internal/types"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(dsn string, connTimeout time.Duration) (*PostgresStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connTimeout)
	defer cancel()

	conn, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{
		pool: conn,
	}, nil
}

func (s *PostgresStorage) ShutDown() {
	s.pool.Close()
}

func (s *PostgresStorage) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// language=PostgreSQL
const createUserQuery = `
	INSERT INTO users (name, password)
	VALUES ($1, $2) ON CONFLICT DO NOTHING 
`

func (s *PostgresStorage) CreateUser(ctx context.Context, name string, password string) (bool, error) {
	t, err := s.pool.Exec(ctx, createUserQuery, name, password)

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
const getUserQuery = `SELECT id, name, password FROM users WHERE name = $1`

func (s *PostgresStorage) GetUser(ctx context.Context, name string) (*types.User, error) {
	var user types.User

	row := s.pool.QueryRow(ctx, getUserQuery, name)
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
const getUserByIDQuery = `SELECT id, name, password FROM users WHERE id = $1`

func (s *PostgresStorage) GetUserByID(ctx context.Context, id int) (*types.User, error) {
	var user types.User

	row := s.pool.QueryRow(ctx, getUserByIDQuery, id)
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
const createOrderQuery = `
	INSERT INTO orders (id, user_id, status, accrual, uploaded_at)
	VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING 
`

func (s *PostgresStorage) CreateOrder(ctx context.Context, userID int, orderID string) (types.Order, error) {
	order := types.Order{
		ID:         orderID,
		UserID:     userID,
		Status:     types.OrderStatusNew,
		UploadedAt: time.Now(),
	}

	_, err := s.pool.Exec(ctx, createOrderQuery, order.ID, order.UserID, order.Status, 0, order.UploadedAt)

	return order, err
}

// language=PostgreSQL
const getOrderByIDQuery = `SELECT user_id, status, accrual, uploaded_at FROM orders WHERE id = $1`

func (s *PostgresStorage) GetOrderByID(ctx context.Context, id string) (*types.Order, error) {
	order := types.Order{
		ID: id,
	}

	row := s.pool.QueryRow(ctx, getOrderByIDQuery, id)
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
const getUserOrdersQuery = `SELECT id, user_id, status, accrual, uploaded_at FROM orders WHERE user_id = $1`

func (s *PostgresStorage) GetUserOrders(ctx context.Context, userID int) ([]types.Order, error) {
	var result []types.Order
	rows, err := s.pool.Query(ctx, getUserOrdersQuery, userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		order := types.Order{}
		err := rows.Scan(&order.ID, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}

		result = append(result, order)
	}

	return result, nil
}

// language=PostgreSQL
const updateOrderQuery = `UPDATE orders SET status = $1, accrual = $2 WHERE id = $3`

func (s *PostgresStorage) UpdateOrder(ctx context.Context, orderID string, status types.OrderStatus, accrual float32) error {
	_, err := s.pool.Exec(ctx, updateOrderQuery, status, accrual, orderID)
	return err
}

// language=PostgreSQL
const createWithdrawSQL = `INSERT INTO withdrawals (user_id, order_id, sum, processed_at)  VALUES ($1, $2, $3, $4)`

func (s *PostgresStorage) CreateWithdraw(ctx context.Context, userID int, orderID string, sum float32) error {
	_, err := s.pool.Exec(ctx, createWithdrawSQL, userID, orderID, sum, time.Now())
	return err
}

// language=PostgreSQL
const getWithdrawalsSumQuery = `SELECT COALESCE(sum(sum), 0) as sum FROM withdrawals WHERE user_id = $1`

func (s *PostgresStorage) GetWithdrawalsSum(ctx context.Context, userID int) (float32, error) {
	var sum float32

	row := s.pool.QueryRow(ctx, getWithdrawalsSumQuery, userID)
	err := row.Scan(&sum)
	if err != nil && err != pgx.ErrNoRows {
		return 0, err
	}

	return sum, nil
}

// language=PostgreSQL
const getAccrualSumQuery = `SELECT COALESCE(sum(accrual), 0) as balance FROM orders WHERE user_id = $1 AND status = $2`

func (s *PostgresStorage) GetAccrualSum(ctx context.Context, userID int) (float32, error) {
	var balance float32

	row := s.pool.QueryRow(ctx, getAccrualSumQuery, userID, types.OrderStatusProcessed)
	err := row.Scan(&balance)
	if err != nil && err != pgx.ErrNoRows {
		return 0, err
	}

	return balance, nil
}

// language=PostgreSQL
const getUserWithdrawalsQuery = `SELECT id, user_id, order_id, sum, processed_at FROM withdrawals WHERE user_id = $1`

func (s *PostgresStorage) GetWithdrawals(ctx context.Context, userID int) ([]types.Withdraw, error) {
	var result []types.Withdraw
	rows, err := s.pool.Query(ctx, getUserWithdrawalsQuery, userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		withdraw := types.Withdraw{}
		err := rows.Scan(&withdraw.ID, &withdraw.UserID, &withdraw.OrderID, &withdraw.Sum, &withdraw.ProcessedAt)
		if err != nil {
			return nil, err
		}

		result = append(result, withdraw)
	}

	return result, nil
}

// language=PostgreSQL
const createUsersTableQuery = `
	CREATE TABLE IF NOT EXISTS users (
		id serial constraint users_pk primary key,
		name varchar(255) not null unique,
		password varchar(255) not null
	)
`

// language=PostgreSQL
const createOrdersTableQuery = `
	CREATE TABLE IF NOT EXISTS orders (
		id varchar(255) constraint orders_pk primary key,
		user_id integer,
		status varchar(255) not null,
		accrual float8 not null,
		uploaded_at timestamp not null
	)
`

// language=PostgreSQL
const createWithdrawalsTableQuery = `
	CREATE TABLE IF NOT EXISTS withdrawals (
		id serial constraint withdraw_pk primary key,
		user_id integer,
		order_id varchar(255),
		sum float8 not null,
		processed_at timestamp not null
	)
`

func (s *PostgresStorage) Migrate(ctx context.Context) error {
	migrations := []string{
		createUsersTableQuery,
		createOrdersTableQuery,
		createWithdrawalsTableQuery,
	}

	for _, m := range migrations {
		_, err := s.pool.Exec(ctx, m)
		if err != nil {
			return err
		}
	}

	return nil
}
