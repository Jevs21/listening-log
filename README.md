# listening-log

Spotify listening tracker with playback polling, listen analysis, and dashboards.

## Setup

```bash
cp .env.example .env
# Fill in CLIENT_ID, CLIENT_SECRET, SPOTIFY_ALLOWED_USER_ID
docker compose up --build
```

App runs on `localhost:${PORT}` (default `8080`), Metabase on `localhost:${METABASE_PORT}` (default `3000`).

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Host port for the app |
| `METABASE_PORT` | `3000` | Host port for Metabase |
| `CLIENT_ID` | *(required)* | Spotify OAuth client ID |
| `CLIENT_SECRET` | *(required)* | Spotify OAuth client secret |
| `SPOTIFY_ALLOWED_USER_ID` | *(required)* | Spotify user ID allowed to auth |
| `DATABASE_URL` | `postgres://listening_log:listening_log@localhost:5432/listening_log?sslmode=disable` | Postgres connection string |
| `MB_ADMIN_EMAIL` | `admin@listening-log.local` | Metabase initial admin email |
| `MB_ADMIN_PASSWORD` | `changeme` | Metabase initial admin password |
| `MB_ADMIN_FIRST_NAME` | `Admin` | Metabase admin first name |
| `MB_ADMIN_LAST_NAME` | `User` | Metabase admin last name |

## Data Export / Import

```bash
# Export
docker compose exec app dbtools export /data/backup.tar.gz

# Import (replace all data)
docker compose exec app dbtools import --mode=overwrite /data/backup.tar.gz

# Import (merge, skip duplicates, update newer metadata)
docker compose exec app dbtools import --mode=merge /data/backup.tar.gz
```

Archives are saved to `./backups/` on the host.
