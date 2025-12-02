package pageview

import (
	"log"
	"net/http"
	"os/user"
	"time"

	"github.com/zackb/updog/domain"
	"github.com/zackb/updog/httpx"
)

type ApiRequest struct {
	W        http.ResponseWriter
	R        *http.Request
	User     *user.User
	DomainID string
	From, To time.Time
}

type ApiHandler func(*ApiRequest) error

type ApiError struct {
	Message    string
	HTTPStatus int
}

func NewApiError(message string, httpStatus int) *ApiError {
	return &ApiError{
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

func (e *ApiError) Error() string {
	return e.Message
}

func (h *Handler) WithApi(a ApiHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := httpx.UserIDFromRequest(r)
		if userID == "" {
			httpx.JSONError(w, "Bad state", http.StatusInternalServerError)
			return
		}

		from, to, err := httpx.ParseTimeParams(r)
		if err != nil {
			log.Printf("Failed to parse time params: %v", err)
			httpx.JSONError(w, err.Error(), http.StatusBadRequest)
			return
		}

		domainID, err := h.resolveDomainID(r, userID)

		if err != nil || domainID == "" {
			log.Printf("Failed to resolve domain: %v", err)
			httpx.JSONError(w, "Failed to resolve domain", http.StatusInternalServerError)
			return
		}

		apiReq := &ApiRequest{
			W:        w,
			R:        r,
			DomainID: domainID,
			From:     from,
			To:       to,
		}

		err = a(apiReq)
		if err != nil {
			if apiErr, ok := err.(*ApiError); ok {
				httpx.JSONError(w, apiErr.Message, apiErr.HTTPStatus)
			} else {
				httpx.JSONError(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}
	}
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
