# Phase 13 — Image Grid Tooltips

Builds on phase 8 (ImageGrid component).

## Goal

Show a custom-styled tooltip when hovering over image grid items, displaying track/album name, artist, and relative last-played time.

## Scope

### In scope

- Extend the `/api/image-grid` response to include `artist_name` and `updated_at`
- New `Tooltip` component matching the existing dark theme
- Relative time display (e.g. "3 hours ago")

### Out of scope

- Touch/mobile tooltip behavior (hover only)
- Click interactions on grid items
- Tooltip animations beyond a simple fade-in

## API changes

### `GET /api/image-grid?mode={tracks|albums}`

Add two fields to each item in the response:

```json
{
  "images": [
    {
      "url": "https://i.scdn.co/image/...",
      "album_name": "Album Title",
      "artist_name": "Artist Name",
      "updated_at": "2026-06-17T14:30:00Z"
    }
  ]
}
```

**Backend query changes in `server/db/image_grid.go`:**

| Mode | Artist source |
|------|---------------|
| `tracks` | `JOIN artist a ON a.spotify_id = t.artist_id` — select `a.name` and `t.updated_at` |
| `albums` | Join through the most recent track on that album: subquery or join `track t ON t.album_id = al.spotify_id` with `ORDER BY t.updated_at DESC LIMIT 1` per album to get artist name. Select `al.updated_at` for the timestamp. |

Update `ImageGridItem` struct:

```go
type ImageGridItem struct {
	URL        string `json:"url"`
	AlbumName  string `json:"album_name"`
	ArtistName string `json:"artist_name"`
	UpdatedAt  string `json:"updated_at"`
}
```

## Frontend changes

### New file: `client/src/components/Tooltip.tsx`

A reusable tooltip component. Renders a positioned `div` near the hovered element.

**Behavior:**
- Appears on `mouseenter`, hides on `mouseleave`
- Positioned above the target element, centered horizontally
- Falls below if there's not enough space above

**Styling (inline styles or a small CSS block in `index.css`):**
- `background: var(--surface)`
- `border: 1px solid var(--border)`
- `color: var(--text)`
- `font-size: 0.75rem`
- `padding: 0.4rem 0.6rem`
- `border-radius: 4px`
- `pointer-events: none`
- `z-index: 10`
- `white-space: nowrap`
- Subtle `opacity` transition (~150ms)

### Tooltip content layout

```
Track Name / Album Name
Artist Name · 3 hours ago
```

Line 1: `font-family: var(--font-heading)`, `color: var(--text)`
Line 2: `color: var(--text-secondary)`, `font-size: 0.7rem`

In tracks mode, line 1 shows the track name (which is the `album_name` field — but see note below). In albums mode, line 1 shows the album name.

**Note:** The current API returns `album_name` for both modes. To show the *track* name in tracks mode, either:
- Add a `track_name` field to the response (only populated in tracks mode), or
- Rename `album_name` to `name` and have it represent the track name in tracks mode and album name in albums mode.

Pick whichever is simpler. If adding `track_name`, the tooltip shows `track_name` on line 1 when present, falling back to `album_name`.

### Relative time helper

Add a small `timeAgo(dateString: string): string` utility (can live in `Tooltip.tsx` or a utils file). Handles: "just now", "X minutes ago", "X hours ago", "X days ago", "X weeks ago". No library needed.

### Changes to `ImageGrid.tsx`

Wrap each `<img>` in a container `<div>` with `position: relative`. Attach mouse event handlers that manage tooltip visibility state. Render `<Tooltip>` as a child of each grid cell (or a single tooltip repositioned on hover).

**Implementation choice:** Either one `<Tooltip>` per grid cell (simpler) or a single shared `<Tooltip>` that repositions (fewer DOM nodes). Either is fine.

### Update `client/src/api/imageGrid.ts`

Update the `ImageGridItem` type to include the new fields:

```ts
interface ImageGridItem {
  url: string;
  album_name: string;
  artist_name: string;
  updated_at: string;
  track_name?: string;  // only if using the separate field approach
}
```

## Definition of done

- [ ] `/api/image-grid` response includes `artist_name` and `updated_at` for each item
- [ ] Hovering over a grid image shows a styled tooltip with name, artist, and relative time
- [ ] Tooltip matches the app's dark theme (uses CSS variables from `index.css`)
- [ ] Tooltip positions above the image (or below if near viewport top)
- [ ] Tooltip disappears on mouse leave
- [ ] Relative time displays correctly (minutes, hours, days)
- [ ] No layout shift or overflow issues from tooltip rendering
