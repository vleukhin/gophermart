package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/rs/zerolog/log"

	"github.com/vleukhin/gophermart/internal/services/users"
)

type UsersController struct {
	service users.Service
}

func NewUserController(service users.Service) UsersController {
	return UsersController{service: service}
}

type (
	Credentials struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
)

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
		if errors.Is(err, users.ErrUsernameTaken) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		errorLogger.Err(err).Msg("Failed to create users")
		w.WriteHeader(http.StatusInternalServerError)
	}

	c.authCookie(w, tokenString, ttl)
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
		errorResponse(w, err, errorLogger)
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
		errorLogger.Err(err).Msg("Failed to log in users")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.authCookie(w, tokenString, ttl)
}

func (c UsersController) authCookie(w http.ResponseWriter, token string, ttl time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
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

		r = r.WithContext(context.WithValue(r.Context(), users.AuthUserID, claims.UserID))

		next.ServeHTTP(w, r)
	})
}
