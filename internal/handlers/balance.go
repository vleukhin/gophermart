package handlers

import (
	"encoding/json"
	"github.com/vleukhin/gophermart/internal/services/balance"
	"github.com/vleukhin/gophermart/internal/services/users"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/rs/zerolog/log"
)

type (
	BalanceController struct {
		balanceService *balance.Service
		usersService   *users.Service
	}
	WithdrawParams struct {
		Order string  `json:"order"`
		Sum   float32 `json:"sum"`
	}
)

func NewBalanceController(balanceService *balance.Service, usersService *users.Service) BalanceController {
	return BalanceController{
		balanceService: balanceService,
		usersService:   usersService,
	}
}

func (c BalanceController) Balance(w http.ResponseWriter, r *http.Request) {
	errorLogger := log.Error().Str("method", "BalanceController::Balance")
	userID := c.usersService.GetAuthUserID(r.Context())
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	bal, err := c.balanceService.Balance(r.Context(), userID)
	if err != nil {
		log.Error().Msg(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(bal)
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

func (c BalanceController) Withdraw(w http.ResponseWriter, r *http.Request) {
	var params WithdrawParams
	var errorLogger = log.Error().Str("method", "BalanceController::CreateWithdraw")

	userID := c.usersService.GetAuthUserID(r.Context())
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
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
		log.Error().Msg(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !success {
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}
}

func (c BalanceController) WithdrawalsList(w http.ResponseWriter, r *http.Request) {
	errorLogger := log.Error().Str("method", "OrdersController::List")
	userID := c.usersService.GetAuthUserID(r.Context())
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	list, err := c.balanceService.WithdrawalsList(r.Context(), userID)
	if err != nil {
		errorLogger.Msg(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(list) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response, err := json.Marshal(list)
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
