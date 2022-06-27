package services

import (
	"context"
	"github.com/vleukhin/gophermart/internal/storage"
	"github.com/vleukhin/gophermart/internal/types"
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

func (s OrdersService) Create(ctx context.Context, userID, orderID int) error {
	order, err := s.storage.CreateOrder(ctx, userID, orderID)
	if err != nil {
		return err
	}

	s.ordersCh <- order

	return nil
}

func (s OrdersService) GetById(ctx context.Context, orderID int) (*types.Order, error) {
	return s.storage.GetOrderByID(ctx, orderID)
}

func (s OrdersService) ValidateOrderID(id int) bool {
	return true
}

func (s OrdersService) MarkOrderAsProcessed(ctx context.Context, orderID, accrual int) error {
	return s.storage.UpdateOrders(ctx, orderID, types.OrderStatusProcessed, accrual)
}

func (s OrdersService) MarkOrderAsProcessing(ctx context.Context, orderID int) error {
	return s.storage.UpdateOrders(ctx, orderID, types.OrderStatusProcessing, 0)
}

func (s OrdersService) MarkOrderAsInvalid(ctx context.Context, orderID, accrual int) error {
	return s.storage.UpdateOrders(ctx, orderID, types.OrderStatusProcessing, accrual)
}
