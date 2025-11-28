package handler

import (
	"encoding/json"
	"net/http"

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
// <script>
//
//	navigator.sendBeacon("/pageview", JSON.stringify({
//		domain: location.hostname
//		path: location.pathname,
//		ref: document.referrer
//	}));
//
// </script>
func Handler(d *db.DB, ds domain.Storage, en *enrichment.Enricher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req PageviewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.JSONError(w, "invalid JSON", http.StatusBadRequest)
			return
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
		_ = db.GetOrCreateDimension(r.Context(), d, region, "name", region.Name)

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
		}

		// TODO: queue and model this better
		_, err = d.Db.NewInsert().Model(pv).Exec(r.Context())
		if err != nil {
			httpx.JSONError(w, "failed to insert pageview", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
