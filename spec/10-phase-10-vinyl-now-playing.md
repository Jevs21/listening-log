# Phase 10: Vinyl Record Now-Playing Component

Builds on: Phase 9 (now-playing staleness logic)

## Goal

Replace the plain-text NowPlaying component with a top-down vinyl record visual. The album art sits in the center of the record. The record spins via CSS animation when a track is currently playing and stops when stale.

## Scope

### In scope

- Add `album_image_url` to the now-playing API response (backend query + Go struct + TS type)
- Build the vinyl record visual using plain CSS (no canvas, no SVG, no image assets)
- CSS spin animation controlled by playing/stale state
- Track info text below the record

### Out of scope

- Tonearm or other turntable visuals
- Visual differentiation for stale beyond stopping the spin (no fade, no overlay)
- Click/interaction on the record
- Responsive breakpoints (fixed size for now)

## API Changes

### `GET /api/now-playing` response

Add `album_image_url` field (nullable string):

```json
{
  "track": {
    "spotify_id": "...",
    "name": "...",
    "artist_name": "...",
    "album_name": "...",
    "duration_ms": 210000,
    "is_explicit": false,
    "updated_at": "...",
    "album_image_url": "https://i.scdn.co/image/..."
  }
}
```

### Backend changes

**`server/db/now_playing.go`**

- Add `AlbumImageURL` field to `NowPlayingTrack` struct (`*string`, json `album_image_url`)
- Join `album_image` in the query, pick the 300×300 image (or closest, ordered by `width DESC`, limit 1 via correlated subquery):

```sql
SELECT
    t.spotify_id, t.name,
    a.name AS artist_name,
    al.name AS album_name,
    t.duration_ms, t.explicit, t.updated_at,
    (SELECT ai.url FROM album_image ai
     WHERE ai.album_id = t.album_id
     ORDER BY ABS(ai.width - 300) ASC
     LIMIT 1) AS album_image_url
FROM track t
JOIN artist a  ON t.artist_id = a.spotify_id
JOIN album  al ON t.album_id  = al.spotify_id
ORDER BY t.updated_at DESC
LIMIT 1
```

**`server/handlers/now_playing.go`** — no changes needed (struct serializes automatically).

### Frontend changes

**`client/src/api/nowPlaying.ts`**

Add to `NowPlayingTrack` interface:

```ts
album_image_url: string | null;
```

**`client/src/components/NowPlaying.tsx`**

Replace the `<p>` with the vinyl record layout. Structure:

```
div.now-playing
  div.record-container
    div.record          ← the black vinyl disc, spins
      div.grooves       ← subtle concentric ring effect (CSS box-shadow or repeating-radial-gradient)
      div.label         ← center circle, holds album art
        img             ← album art (150×150, the 300px image scaled 50%)
  div.track-info
    p.track-name        ← "Track Name"
    p.track-artist      ← "by Artist — Album"
    p.track-status      ← "Now playing" or "Last played (2 mins ago)"
```

**`client/src/components/NowPlaying.css`** (new file)

CSS constants to define at the top of the component or as CSS custom properties:

| Property | Value | Notes |
|---|---|---|
| `--record-size` | `300px` | Outer diameter of the vinyl disc |
| `--label-size` | `150px` | Center label / album art diameter |
| `--spin-duration` | `1.8s` | ~33⅓ RPM (one full rotation) |

Key CSS details:

- `.record`: `border-radius: 50%`, black background, centered, `width/height: var(--record-size)`
- `.record.spinning`: `animation: spin var(--spin-duration) linear infinite`
- `.record:not(.spinning)`: no animation (default)
- `.grooves`: `repeating-radial-gradient` or concentric `box-shadow` rings in slightly lighter black/dark-gray to suggest grooves
- `.label`: centered inside `.record` via flexbox, `border-radius: 50%`, `overflow: hidden`, `width/height: var(--label-size)`
- `img` inside `.label`: `width: 100%; height: 100%; object-fit: cover`
- `@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }`

Component logic:

```tsx
const isStale = Date.now() - new Date(updated_at).getTime() > STALE_MS;
// ...
<div className={`record ${isStale ? '' : 'spinning'}`}>
```

When `album_image_url` is null, show the label circle with a solid dark-gray fill (no broken image).

## Definition of Done

- [ ] `/api/now-playing` returns `album_image_url` (300px-ish image) or `null`
- [ ] NowPlaying component renders a circular black vinyl record with visible groove texture
- [ ] Album art displays in the center at 150×150px
- [ ] Record spins at ~33⅓ RPM (`1.8s` per rotation) when track is not stale
- [ ] Record stops spinning when track is stale
- [ ] Track name, artist, album, and playing status display below the record
- [ ] No album art gracefully handled (solid label, no broken `<img>`)
- [ ] No external image assets or SVGs used — pure CSS for the record visuals
