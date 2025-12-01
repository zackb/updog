package frontend

import (
	"context"
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
	"github.com/zackb/updog/httpx"
	"github.com/zackb/updog/id"
	"github.com/zackb/updog/pageview"
	"github.com/zackb/updog/settings"
	"github.com/zackb/updog/user"
)

const (
	contextKeyUser = "user"
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
	mux.HandleFunc("/dashboard", f.authMiddleware(f.dashboard))
	mux.HandleFunc("/realtime", f.authMiddleware(f.realtime))
	mux.HandleFunc("/domains", f.authMiddleware(f.domains))
	mux.HandleFunc("/domains/verify", f.authMiddleware(f.verifyDomain))
	mux.HandleFunc("/visitors", f.authMiddleware(f.visitors))
	mux.HandleFunc("/pages", f.authMiddleware(f.pages))
	mux.HandleFunc("/settings", f.authMiddleware(f.settings))
	mux.HandleFunc("/", f.index)
}

func (f *Frontend) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := f.auth.IsAuthenticated(r)
		if token == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, err := f.db.UserStorage().ReadUser(r.Context(), token.ClientId)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyUser, user)
		next(w, r.WithContext(ctx))
	}
}

func (f *Frontend) dashboard(w http.ResponseWriter, r *http.Request) {
	user := f.userFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	domains, selectedDomain, err := f.getDomainsAndSelected(r, user)
	if err != nil {
		log.Printf("Failed to list domains: %v", err)
		http.Error(w, "Failed to load dashboard", http.StatusInternalServerError)
		return
	}

	// fetch pageview stats for selected domain
	stats := &DashboardStats{}
	if selectedDomain != nil {
		stats.SelectedDomain = selectedDomain
		// get pageviews for the last 30 days
		start, end, err := httpx.ParseTimeParams(r)

		// get aggregated stats
		agg, err := f.ps.GetAggregatedStats(r.Context(), selectedDomain.ID, start, end)
		if err != nil {
			log.Printf("Failed to get aggregated stats: %v", err)
		} else {
			stats.Aggregated = agg
			stats.TotalPageviews = int(agg.TotalPageviews)
		}

		// graph data
		// TODO: hourly vs daily vs monthly
		// end and start will be static for this chart
		graphEnd := time.Now().UTC()
		graph, err := f.ps.GetHourlyStats(r.Context(), selectedDomain.ID, graphEnd.Add(-23*time.Hour), graphEnd)
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
		topPages, err := f.ps.GetTopPages(r.Context(), selectedDomain.ID, start, end, 5)
		if err != nil {
			log.Printf("Failed to get top pages: %v", err)
		} else {
			stats.TopPages = topPages
		}

		// device usage
		deviceUsage, err := f.ps.GetDeviceUsage(r.Context(), selectedDomain.ID, start, end)
		if err != nil {
			log.Printf("Failed to get device usage: %v", err)
		} else {
			stats.DeviceUsage = deviceUsage
		}
	}

	data := PageData{
		Title:   "Dashboard",
		User:    user,
		Stats:   stats,
		Slug:    "dashboard",
		Domains: domains,
	}

	if err := tmpl.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		http.Error(w, "Failed to render analytics page", http.StatusInternalServerError)
		log.Printf("Template execution error: %v", err)
	}
}

func (f *Frontend) realtime(w http.ResponseWriter, r *http.Request) {
	user := f.userFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	domains, selectedDomain, err := f.getDomainsAndSelected(r, user)
	if err != nil {
		log.Printf("Failed to list domains: %v", err)
	}

	data := PageData{
		Title:   "Real-time",
		User:    user,
		Slug:    "realtime",
		Domains: domains,
		Stats: &DashboardStats{
			SelectedDomain: selectedDomain,
		},
	}

	if err := tmpl.ExecuteTemplate(w, "realtime.html", data); err != nil {
		http.Error(w, "Failed to render realtime page", http.StatusInternalServerError)
	}
}

