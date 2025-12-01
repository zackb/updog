package httpx

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// ParseTimeParams parses "from" and "to" time parameters from the request URL.
func ParseTimeParams(r *http.Request) (time.Time, time.Time, error) {
	from := time.Now().UTC().AddDate(0, 0, -7) // default to 7 days ago
	to := time.Now().UTC()
	var err error

	f := r.URL.Query().Get("from")
	if f != "" {
		from, err = ParseTimeParam(f)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("Invalid 'from' date: %v", err)
		}
	}
	t := r.URL.Query().Get("to")
	if t != "" {
		to, err = ParseTimeParam(t)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("Invalid 'to' date: %v", err)
		}
	}
	return from, to, nil
}

// ParseTimeParam parses a time string from a query parameter.
// It supports:
// - RFC3339 (e.g. "2025-11-29T19:42:07Z")
// - "2006-01-02" (e.g. "2025-11-29")
// - Unix timestamp in seconds (e.g. "1732909327")
func ParseTimeParam(param string) (time.Time, error) {
	if param == "" {
		return time.Time{}, fmt.Errorf("empty time parameter")
	}

	// unix timestamp in seconds
	if i, err := strconv.ParseInt(param, 10, 64); err == nil {
		return time.Unix(i, 0).UTC(), nil
	}

	// RFC3339
	if t, err := time.Parse(time.RFC3339, param); err == nil {
		return t.UTC(), nil
	}

	// date only
	if t, err := time.Parse("2006-01-02", param); err == nil {
		return t.UTC(), nil
	}

	// date with time but no timezone
	if t, err := time.Parse("2006-01-02T15:04:05", param); err == nil {
		return t.UTC(), nil
	}

	return time.Time{}, fmt.Errorf("invalid time format: %s", param)
}
