package handlers

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"github.com/vleukhin/gophermart/internal/services"
	"io"
	"io/ioutil"
	"net/http"
)

type UserController struct {
	service services.UserService
}

func NewUserController(service services.UserService) UserController {
	return UserController{service: service}
}

type (
	RegisterParams struct {
		Name     string `json:"name"`
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

	err = c.service.UserRegister(params.Name, params.Password)
	if err != nil {
		errorLogger.Err(err).Msg("Failed to create user")
	}
}
