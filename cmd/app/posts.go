package main

import (
	"database/sql"
	"html/template"
	"net/http"
)

type post struct {
	Username  string
	Body      string
	CreatedAt sql.NullTime
}

func createPostsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id),
			body TEXT NOT NULL
		);
		-- Keep old posts without a date: their creation time is unknown.
		ALTER TABLE posts ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ;
		ALTER TABLE posts ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP;
	`)
	return err
}

func (app *application) feedPage(w http.ResponseWriter, r *http.Request) {
	rows, err := app.db.Query(`
		SELECT users.username, posts.body, posts.created_at
		FROM posts
		JOIN users ON users.id = posts.user_id
		ORDER BY posts.id DESC
	`)
	if err != nil {
		http.Error(w, "could not load posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []post
	for rows.Next() {
		var post post
		if err := rows.Scan(&post.Username, &post.Body, &post.CreatedAt); err != nil {
			http.Error(w, "could not load posts", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "could not load posts", http.StatusInternalServerError)
		return
	}

	_, loggedIn := app.getCurrentUser(r)
	data := struct {
		LoggedIn bool
		Posts    []post
	}{
		LoggedIn: loggedIn,
		Posts:    posts,
	}

	tmpl, err := template.ParseFiles("templates/feed.html")
	if err != nil {
		http.Error(w, "could not load page", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "could not render page", http.StatusInternalServerError)
	}
}

func (app *application) createPost(w http.ResponseWriter, r *http.Request) {
	userID, ok := app.getCurrentUser(r)
	if !ok {
		http.Error(w, "log in to create a post", http.StatusUnauthorized)
		return
	}

	body := r.FormValue("body")
	if body == "" {
		http.Error(w, "post cannot be empty", http.StatusBadRequest)
		return
	}

	// TODO: add a post length limit
	if _, err := app.db.Exec(
		"INSERT INTO posts (user_id, body) VALUES ($1, $2)",
		userID,
		body,
	); err != nil {
		http.Error(w, "could not create post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
