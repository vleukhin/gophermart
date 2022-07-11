package storage

import (
	"context"
	"github.com/vleukhin/gophermart/internal/types"
)

type EmptyStorage struct {
}

func (e EmptyStorage) Ping(_ context.Context) error {
	return nil
}

func (e EmptyStorage) ShutDown() {
}

func (e EmptyStorage) Migrate(_ context.Context) error {
	return nil
}

func (e EmptyStorage) CreateUser(_ context.Context, _ string, _ string) (created bool, err error) {
	return true, nil
}

func (e EmptyStorage) GetUser(_ context.Context, _ string) (*types.User, error) {
	return &types.User{}, nil
}

func (e EmptyStorage) GetUserByID(_ context.Context, _ int) (*types.User, error) {
	return &types.User{}, nil
}

func (e EmptyStorage) CreateOrder(_ context.Context, _ int, _ string) (types.Order, error) {
	return types.Order{}, nil
}

func (e EmptyStorage) GetOrderByID(_ context.Context, _ string) (*types.Order, error) {
	return &types.Order{}, nil
}

func (e EmptyStorage) GetUserOrders(_ context.Context, _ int) ([]types.Order, error) {
	return []types.Order{}, nil
}

func (e EmptyStorage) UpdateOrder(_ context.Context, _ string, _ types.OrderStatus, _ float32) error {
	return nil
}

func (e EmptyStorage) GetAccrualSum(_ context.Context, _ int) (float32, error) {
	return 0, nil
}

func (e EmptyStorage) CreateWithdraw(_ context.Context, _ int, _ string, _ float32) error {
	return nil
}

func (e EmptyStorage) GetWithdrawalsSum(_ context.Context, _ int) (float32, error) {
	return 0, nil
}

func (e EmptyStorage) GetWithdrawals(_ context.Context, _ int) ([]types.Withdraw, error) {
	return []types.Withdraw{}, nil
}
