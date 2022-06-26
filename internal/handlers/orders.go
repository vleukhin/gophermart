package handlers

import (
	"fmt"
	"net/http"

	"github.com/vleukhin/gophermart/internal/services"
)

type OrdersController struct {
	service services.OrdersService
}

func NewOrdersController(service services.OrdersService) OrdersController {
	return OrdersController{service: service}
}

func (c OrdersController) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(AuthUserID)
	if userID == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Write([]byte(fmt.Sprintf("Welcome %d!", userID.(int))))
}
