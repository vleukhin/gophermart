package accrual

type (
	Service interface {
		GetOrderInfo(orderID string) (OrderInfo, error)
	}

	OrderInfo struct {
		OrderID string `json:"order"`
		Status  string `json:"status"`
		Accrual int    `json:"accrual"`
	}
)
