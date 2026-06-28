# Phase 31 — Data Export / Import CLI

Builds on: Phase 16 (Dockerize & Postgres), Phase 17 (Incremental Listen Analysis)

## Goal

A Go CLI tool that exports core database tables to a tarball of CSVs and imports them back with either **overwrite** or **merge** semantics. Runs inside Docker — no UI exposure. After any import, the `listen` table is truncated and analysis restarts from scratch.

## Scope

### In

- Export `playback_log`, `artist`, `album`, `album_image`, `track`, `song_suggestion` to CSV
- Bundle CSVs into a `.tar.gz` archive
- Import from a `.tar.gz` with two modes: `--mode=overwrite` and `--mode=merge`
- Reset analysis state (`listen`, `analysis_cursor`) after import
- Go CLI at `cmd/dbtools/main.go`, invoked via `docker compose exec`

### Out

- UI or API access to export/import
- Export of derived tables (`listen`, `analysis_cursor`)
- Scheduled/automatic backups

## CLI Interface

```bash
# Export
docker compose exec app dbtools export /data/backup.tar.gz

# Import — overwrite
docker compose exec app dbtools import --mode=overwrite /data/backup.tar.gz

# Import — merge
docker compose exec app dbtools import --mode=merge /data/backup.tar.gz
```

The `/data` path maps to a bind-mount volume so archives are accessible from the host.

## File Structure

```
server/
  cmd/
    dbtools/
      main.go          # CLI entry point (flag parsing, subcommand dispatch)
  dbtools/
    export.go          # Export logic
    import.go          # Import logic (overwrite + merge)
    csv.go             # CSV read/write helpers per table
    reset.go           # Truncate listen + reset analysis_cursor
```

## Export Format

Archive: `backup.tar.gz` containing:

```
playback_log.csv
artist.csv
album.csv
album_image.csv
track.csv
song_suggestion.csv
manifest.json          # { "exported_at": "...", "tables": { "playback_log": 48210, ... } }
```

Each CSV uses headers matching column names. All timestamps formatted as RFC 3339. Booleans as `true`/`false`. NULLs as empty strings.

### Column order per table

| Table | Columns |
|---|---|
| `playback_log` | `id, polled_at, track_id, progress_ms, duration_ms, is_playing, popularity, device_name, device_type, shuffle_state, repeat_state, context_uri` |
| `artist` | `spotify_id, name, created_at, updated_at` |
| `album` | `spotify_id, name, album_type, total_tracks, release_date, created_at, updated_at` |
| `album_image` | `id, album_id, url, width, height` |
| `track` | `spotify_id, name, album_id, artist_id, duration_ms, track_number, disc_number, explicit, isrc, created_at, updated_at` |
| `song_suggestion` | `id, link, message, source, ip_address, created_at` |

## Import — Overwrite Mode

1. Begin transaction
2. Truncate in FK-safe order: `album_image`, `track`, `album`, `artist`, `playback_log`, `song_suggestion`, `listen`, `analysis_cursor`
3. Load CSVs using `COPY FROM` in FK-safe order: `artist` → `album` → `album_image` → `track` → `playback_log` → `song_suggestion`
4. Reset sequences (`setval`) to max ID + 1 for serial columns
5. Commit transaction
6. Print row counts per table

## Import — Merge Mode

Per-table merge strategy:

| Table | Dedup key | Conflict resolution |
|---|---|---|
| `playback_log` | `polled_at` | Skip if `polled_at` already exists |
| `artist` | `spotify_id` | Update if imported `updated_at` is newer |
| `album` | `spotify_id` | Update if imported `updated_at` is newer |
| `album_image` | `(album_id, url)` | Skip if pair already exists |
| `track` | `spotify_id` | Update if imported `updated_at` is newer |
| `song_suggestion` | `(ip_address, created_at)` | Skip if pair already exists |

Implementation: load CSV into a temp table, then `INSERT ... ON CONFLICT` with appropriate logic per table.

After merge completes:
1. Truncate `listen`
2. Reset `analysis_cursor` to `last_id = 0`
3. Print per-table counts: inserted, updated, skipped

## Docker Changes

### `docker-compose.yml`

Add a bind-mount to the `app` service for archive access:

```yaml
volumes:
  - ./backups:/data
```

### `Dockerfile`

Build the `dbtools` binary alongside the main server binary:

```dockerfile
RUN go build -o /app/dbtools ./cmd/dbtools
```

Ensure it's copied to the final stage.

## Analysis Reset (`reset.go`)

```sql
TRUNCATE TABLE listen;
INSERT INTO analysis_cursor (job_name, last_id, updated_at)
VALUES ('listen', 0, NOW())
ON CONFLICT (job_name) DO UPDATE SET last_id = 0, updated_at = NOW();
```

The existing 5-minute analysis worker will pick up from ID 0 on its next tick and re-derive all `listen` rows.

## Definition of Done

- [ ] `docker compose exec app dbtools export /data/backup.tar.gz` produces a valid archive with all 6 tables + manifest
- [ ] Exported CSVs are valid, human-readable, and round-trip cleanly (export → overwrite import → export produces identical CSVs)
- [ ] `--mode=overwrite` truncates all tables (including `listen`/`analysis_cursor`) and loads data
- [ ] `--mode=merge` inserts new rows, updates metadata with newer `updated_at`, skips duplicates
- [ ] Both import modes reset analysis state so `listen` is re-derived
- [ ] Sequences are reset correctly after overwrite (new inserts get IDs above imported max)
- [ ] Import aborts cleanly on error (transaction rollback, no partial state)
- [ ] `./backups/` directory is bind-mounted and accessible from host
- [ ] `dbtools` binary is built and available in the Docker image
