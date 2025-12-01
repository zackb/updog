package frontend

import (
	"log"
	"net/http"

	"github.com/zackb/updog/env"
	"github.com/zackb/updog/id"
	"github.com/zackb/updog/settings"
	"github.com/zackb/updog/user"
)

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
