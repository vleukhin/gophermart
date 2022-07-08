package internal

import (
	"compress/gzip"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/vleukhin/gophermart/internal/handlers"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func NewRouter(app *Application) *mux.Router {
	userController := handlers.NewUserController(app.UsersService)
	ordersController := handlers.NewOrdersController(app.UsersService, app.OrdersService)
	balanceController := handlers.NewBalanceController(app.BalanceService, app.UsersService)

	r := mux.NewRouter().PathPrefix("/api").Subrouter()
	r.Use(gzipEncode)

	r.HandleFunc("/user/register", userController.Register).Methods(http.MethodPost)
	r.HandleFunc("/user/login", userController.Login).Methods(http.MethodPost)

	authRoutes := r.PathPrefix("").Subrouter()
	authRoutes.Use(userController.AuthMiddleware)

	authRoutes.HandleFunc("/user/orders", ordersController.Create).Methods(http.MethodPost)
	authRoutes.HandleFunc("/user/orders", ordersController.List).Methods(http.MethodGet)
	authRoutes.HandleFunc("/user/balance", balanceController.Balance).Methods(http.MethodGet)
	authRoutes.HandleFunc("/user/balance/withdraw", balanceController.Withdraw).Methods(http.MethodPost)
	authRoutes.HandleFunc("/user/withdrawals", balanceController.WithdrawalsList).Methods(http.MethodGet)

	return r
}

func gzipEncode(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestCompression)
		if err != nil {
			log.Error().Msg("Failed to create gzip writer: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func(gz *gzip.Writer) {
			err := gz.Close()
			if err != nil {
				log.Error().Msg("Failed to close gzip writer: " + err.Error())
			}
		}(gz)

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}
