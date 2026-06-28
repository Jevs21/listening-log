# listening-log

Spotify listening tracker with playback polling, listen analysis, and dashboards.

## Setup

```bash
cp .env.example .env
# Fill in CLIENT_ID, CLIENT_SECRET, SPOTIFY_ALLOWED_USER_ID
docker compose up --build
```

App runs on `localhost:8080`, Metabase on `localhost:3000`.

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
