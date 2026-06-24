# Phase 20 — Metabase Dashboards as Code

Builds on phase 19 (metabase auto-setup).

## Goal

Define Metabase saved questions and dashboards as JSON files in the repo. The existing `metabase-setup` service loads them via the Metabase API after initial setup, so `docker compose up` on a fresh volume yields a fully configured Metabase with dashboards ready to use. Include one starter dashboard with two example visualizations.

## Scope

### In scope

- JSON definition format for saved questions (native SQL cards) and dashboards
- Loader script that reads definitions and creates them via Metabase API
- Integration into the existing `metabase/setup.sh` flow
- One example dashboard ("Listening Overview") with two questions
- Idempotent: skips creation if a question/dashboard with the same name already exists

### Out of scope

- Exporting dashboards from Metabase back to code (bidirectional sync)
- Metabase collections / folder organization
- Dashboard permissions or row-level security
- Embedding dashboards into the web app

## Definition format

Each dashboard is a single JSON file in `metabase/dashboards/`. The file contains both the questions and the dashboard layout that references them.

Create `metabase/dashboards/listening-overview.json`:

```json
{
  "name": "Listening Overview",
  "description": "High-level listening activity and top artists.",
  "questions": [
    {
      "ref": "listens-per-day",
      "name": "Listens Per Day",
      "display": "line",
      "native_query": "SELECT DATE(l.started_at) AS day, COUNT(*) AS listens FROM listen l WHERE l.started_at >= CURRENT_DATE - INTERVAL '30 days' GROUP BY day ORDER BY day",
      "visualization_settings": {
        "graph.x_axis.title_text": "Day",
        "graph.y_axis.title_text": "Listens"
      }
    },
    {
      "ref": "top-artists",
      "name": "Top Artists (Last 30 Days)",
      "display": "bar",
      "native_query": "SELECT a.name AS artist, COUNT(*) AS listens FROM listen l JOIN track t ON t.spotify_id = l.track_id JOIN artist a ON a.spotify_id = t.artist_id WHERE l.started_at >= CURRENT_DATE - INTERVAL '30 days' GROUP BY a.name ORDER BY listens DESC LIMIT 15",
      "visualization_settings": {
        "graph.x_axis.title_text": "Artist",
        "graph.y_axis.title_text": "Listens"
      }
    }
  ],
  "cards": [
    { "ref": "listens-per-day", "row": 0, "col": 0, "size_x": 12, "size_y": 6 },
    { "ref": "top-artists",     "row": 6, "col": 0, "size_x": 12, "size_y": 8 }
  ]
}
```

Key design decisions:

- **`ref`** is a local identifier that links questions to dashboard card placements. It is not sent to Metabase.
- **`display`** maps to Metabase's card `display` field (`line`, `bar`, `table`, `scalar`, `pie`, etc.).
- **`native_query`** — all questions use native (raw SQL) mode. This is the most portable and readable format for code-defined queries. Queries run against the replica database added during setup.
- **`visualization_settings`** is passed through directly to the Metabase API.
- **`cards`** defines the dashboard grid layout. `row`/`col`/`size_x`/`size_y` map directly to the Metabase dashboard card placement API.

## Loader script

Create `metabase/load-dashboards.sh`. This script:

1. Accepts `MB_URL` and `SESSION_TOKEN` as environment variables (passed from `setup.sh`)
2. Finds the database ID for "Listening Log (Replica)" via `GET /api/database`
3. Iterates over each `*.json` file in `/dashboards/`
4. For each file:
   - Reads the JSON
   - For each question in `questions[]`:
     - Checks if a card with that name already exists (`GET /api/card` and filter by name)
     - If not, creates it via `POST /api/card` with `dataset_query.type: "native"`, `dataset_query.native.query`, and `database` set to the replica DB ID
     - Records the Metabase-assigned card ID, keyed by `ref`
   - Checks if a dashboard with that name already exists (`GET /api/dashboard` and filter)
   - If not, creates the dashboard via `POST /api/dashboard`
   - For each entry in `cards[]`, adds the card to the dashboard via `POST /api/dashboard/:id` (PUT with the full cards array), mapping `ref` to the Metabase card ID from the previous step

