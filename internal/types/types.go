package types

import "time"

type (
	User struct {
		ID       int
		Name     string
		Password string
	}
	Order struct {
		ID         string      `json:"number"`
		UserID     int         `json:"-"`
		Status     OrderStatus `json:"status"`
		Accrual    float32     `json:"accrual,omitempty"`
		UploadedAt time.Time   `json:"uploaded_at"`
	}
	OrderStatus string
	Withdraw    struct {
		ID          int       `json:"-"`
		UserID      int       `json:"-"`
		OrderID     string    `json:"order"`
		Sum         float32   `json:"sum"`
		ProcessedAt time.Time `json:"processed_at"`
	}
	Balance struct {
		Current   float32 `json:"current"`
		Withdrawn float32 `json:"withdrawn"`
	}
)

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

func (o Order) isNew() bool {
	return o.Status == OrderStatusNew
}
