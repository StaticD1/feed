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
	tmpl, err := template.ParseFiles("templates/register.html")
	if err != nil {
		http.Error(w, "could not load page", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "could not render page", http.StatusInternalServerError)
	}
}

func (app *application) register(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" || password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "could not create user", http.StatusInternalServerError)
		return
	}

	if _, err := app.db.Exec(
		"INSERT INTO users (username, password_hash) VALUES ($1, $2)",
		username,
		string(hash),
	); err != nil {
		http.Error(w, "could not create user", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
