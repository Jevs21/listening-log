# Phase 17 — Incremental Listen Analysis

> Builds on Phase 3 (playback_log) and Phase 16 (Postgres). Introduces an
> incremental analysis pipeline that converts raw polls into resolved
> "listen" records for downstream visualization in tools like
> Superset/Metabase. Also refactors the `db` package to support the new
> patterns cleanly.

## Goal

Derive a `listen` table from `playback_log` where each row represents one
track play — when it started, when it ended, how far the user got, and
whether it was skipped. Process incrementally using a cursor so only new
polls are analyzed on each run.

## Scope

### In scope
- New `listen` table.
- New `analysis_cursor` table (keyed by job name, stores last processed
  `playback_log.id`). Designed to support future analysis jobs beyond
  listens.
- Go analysis worker that runs as a goroutine inside the existing server
  on a timer (every 5 minutes).
- Listen resolution logic grouping consecutive polls into single listens.
- Refactor: wrap `*sql.DB` in a `db.DB` struct so all database methods
  become receivers instead of free functions taking `*sql.DB`.
- Refactor: add transaction helper to `db.DB` for atomic multi-step
  operations.
- Refactor: wrap scraper metadata + playback inserts in a transaction.

### Out of scope
- Read replica setup (separate spec).
- Other analysis tables (daily aggregates, time-of-day patterns, etc.) —
  future specs will reuse the cursor infrastructure.
- Frontend/API for the `listen` table.
- Backfilling historical `playback_log` data (the worker will process it
  naturally on first run — it just starts from cursor 0).

## Refactors

### 1. `db.DB` struct wrapper

Currently every function in `server/db/*.go` takes `*sql.DB` as its first
argument:

```go
func InsertPlaybackLog(database *sql.DB, log PlaybackLog) error
func UpsertArtist(database *sql.DB, artist spotify.Artist) error
func GetNowPlaying(database *sql.DB) (*NowPlayingResult, error)
// ... etc
```

Introduce a thin wrapper:

```go
// server/db/db.go
type DB struct {
    *sql.DB
}

func Open(connStr string) (*DB, error) {
    sqlDB, err := sql.Open("pgx", connStr)
    if err != nil {
        return nil, err
    }
    // apply schema...
    return &DB{sqlDB}, nil
}
```

Then convert all existing functions to methods on `*DB`:

```go
func (d *DB) InsertPlaybackLog(log PlaybackLog) error
func (d *DB) UpsertArtist(artist spotify.Artist) error
func (d *DB) GetNowPlaying() (*NowPlayingResult, error)
```

This removes the stutter of passing `db` everywhere and makes it natural
to add new methods for analysis queries. All callers (handlers, scraper,
analysis worker) receive a `*db.DB` instead of `*sql.DB`.

### 2. Transaction helper

Add a `WithTx` method to `db.DB` that wraps a callback in a transaction:

```go
func (d *DB) WithTx(fn func(tx *sql.Tx) error) error {
    tx, err := d.Begin()
    if err != nil {
        return err
    }
    if err := fn(tx); err != nil {
        tx.Rollback()
        return err
    }
    return tx.Commit()
}
```

Database methods that need to participate in transactions should accept
an optional `*sql.Tx` or use an `Executor` interface:

```go
type Executor interface {
    ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
    QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
    QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
```

This lets the same method work standalone or inside a transaction.

### 3. Wrap scraper writes in a transaction

The scraper currently runs 5 independent inserts per poll cycle
(artist → album → album images → track → playback_log). Wrap these in a
single transaction using `WithTx` so a failure partway through doesn't
leave partial state.

## Data model

### `analysis_cursor`

