# Phase 16 — Dockerize & Switch to PostgreSQL

## Goal

Containerize the application with Docker and replace the SQLite database with PostgreSQL. The result is a single `docker compose up` that runs the Go server (with built client) and a Postgres instance.

## Scope

### In scope
- Multi-stage `Dockerfile` (build client → build server → minimal runtime image)
- `docker-compose.yml` with app + Postgres services
- `.dockerignore`
- Swap `modernc.org/sqlite` for `github.com/jackc/pgx/v5` (most actively maintained Go Postgres driver)
- Rewrite `schema.sql` for Postgres syntax
- Update all `db/*.go` queries (`?` → `$N` placeholders, SQLite functions → Postgres equivalents)
- Update `config.go` to accept a `DATABASE_URL` connection string instead of a file path
- Update `db.go` to connect via connection string and run schema with Postgres
- Remove `data/` directory references, `db_backup.sqlite`, and SQLite-related `.gitignore` entries
- Remove `godotenv` dependency (env vars come from docker-compose `env_file`)
- Update `.env.example` with new vars

### Out of scope
- Data migration from SQLite to Postgres
- Dev-mode hot-reload / volume-mount setup
- CI/CD pipeline changes
- Health-check wait logic beyond Postgres `depends_on` with `condition: service_healthy`

## Data model

Rewrite `server/db/schema.sql` for Postgres:

```sql
CREATE TABLE IF NOT EXISTS spotify_auth (
    id            INTEGER PRIMARY KEY CHECK (id = 1),
    access_token  TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    scope         TEXT NOT NULL,
    expiry        BIGINT NOT NULL,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO spotify_auth (id, access_token, refresh_token, scope, expiry)
VALUES (1, '', '', '', 0)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS playback_log (
    id              SERIAL PRIMARY KEY,
    polled_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    track_id        TEXT NOT NULL,
    progress_ms     INTEGER NOT NULL,
    duration_ms     INTEGER NOT NULL,
    is_playing      BOOLEAN NOT NULL,
    popularity      INTEGER NOT NULL,
    device_name     TEXT NOT NULL,
    device_type     TEXT NOT NULL,
    shuffle_state   BOOLEAN NOT NULL,
    repeat_state    TEXT NOT NULL,
    context_uri     TEXT
);

CREATE INDEX IF NOT EXISTS idx_playback_log_polled_at ON playback_log(polled_at);
CREATE INDEX IF NOT EXISTS idx_playback_log_track_id ON playback_log(track_id);

CREATE TABLE IF NOT EXISTS artist (
    spotify_id   TEXT PRIMARY KEY,
    name         TEXT NOT NULL,
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS album (
    spotify_id    TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    album_type    TEXT,
    total_tracks  INTEGER,
    release_date  TEXT,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS album_image (
    id             SERIAL PRIMARY KEY,
    album_id       TEXT NOT NULL REFERENCES album(spotify_id),
    url            TEXT NOT NULL,
    width          INTEGER,
    height         INTEGER
);

CREATE INDEX IF NOT EXISTS idx_album_image_album_id ON album_image(album_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_album_image_unique ON album_image(album_id, url);

CREATE TABLE IF NOT EXISTS track (
    spotify_id    TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    album_id      TEXT NOT NULL REFERENCES album(spotify_id),
    artist_id     TEXT NOT NULL REFERENCES artist(spotify_id),
    duration_ms   INTEGER NOT NULL,
    track_number  INTEGER,
    disc_number   INTEGER,
    explicit      BOOLEAN,
    isrc          TEXT,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_track_album_id ON track(album_id);
CREATE INDEX IF NOT EXISTS idx_track_artist_id ON track(artist_id);

CREATE TABLE IF NOT EXISTS song_suggestion (
    id         SERIAL PRIMARY KEY,
    link       TEXT    NOT NULL DEFAULT '',
    message    TEXT    NOT NULL DEFAULT '',
    source     TEXT    NOT NULL DEFAULT 'home',
    ip_address TEXT    NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_song_suggestion_ip_created ON song_suggestion(ip_address, created_at);
```

Key changes from SQLite schema:
- `INTEGER PRIMARY KEY AUTOINCREMENT` → `SERIAL PRIMARY KEY`
- `DATETIME` → `TIMESTAMP`
- `INTEGER` booleans → `BOOLEAN` (`is_playing`, `shuffle_state`, `explicit`)
- `INSERT ... WHERE NOT EXISTS` seed → `INSERT ... ON CONFLICT DO NOTHING`
- `expiry INTEGER` → `expiry BIGINT` (unix timestamps)

## Backend changes

### `server/db/db.go`
- Replace `modernc.org/sqlite` import with `github.com/jackc/pgx/v5/stdlib`
- Accept a connection string (`DATABASE_URL`) instead of a file path
- Use `sql.Open("pgx", connStr)` instead of `sql.Open("sqlite", path)`
- Remove `os.MkdirAll` (no data directory needed)

### `server/db/auth.go`
- `?` → `$1, $2, ...` in all queries

