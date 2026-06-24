#!/bin/sh
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
  echo "Metabase is already configured. Skipping initial setup."
  # Log in to get a session token for dashboard loading
  SESSION_TOKEN=$(curl -s -X POST "$MB_URL/api/session" \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"${MB_ADMIN_EMAIL}\", \"password\": \"${MB_ADMIN_PASSWORD}\"}" | jq -r '.id // empty')
else
  echo "Running Metabase setup..."

  SETUP_RESPONSE=$(curl -s -X POST "$MB_URL/api/setup" \
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
    }")

  echo "Setup response: $SETUP_RESPONSE"

  # Extract session token from setup response (used to authenticate further API calls)
  SESSION_TOKEN=$(echo "$SETUP_RESPONSE" | jq -r '.id // empty')

  if [ -z "$SESSION_TOKEN" ]; then
    echo "Warning: could not extract session token from setup response."
    echo "Metabase setup may have failed."
    exit 1
  fi

  # Verify the database was added; if not, add it explicitly
  DB_COUNT=$(curl -s "$MB_URL/api/database" \
    -H "X-Metabase-Session: $SESSION_TOKEN" | jq '[.data[] | select(.name == "Listening Log (Replica)")] | length')

  if [ "$DB_COUNT" -eq 0 ]; then
    echo "Database not created during setup. Adding it now..."
    curl -s -X POST "$MB_URL/api/database" \
      -H "Content-Type: application/json" \
      -H "X-Metabase-Session: $SESSION_TOKEN" \
      -d "{
        \"engine\": \"postgres\",
        \"name\": \"Listening Log (Replica)\",
        \"details\": {
          \"host\": \"db-replica\",
          \"port\": 5432,
          \"dbname\": \"listening_log\",
          \"user\": \"listening_log\",
          \"password\": \"listening_log\"
        }
      }"
    echo ""
    echo "Database added."
  else
    echo "Database already configured."
  fi

  echo "Metabase setup complete."
fi

# Dashboard loading runs in both paths
if [ -n "$SESSION_TOKEN" ]; then
  echo "Loading dashboards..."
  SESSION_TOKEN="$SESSION_TOKEN" MB_URL="$MB_URL" /load-dashboards.sh
fi
