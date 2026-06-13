CREATE TABLE IF NOT EXISTS spotify_auth (
    id            INTEGER PRIMARY KEY CHECK (id = 1),
    access_token  TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    scope         TEXT NOT NULL,
    expiry        INTEGER NOT NULL,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO spotify_auth (id, access_token, refresh_token, scope, expiry)
SELECT 1, '', '', '', 0
WHERE NOT EXISTS (SELECT 1 FROM spotify_auth WHERE id = 1);

CREATE TABLE IF NOT EXISTS playback_log (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    polled_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    track_id        TEXT NOT NULL,
    progress_ms     INTEGER NOT NULL,
    duration_ms     INTEGER NOT NULL,
    is_playing      INTEGER NOT NULL,
    popularity      INTEGER NOT NULL,
    device_name     TEXT NOT NULL,
    device_type     TEXT NOT NULL,
    shuffle_state   INTEGER NOT NULL,
    repeat_state    TEXT NOT NULL,
    context_uri     TEXT
);

CREATE INDEX IF NOT EXISTS idx_playback_log_polled_at ON playback_log(polled_at);
CREATE INDEX IF NOT EXISTS idx_playback_log_track_id ON playback_log(track_id);
