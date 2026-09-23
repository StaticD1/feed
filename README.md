# Feed

A small educational web application written in Go.

The interface supports `en-US` and `ru-RU`. The language selector saves your
choice in a cookie for 365 days. Without a supported cookie value, Feed uses
`Accept-Language`, falling back to `en-US`.

Post creation times are saved automatically in PostgreSQL and displayed in UTC,
with an explicit `UTC` label. Dates use the selected locale, for example
`Wednesday, September 23, 2026, 2:30 PM UTC` or
`Среда, 23 сентября 2026 г., 14:30 UTC`. Changing the language changes the
format, not the time zone.

On startup, the application adds the `created_at` column if it is missing,
preserving existing posts. Older posts display `Creation time unknown` because
their original creation times were not recorded.

Translations are embedded from `internal/localization/locales/*.json` at build
time. After editing a translation, rebuild and restart the application (or
restart `go run`). HTML templates still load from `templates/` on each request;
run the application from the repository root.


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
