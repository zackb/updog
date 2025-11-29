package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zackb/updog/auth"
	"github.com/zackb/updog/db"
	"github.com/zackb/updog/domain"
	"github.com/zackb/updog/httpx"
	"github.com/zackb/updog/httpx/middleware"
	"github.com/zackb/updog/pageview"
	"github.com/zackb/updog/user"
)

type API struct {
	db   *db.DB
	auth *auth.Service
}

func NewAPI(db *db.DB, auth *auth.Service) *API {
	return &API{
		db:   db,
		auth: auth,
	}
}
func (a *API) Routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.LoggerMiddleware, middleware.JsonContentTypeMiddleware, middleware.CorsMiddleware)

	r.Get("/api/v1/healthz", healthCheckHandler)

	r.Route("/api/v1", func(api chi.Router) {
		us := a.db.UserStorage()
		ds := a.db.DomainStorage()
		ps := a.db.PageviewStorage()
		api.Mount("/pageviews", pageview.NewHandler(ps, ds, a.auth).Routes())
		api.Mount("/domains", domain.NewHandler(ds, a.auth).Routes())
		api.Mount("/users", user.NewHandler(us, a.auth).Routes())

		// auth
		api.Post("/auth/login", a.handleLogin)
		api.Post("/auth/logout", a.handleLogout)
		api.Post("/auth/register", a.handleRegister)
		api.Post("/auth/verify", a.handleVerify)
	})
	return r
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{"status": "OK"}
	err := json.NewEncoder(w).Encode(data)
	httpx.CheckError(w, err)
}