func (f *Frontend) domains(w http.ResponseWriter, r *http.Request) {
	user := f.userFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// POST create domain
	if r.Method == http.MethodPost {
		// Create new domain
		name := r.FormValue("name")
		if name == "" {
			http.Error(w, "Domain name is required", http.StatusBadRequest)
			return
		}

		domain := &domain.Domain{
			ID:                id.NewID(),
			Name:              name,
			UserID:            user.ID,
			VerificationToken: id.NewID(),
			Verified:          false,
		}

		_, err := f.db.DomainStorage().CreateDomain(r.Context(), domain)
		if err != nil {
			log.Printf("Failed to create domain: %v", err)
			http.Error(w, "Failed to create domain", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/domains", http.StatusSeeOther)
		return
	}

	// GET list domains
	domains, selectedDomain, err := f.getDomainsAndSelected(r, user)
	if err != nil {
		log.Printf("Failed to list domains: %v", err)
		http.Error(w, "Failed to list domains", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Title:   "Domains",
		User:    user,
		Domains: domains,
		Slug:    "domains",
		Stats: &DashboardStats{
			SelectedDomain: selectedDomain,
		},
	}

	if err := tmpl.ExecuteTemplate(w, "domains.html", data); err != nil {
		http.Error(w, "Failed to render domains page", http.StatusInternalServerError)
	}
}

func (f *Frontend) verifyDomain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	domainID := r.FormValue("domain_id")
	if domainID == "" {
		http.Error(w, "Domain ID is required", http.StatusBadRequest)
		return
	}

	d, err := f.db.DomainStorage().ReadDomain(r.Context(), domainID)
	if err != nil {
		log.Printf("Failed to read domain: %v", err)
		http.Error(w, "Domain not found", http.StatusNotFound)
		return
	}

	// verify ownership by checking for the file
	verificationURL := "https://" + d.Name + "/updog_" + d.VerificationToken + ".txt"
	client := &http.Client{}
	req, err := http.NewRequest("GET", verificationURL, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Set a realistic User-Agent and Accept header
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MyApp/1.0)")
	req.Header.Set("Accept", "text/plain")

	resp, err := client.Do(req)
	// resp, err := http.Get(verificationURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Verification failed for %s %d: %v", d.Name, resp.StatusCode, err)
		http.Error(w, "Verification failed. Please ensure the file exists at: "+verificationURL, http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	// update domain as verified
	d.Verified = true
	_, err = f.db.Db.NewUpdate().
		Model(d).
		Column("verified").
		Where("id = ?", d.ID).
		Exec(r.Context())

	if err != nil {
		log.Printf("Failed to update domain: %v", err)
		http.Error(w, "Failed to verify domain", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/domains", http.StatusSeeOther)
}

func (f *Frontend) settings(w http.ResponseWriter, r *http.Request) {
	user := f.userFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	domains, selectedDomain, err := f.getDomainsAndSelected(r, user)
	if err != nil {
		log.Printf("Failed to list domains: %v", err)
	}

	data := PageData{
		Title:   "Settings",
		User:    user,
		Slug:    "settings",
		Domains: domains,
		Stats: &DashboardStats{
			SelectedDomain: selectedDomain,
		},
	}

	// POST update settings
	if r.Method == http.MethodPost {
		disableSignups := r.FormValue(settings.SettingDisableSignups) == "on"
		if err := f.db.SetValueAsBool(r.Context(), settings.SettingDisableSignups, disableSignups); err != nil {
			log.Printf("Failed to update settings: %v", err)
			http.Error(w, "Failed to update settings", http.StatusInternalServerError)
			return
		}

		// redirect to refresh the page and show updated state
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
		return
	}

	// GET settings
	disableSignups, err := f.db.ReadValueAsBool(r.Context(), settings.SettingDisableSignups)
	if err != nil {
		log.Printf("Failed to read settings: %v", err)
		data.Error = "Failed to load settings"
	}

	data.Data = map[string]any{
		"DisableSignups": disableSignups,
	}

	if err := tmpl.ExecuteTemplate(w, "settings.html", data); err != nil {
		http.Error(w, "Failed to render settings page", http.StatusInternalServerError)
	}
}

func (f *Frontend) visitors(w http.ResponseWriter, r *http.Request) {
	user := f.userFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	domains, selectedDomain, err := f.getDomainsAndSelected(r, user)
	if err != nil {
		log.Printf("Failed to list domains: %v", err)
	}

	data := PageData{
		Title:   "Visitors",
		User:    user,
		Slug:    "visitors",
		Domains: domains,
		Stats: &DashboardStats{
			SelectedDomain: selectedDomain,
		},
		AdditionalStyles: []string{
			"https://unpkg.com/leaflet@1.9.4/dist/leaflet.css",
		},
		AdditionalScripts: []string{
			"https://unpkg.com/leaflet@1.9.4/dist/leaflet.js",
		},
	}

	if err := tmpl.ExecuteTemplate(w, "visitors.html", data); err != nil {
		http.Error(w, "Failed to render visitors page", http.StatusInternalServerError)
	}
}

func (f *Frontend) pages(w http.ResponseWriter, r *http.Request) {
	user := f.userFromRequest(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	domains, selectedDomain, err := f.getDomainsAndSelected(r, user)
	if err != nil {
		log.Printf("Failed to list domains: %v", err)
	}

	stats := &DashboardStats{
		SelectedDomain: selectedDomain,
	}

	if selectedDomain != nil {
		start, end, _ := httpx.ParseTimeParams(r)
		// Get top 100 pages
		topPages, err := f.ps.GetTopPages(r.Context(), selectedDomain.ID, start, end, 100)
		if err != nil {
			log.Printf("Failed to get top pages: %v", err)
		} else {
			stats.TopPages = topPages
		}
	}

	data := PageData{
		Title:   "Pages",
		User:    user,
		Slug:    "pages",
		Domains: domains,
		Stats:   stats,
	}

	if err := tmpl.ExecuteTemplate(w, "pages.html", data); err != nil {
		http.Error(w, "Failed to render pages page", http.StatusInternalServerError)
	}
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

func (f *Frontend) logout(w http.ResponseWriter, r *http.Request) {
	// clear the authentication cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		HttpOnly: true,
		Secure:   !env.IsDev(),
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (f *Frontend) login(w http.ResponseWriter, r *http.Request) {

	data := PageData{
		Title: "Login",
	}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		// find user by email
		user, err := f.db.UserStorage().ReadUserByEmail(r.Context(), email)
		if err != nil {
			data.Error = "Invalid email or password"
			tmpl.ExecuteTemplate(w, "login.html", data)
			return
		}

		// validate password
		if user.EncryptedPassword == "" || !user.Validate(password) {
			data.Error = "Invalid email or password"
			tmpl.ExecuteTemplate(w, "login.html", data)
			return
		}

		// user is validated, create a token
		token, _, err := f.auth.CreateToken(user.ID)
		if err != nil {
			data.Error = "Internal error"
			tmpl.ExecuteTemplate(w, "login.html", data)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   !env.IsDev(),
		})

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	tmpl.ExecuteTemplate(w, "login.html", data)
}

func (f *Frontend) join(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Sign Up",
	}

	disableSignups, err := f.db.ReadValueAsBool(r.Context(), settings.SettingDisableSignups)

	if err != nil {
		log.Printf("Failed to read settings: %v", err)
		data.Error = "Failed to load settings"
	}

	if disableSignups {
		data.Error = "Signups are currently disabled"
	}

	if data.Error != "" {
		tmpl.ExecuteTemplate(w, "signup.html", data)
		return
	}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		if email == "" || password == "" {
			data.Error = "Failed to create user."
			tmpl.ExecuteTemplate(w, "signup.html", data)
			return
		}

		epass, err := user.HashPassword(password)
		if err != nil {
			log.Printf("Failed to hash password: %v", err)
			data.Error = "Sorry! An internal error occurred."
			tmpl.ExecuteTemplate(w, "signup.html", data)
			return
		}

		u := &user.User{
			ID:                id.NewID(),
			Email:             email,
			Name:              user.NameFromEmail(email),
			Initials:          user.InitialsFromEmail(email),
			EncryptedPassword: epass,
		}

		if err := f.db.CreateUser(r.Context(), u); err != nil {
			log.Printf("Failed to create user: %v", err)
			data.Error = "Failed to create user."
			tmpl.ExecuteTemplate(w, "signup.html", data)
			return
		}

		// Auto login after signup
		token, _, err := f.auth.CreateToken(u.ID)
		if err == nil {
			http.SetCookie(w, &http.Cookie{
				Name:     "token",
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				Secure:   !env.IsDev(),
			})
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tmpl.ExecuteTemplate(w, "signup.html", data)
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

func (f *Frontend) userFromRequest(r *http.Request) *user.User {
	if u, ok := r.Context().Value(contextKeyUser).(*user.User); ok {
		return u
	}
	return nil
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
