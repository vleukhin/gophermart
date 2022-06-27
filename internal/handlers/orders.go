package handlers

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"github.com/vleukhin/gophermart/internal/services"
	"io"
	"io/ioutil"
	"net/http"
)

type OrdersController struct {
	usersService   *services.UsersService
	ordersService  *services.OrdersService
	accrualService *services.AccrualService
}

func NewOrdersController(usersService *services.UsersService, ordersService *services.OrdersService) OrdersController {
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

	errorLogger := log.Error().Str("method", "OrdersController::List")
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			errorLogger.Err(err).Msg("Failed to close request body")
		}
	}(r.Body)

	orders, err := c.ordersService.List(r.Context(), userID)
	if err != nil {
		errorLogger.Err(err).Msg("Failed to get orders")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response, err := json.Marshal(orders)
	if err != nil {
		errorLogger.Err(err).Msg("Failed to marshal JSON")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(response)
	if err != nil {
		errorLogger.Err(err).Msg("Failed to write response body")
		w.WriteHeader(http.StatusInternalServerError)
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

	orderID := string(body)

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
