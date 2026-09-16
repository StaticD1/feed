# Feed

A small educational web application written in Go.

## Run

Start PostgreSQL:

```sh
docker compose up -d postgres
```

Run the application:

```sh
DATABASE_URL='postgres://feed:feed@localhost:5432/feed?sslmode=disable' go run ./cmd/app
```

Stop PostgreSQL:

```sh
docker compose down
```
