# Phase 30 — Dynamic OAuth Redirect URI

Builds on: Phase 1 (auth flow), Phase 16 (Docker setup)

## Goal

Make the Spotify OAuth callback URL and post-login redirect derive from the incoming request instead of a static env var, so auth works from any access point (local `10.0.0.X:8080`, Cloudflare-proxied `https://myurl.com`, etc.).

## Scope

### In

- Dynamically build the redirect URI from the request's scheme + host
- Detect scheme correctly behind Cloudflare (`X-Forwarded-Proto` / `X-Forwarded-Scheme`)
- Redirect back to the origin URL after successful auth callback
- Remove reliance on `SPOTIFY_REDIRECT_URI` env var

### Out

- Changes to the Spotify Developer Dashboard (user registers URIs manually)
- Changes to the allowed-user gating logic
- HTTPS termination or proxy config

## Backend changes

### `server/config/config.go`

Remove `SpotifyRedirectURI` field and its env-var loading + fallback.

### `server/handlers/auth.go`

#### Helper: `buildBaseURL(c *gin.Context) string`

Derive scheme + host from the request:

1. Scheme: check `X-Forwarded-Proto` header first, fall back to `c.Request.TLS != nil`
2. Host: check `X-Forwarded-Host` header first, fall back to `c.Request.Host`
3. Return `scheme://host` (no trailing slash)

#### `Login`

- Build redirect URI dynamically: `buildBaseURL(c) + "/api/auth/callback"`
- Store the base URL in the `oauth_state` cookie value (or a second cookie) so the callback can reconstruct the same redirect URI and knows where to send the user after login

#### `Callback`

- Recover the base URL from the cookie set during login
- Use `baseURL + "/api/auth/callback"` as the redirect URI passed to `spotify.ExchangeCode` — this **must** match what was sent in the authorize step
- After successful token exchange, redirect to `baseURL + "/"` instead of the static `CLIENT_BASE_URL`

### Cookie approach for passing base URL

The simplest option: encode the origin base URL alongside the OAuth state. Two approaches (pick whichever is cleaner):

- **Option A**: Pack into one cookie value, e.g. `state|baseURL`, split on `|` in the callback
- **Option B**: Set a second cookie `oauth_origin` with the base URL, same max-age (300s) and attributes

Either way, the cookie must use `SameSite=Lax`, `HttpOnly=true`, `Path=/` — matching the existing `oauth_state` cookie.

### `server/config/config.go`

`CLIENT_BASE_URL` env var can also be removed if it is only used for the post-login redirect. Check for other usages first — if it's used elsewhere, leave it.

### `server/spotify/spotify.go`

No changes — `AuthorizeURL` and `ExchangeCode` already accept `redirectURI` as a parameter.

## Definition of done

- [ ] `SPOTIFY_REDIRECT_URI` env var is no longer read or required
- [ ] Hitting `/api/auth/login` from `http://10.0.0.X:8080` sends the callback URI as `http://10.0.0.X:8080/api/auth/callback` to Spotify
- [ ] Hitting `/api/auth/login` from `https://myurl.com` sends the callback URI as `https://myurl.com/api/auth/callback` to Spotify
- [ ] After successful auth, user is redirected back to the same origin they started from
- [ ] Scheme detection uses `X-Forwarded-Proto` when present (Cloudflare sets this)
- [ ] Existing `SPOTIFY_ALLOWED_USER_ID` gating still works unchanged
- [ ] `.env.example` updated to remove `SPOTIFY_REDIRECT_URI` (and `CLIENT_BASE_URL` if removed)
