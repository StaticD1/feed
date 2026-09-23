# Feed

A small educational web application written in Go.

Post creation times are saved automatically in PostgreSQL and displayed in UTC
(for example, `Wed, 23.09.2026 14:30 UTC`). The display format is defined in
`templates/feed.html`.

On startup, the application adds the `created_at` column if it is missing,
preserving existing posts. Older posts display `Creation time unknown` because
their original creation times were not recorded.

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
