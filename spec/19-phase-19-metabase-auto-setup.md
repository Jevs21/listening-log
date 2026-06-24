# Phase 19 — Metabase Auto-Setup

Builds on phase 18 (metabase + read replica).

## Goal

Automate the Metabase setup wizard so that `docker compose up` yields a fully configured Metabase instance with no manual interaction. Admin credentials and the replica database connection are configured via environment variables and an init script that calls the Metabase setup API.

## Scope

### In scope

- Shell script that completes Metabase setup via `POST /api/setup`
- One-shot `metabase-setup` Docker Compose service that runs the script and exits
- Environment variables for admin credentials, added to `.env.example`
- Idempotent: skips setup if Metabase is already configured

### Out of scope

- Metabase Pro/Enterprise config file (`config.yml`) — not available on OSS
- Creating dashboards, questions, or saved queries
- Additional Metabase users beyond the initial admin

## Why not the config file?

Metabase's `config.yml` / `MB_CONFIG_FILE_PATH` feature requires a Pro or Enterprise license. The OSS image (`metabase/metabase:latest`) does not support it. The `/api/setup` endpoint is the supported path for programmatic setup on the free tier.

## Environment variables

Add to `.env.example`:

```
MB_ADMIN_EMAIL=admin@listening-log.local
MB_ADMIN_PASSWORD=changeme
MB_ADMIN_FIRST_NAME=Admin
MB_ADMIN_LAST_NAME=User
```

These are passed to the `metabase-setup` service via `env_file: .env`.

## Setup script

Create `metabase/setup.sh`:

```bash
#!/bin/bash
set -e

MB_URL="http://metabase:3000"

# Wait for Metabase to be ready (health endpoint returns 200)
echo "Waiting for Metabase to be ready..."
until curl -s "$MB_URL/api/health" | grep -q "ok"; do
  sleep 2
done

# Get the setup token from session properties
# If setup-token is null/missing, Metabase is already configured
SETUP_TOKEN=$(curl -s "$MB_URL/api/session/properties" | jq -r '.["setup-token"] // empty')

if [ -z "$SETUP_TOKEN" ]; then
  echo "Metabase is already configured. Skipping setup."
  exit 0
fi

echo "Running Metabase setup..."

curl -s -X POST "$MB_URL/api/setup" \
  -H "Content-Type: application/json" \
  -d "{
    \"token\": \"$SETUP_TOKEN\",
    \"user\": {
      \"first_name\": \"${MB_ADMIN_FIRST_NAME}\",
      \"last_name\": \"${MB_ADMIN_LAST_NAME}\",
      \"email\": \"${MB_ADMIN_EMAIL}\",
      \"password\": \"${MB_ADMIN_PASSWORD}\",
      \"site_name\": \"listening-log\"
    },
    \"database\": {
      \"engine\": \"postgres\",
      \"name\": \"Listening Log (Replica)\",
      \"details\": {
        \"host\": \"db-replica\",
        \"port\": 5432,
        \"dbname\": \"listening_log\",
        \"user\": \"listening_log\",
        \"password\": \"listening_log\"
      }
    },
    \"prefs\": {
      \"site_name\": \"listening-log\",
      \"allow_tracking\": false
    }
  }"

echo ""
echo "Metabase setup complete."
```

## Docker Compose changes

Add a `metabase-setup` service to `docker-compose.yml`:

```yaml
metabase-setup:
  image: alpine:latest
  entrypoint: ["/bin/sh", "-c", "apk add --no-cache curl jq && /setup.sh"]
  volumes:
    - ./metabase/setup.sh:/setup.sh:ro
  env_file: .env
  depends_on:
    metabase:
      condition: service_healthy
  restart: "no"
```

Add a healthcheck to the existing `metabase` service so `metabase-setup` can depend on it:

```yaml
metabase:
  # ... existing config ...
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:3000/api/health"]
    interval: 10s
    timeout: 5s
    retries: 20
    start_period: 60s
```

The `start_period` is generous because Metabase takes a while to initialize on first boot (database migrations, etc.).

## File structure

```
metabase/
  setup.sh
```

## Definition of done

- [ ] `docker compose up` starts Metabase and the setup service runs automatically
- [ ] On a fresh volume (no prior state), Metabase setup completes without manual interaction
- [ ] After setup, Metabase at `http://localhost:3000` shows the home screen (not the setup wizard)
- [ ] Admin can log in with the credentials from `.env`
- [ ] The "Listening Log (Replica)" database appears in Metabase's admin database list
- [ ] On subsequent `docker compose up` (with existing volume), the setup service detects Metabase is already configured and exits cleanly
- [ ] `.env.example` contains placeholder values for `MB_ADMIN_*` variables
