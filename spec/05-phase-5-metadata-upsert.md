# Phase 5: Metadata Upsert

> Builds on Phase 4. Metadata tables now track when rows were last seen via
> an `updated_at` column. The scraper switches from insert-or-ignore to
> upsert so that `updated_at` is refreshed on every poll, while all other
> columns remain unchanged on collision.

## Goal

Know when the scraper last encountered a given artist, album, or track
without altering any stored metadata values. This enables future queries
like "artists not seen in 90 days" or "recently active albums."

## Scope

### In scope
- Add `updated_at` column to `artist`, `album`, and `track` tables.
- Change `INSERT OR IGNORE` to `INSERT ... ON CONFLICT ... DO UPDATE` in
  the scraper, setting only `updated_at = CURRENT_TIMESTAMP` on conflict.
- Schema updated in `schema.sql` (fresh database assumed, no migration).

### Out of scope
- `album_image` table (image data doesn't meaningfully change).
- Updating any data columns on conflict (name, release_date, etc.).
- Backfilling `updated_at` for existing rows.
- Any frontend or analysis work.

## Data model changes

### `artist`

```sql
CREATE TABLE IF NOT EXISTS artist (
    spotify_id   TEXT PRIMARY KEY,
    name         TEXT NOT NULL,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### `album`

```sql
CREATE TABLE IF NOT EXISTS album (
    spotify_id    TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    album_type    TEXT,
    total_tracks  INTEGER,
    release_date  TEXT,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

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
    explicit      INTEGER,
    isrc          TEXT,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_track_album_id ON track(album_id);
CREATE INDEX IF NOT EXISTS idx_track_artist_id ON track(artist_id);
```

On first insert, `updated_at` defaults to `CURRENT_TIMESTAMP` (same as
`created_at`). On subsequent encounters, only `updated_at` is overwritten.

## Scraper changes

Replace each `INSERT OR IGNORE` with an upsert. Example for `artist`:

```go
INSERT INTO artist (spotify_id, name)
VALUES (?, ?)
ON CONFLICT(spotify_id) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
```

Same pattern for `album` and `track`. `album_image` stays as
`INSERT OR IGNORE` — no change.

## Definition of done

1. `artist`, `album`, and `track` tables include an `updated_at` column
   after scraper boot.
2. First insert sets `updated_at` equal to `created_at`.
3. Re-encountering the same Spotify ID updates only `updated_at` — all
   other columns remain unchanged.
4. `album_image` behavior is unchanged (`INSERT OR IGNORE`).
5. Existing `playback_log` behavior is unchanged.
