package main

import (
	"database/sql"
	"html/template"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func createUsersTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL
		)
	`)
	return err
}

func (app *application) registerPage(w http.ResponseWriter, r *http.Request) {
	data := newPageData(r)
	tmpl, err := template.ParseFiles("templates/register.html", "templates/language-switcher.html")
	if err != nil {
		http.Error(w, data.I18n.T("errors.load_page"), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, data.I18n.T("errors.render_page"), http.StatusInternalServerError)
	}
}

func (app *application) register(w http.ResponseWriter, r *http.Request) {
	localizer := requestLocalizer(r)
	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" || password == "" {
		http.Error(w, localizer.T("errors.credentials_required"), http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, localizer.T("errors.create_user"), http.StatusInternalServerError)
		return
	}

	if _, err := app.db.Exec(
		"INSERT INTO users (username, password_hash) VALUES ($1, $2)",
		username,
		string(hash),
	); err != nil {
		http.Error(w, localizer.T("errors.create_user"), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
