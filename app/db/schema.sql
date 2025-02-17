CREATE TABLE IF NOT EXISTS albums (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    release_date DATE,
    spotify_url TEXT,
    album_type TEXT,
    cover_image TEXT
);

CREATE TABLE IF NOT EXISTS artists (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    spotify_url TEXT
);

CREATE TABLE IF NOT EXISTS tracks (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    duration_ms INTEGER,
    explicit BOOLEAN,
    popularity INTEGER,
    spotify_url TEXT,
    track_number INTEGER,
    album_id TEXT NOT NULL,
    FOREIGN KEY (album_id) REFERENCES albums(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS track_artists (
    track_id TEXT NOT NULL,
    artist_id TEXT NOT NULL,
    PRIMARY KEY (track_id, artist_id),
    FOREIGN KEY (track_id) REFERENCES tracks(id) ON DELETE CASCADE,
    FOREIGN KEY (artist_id) REFERENCES artists(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS play_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    progress_ms INTEGER,
    is_playing BOOLEAN,
    device_id TEXT,
    device_name TEXT,
    volume_percent INTEGER,
    repeat_state TEXT,
    shuffle_state BOOLEAN,
    track_id TEXT NOT NULL,
    FOREIGN KEY (track_id) REFERENCES tracks(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS spotify_auth (
    id INTEGER PRIMARY KEY CHECK (id = 1),  -- Ensures a single row (ID always 1)
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    expiry INTEGER NOT NULL 
);
INSERT INTO spotify_auth (id, access_token, refresh_token, expiry) 
SELECT 1, '', '', 0 
WHERE NOT EXISTS (SELECT 1 FROM spotify_auth WHERE id = 1);