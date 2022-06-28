package services

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/vleukhin/gophermart/internal/storage"
	"github.com/vleukhin/gophermart/internal/types"
	"strconv"
)

type OrdersService struct {
	storage           storage.Storage
	processCh         chan types.Order
	processedOrdersCh chan OrderInfo
}

func NewOrdersService(storage storage.Storage, processCh chan types.Order) *OrdersService {
	processedOrdersCh := make(chan OrderInfo, 1)

	service := &OrdersService{
		storage:           storage,
		processCh:         processCh,
		processedOrdersCh: processedOrdersCh,
	}

	go service.updateProcessedOrders()

	return service
}

func (s OrdersService) ProcessedOrdersChan() chan OrderInfo {
	return s.processedOrdersCh
}

func (s OrdersService) updateProcessedOrders() {
	ctx := context.TODO()
	var err error
	for info := range s.processedOrdersCh {
		switch info.Status {
		case types.OrderStatusProcessed:
			err = s.MarkOrderAsProcessed(ctx, info.Order, info.Accrual)
		case types.OrderStatusInvalid:
			err = s.MarkOrderAsInvalid(ctx, info.Order)
		case types.OrderStatusProcessing:
			err = s.MarkOrderAsProcessing(ctx, info.Order)
		}

		if err != nil {
			log.Error().Str("order", info.Order).Err(err).Msg("Failed to update order")
			s.ProcessedOrdersChan() <- info
			continue
		}
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

	s.processCh <- order

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

func (s OrdersService) MarkOrderAsInvalid(ctx context.Context, orderID string) error {
	return s.storage.UpdateOrder(ctx, orderID, types.OrderStatusProcessing, 0)
}
