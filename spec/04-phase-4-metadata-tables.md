# Phase 4: Metadata Lookup Tables

> Builds on Phase 3. The scraper now upserts track, album, and artist
> metadata from the now-playing response into lookup tables alongside each
> playback log insert. These tables store data that *is* retrievable by
> Spotify ID, giving us local access without extra API calls.

## Goal

On each poll that produces a `playback_log` row, also insert the track,
album, and artist into their respective tables if they don't already exist
(keyed on Spotify ID). All data comes from the existing now-playing
response — no additional API calls.

## Scope

### In scope
- New `artist`, `album`, `album_image`, and `track` tables.
- Expand Go structs (`Track`, `Album`, `Artist`) to parse all useful fields
  from the now-playing response.
- Scraper inserts into the new tables (insert-or-ignore) before inserting
  the `playback_log` row.
- Schema applied at boot alongside existing tables.

### Out of scope
- Updating existing rows if metadata changes (future upsert phase).
- Track-to-artist many-to-many join table (store primary artist only).
- Additional API calls to enrich metadata.
- Any frontend or analysis work.

## Data model

### `artist`

```sql
CREATE TABLE IF NOT EXISTS artist (
    spotify_id   TEXT PRIMARY KEY,
    name         TEXT NOT NULL,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

The now-playing response only provides `id` and `name` for artists. Extra
fields (genres, images, follower count) require a separate API call and are
out of scope.

### `album`

```sql
CREATE TABLE IF NOT EXISTS album (
    spotify_id    TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    album_type    TEXT,          -- "album", "single", "compilation"
    total_tracks  INTEGER,
    release_date  TEXT,          -- "2024", "2024-01", or "2024-01-15"
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### `album_image`

```sql
CREATE TABLE IF NOT EXISTS album_image (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    album_id       TEXT NOT NULL REFERENCES album(spotify_id),
    url            TEXT NOT NULL,
    width          INTEGER,
    height         INTEGER
);

CREATE INDEX IF NOT EXISTS idx_album_image_album_id ON album_image(album_id);
```

Spotify returns multiple image sizes per album (typically 640, 300, 64).
One row per image.

### `track`

```sql
CREATE TABLE IF NOT EXISTS track (
    spotify_id    TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    album_id      TEXT NOT NULL REFERENCES album(spotify_id),
    artist_id     TEXT NOT NULL REFERENCES artist(spotify_id),
    duration_ms   INTEGER NOT NULL,
    track_number  INTEGER,
    disc_number   INTEGER,
    explicit      INTEGER,       -- 1 = explicit, 0 = not
    isrc          TEXT,          -- from external_ids
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_track_album_id ON track(album_id);
CREATE INDEX IF NOT EXISTS idx_track_artist_id ON track(artist_id);
```

`artist_id` refers to the first (primary) artist in the response array.

## Struct changes

Expand the existing Go structs in `server/spotify/spotify.go`:

```go
type Artist struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type Album struct {
    ID           string   `json:"id"`
    Name         string   `json:"name"`
    AlbumType    string   `json:"album_type"`
    TotalTracks  int      `json:"total_tracks"`
    ReleaseDate  string   `json:"release_date"`
    Images       []Image  `json:"images"`
}

type Image struct {
    URL    string `json:"url"`
    Width  int    `json:"width"`
    Height int    `json:"height"`
}

type Track struct {
    ID          string      `json:"id"`
    Name        string      `json:"name"`
    Album       Album       `json:"album"`
    Artists     []Artist    `json:"artists"`
    DurationMs  int         `json:"duration_ms"`
    Popularity  int         `json:"popularity"`
    IsLocal     bool        `json:"is_local"`
    TrackNumber int         `json:"track_number"`
    DiscNumber  int         `json:"disc_number"`
    Explicit    bool        `json:"explicit"`
    ExternalIDs ExternalIDs `json:"external_ids"`
}

type ExternalIDs struct {
    ISRC string `json:"isrc"`
}
```

## Scraper changes

The poll loop (after filtering, before `playback_log` insert) adds:

1. `INSERT OR IGNORE` the primary artist (first in `Artists` array).
2. `INSERT OR IGNORE` the album.
3. For each album image, `INSERT OR IGNORE` (keyed on `album_id` + `url`).
4. `INSERT OR IGNORE` the track.
5. Insert `playback_log` row (existing behavior).

All five inserts use the same data from the single now-playing response.
No new API calls. Each insert-or-ignore is a no-op if the Spotify ID
already exists, so the overhead per poll is negligible after the first
encounter.

For album images, add a unique constraint on `(album_id, url)` to support
insert-or-ignore:

```sql
CREATE UNIQUE INDEX IF NOT EXISTS idx_album_image_unique
    ON album_image(album_id, url);
```

## Definition of done

1. All four new tables exist after scraper boot.
2. Playing a never-before-seen track inserts rows into `artist`, `album`,
   `album_image`, and `track`.
3. Playing an already-seen track inserts only a `playback_log` row — no
   duplicates or errors in the metadata tables.
4. `track.artist_id` points to the first artist in the response.
5. Album images are stored with their dimensions.
6. Existing `playback_log` behavior is unchanged.
