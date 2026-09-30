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
	data := newPageData(r)
	tmpl, err := template.ParseFiles("templates/login.html", "templates/language-switcher.html")
	if err != nil {
		http.Error(w, data.I18n.T("errors.load_page"), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, data.I18n.T("errors.render_page"), http.StatusInternalServerError)
	}
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	localizer := requestLocalizer(r)
	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" || password == "" {
		http.Error(w, localizer.T("errors.credentials_required"), http.StatusBadRequest)
		return
	}

	var userID int64
	var passwordHash string
	err := app.db.QueryRow(
		"SELECT id, password_hash FROM users WHERE username = $1",
		username,
	).Scan(&userID, &passwordHash)
	if err == sql.ErrNoRows {
		http.Error(w, localizer.T("errors.invalid_credentials"), http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, localizer.T("errors.login"), http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		http.Error(w, localizer.T("errors.invalid_credentials"), http.StatusUnauthorized)
		return
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		http.Error(w, localizer.T("errors.create_session"), http.StatusInternalServerError)
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

// TODO: add logout
