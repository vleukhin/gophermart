package handlers

import (
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/vleukhin/gophermart/internal/services"
)

type OrdersController struct {
	usersService  services.UsersService
	ordersService services.OrdersService
}

func NewOrdersController(usersService services.UsersService, ordersService services.OrdersService) OrdersController {
	return OrdersController{
		usersService:  usersService,
		ordersService: ordersService,
	}
}

func (c OrdersController) List(w http.ResponseWriter, r *http.Request) {
	userID := c.usersService.GetAuthUserID(r.Context())
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

}

func (c OrdersController) Create(w http.ResponseWriter, r *http.Request) {
	userID := c.usersService.GetAuthUserID(r.Context())
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	errorLogger := log.Error().Str("method", "OrdersController::Create")
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			errorLogger.Err(err).Msg("Failed to close request body")
		}
	}(r.Body)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorLogger.Err(err).Msg("Failed read request body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	orderID, err := strconv.Atoi(string(body))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !c.ordersService.ValidateOrderID(orderID) {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}

	existsOrder, err := c.ordersService.GetById(r.Context(), orderID)
	if err != nil {
		errorLogger.Err(err).Msg("Failed to get order")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if existsOrder != nil {
		if existsOrder.UserID == userID {
			w.WriteHeader(http.StatusOK)
			return
		} else {
			w.WriteHeader(http.StatusConflict)
			return
		}
	}

	err = c.ordersService.Create(r.Context(), userID, orderID)
	if err != nil {
		errorLogger.Err(err).Msg("Failed to create order")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	return
}
