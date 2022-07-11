package orders

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/vleukhin/gophermart/internal/services/accrual"
	"github.com/vleukhin/gophermart/internal/storage"
	"github.com/vleukhin/gophermart/internal/types"
)

const workersNumber = 2

type Service struct {
	storage        storage.Storage
	ordersCh       chan job
	ordersInfoCh   chan accrual.OrderInfo
	accrualService accrual.Service
	validator      OrderValidator
}

func NewService(storage storage.Storage, accrualService accrual.Service) *Service {
	ordersCh := make(chan job)
	ordersInfoCh := make(chan accrual.OrderInfo)
	for i := 0; i < workersNumber; i++ {
		w := newWorker(accrualService, ordersCh, ordersInfoCh)
		go w.Run()
	}

	service := &Service{
		storage:        storage,
		accrualService: accrualService,
		ordersCh:       ordersCh,
		ordersInfoCh:   ordersInfoCh,
		validator:      luhnValidator{},
	}

	go service.updateProcessedOrders()

	return service
}

func (s *Service) updateProcessedOrders() {
	ctx := context.TODO()
	var err error
	for info := range s.ordersInfoCh {
		switch info.Status {
		case string(types.OrderStatusProcessed):
			err = s.MarkOrderAsProcessed(ctx, info.OrderID, info.Accrual)
		case string(types.OrderStatusInvalid):
			err = s.MarkOrderAsInvalid(ctx, info.OrderID)
		case string(types.OrderStatusProcessing):
			err = s.MarkOrderAsProcessing(ctx, info.OrderID)
		default:
			log.Warn().Str("status", info.Status).Str("order", info.OrderID).Msg("Unknown order status")
		}

		if err != nil {
			log.Error().Str("order", info.OrderID).Err(err).Msg("Failed to update order")
			s.ordersCh <- newJob(info.OrderID, 10)
			continue
		}
	}
}

func (s *Service) List(ctx context.Context, userID int) ([]types.Order, error) {
	return s.storage.GetUserOrders(ctx, userID)
}

func (s *Service) Create(ctx context.Context, userID int, orderID string) error {
	order, err := s.storage.CreateOrder(ctx, userID, orderID)
	if err != nil {
		return err
	}

	s.Process(order.ID)

	return nil
}

func (s *Service) Process(orderID string) {
	s.ordersCh <- newJob(orderID, 10)
}

func (s *Service) GetByID(ctx context.Context, orderID string) (*types.Order, error) {
	return s.storage.GetOrderByID(ctx, orderID)
}

func (s *Service) ValidateOrderID(id string) bool {
	return s.validator.OrderNumberIsValid(id)
}

func (s *Service) MarkOrderAsProcessed(ctx context.Context, orderID string, accrual float32) error {
	return s.storage.UpdateOrder(ctx, orderID, types.OrderStatusProcessed, accrual)
}

func (s *Service) MarkOrderAsProcessing(ctx context.Context, orderID string) error {
	return s.storage.UpdateOrder(ctx, orderID, types.OrderStatusProcessing, 0)
}

func (s *Service) MarkOrderAsInvalid(ctx context.Context, orderID string) error {
	return s.storage.UpdateOrder(ctx, orderID, types.OrderStatusProcessing, 0)
}

func (s *Service) ShutDown() {
	close(s.ordersCh)
	close(s.ordersInfoCh)
}
