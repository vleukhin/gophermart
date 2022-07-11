package accrual

import (
	"errors"
)

type MockAccrual struct {
	orders map[string]OrderInfo
}

func NewMockAccrualService(orders map[string]OrderInfo) Service {
	return &MockAccrual{
		orders: orders,
	}
}

func (s MockAccrual) GetOrderInfo(orderID string) (OrderInfo, error) {
	info, ok := s.orders[orderID]
	if !ok {
		return OrderInfo{}, errors.New("order not found")
	}

	return info, nil
}
