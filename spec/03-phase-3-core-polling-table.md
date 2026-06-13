# Phase 3: Core Polling Table

> Builds on Phase 2. The scraper now writes each poll result to a
> `playback_log` table instead of just printing to stdout. This table is the
> permanent source of truth for all future analysis and cannot be backfilled,
> so the schema is designed to be final.

## Goal

Every 30 seconds, if a non-local track is playing (or paused), insert a row
capturing the playback state at that moment. Skip polls where nothing is
playing, an episode/ad is playing, or a local file is playing.

## Scope

### In scope
- New `playback_log` table in the shared SQLite database.
- Scraper writes a row on each successful poll (instead of / in addition to
  stdout).
- Schema applied at boot alongside existing `spotify_auth` schema.

### Out of scope
- Derived/analysis tables (Phase 4+).
- Deduplication or session grouping — raw polls only.
- Podcast/episode tracking.
- Local file tracking.
- Lookup tables for track/album/artist metadata (Phase 4+).

## Data model

### Design principle

Store only what cannot be retrieved later from the Spotify API by ID, plus
temporal and device state unique to the moment of the poll. Track names,
artist names, album metadata, and artwork are all retrievable by ID and
belong in later lookup/cache tables.

### Schema

Add to `server/db/schema.sql`:

```sql
CREATE TABLE IF NOT EXISTS playback_log (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    polled_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    track_id        TEXT NOT NULL,       -- spotify track ID (e.g. "4iV5W9uYEdYUVa79Axb7Rh")
    progress_ms     INTEGER NOT NULL,    -- playback position in the track
    duration_ms     INTEGER NOT NULL,    -- total track length
    is_playing      INTEGER NOT NULL,    -- 1 = playing, 0 = paused
    popularity      INTEGER NOT NULL,    -- 0-100, changes over time
    device_name     TEXT NOT NULL,       -- e.g. "Kitchen speaker"
    device_type     TEXT NOT NULL,       -- e.g. "computer", "smartphone"
    shuffle_state   INTEGER NOT NULL,    -- 1 = on, 0 = off
    repeat_state    TEXT NOT NULL,       -- "off", "context", or "track"
    context_uri     TEXT                 -- e.g. "spotify:playlist:37i9d..." (nullable, can be absent)
);

CREATE INDEX IF NOT EXISTS idx_playback_log_polled_at ON playback_log(polled_at);
CREATE INDEX IF NOT EXISTS idx_playback_log_track_id ON playback_log(track_id);
```

### Column rationale

| Column | Why it's here |
|---|---|
| `polled_at` | Our clock, not Spotify's. When we observed this state. |
| `track_id` | Key to fetch all track/album/artist metadata from Spotify later. |
| `progress_ms` | Reconstructs skip behavior, partial listens, replays. |
| `duration_ms` | Needed alongside progress to compute % complete without a lookup. |
| `is_playing` | Distinguishes active listening from paused-on-screen. |
| `popularity` | Changes over time; cannot be backfilled historically. |
| `device_name` | Which device was active. Not retrievable after the fact. |
| `device_type` | Device category. Not retrievable after the fact. |
| `shuffle_state` | Playback mode. Not retrievable after the fact. |
| `repeat_state` | Playback mode. Not retrievable after the fact. |
| `context_uri` | What triggered playback (playlist, album, artist page). Private playlists and generated mixes disappear — can't be recovered. |

### What's deliberately omitted

| Field | Reason |
|---|---|
| Track/artist/album names | Retrievable by `track_id` via Spotify API. |
| Artist IDs | Retrievable from track metadata. |
| Album ID, art, release date | Retrievable from track metadata. |
| `isrc` / external IDs | Retrievable from track metadata. |
| `explicit`, `track_number`, `disc_number` | Retrievable from track metadata. |
| `album_type`, `total_tracks` | Retrievable from track metadata. |
| Spotify `timestamp` field | Redundant with our own `polled_at`. |
| `available_markets` | Irrelevant to listening history. |
| `href`, `external_urls`, `uri` | Derivable from IDs. |
| `preview_url`, `is_playable`, `linked_from`, `restrictions` | Not useful for history. |
| `actions` block | Describes available controls, not listening data. |
| `device.id`, `volume_percent`, etc. | Noise. Name and type are sufficient. |

## Scraper changes

The scraper (from Phase 2) gains a DB write step in its poll loop:

1. Call currently-playing endpoint.
2. If response is empty, `currently_playing_type` is not `"track"`, or
   `item.is_local` is `true` → skip (no row inserted, no stdout).
3. Otherwise, insert a row into `playback_log`.
4. Continue printing to stdout as before.

The scraper should also apply the `playback_log` schema on startup, same
pattern as the server does for `spotify_auth`.

## Storage estimate

At 30-second intervals, ~2,880 rows/day, ~1M rows/year. Each row is
roughly 200 bytes → ~200 MB/year. Well within SQLite's comfort zone.

## Definition of done

1. `playback_log` table exists after server or scraper boot.
2. Running the scraper while music plays inserts one row per poll cycle.
3. Paused tracks still get logged (with `is_playing = 0`).
4. Local files, episodes, ads, and "nothing playing" produce no rows.
5. `polled_at` reflects when the scraper observed the state, not Spotify's
   timestamp.
6. Existing stdout logging still works.
7. Indexes on `polled_at` and `track_id` are present.
