package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/vleukhin/gophermart/internal/services/users"
)

func jsonResponse(w http.ResponseWriter, v any, log *zerolog.Event) {
	response, err := json.Marshal(v)
	if err != nil {
		errorResponse(w, err, log)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(response)
	if err != nil {
		errorResponse(w, err, log)
		return
	}
}

func errorResponse(w http.ResponseWriter, err error, log *zerolog.Event) {
	w.WriteHeader(http.StatusInternalServerError)
	log.Err(err)
}

func checkAuth(w http.ResponseWriter, r *http.Request, service *users.Service) int {
	userID := service.GetAuthUserID(r.Context())
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
	}

	return userID
}
