package services

import (
	"context"
	"github.com/vleukhin/gophermart/internal/storage"
	"github.com/vleukhin/gophermart/internal/types"
	"strconv"
)

type OrdersService struct {
	storage  storage.Storage
	ordersCh chan types.Order
}

func NewOrdersService(storage storage.Storage, ordersCh chan types.Order) *OrdersService {
	return &OrdersService{
		storage:  storage,
		ordersCh: ordersCh,
	}
}

func (s OrdersService) List(ctx context.Context, userID int) ([]types.Order, error) {
	return s.storage.GetUserOrders(ctx, userID)
}

func (s OrdersService) Create(ctx context.Context, userID int, orderID string) error {
	order, err := s.storage.CreateOrder(ctx, userID, orderID)
	if err != nil {
		return err
	}

	s.ordersCh <- order

	return nil
}

func (s OrdersService) GetById(ctx context.Context, orderID string) (*types.Order, error) {
	return s.storage.GetOrderByID(ctx, orderID)
}

func (s OrdersService) ValidateOrderID(id string) bool {
	_, err := strconv.Atoi(id)
	if err != nil {
		return false
	}

	return true
}

func (s OrdersService) MarkOrderAsProcessed(ctx context.Context, orderID string, accrual int) error {
	return s.storage.UpdateOrder(ctx, orderID, types.OrderStatusProcessed, accrual)
}

func (s OrdersService) MarkOrderAsProcessing(ctx context.Context, orderID string) error {
	return s.storage.UpdateOrder(ctx, orderID, types.OrderStatusProcessing, 0)
}

func (s OrdersService) MarkOrderAsInvalid(ctx context.Context, orderID string, accrual int) error {
	return s.storage.UpdateOrder(ctx, orderID, types.OrderStatusProcessing, accrual)
}
