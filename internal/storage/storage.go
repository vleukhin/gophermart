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
}

type UsersStorage interface {
	CreateUser(ctx context.Context, name string, password string) (created bool, err error)
	GetUser(ctx context.Context, name string) (*types.User, error)
	GetUserByID(ctx context.Context, id int) (*types.User, error)
}

type OrdersStorage interface {
	CreateOrder(ctx context.Context, userID, orderID int) (types.Order, error)
	GetOrderByID(ctx context.Context, id int) (*types.Order, error)
	GetUserOrders(ctx context.Context, userID int) ([]types.Order, error)
	UpdateOrders(ctx context.Context, orderID int, status types.OrderStatus, accrual int) error
}
