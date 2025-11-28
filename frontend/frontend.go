package frontend

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"

	"github.com/zackb/updog/auth"
	"github.com/zackb/updog/db"
	"github.com/zackb/updog/env"
	"github.com/zackb/updog/user"
)

const (
	contextKeyUser = "user"
)

//go:embed views/*.html
var viewsFS embed.FS

//go:embed public/img/* public/script/* public/style/*
var staticFS embed.FS

var tmpl = template.Must(template.ParseFS(viewsFS, "views/*.html"))

var staticHandler http.Handler

type Frontend struct {
	auth          *auth.Service
	db            *db.DB
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
		staticHandler: staticHandler,
	}, nil
}

func (f *Frontend) Routes(mux *http.ServeMux) {

	// serve static files from the public directory
	mux.Handle("/static/", http.StripPrefix("/static/", staticHandler))

	mux.HandleFunc("/logout", f.logout)
	mux.HandleFunc("/join", f.join)
	mux.HandleFunc("/dashboard", f.dashboard)
	mux.HandleFunc("/", f.index)
}

func (f *Frontend) dashboard(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.ExecuteTemplate(w, "dashboard.html", nil); err != nil {
		http.Error(w, "Failed to render analytics page", http.StatusInternalServerError)
	}
}

func (f *Frontend) index(w http.ResponseWriter, r *http.Request) {

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

func (f *Frontend) join(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		if email == "" || password == "" {
			http.Error(w, "Email and password are required", http.StatusBadRequest)
			return
		}

		epass, err := user.HashPassword(password)
		if err != nil {
			log.Printf("Failed to hash password: %v", err)
			http.Error(w, "Sorry! An internal error occurred", http.StatusInternalServerError)
			return
		}

		user := &user.User{
			Email:             email,
			EncryptedPassword: epass,
		}

		if err := f.db.CreateUser(r.Context(), user); err != nil {
			log.Printf("Failed to create user: %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tmpl.ExecuteTemplate(w, "join.html", nil)
}

func initTemplatesAndStatic() error {
	if env.IsDev() {
		// dev mode from disk
		tmpl = template.Must(template.ParseGlob(filepath.Join("frontend", "views", "*.html")))

		// serve static files from disk
		staticHandler = http.FileServer(http.Dir(filepath.Join("frontend", "public")))
	} else {
		// production from embedded FS
		tmpl = template.Must(template.ParseFS(viewsFS, "views/*.html"))

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
