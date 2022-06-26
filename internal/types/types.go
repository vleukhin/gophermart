package types

import "time"

type User struct {
	ID       int
	Name     string
	Password string
}

type Order struct {
	ID         int
	UserID     int
	Status     string
	Accrual    int
	UploadedAt time.Time
}
