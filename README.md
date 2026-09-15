# Feed

A small educational web application written in Go.

## Run

Create an empty PostgreSQL database, then run the application from the project
root:

```sh
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/feed?sslmode=disable'
go run ./cmd/app
```

The server listens on <http://localhost:8080>. Database tables are created when
the application starts.
