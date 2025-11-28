package httpx

import (
	"encoding/json"
	"log"
	"net/http"
)

const (
	ContextKeyUserID = "userID"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

func InvalidCredentials(w http.ResponseWriter) {
	JSONError(w, "Invalid credentials", http.StatusUnauthorized)
}

func JSONError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error: msg,
	})
}

func CheckError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	log.Println(err, "Checking error")
	w.WriteHeader(http.StatusInternalServerError)
	jsn, e := json.Marshal(err)
	if e != nil {
		log.Println(e, "marhsalling json error")
	}
	JSONError(w, string(jsn), http.StatusInternalServerError)
	return true
}

// Helper to get userID from request
func UserIDFromRequest(r *http.Request) string {
	if u, ok := r.Context().Value(ContextKeyUserID).(string); ok {
		return u
	}
	return ""
}
