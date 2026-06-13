# Phase 8 — ImageGrid Component

Builds on phase 6 (NowPlaying) and phase 4/5 (metadata tables including `album_image`).

## Goal

Add an `ImageGrid` component below `NowPlaying` that displays a grid of 64×64 album art thumbnails. A dropdown lets the user switch between "Recent Tracks" and "Recent Albums" orderings.

## Scope

### In scope

- New API endpoint returning album image URLs for recent tracks or albums
- React component with select dropdown and 4-column image grid
- Polling on a longer interval than NowPlaying

### Out of scope

- "Recent Artists" mode
- Click/interaction behavior on thumbnails
- Styled/responsive layout beyond bare-minimum grid

## API changes

### `GET /api/image-grid?mode={tracks|albums}`

Returns up to 50 album art URLs (64×64 size preferred).

**Query logic by mode:**

| Mode | Query |
|------|-------|
| `tracks` | Select distinct `album_image.url` joined through `track → album → album_image` ordered by `track.updated_at DESC`, deduplicated by `track.spotify_id`, filtered to `width = 64` or smallest available, limit 50 |
| `albums` | Select `album_image.url` joined through `album → album_image` ordered by `album.updated_at DESC`, deduplicated by `album.spotify_id`, filtered to `width = 64` or smallest available, limit 50 |

**Response:**

```json
{
  "images": [
    { "url": "https://i.scdn.co/image/...", "album_name": "Album Title" }
  ]
}
```

Include `album_name` for `alt` text on `<img>` tags.

**Handler:** `handlers.ImageGrid(database)` in new file `server/handlers/image_grid.go`. Validate `mode` query param, default to `tracks` if missing.

**DB function:** `db.GetImageGrid(database, mode, limit)` in new file `server/db/image_grid.go`.

For image size selection: prefer `width = 64`. If no 64px image exists for an album, fall back to the smallest available (`ORDER BY width ASC LIMIT 1` per album).

## Frontend changes

### New files

| File | Purpose |
|------|---------|
| `client/src/components/ImageGrid.tsx` | Component with dropdown + grid |
| `client/src/hooks/useImageGrid.ts` | React Query hook |
| `client/src/api/imageGrid.ts` | Fetch function |

### Constants

```ts
const IMAGE_GRID_MAX = 50;          // matches backend limit
const IMAGE_GRID_POLL_MS = 30_000;  // 30s polling interval
```

Define in `imageGrid.ts` or a shared constants file.

### Component structure

```tsx
<div>
  <select value={mode} onChange={...}>
    <option value="tracks">Recent Tracks</option>
    <option value="albums">Recent Albums</option>
  </select>
  <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 64px)", gap: "4px" }}>
    {images.map(img => (
      <img key={img.url} src={img.url} alt={img.album_name} width={64} height={64} />
    ))}
  </div>
</div>
```

### Hook

```ts
useQuery({
  queryKey: ["image-grid", mode],
  queryFn: () => fetchImageGrid(mode),
  refetchInterval: IMAGE_GRID_POLL_MS,
});
```

### Integration

Render `<ImageGrid />` in `App.tsx` below `<NowPlaying />`, inside the same `connected` conditional.

## Backend constant

```go
const ImageGridMaxResults = 50
```

Define in `server/handlers/image_grid.go` or `server/db/image_grid.go`.

## Route registration

In `server/main.go`:

```go
r.GET("/api/image-grid", handlers.ImageGrid(database))
```

## Definition of done

- [ ] `GET /api/image-grid?mode=tracks` returns up to 50 album image URLs ordered by `track.updated_at DESC`
- [ ] `GET /api/image-grid?mode=albums` returns up to 50 album image URLs ordered by `album.updated_at DESC`
- [ ] Results are deduplicated (no repeated album art in a single response)
- [ ] 64px images preferred, smallest available used as fallback
- [ ] `ImageGrid` component renders below `NowPlaying` with a 4-column grid of 64×64 thumbnails
- [ ] Dropdown switches between "Recent Tracks" and "Recent Albums" modes
- [ ] Component polls at 30s interval via React Query
- [ ] Max result count is controlled by a single constant on both backend and frontend
