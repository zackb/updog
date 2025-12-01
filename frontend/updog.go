package frontend

import (
	"log"
	"net/http"
	"time"

	"github.com/zackb/updog/domain"
	"github.com/zackb/updog/httpx"
	"github.com/zackb/updog/user"
)

const (
	contextKeyUser = "user"
)

type UpdogRequest struct {
	W              http.ResponseWriter
	R              *http.Request
	User           *user.User
	Domains        []*domain.Domain
	SelectedDomain *domain.Domain
	Start, End     time.Time
}

type UpdogHandler func(*UpdogRequest) error

func (f *Frontend) WithUpdog(h UpdogHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := f.userFromRequest(r)
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		domains, selectedDomain, err := f.getDomainsAndSelected(r, user)
		if err != nil {
			log.Printf("Failed to list domains: %v", err)
			http.Error(w, "Failed to load domains", http.StatusInternalServerError)
			return
		}

		req := &UpdogRequest{
			W:              w,
			R:              r,
			User:           user,
			Domains:        domains,
			SelectedDomain: selectedDomain,
		}

		start, end, err := httpx.ParseTimeParams(r)
		if err != nil {
			http.Error(w, "Invalid time parameters", http.StatusBadRequest)
			return
		}

		req.Start = start
		req.End = end

		if err := h(req); err != nil {
			log.Printf("Handler error: %v", err)
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}
	}
}

func (f *Frontend) getDomainsAndSelected(r *http.Request, user *user.User) ([]*domain.Domain, *domain.Domain, error) {
	domains, err := f.db.DomainStorage().ListDomainsByUser(r.Context(), user.ID)
	if err != nil {
		return nil, nil, err
	}

	var selectedDomain *domain.Domain

	// check for cookie
	// TODO: change to query param
	if cookie, err := r.Cookie("selected_domain_id"); err == nil && cookie.Value != "" {
		for _, d := range domains {
			if d.ID == cookie.Value {
				selectedDomain = d
				break
			}
		}
	}

	// fallback: first verified domain
	if selectedDomain == nil {
		for _, d := range domains {
			if d.Verified {
				selectedDomain = d
				break
			}
		}
	}

	// fallback: first domain
	if selectedDomain == nil && len(domains) > 0 {
		selectedDomain = domains[0]
	}

	return domains, selectedDomain, nil
}

func (f *Frontend) userFromRequest(r *http.Request) *user.User {
	if u, ok := r.Context().Value(contextKeyUser).(*user.User); ok {
		return u
	}
	return nil
}
