# Phase 32 — Configurable Port Mapping

Builds on phase 16 (Dockerize and Postgres).

## Goal

The `PORT` and a new `METABASE_PORT` env var should control the host-side port mapping in Docker Compose. Internal container ports stay fixed (app: 8080, Metabase: 3000). Changing these env vars is all that's needed to deploy on different ports.

## Scope

### In scope

- `docker-compose.yml` host port mappings use env vars
- New `METABASE_PORT` env var with default `3000`
- `.env.example` updated with `METABASE_PORT`
- `README.md` documents all env vars and their defaults

### Out of scope

- Internal container ports (remain hardcoded)
- Vite dev proxy (local dev only, not Docker deployment)
- Dockerfile `EXPOSE` (documentation-only directive, irrelevant at runtime)

## Changes

### `docker-compose.yml`

| Service  | Current         | New                               |
|----------|-----------------|-----------------------------------|
| app      | `"8080:8080"`   | `"${PORT:-8080}:8080"`            |
| metabase | `"3000:3000"`   | `"${METABASE_PORT:-3000}:3000"`   |

No other lines change. Internal container ports, healthchecks, and inter-container URLs (e.g., `http://metabase:3000` in the app proxy) are unaffected because they use the Docker network, not host ports.

### `.env.example`

Add `METABASE_PORT=3000` alongside existing `PORT=8080`.

### `README.md`

Add an **Environment Variables** section listing every `.env` variable and its default:

| Variable                 | Default                          | Description                        |
|--------------------------|----------------------------------|------------------------------------|
| `PORT`                   | `8080`                           | Host port for the app              |
| `METABASE_PORT`          | `3000`                           | Host port for Metabase             |
| `CLIENT_ID`              | *(required)*                     | Spotify OAuth client ID            |
| `CLIENT_SECRET`          | *(required)*                     | Spotify OAuth client secret        |
| `SPOTIFY_ALLOWED_USER_ID`| *(required)*                     | Spotify user ID allowed to auth    |
| `DATABASE_URL`           | `postgres://...localhost:5432/…` | Postgres connection string         |
| `MB_ADMIN_EMAIL`         | `admin@listening-log.local`      | Metabase initial admin email       |
| `MB_ADMIN_PASSWORD`      | `changeme`                       | Metabase initial admin password    |
| `MB_ADMIN_FIRST_NAME`    | `Admin`                          | Metabase admin first name          |
| `MB_ADMIN_LAST_NAME`     | `User`                           | Metabase admin last name           |

Update the existing "App runs on `localhost:8080`, Metabase on `localhost:3000`" line to reference the env vars.

## Definition of done

- [ ] `PORT=9090 docker compose up` exposes the app on host port 9090
- [ ] `METABASE_PORT=4000 docker compose up` exposes Metabase on host port 4000
- [ ] Omitting both vars defaults to 8080 / 3000 (existing behavior)
- [ ] App's Metabase proxy (`/metabase/*`) still works regardless of host port
- [ ] Spotify OAuth flow still works regardless of host port
- [ ] `.env.example` lists `PORT` and `METABASE_PORT`
- [ ] `README.md` has a complete env var reference table
