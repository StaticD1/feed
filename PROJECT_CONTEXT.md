# Project context for coding agents

## Project purpose

Feed is an educational web application that grows through homework assignments.
Visitors read a public text feed; registered users log in and publish posts.
The interface supports English and Russian. The project favors an implementation
that a student can follow through a few Go files.

This document describes the current implementation and where its parts connect.
It is a reference for understanding and changing the project, not a backlog.
[AGENTS.md](AGENTS.md) contains working instructions;
[README.md](README.md) contains setup and usage;
[CHANGELOG.md](CHANGELOG.md) records releases.

## Architecture and code map

The module is `github.com/StaticD1/feed`, with Go 1.25.5 declared in
[go.mod](go.mod). There is one application process and one PostgreSQL database.
The server uses `net/http`, `html/template`, and `database/sql` with the pgx
driver. There is no separate frontend application, JSON API, ORM, or service layer.

| File or directory | Responsibility and entry points |
| --- | --- |
| [cmd/app/main.go](cmd/app/main.go) | `application` state, startup, route registration, HTTP server |
| [cmd/app/users.go](cmd/app/users.go) | `createUsersTable`, registration page and submission |
| [cmd/app/auth.go](cmd/app/auth.go) | Login page, `login`, and `getCurrentUser` |
| [cmd/app/posts.go](cmd/app/posts.go) | `createPostsTable`, `post` view model, `feedPage`, and `createPost` |
| [cmd/app/localization.go](cmd/app/localization.go) | `withLocale`, request context, `pageData`, and `changeLocale` |
| [internal/localization/localization.go](internal/localization/localization.go) | Embedded catalogs, supported locales, language matching, and regional formatters |
| [internal/localization/locales](internal/localization/locales) | English and Russian JSON message catalogs |
| [templates](templates) | Feed, login, registration, and shared language selector templates |
| [compose.yaml](compose.yaml) | PostgreSQL 16 Alpine service, port mapping, and persistent volume |
| [.env.example](.env.example) | Example `DATABASE_URL` for the local database |

Handlers are methods of `application`. They validate form input, query the
database, and render templates or redirect directly. The shared application state
consists of a database handle, a translation catalog, and a session map protected
by a `sync.RWMutex`.

## Startup and request lifecycle

Startup in `main` proceeds in this order:

1. Read `DATABASE_URL`; exit if it is empty.
2. Load the embedded translation catalogs.
3. Open and ping PostgreSQL.
4. Create the users table, then prepare the posts table and timestamp column.
5. Initialize the empty session map and register routes.
6. Serve `app.withLocale(mux)` on `:8080`.

For each request, `withLocale` selects a localizer and puts it in the request
context. It sets `Content-Language` and adds `Accept-Language` and `Cookie` to
`Vary`. Handlers retrieve that localizer through `requestLocalizer`, which
assumes the middleware has already run.

HTML handlers create `pageData` with the localizer, the current URL path as
`ReturnTo`, and UTC as the time zone. Each page parses its own template together
with `templates/language-switcher.html`. Successful form submissions use
HTTP 303 redirects; errors use localized text through `http.Error`.

## User and post flows

### Registration and login

`POST /register` reads `username` and `password`, rejects empty strings, hashes
the password with bcrypt's default cost, and inserts the user. The database
enforces username uniqueness. Success redirects to `/login`; registration does
not create a session. Insert errors currently share a generic HTTP 400 response.

`POST /login` looks up the user by username and compares the password with the
stored bcrypt hash. Invalid credentials return HTTP 401. A successful login
generates 32 random bytes, hex-encodes them, and stores the token-to-user-ID mapping
in `application.sessions`. The response sets a `session` cookie with `Path=/`,
`HttpOnly`, and `SameSite=Lax`, then redirects to the feed.

`getCurrentUser` reads the cookie and checks the map under a read lock. The cookie
has no explicit expiry or `Secure` flag. The server has no session expiry,
revocation, or logout mechanism. Restarting the process invalidates all sessions;
users and posts remain in PostgreSQL.

### Reading and writing posts

`GET /` joins posts to users and loads username, body, and timestamp. Ordering is
`posts.id DESC`, not timestamp order. All rows are loaded without pagination.
The `post` type is a rendering model; it does not contain the database IDs.

The feed is public. A valid session controls whether the post form is displayed.
`POST /posts` also checks authentication on the server and returns HTTP 401 for
a missing or unknown session. The author ID comes from that session. The handler
rejects an empty body, inserts the post with parameterized SQL, and redirects
to `/`.

