package user

import (
	"github.com/go-chi/chi/v5"
	"github.com/zackb/updog/auth"
	"github.com/zackb/updog/httpx/middleware"
)

type Handler struct {
	store Storage
	auth  *auth.Service
}

func NewHandler(store Storage, auth *auth.Service) *Handler {
	return &Handler{
		store: store,
		auth:  auth,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	// r.Post("/", h.handleCreateUser)
	r.Group(func(protected chi.Router) {
		protected.Use(middleware.AuthMiddleware(h.auth))
		// protected.Get("/users", ...)
	})

	return r
}
