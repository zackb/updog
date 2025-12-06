package frontend

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/zackb/updog/auth"
	"github.com/zackb/updog/db"
	"github.com/zackb/updog/domain"
	"github.com/zackb/updog/env"
	"github.com/zackb/updog/id"
	"github.com/zackb/updog/pageview"
	"github.com/zackb/updog/settings"
)

//go:embed views/*.html
var viewsFS embed.FS

//go:embed public/img/* public/script/* public/style/*
var staticFS embed.FS

var tmpl = template.Must(template.New("").Funcs(Funcs).
	ParseFS(viewsFS, "views/*.html"))

var staticHandler http.Handler

type Frontend struct {
	auth          *auth.Service
	db            *db.DB
	ps            pageview.Storage
	staticHandler http.Handler
}

func NewFrontend(authSvc *auth.Service, database *db.DB) (*Frontend, error) {
	if err := initTemplatesAndStatic(); err != nil {
		log.Fatalf("Failed to initialize templates and static files: %v", err)
		return nil, err
	}

	return &Frontend{
		auth:          authSvc,
		db:            database,
		ps:            database.PageviewStorage(),
		staticHandler: staticHandler,
	}, nil
}

func (f *Frontend) Routes(mux *http.ServeMux) {

	// serve static files from the public directory
	mux.Handle("/static/", http.StripPrefix("/static/", staticHandler))

	mux.HandleFunc("/logout", f.logout)
	mux.HandleFunc("/join", f.join)
	mux.HandleFunc("/login", f.login)
	mux.HandleFunc("/dashboard", f.WithAuthenticated(f.WithUpdog(f.dashboard)))
	mux.HandleFunc("/realtime", f.WithAuthenticated(f.WithUpdog(f.realtime)))
	mux.HandleFunc("/domains", f.WithAuthenticated(f.WithUpdog(f.domains)))
	mux.HandleFunc("/domains/verify", f.WithAuthenticated(f.WithUpdog(f.verifyDomain)))
	mux.HandleFunc("/visitors", f.WithAuthenticated(f.WithUpdog(f.visitors)))
	mux.HandleFunc("/pages", f.WithAuthenticated(f.WithUpdog(f.pages)))
	mux.HandleFunc("/settings", f.WithAuthenticated(f.WithUpdog(f.settings)))
	mux.HandleFunc("/", f.index)
}

func (f *Frontend) dashboard(req *UpdogRequest) error {

	ctx := req.R.Context()

	// fetch pageview stats for selected domain
	stats := &DashboardStats{}
	if req.SelectedDomain != nil {
		stats.SelectedDomain = req.SelectedDomain
		// get pageviews for the last 30 days

		// get aggregated stats
		agg, err := f.ps.GetAggregatedStats(ctx, req.SelectedDomain.ID, req.Start, req.End)
		if err != nil {
			log.Printf("Failed to get aggregated stats: %v", err)
		} else {
			stats.Aggregated = agg
			stats.TotalPageviews = int(agg.TotalPageviews)
		}

		// graph data
		// end and start will be static for this chart
		graphEnd := time.Now().UTC()
		resolution := req.R.URL.Query().Get("resolution")
		if resolution == "" {
			resolution = "hourly"
		}
		stats.GraphResolution = resolution

		var graph []*pageview.AggregatedPoint

		switch resolution {
		case "daily":
			graph, err = f.ps.GetDailyStats(ctx, req.SelectedDomain.ID, graphEnd.AddDate(0, 0, -23), graphEnd)
		case "monthly":
			graph, err = f.ps.GetMonthlyStats(ctx, req.SelectedDomain.ID, graphEnd.AddDate(0, -23, 0), graphEnd)
		default: // hourly
			graph, err = f.ps.GetHourlyStats(ctx, req.SelectedDomain.ID, graphEnd.Add(-23*time.Hour), graphEnd)
		}

		if err != nil {
			log.Printf("Failed to get graph data: %v", err)
		} else {
			stats.GraphData = graph
			for _, d := range graph {
				if d.Count > stats.MaxViews {
					stats.MaxViews = d.Count
				}
			}
		}

		// top pages
		topPages, err := f.ps.GetTopPages(ctx, req.SelectedDomain.ID, req.Start, req.End, 5)
		if err != nil {
			log.Printf("Failed to get top pages: %v", err)
		} else {
			stats.TopPages = topPages
		}

		// device usage
		deviceUsage, err := f.ps.GetDeviceUsage(ctx, req.SelectedDomain.ID, req.Start, req.End)
		if err != nil {
			log.Printf("Failed to get device usage: %v", err)
		} else {
			stats.DeviceUsage = deviceUsage
		}
	}

	data := PageData{
		Title:   "Dashboard",
		User:    req.User,
		Stats:   stats,
		Slug:    "dashboard",
		Domains: req.Domains,
	}

	return tmpl.ExecuteTemplate(req.W, "dashboard.html", data)
}

