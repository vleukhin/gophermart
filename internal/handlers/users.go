package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/rs/zerolog/log"

	"github.com/vleukhin/gophermart/internal/services"
)

type UsersController struct {
	service services.UsersService
}

func NewUserController(service services.UsersService) UsersController {
	return UsersController{service: service}
}

type (
	Credentials struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	ContextKey string
)

const AuthUserID ContextKey = "userID"

func (c UsersController) Register(w http.ResponseWriter, r *http.Request) {
	var params Credentials
	errorLogger := log.Error().Str("method", "UsersController::Register")
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

	tokenString, ttl, err := c.service.Register(r.Context(), params.Login, params.Password)
	if err != nil {
		if errors.Is(err, services.ErrUsernameTaken) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		errorLogger.Err(err).Msg("Failed to create user")
		w.WriteHeader(http.StatusInternalServerError)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  ttl,
		Path:     "/",
		HttpOnly: true,
	})
}

func (c UsersController) Login(w http.ResponseWriter, r *http.Request) {
	var params Credentials
	errorLogger := log.Error().Str("method", "UsersController::Login")
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

	tokenString, ttl, err := c.service.Login(r.Context(), params.Login, params.Password)
	if err != nil {
		errorLogger.Err(err).Msg("Failed to log in user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  ttl,
		Path:     "/",
		HttpOnly: true,
	})
}

func (p Credentials) isValid() bool {
	return p.Login != "" && p.Password != ""
}

func (c UsersController) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := c.service.CheckAuth(r)
		if err != nil {
			if err == http.ErrNoCookie || err == jwt.ErrSignatureInvalid {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			log.Error().Err(err)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), AuthUserID, claims.UserID))

		next.ServeHTTP(w, r)
	})
}
