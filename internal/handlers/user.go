package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/vleukhin/gophermart/internal/services"
)

type UserController struct {
	service services.UserService
}

func NewUserController(service services.UserService) UserController {
	return UserController{service: service}
}

type (
	RegisterParams struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
)

func (c UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var params RegisterParams
	errorLogger := log.Error().Str("method", "UserController::CreateUser")
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

	if !params.isValid() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = c.service.UserRegister(params.Login, params.Password)
	if err != nil {
		if errors.Is(err, services.ErrLUsernameTaken) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		errorLogger.Err(err).Msg("Failed to create user")
	}
}

func (p RegisterParams) isValid() bool {
	return p.Login != "" && p.Password != ""
}
