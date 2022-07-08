package storage

import (
	"context"

	"github.com/vleukhin/gophermart/internal/types"
)

type Storage interface {
	Ping(ctx context.Context) error
	ShutDown()
	Migrate(ctx context.Context) error

	UsersStorage
	OrdersStorage
	BalanceStorage
}

type UsersStorage interface {
	CreateUser(ctx context.Context, name string, password string) (created bool, err error)
	GetUser(ctx context.Context, name string) (*types.User, error)
	GetUserByID(ctx context.Context, id int) (*types.User, error)
}

type OrdersStorage interface {
	CreateOrder(ctx context.Context, userID int, orderID string) (types.Order, error)
	GetOrderByID(ctx context.Context, id string) (*types.Order, error)
	GetUserOrders(ctx context.Context, userID int) ([]types.Order, error)
	UpdateOrder(ctx context.Context, orderID string, status types.OrderStatus, accrual float32) error
}

type BalanceStorage interface {
	GetBalance(ctx context.Context, userID int) (float32, error)
	CreateWithdraw(ctx context.Context, userID int, amount float32) error
	GetWithDrawAmount(ctx context.Context) (float32, error)
}
