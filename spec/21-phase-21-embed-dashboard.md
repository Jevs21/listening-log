# Phase 21 — Embed Listening Overview Dashboard

Builds on phase 20 (dashboards as code) and phase 14 (stats page).

## Goal

Replace the "stats coming soon" placeholder on `/stats` with the Metabase "Listening Overview" dashboard embedded in an iframe. The dashboard is publicly shared within the Docker network, and the Go backend provides the iframe URL to the frontend.

## Scope

### In scope

- Enable Metabase public sharing and create a public link for the dashboard (in setup scripts)
- Write the public UUID to a shared volume so the Go server can read it
- New `METABASE_URL` env var for the browser-accessible Metabase URL
- Go endpoint `GET /api/stats/dashboard` returning the iframe embed URL
- `Loading` component for reuse
- `StatsPage` renders the dashboard in a full-content-area iframe with no Metabase chrome

### Out of scope

- Embedding multiple dashboards or dynamic dashboard selection
- Metabase signed/JWT embedding
- Styling the Metabase dashboard itself (colors, fonts)
- Any Metabase configuration beyond enabling public sharing

## Setup script changes

### `metabase/load-dashboards.sh`

After creating (or finding) a dashboard, enable public sharing and create a public link. Write the UUID to a shared volume.

At the top of the script, add:

```bash
SHARED_DIR="/shared"
mkdir -p "$SHARED_DIR"
```

After the "Create dashboard" block, add a new section that runs for both new and existing dashboards (move the existing-dashboard `continue` down past this):

```bash
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
```

The existing-dashboard `continue` should be removed so that the public link and card-adding logic runs regardless. Guard the card-adding section with `if [ -n "$DASH_ID" ]` (only add cards for newly created dashboards).

### `docker-compose.yml`

Add a `metabase-shared` volume and mount it to both `app` and `metabase-setup`:

```yaml
app:
  volumes:
    - metabase-shared:/shared:ro

metabase-setup:
  volumes:
    # ... existing mounts ...
    - metabase-shared:/shared
```

Add to the `volumes:` section:

```yaml
metabase-shared:
```

## Environment changes

Add to `.env.example`:

```
METABASE_URL=http://localhost:3000
```

This is the browser-accessible Metabase URL (not the internal Docker URL). The Go server uses it to construct the iframe embed URL.

## Backend changes

### `server/config/config.go`

Add `MetabaseURL string` to the `Config` struct. Load from `METABASE_URL` env var, default to `http://localhost:3000`.

### `server/handlers/stats.go` (new)

Single handler:

```go
func DashboardURL(metabaseURL string) gin.HandlerFunc {
    return func(c *gin.Context) {
        uuid, err := os.ReadFile("/shared/dashboard-uuid")
        if err != nil {
            c.JSON(http.StatusServiceUnavailable, gin.H{"error": "dashboard not available yet"})
            return
        }

        url := fmt.Sprintf("%s/public/dashboard/%s#bordered=false&titled=false&theme=transparent",
            strings.TrimRight(metabaseURL, "/"),
            strings.TrimSpace(string(uuid)),
        )

        c.JSON(http.StatusOK, gin.H{"url": url})
    }
}
```

The `#bordered=false&titled=false&theme=transparent` parameters strip the Metabase header, title, and background.

### `server/main.go`

Register the route:

```go
r.GET("/api/stats/dashboard", handlers.DashboardURL(cfg.MetabaseURL))
```

## Frontend changes

### `client/src/components/Loading.tsx` (new)

Minimal loading component for reuse:

```tsx
export function Loading() {
  return <p className="app-description">loading</p>;
}
```

Uses the existing `app-description` class for consistent styling. No CSS file needed.

### `client/src/api/stats.ts` (new)

```ts
export async function getDashboardURL(): Promise<{ url: string }> {
  const res = await fetch("/api/stats/dashboard");
  if (!res.ok) throw new Error("dashboard not available");
  return res.json();
}
```

### `client/src/pages/StatsPage.tsx`

Replace the placeholder content with:

1. Fetch suggestion check and dashboard URL in parallel on mount
2. Redirect to gate page if no suggestion (existing behavior)
3. Show `<Loading />` while fetching
4. Once the URL is loaded, render a full-content-area iframe

The iframe should:
- Have no border (`border: none`)
- Fill the available viewport height (use `calc(100vh - <header offset>)` or a simple fixed approach)
- Set `width: 100%`

```tsx
import { useEffect, useState } from "react";
import { Navigate } from "react-router-dom";
import { checkSuggestion } from "../api/suggestions";
import { getDashboardURL } from "../api/stats";
import { Loading } from "../components/Loading";

export function StatsPage() {
  const [hasSuggested, setHasSuggested] = useState<boolean | null>(null);
  const [dashboardURL, setDashboardURL] = useState<string | null>(null);

  useEffect(() => {
    checkSuggestion().then((data) => setHasSuggested(data.has_suggested));
    getDashboardURL().then((data) => setDashboardURL(data.url)).catch(() => {});
  }, []);

  if (hasSuggested === null) return null;
  if (!hasSuggested) return <Navigate to="/woah-hold-it-right-there-buckaroo" replace />;
  if (!dashboardURL) return <Loading />;

  return (
    <iframe
      src={dashboardURL}
      style={{ width: "100%", height: "100vh", border: "none" }}
      title="Listening Overview"
    />
  );
}
```

## File structure

```
server/
  handlers/stats.go          (new)
  config/config.go           (modified)
  main.go                    (modified — new route)
client/src/
  api/stats.ts               (new)
  components/Loading.tsx      (new)
  pages/StatsPage.tsx         (modified)
metabase/
  load-dashboards.sh          (modified — public sharing + UUID file)
docker-compose.yml             (modified — shared volume)
.env.example                   (modified — METABASE_URL)
```

## Definition of done

- [ ] `docker compose up` on a fresh volume enables public sharing and creates a public link for the Listening Overview dashboard
- [ ] Public UUID is written to the shared volume at `/shared/dashboard-uuid`
- [ ] `GET /api/stats/dashboard` returns `{ "url": "http://localhost:3000/public/dashboard/<uuid>#bordered=false&titled=false&theme=transparent" }`
- [ ] `GET /api/stats/dashboard` returns 503 if the UUID file doesn't exist yet (metabase-setup hasn't run)
- [ ] `/stats` page renders the Metabase dashboard in a full-viewport iframe with no Metabase chrome
- [ ] `/stats` still redirects to the gate page if the user hasn't submitted a suggestion
- [ ] `Loading` component shows while dashboard URL is being fetched
- [ ] On subsequent `docker compose up`, the existing public UUID is reused (not recreated)
- [ ] `.env.example` includes `METABASE_URL=http://localhost:3000`
