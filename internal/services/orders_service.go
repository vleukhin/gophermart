package services

import (
	"context"
	"github.com/vleukhin/gophermart/internal/storage"
	"github.com/vleukhin/gophermart/internal/types"
)

type OrdersService struct {
	storage storage.Storage
}

func NewOrdersService(storage storage.Storage) OrdersService {
	return OrdersService{
		storage: storage,
	}
}

func (s OrdersService) List(ctx context.Context, userID int) ([]types.Order, error) {
	return nil, nil
}
