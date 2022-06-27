package types

import "time"

type (
	User struct {
		ID       int
		Name     string
		Password string
	}
	Order struct {
		ID         int       `json:"number"`
		UserID     int       `json:"-"`
		Status     string    `json:"status"`
		Accrual    int       `json:"accrual,omitempty"`
		UploadedAt time.Time `json:"uploaded_at"`
	}
	OrderStatus string
)

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)
