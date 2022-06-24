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
	AuthParams struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
)

func (c UserController) Register(w http.ResponseWriter, r *http.Request) {
	var params AuthParams
	errorLogger := log.Error().Str("method", "UserController::Register")
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

	err = c.service.Register(r.Context(), params.Login, params.Password)
	if err != nil {
		if errors.Is(err, services.ErrUsernameTaken) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		errorLogger.Err(err).Msg("Failed to create user")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (c UserController) Login(w http.ResponseWriter, r *http.Request) {
	var params AuthParams
	errorLogger := log.Error().Str("method", "UserController::Login")
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

	loggedIn, err := c.service.Login(r.Context(), params.Login, params.Password)
	if err != nil {
		errorLogger.Err(err).Msg("Failed to log in user")
		w.WriteHeader(http.StatusInternalServerError)
	}

	if !loggedIn {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
}

func (p AuthParams) isValid() bool {
	return p.Login != "" && p.Password != ""
}
