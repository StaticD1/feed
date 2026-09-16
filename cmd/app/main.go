package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type application struct {
	db         *sql.DB
	sessions   map[string]int64
	sessionsMu sync.RWMutex
}

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	if err := createUsersTable(db); err != nil {
		log.Fatal(err)
	}

	app := &application{
		db:       db,
		sessions: make(map[string]int64),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /register", app.registerPage)
	mux.HandleFunc("POST /register", app.register)
	mux.HandleFunc("GET /login", app.loginPage)
	mux.HandleFunc("POST /login", app.login)

	log.Println("server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
