# Phase 2: Scraper MVP

> Builds on Phase 1. A standalone Go service that reads Spotify credentials
> from the shared SQLite database, polls the currently-playing endpoint, and
> prints the result to the console.

## Goal

Run `./scraper` alongside the server. It reads tokens from
`./data/database.sqlite`, polls Spotify every 30 seconds, and prints what's
currently playing to stdout. It owns token refresh.

## Scope

### In scope
- `./scraper` — separate Go module, separate binary.
- Read `spotify_auth` row from the shared SQLite DB (same file as Phase 1).
- Call `GET https://api.spotify.com/v1/me/player/currently-playing`.
- Print track name, artist, album to stdout on each poll.
- Refresh the access token when expired, update the DB row.
- 30-second polling interval via a ticker.

### Out of scope
- Writing play history to the DB (Phase 3).
- Any web UI changes.
- Docker / deployment packaging.

## Directory structure

```
listening-log/
├── scraper/
│   ├── go.mod
│   ├── main.go            # ticker loop, orchestration
│   ├── config/config.go   # loads .env (same vars as server)
│   ├── db/db.go           # open sqlite, read/update spotify_auth
│   └── spotify/spotify.go # currently-playing call, token refresh
├── server/                # (Phase 1, unchanged)
├── client/                # (Phase 1, unchanged)
├── data/
│   └── database.sqlite    # shared
└── .env
```

## Config

Reads the same `.env` as the server. Only needs:

```dotenv
CLIENT_ID=...
CLIENT_SECRET=...
DATABASE_PATH=./data/database.sqlite
```

## Token refresh

The scraper owns refresh. Before each API call:
1. Read `expiry` from `spotify_auth`.
2. If expired (or within 60s of expiry), POST to
   `https://accounts.spotify.com/api/token` with `grant_type=refresh_token`.
3. Update `access_token`, `expiry` in the DB. Keep `refresh_token` unless
   Spotify returns a new one.

## Console output

```
2024-03-15 10:32:00  ♫ "Song Name" — Artist Name (Album Name)
2024-03-15 10:32:30  ♫ "Song Name" — Artist Name (Album Name)
2024-03-15 10:33:00  ⏸ Nothing playing
```

## Definition of done

1. `cd scraper && go build -o ../scraper .` succeeds (no cgo).
2. Run `./scraper` — it reads tokens from the shared DB.
3. If Spotify is playing, prints track/artist/album to stdout every 30s.
4. If nothing is playing, prints a "nothing playing" line.
5. When the access token expires, the scraper refreshes it and continues.
