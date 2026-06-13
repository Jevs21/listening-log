# Phase 1: Web App Scaffolding & Spotify Auth

> Builds on [`00-overview.md`](./00-overview.md). This phase delivers the
> foundation everything else sits on: a React client, a Go + Gin server, a
> SQLite database, and a complete Spotify OAuth flow that lands the user's
> tokens in the database for later phases (the scraper, analysis) to consume.
>
> **Scope: single user, minimal prototype.** One Spotify account, one row of
> credentials. No user accounts, no sessions, no JWT.

## 1. Goal

Open the web app, click "Connect Spotify", complete Spotify's consent, and be
returned to the app in a connected state. The server persists the Spotify
access + refresh tokens to `./data/database.sqlite` so the Phase 2 scraper can
read them and begin polling.

There is **no scraping and no listening-data UI** in this phase — only the
scaffolding and the auth handshake.

## 2. Scope

### In scope
- `./client` — Vite + TypeScript + React app (bare-minimum MVP).
- `./server` — Go + Gin HTTP server.
- `./data/database.sqlite` — SQLite database, created by the server at boot.
- Spotify **Authorization Code** OAuth flow, with the code→token exchange done
  **server-side** using the client secret.
- A single-row `spotify_auth` table holding the one user's tokens.
- A dev workflow where `pnpm dev` (client) and `./server` (server) run in two
  terminals, plus a prod path where the server serves the built client.

### Out of scope (deferred)
- The scraper / `now-playing` polling → **Phase 2**.
- Any listening-history pages, charts, "most played", ratings, etc.
- Multi-user, user accounts, login/sessions, JWT. "Connected" simply means
  valid Spotify tokens exist in the DB.
- Token **refresh** for calling the Spotify data API (the scraper owns that in
  Phase 2). Phase 1 only obtains and stores the initial token set.
- Docker / deployment packaging (the prototype's compose files are not carried
  forward).

## 3. Architecture

```
┌────────────────────┐         ┌─────────────────────────┐        ┌──────────────────┐
│  Client (React)    │  /api   │   Server (Go + Gin)     │        │  Spotify OAuth   │
│  Vite + TS         │ ───────▶│  - OAuth orchestration  │ ──────▶│  accounts.spotify│
│  :5173 (dev)       │◀─────── │  - token persistence    │◀────── │  api.spotify.com │
│                    │         │            │            │        └──────────────────┘
└────────────────────┘         │            ▼            │
                               │   ./data/database.sqlite│
                               └─────────────────────────┘
```

- **Dev:** Vite dev server on `:5173` proxies `/api/*` to the Go server on
  `:8080`. Run `pnpm dev` and `./server` in separate terminals.
- **Prod:** `go build` produces `./server`, which serves the compiled client
  from `./client/dist` for non-`/api` routes (SPA fallback to `index.html`) and
  handles `/api/*` itself. Single process, single port.

## 4. Directory structure

```
listening-log/
├── client/
│   ├── index.html
│   ├── package.json            # pnpm
│   ├── pnpm-lock.yaml
│   ├── vite.config.ts          # /api proxy → :8080 in dev
│   ├── tsconfig.json
│   └── src/
│       ├── main.tsx
│       ├── App.tsx             # reads /api/status, shows Login or Home
│       └── api/client.ts       # thin fetch wrapper
├── server/
│   ├── go.mod
│   ├── main.go                 # Gin setup, routes, static serving
│   ├── config/config.go        # env loading (.env)
│   ├── spotify/spotify.go      # authorize URL, code exchange
│   ├── db/
│   │   ├── db.go               # open sqlite (modernc), apply schema
│   │   ├── schema.sql          # CREATE TABLE IF NOT EXISTS ...
│   │   └── auth.go             # get/set spotify creds
│   └── handlers/
│       └── auth.go             # /api/auth/*, /api/status
├── data/
│   └── database.sqlite         # gitignored; created on first run
├── .env                        # CLIENT_ID, CLIENT_SECRET, ...
└── spec/
    ├── 00-overview.md
    └── 01-phase-1-web-app-scaffolding.md   # this file
```

## 5. Tech stack & key dependencies

| Concern        | Choice                                          |
| -------------- | ----------------------------------------------- |
| Client build   | Vite + TypeScript + React                       |
| Client pkg mgr | pnpm                                            |
| Server         | Go (latest, ≥1.24) + `github.com/gin-gonic/gin` |
| SQLite driver  | `modernc.org/sqlite` (pure-Go, **no cgo**)      |
| Schema         | single `schema.sql` applied at boot             |
| Env loading    | `github.com/joho/godotenv` (dev) + `os.Getenv`  |

> Pure-Go SQLite + no cgo means `go build` yields a single static `./server`
> binary with no C toolchain required.

## 6. Configuration (`.env`)

Shared, read by the server. `CLIENT_SECRET` never leaves the server.

```dotenv
# Spotify app credentials (from developer.spotify.com)
CLIENT_ID=...
CLIENT_SECRET=...

# OAuth redirect — must be registered EXACTLY in the Spotify app dashboard
SPOTIFY_REDIRECT_URI=http://127.0.0.1:8080/api/auth/callback

# Where to send the browser after a successful login (the SPA)
CLIENT_BASE_URL=http://127.0.0.1:5173

# Server
PORT=8080
DATABASE_PATH=./data/database.sqlite
```

> Spotify now requires loopback redirects to use `127.0.0.1` rather than
> `localhost`; use `127.0.0.1` consistently in the dashboard and `.env`.

## 7. OAuth scopes

Request the broad, future-proof set up front so later phases never force a
re-consent:

```
user-read-currently-playing
user-read-playback-state
user-read-recently-played
```