### `server/db/playback.go`
- `?` → `$1` through `$10`
- Remove manual `bool → int` conversion for `is_playing` and `shuffle_state` — pass bools directly

### `server/db/metadata.go`
- `?` → `$N` in all queries
- `INSERT OR IGNORE INTO album_image` → `INSERT INTO album_image ... ON CONFLICT (album_id, url) DO NOTHING`
- Remove manual `bool → int` conversion for `explicit` — pass bool directly

### `server/db/suggestions.go`
- `?` → `$N` in all queries
- `datetime('now', '-1 hour')` → `NOW() - INTERVAL '1 hour'`

### `server/db/image_grid.go`
- `?` → `$1` for LIMIT parameter

### `server/db/now_playing.go`
- No changes expected (no parameterized queries, standard SQL only)

### `server/config/config.go`
- Remove `godotenv` import and `godotenv.Load()` call
- Replace `DatabasePath` field with `DatabaseURL string` (env: `DATABASE_URL`)
- Default `DATABASE_URL` to `postgres://listening_log:listening_log@localhost:5432/listening_log?sslmode=disable`
- Remove `ClientBaseURL` default of `5173` dev port — default to empty (prod doesn't need it set since SPA is served by Go)

### `server/go.mod`
- Remove `modernc.org/sqlite` (and its transitive deps)
- Remove `github.com/joho/godotenv`
- Add `github.com/jackc/pgx/v5`

### `server/main.go`
- Pass `cfg.DatabaseURL` to `db.Open()` instead of `cfg.DatabasePath`
- Fix SPA middleware path resolution for Docker: check `../client/dist` first, fall back to `client/dist` (resolves to `/client/dist` when CWD is `/` in the container)

## File structure (new files)

```
Dockerfile
docker-compose.yml
.dockerignore
```

### `Dockerfile`

Multi-stage build:

```dockerfile
# Stage 1: Build client
FROM node:22-alpine AS client-build
WORKDIR /app/client
COPY client/package.json client/package-lock.json ./
RUN npm ci
COPY client/ ./
RUN npm run build

# Stage 2: Build server
FROM golang:1.24-alpine AS server-build
WORKDIR /app/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
COPY --from=client-build /app/client/dist ../client/dist
RUN CGO_ENABLED=0 go build -o /server .

# Stage 3: Runtime
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=server-build /server /server
COPY --from=client-build /app/client/dist /client/dist
EXPOSE 8080
CMD ["/server"]
```

Note: The server's SPA middleware looks for `../client/dist` relative to the binary. Since the binary runs from `/`, the client dist at `/client/dist` matches `../client/dist` resolved from the working directory. Adjust the `spaMiddleware` path in `main.go` if needed — simplest fix is to make the static dir path configurable or use an absolute path `/client/dist`.

### `docker-compose.yml`

```yaml
services:
  db:
    image: postgres:17-alpine
    environment:
      POSTGRES_USER: listening_log
      POSTGRES_PASSWORD: listening_log
      POSTGRES_DB: listening_log
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "listening_log"]
      interval: 5s
      timeout: 3s
      retries: 5

  app:
    build: .
    ports:
      - "8080:8080"
    env_file: .env
    environment:
      DATABASE_URL: postgres://listening_log:listening_log@db:5432/listening_log?sslmode=disable
    depends_on:
      db:
        condition: service_healthy

volumes:
  pgdata:
```

The `env_file: .env` loads Spotify credentials and any overrides. The `DATABASE_URL` in `environment` overrides anything in `.env` to point at the compose Postgres service.

### `.dockerignore`

```
.git
.env
data/
client/node_modules/
client/dist/
server/server
db_backup.sqlite
*.md
spec/
```

### `.env.example` (update)

```
CLIENT_ID=
CLIENT_SECRET=
SPOTIFY_REDIRECT_URI=http://127.0.0.1:8080/api/auth/callback
CLIENT_BASE_URL=
PORT=8080
DATABASE_URL=postgres://listening_log:listening_log@localhost:5432/listening_log?sslmode=disable
SPOTIFY_ALLOWED_USER_ID=
```

## Cleanup

- Delete `server/data/` directory reference from code
- Delete `db_backup.sqlite` from repo root
- Remove `data/` from `.gitignore` (no longer relevant)
- Remove SQLite-related `.gitignore` entries (`database.db`)

## Definition of done

- [ ] SPA middleware serves the frontend correctly in Docker (path fallback works)
- [ ] `docker compose up --build` starts both containers and the app is accessible on `:8080`
- [ ] Postgres is the only database — no SQLite driver in `go.mod`
- [ ] Schema is applied automatically on startup (tables created in Postgres)
- [ ] Spotify OAuth flow works end-to-end through the containerized app
- [ ] Scraper polls and writes playback data to Postgres
- [ ] All existing API endpoints return data from Postgres
- [ ] Song suggestion rate limiting works with Postgres time functions
- [ ] `godotenv` removed — env vars come from docker-compose `env_file` / `environment`
- [ ] `.env.example` updated with `DATABASE_URL` replacing `DATABASE_PATH`
- [ ] No references to SQLite, `data/` directory, or `db_backup.sqlite` remain in code
