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
	return s.storage.GetUserOrders(ctx, userID)
}

func (s OrdersService) Create(ctx context.Context, userID, orderID int) error {
	return s.storage.CreateOrder(ctx, userID, orderID)
}

func (s OrdersService) GetById(ctx context.Context, orderID int) (*types.Order, error) {
	return s.storage.GetOrderByID(ctx, orderID)
}

func (s OrdersService) ValidateOrderID(id int) bool {
	return true
}
