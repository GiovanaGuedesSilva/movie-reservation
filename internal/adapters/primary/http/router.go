package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter creates and configures the application router with global
// middlewares and the health check endpoint.
//
// The env parameter is embedded in the /health response so clients can verify
// which environment the API is serving (e.g. "development", "production").
func NewRouter(env string) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"env":    env,
		})
	})

	r.Route("/api/v1", func(r chi.Router) {
		// Feature routes will be registered here in upcoming stages.
		// Example:
		//   r.Mount("/auth",         authHandler.Routes())
		//   r.Mount("/movies",       movieHandler.Routes())
		//   r.Mount("/showtimes",    showtimeHandler.Routes())
		//   r.Mount("/reservations", reservationHandler.Routes())
	})

	return r
}
