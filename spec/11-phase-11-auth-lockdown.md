# Phase 11: Auth Lockdown

Builds on: phase 1 (auth flow), phase 6 (status endpoint)

## Goal

Restrict Spotify OAuth so only the owner's account can authenticate. Replace the "Spotify connected" text / "Connect Spotify" button with a colored status dot at the bottom of the page.

## Scope

### In scope
- Spotify user ID allowlist (single ID via env var)
- Fetch user profile in callback to verify identity
- Log the authenticated Spotify user ID in callback (so the owner can grab it)
- Silent redirect to homepage on rejected auth
- Replace connection UI with a status dot (green=connected, red=disconnected)
- Dot is clickable: red starts auth flow, green does nothing

### Out of scope
- Multi-user support / comma-separated allowlist
- Any session or password system
- Removing the `/api/auth/login` endpoint

## Backend changes

### `server/config/config.go`

Add field to `Config`:

```go
SpotifyAllowedUserID string
```

Load from env:

```go
SpotifyAllowedUserID: os.Getenv("SPOTIFY_ALLOWED_USER_ID"),
```

No default — empty string means **no restriction** (so the app still works before the owner sets it).

### `server/spotify/spotify.go`

Add a function to fetch the current user's profile:

```go
type UserProfile struct {
    ID string `json:"id"`
}

func GetCurrentUser(accessToken string) (*UserProfile, error)
```

Calls `GET https://api.spotify.com/v1/me` with Bearer token. Returns the user ID.

### `server/handlers/auth.go`

In `Callback`, after successful token exchange:

1. Call `spotify.GetCurrentUser(token.AccessToken)` to get the user's Spotify ID
2. Log it: `log.Printf("spotify auth callback: user_id=%s", profile.ID)`
3. If `h.Cfg.SpotifyAllowedUserID` is non-empty AND `profile.ID != h.Cfg.SpotifyAllowedUserID`:
   - Log the rejection: `log.Printf("spotify auth rejected: user_id=%s not allowed", profile.ID)`
   - Redirect to `h.Cfg.ClientBaseURL` (silent redirect, no error)
   - Return early — do NOT save tokens
4. Otherwise proceed with `db.UpsertAuth` as before

## Frontend changes

### `client/src/App.tsx`

- Remove the `<p>Spotify connected</p>` line from the connected view
- Remove the `<a>/<button>Connect Spotify</button></a>` from the disconnected view
- Both views now share the same layout (title, description, content, then dot at bottom)
- Consolidate into a single return: always render `NowPlaying` and `ImageGrid` (they already handle the not-connected case gracefully since they just show nothing), with the status dot at the bottom

Add a status dot component at the bottom of the page, centered:

```tsx
<div style={{ textAlign: "center", marginTop: "2rem" }}>
  {connected ? (
    <span style={{ display: "inline-block", width: 12, height: 12, borderRadius: "50%", backgroundColor: "#22c55e" }} />
  ) : (
    <a href={`${SERVER_URL}/api/auth/login`}>
      <span style={{ display: "inline-block", width: 12, height: 12, borderRadius: "50%", backgroundColor: "#ef4444", cursor: "pointer" }} />
    </a>
  )}
</div>
```

## Definition of done

- [ ] `SPOTIFY_ALLOWED_USER_ID` env var is read into config
- [ ] Callback logs the Spotify user ID on every auth attempt
- [ ] If `SPOTIFY_ALLOWED_USER_ID` is set and the authenticating user doesn't match, tokens are NOT saved and user is silently redirected home
- [ ] If `SPOTIFY_ALLOWED_USER_ID` is empty, auth works as before (no restriction)
- [ ] "Spotify connected" text and "Connect Spotify" button are gone
- [ ] Green dot appears centered at page bottom when connected (not clickable / no action)
- [ ] Red dot appears centered at page bottom when disconnected, clicking it starts auth flow
- [ ] Server compiles and starts cleanly
