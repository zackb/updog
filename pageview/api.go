package pageview

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/zackb/updog/auth"
	"github.com/zackb/updog/domain"
	"github.com/zackb/updog/httpx"
	"github.com/zackb/updog/httpx/middleware"
)

type Handler struct {
	store       Storage
	domainStore domain.Storage
	auth        *auth.Service
}

func NewHandler(store Storage, domainStore domain.Storage, auth *auth.Service) *Handler {
	return &Handler{
		store:       store,
		domainStore: domainStore,
		auth:        auth,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(protected chi.Router) {
		protected.Use(middleware.AuthMiddleware(h.auth))
		protected.Get("/", h.WithApi(h.handleListPageviews))
		protected.Get("/visitors", h.WithApi(h.handleGetVisitors))
		protected.Get("/hourly", h.WithApi(h.handleGetHourlyStats))
		protected.Get("/daily", h.WithApi(h.handleGetDailyStats))
		protected.Get("/monthly", h.WithApi(h.handleGetMonthlyStats))
		protected.Get("/stats", h.WithApi(h.handleGetAggregatedStats))
		// TODO: remove this
		protected.Get("/rollup", h.handleRollup)
	})

	return r
}

func (h *Handler) handleGetVisitors(req *ApiRequest) error {

	visitors, err := h.store.GetGeoStats(req.R.Context(), req.DomainID, req.From, req.To)
	if err != nil {
		return NewApiError("Error reading visitors", http.StatusInternalServerError)
	}
	return json.NewEncoder(req.W).Encode(visitors)
}

func (h *Handler) handleRollup(w http.ResponseWriter, r *http.Request) {
	day := time.Now().AddDate(0, 0, -1)
	dayStr := r.URL.Query().Get("day")
	if dayStr != "" {
		var err error
		day, err = time.Parse("2006-01-02", dayStr)
		if err != nil {
			http.Error(w, "Invalid 'day' date", http.StatusBadRequest)
			return
		}
	}

	err := h.store.RunDailyRollup(r.Context(), day)
	if err != nil {
		log.Println("Error rolling up pageviews:", err)
		httpx.JSONError(w, "Error rolling up pageviews", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleListPageviews(req *ApiRequest) error {
	pvs, err := h.store.ListPageviewsByDomainID(req.R.Context(), req.DomainID, req.From, req.To, 1000, 0)
	if err != nil {
		return NewApiError("Error reading pageviews", http.StatusInternalServerError)
	}
	dtos := ToPageviewDTOs(pvs)
	return json.NewEncoder(req.W).Encode(dtos)
}

func (h *Handler) handleGetHourlyStats(req *ApiRequest) error {
	return h.handleGetStats(req, h.store.GetHourlyStats)
}

func (h *Handler) handleGetDailyStats(req *ApiRequest) error {
	return h.handleGetStats(req, h.store.GetDailyStats)
}

func (h *Handler) handleGetMonthlyStats(req *ApiRequest) error {
	return h.handleGetStats(req, h.store.GetMonthlyStats)
}

func (h *Handler) handleGetAggregatedStats(req *ApiRequest) error {

	stats, err := h.store.GetAggregatedStats(req.R.Context(), req.DomainID, req.From, req.To)

	if err != nil {
		log.Println("Error reading stats:", err)
		return NewApiError("Error reading stats", http.StatusInternalServerError)
	}
	return json.NewEncoder(req.W).Encode(stats)
}

func (h *Handler) handleGetStats(req *ApiRequest, statsFunc func(context.Context, string, time.Time, time.Time) ([]*AggregatedPoint, error)) error {
	stats, err := statsFunc(req.R.Context(), req.DomainID, req.From, req.To)
	if err != nil {
		return NewApiError("Error reading stats", http.StatusInternalServerError)
	}
	return json.NewEncoder(req.W).Encode(stats)
}
