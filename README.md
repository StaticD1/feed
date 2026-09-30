# Feed

Feed is a small web application written in Go. Visitors can read a public feed,
and registered users can log in and publish text posts. The interface is available
in English and Russian.

This is a learning project, developed one homework assignment at a time. It keeps
the implementation straightforward: Go HTTP handlers, HTML templates, and SQL
queries against PostgreSQL.

## Features

- User registration with unique usernames and bcrypt password hashing.
- Login with cookie-based, in-memory sessions.
- A public feed showing the most recently inserted posts first.
- Post creation for authenticated users.
- Publication timestamps with localized date and time formatting.
- English (`en-US`) and Russian (`ru-RU`) translations with a language selector.
- Persistent storage for users and posts in PostgreSQL.
- Docker Compose configuration for the development database.

## Technology

| Component | Implementation |
| --- | --- |
| Application | Go, `net/http`, and `html/template` |
| Database | PostgreSQL 16 in Docker Compose |
| Database access | `database/sql` with the `pgx/v5` driver; SQL written directly in Go |
| Password hashing | `golang.org/x/crypto/bcrypt` |
| Translations | `go-i18n/v2` with embedded JSON catalogs |
| Regional formatting | `monday` for dates and times; `golang.org/x/text` for language matching and number formatting |

Dependency versions are recorded in [go.mod](go.mod). There is no frontend build
step or JavaScript package installation.

## Getting started

### Requirements

- Go 1.25.5 or later, as specified in `go.mod`.
- Docker with the Docker Compose plugin, or an existing PostgreSQL instance.
- Available ports `8080` for the application and `5432` for the Compose database.

### Start the database

From the repository root, run:

```sh
docker compose up -d postgres
docker compose exec postgres pg_isready -U feed -d feed
```

Wait until the second command reports that PostgreSQL is accepting connections.
Compose starts only PostgreSQL; the Go application runs separately on your host.

### Start the application

Run this command from the repository root in a POSIX shell, such as Bash:

```sh
DATABASE_URL='postgres://feed:feed@localhost:5432/feed?sslmode=disable' go run ./cmd/app
```

Open [http://localhost:8080](http://localhost:8080). Go downloads the required
modules on the first run. The application connects to PostgreSQL and prepares its
tables before starting the HTTP server.

To try the application:

1. Open [/register](http://localhost:8080/register) and create an account.
2. Log in on the page shown after registration.
3. Write a post on the feed and submit it.
4. Use the language selector to switch between English and Russian.

The feed is also readable without logging in. To view it as a guest, use a private
browser window. There is no logout button yet.

### Configuration

| Setting | Purpose |
| --- | --- |
| `DATABASE_URL` | Required PostgreSQL connection string. The application exits if it is unset or the database cannot be reached. |
| HTTP address | Fixed at `:8080` in `cmd/app/main.go`; it is not configurable through an environment variable. |

[.env.example](.env.example) contains the connection string for the Compose
database. The application does **not** load `.env` files automatically. Set
`DATABASE_URL` in the process environment, as in the command above. A local `.env`
file is ignored by Git, but creating one alone does not configure `go run`.

For an existing PostgreSQL instance, use its connection string instead. The
database must already exist, and the application user needs permission to create
and alter tables. The example credentials and disabled SSL are intended for local
development. The HTTP server listens on all interfaces on port `8080`.

### Stop the application and database

Press `Ctrl+C` in the application terminal, then run:

```sh
docker compose down
```

Database records remain in the named `feed_postgres_data` volume. Running
`docker compose down -v` deletes that volume and its users and posts. Stopping or
restarting the Go process clears all login sessions, so users must log in again.

## Project structure

```text
.
├── cmd/app/
│   ├── main.go                 # Startup, database connection, and routes
│   ├── users.go                # Users table and registration
│   ├── auth.go                 # Login and session lookup
│   ├── posts.go                # Posts table, public feed, and post creation
│   └── localization.go         # Locale middleware and language switching
├── internal/localization/
│   ├── localization.go        # Translation catalog and regional formatters
│   └── locales/
│       ├── en-US.json
│       └── ru-RU.json
├── templates/                 # Feed, authentication pages, and language selector
├── .env.example               # Example database connection string
├── compose.yaml               # PostgreSQL service and persistent volume
├── go.mod
├── go.sum
├── AGENTS.md                  # Repository instructions for coding agents
├── PROJECT_CONTEXT.md         # Architecture and implementation context for agents
├── CHANGELOG.md
└── README.md
```

## Routes

These routes serve HTML pages and process form submissions; there is no JSON API.

| Method | Path | Behavior |
| --- | --- | --- |
| `GET` | `/` | Show the public feed and, when logged in, the post form. |
| `GET` | `/register` | Show the registration form. |
| `POST` | `/register` | Create a user and redirect to `/login`. |
| `GET` | `/login` | Show the login form. |
| `POST` | `/login` | Create a session and redirect to `/`. |
| `POST` | `/posts` | Create a post and redirect to `/`; requires a valid session. |
| `POST` | `/locale` | Save the selected language and return to the originating page. |

## Data and sessions

The application creates two tables on startup:

- `users`: ID, unique username, and password hash.
- `posts`: ID, author ID referencing `users`, post body, and creation timestamp.

Schema setup lives in `createUsersTable` and `createPostsTable`; there is no
separate migration command. Startup also adds the `created_at` column if it is
missing. New posts receive a database-generated `TIMESTAMPTZ` value. Existing
posts without recorded timestamps retain `NULL` and display “Creation time
unknown” rather than an invented date.

Login creates a random session token and sends it in an `HttpOnly`, `SameSite=Lax`
cookie. The token-to-user mapping is stored in a mutex-protected map inside the
Go process. Sessions are not shared between application instances or persisted in
PostgreSQL.

## Localization

The language is selected in this order:

1. A supported value in the `locale` cookie.
2. A supported language matched from the browser's `Accept-Language` header.
3. English (`en-US`) as the fallback.

The language selector stores the choice for 365 days. Changing the language
changes labels, error messages, and date formatting; it does not translate user
posts or change the time zone. Publication times are always displayed in UTC
with an explicit `UTC` label.

Translations are embedded from `internal/localization/locales/*.json` at build
time. Update both catalogs when adding interface text, then rebuild and restart
the application, or restart `go run`. HTML templates are loaded from `templates/`
on each request, so template changes do not require rebuilding. The working
directory must still be the repository root.

## Development checks

Run from the repository root:

```sh
go test ./...
go vet ./...
go build -o bin/feed ./cmd/app
```

There are currently no automated test files. `go test ./...` checks that the
packages compile but does not verify the application's behavior. The checks above
do not require a running database. To run the built binary, set `DATABASE_URL` and
execute `./bin/feed` from the repository root so it can find the templates.

For a manual check, register a new user, log in, publish a post, and view the feed
in both languages. Check invalid login credentials and an empty post as well.
Restarting the application should preserve posts and require a new login.

## Current limitations

This project is intended for learning and local development. Its current scope
does not include:

- Logout, server-side session expiry, or persistent sessions.
- Post editing, deletion, pagination, or an application-defined post length limit.
- Password recovery, email verification, or a password strength policy.
- Dedicated CSRF protection or request rate limiting.
- Application-managed HTTPS; the session cookie does not currently set `Secure`.
- An automated test suite or a separate database migration system.

These are limits of the current implementation, not a schedule for future work.
See [CHANGELOG.md](CHANGELOG.md) for release history. Coding agents can use
[AGENTS.md](AGENTS.md) for working instructions and
[PROJECT_CONTEXT.md](PROJECT_CONTEXT.md) for architecture and implementation details.
