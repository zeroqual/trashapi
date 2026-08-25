package server

import (
	"net/http"
	"trash/api/internal/user"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r *chi.Mux, h *user.UserHandler) {
	r.Get("/health", healthHandler)

	//public routes
	r.Group(func(r chi.Router) {
		r.Get("/health", healthHandler)
		r.Post("/sign-up", h.Register)
		r.Post("/sign-in", h.Login)
	})

}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
