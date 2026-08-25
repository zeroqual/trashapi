package server

import (
	"net/http"
	"trash/api/internal/user"
	"trash/api/pkg/helpers"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r *chi.Mux, h *user.UserHandler, j *helpers.JwtManager) {
	r.Get("/health", healthHandler)

	//public routes
	r.Group(func(r chi.Router) {
		r.Get("/health", healthHandler)
		r.Post("/sign-up", h.Register)
		r.Post("/sign-in", h.Login)
		r.Post("/refresh", h.Refresh)
	})

	//protected routes
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(*j))
		r.Get("/me", h.Me)

		r.With(RequirePermission(helpers.UserUpdate)).Post("/{id}", h.Update)
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
