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

CREATE TABLE IF NOT EXISTS artist (
    spotify_id   TEXT PRIMARY KEY,
    name         TEXT NOT NULL,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS album (
    spotify_id    TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    album_type    TEXT,
    total_tracks  INTEGER,
    release_date  TEXT,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS album_image (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    album_id       TEXT NOT NULL REFERENCES album(spotify_id),
    url            TEXT NOT NULL,
    width          INTEGER,
    height         INTEGER
);

CREATE INDEX IF NOT EXISTS idx_album_image_album_id ON album_image(album_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_album_image_unique ON album_image(album_id, url);

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

CREATE TABLE IF NOT EXISTS song_suggestion (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    link       TEXT    NOT NULL DEFAULT '',
    message    TEXT    NOT NULL DEFAULT '',
    source     TEXT    NOT NULL DEFAULT 'home',
    ip_address TEXT    NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_song_suggestion_ip_created ON song_suggestion(ip_address, created_at);
