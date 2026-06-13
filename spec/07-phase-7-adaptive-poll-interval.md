# Phase 7: Adaptive Poll Interval

> Builds on Phase 3. Replaces the fixed 30s `gocron` scheduler with a
> custom polling loop that switches between fast and slow intervals based
> on whether Spotify reports an active playback state.

## Goal

Poll every 15 seconds while the user has an active playback session
(playing or paused), and every 30 seconds when nothing is reported.
This gives higher-fidelity data during listening sessions without
wasting API calls during idle periods.

## Scope

### In scope
- Remove `gocron` dependency.
- Custom polling loop with dynamic interval in `server/main.go`.
- Configurable interval constants.
- `scraper.Poll` returns a signal indicating whether playback was active.

### Out of scope
- Cooldown/hysteresis before switching back to slow (instant for now).
- Persisting or exposing the current polling mode via API.
- Changes to what gets logged or skipped — filtering rules stay the same.

## Backend changes

### Constants

Add to `server/scraper/scraper.go`:

```go
const (
    PollIntervalActive = 15 * time.Second
    PollIntervalIdle   = 30 * time.Second
)
```

### `scraper.Poll` return value

Change the signature of `Poll` from:

```go
func Poll(database *sql.DB, cfg config.Config)
```

to:

```go
func Poll(database *sql.DB, cfg config.Config) bool
```

Returns `true` when the Spotify API returned a non-nil playback state
(i.e. `PlaybackState` is not nil and `Item` is not nil), regardless of
`IsPlaying`. Returns `false` when nothing is playing, or the API call
failed.

The existing skip conditions (non-track, local file) still skip the
**database write** but should still return `true` — the user has an
active session even if we don't log the current item.

### Polling loop in `main.go`

Replace the `gocron` scheduler block with a custom loop:

```go
go func() {
    timer := time.NewTimer(0) // fire immediately on startup
    defer timer.Stop()
    for {
        <-timer.C
        active := scraper.Poll(database, cfg)
        if active {
            timer.Reset(scraper.PollIntervalActive)
        } else {
            timer.Reset(scraper.PollIntervalIdle)
        }
    }
}()
```

Remove the `gocron` import and dependency (`go mod tidy`).

## Definition of done

1. `gocron` is no longer a dependency (`go.mod`, `go.sum` updated).
2. Scraper polls immediately on startup, then at the appropriate interval.
3. While Spotify reports any playback state (playing or paused), the next
   poll fires after 15 seconds.
4. When Spotify reports nothing playing (nil / 204), the next poll fires
   after 30 seconds.
5. Skipped items (episodes, local files) still count as "active" — fast
   polling continues.
6. All existing logging and database behavior is unchanged.
7. Startup log line reflects the two intervals, e.g.
   `"scraper started — polling every 15s (active) / 30s (idle)"`.
