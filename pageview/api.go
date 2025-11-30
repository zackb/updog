package pageview

import (
	"context"
	"encoding/json"
	"fmt"
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
		protected.Get("/", h.handleListPageviews)
		protected.Get("/hourly", h.handleGetHourlyStats)
		protected.Get("/daily", h.handleGetDailyStats)
		protected.Get("/monthly", h.handleGetMonthlyStats)
		// TODO: remove this
		protected.Get("/rollup", h.handleRollup)
	})

	return r
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

func (h *Handler) handleListPageviews(w http.ResponseWriter, r *http.Request) {
	userID := httpx.UserIDFromRequest(r)
	if userID == "" {
		httpx.JSONError(w, "Bad state", http.StatusInternalServerError)
		return
	}

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

	domainID, err := h.resolveDomainID(r, userID)

	if err != nil {
		log.Printf("Failed to resolve domain: %v", err)
		httpx.JSONError(w, "Failed to resolve domain", http.StatusInternalServerError)
		return
	}

	pvs, err := h.store.ListPageviewsByDomainID(r.Context(), domainID, from, to, 1000, 0)
	if err != nil {
		log.Println("Error reading pageviews:", err)
		httpx.JSONError(w, "Error reading pageviews", http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(pvs)
	httpx.CheckError(w, err)
}

func (h *Handler) handleGetHourlyStats(w http.ResponseWriter, r *http.Request) {
	h.handleGetStats(w, r, h.store.GetHourlyStats)
}

func (h *Handler) handleGetDailyStats(w http.ResponseWriter, r *http.Request) {
	h.handleGetStats(w, r, h.store.GetDailyStats)
}

func (h *Handler) handleGetMonthlyStats(w http.ResponseWriter, r *http.Request) {
	h.handleGetStats(w, r, h.store.GetMonthlyStats)
}

func (h *Handler) handleGetStats(w http.ResponseWriter, r *http.Request, statsFunc func(context.Context, string, time.Time, time.Time) ([]*AggregatedPoint, error)) {
	userID := httpx.UserIDFromRequest(r)
	if userID == "" {
		httpx.JSONError(w, "Bad state", http.StatusInternalServerError)
		return
	}

	from, to, err := h.parseTimeParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	domainID, err := h.resolveDomainID(r, userID)
	if err != nil {
		log.Printf("Failed to resolve domain: %v", err)
		http.Error(w, "Failed to resolve domain", http.StatusInternalServerError)
		return
	}
	if domainID == "" {
		http.Error(w, "No domain found", http.StatusNotFound)
		return
	}

	stats, err := statsFunc(r.Context(), domainID, from, to)
	if err != nil {
		log.Println("Error reading stats:", err)
		httpx.JSONError(w, "Error reading stats", http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(stats)
	httpx.CheckError(w, err)
}

func (h *Handler) parseTimeParams(r *http.Request) (time.Time, time.Time, error) {
	from := time.Now().AddDate(0, 0, -7)
	to := time.Now()
	var err error

	f := r.URL.Query().Get("from")
	if f != "" {
		from, err = httpx.ParseTimeParam(f)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("Invalid 'from' date: %v", err)
		}
	}
	t := r.URL.Query().Get("to")
	if t != "" {
		to, err = httpx.ParseTimeParam(t)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("Invalid 'to' date: %v", err)
		}
	}
	return from, to, nil
}

// resolveDomainID determines the domain ID to use based on the request parameters and user ownership.
func (h *Handler) resolveDomainID(r *http.Request, userID string) (string, error) {
	domains, err := h.domainStore.ListDomainsByUser(r.Context(), userID)
	if err != nil {
		return "", err
	}

	requestedDomainID := r.URL.Query().Get("domain_id")
	if requestedDomainID != "" {
		for _, d := range domains {
			if d.ID == requestedDomainID {
				return d.ID, nil
			}
		}
		// user requested a domain they don't own or doesn't exist
		return "", nil
	}

	requestedDomainName := r.URL.Query().Get("domain")
	if requestedDomainName != "" {
		for _, d := range domains {
			if d.Name == requestedDomainName {
				return d.ID, nil
			}
		}
		// user requested a domain they don't own or doesn't exist
		return "", nil
	}

	// default logic
	var selectedDomain *domain.Domain
	for _, d := range domains {
		if d.Verified {
			selectedDomain = d
			break
		}
	}

	if selectedDomain == nil && len(domains) > 0 {
		selectedDomain = domains[0]
	}

	if selectedDomain != nil {
		return selectedDomain.ID, nil
	}

	return "", nil
}