```bash
#!/bin/sh
set -e

# Expects: MB_URL, SESSION_TOKEN
AUTH_HEADER="X-Metabase-Session: $SESSION_TOKEN"
DASHBOARD_DIR="/dashboards"

# Find the replica database ID
DB_ID=$(curl -s "$MB_URL/api/database" \
  -H "$AUTH_HEADER" | jq '.data[] | select(.name == "Listening Log (Replica)") | .id')

if [ -z "$DB_ID" ]; then
  echo "Error: could not find 'Listening Log (Replica)' database in Metabase."
  exit 1
fi

echo "Using database ID: $DB_ID"

for file in "$DASHBOARD_DIR"/*.json; do
  [ -f "$file" ] || continue
  echo "Processing $(basename "$file")..."

  DASHBOARD_NAME=$(jq -r '.name' "$file")
  DASHBOARD_DESC=$(jq -r '.description // ""' "$file")

  # --- Create questions ---
  # Build a ref->metabase_id map
  REF_MAP="{}"
  QUESTION_COUNT=$(jq '.questions | length' "$file")
  i=0
  while [ "$i" -lt "$QUESTION_COUNT" ]; do
    REF=$(jq -r ".questions[$i].ref" "$file")
    NAME=$(jq -r ".questions[$i].name" "$file")
    DISPLAY=$(jq -r ".questions[$i].display" "$file")
    QUERY=$(jq -r ".questions[$i].native_query" "$file")
    VIZ=$(jq -c ".questions[$i].visualization_settings // {}" "$file")

    # Check if card already exists by name
    EXISTING_ID=$(curl -s "$MB_URL/api/card" -H "$AUTH_HEADER" | \
      jq -r ".[] | select(.name == \"$NAME\") | .id" | head -1)

    if [ -n "$EXISTING_ID" ]; then
      echo "  Question '$NAME' already exists (id=$EXISTING_ID), skipping."
      CARD_ID="$EXISTING_ID"
    else
      echo "  Creating question '$NAME'..."
      CARD_ID=$(curl -s -X POST "$MB_URL/api/card" \
        -H "Content-Type: application/json" \
        -H "$AUTH_HEADER" \
        -d "$(jq -n \
          --arg name "$NAME" \
          --arg display "$DISPLAY" \
          --arg query "$QUERY" \
          --argjson db_id "$DB_ID" \
          --argjson viz "$VIZ" \
          '{
            name: $name,
            display: $display,
            dataset_query: {
              type: "native",
              native: { query: $query },
              database: $db_id
            },
            visualization_settings: $viz
          }')" | jq -r '.id')
      echo "  Created question '$NAME' (id=$CARD_ID)"
    fi

    REF_MAP=$(echo "$REF_MAP" | jq --arg ref "$REF" --argjson id "$CARD_ID" '. + {($ref): $id}')
    i=$((i + 1))
  done

  # --- Create dashboard ---
  EXISTING_DASH=$(curl -s "$MB_URL/api/dashboard" -H "$AUTH_HEADER" | \
    jq -r ".[] | select(.name == \"$DASHBOARD_NAME\") | .id" | head -1)

  if [ -n "$EXISTING_DASH" ]; then
    echo "  Dashboard '$DASHBOARD_NAME' already exists (id=$EXISTING_DASH), skipping."
    continue
  fi

  echo "  Creating dashboard '$DASHBOARD_NAME'..."
  DASH_ID=$(curl -s -X POST "$MB_URL/api/dashboard" \
    -H "Content-Type: application/json" \
    -H "$AUTH_HEADER" \
    -d "$(jq -n --arg name "$DASHBOARD_NAME" --arg desc "$DASHBOARD_DESC" \
      '{ name: $name, description: $desc }')" | jq -r '.id')

  echo "  Created dashboard (id=$DASH_ID)"

  # --- Add cards to dashboard ---
  CARDS_PAYLOAD=$(jq -c --argjson ref_map "$REF_MAP" \
    '[.cards[] | {
      id: -1,
      card_id: ($ref_map[.ref] | tonumber),
      row: .row,
      col: .col,
      size_x: .size_x,
      size_y: .size_y
    }]' "$file")

  curl -s -X PUT "$MB_URL/api/dashboard/$DASH_ID" \
    -H "Content-Type: application/json" \
    -H "$AUTH_HEADER" \
    -d "$(jq -n --argjson cards "$CARDS_PAYLOAD" '{ cards: $cards }')" > /dev/null

  echo "  Added cards to dashboard."
done

echo "Dashboard loading complete."
```

## Changes to setup.sh

After the existing setup logic (admin user + database verification), add a call to the loader:

```bash
# At the end of setup.sh, after "Metabase setup complete."

echo "Loading dashboards..."
SESSION_TOKEN="$SESSION_TOKEN" MB_URL="$MB_URL" /load-dashboards.sh
```

For the case where Metabase was already configured (the early `exit 0` path), the loader should still run so that new dashboard definitions are picked up on subsequent boots. Restructure setup.sh so the early exit skips only the `/api/setup` call, not the dashboard loading. The loader needs a session token, so add a login step:

```bash
if [ -z "$SETUP_TOKEN" ]; then
  echo "Metabase is already configured. Skipping initial setup."
  # Log in to get a session token for dashboard loading
  SESSION_TOKEN=$(curl -s -X POST "$MB_URL/api/session" \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"${MB_ADMIN_EMAIL}\", \"password\": \"${MB_ADMIN_PASSWORD}\"}" | jq -r '.id // empty')
else
  # ... existing setup logic, SESSION_TOKEN comes from setup response ...
fi

# Dashboard loading runs in both paths
if [ -n "$SESSION_TOKEN" ]; then
  echo "Loading dashboards..."
  SESSION_TOKEN="$SESSION_TOKEN" MB_URL="$MB_URL" /load-dashboards.sh
fi
```

## Docker Compose changes

Mount the dashboards directory and loader script into the `metabase-setup` service:

```yaml
metabase-setup:
  image: alpine:latest
  entrypoint: ["/bin/sh", "-c", "apk add --no-cache curl jq && /setup.sh"]
  volumes:
    - ./metabase/setup.sh:/setup.sh:ro
    - ./metabase/load-dashboards.sh:/load-dashboards.sh:ro
    - ./metabase/dashboards:/dashboards:ro
  env_file: .env
  depends_on:
    metabase:
      condition: service_healthy
  restart: "no"
```

## File structure

```
metabase/
  setup.sh                              (modified)
  load-dashboards.sh                    (new)
  dashboards/
    listening-overview.json             (new)
```

## Definition of done

- [ ] `docker compose up` on a fresh volume creates the admin user, replica DB connection, **and** the "Listening Overview" dashboard with both questions
- [ ] Dashboard at `http://localhost:3000` shows "Listening Overview" with "Listens Per Day" (line chart) and "Top Artists" (bar chart)
- [ ] On subsequent `docker compose up` (existing volume), the loader detects existing questions/dashboards by name and skips them
- [ ] Adding a new `.json` file to `metabase/dashboards/` and restarting creates the new dashboard on next boot
- [ ] Adding a new question to an existing JSON file creates the question (but does not duplicate existing ones)
- [ ] `load-dashboards.sh` exits cleanly if no JSON files exist in the dashboards directory
