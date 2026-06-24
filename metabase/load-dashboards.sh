#!/bin/sh
set -e

# Expects: MB_URL, SESSION_TOKEN
AUTH_HEADER="X-Metabase-Session: $SESSION_TOKEN"
DASHBOARD_DIR="/dashboards"
SHARED_DIR="/shared"
mkdir -p "$SHARED_DIR"

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
    echo "  Dashboard '$DASHBOARD_NAME' already exists (id=$EXISTING_DASH)."
  else
    echo "  Creating dashboard '$DASHBOARD_NAME'..."
    DASH_ID=$(curl -s -X POST "$MB_URL/api/dashboard" \
      -H "Content-Type: application/json" \
      -H "$AUTH_HEADER" \
      -d "$(jq -n --arg name "$DASHBOARD_NAME" --arg desc "$DASHBOARD_DESC" \
        '{ name: $name, description: $desc }')" | jq -r '.id')

    echo "  Created dashboard (id=$DASH_ID)"

    # --- Add cards to dashboard ---
    CARDS_PAYLOAD=$(jq -c --argjson ref_map "$REF_MAP" \
      '[.cards | to_entries[] | {
        id: (-.key - 1),
        card_id: ($ref_map[.value.ref] | tonumber),
        row: .value.row,
        col: .value.col,
        size_x: .value.size_x,
        size_y: .value.size_y
      }]' "$file")

    curl -s -X PUT "$MB_URL/api/dashboard/$DASH_ID" \
      -H "Content-Type: application/json" \
      -H "$AUTH_HEADER" \
      -d "$(jq -n --argjson dashcards "$CARDS_PAYLOAD" '{ dashcards: $dashcards }')" > /dev/null

    echo "  Added cards to dashboard."
  fi

  # Use DASH_ID from creation, or EXISTING_DASH if it already existed
  FINAL_DASH_ID="${DASH_ID:-$EXISTING_DASH}"

  # --- Enable public sharing (idempotent) ---
  curl -s -X PUT "$MB_URL/api/setting/enable-public-sharing" \
    -H "Content-Type: application/json" \
    -H "$AUTH_HEADER" \
    -d '{"value": true}' > /dev/null

  # --- Get or create public link ---
  PUBLIC_UUID=$(curl -s "$MB_URL/api/dashboard/$FINAL_DASH_ID" \
    -H "$AUTH_HEADER" | jq -r '.public_uuid // empty')

  if [ -z "$PUBLIC_UUID" ]; then
    echo "  Creating public link for dashboard '$DASHBOARD_NAME'..."
    PUBLIC_UUID=$(curl -s -X POST "$MB_URL/api/dashboard/$FINAL_DASH_ID/public_link" \
      -H "$AUTH_HEADER" | jq -r '.uuid')
    echo "  Public UUID: $PUBLIC_UUID"
  else
    echo "  Dashboard '$DASHBOARD_NAME' already has public link (uuid=$PUBLIC_UUID)"
  fi

  # Write UUID to shared volume for the app to read
  echo "$PUBLIC_UUID" > "$SHARED_DIR/dashboard-uuid"
  echo "  Wrote public UUID to $SHARED_DIR/dashboard-uuid"
done

echo "Dashboard loading complete."
