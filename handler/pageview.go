package handler

import (
	"encoding/json"
	"net/http"

	"github.com/zackb/updog/db"
	"github.com/zackb/updog/domain"
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
func Handler(d *db.DB, ds domain.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req PageviewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.JSONError(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// TODO: find domain
		dsomain, _ := ds.ReadDomainByName(r.Context(), req.Domain)
		if dsomain == nil {
			httpx.JSONError(w, "domain not found", http.StatusNotFound)
			return
		}
		domain := &domain.Domain{Name: ""}

		country := &pageview.Country{Name: ""}
		_ = db.GetOrCreateDimension(r.Context(), d, country, "name", country.Name)

		region := &pageview.Region{Name: "", CountryID: country.ID}
		_ = db.GetOrCreateDimension(r.Context(), d, region, "name", region.Name)

		browser := &pageview.Browser{Name: ""}
		_ = db.GetOrCreateDimension(r.Context(), d, browser, "name", browser.Name)

		os := &pageview.OperatingSystem{Name: ""}
		_ = db.GetOrCreateDimension(r.Context(), d, os, "name", os.Name)

		deviceType := &pageview.DeviceType{Name: ""}
		_ = db.GetOrCreateDimension(r.Context(), d, deviceType, "name", deviceType.Name)

		language := &pageview.Language{Code: ""}
		_ = db.GetOrCreateDimension(r.Context(), d, language, "code", language.Code)

		referrer := &pageview.Referrer{Host: req.Referrer}
		_ = db.GetOrCreateDimension(r.Context(), d, referrer, "host", referrer.Host)

		// Insert Pageview
		pv := &pageview.Pageview{
			DomainID:     domain.ID,
			Path:         req.Path,
			CountryID:    country.ID,
			RegionID:     region.ID,
			BrowserID:    browser.ID,
			OSID:         os.ID,
			DeviceTypeID: deviceType.ID,
			LanguageID:   language.ID,
			ReferrerID:   referrer.ID,
		}

		_, err := d.Db.NewInsert().Model(pv).Exec(r.Context())
		if err != nil {
			httpx.JSONError(w, "failed to insert pageview", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
