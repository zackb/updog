package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/zackb/updog/env"
	"github.com/zackb/updog/httpx"
	"github.com/zackb/updog/user"
)

func (a *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&creds)
	if httpx.CheckError(w, err) {
		return
	}

	// find user by email
	user, err := a.db.UserStorage().ReadUserByEmail(r.Context(), creds.Email)
	if err != nil {
		httpx.InvalidCredentials(w)
		return
	}

	// validate password
	if user.EncryptedPassword == "" || !user.Validate(creds.Password) {
		httpx.InvalidCredentials(w)
		return
	}

	// user is validated, create a token
	token, _, err := a.auth.CreateToken(user.ID)
	if err != nil {
		httpx.InvalidCredentials(w)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   !env.IsDev(),
	})

	data := map[string]any{"token": token}
	muser := map[string]any{"id": user.ID, "email": user.Email}
	data["user"] = muser
	err = json.NewEncoder(w).Encode(data)
	httpx.CheckError(w, err)
}

func (a *API) handleLogout(w http.ResponseWriter, r *http.Request) {
	// clear the token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   !env.IsDev(),
	})

	data := map[string]string{"message": "Logged out successfully"}
	err := json.NewEncoder(w).Encode(data)
	httpx.CheckError(w, err)
}

func (a *API) handleRegister(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&creds)
	if httpx.CheckError(w, err) {
		return
	}
	userStore := a.db.UserStorage()
	if creds.Email == "" || creds.Password == "" {
		httpx.JSONError(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// check if user already exists
	_, err = userStore.ReadUserByEmail(ctx, creds.Email)
	if err == nil {
		httpx.JSONError(w, "User already exists", http.StatusBadRequest)
		return
	}

	epass, err := user.HashPassword(creds.Password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		httpx.JSONError(w, "Sorry! An internal error occurred", http.StatusInternalServerError)
		return
	}

	user := &user.User{
		Email:             creds.Email,
		EncryptedPassword: epass,
	}

	if err := userStore.CreateUser(ctx, user); err != nil {
		log.Printf("Failed to create user: %v", err)
		httpx.JSONError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	token, _, err := a.auth.CreateToken(user.ID)
	if err != nil {
		httpx.InvalidCredentials(w)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   !env.IsDev(),
	})

	data := map[string]any{"token": token}
	muser := map[string]any{"id": user.ID, "email": user.Email}
	data["user"] = muser
	err = json.NewEncoder(w).Encode(data)
	httpx.CheckError(w, err)
}

func (a *API) handleVerify(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	// check if the user is authenticated
	token := a.auth.IsAuthenticated(r)
	if token == nil {
		httpx.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := a.db.UserStorage().ReadUser(ctx, token.ClientId)
	if err != nil {
		httpx.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	data := map[string]any{
		"id":    user.ID,
		"email": user.Email,
	}

	err = json.NewEncoder(w).Encode(data)
	httpx.CheckError(w, err)
}