```sql
CREATE TABLE IF NOT EXISTS analysis_cursor (
    job_name    TEXT PRIMARY KEY,
    last_id     BIGINT NOT NULL DEFAULT 0,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

One row per analysis job. For this phase, one row: `job_name = 'listen'`.
Future analysis phases insert their own row without schema changes.

### `listen`

```sql
CREATE TABLE IF NOT EXISTS listen (
    id                  SERIAL PRIMARY KEY,
    track_id            TEXT NOT NULL,
    started_at          TIMESTAMP NOT NULL,
    ended_at            TIMESTAMP NOT NULL,
    duration_ms         INTEGER NOT NULL,     -- ended_at - started_at in ms
    progress_ms         INTEGER NOT NULL,     -- max progress_ms reached across polls
    duration_track_ms   INTEGER NOT NULL,     -- track's total duration (from playback_log.duration_ms)
    poll_count          INTEGER NOT NULL,     -- number of playback_log rows in this listen
    skipped             BOOLEAN NOT NULL,     -- progress_ms < 10% of duration_track_ms
    context_uri         TEXT,                 -- from first poll in the group
    device_name         TEXT NOT NULL,        -- from first poll in the group
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_listen_track_id ON listen(track_id);
CREATE INDEX IF NOT EXISTS idx_listen_started_at ON listen(started_at);
```

### Column rationale

| Column | Why |
|---|---|
| `duration_ms` | Wall-clock time spent on this listen (includes pauses). |
| `progress_ms` | How far into the track the user actually got. More accurate than wall-clock for "did they hear it?" |
| `duration_track_ms` | Denominator for skip detection. Denormalized here to avoid joining `playback_log` or `track` when querying. |
| `poll_count` | Useful for data quality — a listen with 1 poll is lower confidence than one with 20. |
| `skipped` | Pre-computed flag: `progress_ms < 0.10 * duration_track_ms`. |
| `context_uri` | What playlist/album triggered playback. Taken from the first poll. |
| `device_name` | Which device. Taken from the first poll. |

## Backend changes

### Modified: `server/db/db.go`

- Introduce `DB` struct wrapping `*sql.DB` (see Refactors §1).
- Add `Executor` interface (see Refactors §2).
- Add `WithTx` method (see Refactors §2).
- `Open()` returns `*DB` instead of `*sql.DB`.

### Modified: all existing `server/db/*.go` files

Convert free functions to `*DB` methods. Update signatures to accept
`Executor` where they may participate in transactions (metadata and
playback inserts).

### Modified: `server/handlers/*.go`

Update to receive `*db.DB` instead of `*sql.DB`. Call methods on the
struct instead of passing db as first arg.

### Modified: `server/scraper/scraper.go`

- Accept `*db.DB` instead of `*sql.DB`.
- Wrap the metadata + playback insert sequence in `db.WithTx`.

### Modified: `server/main.go`

- `db.Open()` now returns `*db.DB` — update variable type.
- Start analysis worker after DB init.

### New file: `server/db/analysis.go`

Methods on `*DB`:

- `GetAnalysisCursor(jobName) -> lastID` — returns `last_id` for the
  given job, or 0 if no row exists.
- `SetAnalysisCursor(ex Executor, jobName, lastID)` — upserts the cursor
  row. Accepts `Executor` so it can run inside a transaction.
- `InsertListen(ex Executor, listen)` — inserts a single `listen` row.
  Accepts `Executor` for transaction support.
- `GetUnprocessedPolls(afterID, limit) -> []PlaybackLogRow` — fetches
  `playback_log` rows with `id > afterID` ordered by `id ASC`, with a
  limit (e.g., 10,000) to bound memory.

### New file: `server/analysis/listen.go`

The listen resolution logic:

```
func ProcessNewPolls(db) error:
    1. cursor = db.GetAnalysisCursor("listen")
    2. polls = db.GetUnprocessedPolls(cursor, 10000)
    3. if empty, return
    4. listens, newCursor = resolveListens(polls)
    5. db.WithTx: for each listen → InsertListen(tx, listen)
                  then SetAnalysisCursor(tx, "listen", newCursor)
```

The batch of listen inserts + cursor update happens atomically in one
transaction. If anything fails, the cursor doesn't advance and the next
run re-processes.

### Listen resolution algorithm

Walk the polls in `id` order (which is `polled_at` order). Maintain a
"current listen" accumulator:

1. **Start a new listen** when:
   - There is no current listen (first poll, or after closing the previous).
   - `track_id` differs from the current listen's `track_id`.
   - Gap between this poll's `polled_at` and the previous poll's `polled_at`
     exceeds **120 seconds**.

2. **Extend the current listen** when:
   - Same `track_id` and gap ≤ 120 seconds.
   - Update `ended_at`, `poll_count`, and `progress_ms` (track the max
     `progress_ms` seen).

3. **Close & emit a listen** when:
   - A new listen starts (step 1) — close the previous one.
   - End of the batch — but **do not close** the last listen in progress.
     It may continue in the next batch. Instead, set the cursor to the
     `id` of the last poll that belongs to a *completed* listen. The
     in-progress listen's polls will be re-read next run.

This ensures no listen is split across batches or prematurely closed.

### Computed fields on close

- `duration_ms` = `ended_at - started_at` in milliseconds.
- `skipped` = `progress_ms < 0.10 * duration_track_ms`.

### New file: `server/analysis/worker.go`

A goroutine started from `main.go`:

```go
func StartWorker(db *db.DB, interval time.Duration) {
    ticker := time.NewTicker(interval)
    go func() {
        for range ticker.C {
            if err := ProcessNewPolls(db); err != nil {
                log.Printf("listen analysis error: %v", err)
            }
        }
    }()
}
```

Called from `main.go` after DB init:

```go
analysis.StartWorker(database, 5*time.Minute)
```

### `server/db/schema.sql`

Append the `analysis_cursor` and `listen` table definitions.

## Definition of done

- [ ] `db.DB` struct wrapper is in place; all existing functions converted to methods.
- [ ] `Executor` interface and `WithTx` helper exist and work.
- [ ] Scraper metadata + playback inserts wrapped in a transaction.
- [ ] All handlers and scraper updated to use `*db.DB`.
- [ ] `analysis_cursor` and `listen` tables exist after server boot.
- [ ] Worker goroutine starts automatically with the server.
- [ ] After the worker runs, `listen` rows appear for completed track plays.
- [ ] A track played across two worker cycles is not split into two listens.
- [ ] A track where max `progress_ms` < 10% of `duration_track_ms` has `skipped = true`.
- [ ] `analysis_cursor` advances only past fully resolved listens (in-progress listen polls are re-read).
- [ ] Listen inserts + cursor update are atomic (single transaction).
- [ ] Processing 0 new polls is a no-op (no errors, no empty inserts).
- [ ] Worker logs errors but does not crash the server.
- [ ] Running the worker against an existing `playback_log` with thousands of rows processes them incrementally in batches.
- [ ] Existing functionality (auth, now-playing, image grid, suggestions, scraper) is unaffected.
