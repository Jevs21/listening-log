# Phase 29 — Reverse-Proxy Dashboard Embeds

Builds on phase 21 (embed dashboard).

## Goal

Replace the direct-to-Metabase iframe URL with a reverse proxy through the Go app so the embedded dashboard loads from any origin — localhost, LAN IP, or external domain via Cloudflare — without exposing Metabase externally.

## Problem

The `GET /api/stats/dashboard` handler returns an iframe URL using `METABASE_URL` (hardcoded to `http://localhost:3000`). The browser must reach Metabase directly, which fails from any host other than localhost. Metabase port 3000 is only accessible in-network, not through Cloudflare.

## Scope

### In scope

- Go reverse proxy that forwards `/metabase/public/dashboard/*` to internal Metabase
- Update iframe URL construction to use relative paths
- Remove `METABASE_URL` env var (replace with internal Docker URL)
- Restrict proxy to public dashboard embed routes only

### Out of scope

- Proxying the Metabase admin UI or any non-embed routes
- Authentication/authorization on the proxy (public dashboards are already token-gated by UUID)
- Changes to Metabase's Docker port mapping (3000 stays exposed on host for local dev/admin access)

## Backend changes

### `server/config/config.go`

Rename `MetabaseURL` to `MetabaseInternalURL`. Change the env var to `METABASE_INTERNAL_URL`, default `http://metabase:3000`. This is the Docker-internal address, never sent to the browser.

### `server/handlers/stats.go`

Change `DashboardURL` to return a **relative** URL instead of an absolute one:

```go
url := fmt.Sprintf("/metabase/public/dashboard/%s#theme=night&background=false&bordered=false&titled=false",
    strings.TrimSpace(string(uuid)),
)
```

The handler no longer needs `metabaseURL` as a parameter — it only needs to read the UUID file. Update the function signature accordingly.

### `server/handlers/metabase_proxy.go` (new)

Reverse proxy handler scoped to public dashboard embeds only:

```go
func MetabaseProxy(metabaseInternalURL string) gin.HandlerFunc {
    target, _ := url.Parse(metabaseInternalURL)
    proxy := httputil.NewSingleHostReverseProxy(target)

    return func(c *gin.Context) {
        // Only allow /metabase/public/dashboard/* and static assets needed by embeds
        path := c.Request.URL.Path
        stripped := strings.TrimPrefix(path, "/metabase")

        allowed := strings.HasPrefix(stripped, "/public/dashboard/") ||
            strings.HasPrefix(stripped, "/app/") ||
            strings.HasPrefix(stripped, "/api/") ||
            strings.HasPrefix(stripped, "/public/")

        if !allowed {
            c.JSON(http.StatusForbidden, gin.H{"error": "not allowed"})
            return
        }

        c.Request.URL.Path = stripped
        c.Request.Host = target.Host
        proxy.ServeHTTP(c.Writer, c.Request)
    }
}
```

Metabase public dashboard pages load JS/CSS from `/app/` and make API calls to `/api/` — the proxy must pass these through for the embed to render.

### `server/main.go`

Remove the `metabaseURL` argument from `DashboardURL`. Add the proxy route:

```go
r.GET("/api/stats/dashboard", handlers.DashboardURL())
r.Any("/metabase/*path", handlers.MetabaseProxy(cfg.MetabaseInternalURL))
```

## Environment changes

### `.env.example`

Replace:
```
METABASE_URL=http://localhost:3000
```

With:
```
METABASE_INTERNAL_URL=http://metabase:3000
```

### `.env`

Same change. The value `http://metabase:3000` uses Docker's internal DNS and works inside the Docker network. For local development outside Docker, override to `http://localhost:3000`.

## Frontend changes

None. `StatsPage.tsx` already uses the URL returned by `/api/stats/dashboard` as-is. Since it becomes a relative path, the iframe will resolve against the current origin automatically.

## Definition of done

- [ ] `GET /api/stats/dashboard` returns a relative URL like `/metabase/public/dashboard/<uuid>#theme=night&...`
- [ ] Visiting the app at `localhost:8080/stats` loads the dashboard embed through the proxy
- [ ] Visiting the app at `10.0.0.X:8080/stats` loads the dashboard embed through the proxy
- [ ] Visiting the app through an external domain (Cloudflare) loads the dashboard embed through the proxy
- [ ] `GET /metabase/public/dashboard/<uuid>` proxies to Metabase and returns the embed page
- [ ] `GET /metabase/app/*` and `GET /metabase/api/*` proxy through (needed for embed assets)
- [ ] `GET /metabase/collection/` or other non-embed routes return 403
- [ ] Metabase admin UI is NOT accessible through the proxy
- [ ] `METABASE_URL` env var is replaced with `METABASE_INTERNAL_URL`
- [ ] Metabase port 3000 remains mapped to host in `docker-compose.yml` for local admin access