func (f *Frontend) realtime(req *UpdogRequest) error {

	data := PageData{
		Title:   "Real-time",
		User:    req.User,
		Slug:    "realtime",
		Domains: req.Domains,
		Stats: &DashboardStats{
			SelectedDomain: req.SelectedDomain,
		},
	}

	return tmpl.ExecuteTemplate(req.W, "realtime.html", data)
}

func (f *Frontend) domains(req *UpdogRequest) error {

	ctx := req.R.Context()

	// POST create domain
	if req.R.Method == http.MethodPost {
		// create new domain
		name := req.R.FormValue("name")
		if name == "" {
			return NewUpError("Domain name is required", http.StatusBadRequest)
		}

		domain := &domain.Domain{
			ID:                id.NewID(),
			Name:              name,
			UserID:            req.User.ID,
			VerificationToken: id.NewID(),
			Verified:          false,
		}

		_, err := f.db.DomainStorage().CreateDomain(ctx, domain)
		if err != nil {
			log.Printf("Failed to create domain: %v", err)
			return NewUpError("Failed to create domain", http.StatusInternalServerError)
		}

		http.Redirect(req.W, req.R, "/domains", http.StatusSeeOther)

		return nil
	}

	data := PageData{
		Title:   "Domains",
		User:    req.User,
		Domains: req.Domains,
		Slug:    "domains",
		Stats: &DashboardStats{
			SelectedDomain: req.SelectedDomain,
		},
	}

	return tmpl.ExecuteTemplate(req.W, "domains.html", data)
}

func (f *Frontend) verifyDomain(req *UpdogRequest) error {

	ctx := req.R.Context()

	if req.R.Method != http.MethodPost {
		return NewUpError("Method not allowed", http.StatusMethodNotAllowed)
	}

	domainID := req.R.FormValue("domain_id")
	if domainID == "" {
		return NewUpError("Domain ID is required", http.StatusBadRequest)
	}

	d, err := f.db.DomainStorage().ReadDomain(ctx, domainID)
	if err != nil {
		log.Printf("Failed to read domain: %v", err)
		return NewUpError("Failed to read domain", http.StatusInternalServerError)
	}

	// verify ownership by checking for the file
	verificationURL := "https://" + d.Name + "/updog_" + d.VerificationToken + ".txt"
	client := &http.Client{}
	hreq, err := http.NewRequest("GET", verificationURL, nil)
	if err != nil {
		return NewUpError("Failed to create verification request", http.StatusInternalServerError)
	}

	// set a realistic User-Agent and Accept header
	hreq.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Updog/1.0)")
	hreq.Header.Set("Accept", "text/plain")

	resp, err := client.Do(hreq)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Verification failed for %s %d: %v", d.Name, resp.StatusCode, err)
		return NewUpError("Verification file not found", http.StatusBadRequest)
	}
	defer resp.Body.Close()

	err = f.db.DomainStorage().VerifyDomain(ctx, domainID)

	if err != nil {
		log.Printf("Failed to update domain: %v", err)
		return NewUpError("Failed to update domain", http.StatusInternalServerError)
	}

	http.Redirect(req.W, req.R, "/domains", http.StatusSeeOther)

	return nil
}