## 8. Data model

SQLite, created via a single `schema.sql` (`CREATE TABLE IF NOT EXISTS`) applied
when the server boots. Phase 1 introduces only the single-row auth table;
listening-data tables arrive with the scraper in Phase 2.

`db/schema.sql`:

```sql
CREATE TABLE IF NOT EXISTS spotify_auth (
    id            INTEGER PRIMARY KEY CHECK (id = 1),  -- single row
    access_token  TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    scope         TEXT NOT NULL,
    expiry        INTEGER NOT NULL,                    -- unix seconds
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO spotify_auth (id, access_token, refresh_token, scope, expiry)
SELECT 1, '', '', '', 0
WHERE NOT EXISTS (SELECT 1 FROM spotify_auth WHERE id = 1);
```

Notes:
- Exactly one row (`id = 1`), like the prototype. Auth = upsert that row.
- If Spotify omits a new `refresh_token` on a re-auth, keep the existing one
  (the prototype already handled this fallback).
- "Connected" = the row has a non-empty `refresh_token`.
- The Phase 2 scraper reads `spotify_auth` to know how to poll.

## 9. Auth flow (detailed)

Server-side Authorization Code exchange. No sessions — the only persisted state
is the token row.

```
1. Client: user clicks "Connect Spotify" → navigates to /api/auth/login.

2. Server (/api/auth/login): generate a random `state`, set it in a short-lived
   HttpOnly cookie (CSRF), and 302-redirect to:
   https://accounts.spotify.com/authorize?
        client_id=...&response_type=code&redirect_uri=SPOTIFY_REDIRECT_URI
        &scope=<§7 scopes>&state=<state>

3. User consents → Spotify 302 → SPOTIFY_REDIRECT_URI
        i.e. GET /api/auth/callback?code=...&state=...

4. Server (/api/auth/callback):
   a. Verify `state` against the cookie (reject on mismatch).
   b. POST https://accounts.spotify.com/api/token
        grant_type=authorization_code, code, redirect_uri,
        Authorization: Basic base64(CLIENT_ID:CLIENT_SECRET)
      → { access_token, refresh_token, expires_in, scope }
   c. Upsert spotify_auth row (id=1) with the tokens, scope, and
      expiry = now + expires_in.
   d. 302 redirect the browser to CLIENT_BASE_URL.

5. Client: on load, calls GET /api/status → { connected: true }, shows Home.
```

CSRF: validate `state` in step 4a. A short-lived HttpOnly cookie set in step 2
is sufficient for this local single-user app.

## 10. API endpoints

| Method | Path                 | Purpose                                                                |
| ------ | -------------------- | --------------------------------------------------------------------- |
| GET    | `/api/auth/login`    | Sets `state` cookie, redirects to Spotify authorize URL.              |
| GET    | `/api/auth/callback` | Spotify redirect target; exchanges code, persists tokens, redirects to client. |
| GET    | `/api/status`        | `{ "connected": true|false }` — whether valid Spotify tokens exist.   |
| GET    | `/api/health`        | Liveness check (`{ "status": "ok" }`).                                |

## 11. Client (React) — minimal

Two states, gated on `GET /api/status`:

- **Login** — shown when `connected: false`. A single "Connect Spotify" button
  that navigates to `/api/auth/login`.
- **Home** — shown when `connected: true`. A placeholder ("Spotify connected ✓")
  confirming the handshake succeeded. (Listening data is Phase 2+.)

`src/api/client.ts` is a thin `fetch` wrapper. Styling is intentionally minimal
— this is scaffolding.

`vite.config.ts` proxies `/api` → `http://127.0.0.1:8080` in dev so the client
and server share an origin from the browser's perspective.

## 12. Dev & build workflow

**Dev (two terminals):**
```bash
# terminal 1 — client
cd client && pnpm install && pnpm dev          # http://127.0.0.1:5173

# terminal 2 — server
cd server && go run .                            # http://127.0.0.1:8080
# (or: go build -o ../server . && ../server)
```

**Prod build:**
```bash
cd client && pnpm build                          # → client/dist
cd ../server && go build -o ../server .
./server                                          # serves dist + /api on :PORT
```

The server serves `../client/dist` (path configurable) when present, with SPA
fallback: any non-`/api` route returns `index.html`.

## 13. Security considerations

- `CLIENT_SECRET` lives only on the server / in `.env`; never shipped to the
  browser. The server-side code exchange keeps the secret off the client.
- Validate the OAuth `state` to prevent CSRF on the callback.
- `.env`, `./data/database.sqlite`, `./server`, and `client/dist` are
  gitignored.
- Spotify tokens are stored unencrypted in SQLite — acceptable for a local,
  single-user, personal MVP.

## 14. Definition of done

1. Fresh clone → set `.env` → run client and server in two terminals.
2. Visiting `http://127.0.0.1:5173` shows the **Login** page.
3. Clicking "Connect Spotify" completes Spotify consent and returns to the app
   showing "Spotify connected ✓".
4. `./data/database.sqlite` has the `spotify_auth` row (id=1) populated with
   non-empty access + refresh tokens and a future `expiry`.
5. Re-authenticating updates the same row (no duplicates); a missing
   `refresh_token` in the response keeps the existing one.
6. `GET /api/status` reflects connected/not-connected correctly.
7. `go build` succeeds with **no cgo**; `./server` serves the built client and
   the API on a single port.

## 15. Open questions / future hooks

- **Token refresh ownership:** Phase 1 stores tokens but does not refresh them.
  Confirm Phase 2 (scraper) owns refresh.
- **Single → multi-user:** the single-row `spotify_auth` table is deliberately
  the prototype shape; revisit only if multi-user is ever wanted.
