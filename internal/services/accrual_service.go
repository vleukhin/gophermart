package services

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/vleukhin/gophermart/internal/types"
	"strconv"
	"time"
)

type AccrualService struct {
	OrdersService *OrdersService
	ordersCh      chan types.Order
}

func NewAccrualService(ordersService *OrdersService, ordersCh chan types.Order) *AccrualService {
	return &AccrualService{
		OrdersService: ordersService,
		ordersCh:      ordersCh,
	}
}

func (s *AccrualService) Run(ctx context.Context) {
	for {
		select {
		case order, ok := <-s.ordersCh:
			if !ok {
				log.Debug().Msg("Accrual service stopped: orders channel closed")
				return
			}
			err := s.processOrder(order)
			if err != nil {
				log.Error().Err(err).Msg("Failed to process order")
			}
		case <-ctx.Done():
			log.Error().Err(ctx.Err())
			return
		}
	}
}

func (s *AccrualService) processOrder(order types.Order) error {
	id, _ := strconv.Atoi(order.ID)
	ctx := context.Background()
	err := s.OrdersService.MarkOrderAsProcessing(ctx, id)
	if err != nil {
		return err
	}

	accrual, err := s.getOrderAccrual(order)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get order accrual")
	}

	return s.OrdersService.MarkOrderAsProcessed(ctx, id, accrual)
}

func (s *AccrualService) getOrderAccrual(order types.Order) (int, error) {
	time.Sleep(5 * time.Second)
	return 500, nil
}