func (f *Frontend) settings(req *UpdogRequest) error {

	ctx := req.R.Context()

	data := PageData{
		Title:   "Settings",
		User:    req.User,
		Slug:    "settings",
		Domains: req.Domains,
		Stats: &DashboardStats{
			SelectedDomain: req.SelectedDomain,
		},
	}

	// POST update settings
	if req.R.Method == http.MethodPost {
		disableSignups := req.R.FormValue(settings.SettingDisableSignups) == "on"
		if err := f.db.SetValueAsBool(ctx, settings.SettingDisableSignups, disableSignups); err != nil {
			log.Printf("Failed to update settings: %v", err)
			return NewUpError("Failed to update settings", http.StatusInternalServerError)
		}

		// redirect to refresh the page and show updated state
		http.Redirect(req.W, req.R, "/settings", http.StatusSeeOther)
		return nil
	}

	// GET settings
	disableSignups, err := f.db.ReadValueAsBool(ctx, settings.SettingDisableSignups)
	if err != nil {
		log.Printf("Failed to read settings: %v", err)
		data.Error = "Failed to load settings"
	}

	data.Data = map[string]any{
		"DisableSignups": disableSignups,
	}

	return tmpl.ExecuteTemplate(req.W, "settings.html", data)
}

func (f *Frontend) visitors(req *UpdogRequest) error {

	data := PageData{
		Title:   "Visitors",
		User:    req.User,
		Slug:    "visitors",
		Domains: req.Domains,
		Stats: &DashboardStats{
			SelectedDomain: req.SelectedDomain,
		},
		AdditionalStyles: []string{
			"https://unpkg.com/leaflet@1.9.4/dist/leaflet.css",
		},
		AdditionalScripts: []string{
			"https://unpkg.com/leaflet@1.9.4/dist/leaflet.js",
		},
	}

	return tmpl.ExecuteTemplate(req.W, "visitors.html", data)
}

func (f *Frontend) pages(req *UpdogRequest) error {

	ctx := req.R.Context()

	data := PageData{
		Title:   "Pages",
		User:    req.User,
		Slug:    "pages",
		Domains: req.Domains,
		Stats: &DashboardStats{
			SelectedDomain: req.SelectedDomain,
		},
	}

	if req.SelectedDomain != nil {
		// Get top 100 pages
		topPages, err := f.ps.GetTopPages(ctx, req.SelectedDomain.ID, req.Start, req.End, 100)
		if err != nil {
			log.Printf("Failed to get top pages: %v", err)
		} else {
			data.Stats.TopPages = topPages
		}
	}

	return tmpl.ExecuteTemplate(req.W, "pages.html", data)
}

func (f *Frontend) index(w http.ResponseWriter, r *http.Request) {

	if f.auth.IsAuthenticated(r) != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	data := PageData{
		Title: "Welcome to Updog",
		User:  f.userFromRequest(r),
	}

	if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Failed to render index page", http.StatusInternalServerError)
	}
}

func initTemplatesAndStatic() error {
	if env.IsDev() {
		// dev mode from disk
		tmpl = template.Must(template.New("").Funcs(Funcs).
			ParseGlob(filepath.Join("frontend", "views", "*.html")))

		// serve static files from disk
		staticHandler = http.FileServer(http.Dir(filepath.Join("frontend", "public")))
	} else {
		// production from embedded FS
		tmpl = template.Must(template.New("").Funcs(Funcs).
			ParseFS(viewsFS, "views/*.html"))

		// serve static files from embedded FS
		staticSub, err := fs.Sub(staticFS, "public")
		if err != nil {
			return err
		}
		staticHandler = http.FileServer(http.FS(staticSub))
	}
	return nil
}
