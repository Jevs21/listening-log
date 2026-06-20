CREATE TABLE IF NOT EXISTS spotify_auth (
    id            INTEGER PRIMARY KEY CHECK (id = 1),
    access_token  TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    scope         TEXT NOT NULL,
    expiry        BIGINT NOT NULL,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO spotify_auth (id, access_token, refresh_token, scope, expiry)
VALUES (1, '', '', '', 0)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS playback_log (
    id              SERIAL PRIMARY KEY,
    polled_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    track_id        TEXT NOT NULL,
    progress_ms     INTEGER NOT NULL,
    duration_ms     INTEGER NOT NULL,
    is_playing      BOOLEAN NOT NULL,
    popularity      INTEGER NOT NULL,
    device_name     TEXT NOT NULL,
    device_type     TEXT NOT NULL,
    shuffle_state   BOOLEAN NOT NULL,
    repeat_state    TEXT NOT NULL,
    context_uri     TEXT
);

CREATE INDEX IF NOT EXISTS idx_playback_log_polled_at ON playback_log(polled_at);
CREATE INDEX IF NOT EXISTS idx_playback_log_track_id ON playback_log(track_id);

CREATE TABLE IF NOT EXISTS artist (
    spotify_id   TEXT PRIMARY KEY,
    name         TEXT NOT NULL,
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS album (
    spotify_id    TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    album_type    TEXT,
    total_tracks  INTEGER,
    release_date  TEXT,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS album_image (
    id             SERIAL PRIMARY KEY,
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
    explicit      BOOLEAN,
    isrc          TEXT,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_track_album_id ON track(album_id);
CREATE INDEX IF NOT EXISTS idx_track_artist_id ON track(artist_id);

CREATE TABLE IF NOT EXISTS song_suggestion (
    id         SERIAL PRIMARY KEY,
    link       TEXT    NOT NULL DEFAULT '',
    message    TEXT    NOT NULL DEFAULT '',
    source     TEXT    NOT NULL DEFAULT 'home',
    ip_address TEXT    NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_song_suggestion_ip_created ON song_suggestion(ip_address, created_at);

CREATE TABLE IF NOT EXISTS analysis_cursor (
    job_name    TEXT PRIMARY KEY,
    last_id     BIGINT NOT NULL DEFAULT 0,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS listen (
    id                  SERIAL PRIMARY KEY,
    track_id            TEXT NOT NULL,
    started_at          TIMESTAMP NOT NULL,
    ended_at            TIMESTAMP NOT NULL,
    duration_ms         INTEGER NOT NULL,
    progress_ms         INTEGER NOT NULL,
    duration_track_ms   INTEGER NOT NULL,
    poll_count          INTEGER NOT NULL,
    skipped             BOOLEAN NOT NULL,
    context_uri         TEXT,
    device_name         TEXT NOT NULL,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_listen_track_id ON listen(track_id);
CREATE INDEX IF NOT EXISTS idx_listen_started_at ON listen(started_at);