Current validation checks exact empty strings: it does not trim whitespace or
apply a post length limit. Templates render post bodies as escaped text.

## Database model and schema evolution

| Table | Columns and constraints |
| --- | --- |
| `users` | `id BIGSERIAL PRIMARY KEY`; `username TEXT UNIQUE NOT NULL`; `password_hash TEXT NOT NULL` |
| `posts` | `id BIGSERIAL PRIMARY KEY`; `user_id BIGINT NOT NULL REFERENCES users(id)`; `body TEXT NOT NULL`; nullable `created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP` |

Schema setup runs inside `createUsersTable` and `createPostsTable` on every
startup. There is no migration directory or version table. The database itself
must already exist, and the connection role must be able to create and alter
tables.

The timestamp upgrade deliberately adds a nullable column before setting its
default. This leaves older rows with `NULL` because their original creation times
are unknown. New inserts omit the column and use the database default.
`sql.NullTime` carries that distinction into the template, which displays either
the formatted UTC time or the localized unknown-time message.

An application restart preserves database records but clears sessions.
`docker compose down` preserves the database volume; adding `-v` removes it.

## Localization and rendering

The supported locales are `en-US` and `ru-RU`; the first entry in
`supportedLocales` is the fallback. Locale selection uses a supported cookie
value first, then a sufficiently confident match from `Accept-Language`, then
English. Explicit choices use `Lookup`, which accepts supported tags with
equivalent casing, such as `en-us`.

`POST /locale` validates the submitted locale and stores the canonical tag in a
365-day cookie. Its redirect target is restricted to `/`, `/login`, and
`/register`; any other value falls back to `/`. The locale cookie uses
`HttpOnly`, `SameSite=Lax`, and `Secure` when `r.TLS` is non-nil.

`go-i18n/v2` translates message keys from both JSON catalogs. `monday` formats
dates and times, while `golang.org/x/text` provides language matching and number
formatting. `Localizer.T` accepts optional named template parameters;
`DateTime` combines the date, time, and zone using the `datetime.full` message.
The `Integer` and `Decimal` helpers exist but are not used by the current pages.

Publication times are displayed in UTC regardless of language. The feed also
includes a machine-readable timestamp in the HTML `time` element. Usernames and
post bodies are not translated.

Catalogs are embedded at build time and require rebuilding and restarting after
changes. Templates are read from disk on every request, relative to the process
working directory. Running from the repository root is therefore required.

## Change impact and verification

| Change | Connected areas to inspect |
| --- | --- |
| New page or route | Route registration, locale middleware, page data, template, and both message catalogs |
| New interface message | Matching keys and placeholders in both catalogs, plus Go or template call sites |
| Authentication change | Login, cookie settings, synchronized session access, feed form visibility, and post authorization |
| Post or schema change | Startup SQL, insert/query statements, rendering model, template, and existing database rows |
| Date or locale change | Locale matching, formatter, nullable timestamps, UTC display, and language selector |
| Configuration change | Startup, Compose, example environment, README, and this context |

There are currently no automated test files or CI configuration. The Go commands
in AGENTS check compilation and static issues; they do not prove HTTP or database
behavior. Relevant manual scenarios include:

- Registration, duplicate usernames, valid and invalid login, unauthenticated post
  submission, and session loss after restart.
- Post submission, empty input rejection, feed order, escaped user content, and
  persistence after restarting the application.
- Both interface languages, cookie precedence, header fallback, invalid locale
  rejection, restricted return URLs, and dates that still show UTC.
- Fresh schema setup and upgrades with existing records, including posts whose
  creation time is unknown.

Use a disposable database for checks that create or alter records. Startup
instructions are in README. Compose starts only PostgreSQL, the application does
not load `.env`, and the HTTP address is fixed at `:8080`.

## Current boundaries

The implementation has no post editing or deletion, pagination, password recovery,
email verification, password strength policy, dedicated CSRF protection, request
rate limiting, or application-managed HTTPS. Logout and a post length limit are
explicit TODOs. These gaps describe the current homework scope and do not imply
that an agent should implement them during an unrelated task.

## Documentation references

The separation between brief working instructions and this project reference
follows [OpenAI's Codex best practices](https://learn.chatgpt.com/guides/best-practices)
and [AGENTS.md guidance](https://learn.chatgpt.com/docs/agent-configuration/agents-md).
The [OpenAI community discussion on Markdown practices](https://community.openai.com/t/what-are-you-md-files-best-practices/1386098/2)
also discusses keeping durable instructions separate from ordinary project
documentation. Community advice is not an official requirement. These sources
were reviewed on September 30, 2026.
