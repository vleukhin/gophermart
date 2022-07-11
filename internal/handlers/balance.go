package handlers

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/vleukhin/gophermart/internal/services/balance"
	"github.com/vleukhin/gophermart/internal/services/users"
)

type (
	BalanceController struct {
		balanceService balance.Service
		usersService   users.Service
	}
	WithdrawParams struct {
		Order string  `json:"order"`
		Sum   float32 `json:"sum"`
	}
)

func NewBalanceController(balanceService balance.Service, usersService users.Service) BalanceController {
	return BalanceController{
		balanceService: balanceService,
		usersService:   usersService,
	}
}

func (c BalanceController) Balance(w http.ResponseWriter, r *http.Request) {
	errorLogger := log.Error().Str("method", "BalanceController::Balance")
	userID := checkAuth(w, r, c.usersService)
	if userID == 0 {
		return
	}

	bal, err := c.balanceService.Balance(r.Context(), userID)
	if err != nil {
		errorResponse(w, err, errorLogger)
		return
	}

	jsonResponse(w, bal, errorLogger)
}

func (c BalanceController) Withdraw(w http.ResponseWriter, r *http.Request) {
	var params WithdrawParams
	var errorLogger = log.Error().Str("method", "BalanceController::CreateWithdraw")

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
		log.Error().Msg(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	success, err := c.balanceService.CreateWithdraw(r.Context(), userID, params.Order, params.Sum)
	if err != nil {
		errorResponse(w, err, errorLogger)
		return
	}

	if !success {
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}
}

func (c BalanceController) WithdrawalsList(w http.ResponseWriter, r *http.Request) {
	errorLogger := log.Error().Str("method", "OrdersController::List")
	userID := checkAuth(w, r, c.usersService)
	if userID == 0 {
		return
	}

	list, err := c.balanceService.WithdrawalsList(r.Context(), userID)
	if err != nil {
		errorResponse(w, err, errorLogger)
		return
	}

	if len(list) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	jsonResponse(w, list, errorLogger)
}
