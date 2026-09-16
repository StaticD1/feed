package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"html/template"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func (app *application) loginPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		http.Error(w, "could not load page", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "could not render page", http.StatusInternalServerError)
	}
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" || password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	var userID int64
	var passwordHash string
	err := app.db.QueryRow(
		"SELECT id, password_hash FROM users WHERE username = $1",
		username,
	).Scan(&userID, &passwordHash)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "could not log in", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}
	token := hex.EncodeToString(tokenBytes)

	app.sessionsMu.Lock()
	app.sessions[token] = userID
	app.sessionsMu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) getCurrentUser(r *http.Request) (int64, bool) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return 0, false
	}

	app.sessionsMu.RLock()
	userID, ok := app.sessions[cookie.Value]
	app.sessionsMu.RUnlock()
	return userID, ok
}

// TODO: add logout if the application ever needs it.
