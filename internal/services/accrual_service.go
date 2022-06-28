package services

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/vleukhin/gophermart/internal/types"
)

type AccrualService struct {
	ordersService   *OrdersService
	ordersCh        chan types.Order
	client          http.Client
	processMaxTries int
	accrualAddr     string
}

type OrderInfo struct {
	Order   string            `json:"order"`
	Status  types.OrderStatus `json:"status"`
	Accrual int               `json:"accrual"`
}

func NewAccrualService(addr string, ordersService *OrdersService, ordersCh chan types.Order) *AccrualService {
	client := http.Client{}
	client.Timeout = time.Second * 5

	return &AccrualService{
		ordersService:   ordersService,
		ordersCh:        ordersCh,
		client:          client,
		processMaxTries: 10,
		accrualAddr:     addr,
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
			go s.processOrder(ctx, order)
		case <-ctx.Done():
			log.Error().Err(ctx.Err())
			return
		}
	}
}

func (s *AccrualService) processOrder(ctx context.Context, order types.Order) {
	var (
		try   int
		delay = time.Millisecond * 100
		info  OrderInfo
		err   error
	)

	for {
		try++
		log.Debug().Str("order", order.ID).Int("try", try).Msgf("Processing order")
		if try > 0 {
			if try > s.processMaxTries {
				log.Error().Str("order", order.ID).Msgf("Accrual service is unavailable after %d tries", s.processMaxTries)
				return
			}
			time.Sleep(delay)
			delay *= 2
		}

		info, err = s.getOrderInfo(order)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get order info")
			continue
		}

		break
	}

	s.ordersService.ProcessedOrdersChan() <- info
}

func (s *AccrualService) getOrderInfo(order types.Order) (OrderInfo, error) {
	var info OrderInfo
	response, err := s.client.Get(s.accrualAddr + "/api/orders/" + order.ID)
	if err != nil {
		return info, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("Failed to close response body")
		}
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		return info, errors.New("bad status code: " + strconv.Itoa(response.StatusCode))
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return info, err
	}

	err = json.Unmarshal(body, &info)
	if err != nil {
		return info, err
	}

	return info, nil
}
