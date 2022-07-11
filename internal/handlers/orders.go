package handlers

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/vleukhin/gophermart/internal/services/orders"
	"github.com/vleukhin/gophermart/internal/services/users"
)

type OrdersController struct {
	usersService  *users.Service
	ordersService *orders.Service
}

func NewOrdersController(usersService *users.Service, ordersService *orders.Service) OrdersController {
	return OrdersController{
		usersService:  usersService,
		ordersService: ordersService,
	}
}

func (c OrdersController) List(w http.ResponseWriter, r *http.Request) {
	errorLogger := log.Error().Str("method", "OrdersController::List")
	userID := checkAuth(w, r, c.usersService)
	if userID == 0 {
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			errorLogger.Err(err).Msg("Failed to close request body")
		}
	}(r.Body)

	ordersList, err := c.ordersService.List(r.Context(), userID)
	if err != nil {
		errorLogger.Err(err).Msg("Failed to get ordersList")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(ordersList) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	jsonResponse(w, ordersList, errorLogger)
}

func (c OrdersController) Create(w http.ResponseWriter, r *http.Request) {
	errorLogger := log.Error().Str("method", "OrdersController::Create")
	userID := checkAuth(w, r, c.usersService)
	if userID == 0 {
		return
	}

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

	orderID := string(body)

	if !c.ordersService.ValidateOrderID(orderID) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	existsOrder, err := c.ordersService.GetByID(r.Context(), orderID)
	if err != nil {
		errorResponse(w, err, errorLogger)
		return
	}

	if existsOrder != nil {
		if existsOrder.UserID == userID {
			w.WriteHeader(http.StatusOK)
			c.ordersService.Process(existsOrder.ID)
			return
		} else {
			w.WriteHeader(http.StatusConflict)
			return
		}
	}

	err = c.ordersService.Create(r.Context(), userID, orderID)
	if err != nil {
		errorResponse(w, err, errorLogger)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
