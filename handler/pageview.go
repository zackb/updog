package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/zackb/updog/db"
	"github.com/zackb/updog/domain"
	"github.com/zackb/updog/enrichment"
	"github.com/zackb/updog/httpx"
	"github.com/zackb/updog/pageview"
)

type PageviewRequest struct {
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	Referrer string `json:"ref"`
}

// Handler handles incoming pageview tracking requests.
func Handler(d *db.DB, ds domain.Storage, en *enrichment.Enricher, gif bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req PageviewRequest

		if r.Method == http.MethodPost {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				httpx.JSONError(w, "invalid JSON", http.StatusBadRequest)
				return
			}
		} else {
			req.Domain = r.URL.Query().Get("domain")
			req.Path = r.URL.Query().Get("path")
			req.Referrer = r.URL.Query().Get("ref")
		}

		if req.Domain == "" || req.Path == "" {
			httpx.JSONError(w, "missing required parameters", http.StatusBadRequest)
			return
		}

		// verify the request is coming from the claimed domain
		origin := r.Header.Get("Origin")
		referer := r.Referer()

		// if we have an origin header, it must match the claimed domain
		if origin != "" {
			u, err := url.Parse(origin)
			if err == nil {
				originHost := u.Hostname()
				if originHost != req.Domain && originHost != "localhost" && originHost != "127.0.0.1" && !strings.HasSuffix(originHost, "localhost") {
					log.Printf("Origin mismatch: %s != %s", originHost, req.Domain)
				}
			}
		} else if referer != "" {
			// if no origin, check referer
			u, err := url.Parse(referer)
			if err == nil {
				refererHost := u.Hostname()
				if refererHost != req.Domain && refererHost != "localhost" && refererHost != "127.0.0.1" && !strings.HasSuffix(refererHost, "localhost") {
					log.Printf("Referer mismatch: %s != %s", refererHost, req.Domain)
				}
			}
		}

		dsomain, _ := ds.ReadDomainByName(r.Context(), req.Domain)
		if dsomain == nil {
			httpx.JSONError(w, "domain not found", http.StatusNotFound)
			return
		}

		entry, err := en.Enrich(r)

		if httpx.CheckError(w, err) {
			return
		}

		country := &pageview.Country{Name: entry.Country}
		_ = db.GetOrCreateDimension(r.Context(), d, country, "name", country.Name)

		region := &pageview.Region{Name: entry.Region, CountryID: country.ID}
		_ = db.GetOrCreateRegion(r.Context(), d, region)

		browser := &pageview.Browser{Name: entry.Browser}
		_ = db.GetOrCreateDimension(r.Context(), d, browser, "name", browser.Name)

		os := &pageview.OperatingSystem{Name: entry.OS}
		_ = db.GetOrCreateDimension(r.Context(), d, os, "name", os.Name)

		deviceType := &pageview.DeviceType{Name: entry.DeviceType}
		_ = db.GetOrCreateDimension(r.Context(), d, deviceType, "name", deviceType.Name)

		language := &pageview.Language{Code: r.Header.Get("Accept-Language")}
		_ = db.GetOrCreateDimension(r.Context(), d, language, "code", language.Code)

		referrer := &pageview.Referrer{Host: req.Referrer}
		_ = db.GetOrCreateDimension(r.Context(), d, referrer, "host", referrer.Host)

		// Insert Pageview
		pv := &pageview.Pageview{
			DomainID:     dsomain.ID,
			Path:         req.Path,
			CountryID:    country.ID,
			RegionID:     region.ID,
			BrowserID:    browser.ID,
			OSID:         os.ID,
			DeviceTypeID: deviceType.ID,
			LanguageID:   language.ID,
			ReferrerID:   referrer.ID,
			VisitorID:    entry.VisitorID,
			Timestamp:    time.Now().UTC(),
		}

		// TODO: queue and model this better
		_, err = d.Db.NewInsert().Model(pv).Exec(r.Context())
		if err != nil {
			httpx.JSONError(w, "failed to insert pageview", http.StatusInternalServerError)
			return
		}

		if gif {
			w.Header().Set("Content-Type", "image/gif")
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Write([]byte{
				0x47, 0x49, 0x46, 0x38, 0x39, 0x61,
				0x01, 0x00, 0x01, 0x00,
				0x80, 0x00, 0x00,
				0x00, 0x00, 0x00,
				0xFF, 0xFF, 0xFF,
				0x21, 0xF9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00,
				0x2C, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00,
				0x02, 0x02, 0x44, 0x01, 0x00, 0x3B,
			})
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}
