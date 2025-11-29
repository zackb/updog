package pageview

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/zackb/updog/auth"
	"github.com/zackb/updog/httpx"
	// "github.com/zackb/updog/httpx/middleware"
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

	r.Group(func(protected chi.Router) {
		// protected.Use(middleware.AuthMiddleware(h.auth))
		protected.Get("/", h.handleListPageviews)
	})

	return r
}

func (h *Handler) handleListPageviews(w http.ResponseWriter, r *http.Request) {
	/*
		userID := httpx.UserIDFromRequest(r)
		if userID == "" {
			httpx.JSONError(w, "Bad state", http.StatusInternalServerError)
			return
		}
	*/

	from := time.Now().AddDate(0, 0, -7)
	to := time.Now()
	var err error

	f := r.URL.Query().Get("from")
	if f != "" {
		from, err = time.Parse(f, "2006-01-02")
		if err != nil {
			http.Error(w, "Invalid 'from' date", http.StatusBadRequest)
			return
		}
	}
	t := r.URL.Query().Get("to")
	if t != "" {
		to, err = time.Parse(t, "2006-01-02")
		if err != nil {
			http.Error(w, "Invalid 'to' date", http.StatusBadRequest)
			return
		}
	}

	// TOOD: get domains for user
	// TODO: limit and offset
	pvs, err := h.store.ListPageviewsByDomainID(r.Context(), "1111", from, to, 1000, 0)
	if err != nil {
		log.Println("Error reading pageviews:", err)
		httpx.JSONError(w, "Error reading pageviews", http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(pvs)
	httpx.CheckError(w, err)
}
